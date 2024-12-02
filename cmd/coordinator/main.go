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

	expectedNodes, err := coordinator2.GetExpectedNodes()
	if err != nil {
		slog.Error("error getting expected nodes", "error", err.Error())
		return
	}

	transactionLogFile, err := coordinator2.GetTransactionLogFile()
	if err != nil {
		slog.Error("error getting transaction log file", "error", err.Error())
		return
	}

	// Set up the healthcheck service
	config := healthcheck.NewHealthServiceConfigFromEnv()
	service := healthcheck.NewHealthService(config)

	go service.Run(ctx)

	coordinator, err := coordinator2.NewEndCoordinatorController(outputType, expectedNodes)
	defer coordinator.Close()
	if err != nil {
		slog.Error("error creating coordinator", "error", err.Error())
		return
	}

	slog.Info("starting coordinator")
	go func() {
		if err := coordinator.Run(ctx, transactionLogFile); err != nil {
			slog.Error("error running coordinator", "error", err.Error())
			return
		}
	}()

	utils.BlockUntilSignal(signal, coordinator.Done(), cancel)
}
