package src

import (
	"context"
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"os"
	"strconv"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/network"
)

const ServerPortEnv = "SERVER_PORT"

type ServerConfig struct {
	ServicePort int
}

func GetServerConfigFromEnv() (*ServerConfig, error) {
	value, ok := os.LookupEnv(ServerPortEnv)
	if !ok {
		return nil, fmt.Errorf("environment variable %s not set", ServerPortEnv)
	}

	port, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("environment variable %s is not a number", ServerPortEnv)
	}

	return &ServerConfig{ServicePort: port}, nil
}

type Server struct {
	socket       *network.SocketTcp
	deleteSocket func()
	client       *Client
	deleteClient func()

	ioManager       *client.IOManager
	clientIdCounter uint32
	done            chan struct{}
}

func NewServer(serverConfig *ServerConfig, ioManager *client.IOManager) (*Server, func()) {
	server := &Server{ioManager: ioManager, done: make(chan struct{})}
	cleanup := func() {
		deleteServer(server)
	}
	server.setSocket(serverConfig)

	return server, cleanup
}

func deleteServer(s *Server) {
	s.deleteSocket()
	s.deleteClient()
	s.ioManager.Close()
}

func (s *Server) GetClientId() uint32 {
	defer func() { s.clientIdCounter++ }()
	return s.clientIdCounter
}

func (s *Server) GetDone() <-chan struct{} {
	return s.done
}

func (s *Server) DoneSignal() {
	s.done <- struct{}{}
}

func (s *Server) Run(ctx context.Context) error {
	defer s.DoneSignal()

	for {
		select {
		case <-ctx.Done():
			return nil

		default:
			if err := s.Listen(); err != nil {
				return fmt.Errorf("error when listenning %s", err)
			}

			if err := s.Accept(); err != nil {
				return fmt.Errorf("error when accepting connection %s", err)
			}

			if err := s.Execute(); err != nil {
				return fmt.Errorf("error when executing command %s", err)
			}
		}
	}
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

	client, deleteClient, err := NewClient(
		clientSocket,
		s.GetClientId(),
		deleteClientSocket,
	)
	if err != nil {
		return err
	}
	s.client = client
	s.deleteClient = deleteClient
	return nil
}

func (s *Server) Execute() error {
	return s.client.Execute(s.ioManager)
}
