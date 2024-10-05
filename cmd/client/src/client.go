package src

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/network"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
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
	return c.protocol.Sync()
}

func (c *Client) Execute() error {
	sender := NewSender(c.clientConfig, c.protocol)
	receiver := NewReceiver(c.clientConfig, c.protocol)
	senderThread := utils.NewThread(sender)
	receiverThread := utils.NewThread(receiver)
	senderThread.Run()
	receiverThread.Run()

	if err := senderThread.Join(); err != nil {
		return err
	}
	if err := receiverThread.Join(); err != nil {
		return err
	}
	return nil
}
