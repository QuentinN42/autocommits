package github

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/QuentinN42/autocommits/pkg/types"
	"github.com/google/go-github/v58/github"
)

type RateLimit struct {
	Remaining int
	Reset     time.Time
}

type GitHub struct {
	username   string
	repository string

	client    *github.Client
	rateLimit RateLimit
}

func New(ctx context.Context) (*GitHub, error) {
	username := os.Getenv("GH_USER")
	if username == "" {
		return nil, fmt.Errorf("GH_USER is not set")
	}
	repository := os.Getenv("GH_REPO")
	if repository == "" {
		return nil, fmt.Errorf("GH_REPO is not set")
	}
	token := os.Getenv("GH_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GH_REPO is not set")
	}
	client := github.NewClient(nil).WithAuthToken(token)
	gh := &GitHub{
		username:   username,
		repository: repository,
		client:     client,
		rateLimit:  RateLimit{},
	}
	err := gh.updateRateLimit(ctx)
	if err != nil {
		return nil, err
	}
	return gh, nil
}

func (w *GitHub) updateRateLimit(ctx context.Context) error {
	rateLimit, _, err := w.client.RateLimit.Get(ctx)
	if err != nil {
		return err
	}
	w.rateLimit.Remaining = rateLimit.Core.Remaining
	w.rateLimit.Reset = rateLimit.Core.Reset.Time
	return nil
}

func (w *GitHub) canRequest(ctx context.Context) (bool, error) {
	if w.rateLimit.Reset.Before(time.Now()) {
		err := w.updateRateLimit(ctx)
		if err != nil {
			return false, err
		}
	}
	return w.rateLimit.Remaining > 100, nil
}

func (w *GitHub) Work(ctx context.Context, todo types.Todo) {
	// logger.Debug(ctx, "Working on %s", todo.ID)
}
