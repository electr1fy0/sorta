package internal

import (
	"regexp"
)

const (
	ansiReset  = "[0m"
	ansiRed    = "[31m"
	ansiGreen  = "[32m"
	ansiYellow = "[33m"
)

type ConfigData struct {
	Foldernames []string
	Matchers    [][]Matcher
	Blacklist   []string
}

type Matcher struct {
	Raw   string
	Regex *regexp.Regexp
}

type Sorter interface {
	Decide(filePaths []FileEntry) ([]FileOperation, error)
}

type OperationType int

const (
	OpMove OperationType = iota
	OpDelete
	OpSkip
	OpUndo
)

type TransactionType int

const (
	TAction TransactionType = iota
	TUndo
)

type Transaction struct {
	ID           string
	TType        TransactionType
	Operations   []FileOperation
	Irreversible bool
}

type FileEntry struct {
	RootDir    string
	SourcePath string
	Size       int64
}

type FileOperation struct {
	OpType   OperationType
	File     FileEntry
	DestPath string
	Size     int64
}

type Executor struct {
	DryRun     bool
	Operations []FileOperation
	Blacklist  []string
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
	Deleted int
	Errors  []error
}

type Reporter struct {
	DryRun bool
}
