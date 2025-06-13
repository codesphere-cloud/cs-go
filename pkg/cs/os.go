package cs

import (
	"fmt"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/memory"
)

type FileSystem struct {
	storage.Storer
	billy.Filesystem
}

func NewOSFileSystem(root string) *FileSystem {
	system := osfs.New(root)
	return &FileSystem{
		filesystem.NewStorage(system, nil),
		system,
	}
}

func NewMemFileSystem() *FileSystem {
	return &FileSystem{
		memory.NewStorage(),
		memfs.New(),
	}
}

func (f *FileSystem) FileExists(filename string) bool {
	info, err := f.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func (f *FileSystem) DirExists(dirname string) bool {
	info, err := f.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}

	return info.IsDir()
}

func (f *FileSystem) CreateDirectory(dirname string) error {
	if f.DirExists(dirname) {
		return nil
	}

	err := os.MkdirAll(dirname, 0755)
	if err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	return nil
}

func (f *FileSystem) CreateFile(filename string) (billy.File, error) {
	if f.FileExists(filename) {
		return nil, fmt.Errorf("file already exists: %s", filename)
	}

	file, err := f.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %v", err)
	}

	return file, nil
}
