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

	joiner, err := controllers.NewJoiner()
	if err != nil {
		slog.Error("error creating joiner", "error", err)
		return
	}
	defer joiner.Destroy()

	slog.Info("joiner started")
	go func() {
		err = joiner.Run(ctx)
		if err != nil {
			slog.Error("error running joiner", "error", err.Error())
			return
		}
	}()

	utils.BlockUntilSignal(signal, joiner.Done(), cancel)
}
