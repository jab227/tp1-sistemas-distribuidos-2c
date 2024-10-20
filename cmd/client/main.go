package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/logging"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/cmd/client/src"
)

func loadConfig(path string) (*src.ClientConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}

	bytesValues, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config src.ClientConfig
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

	if err = logging.InitLoggerWithString(clientConfig.LoggerLevel); err != nil {
		fmt.Println(err)
		return
	}

	slog.Debug("Configuration debug",
		"ClientConfig", clientConfig,
		"ReviewsBatch", clientConfig.ReviewsBatch,
		"GamesBatch", clientConfig.GamesBatch,
	)

	games_file := ""
	reviews_file := ""
	output_file := ""
	fmt.Println("Enter the game's file name:")
	fmt.Scanln(&games_file)
	clientConfig.GamesBatch.Path += games_file
	fmt.Println("Enter the review's file name:")
	fmt.Scanln(&reviews_file)
	clientConfig.ReviewsBatch.Path += reviews_file
	fmt.Println("Enter the output's file name:")
	fmt.Scanln(&output_file)
	clientConfig.OutputFile += output_file

	client, deleteClient := src.NewClient(clientConfig)
	defer deleteClient()

	if err := client.Connect(); err != nil {
		slog.Error("error connecting to server:", "error", err)
		return
	}
	if err := client.Execute(); err != nil {
		slog.Error("error executing command:", "error", err)
		return
	}
}
