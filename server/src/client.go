package src

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/network"
)

type Client struct {
	socket       *network.SocketTcp
	deleteSocket func()
	protocol     *communication.Protocol
}

func NewClient(socket *network.SocketTcp, deleteSocket func()) (*Client, func()) {
	client := &Client{socket: socket, deleteSocket: deleteSocket}
	client.protocol = communication.NewProtocol(socket)
	cleanup := func() {
		deleteClient(client)
	}
	return client, cleanup
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

		fmt.Println(*msgData.Payload.Header)
		fmt.Print(string(msgData.Payload.Payload.Data))
	}
}
