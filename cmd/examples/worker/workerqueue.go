package main

import (
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/rabbitmq"
)

func producer(conn *rabbitmq.Connection) {
	wq := rabbitmq.WorkerQueue{}
	err := wq.Connect(conn,
		rabbitmq.WorkerQueueConfig{
			Name:          "example",
			Timeout:       5,
			PrefetchCount: 1,
		},
	)
	if err != nil {
		panic(err)
	}
	defer wq.Close()

	for i := 0; i < 20; i++ {
		msg := fmt.Sprintf("Message with id: %d", i)
		_, err = wq.Write([]byte(msg))
		if err != nil {
			panic(err)
		}
	}
}

func consumer(conn *rabbitmq.Connection, id string) {
	wq := rabbitmq.WorkerQueue{}
	err := wq.Connect(conn,
		rabbitmq.WorkerQueueConfig{
			Name:          "example",
			Timeout:       5,
			PrefetchCount: 1,
		},
	)
	if err != nil {
		panic(err)
	}
	defer wq.Close()

	for msg := range wq.Consumer {
		fmt.Printf("Consumer %s - Received %s\n", id, msg.Body)
		msg.Ack(false)
	}
}

func main() {
	conn := rabbitmq.Connection{}
	err := conn.Connect("localhost", "5672", "user", "password")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	var forever chan struct{}

	go producer(&conn)
	go consumer(&conn, "1")
	go consumer(&conn, "2")

	<-forever
}
