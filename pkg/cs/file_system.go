// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package cs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/cache"
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
		filesystem.NewStorage(system, cache.NewObjectLRUDefault()),
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

	err := f.MkdirAll(dirname, 0755)
	if err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	return nil
}

func (f *FileSystem) CreateFile(filename string) (billy.File, error) {
	if f.FileExists(filename) {
		return nil, fmt.Errorf("file already exists: %s", filename)
	}

	file, err := f.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %w", err)
	}

	return file, nil
}

// WriteFile creates a file at the specified path and writes data to it.
// If the directory does not exist, it will be created.
// If the file in the directory already exists, it returns an error.
func (f *FileSystem) WriteFile(path string, filename string, data []byte, force bool) error {
	if !f.DirExists(path) {
		err := f.CreateDirectory(path)
		if err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
	}

	fullfilename := filepath.Join(path, filename)
	exists := f.FileExists(fullfilename)

	if exists && !force {
		return fmt.Errorf("file already exists: %s", fullfilename)
	}

	file, err := f.Create(fullfilename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}

	return nil
}
