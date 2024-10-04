package env

import (
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/rabbitmq"
)

const FanoutPublisherExchange = "FANOUT_PUBLISHER_EXCHANGE"
const FanoutPublisherTimeout = "FANOUT_PUBLISHER_TIMEOUT"

func GetFanoutPublisherConfig() (*rabbitmq.FanoutPublisherConfig, error) {
	exchange, err := GetFromEnv(FanoutPublisherExchange)
	if err != nil {
		return nil, err
	}

	timeout, err := GetFromEnvUint(FanoutPublisherTimeout)
	if err != nil {
		return nil, err
	}
	if *timeout <= 0 {
		return nil, fmt.Errorf(
			"environment variable %s must be a positive integer: %d",
			FanoutPublisherTimeout,
			timeout,
		)
	}

	return &rabbitmq.FanoutPublisherConfig{
		Exchange: *exchange,
		Timeout:  uint8(*timeout),
	}, nil
}

const FanoutSubscriberExchange = "FANOUT_SUBSCRIBER_EXCHANGE"
const FanoutSubscriberQueueName = "FANOUT_SUBSCRIBER_QUEUE_NAME"

func GetFanoutSubscriberConfig() (*rabbitmq.FanoutSubscriberConfig, error) {
	exchange, err := GetFromEnv(FanoutSubscriberExchange)
	if err != nil {
		return nil, err
	}

	name, err := GetFromEnv(FanoutSubscriberQueueName)
	if err != nil {
		return nil, err
	}

	return &rabbitmq.FanoutSubscriberConfig{
		Exchange: *exchange,
		Queue:    *name,
	}, nil
}
