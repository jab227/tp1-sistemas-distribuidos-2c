package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"

	internalClient "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/logging"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

func loadConfig(path string) (*internalClient.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}

	bytesValues, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config internalClient.Config
	if err := json.Unmarshal(bytesValues, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return &config, nil
}

func main() {
	clientConfig, err := loadConfig(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := logging.InitLoggerWithString(clientConfig.LoggerLevel); err != nil {
		fmt.Println(err)
		return
	}

	client, err := internalClient.NewClient(clientConfig)
	defer client.Close()
	if err != nil {
		slog.Error(err.Error())
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	signal := utils.MakeSignalHandler()
	done := make(chan struct{}, 1)
	go func() {
		clientDone := make(chan struct{}, 1)
		go func() {
			if err := client.Run(); err != nil {
				slog.Error("stopping client")
			}
			clientDone <- struct{}{}
		}()
		select {
		case <-ctx.Done():
			slog.Error("context error", "error", ctx.Err())
			done <- struct{}{}			
			return
		case <-clientDone:
			done <- struct{}{}
			return
		}
	}()
	utils.BlockUntilSignal(signal, done, cancel)
}
