package internal

const (
	ansiReset  = "[0m"
	ansiRed    = "[31m"
	ansiGreen  = "[32m"
	ansiYellow = "[33m"
)

type ConfigData struct {
	Foldernames []string
	Keywords    [][]string
	Blacklist   []string
}

type Sorter interface {
	Decide(filePaths []FileEntry) ([]FileOperation, error)
}

type OperationType int

const (
	OpMove OperationType = iota
	OpDelete
	OpSkip
	OpUndo
)

type TransactionType int

const (
	TAction TransactionType = iota
	TUndo
)

type Transaction struct {
	ID         string
	TType      TransactionType
	Operations []FileOperation
}

type FileEntry struct {
	RootDir    string
	SourcePath string
	Size       int64
}

type FileOperation struct {
	OpType   OperationType
	File     FileEntry
	DestPath string
	Size     int64
}

type Executor struct {
	DryRun      bool
	Interactive bool
	Operations  []FileOperation
	Blacklist   []string
}

type ConfigSorter struct {
	configData *ConfigData
}

type DuplicateFinder struct {
	hashes map[string]string
}

type Renamer struct {
}

type SortResult struct {
	Moved   int
	Skipped int
	Deleted int
	Errors  []error
}

type Reporter struct {
	DryRun bool
}
