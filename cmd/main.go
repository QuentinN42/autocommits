package main

import "github.com/QuentinN42/autocommits/pkg/svc"

func main() {
	svc := svc.New()

	svc.Run()
}
