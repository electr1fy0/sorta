//go:build unix

package hash

import (
	"os"
	"syscall"
)

func inodeFromFileInfo(info os.FileInfo) uint64 {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0
	}
	return stat.Ino
}
