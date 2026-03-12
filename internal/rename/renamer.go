package rename

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/electr1fy0/sorta/internal/core"
	"github.com/electr1fy0/sorta/templates"
	"google.golang.org/genai"
)

var defaultPrompt = templates.DefaultPrompt

type Renamer struct{}

func NewRenamer() *Renamer {
	return &Renamer{}
}

func (r *Renamer) Decide(ctx context.Context, files []core.FileEntry) ([]core.FileOperation, error) {
	if os.Getenv("GEMINI_API_KEY") == "" {
		return nil, fmt.Errorf("Missing GEMINI_API_KEY environment variable")
	}
	if len(files) == 0 {
		return nil, nil
	}

	filenames := make([]string, len(files))

	for i, f := range files {
		filenames[i] = filepath.Base(f.SourcePath)
	}

	marshalledPayload, err := json.Marshal(filenames)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal filenames: %w", err)
	}

	prompt := defaultPrompt

	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}
	status := make(chan struct{})
	go func() {
		dots := []string{"", ".", "..", "..."}
		i := 0

		for {
			select {
			case <-status:
				return

			default:
				fmt.Printf("\rConversing with Gemini%s     ", dots[i])
				i = (i + 1) % len(dots)
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash-lite", genai.Text(prompt+"\n"+string(marshalledPayload)), nil)
	close(status)
	fmt.Print("\r                             \n")

	if err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}

	raw := resp.Text()
	raw = strings.TrimSpace(raw)

	var newnames []string
	if err := json.Unmarshal([]byte(raw), &newnames); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w. Raw output: %s", err, raw)
	}

	if len(newnames) != len(files) {
		return nil, fmt.Errorf("integrity error: sent %d files, received %d names", len(files), len(newnames))
	}

	ops := make([]core.FileOperation, 0, len(files))
	seen := make(map[string]bool)

	for i, newName := range newnames {
		if strings.TrimSpace(newName) == "" {
			newName = filenames[i]
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
