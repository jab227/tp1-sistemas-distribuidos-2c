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

	if err := server.Listen(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if err := server.Accept(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if err := server.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
}
