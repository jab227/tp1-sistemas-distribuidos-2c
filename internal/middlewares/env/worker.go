package env

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/rabbitmq"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

const InputWorkerQueueName = "INPUT_WORKER_QUEUE"
const InputWorkerQueueTimeout = "INPUT_WORKER_QUEUE_TIMEOUT"
const InputWorkerQueueCount = "INPUT_WORKER_QUEUE_COUNT"

func GetInputWorkerQueueConfig() (*rabbitmq.WorkerQueueConfig, error) {
	name, err := utils.GetFromEnv(InputWorkerQueueName)
	if err != nil {
		return nil, err
	}

	timeout, err := utils.GetFromEnvUint(InputWorkerQueueTimeout)
	if err != nil {
		return nil, err
	}

	if *timeout <= 0 {
		return nil, fmt.Errorf(
			"environment variable %s must be a positive integer: %d",
			InputWorkerQueueTimeout,
			timeout,
		)
	}

	count, err := utils.GetFromEnvInt(InputWorkerQueueCount)
	if err != nil {
		return nil, err
	}
	if *count <= 0 {
		return nil, fmt.Errorf(
			"environment variable %s must be a positive integer: %d",
			InputWorkerQueueCount,
			count,
		)
	}

	return &rabbitmq.WorkerQueueConfig{
		Name:          *name,
		Timeout:       uint8(*timeout),
		PrefetchCount: int(*count),
	}, nil
}

const OutputWorkerQueueName = "OUTPUT_WORKER_QUEUE"
const OutputWorkerQueueTimeout = "OUTPUT_WORKER_QUEUE_TIMEOUT"
const OutputWorkerQueueCount = "OUTPUT_WORKER_QUEUE_COUNT"

func GetOutputWorkerQueueConfig() (*rabbitmq.WorkerQueueConfig, error) {
	name, err := utils.GetFromEnv(OutputWorkerQueueName)
	if err != nil {
		return nil, err
	}

	timeout, err := utils.GetFromEnvUint(OutputWorkerQueueTimeout)
	if err != nil {
		return nil, err
	}
	if *timeout <= 0 {
		return nil, fmt.Errorf(
			"environment variable %s must be a positive integer: %d",
			OutputWorkerQueueTimeout,
			timeout,
		)
	}

	count, err := utils.GetFromEnvInt(OutputWorkerQueueCount)
	if err != nil {
		return nil, err
	}
	if *count <= 0 {
		return nil, fmt.Errorf(
			"environment variable %s must be a positive integer: %d",
			OutputWorkerQueueCount,
			count,
		)
	}

	return &rabbitmq.WorkerQueueConfig{
		Name:          *name,
		Timeout:       uint8(*timeout),
		PrefetchCount: int(*count),
	}, nil
}
