package svc

import (
	"context"

	"github.com/QuentinN42/autocommits/pkg/github"
	"github.com/QuentinN42/autocommits/pkg/gitlab"
	"github.com/QuentinN42/autocommits/pkg/logger"
)

type Service struct {
	gl *gitlab.Gitlab
	gh *github.GitHub
}

func New(ctx context.Context) (*Service, error) {
	gl, err := gitlab.New(ctx)
	if err != nil {
		return nil, err
	}
	gh, err := github.New(ctx)
	if err != nil {
		return nil, err
	}
	return &Service{
		gl: gl,
		gh: gh,
	}, nil
}

func (s *Service) Run(ctx context.Context) error {
	todos, err := s.gl.GetCommits(ctx)
	if err != nil {
		logger.Error(ctx, "Could not fetch commits from GitLab: %v", err)
		return err
	}

	for _, todo := range todos {
		s.gh.Work(ctx, todo)
	}
	return nil
}
