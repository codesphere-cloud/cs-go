// Copyright (c) Codesphere Inc.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/codesphere-cloud/cs-go/pkg/util"
	"github.com/go-git/go-git/v5"
	//"github.com/go-git/go-git/v5/plumbing"
)

type Git interface {
	CloneRepository(fs *util.FileSystem, url string, branch string, path string) (*git.Repository, error)
	GetRemoteUrl(remoteName string) (string, error)
	Checkout(branch string, createBranch bool) error
	AddAll() error
	HasChanges(remote string, branch string) (bool, error)
	PrintStatus() error
	Commit(message string) error
	Push(remote string, branch string) error
}

type GitService struct {
	fs *util.FileSystem
}

// Checkout checks out the specified branch, if createBranch is true, it will create the branch
func (g *GitService) Checkout(branch string, createBranch bool) error {
	cmd := gitCommand("checkout", branch)
	if createBranch {
		cmd = gitCommand("checkout", "-b", branch)
	}
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}
	return nil
}

// AddAll stages all changes in the repository using `git add .`
func (g *GitService) AddAll() error {
	cmd := gitCommand("add", ".")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}
	return nil
}

// HasChanges checks if there are any changes compared to the specified remote and branch without printing the diff
func (g *GitService) HasChanges(remote string, branch string) (bool, error) {
	cmd := gitCommand("diff", remote+"/"+branch, "--cached", "--exit-code", "--quiet")
	err := cmd.Run()
	if err == nil {
		return false, nil
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() == 1 {
			return true, nil
		}
	}
	return false, fmt.Errorf("failed to check for changes compared to remote: %w", err)
}

// PrintStatus prints the output of `git status --short` to stdout
func (g *GitService) PrintStatus() error {
	cmd := gitCommand("status", "--short")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to print git status: %w", err)
	}
	return nil
}

// Commit commits the changes with the specified commit message
func (g *GitService) Commit(message string) error {
	cmd := gitCommand("commit", "-m", message)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}
	return nil
}

// Push pushes the changes to the specified remote and branch
func (g *GitService) Push(remote string, branch string) error {
	cmd := gitCommand("push", remote, "HEAD:"+branch)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}
	return nil
}

// GetCurrentBranch returns the name of the current branch
func gitCommand(args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("Running git command: git %s\n", strings.Join(args, " "))
	return cmd
}

// NewGitService creates a new GitService with the provided FileSystem
func NewGitService(fs *util.FileSystem) *GitService {
	return &GitService{
		fs: fs,
	}
}

// CloneRepository clones the repository from the specified URL and branch to the specified path using go-git library
func (g *GitService) CloneRepository(fs *util.FileSystem, url string, branch string, path string) (*git.Repository, error) {
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

// GetRemoteUrl returns the URL of the specified remote using go-git library
func (g *GitService) GetRemoteUrl(remoteName string) (string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("error opening repository: %w", err)
	}
	remote, err := repo.Remote(remoteName)
	if err != nil {
		return "", fmt.Errorf("error getting remote: %w", err)
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return "", fmt.Errorf("no URLs found for remote %s", remoteName)
	}

	return urls[0], nil
}
