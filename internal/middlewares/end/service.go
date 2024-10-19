package end

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/env"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/rabbitmq"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type Service struct {
	fanoutPub   *rabbitmq.FanoutPublisher
	fanoutSub   *rabbitmq.FanoutSubscriber
	coordinator *rabbitmq.WorkerQueue
	notify      chan protocol.Message
}

type ServiceOptions struct {
	Exchange         string
	SubscriberQueue  string
	CoordinatorQueue string
	Timeout          uint8
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
		Timeout:          uint8(*timeout),
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
	}, nil
}

func (s *Service) Destroy() {
	s.fanoutPub.Close()
	s.fanoutSub.Close()
	s.coordinator.Close()
}

func (s *Service) Run(ctx context.Context) (chan<- protocol.Message, <-chan protocol.DataType) {
	tx := make(chan protocol.Message, 1)
	rx := make(chan protocol.DataType, 1)
	go func() {
		consumerCh := s.fanoutSub.GetConsumer()
		for {
			select {
			// FROM THE QUEUE
			case msg := <-tx:
				utils.Assert(msg.ExpectKind(protocol.End), "must be an END message")
				utils.Assert(msg.HasGameData() || msg.HasReviewData(), "unreachable")

				marshaledMsg := msg.Marshal()
				s.fanoutPub.Write(marshaledMsg, "")
			// FROM MY BROTHERS
			// NEED SOME OATS BROTHER
			case delivery := <-consumerCh:
				msgBytes := delivery.Body
				var msg protocol.Message
				if err := msg.Unmarshal(msgBytes); err != nil {
					slog.Error("couldn't unmarshal message", "error", err)
					continue
				}
				utils.Assert(msg.ExpectKind(protocol.End), "must be an END message")
				// Notify that I received an END
				var t protocol.DataType
				if msg.HasGameData() {
					t = protocol.Games
				} else if msg.HasReviewData() {
					t = protocol.Reviews
				} else {
					utils.Assert(false, "unreachable")
				}
				rx <- t
				// Acknowledge
				delivery.Ack(false)
			case msg := <-s.notify:
				s.coordinator.Write(msg.Marshal(), "")
			case <-ctx.Done():
				slog.Error("context error", "error", ctx.Err())
				return
			}
		}
	}()
	return tx, rx
}

func (s *Service) NotifyCoordinator(d protocol.DataType, opts protocol.MessageOptions) {
	msg := protocol.NewEndMessage(d, opts)
	s.notify <- msg
}
