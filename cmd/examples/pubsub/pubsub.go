package main

import (
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"time"
)

func producer(conn *amqp091.Connection) {
	publisher := rabbitmq.Publisher{}
	err := publisher.Connect(
		conn,
		rabbitmq.PublisherConfig{
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

func consumer(conn *amqp091.Connection, id string) {
	subscriber := rabbitmq.Subscriber{}
	err := subscriber.Connect(
		conn,
		rabbitmq.SubscriberConfig{
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
	conn, err := amqp091.Dial("amqp://user:password@localhost:5672/")
	if err != nil {
		panic(err)
	}

	var forever chan struct{}

	go consumer(conn, "1")
	go consumer(conn, "2")

	// Add a small delay to ensure consumers are ready
	time.Sleep(2 * time.Second)
	go producer(conn)

	<-forever
}
