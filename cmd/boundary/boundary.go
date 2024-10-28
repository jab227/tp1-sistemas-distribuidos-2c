package main

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/boundary"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/logging"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"log/slog"
)

func main() {
	err := logging.InitLoggerWithEnv()
	if err != nil {
		slog.Error("error creating logger", "error", err.Error())
		return
	}

	config, err := boundary.GetBoundaryConfigFromEnv()
	if err != nil {
		slog.Error("error getting bound runtime config", "error", err.Error())
		return
	}

	ioManager := client.IOManager{}
	if err := ioManager.Connect(client.InputWorker, client.OutputWorker); err != nil {
		slog.Error("error connecting to io manager", "error", err.Error())
		return
	}

	boundaryStruct := boundary.NewBoundary(config, &ioManager)
	if err := boundaryStruct.Run(); err != nil {
		slog.Error("error running boundary", "error", err.Error())
		return
	}
}
