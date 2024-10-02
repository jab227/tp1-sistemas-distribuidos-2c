package comms

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/env"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

type InputType int

// TODO - Add None
const (
	InputWorker InputType = iota
	FanoutSubscriber
	DirectSubscriber
)

type OutputType int

// TODO - Add None
const (
	OutputWorker OutputType = iota
	FanoutPublisher
	DirectPublisher
)

type IOManager struct {
	InputType  InputType
	OutputType OutputType
	Connection *rabbitmq.Connection

	InputWorker      *rabbitmq.WorkerQueue
	FanoutSubscriber *rabbitmq.FanoutSubscriber
	DirectSubscriber *rabbitmq.DirectSubscriber

	OutputWorker    *rabbitmq.WorkerQueue
	FanoutPublisher *rabbitmq.FanoutPublisher
	DirectPublisher *rabbitmq.DirectPublisher
}

func (m *IOManager) connectInput(conn *rabbitmq.Connection, input InputType) error {
	switch input {
	case InputWorker:
		m.InputWorker = &rabbitmq.WorkerQueue{}
		config, err := env.GetOutputWorkerQueueConfig()
		if err != nil {
			return err
		}
		err = m.InputWorker.Connect(conn, *config)
		if err != nil {
			return err
		}

	case FanoutSubscriber:
		m.FanoutSubscriber = &rabbitmq.FanoutSubscriber{}
		config, err := env.GetFanoutSubscriberConfig()
		if err != nil {
			return err
		}
		err = m.FanoutSubscriber.Connect(conn, *config)
		if err != nil {
			return err
		}
	case DirectSubscriber:
		m.DirectSubscriber = &rabbitmq.DirectSubscriber{}
		config, err := env.GetDirectSubscriberConfig()
		if err != nil {
			return err
		}
		err = m.DirectSubscriber.Connect(conn, *config)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *IOManager) connectOutput(conn *rabbitmq.Connection, output OutputType) error {
	switch output {
	case OutputWorker:
		m.OutputWorker = &rabbitmq.WorkerQueue{}
		config, err := env.GetOutputWorkerQueueConfig()
		if err != nil {
			return err
		}
		err = m.OutputWorker.Connect(conn, *config)
		if err != nil {
			return err
		}
	case FanoutPublisher:
		m.FanoutPublisher = &rabbitmq.FanoutPublisher{}
		config, err := env.GetFanoutPublisherConfig()
		if err != nil {
			return err
		}
		err = m.FanoutPublisher.Connect(conn, *config)
		if err != nil {
			return err
		}
	case DirectPublisher:
		m.DirectPublisher = &rabbitmq.DirectPublisher{}
		config, err := env.GetDirectPublisherConfig()
		if err != nil {
			return err
		}
		err = m.DirectPublisher.Connect(conn, *config)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *IOManager) Connect(input InputType, output OutputType) error {
	conn, err := env.GetConnection()
	if err != nil {
		return err
	}

	err = m.connectInput(conn, input)
	if err != nil {
		return err
	}

	err = m.connectOutput(conn, output)
	if err != nil {
		return err
	}

	m.InputType = input
	m.OutputType = output
	m.Connection = conn

	return nil
}

func (m *IOManager) Read() amqp091.Delivery {
	switch m.InputType {
	case InputWorker:
		return m.InputWorker.Read()
	case FanoutSubscriber:
		return m.FanoutSubscriber.Read()
	case DirectSubscriber:
		return m.DirectSubscriber.Read()
	default:
		panic("this should never happen")
	}
}

func (m *IOManager) Write(msg []byte) error {
	switch m.OutputType {
	case OutputWorker:
		_, err := m.OutputWorker.Write(msg)
		return err
	case FanoutPublisher:
		return m.FanoutPublisher.Publish(msg)
	default:
		panic("this should never happen")
	}
}

func (m *IOManager) Close() {
	switch m.InputType {
	case InputWorker:
		m.InputWorker.Close()
	case FanoutSubscriber:
		m.FanoutSubscriber.Close()
	case DirectSubscriber:
		m.DirectSubscriber.Close()
	}

	switch m.OutputType {
	case OutputWorker:
		m.OutputWorker.Close()
	case FanoutPublisher:
		m.FanoutPublisher.Close()
	case DirectPublisher:
		m.DirectPublisher.Close()
	}

	m.Connection.Close()
}
