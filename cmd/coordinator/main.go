package main

import (
	"context"
	coordinator2 "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/coordinator"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/healthcheck"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/logging"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"log/slog"
)

func main() {
	if err := logging.InitLoggerWithEnv(); err != nil {
		slog.Error("error creating logger", "error", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	signal := utils.MakeSignalHandler()

	outputType, err := coordinator2.GetOutputType()
	if err != nil {
		slog.Error("error getting output type", "error", err.Error())
		return
	}

	expectedGames, err := coordinator2.GetExpectedGames()
	if err != nil {
		slog.Error("error getting expected games", "error", err.Error())
		return
	}

	expectedReviews, err := coordinator2.GetExpectedRevisions()
	if err != nil {
		slog.Error("error getting expected revisions", "error", err.Error())
		return
	}

	// Set up the healthcheck service
	config := healthcheck.NewHealthServiceConfigFromEnv()
	service := healthcheck.NewHealthService(config)

	go service.Run(ctx)

	coordinator, err := coordinator2.NewEndCoordinator(outputType, expectedGames, expectedReviews)
	defer coordinator.Close()
	if err != nil {
		slog.Error("error creating coordinator", "error", err.Error())
		return
	}

	slog.Info("starting coordinator")
	go func() {
		if err := coordinator.Run(ctx); err != nil {
			slog.Error("error running coordinator", "error", err.Error())
			return
		}
	}()

	utils.BlockUntilSignal(signal, coordinator.GetDone(), cancel)
}
