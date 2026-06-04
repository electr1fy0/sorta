package sorter

import (
	"context"
	"testing"

	"github.com/electr1fy0/sorta/internal/config"
	"github.com/electr1fy0/sorta/internal/core"
)

func TestConfigSorterDecide(t *testing.T) {
	t.Run("moves files to matching folders", func(t *testing.T) {
		s, err := NewRuleSorter([]config.RuleSpec{
			{Folder: "Images", Keywords: []string{"jpg"}},
			{Folder: "Docs", Keywords: []string{"pdf"}},
		})
		if err != nil {
			t.Fatal(err)
		}

		files := []core.FileEntry{
			{RootDir: "/dir", SourcePath: "/dir/photo.jpg", Size: 100},
			{RootDir: "/dir", SourcePath: "/dir/doc.pdf", Size: 200},
			{RootDir: "/dir", SourcePath: "/dir/script.go", Size: 300},
		}

		ops, err := s.Decide(context.Background(), files)
		if err != nil {
			t.Fatal(err)
		}
		if len(ops) != 3 {
			t.Fatalf("expected 3 ops, got %d", len(ops))
		}

		if ops[0].OpType != core.OpMove || ops[0].DestPath != "/dir/Images/photo.jpg" {
			t.Errorf("unexpected first op: %+v", ops[0])
		}
		if ops[1].OpType != core.OpMove || ops[1].DestPath != "/dir/Docs/doc.pdf" {
			t.Errorf("unexpected second op: %+v", ops[1])
		}
		if ops[2].OpType != core.OpSkip {
			t.Errorf("expected skip for unmatched file, got %+v", ops[2])
		}
	})

	t.Run("handles empty file list", func(t *testing.T) {
		s, err := NewRuleSorter([]config.RuleSpec{
			{Folder: "Images", Keywords: []string{"jpg"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		ops, err := s.Decide(context.Background(), nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(ops) != 0 {
			t.Errorf("expected 0 ops, got %d", len(ops))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		s, err := NewRuleSorter([]config.RuleSpec{
			{Folder: "Images", Keywords: []string{"jpg"}},
		})
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err = s.Decide(ctx, []core.FileEntry{
			{RootDir: "/dir", SourcePath: "/dir/photo.jpg"},
		})
		if err == nil {
			t.Error("expected error from cancelled context")
		}
	})
}

func TestNewRuleSorterErrors(t *testing.T) {
	_, err := NewRuleSorter([]config.RuleSpec{})
	if err == nil {
		t.Error("expected error for empty rules")
	}
}

func TestGetBlacklist(t *testing.T) {
	s, err := NewRuleSorter([]config.RuleSpec{
		{Folder: "Images", Keywords: []string{"jpg"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if bl := s.GetBlacklist(); bl == nil {
		t.Error("expected non-nil blacklist")
	}
}
