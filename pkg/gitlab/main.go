package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
				ID:   fmt.Sprintf("%s-%d", k, i),
				Date: date,
			})
		}
	}
	return result
}

// Fetches commits from GitLab
//
// https://gitlab.com/users/[username]/calendar.json
func GetCommits(ctx context.Context, username string) ([]types.Todo, error) {
	url := "https://gitlab.com/users/" + username + "/calendar.json"

	model, err := getModel(ctx, url)
	if err != nil {
		return nil, err
	}

	return parseModel(ctx, *model), nil
}
