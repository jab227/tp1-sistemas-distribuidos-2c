package main

import (
	"context"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/controllers"
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

	percentile, err := controllers.NewPercentile()
	defer percentile.Close()
	if err != nil {
		slog.Error("error creating percentile", "error", err)
		return
	}

	slog.Info("review percentile")
	go func() {
		err = percentile.Run(ctx)
		if err != nil {
			slog.Error("error running percentile", "error", err.Error())
			return
		}
	}()

	utils.BlockUntilSignal(signal, percentile.Done(), cancel)
}
