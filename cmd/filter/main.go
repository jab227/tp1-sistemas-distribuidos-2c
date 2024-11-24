package main

import (
	"context"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/healthcheck"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/controllers"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/logging"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
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

	filterName, err := utils.GetFromEnv("FILTER_NAME")
	if err != nil {
		slog.Error("couldn't read filter name", "error", err)
		return
	}
	filter, err := controllers.NewFilter(*filterName)
	if err != nil {
		slog.Error("error creating filter", "error", err)
		return
	}
	defer filter.Close()

	slog.Info("filter started", "filter", *filterName)
	go func() {
		err = filter.Run(ctx)
		if err != nil {
			slog.Error("error running filter", "filter", *filterName, "error", err.Error())
			return
		}
	}()

	utils.BlockUntilSignal(signal, filter.Done(), cancel)
}
