// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"fmt"
	"os"

	"github.com/codesphere-cloud/cs-go/pkg/cs"
	"github.com/go-git/go-git/v5"
	//"github.com/go-git/go-git/v5/plumbing"
)

type Git interface {
	CloneRepository(fs *cs.FileSystem, url string, branch string, path string) (*git.Repository, error)
}

type GitService struct {
	fs *cs.FileSystem
}

func NewGitService(fs *cs.FileSystem) *GitService {
	return &GitService{
		fs: fs,
	}
}

func (g *GitService) CloneRepository(fs *cs.FileSystem, url string, branch string, path string) (*git.Repository, error) {
	repo, err := git.PlainClone(path, false, &git.CloneOptions{
		//repo, err := git.Clone(fs.Storer, fs, &git.CloneOptions{
		//ReferenceName: plumbing.NewBranchReferenceName(branch),
		Progress: os.Stdout,
		URL:      url,
	})
	if err != nil {
		return nil, fmt.Errorf("error cloning repository: %w", err)
	}

	return repo, nil
}
