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
	socket             *network.SocketTcp
	deleteSocket       func()
	clientSocket       *network.SocketTcp
	deleteClientSocket func()
	clientProtocol     *communication.Protocol
}

func NewServer(serverConfig *ServerConfig) (*Server, func()) {
	server := &Server{}
	cleanup := func() {
		deleteServer(server)
	}
	server.setSocket(serverConfig)
	return server, cleanup
}

func deleteServer(s *Server) {
	s.deleteSocket()
	s.deleteClientSocket()
}

func (s *Server) setSocket(serverConfig *ServerConfig) {
	const serverName = "localhost"
	serverAddress := fmt.Sprintf("%v:%v", serverName, serverConfig.ServicePort)
	socket, deleteSocket := network.NewSocketTcp(serverAddress)
	s.socket = socket
	s.deleteSocket = deleteSocket
}

func (s *Server) Listen() error {
	return s.socket.Listen()
}

func (s *Server) Accept() error {
	clientSocket, deleteClientSocket, err := s.socket.Accept()
	if err != nil {
		return err
	}
	s.clientSocket = clientSocket
	s.deleteClientSocket = deleteClientSocket
	s.clientProtocol = communication.NewProtocol(s.clientSocket)
	return nil
}

func (s *Server) Execute() error {
	for {
		msgData, err := s.clientProtocol.RecvDataMessage()
		if err != nil {
			return err
		}

		fmt.Println(*msgData.Payload.Header)
		fmt.Print(string(msgData.Payload.Payload.Data))
	}
}
