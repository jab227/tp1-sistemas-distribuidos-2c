package client

import (
	"fmt"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/env"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/rabbitmq"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type InputType int

const (
	NoneInput InputType = iota
	InputWorker
	FanoutSubscriber
	DirectSubscriber
)

type OutputType int

const (
	NoneOutput OutputType = iota
	OutputWorker
	FanoutPublisher
	DirectPublisher
	Router
)

type IOManager struct {
	Conn rabbitmq.Connection

	InputType InputType
	Input     rabbitmq.InputHandler

	OutputType OutputType
	Output     rabbitmq.OutputHandler
}

func (m *IOManager) connectInput(conn *rabbitmq.Connection, input InputType) error {
	switch input {
	case NoneInput:
		return nil
	case InputWorker:
		config, err := env.GetInputWorkerQueueConfig()
		if err != nil {
			return err
		}
		m.Input = rabbitmq.NewWorkerQueue(*config)
	case FanoutSubscriber:
		config, err := env.GetFanoutSubscriberConfig()
		if err != nil {
			return err
		}
		m.Input = rabbitmq.NewFanoutSubscriber(*config)
	case DirectSubscriber:
		config, err := env.GetDirectSubscriberConfig()
		if err != nil {
			return err
		}
		m.Input = rabbitmq.NewDirectSubscriber(*config)
	}

	err := m.Input.Connect(conn)
	if err != nil {
		return err
	}
	return nil
}

func (m *IOManager) connectOutput(conn *rabbitmq.Connection, output OutputType) error {
	switch output {
	case NoneOutput:
		return nil
	case OutputWorker:
		config, err := env.GetOutputWorkerQueueConfig()
		if err != nil {
			return err
		}
		m.Output = rabbitmq.NewWorkerQueue(*config)
	case FanoutPublisher:
		config, err := env.GetFanoutPublisherConfig()
		if err != nil {
			return err
		}
		m.Output = rabbitmq.NewFanoutPublisher(*config)
	case DirectPublisher:
		config, err := env.GetDirectPublisherConfig()
		if err != nil {
			return err
		}
		m.Output = rabbitmq.NewDirectPublisher(*config)
	case Router:
		config, err := env.GetDirectPublisherConfig()
		routerType := env.GetRouterTypeFromEnv()
		tags, err := env.GetRouterTags()
		if err != nil {
			return fmt.Errorf("couldn't get tags from env: %w", err)
		}
		slog.Debug("tag values", "values", tags)
		selector := makeSelector(routerType, len(tags))
		router := rabbitmq.NewRouter(*config, tags, selector)
		m.Output = &router
	}

	err := m.Output.Connect(conn)
	if err != nil {
		return err
	}
	return nil
}

func makeSelector(routerType string, tagCount int) rabbitmq.RouteSelector {
	switch routerType {
	case env.RouterRoundRobin:
		slog.Debug("use round robin router")
		r := rabbitmq.NewRoundRobinRouter(tagCount)
		return &r
	case env.RouterGameReview:
		slog.Debug("use game-review router")
		return rabbitmq.GameReviewRouter{}
	case env.RouterID:
		slog.Debug("use id router")
		return rabbitmq.NewIDRouter(tagCount)
	default:
		utils.Assertf(false, "unknown router type %s", routerType)
		return nil
	}
}

func (m *IOManager) Connect(input InputType, output OutputType) error {
	conn, err := env.GetConnection()
	if err != nil {
		return err
	}
	m.Conn = *conn

	if err = m.connectInput(conn, input); err != nil {
		return err
	}
	m.InputType = input

	if err = m.connectOutput(conn, output); err != nil {
		return err
	}
	m.OutputType = output

	return nil
}

func (m *IOManager) Write(msg []byte, tag string) error {
	if m.OutputType == NoneOutput {
		panic("no output was configured")
	}

	return m.Output.Write(msg, tag)
}

func (m *IOManager) Close() {
	if m.InputType != NoneInput {
		m.Input.Close()
	}

	if m.OutputType != NoneOutput {
		m.Output.Close()
	}

	m.Conn.Close()
}
