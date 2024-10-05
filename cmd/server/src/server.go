package src

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/network"
)

type ServerConfig struct {
	ServicePort int
}

type Server struct {
	socket       *network.SocketTcp
	deleteSocket func()
	client       *Client
	deleteClient func()
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
	s.deleteClient()
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

	client, deleteClient := NewClient(clientSocket, deleteClientSocket)
	s.client = client
	s.deleteClient = deleteClient
	return nil
}

func (s *Server) Execute() error {
	return s.client.Execute()
}
