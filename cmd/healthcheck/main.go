package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"

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

	serviceConfig := healthcheck.NewHealthServiceConfigFromEnv()
	service := healthcheck.NewHealthService(serviceConfig)

	go service.Run(ctx)

	signal := utils.MakeSignalHandler()

	// TODO - Add list of excluded as config
	discoveryConfig, err := healthcheck.NewDockerDiscoveryConfigFromEnv()
	if err != nil {
		slog.Error("error creating discovery config", "error", err.Error())
		return
	}

	nodesList, err := healthcheck.GetDockerNodes(discoveryConfig.Excluded, discoveryConfig.Network)
	fmt.Printf("%v", nodesList)
	if err != nil {
		slog.Error("error getting docker nodes list", "error", err.Error())
	}
	slog.Info("got docker nodes list", "nodes", nodesList)
	config, err := healthcheck.NewHealthConfigFromEnv(nodesList)
	if err != nil {
		slog.Error("couldn't read config from env", "error", err)
		os.Exit(1)
	}
	healthController := healthcheck.NewHealthController(config)

	go func() {
		if err = healthController.Run(ctx); err != nil {
			slog.Error("error running healthcheck", "error", err.Error())
			return
		}
	}()

	utils.BlockUntilSignal(signal, healthController.Done(), cancel)
}
