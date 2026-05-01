package agent

import "encoding/json"

type DirectorySnapshot struct {
	FileCount       int            `json:"file_count"`
	FolderCount     int            `json:"folder_count"`
	ObservedFiles   []string       `json:"observed_files"`
	ObservedFolders []string       `json:"observed_folders"`
	ExtensionCounts map[string]int `json:"extension_counts"`
}

type PromptInput struct {
	Directory   string            `json:"directory"`
	Instruction string            `json:"instruction"`
	Observed    DirectorySnapshot `json:"observed_names"`
	UserHints   []string          `json:"user_hints"`
}

func GoalSystemPrompt() string {
	return `You are the planning engine for the sorta CLI.

Return exactly one JSON object and nothing else.

The JSON object must have this shape:
{
  "summary": "short summary",
  "request_filenames": ["Ask for filenames or examples if needed"],
  "actions": [
    {"kind": "dedupe"},
    {"kind": "sort_rule", "sort_rule": {"folder": "OS", "keywords": ["os", "operating systems"]}},
    {"kind": "rename", "rename": {"files": [], "hints": ["all caps"]}}
  ]
}

Rules:
- Allowed action kinds are only: "dedupe", "sort_rule", "rename".
- If the goal is about sorting, grouping, or categorizing and you do not have enough filename examples to infer good rules, return one or more short prompts in "request_filenames" and leave "actions" empty for this turn.
- Use "sort_rule" for grouping behavior.
- For renaming the whole directory, set "rename.files" to an empty array.
- For renaming a subset, "rename.files" must be relative file paths.
- For the "rename" action, use the "hints" field (string array) to pass specific renaming instructions from the goal (e.g., "uppercase", "remove dates", "add prefix X").
- Do not include any extra top-level keys.
- Do not edit or mention persistent config files.
- Keep plans minimal and directly executable.
- If the request cannot be completed safely, return an empty actions array with a short explanation in "summary".`
}

func BuildUserPrompt(input PromptInput) (string, error) {
	data, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
