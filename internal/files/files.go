// Package files is the local-filesystem adapter for ensure.FS.
package files

import (
	"io/fs"
	"os"
)

// OS is the local filesystem. The zero value is ready to use.
type OS struct{}

// Lstat returns file info without following a symlink.
func (OS) Lstat(path string) (fs.FileInfo, error) {
	return os.Lstat(path)
}

// Readlink returns the destination of a symlink.
func (OS) Readlink(path string) (string, error) {
	return os.Readlink(path)
}

// ReadDir lists names in a directory.
func (OS) ReadDir(path string) ([]fs.DirEntry, error) {
	return os.ReadDir(path)
}

// MkdirAll creates path and any missing parents.
func (OS) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Chmod sets mode bits on path.
func (OS) Chmod(path string, perm fs.FileMode) error {
	return os.Chmod(path, perm)
}

// Rename moves src to dst.
func (OS) Rename(src, dst string) error {
	return os.Rename(src, dst)
}

// Remove deletes a file or empty directory.
func (OS) Remove(path string) error {
	return os.Remove(path)
}

// RemoveAll deletes path and everything under it.
func (OS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Symlink creates newname pointing at oldname.
func (OS) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}
