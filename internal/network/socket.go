package network

import (
	"bufio"
	"net"
)

const (
	tpcNetwork = "tcp"
	bufferSize = 4 * 1024
)

type SocketTcp struct {
	address        string
	connection     net.Conn
	listener       net.Listener
	bufferedReader *bufio.Reader
}

func NewSocketTcp(address string) (*SocketTcp, func()) {
	socket := &SocketTcp{address: address}
	cleanup := func() {
		deleteSocketTcp(socket)
	}
	return socket, cleanup
}

func deleteSocketTcp(s *SocketTcp) error {
	if err := s.closeConnection(); err != nil {
		return err
	}
	if err := s.closeListener(); err != nil {
		return err
	}
	return nil
}

func (s *SocketTcp) closeConnection() error {
	if s.connection == nil {
		return nil
	}
	return s.connection.Close()
}

func (s *SocketTcp) closeListener() error {
	if s.listener == nil {
		return nil
	}
	return s.listener.Close()
}

func (s *SocketTcp) Connect() error {
	connection, err := net.Dial(tpcNetwork, s.address)
	if err != nil {
		return err
	}
	s.prepareConnection(connection)
	return nil
}

func (s *SocketTcp) Accept() (*SocketTcp, func(), error) {
	connection, err := s.listener.Accept()
	if err != nil {
		return nil, nil, err
	}

	socket, deleteSocket := NewSocketTcp(connection.RemoteAddr().String())
	socket.prepareConnection(connection)
	return socket, deleteSocket, nil
}

func (s *SocketTcp) prepareConnection(connection net.Conn) {
	s.connection = connection
	s.bufferedReader = bufio.NewReaderSize(s.connection, bufferSize)
}

func (s *SocketTcp) Listen() error {
	listener, err := net.Listen(tpcNetwork, s.address)
	if err != nil {
		return err
	}
	s.listener = listener
	return nil
}

func (s *SocketTcp) Send(data []byte) error {
	remainingBytes := len(data)
	for remainingBytes > 0 {
		n, err := s.connection.Write(data)
		if err != nil {
			return err
		}

		remainingBytes -= n
		data = data[n:]
	}
	return nil
}

func (s *SocketTcp) Receive(buffer []byte) error {
	remainingBytes := len(buffer)
	for remainingBytes > 0 {
		n, err := s.bufferedReader.Read(buffer)
		if err != nil {
			return err
		}

		remainingBytes -= n
		buffer = buffer[n:]
	}
	return nil
}

func (s *SocketTcp) GetConnection() net.Conn {

}
