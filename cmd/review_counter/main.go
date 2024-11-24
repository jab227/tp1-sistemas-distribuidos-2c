package main

import (
	"context"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/controllers"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/healthcheck"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/logging"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"log/slog"
)

func main() {
	err := logging.InitLoggerWithEnv()
	if err != nil {
		slog.Error("error creating logger", "error", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	signal := utils.MakeSignalHandler()

	// Set up the healthcheck service
	config := healthcheck.NewHealthServiceConfigFromEnv()
	service := healthcheck.NewHealthService(config)

	go service.Run(ctx)

	reviewCounter, err := controllers.NewReviewCounter()
	defer reviewCounter.Close()
	if err != nil {
		slog.Error("error creating review", "error", err)
		return
	}

	slog.Info("review counter started")
	go func() {
		err = reviewCounter.Run(ctx)
		if err != nil {
			slog.Error("error running review counter", "error", err.Error())
			return
		}
	}()

	utils.BlockUntilSignal(signal, reviewCounter.Done(), cancel)
}
