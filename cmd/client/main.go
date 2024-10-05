package main

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/cmd/client/src"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication/message"
)

func main() {
	clientConfig := &src.ClientConfig{
		ServerName:    "localhost",
		ServerPort:    7070,
		TaskQueueSize: 10,
		ReviewsBatch: &src.BatchFileConfig{
			DataType:       message.Reviews,
			Path:           "/home/daniel/Documents/facultad/2024-2C/75.74/tps/tp1-sistemas-distribuidos-2c/datasets/reviews.csv",
			NlinesFromDisk: 100,
			BatchSize:      10,
			MaxBytes:       4 * 1024,
		},
		GamesBatch: &src.BatchFileConfig{
			DataType:       message.Games,
			Path:           "/home/daniel/Documents/facultad/2024-2C/75.74/tps/tp1-sistemas-distribuidos-2c/datasets/games.csv",
			NlinesFromDisk: 100,
			BatchSize:      1,
			MaxBytes:       4 * 1024,
		},
	}

	client, deleteClient := src.NewClient(clientConfig)
	defer deleteClient()

	if err := client.Connect(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if err := client.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
}
