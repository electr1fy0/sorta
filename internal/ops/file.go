package ops

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/electr1fy0/sorta/internal/core"
	"github.com/electr1fy0/sorta/internal/ignore"
)

var (
	DuplNuke         = false
	RecurseLevel int = 1 << 10
)

func FilterFiles(rootDir string, sorter core.Sorter, executor *Executor, reporter *Reporter) (*core.SortResult, error) {
	operations, err := PlanOperations(rootDir, sorter)
	if err != nil {
		return nil, err
	}

	return ApplyOperations(rootDir, operations, executor, reporter)
}

func PlanOperations(rootDir string, sorter core.Sorter) ([]core.FileOperation, error) {
	return PlanOperationsWithIgnoreCtx(context.Background(), rootDir, sorter, nil)
}

func PlanOperationsWithIgnore(rootDir string, sorter core.Sorter, ignoreMatcher *ignore.IgnoreMatcher) ([]core.FileOperation, error) {
	return PlanOperationsWithIgnoreCtx(context.Background(), rootDir, sorter, ignoreMatcher)
}

func PlanOperationsWithIgnoreCtx(ctx context.Context, rootDir string, sorter core.Sorter, ignoreMatcher *ignore.IgnoreMatcher) ([]core.FileOperation, error) {
	var files []core.FileEntry

	walkErr := WalkFilesWithIgnoreCtx(ctx, rootDir, ignoreMatcher, func(file core.FileEntry) error {
		files = append(files, file)
		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	operations, err := sorter.Decide(ctx, files)
	if err != nil {
		return nil, err
	}
	sortOperationsDeterministically(operations)
	return operations, nil
}

func ApplyOperations(rootDir string, operations []core.FileOperation, executor *Executor, reporter *Reporter) (*core.SortResult, error) {
	return ApplyOperationsCtx(context.Background(), rootDir, operations, executor, reporter)
}

type rollbackAction struct {
	From string
	To   string
}

func ApplyOperationsCtx(ctx context.Context, rootDir string, operations []core.FileOperation, executor *Executor, reporter *Reporter) (*core.SortResult, error) {
	result := &core.SortResult{}
	txnDir, err := createTransactionDir(rootDir)
	if err != nil {
		return result, fmt.Errorf("failed to create transaction dir: %w", err)
	}
	rollback := make([]rollbackAction, 0, len(operations)+1)
	failWithRollback := func(baseErr error, actions []rollbackAction) error {
		rollbackErr := rollbackAll(actions)
		if rollbackErr != nil {
			return fmt.Errorf("%w (rollback failed: %v)", baseErr, rollbackErr)
		}
		return baseErr
	}

	for _, op := range operations {
		if err := ctx.Err(); err != nil {
			return result, failWithRollback(fmt.Errorf("operation cancelled: %w", err), rollback)
		}

		moved, rb, err := applyAtomicOperation(op, executor, txnDir, len(rollback))
		if moved || err != nil {
			reporter.Report(op, err)
		}

		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", filepath.Base(op.File.SourcePath), err))
			return result, failWithRollback(fmt.Errorf("failed to apply operations: %w", err), append(rollback, rb...))
		}
		rollback = append(rollback, rb...)

		if moved {
			switch op.OpType {
			case core.OpMove:
				result.Moved++
			case core.OpDedupe:
				result.Deduped++
			case core.OpRename:
				result.Renamed++
			case core.OpDelete:
				result.Deleted++
			}
		}
		if op.OpType == core.OpSkip {
			result.Skipped++
		}
	}

	var nukedCount int
	if DuplNuke {
		nc, stagedRollback, err := stageDuplicateNuke(rootDir, txnDir)
		nukedCount = nc
		if err != nil {
			return result, failWithRollback(fmt.Errorf("failed to stage duplicates folder: %w", err), rollback)
		}
		rollback = append(rollback, stagedRollback...)
	}

	id := time.Now().UTC().Format(time.RFC3339Nano)
	transaction := core.Transaction{TType: core.TAction, Operations: operations, ID: id, Irreversible: DuplNuke}
	if err := LogToHistory(transaction); err != nil {
		return result, failWithRollback(fmt.Errorf("failed to log history: %w", err), rollback)
	}

	if err := os.RemoveAll(txnDir); err != nil {
		return result, fmt.Errorf("failed to finalize transaction cleanup: %w", err)
	}

	if err := cleanEmptyFolders(rootDir); err != nil {
		return result, err
	}

	if DuplNuke {
		result.Deleted += nukedCount
	}
	return result, nil
}

func cleanEmptyFolders(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			if err := cleanEmptyFolders(path); err != nil {
				return err
			}

			subEntries, err := os.ReadDir(path)
			if err != nil {
				continue
			}

			onlyDSStore := len(subEntries) == 1 && subEntries[0].Name() == ".DS_Store"
			if onlyDSStore {
				_ = os.Remove(filepath.Join(path, ".DS_Store"))
			}
			if len(subEntries) == 0 || onlyDSStore {
				if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove empty dir %q: %w", path, err)
				}
			}
		}
	}

	return nil
}

