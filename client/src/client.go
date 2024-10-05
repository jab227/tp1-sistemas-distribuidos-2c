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
	client := &Client{clientConfig: clientConfig}
	cleanup := func() {
		deleteClient(client)
	}
	client.setSocket(clientConfig)
	return client, cleanup
}

func deleteClient(c *Client) {
	c.deleteSocket()
}

func (c *Client) setSocket(clientConfig *ClientConfig) {
	serverAddress := fmt.Sprintf("%v:%v", clientConfig.ServerName, clientConfig.ServerPort)
	socket, deleteSocket := network.NewSocketTcp(serverAddress)
	c.socket = socket
	c.deleteSocket = deleteSocket
}

func (c *Client) Connect() error {
	if err := c.socket.Connect(); err != nil {
		return err
	}
	c.protocol = communication.NewProtocol(c.socket)
	return nil
}

func (c *Client) Execute() error {
	taskQueue := NewBlockingQueue[*message.DataMessageConfig](c.clientConfig.TaskQueueSize)

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
	fileSender := NewFileSender(c.protocol, taskQueue)

	reviewsBatchThread := NewThread(reviewsBatch)
	gamesBatchThread := NewThread(gamesBatch)
	fileSenderThread := NewThread(fileSender)
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

	for {
		result, err := c.protocol.RecvResultMessage()
		if err != nil {
			return err
		}
		fmt.Println(result)
	}
}
