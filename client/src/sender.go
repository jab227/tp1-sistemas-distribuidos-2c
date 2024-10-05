package src

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication/message"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type Sender struct {
	clientConfig *ClientConfig
	protocol     *communication.Protocol
}

func NewSender(clientConfig *ClientConfig, protocol *communication.Protocol) *Sender {
	return &Sender{clientConfig: clientConfig, protocol: protocol}
}

func (s *Sender) Run(join chan error) {
	if err := s.send(); err != nil {
		join <- err
	}
	join <- nil
}

func (s *Sender) send() error {
	taskQueue := utils.NewBlockingQueue[*message.DataMessageConfig](s.clientConfig.TaskQueueSize)

	reviewsBatch, deleteReviewsBatch, err := NewBatchFile(s.clientConfig.ReviewsBatch, taskQueue)
	if err != nil {
		return err
	}
	defer deleteReviewsBatch()
	gamesBatch, deleteGamesBatch, err := NewBatchFile(s.clientConfig.GamesBatch, taskQueue)
	if err != nil {
		return err
	}
	defer deleteGamesBatch()
	fileSender := NewFileSender(s.protocol, taskQueue)

	reviewsBatchThread := utils.NewThread(reviewsBatch)
	gamesBatchThread := utils.NewThread(gamesBatch)
	fileSenderThread := utils.NewThread(fileSender)
	gamesBatchThread.Run()
	reviewsBatchThread.Run()
	fileSenderThread.Run()

	if err := reviewsBatchThread.Join(); err != nil {
		return err
	}
	if err := gamesBatchThread.Join(); err != nil {
		return err
	}
	taskQueue.Close()
	if err := fileSenderThread.Join(); err != nil {
		return err
	}
	return nil
}
