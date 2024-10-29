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

type MessageInfo struct {
	DataType protocol.DataType
	Options  protocol.MessageOptions
}

type ServiceOptions struct {
	Exchange         string
	SubscriberQueue  string
	CoordinatorQueue string
	Timeout          uint8
}

type WorkerQueueService struct {
	fanoutPub   *rabbitmq.FanoutPublisher
	fanoutSub   *rabbitmq.FanoutSubscriber
	coordinator *rabbitmq.WorkerQueue
	notify      chan protocol.Message
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

func NewWorkerQueueService(opts *ServiceOptions) (*WorkerQueueService, error) {
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

	return &WorkerQueueService{
		fanoutPub:   fanoutPub,
		fanoutSub:   fanoutSub,
		coordinator: coordinator,
		notify:      make(chan protocol.Message),
	}, nil
}

func (s *WorkerQueueService) Destroy() {
	s.fanoutPub.Close()
	s.fanoutSub.Close()
	s.coordinator.Close()
}

func (s *WorkerQueueService) Run(ctx context.Context) (chan<- protocol.Message, <-chan MessageInfo) {
	tx := make(chan protocol.Message, 1)
	rx := make(chan MessageInfo, 1)
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
				} else {
					t = protocol.Reviews
				}
				rx <- MessageInfo{
					Options: protocol.MessageOptions{
						MessageID: msg.GetMessageID(),
						ClientID:  msg.GetClientID(),
						RequestID: msg.GetRequestID(),
					},
					DataType: t,
				}
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

func (s *WorkerQueueService) NotifyCoordinator(d protocol.DataType, opts protocol.MessageOptions) {
	msg := protocol.NewEndMessage(d, opts)
	s.notify <- msg
}

type RouterService struct {
	coordinator *rabbitmq.WorkerQueue
	notify      chan protocol.Message
}

func NewRouterService(opts *ServiceOptions) (*RouterService, error) {
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
	return &RouterService{
		coordinator: coordinator,
		notify:      make(chan protocol.Message, 1),
	}, nil
}

func (r *RouterService) Run(ctx context.Context) {
	go func() {
		for {
			select {
			case msg := <-r.notify:
				r.coordinator.Write(msg.Marshal(), "")
			case <-ctx.Done():
				slog.Error("context error", "message", ctx.Err())
			}
		}
	}()
}

func (r *RouterService) NotifyCoordinator(d protocol.DataType, opts protocol.MessageOptions) {
	msg := protocol.NewEndMessage(d, opts)
	r.notify <- msg
}
