package src

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication/message"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/network"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/utils"
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

func (c *Client) Run() error {
	if err := c.socket.Connect(); err != nil {
		return err
	}
	c.protocol = communication.NewProtocol(c.socket)

	// Preparing for file reading
	nlines := 100
	fpath := "/home/daniel/Documents/facultad/2024-2C/75.74/tps/tp1-sistemas-distribuidos-2c/datasets/reviews.csv"
	fileReader, deleteFileLinesReader, err := NewFileLinesReader(fpath, nlines)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer deleteFileLinesReader()

	// Sending reviews
	lines, more, err := fileReader.Read()
	batchSize := 10
	maxBytes := 4 * 1024
	batch := utils.NewBatchLines(lines, batchSize, maxBytes)
	if len(lines) == 0 {
		return err
	}

	c.sendStart()
	for {
		batch.Run(c.sendData)
		if !more {
			break
		}
		lines, more, _ = fileReader.Read()
		batch.Reset(lines)
	}
	c.sendEnd()
	return nil
}

func (c *Client) sendStart() error {
	dataConfig := &message.DataMessageConfig{
		Start:    true,
		DataType: message.Reviews,
	}
	return c.protocol.SendDataMessage(dataConfig)
}

func (c *Client) sendData(data string) error {
	dataConfig := &message.DataMessageConfig{
		DataType: message.Reviews,
		Data:     []byte(data),
	}
	return c.protocol.SendDataMessage(dataConfig)
}

func (c *Client) sendEnd() error {
	dataConfig := &message.DataMessageConfig{
		End:      true,
		DataType: message.Reviews,
	}
	return c.protocol.SendDataMessage(dataConfig)
}