func TopLargestFiles(rootDir string, n int) error {
	var entries []core.FileEntry
	err := WalkFiles(rootDir, func(file core.FileEntry) error {
		entries = append(entries, file)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No files found.")
		return nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Size > entries[j].Size
	})

	limit := min(len(entries), n)
	fmt.Printf("Top %d largest files in %s:\n", limit, rootDir)
	for _, e := range entries[:limit] {
		rel, err := filepath.Rel(rootDir, e.SourcePath)
		if err != nil {
			rel = e.SourcePath
		}
		fmt.Printf("  %-10s  %s\n", core.HumanReadable(e.Size), rel)
	}
	return nil
}

func WalkFiles(rootDir string, fn func(core.FileEntry) error) error {
	return WalkFilesWithIgnoreCtx(context.Background(), rootDir, nil, fn)
}

func WalkFilesWithIgnore(rootDir string, ignoreMatcher *ignore.IgnoreMatcher, fn func(core.FileEntry) error) error {
	return WalkFilesWithIgnoreCtx(context.Background(), rootDir, ignoreMatcher, fn)
}

func WalkFilesWithIgnoreCtx(ctx context.Context, rootDir string, ignoreMatcher *ignore.IgnoreMatcher, fn func(core.FileEntry) error) error {
	return filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if ignoreMatcher != nil && ignoreMatcher.Match(rootDir, path, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(rootDir, path)
		if err == nil && d.IsDir() && rel != "." {
			depth := strings.Count(rel, string(os.PathSeparator)) + 1
			if depth > RecurseLevel {
				return filepath.SkipDir
			}
		}

		if d.IsDir() {
			return nil
		}

		stat, err := d.Info()
		if err != nil {
			return err
		}

		size := stat.Size()
		return fn(core.FileEntry{RootDir: rootDir, SourcePath: path, Size: size})
	})
}

func sortOperationsDeterministically(ops []core.FileOperation) {
	sort.SliceStable(ops, func(i, j int) bool {
		a := ops[i]
		b := ops[j]

		aSrc := filepath.Clean(a.File.SourcePath)
		bSrc := filepath.Clean(b.File.SourcePath)
		if aSrc != bSrc {
			return aSrc < bSrc
		}

		aDst := filepath.Clean(a.DestPath)
		bDst := filepath.Clean(b.DestPath)
		if aDst != bDst {
			return aDst < bDst
		}

		if a.OpType != b.OpType {
			return a.OpType < b.OpType
		}

		return a.Size < b.Size
	})
}

func createTransactionDir(rootDir string) (string, error) {
	txnID := time.Now().UTC().Format("20060102T150405.000000000Z")
	txnDir := filepath.Join(rootDir, ".sorta", "transactions", txnID)
	if err := os.MkdirAll(txnDir, 0755); err != nil {
		return "", err
	}
	return txnDir, nil
}

func stageDuplicateNuke(rootDir, txnDir string) (int, []rollbackAction, error) {
	duplicatePath := filepath.Join(rootDir, "duplicates")
	duplicates, _ := os.ReadDir(duplicatePath)
	nukedCount := len(duplicates)

	var rollback []rollbackAction
	if _, err := os.Stat(duplicatePath); err == nil {
		staged := filepath.Join(txnDir, "nuked-duplicates")
		if err := os.Rename(duplicatePath, staged); err != nil {
			return 0, nil, fmt.Errorf("failed to rename duplicates to staged: %w", err)
		}
		rollback = append(rollback, rollbackAction{From: staged, To: duplicatePath})
	}
	return nukedCount, rollback, nil
}

func applyAtomicOperation(op core.FileOperation, executor *Executor, txnDir string, idx int) (bool, []rollbackAction, error) {
	switch op.OpType {
	case core.OpMove, core.OpDedupe, core.OpRename:
		moved, err := executor.Execute(op)
		if err != nil || !moved {
			return moved, nil, err
		}
		return true, []rollbackAction{{From: op.DestPath, To: op.File.SourcePath}}, nil
	case core.OpDelete:
		src := op.File.SourcePath
		if src == "" {
			return false, nil, fmt.Errorf("cannot delete: empty source path")
		}
		if _, err := os.Stat(src); err != nil {
			if os.IsNotExist(err) {
				return false, nil, nil
			}
			return false, nil, err
		}
		staged := filepath.Join(txnDir, "deletes", fmt.Sprintf("%06d_%s", idx, filepath.Base(src)))
		if err := os.MkdirAll(filepath.Dir(staged), 0755); err != nil {
			return false, nil, err
		}
		if err := os.Rename(src, staged); err != nil {
			return false, nil, err
		}
		return true, []rollbackAction{{From: staged, To: src}}, nil
	case core.OpSkip:
		return false, nil, nil
	default:
		return false, nil, fmt.Errorf("unsupported operation type: %v", op.OpType)
	}
}

func rollbackAll(actions []rollbackAction) error {
	var rollbackErrors []error
	for i := len(actions) - 1; i >= 0; i-- {
		a := actions[i]
		if a.From == "" || a.To == "" {
			continue
		}
		if _, err := os.Stat(a.From); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			rollbackErrors = append(rollbackErrors, err)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(a.To), 0755); err != nil {
			rollbackErrors = append(rollbackErrors, err)
			continue
		}
		if err := os.Rename(a.From, a.To); err != nil {
			rollbackErrors = append(rollbackErrors, err)
		}
	}
	if len(rollbackErrors) > 0 {
		return fmt.Errorf("%d rollback actions failed", len(rollbackErrors))
	}
	return nil
}
