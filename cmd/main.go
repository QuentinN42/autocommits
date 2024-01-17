package main

import (
	"context"

	"github.com/QuentinN42/autocommits/pkg/svc"
)

func main() {
	ctx := context.Background()
	svc := svc.New(ctx)
	svc.Run(ctx)
}
