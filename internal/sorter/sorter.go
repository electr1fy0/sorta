package sorter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/electr1fy0/sorta/internal/config"
	"github.com/electr1fy0/sorta/internal/core"
)

type ConfigSorter struct {
	configData *config.ConfigData
}

type RuleSorter struct {
	configData *config.ConfigData
}

func NewConfigSorter(folderPath, configPath, inline string) (*ConfigSorter, error) {
	if inline != "" {
		confData, err := config.ParseInline(inline)
		if err != nil {
			return nil, err
		}
		return &ConfigSorter{configData: confData}, nil
	}

	confData, _, err := config.LoadConfig(configPath, folderPath)
	if err != nil {
		return nil, err
	}
	return &ConfigSorter{configData: confData}, nil
}

func NewRuleSorter(rules []config.RuleSpec) (*RuleSorter, error) {
	confData, err := config.BuildConfigData(rules)
	if err != nil {
		return nil, err
	}
	return &RuleSorter{configData: confData}, nil
}

func (s *ConfigSorter) Decide(ctx context.Context, files []core.FileEntry) ([]core.FileOperation, error) {
	return decide(ctx, s.configData, files)
}

func (s *RuleSorter) Decide(ctx context.Context, files []core.FileEntry) ([]core.FileOperation, error) {
	return decide(ctx, s.configData, files)
}

func decide(ctx context.Context, configData *config.ConfigData, files []core.FileEntry) ([]core.FileOperation, error) {
	ops := make([]core.FileOperation, 0, len(files))
	seen := make(map[string]bool)

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		filename := filepath.Base(file.SourcePath)
		destFolder := config.Categorize(*configData, filename)

		if destFolder == "" {
			ops = append(ops, core.FileOperation{OpType: core.OpSkip, File: file})
			continue
		}

		name := filename
		ext := filepath.Ext(name)
		stem := strings.TrimSuffix(name, ext)
		counter := 1
		destPath := filepath.Join(file.RootDir, destFolder, name)
		for seen[destPath] {
			name = fmt.Sprintf("%s_v%d%s", stem, counter, ext)
			destPath = filepath.Join(file.RootDir, destFolder, name)
			counter++
		}

		if _, err := os.Stat(destPath); err == nil {
			ops = append(ops, core.FileOperation{OpType: core.OpSkip, File: file})
			continue
		}

		seen[destPath] = true
		ops = append(ops, core.FileOperation{
			OpType:   core.OpMove,
			File:     file,
			Size:     file.Size,
			DestPath: destPath,
		})
	}

	return ops, nil
}

func (s *ConfigSorter) GetBlacklist() []string {
	return s.configData.Blacklist
}

func (s *RuleSorter) GetBlacklist() []string {
	return s.configData.Blacklist
}
