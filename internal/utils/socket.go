package utils

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// WriteToSocket sends a message through the provided network connection.
// It ensures the entire message is sent, handling short writes. Returns an
// error if the send operation fails for reasons other than short writes.
func WriteToSocket(connection net.Conn, msg []byte) error {
	sentData := 0
	for sentData < len(msg) {
		bytesSent, err := connection.Write(msg[sentData:])
		if err != nil && !errors.Is(err, io.ErrShortWrite) {
			return fmt.Errorf("failed to send message: %v", msg)
		}

		sentData += bytesSent
	}

	return nil
}

// ReadFromSocket reads data from the provided network connection into the given
// buffer. It ensures the buffer is filled up to the specified size, handling
// short reads. Returns an error if the read operation fails.
func ReadFromSocket(connection net.Conn, buffer *[]byte, size int) error {
	recvData := 0
	internalBuffer := *buffer
	for recvData < size {
		bytesRecv, err := connection.Read(internalBuffer[recvData:])
		if err != nil {
			return fmt.Errorf("failed to recv message of size: %v - err %s", size, err)
		}

		recvData += bytesRecv
	}

	return nil
}

// WriteToUDPSocket sends a message to the specified UDP address using the provided
// UDP connection. It ensures the entire message is sent, handling short writes.
// Returns an error if the send operation fails.
func WriteToUDPSocket(connection *net.UDPConn, msg []byte, addr *net.UDPAddr) error {
	sentData := 0
	for sentData < len(msg) {
		bytesSent, err := connection.WriteToUDP(msg[sentData:], addr)
		if err != nil {
			return fmt.Errorf("failed to send message: %v", msg)
		}

		sentData += bytesSent
	}

	return nil
}

// ReadFromUDPSocket reads data from the provided UDP connection into the given
// buffer. It ensures the buffer is filled up to the specified size, handling
// short reads. Returns the sender's address and any error encountered.
func ReadFromUDPSocket(connection *net.UDPConn, buffer *[]byte, size int) (*net.UDPAddr, error) {
	recvData := 0
	internalBuffer := *buffer
	var finalAddr *net.UDPAddr
	for recvData < size {
		bytesRecv, addr, err := connection.ReadFromUDP(internalBuffer[recvData:])
		if err != nil {
			return nil, fmt.Errorf("failed to receive message of size: %v - err %s", size, err)
		}
		recvData += bytesRecv
		finalAddr = addr
	}
	return finalAddr, nil
}

// ReadFromUDPSocketWithTimeout reads data from the provided UDP connection into the given
// buffer. It ensures the buffer is filled up to the specified size, handling
// short reads. Returns the sender's address and any error encountered.
func ReadFromUDPSocketWithTimeout(connection *net.UDPConn, buffer *[]byte, size int, timeout time.Duration) (*net.UDPAddr, error) {
	recvData := 0
	internalBuffer := *buffer
	var finalAddr *net.UDPAddr
	for recvData < size {
		if err := connection.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return nil, fmt.Errorf("failed to set read deadline: %v", err)
		}
		bytesRecv, addr, err := connection.ReadFromUDP(internalBuffer[recvData:])
		if err != nil {
			return nil, err
		}

		recvData += bytesRecv
		finalAddr = addr
	}

	return finalAddr, nil
}
