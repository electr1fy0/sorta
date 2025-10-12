package internal

const (
	ansiReset  = "[0m"
	ansiRed    = "[31m"
	ansiGreen  = "[32m"
	ansiYellow = "[33m"
)

type ConfigData struct {
	foldernames []string
	keywords    [][]string
}

type Sorter interface {
	Sort(baseDir, dir string, filename string, size int64) (FileOperation, error)
}

type OperationType int

const (
	OpMove OperationType = iota
	OpDelete
	OpSkip
)

type FileOperation struct {
	Type       OperationType
	SourcePath string
	DestPath   string
	Filename   string
	Size       int64
}

type Executor struct {
	DryRun bool
}

type ExtensionSorter struct {
	categories map[string][]string
}

type ConfigSorter struct {
	configData *ConfigData
}

type DuplicateFinder struct {
	hashes map[string]string
}

type SortResult struct {
	Moved   int
	Skipped int
	Errors  []error
}

type Reporter struct {
	DryRun bool
}

type FileInfo struct {
	Name string
	Size int64
}
