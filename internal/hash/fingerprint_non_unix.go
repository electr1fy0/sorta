//go:build !unix

package hash

import "os"

func inodeFromFileInfo(_ os.FileInfo) uint64 {
	return 0
}
