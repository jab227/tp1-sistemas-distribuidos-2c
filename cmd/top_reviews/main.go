package main

import (
	"context"
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

	const nKeyName = "N_VALUE"
	n, err := utils.GetFromEnvInt(nKeyName)
	if err != nil {
		slog.Error("error parsing N_VALUE env var", "error", err)
		return
	}

	topReviews, err := controllers.NewTopReviews(int(*n))
	if err != nil {
		slog.Error("error creating top reviews", "error", err)
		return
	}
	defer topReviews.Close()

	slog.Info("top reviews started")
	go func() {
		err = topReviews.Run(ctx)
		if err != nil {
			slog.Error("error running top reviews", "error", err.Error())
			return
		}
	}()

	utils.BlockUntilSignal(signal, topReviews.Done(), cancel)
}
