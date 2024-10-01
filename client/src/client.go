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

func (c *Client) Run() error {
	if err := c.socket.Connect(); err != nil {
		return err
	}

	// Communication library usage
	protocol := communication.NewProtocol(c.socket)

	protocol.SendSyncMessage()

	synAckConfig := &message.SyncAckMessageConfig{
		ClientId:  5,
		RequestId: 77,
	}
	protocol.SendSyncAckMessage(synAckConfig)

	dataConfig := &message.DataMessageConfig{
		ClientId:       5,
		RequestId:      77,
		SequenceNumber: 912,
		Start:          true,
		End:            false,
		DataType:       message.Games,
		Data:           []byte("Data - Hello, world!"),
	}
	protocol.SendDataMessage(dataConfig)

	resultConfig := &message.ResultMessageConfig{
		ClientId:       5,
		RequestId:      77,
		SequenceNumber: 812,
		Start:          false,
		End:            true,
		ResultType:     message.Query5,
		Data:           []byte("Result - Hello, world!"),
	}
	protocol.SendResultMessage(resultConfig)

	return nil
}
