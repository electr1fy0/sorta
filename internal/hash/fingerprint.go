package hash

import (
	"os"
)

type FileFingerprint struct {
	Size    int64  `json:"size"`
	ModTime int64  `json:"modTime"`
	Inode   uint64 `json:"inode"`
}

func GetFingerprint(path string) (FileFingerprint, error) {
	info, err := os.Stat(path)
	if err != nil {
		return FileFingerprint{}, err
	}

	return FileFingerprint{
		Size:    info.Size(),
		ModTime: info.ModTime().UnixNano(),
		Inode:   inodeFromFileInfo(info),
	}, nil
}
