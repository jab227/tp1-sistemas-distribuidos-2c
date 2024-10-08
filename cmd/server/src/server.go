package src

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/results"

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

	inputManager    *client.IOManager
	outputManager   *client.IOManager
	clientIdCounter uint32
	done            chan struct{}
}

func NewServer(serverConfig *ServerConfig, inputManager *client.IOManager, outputManager *client.IOManager) (*Server, func()) {
	server := &Server{inputManager: inputManager, outputManager: outputManager, done: make(chan struct{})}
	cleanup := func() {
		deleteServer(server)
	}
	server.setSocket(serverConfig)

	return server, cleanup
}

func deleteServer(s *Server) {
	s.deleteSocket()
	s.deleteClient()
	s.inputManager.Close()
	s.outputManager.Close()
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

	if err := s.Listen(); err != nil {
		return fmt.Errorf("error when listenning %s", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
			if err := s.Accept(); err != nil {
				return fmt.Errorf("error when accepting connection %s", err)
			}

			if err := s.StartClient(ctx); err != nil {
				return fmt.Errorf("error when executing command %s", err)
			}
		}
	}
}

func (s *Server) setSocket(serverConfig *ServerConfig) {
	serverAddress := fmt.Sprintf("0.0.0.0:%v", serverConfig.ServicePort)
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

func (s *Server) GetClientConn() net.Conn {
	return s.client.socket.GetConnection()
}
func (s *Server) StartClient(ctx context.Context) error {
	if err := s.client.Execute(s.outputManager); err != nil {
		return fmt.Errorf("error receiving data from client %s", err)
	}

	service := results.NewResultsService(s.client.protocol, s.inputManager)
	go service.Run(ctx)
	<-service.Done()
	return nil
}
