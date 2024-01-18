package github

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/QuentinN42/autocommits/pkg/logger"
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

	repositoryPath string

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
	repositoryPath := os.Getenv("GH_REPO_PATH")
	if repositoryPath == "" {
		return nil, fmt.Errorf("GH_REPO is not set")
	}
	_, err := os.Stat(repositoryPath)
	if err != nil {
		logger.Error(ctx, "GH_REPO_PATH is not a valid path: %v", err)
		return nil, fmt.Errorf("GH_REPO_PATH is not a valid path")
	}
	token := os.Getenv("GH_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GH_REPO is not set")
	}
	client := github.NewClient(nil).WithAuthToken(token)
	gh := &GitHub{
		username:       username,
		repository:     repository,
		repositoryPath: repositoryPath,
		client:         client,
		rateLimit:      RateLimit{},
	}
	err = gh.updateRateLimit(ctx)
	if err != nil {
		return nil, err
	}
	return gh, nil
}

func (g *GitHub) updateRateLimit(ctx context.Context) error {
	rateLimit, _, err := g.client.RateLimit.Get(ctx)
	if err != nil {
		return err
	}
	g.rateLimit.Remaining = rateLimit.Core.Remaining
	g.rateLimit.Reset = rateLimit.Core.Reset.Time
	logger.Debug(ctx, "Rate limit: %d until %s", g.rateLimit.Remaining, g.rateLimit.Reset)
	return nil
}

func (g *GitHub) canRequest(ctx context.Context) (bool, error) {
	if g.rateLimit.Reset.Before(time.Now()) {
		err := g.updateRateLimit(ctx)
		if err != nil {
			return false, err
		}
	}
	return g.rateLimit.Remaining > 100, nil
}

func (g *GitHub) createIssue(ctx context.Context, todo types.Todo) (*github.Issue, error) {
	issueReq := &github.IssueRequest{
		Title:    &todo.ID,
		Assignee: &g.username,
	}
	issue, _, err := g.client.Issues.Create(ctx, g.username, g.repository, issueReq)
	return issue, err
}

func (g *GitHub) Work(ctx context.Context, todo types.Todo, commits []string) error {
	logger.Trace(ctx, "Working on %s", todo.ID)
	if !needToWork(todo, commits) {
		logger.Info(ctx, "%s already processed", todo.ID)
		return nil
	}
	ok, err := g.canRequest(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("Rate limit exceeded")
	}
	pr, err := g.createPR(ctx, todo)
	if err != nil {
		logger.Error(ctx, "Could not create PR: %v", err)
		return err
	}
	logger.Debug(ctx, "Created PR: %s", *pr.HTMLURL)
	issue, err := g.createIssue(ctx, todo)
	if err != nil {
		logger.Error(ctx, "Could not create issue: %v", err)
		return err
	}
	logger.Debug(ctx, "Created issue: %s", *issue.HTMLURL)

	err = g.cleanIssuePR(ctx, issue, pr)
	if err != nil {
		return err
	}
	logger.Debug(ctx, "Issue and PR linked")

	return nil
}

func (g *GitHub) cleanIssuePR(ctx context.Context, issue *github.Issue, pr *github.PullRequest) error {
	// set the user as assignee on the PR
	_, _, err := g.client.Issues.AddAssignees(ctx, g.username, g.repository, *pr.Number, []string{g.username})
	if err != nil {
		return err
	}

	// set the user as reviewer on the PR
	_, _, err = g.client.PullRequests.RequestReviewers(ctx, g.username, g.repository, *pr.Number, github.ReviewersRequest{
		Reviewers: []string{g.username},
	})
	if err != nil {
		return err
	}

	// approve the PR
	ok := "APPROVE"
	_, _, err = g.client.PullRequests.CreateReview(ctx, g.username, g.repository, *pr.Number, &github.PullRequestReviewRequest{
		Event: &ok,
	})
	if err != nil {
		return err
	}

	return nil
}

func (g *GitHub) WorkAll(ctx context.Context, todos []types.Todo) (err error) {
	commits, err := g.getAllCommits(ctx)
	if err != nil {
		return err
	}
	for _, todo := range todos {
		err = g.Work(ctx, todo, commits)
		if err != nil {
			return err
		}
		os.Exit(0)
	}
	return nil
}
