package main

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/server/src"
)

func main() {
	serverConfig := &src.ServerConfig{
		ServicePort: 7070,
	}
	server, deleteServer := src.NewServer(serverConfig)
	defer deleteServer()

	if err := server.Run(); err != nil {
		fmt.Println("Server running error")
	}
}
