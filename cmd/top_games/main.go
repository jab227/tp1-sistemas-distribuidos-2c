package main

import (
	"context"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/controllers"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/logging"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/env"
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
	n, err := env.GetFromEnvUint(nKeyName)
	if err != nil {
		slog.Error("error parsing N_VALUE env var", "error", err)
		return
	}

	topGames, err := controllers.NewTopGames(*n)
	if err != nil {
		slog.Error("error creating top games", "error", err)
		return
	}

	slog.Info("top games started")
	go func() {
		err = topGames.Run(ctx)
		if err != nil {
			slog.Error("error running top games", "error", err.Error())
			return
		}
	}()

	utils.BlockUntilSignal(signal, topGames.Done(), cancel)
}
