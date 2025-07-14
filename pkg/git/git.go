// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"fmt"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/go-git/go-git/v5"
)

type Git interface {
	CloneRepository(fs *cs.FileSystem, url string, path string) (*git.Repository, error)
}

type GitService struct {
	fs *cs.FileSystem
}

func NewGitService(fs *cs.FileSystem) Git {
	return &GitService{
		fs: fs,
	}
}

func (g *GitService) CloneRepository(fs *cs.FileSystem, url string, path string) (*git.Repository, error) {
	repo, err := git.Clone(fs.Storer, nil, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		return nil, fmt.Errorf("error cloning repository: %w", err)
	}

	return repo, nil
}
