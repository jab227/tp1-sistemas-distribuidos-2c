package main

import (
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/rabbitmq"
	"time"
)

func producer(conn *rabbitmq.Connection) {
	publisher := rabbitmq.FanoutPublisher{}
	err := publisher.Connect(
		conn,
		rabbitmq.FanoutPublisherConfig{
			Exchange: "exchange_example",
			Timeout:  5,
		},
	)
	if err != nil {
		panic(err)
	}
	defer publisher.Close()

	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("Message with Id: %d", i)
		err = publisher.Publish([]byte(msg))
		fmt.Printf("Sended %s\n", msg)
		if err != nil {
			println(err.Error())
			panic(err)
		}
	}
}

func consumer(conn *rabbitmq.Connection, id string) {
	subscriber := rabbitmq.FanoutSubscriber{}
	err := subscriber.Connect(
		conn,
		rabbitmq.FanoutSubscriberConfig{
			Exchange: "exchange_example",
			Queue:    fmt.Sprintf("queue_example_%s", id),
		},
	)
	if err != nil {
		panic(err)
	}
	defer subscriber.Close()

	for msg := range subscriber.Consumer {
		fmt.Printf("%s - Received message: %s\n", id, msg.Body)
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

	go consumer(&conn, "1")
	go consumer(&conn, "2")

	// Add a small delay to ensure consumers are ready
	time.Sleep(2 * time.Second)
	go producer(&conn)

	<-forever
}
