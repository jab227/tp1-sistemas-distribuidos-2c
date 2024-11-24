package main

import (
	"context"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/healthcheck"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/logging"
)

func main() {
	err := logging.InitLoggerWithEnv()
	if err != nil {
		slog.Error("error creating logger", "error", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	signal := utils.MakeSignalHandler()

	// TODO - Add list of nodes as config
	config := healthcheck.NewHealthConfigFromEnv([]string{"node1", "node2"})
	healthController := healthcheck.NewHealthController(*config)

	go func() {
		if err = healthController.Run(ctx); err != nil {
			slog.Error("error running healthcheck", "error", err.Error())
			return
		}
	}()

	utils.BlockUntilSignal(signal, healthController.Done(), cancel)
}
