package cs

import (
	"fmt"

	"github.com/go-git/go-git/v5"
)

func CloneRepository(fs *FileSystem, url string, path string) (*git.Repository, error) {
	repo, err := git.Clone(fs.Storer, nil, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		return nil, fmt.Errorf("error cloning repository: %v", err)
	}

	return repo, nil
}
