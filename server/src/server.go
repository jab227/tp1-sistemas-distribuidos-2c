package src

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/network"
)

type ServerConfig struct {
	ServicePort int
}

type Server struct {
	socket       *network.SocketTcp
	deleteSocket func()
}

func NewServer(serverConfig *ServerConfig) (*Server, func()) {
	socket, deleteSocket := newSocket(serverConfig)
	server := &Server{socket: socket, deleteSocket: deleteSocket}
	cleanup := func() {
		deleteServer(server)
	}
	return server, cleanup
}

func newSocket(serverConfig *ServerConfig) (*network.SocketTcp, func()) {
	const serverName = "localhost"
	serverAddress := fmt.Sprintf("%v:%v", serverName, serverConfig.ServicePort)
	return network.NewSocketTcp(serverAddress)
}

func deleteServer(c *Server) {
	c.deleteSocket()
}

func (c *Server) Run() error {
	if err := c.socket.Listen(); err != nil {
		return err
	}

	clientSocket, deleteClientSocket, err := c.socket.Accept()
	if err != nil {
		return err
	}
	defer deleteClientSocket()

	// Communication library usage
	protocol := communication.NewProtocol(clientSocket)

	msgSync, _ := protocol.RecvSyncMessage()
	fmt.Println(msgSync.Header.Optype)
	fmt.Println(msgSync.Header.ClientId)
	fmt.Println(msgSync.Header.RequestId)
	fmt.Println(msgSync.Header.PayloadSize)
	fmt.Println(msgSync.Payload)
	fmt.Println("")

	msgSyncAck, _ := protocol.RecvSyncAckMessage()
	fmt.Println(msgSyncAck.Header.Optype)
	fmt.Println(msgSyncAck.Header.ClientId)
	fmt.Println(msgSyncAck.Header.RequestId)
	fmt.Println(msgSyncAck.Header.PayloadSize)
	fmt.Println(msgSyncAck.Payload)
	fmt.Println("")

	msgData, _ := protocol.RecvDataMessage()
	fmt.Println(msgData.Header.Optype)
	fmt.Println(msgData.Header.ClientId)
	fmt.Println(msgData.Header.RequestId)
	fmt.Println(msgData.Header.PayloadSize)
	fmt.Println(msgData.Payload.Header.Optype)
	fmt.Println(msgData.Payload.Header.SequenceNumber)
	fmt.Println(msgData.Payload.Header.Start)
	fmt.Println(msgData.Payload.Header.End)
	fmt.Println(string(msgData.Payload.Payload.Data))
	fmt.Println("")

	msgResult, _ := protocol.RecvResultMessage()
	fmt.Println(msgResult.Header.Optype)
	fmt.Println(msgResult.Header.ClientId)
	fmt.Println(msgResult.Header.RequestId)
	fmt.Println(msgResult.Header.PayloadSize)
	fmt.Println(msgResult.Payload.Header.Optype)
	fmt.Println(msgResult.Payload.Header.SequenceNumber)
	fmt.Println(msgResult.Payload.Header.Start)
	fmt.Println(msgResult.Payload.Header.End)
	fmt.Println(string(msgResult.Payload.Payload.Data))
	return nil
}
