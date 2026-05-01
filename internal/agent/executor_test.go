package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExecutePlanCreatesFoldersAndMovesFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.Setenv("GEMINI_API_KEY", "test-key"); err != nil {
		t.Fatal(err)
	}

	filePath := filepath.Join(dir, "os_notes.pdf")
	if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	plan := ExecutionPlan{
		Summary: "Create pyqs and group OS files",
		Actions: []PlannedAction{
			{Kind: ActionSortRule, SortRule: &SortRuleAction{Folder: "OS", Keywords: []string{"os"}}},
		},
	}

	if err := ExecutePlan(context.Background(), dir, plan); err != nil {
		t.Fatalf("ExecutePlan returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "pyqs")); err != nil {
		t.Fatalf("expected pyqs directory to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "OS", "os_notes.pdf")); err != nil {
		t.Fatalf("expected sorted file to exist: %v", err)
	}
}
