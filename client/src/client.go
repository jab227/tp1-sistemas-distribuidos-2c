package src

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication/message"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/network"
)

type ClientConfig struct {
	ServerName    string
	ServerPort    int
	TaskQueueSize int
	ReviewsBatch  *BatchFileConfig
	GamesBatch    *BatchFileConfig
}

type Client struct {
	socket       *network.SocketTcp
	deleteSocket func()
	protocol     *communication.Protocol
	clientConfig *ClientConfig
}

func NewClient(clientConfig *ClientConfig) (*Client, func()) {
	socket, deleteSocket := newSocket(clientConfig)
	client := &Client{socket: socket, deleteSocket: deleteSocket, clientConfig: clientConfig}
	cleanup := func() {
		deleteClient(client)
	}

	return client, cleanup
}

func newSocket(clientConfig *ClientConfig) (*network.SocketTcp, func()) {
	serverAddress := fmt.Sprintf("%v:%v", clientConfig.ServerName, clientConfig.ServerPort)
	return network.NewSocketTcp(serverAddress)
}

func deleteClient(c *Client) {
	c.deleteSocket()
}

func (c *Client) Connect() error {
	if err := c.socket.Connect(); err != nil {
		return err
	}
	c.protocol = communication.NewProtocol(c.socket)
	return nil
}

func (c *Client) Execute() error {
	// Create common Task Queue
	taskQueue := NewBlockingQueue[*message.DataMessageConfig](c.clientConfig.TaskQueueSize)

	// Create Producers
	reviewsBatch, deleteReviewsBatch, err := NewBatchFile(c.clientConfig.ReviewsBatch, taskQueue)
	if err != nil {
		return err
	}
	defer deleteReviewsBatch()
	gamesBatch, deleteGamesBatch, err := NewBatchFile(c.clientConfig.GamesBatch, taskQueue)
	if err != nil {
		return err
	}
	defer deleteGamesBatch()

	// Create Consumer
	fileSender := NewFileSender(c.protocol, taskQueue)

	// Spawn producers
	reviewsBatchThread := NewThread(reviewsBatch)
	gamesBatchThread := NewThread(gamesBatch)
	gamesBatchThread.Run()
	reviewsBatchThread.Run()
	// Spawn consumer
	fileSenderThread := NewThread(fileSender)
	fileSenderThread.Run()

	// Synchronize consumers and producers
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
