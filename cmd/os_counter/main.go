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

	osCounter, err := controllers.NewOSCounter()
	defer osCounter.Close()
	if err != nil {
		slog.Error("error creating os counter", "error", err)
		return
	}

	slog.Info("os counter started")
	go func() {
		err = osCounter.Run(ctx)
		if err != nil {
			slog.Error("error running os counter", "error", err.Error())
			return
		}
	}()

	utils.BlockUntilSignal(signal, osCounter.Done(), cancel)
}
