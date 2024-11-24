package main

import "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/healthcheck"

func main() {
	config := healthcheck.ControllerConfig{
		Port:                1516,
		NodesPort:           1516,
		HealthCheckInterval: 5,
		MaxRetries:          3,
		MaxTimeout:          4,
		ListOfNodes:         []string{"node1", "node2"},
	}

	controller := healthcheck.NewHealthController(config)

	if err := controller.Run(); err != nil {
		panic(err)
	}
}
