package src

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication/message"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/network"
)

type ClientConfig struct {
	ServerName string
	ServerPort int
}

type Client struct {
	socket       *network.SocketTcp
	deleteSocket func()
	protocol     *communication.Protocol
}

func NewClient(clientConfig *ClientConfig) (*Client, func()) {
	socket, deleteSocket := newSocket(clientConfig)
	client := &Client{socket: socket, deleteSocket: deleteSocket}
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

func (c *Client) Execute() error {
	if err := c.socket.Connect(); err != nil {
		return err
	}
	c.protocol = communication.NewProtocol(c.socket)

	// Create common Task Queue
	taskQueue := NewBlockingQueue[*message.DataMessageConfig](10)

	// Create Producer
	batchConfig := &BatchFileConfig{
		DataType:       message.Reviews,
		Path:           "/home/daniel/Documents/facultad/2024-2C/75.74/tps/tp1-sistemas-distribuidos-2c/datasets/reviews.csv",
		NlinesFromDisk: 100,
		BatchSize:      10,
		MaxBytes:       4 * 1024,
	}
	reviewsBatch, deleteReviewsBatch, err := NewBatchFile(batchConfig, taskQueue)
	if err != nil {
		return err
	}
	defer deleteReviewsBatch()

	// Create Producer
	gamesConfig := &BatchFileConfig{
		DataType:       message.Games,
		Path:           "/home/daniel/Documents/facultad/2024-2C/75.74/tps/tp1-sistemas-distribuidos-2c/datasets/games.csv",
		NlinesFromDisk: 100,
		BatchSize:      1,
		MaxBytes:       4 * 1024,
	}
	gamesBatch, deleteGamesBatch, err := NewBatchFile(gamesConfig, taskQueue)
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
