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
	Sort(filePaths []FilePath) ([]FileOperation, error)
}

type OperationType int

const (
	OpMove OperationType = iota
	OpDelete
	OpSkip
)

type Transaction struct {
	Operations []FileOperation
	ID         string
	Root       string
	Type       TransactionType
}

type FileOperation struct {
	Type       OperationType
	SourcePath string
	DestPath   string
	Filename   string
	Size       int64
}

type Executor struct {
	DryRun      bool
	Interactive bool
	Operations  []FileOperation
	Blacklist   []string
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

type Renamer struct {
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
