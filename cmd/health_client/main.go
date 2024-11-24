package main

import "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/healthcheck"

func main() {
	service := healthcheck.HealthService{
		Port: 1516,
	}

	if err := service.Run(); err != nil {
		panic(err)
	}
}
