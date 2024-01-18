package github

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/QuentinN42/autocommits/pkg/types"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v58/github"
)

const commitPrefix = "auto: "
const changesFile = "changes"

func changeTheFile(basepath string) error {
	path := filepath.Join(basepath, changesFile)
	// read the changes file
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	content, err := io.ReadAll(f)
	f.Close()
	if err != nil {
		return err
	}
	var toWrite string
	if content == nil || len(content) <= 0 || string(content) == "true" {
		toWrite = "false"
	} else {
		toWrite = "true"
	}
	f, err = os.Open(path)
	if err != nil {
		return err
	}
	f.WriteString(toWrite)
	f.Close()
	return nil
}

func (gh *GitHub) createPR(ctx context.Context, todo types.Todo) (*github.PullRequest, error) {
	r, err := git.PlainOpen(gh.repositoryPath)
	if err != nil {
		return nil, err
	}
	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	// check for uncommited changes
	status, err := w.Status()
	if err != nil {
		return nil, err
	}
	if !status.IsClean() {
		return nil, fmt.Errorf("Uncommited changes %s", status.String())
	}

	// checkout master
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.Master,
	})
	if err != nil {
		return nil, err
	}

	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil {
		if err != git.NoErrAlreadyUpToDate {
			return nil, err
		}
	}

	// create new branch
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(todo.ID),
		Create: true,
	})

	err = changeTheFile(gh.repositoryPath)
	if err != nil {
		return nil, err
	}

	// commit the changes
	_, err = w.Commit(commitPrefix+todo.ID, &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name: gh.username,
			When: todo.Date,
		},
	})

	// push set upstream
	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/*:refs/heads/*"},
	})
	if err != nil {
		return nil, err
	}

	res, _, err := gh.client.PullRequests.Create(ctx, gh.username, gh.repository, &github.NewPullRequest{
		Title: github.String(todo.ID),
		Head:  github.String(todo.ID),
		Base:  github.String("master"),
	})

	return res, err
}

func (gh *GitHub) getAllCommits(ctx context.Context) ([]string, error) {
	r, err := git.PlainOpen(gh.repositoryPath)
	if err != nil {
		return []string{}, err
	}
	w, err := r.Worktree()
	if err != nil {
		return []string{}, err
	}

	// check for uncommited changes
	status, err := w.Status()
	if err != nil {
		return []string{}, err
	}
	if !status.IsClean() {
		return []string{}, fmt.Errorf("Uncommited changes %s", status.String())
	}

	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil {
		if err != git.NoErrAlreadyUpToDate {
			return []string{}, err
		}
	}

	// get all commits
	ref, err := r.Head()
	if err != nil {
		return []string{}, err
	}
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return []string{}, err
	}

	var commits []string
	err = cIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c.Message)
		return nil
	})
	if err != nil {
		return []string{}, err
	}
	return commits, nil
}

func needToWork(todo types.Todo, commits []string) bool {
	for _, c := range commits {
		if len(c) < len(commitPrefix) {
			continue
		}
		if c[:len(commitPrefix)] != commitPrefix {
			continue
		}
		// get the first line
		if strings.Contains(c, "\n") {
			c = c[:strings.Index(c, "\n")]
		}
		c = c[len(commitPrefix):]
		if c == todo.ID {
			return false
		}
	}
	return true
}
