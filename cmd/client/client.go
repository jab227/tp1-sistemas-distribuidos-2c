package main

import (
	"encoding/json"
	"fmt"
	internalClient "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/logging"
	"io"
	"log/slog"
	"os"
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

	if err := client.Run(); err != nil {
		panic(err)
	}

}
