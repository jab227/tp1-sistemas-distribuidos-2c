package end

import (
	"fmt"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/env"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/rabbitmq"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"github.com/rabbitmq/amqp091-go"
)

type Service struct {
	fanoutPub   *rabbitmq.FanoutPublisher
	fanoutSub   *rabbitmq.FanoutSubscriber
	coordinator *rabbitmq.WorkerQueue
	notify      chan protocol.Message
	NodeId      uint32
}

type ServiceOptions struct {
	Exchange         string
	SubscriberQueue  string
	CoordinatorQueue string
	Timeout          uint32
	NodeId           uint32
}

func GetServiceOptionsFromEnv() (*ServiceOptions, error) {
	exchange, err := utils.GetFromEnv("END_SERVICE_EXCHANGE")
	if err != nil {
		return nil, err
	}

	subQueue, err := utils.GetFromEnv("END_SERVICE_SUBSCRIBER_QUEUE")
	if err != nil {
		return nil, err
	}

	coordinator, err := utils.GetFromEnv("END_SERVICE_COORDINATOR_QUEUE")
	if err != nil {
		return nil, err
	}

	nodeId, err := utils.GetFromEnvUint("END_SERVICE_NODE_ID")
	if err != nil {
		return nil, err
	}
	slog.Info("Initializing with Node ID", "nodeId", uint32(*nodeId))

	timeout, err := utils.GetFromEnvUint("END_SERVICE_TIMEOUT")
	if err != nil {
		return nil, err
	}
	if *timeout <= 0 {
		return nil, fmt.Errorf(
			"environment variable %s must be a positive integer: %d",
			"END_SERVICE_TIMEOUT",
			timeout,
		)
	}

	return &ServiceOptions{
		Exchange:         *exchange,
		SubscriberQueue:  *subQueue,
		CoordinatorQueue: *coordinator,
		Timeout:          uint32(*timeout),
		NodeId:           uint32(*nodeId),
	}, nil
}

func NewService(opts *ServiceOptions) (*Service, error) {
	conn, err := env.GetConnection()
	if err != nil {
		return nil, err
	}

	coordinator := rabbitmq.NewWorkerQueue(rabbitmq.WorkerQueueConfig{
		Name:          opts.CoordinatorQueue,
		Timeout:       opts.Timeout,
		PrefetchCount: 1,
	})
	if err := coordinator.Connect(conn); err != nil {
		return nil, fmt.Errorf("end service: couldn't create coordinator queue: %w", err)
	}

	fanoutPub := rabbitmq.NewFanoutPublisher(rabbitmq.FanoutPublisherConfig{
		Exchange: opts.Exchange,
		Timeout:  opts.Timeout,
	})
	if err := fanoutPub.Connect(conn); err != nil {
		return nil, fmt.Errorf("end service: couldn't create fanout publisher queue: %w", err)
	}

	fanoutSub := rabbitmq.NewFanoutSubscriber(rabbitmq.FanoutSubscriberConfig{
		Exchange: opts.Exchange,
		Queue:    opts.SubscriberQueue,
	})
	if err := fanoutSub.Connect(conn); err != nil {
		return nil, fmt.Errorf("end service: couldn't create fanout subscriber queue: %w", err)
	}

	return &Service{
		fanoutPub:   fanoutPub,
		fanoutSub:   fanoutSub,
		coordinator: coordinator,
		notify:      make(chan protocol.Message),
		NodeId:      opts.NodeId,
	}, nil
}

func (s *Service) Destroy() {
	s.fanoutPub.Close()
	s.fanoutSub.Close()
	s.coordinator.Close()
}

type MessageInfo struct {
	DataType protocol.DataType
	Options  protocol.MessageOptions
}

type Sender string

const (
	SenderNeighbour Sender = "neighbour"
	SenderPrevious  Sender = "previous"
)

type Delivery struct {
	RecvDelivery amqp091.Delivery
	SenderType   Sender
}

func (s *Service) MergeConsumers(consumer <-chan amqp091.Delivery) <-chan Delivery {
	newConsumer := make(chan Delivery, 1)
	neighboursConsumer := s.fanoutSub.GetConsumer()
	go func() {
		for {
			select {
			case delivery := <-neighboursConsumer:
				newConsumer <- Delivery{
					RecvDelivery: delivery,
					SenderType:   SenderNeighbour,
				}
			case delivery := <-consumer:
				newConsumer <- Delivery{
					RecvDelivery: delivery,
					SenderType:   SenderPrevious,
				}
			}
		}
	}()
	return newConsumer
}

func (s *Service) NotifyCoordinator(endMessage protocol.Message) {
	utils.Assert(endMessage.ExpectKind(protocol.End), "must be an END message")
	utils.Assert(endMessage.HasGameData() || endMessage.HasReviewData(), "unreachable")

	var dataType protocol.DataType
	if endMessage.HasReviewData() {
		dataType = protocol.Reviews
	} else {
		dataType = protocol.Games
	}

	slog.Debug("Notifyng to coordinator", "clientId", endMessage.GetClientID(), "nodeId", s.NodeId)
	msgToSend := protocol.NewEndMessage(dataType, protocol.MessageOptions{
		ClientID:  endMessage.GetClientID(),
		MessageID: s.NodeId,
		RequestID: endMessage.GetRequestID(),
	})

	if err := s.coordinator.Write(msgToSend.Marshal(), ""); err != nil {
		slog.Error("error sending notify coordinator", err)
	}
}

func (s *Service) NotifyNeighbours(endMessage protocol.Message) {
	utils.Assert(endMessage.ExpectKind(protocol.End), "must be an END message")
	utils.Assert(endMessage.HasGameData() || endMessage.HasReviewData(), "unreachable")

	var dataType protocol.DataType
	if endMessage.HasReviewData() {
		dataType = protocol.Reviews
	} else {
		dataType = protocol.Games
	}

	msgToSend := protocol.NewEndMessage(dataType, protocol.MessageOptions{
		ClientID:  endMessage.GetClientID(),
		MessageID: s.NodeId,
		RequestID: endMessage.GetRequestID(),
	})

	if err := s.fanoutPub.Write(msgToSend.Marshal(), ""); err != nil {
		slog.Error("error sending notify neighbours", err)
	}
}
