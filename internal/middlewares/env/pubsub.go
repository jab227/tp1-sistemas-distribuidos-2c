package env

import (
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/rabbitmq"
	"strings"
)

const DirectPublisherExchange = "DIRECT_PUBLISHER_EXCHANGE"
const DirectPublisherTimeout = "DIRECT_PUBLISHER_TIMEOUT"

func GetDirectPublisherConfig() (*rabbitmq.DirectPublisherConfig, error) {
	exchange, err := GetFromEnv(DirectPublisherExchange)
	if err != nil {
		return nil, err
	}

	timeout, err := GetFromEnvUint(DirectPublisherTimeout)
	if err != nil {
		return nil, err
	}
	if *timeout <= 0 {
		return nil, fmt.Errorf(
			"environment variable %s must be a positive integer: %d",
			DirectPublisherTimeout,
			timeout,
		)
	}

	return &rabbitmq.DirectPublisherConfig{
		Exchange: *exchange,
		Timeout:  uint8(*timeout),
	}, nil
}

const DirectSubscriberExchange = "DIRECT_SUBSCRIBER_EXCHANGE"
const DirectSubscriberQueue = "DIRECT_SUBSCRIBER_QUEUE"
const DirectSubscriberKeys = "DIRECT_SUBSCRIBER_KEYS"

func GetDirectSubscriberConfig() (*rabbitmq.DirectSubscriberConfig, error) {
	exchange, err := GetFromEnv(DirectSubscriberExchange)
	if err != nil {
		return nil, err
	}

	queue, err := GetFromEnv(DirectSubscriberQueue)
	if err != nil {
		return nil, err
	}

	keys, err := GetFromEnv(DirectSubscriberKeys)
	if err != nil {
		return nil, err
	}

	keysList := strings.Split(*keys, ",")
	return &rabbitmq.DirectSubscriberConfig{
		Exchange: *exchange,
		Queue:    *queue,
		Keys:     keysList,
	}, nil
}
