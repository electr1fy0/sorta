package core

import (
	"os"
)

type FileSystem interface {
	Rename(oldpath, newpath string) error
	Remove(path string) error
	RemoveAll(path string) error
	MkdirAll(path string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)
	ReadDir(name string) ([]os.DirEntry, error)
	IsNotExist(err error) bool
}

type OSFileSystem struct{}

func (OSFileSystem) Rename(oldpath, newpath string) error { return os.Rename(oldpath, newpath) }
func (OSFileSystem) Remove(path string) error             { return os.Remove(path) }
func (OSFileSystem) RemoveAll(path string) error          { return os.RemoveAll(path) }
func (OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
func (OSFileSystem) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }
func (OSFileSystem) ReadDir(name string) ([]os.DirEntry, error) {
	return os.ReadDir(name)
}
func (OSFileSystem) IsNotExist(err error) bool { return os.IsNotExist(err) }
