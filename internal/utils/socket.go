package utils

import (
	"errors"
	"fmt"
	"io"
	"net"
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
			return fmt.Errorf("failed to recv message of size: %v", size)
		}

		recvData += bytesRecv
	}

	return nil
}
