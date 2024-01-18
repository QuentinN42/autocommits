package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/civil"
	"github.com/QuentinN42/autocommits/pkg/logger"
	"github.com/QuentinN42/autocommits/pkg/types"
)

type ResponseModel = map[string]int

func getModel(ctx context.Context, url string) (*ResponseModel, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	res := ResponseModel{}
	err = json.Unmarshal(resBody, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func parseModel(ctx context.Context, mod ResponseModel) []types.Todo {
	result := []types.Todo{}
	for k, v := range mod {
		date, err := civil.ParseDate(k)
		if err != nil {
			logger.Warning(ctx, "Could not parse date %s", k)
			continue
		}
		for i := 0; i < v; i++ {
			result = append(result, types.Todo{
				ID:   fmt.Sprintf("%s-%d", k, i+1),
				Date: time.Date(date.Year, time.Month(date.Month), date.Day, 0, 0, 0, 0, time.UTC),
			})
		}
	}
	return result
}

type Gitlab struct {
	username string
}

func New(ctx context.Context) (*Gitlab, error) {
	username := os.Getenv("GL_USER")
	if username == "" {
		return nil, fmt.Errorf("GL_USER is not set")
	}
	return &Gitlab{
		username: username,
	}, nil
}

// Fetches commits from GitLab
//
// https://gitlab.com/users/[username]/calendar.json
func (g *Gitlab) GetCommits(ctx context.Context) ([]types.Todo, error) {
	url := "https://gitlab.com/users/" + g.username + "/calendar.json"

	model, err := getModel(ctx, url)
	if err != nil {
		return nil, err
	}

	return parseModel(ctx, *model), nil
}
