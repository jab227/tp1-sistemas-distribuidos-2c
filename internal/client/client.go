package client

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/cprotocol"
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
	if err := c.RecvResults(); err != nil {
		return fmt.Errorf("could not receive results: %w", err)
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
	slog.Info("starts sending games data")
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
	slog.Info("finish sending games data")

	return nil
}

func (c *Client) SendReviews() error {
	reviewsReader, err := NewFileBatcher(c.config.ReviewsConfig)
	if err != nil {
		return fmt.Errorf("could not create reviews reader: %w", err)
	}
	defer reviewsReader.Close()

	slog.Info("starts sending reviews data")
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
	slog.Info("finish sending reviews data")

	return nil
}

func (c *Client) RecvResults() error {
	queriesReceived := make(map[cprotocol.ContentType]struct{})
	for len(queriesReceived) < 5 {
		msg, err := cprotocol.ReadResultMsg(c.conn)
		if err != nil {
			return fmt.Errorf("could not read results message: %w", err)
		}

		if msg.Header.ContentType == cprotocol.Query1 {
			if _, ok := queriesReceived[cprotocol.Query1]; ok {
				continue
			}
			query1 := strings.Split(string(msg.Payload), ",")
			windows := query1[0]
			mac := query1[1]
			linux := query1[2]
			fmt.Fprintf(os.Stdout, "===========\n")
			fmt.Fprintf(os.Stdout, "Query 1:\n")
			fmt.Fprintf(os.Stdout, "===========\n")
			fmt.Fprintf(os.Stdout, "windows: %s\n", windows)
			fmt.Fprintf(os.Stdout, "mac: %s\n", mac)
			fmt.Fprintf(os.Stdout, "linux: %s\n", linux)
			queriesReceived[cprotocol.Query1] = struct{}{}
		} else if msg.Header.ContentType == cprotocol.Query2 {
			if _, ok := queriesReceived[cprotocol.Query2]; ok {
				continue
			}
			query2 := strings.Split(string(msg.Payload), "\n")
			fmt.Fprintf(os.Stdout, "===========\n")
			fmt.Fprintf(os.Stdout, "Query 2:\n")
			fmt.Fprintf(os.Stdout, "===========\n")
			appIdsProcessed := make(map[string]struct{})
			for i, s := range query2 {
				queryFields := strings.Split(s, ",")
				appId := queryFields[0]
				if _, ok := appIdsProcessed[appId]; ok {
					continue
				}
				appIdsProcessed[appId] = struct{}{}
				result := queryFields[1]
				fmt.Fprintf(os.Stdout, "%d: %s\n", i+1, result)
			}
			queriesReceived[cprotocol.Query2] = struct{}{}
		} else if msg.Header.ContentType == cprotocol.Query3 {
			if _, ok := queriesReceived[cprotocol.Query3]; ok {
				continue
			}
			query3 := strings.Split(string(msg.Payload), "\n")
			fmt.Fprintf(os.Stdout, "===========\n")
			fmt.Fprintf(os.Stdout, "Query 3:\n")
			fmt.Fprintf(os.Stdout, "===========\n")
			appIdsProcessed := make(map[string]struct{})
			for i, s := range query3 {
				queryFields := strings.Split(s, ",")
				appId := queryFields[0]
				if _, ok := appIdsProcessed[appId]; ok {
					continue
				}
				appIdsProcessed[appId] = struct{}{}
				result := queryFields[1]
				fmt.Fprintf(os.Stdout, "%d: %s\n", i+1, result)
			}
			queriesReceived[cprotocol.Query3] = struct{}{}
		} else if msg.Header.ContentType == cprotocol.Query4 {
			if _, ok := queriesReceived[cprotocol.Query4]; ok {
				continue
			}
			query4 := strings.Split(string(msg.Payload), "\n")
			fmt.Fprintf(os.Stdout, "===========\n")
			fmt.Fprintf(os.Stdout, "Query 4:\n")
			fmt.Fprintf(os.Stdout, "===========\n")
			appIdsProcessed := make(map[string]struct{})
			for i, s := range query4 {
				queryFields := strings.Split(s, ",")
				appId := queryFields[0]
				if _, ok := appIdsProcessed[appId]; ok {
					continue
				}
				appIdsProcessed[appId] = struct{}{}
				var result string
				if len(queryFields) > 1 {
					result = queryFields[1]
				} else {
					result = ""
				}
				fmt.Fprintf(os.Stdout, "%d: %s\n", i+1, result)
			}
			queriesReceived[cprotocol.Query4] = struct{}{}
		} else if msg.Header.ContentType == cprotocol.Query5 {
			if _, ok := queriesReceived[cprotocol.Query5]; ok {
				continue
			}
			query5 := strings.Split(string(msg.Payload), "\n")
			fmt.Fprintf(os.Stdout, "===========\n")
			fmt.Fprintf(os.Stdout, "Query 5:\n")
			fmt.Fprintf(os.Stdout, "===========\n")
			appIdsProcessed := make(map[string]struct{})
			for i, s := range query5 {
				queryFields := strings.Split(s, ",")
				appId := queryFields[0]
				if _, ok := appIdsProcessed[appId]; ok {
					continue
				}
				appIdsProcessed[appId] = struct{}{}
				result := queryFields[1]
				fmt.Fprintf(os.Stdout, "%d: %s\n", i+1, result)
			}
			queriesReceived[cprotocol.Query5] = struct{}{}
		} else {
			slog.Debug(fmt.Sprintf("Unknown query type: %d\n", msg.Header.ContentType))
		}
	}

	return nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
