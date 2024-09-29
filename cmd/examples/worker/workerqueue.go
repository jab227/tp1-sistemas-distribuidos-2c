package main

import (
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

func producer(conn *amqp091.Connection) {
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

func consumer(conn *amqp091.Connection, id string) {
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
	conn, err := amqp091.Dial("amqp://user:password@localhost:5672/")
	if err != nil {
		panic(err)
	}

	var forever chan struct{}

	go producer(conn)
	go consumer(conn, "1")
	go consumer(conn, "2")

	<-forever
}
