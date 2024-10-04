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

	protocol := communication.NewProtocol(clientSocket)
	for {
		msgData, err := protocol.RecvDataMessage()
		if err != nil {
			fmt.Println(err)
			break
		}

		fmt.Println(*msgData.Payload.Header)
		fmt.Print(string(msgData.Payload.Payload.Data))
	}
	return nil
}
