package main

import (
	"context"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/logging"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/cmd/server/src"
)

func main() {
	err := logging.InitLoggerWithEnv()
	if err != nil {
		slog.Error("error creating logger", "error", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	signal := utils.MakeSignalHandler()

	serverConfig, err := src.GetServerConfigFromEnv()
	if err != nil {
		slog.Error("error getting server config", "error", err.Error())
		return
	}

	inputManager := client.IOManager{}
	if err := inputManager.Connect(client.InputWorker, client.NoneOutput); err != nil {
		slog.Error("error connecting to Input Manager", "error", err.Error())
		return
	}
	defer inputManager.Close()
	outputManager := client.IOManager{}
	if err = outputManager.Connect(client.NoneInput, client.OutputWorker); err != nil {
		slog.Error("error connecting to Output Manager", "error", err.Error())
		return
	}
	defer outputManager.Close()

	server, deleteServer := src.NewServer(serverConfig, &inputManager, &outputManager)
	defer deleteServer()

	slog.Info("starting server")
	go func() {
		err := server.Run(ctx)
		if err != nil {
			slog.Error("error starting server", "error", err.Error())
			return
		}
	}()

	utils.BlockUntilSignal(signal, server.GetDone(), cancel)
}
