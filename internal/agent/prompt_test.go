package agent

import (
	"strings"
	"testing"
)

func TestBuildUserPromptIncludesObservedAndHints(t *testing.T) {
	prompt, err := BuildUserPrompt(PromptInput{
		Directory:   "/tmp/demo",
		Instruction: "group files",
		Observed: DirectorySnapshot{
			ObservedFiles:   []string{"os/unit1.pdf"},
			ObservedFolders: []string{"os"},
		},
		UserHints: []string{"course names"},
	})
	if err != nil {
		t.Fatalf("BuildUserPrompt returned error: %v", err)
	}

	if !strings.Contains(prompt, "\"observed_names\"") {
		t.Fatalf("expected observed_names in prompt: %s", prompt)
	}
	if !strings.Contains(prompt, "\"user_hints\"") {
		t.Fatalf("expected user_hints in prompt: %s", prompt)
	}
}
