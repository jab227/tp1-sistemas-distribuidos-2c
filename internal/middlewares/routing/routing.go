package routing

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
)

type Router struct {
	keys []string
	m    *client.IOManager
}

func NewRouter(input client.InputType, keys []string) (Router, error) {
	manager := &client.IOManager{}
	if err := manager.Connect(input, client.DirectPublisher); err != nil {
		return Router{}, fmt.Errorf("couldn't create router: %w", err)
	}

	return Router{m: manager, keys: keys}, nil
}

