package main

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/client/src"
)

func main() {
	clientConfig := &src.ClientConfig{
		ServerName: "localhost",
		ServerPort: 7070,
	}
	client, deleteClient := src.NewClient(clientConfig)
	defer deleteClient()

	if err := client.Run(); err != nil {
		fmt.Println("Client running error")
	}
}
