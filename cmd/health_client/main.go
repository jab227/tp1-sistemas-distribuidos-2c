package main

import (
	"context"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/healthcheck"
)

func main() {
	service := healthcheck.HealthService{
		Port:    1516,
		Timeout: 2,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := service.Run(ctx); err != nil {
		panic(err)
	}
}
