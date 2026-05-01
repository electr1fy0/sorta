package rename

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/electr1fy0/sorta/internal/core"
	"github.com/electr1fy0/sorta/internal/llm"
	"github.com/electr1fy0/sorta/templates"
)

var defaultPrompt = templates.DefaultPrompt

type Renamer struct {
	client llm.Client
	model  string
	hints  []string
}

func NewRenamer(client llm.Client, model string, hints []string) *Renamer {
	return &Renamer{
		client: client,
		model:  model,
		hints:  append([]string(nil), hints...),
	}
}

func (r *Renamer) Decide(ctx context.Context, files []core.FileEntry) ([]core.FileOperation, error) {
	if len(files) == 0 {
		return nil, nil
	}

	allNewNames := make([]string, 0, len(files))
	batchSize := 10

	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		end = min(len(files), end)

		batch := files[i:end]
		batchFilenames := make([]string, len(batch))
		for j, f := range batch {
			batchFilenames[j] = filepath.Base(f.SourcePath)
		}

		marshalledPayload, err := json.Marshal(batchFilenames)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal filenames: %w", err)
		}

		userPrompt := ""
		if len(r.hints) > 0 {
			hintsJSON, err := json.Marshal(r.hints)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal rename hints: %w", err)
			}
			userPrompt += "USER_HINTS: " + string(hintsJSON) + "\n"
		}
		userPrompt += string(marshalledPayload) + "\nOutput the results as a JSON object."

		raw, err := r.client.Run(ctx, llm.Request{
			Model:        r.model,
			SystemPrompt: defaultPrompt + "\n\nCRITICAL: Return exactly one JSON object with a 'filenames' key containing the renamed strings and nothing else.",
			UserPrompt:   userPrompt,
		})
		if err != nil {
			return nil, fmt.Errorf("rename request failed for batch starting at %d: %w", i, err)
		}
		raw = strings.TrimSpace(raw)

		var response struct {
			Filenames []string `json:"filenames"`
		}
		if err := json.Unmarshal([]byte(raw), &response); err != nil {
			return nil, fmt.Errorf("failed to parse AI response for batch starting at %d: %w. Raw output: %s", i, err, raw)
		}

		if len(response.Filenames) != len(batch) {
			return nil, fmt.Errorf("integrity error in batch starting at %d: sent %d files, received %d names. Raw output: %s", i, len(batch), len(response.Filenames), raw)
		}

		allNewNames = append(allNewNames, response.Filenames...)
	}

	ops := make([]core.FileOperation, 0, len(files))
	seen := make(map[string]bool)

	for i, newName := range allNewNames {
		originalName := filepath.Base(files[i].SourcePath)
		if strings.TrimSpace(newName) == "" {
			newName = originalName
		}

		base := newName
		ext := filepath.Ext(newName)
		nameNoExt := strings.TrimSuffix(base, ext)
		counter := 1

		for seen[newName] {
			newName = fmt.Sprintf("%s_v%d%s", nameNoExt, counter, ext)
			counter++
		}
		seen[newName] = true
		destPath := filepath.Join(filepath.Dir(files[i].SourcePath), newName)

		op := core.FileOperation{
			OpType:   core.OpRename,
			File:     files[i],
			DestPath: destPath,
			Size:     files[i].Size,
		}
		ops = append(ops, op)
	}

	return ops, nil
}
