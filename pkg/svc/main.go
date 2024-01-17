package svc

import (
	"context"

	"github.com/QuentinN42/autocommits/pkg/gitlab"
	"github.com/QuentinN42/autocommits/pkg/logger"
)

type Service struct {
	// ...
}

func New(ctx context.Context) *Service {
	return &Service{}
}

func (*Service) Run(ctx context.Context) {
	_, err := gitlab.GetCommits(ctx, "QuentinN42")
	if err != nil {
		logger.Logger.Error("Could not fetch commits from GitLab", "error", err)
	}
}
