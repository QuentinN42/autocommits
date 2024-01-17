package main

import (
	"context"
	"os"

	"github.com/QuentinN42/autocommits/pkg/logger"
	"github.com/QuentinN42/autocommits/pkg/svc"
)

func main() {
	ctx := context.Background()
	svc, err := svc.New(ctx)
	if err != nil {
		logger.Error(ctx, "Could not initialize service: %v", err)
		os.Exit(1)
	}
	err = svc.Run(ctx)
	if err != nil {
		logger.Error(ctx, "Could not run service: %v", err)
		os.Exit(1)
	}
}
