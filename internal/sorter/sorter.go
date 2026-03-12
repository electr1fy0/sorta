package sorter

import (
	"context"
	"path/filepath"

	"github.com/electr1fy0/sorta/internal/config"
	"github.com/electr1fy0/sorta/internal/core"
)

type ConfigSorter struct {
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

func (s *ConfigSorter) Decide(ctx context.Context, files []core.FileEntry) ([]core.FileOperation, error) {
	ops := make([]core.FileOperation, 0, 10)

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		filename := filepath.Base(file.SourcePath)
		destFolder := config.Categorize(*s.configData, filename)

		if destFolder == "" {
			ops = append(ops, core.FileOperation{OpType: core.OpSkip})
		} else {
			ops = append(ops, core.FileOperation{
				OpType:   core.OpMove,
				File:     file,
				Size:     file.Size,
				DestPath: filepath.Join(file.RootDir, destFolder, filename),
			})
		}
	}

	return ops, nil
}

func (s *ConfigSorter) GetBlacklist() []string {
	return s.configData.Blacklist
}
