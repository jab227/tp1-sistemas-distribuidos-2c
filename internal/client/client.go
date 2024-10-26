package client

import (
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/cprotocol"
	"net"
	"strings"
)

type Client struct {
	conn   net.Conn
	config *Config

	ClientId  uint64
	RequestId uint64
}

func NewClient(config *Config) (*Client, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", config.ServerName, config.ServerPort))
	if err != nil {
		return nil, fmt.Errorf("could not connect to server: %w", err)
	}

	return &Client{
		conn:   conn,
		config: config,
	}, nil
}

// TODO(fede) - Add signal handling
func (c *Client) Run() error {
	if err := c.MakeHandshake(); err != nil {
		return fmt.Errorf("could not make handshake: %w", err)
	}
	if err := c.SendGames(); err != nil {
		return fmt.Errorf("could not send games: %w", err)
	}
	if err := c.SendReviews(); err != nil {
		return fmt.Errorf("could not send reviews: %w", err)
	}

	return nil
}

func (c *Client) MakeHandshake() error {
	if err := cprotocol.SendSyncMsg(c.conn); err != nil {
		return fmt.Errorf("could not send sync message: %w", err)
	}

	ackSyncMsg, err := cprotocol.ReadAckSyncMsg(c.conn)
	if err != nil {
		return fmt.Errorf("could not read ack sync message: %w", err)
	}

	// Defines request and client ids
	c.ClientId = ackSyncMsg.Header.ClientId
	c.RequestId = ackSyncMsg.Header.RequestId

	return nil
}

func (c *Client) SendGames() error {
	gamesReader, err := NewFileBatcher(c.config.GamesConfig)
	if err != nil {
		return fmt.Errorf("could not create games reader: %w", err)
	}
	defer gamesReader.Close()

	// Send the games start msg
	if err := cprotocol.SendStartMsg(c.conn, cprotocol.Games, c.ClientId, c.RequestId); err != nil {
		return fmt.Errorf("could not send start games message: %w", err)
	}

	for lines := range gamesReader.Contents() {
		bytes := []byte(strings.Join(lines, "\n"))
		if err := cprotocol.SendDataMsg(c.conn, cprotocol.Games, c.ClientId, c.RequestId, bytes); err != nil {
			return fmt.Errorf("could not send data message: %w", err)
		}
	}

	if err := cprotocol.SendEndMsg(c.conn, cprotocol.Games, c.ClientId, c.RequestId); err != nil {
		return fmt.Errorf("could not send end game message: %w", err)
	}

	return nil
}

func (c *Client) SendReviews() error {
	reviewsReader, err := NewFileBatcher(c.config.ReviewsConfig)
	if err != nil {
		return fmt.Errorf("could not create reviews reader: %w", err)
	}
	defer reviewsReader.Close()

	if err := cprotocol.SendStartMsg(c.conn, cprotocol.Reviews, c.ClientId, c.RequestId); err != nil {
		return fmt.Errorf("could not send start reviews message: %w", err)
	}

	for lines := range reviewsReader.Contents() {
		bytes := []byte(strings.Join(lines, "\n"))
		if err := cprotocol.SendDataMsg(c.conn, cprotocol.Reviews, c.ClientId, c.RequestId, bytes); err != nil {
			return fmt.Errorf("could not send data message: %w", err)
		}
	}

	if err := cprotocol.SendEndMsg(c.conn, cprotocol.Reviews, c.ClientId, c.RequestId); err != nil {
		return fmt.Errorf("could not send end reviews message: %w", err)
	}

	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
