package src

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/network"
)

type Client struct {
	socket       *network.SocketTcp
	deleteSocket func()
	protocol     *communication.Protocol
}

func NewClient(socket *network.SocketTcp, deleteSocket func()) (*Client, func(), error) {
	client := &Client{socket: socket, deleteSocket: deleteSocket}
	client.protocol = communication.NewProtocol(socket)
	if err := client.protocol.SyncAck(); err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		deleteClient(client)
	}
	return client, cleanup, nil
}

func deleteClient(c *Client) {
	c.deleteSocket()
}

func (c *Client) Execute() error {
	for {
		msgData, err := c.protocol.RecvDataMessage()
		if err != nil {
			return err
		}

		fmt.Println(*msgData.Header)
		fmt.Println(*msgData.Payload.Header)
		fmt.Print(string(msgData.Payload.Payload.Data))
	}
}
