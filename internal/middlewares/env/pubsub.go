package env

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/rabbitmq"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

const DirectPublisherExchange = "DIRECT_PUBLISHER_EXCHANGE"
const DirectPublisherTimeout = "DIRECT_PUBLISHER_TIMEOUT"

func GetDirectPublisherConfig() (*rabbitmq.DirectPublisherConfig, error) {
	exchange, err := utils.GetFromEnv(DirectPublisherExchange)
	if err != nil {
		return nil, err
	}

	timeout, err := utils.GetFromEnvUint(DirectPublisherTimeout)
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

const DirectSubscriberExchange = "DIRECT_SUBSCRIBER_EXCHANGES"
const DirectSubscriberQueue = "DIRECT_SUBSCRIBER_QUEUE"
const DirectSubscriberKeys = "DIRECT_SUBSCRIBER_KEYS"
const DirectSubscriberPrefetchCount = "DIRECT_SUBSCRIBER_PREFETCH_COUNT"

func GetDirectSubscriberConfig() (*rabbitmq.DirectSubscriberConfig, error) {
	exchange, err := utils.GetFromEnv(DirectSubscriberExchange)
	if err != nil {
		return nil, err
	}

	queue, err := utils.GetFromEnv(DirectSubscriberQueue)
	if err != nil {
		return nil, err
	}

	keys, err := utils.GetFromEnv(DirectSubscriberKeys)
	if err != nil {
		return nil, err
	}

	finalPrefetchCount := -1
	prefetchCount, err := utils.GetFromEnvInt(DirectSubscriberPrefetchCount)
	if err != nil {
		slog.Info("No prefetch count specified")
	} else {
		finalPrefetchCount = int(*prefetchCount)
	}

	exchangeList := strings.Split(*exchange, ",")
	keysList := strings.Split(*keys, ",")
	return &rabbitmq.DirectSubscriberConfig{
		Exchange:      exchangeList,
		Queue:         *queue,
		Keys:          keysList,
		PrefetchCount: finalPrefetchCount,
	}, nil
}
