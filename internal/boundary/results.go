package boundary

import (
	"context"
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/cprotocol"
	"log/slog"
	"net"
	"slices"
	"strings"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type query1 struct {
	windows uint32
	mac     uint32
	linux   uint32
}

type query2 []string

type query3 []string

type query4 []string

type query5 []string

type receivedQuerys uint8

const (
	query1Received    receivedQuerys = 1 << 0
	query2Received    receivedQuerys = 1 << 1
	query3Received    receivedQuerys = 1 << 2
	query4Received    receivedQuerys = 1 << 3
	query5Received    receivedQuerys = 1 << 4
	allQuerysReceived receivedQuerys = query1Received | query2Received | query3Received | query4Received | query5Received
)

type results struct {
	q1       query1
	q2       query2
	q3       query3
	q4       query4
	q5       query5
	received receivedQuerys
}

type ResultsService struct {
	client net.Conn
	io     *client.IOManager
	done   chan struct{}
	res    *results
}

// TODO(fede) - Refactor for multiple clients
// I don't own the connection
func NewResultsService(conn net.Conn, io *client.IOManager) *ResultsService {
	return &ResultsService{
		client: conn,
		io:     io,
		done:   make(chan struct{}),
		res:    &results{},
	}
}

func (r *ResultsService) Done() <-chan struct{} {
	return r.done
}

func (r *ResultsService) Run(ctx context.Context) error {
	consumerCh := r.io.Input.GetConsumer()
	defer func() {
		r.done <- struct{}{}
	}()

	for {
		select {
		case delivery := <-consumerCh:
			msgBytes := delivery.Body
			var msg protocol.Message
			if err := msg.Unmarshal(msgBytes); err != nil {
				return fmt.Errorf("couldn't unmarshal protocol message: %w", err)
			}
			if msg.ExpectKind(protocol.Results) {
				queryNumber := msg.GetQueryNumber()
				elements := msg.Elements()
				switch queryNumber {
				case 1:
					slog.Debug("received result", "query", 1, "clientId", msg.GetClientID())
					for _, element := range elements.Iter() {
						r.res.q1 = query1{
							windows: element.ReadUint32(),
							mac:     element.ReadUint32(),
							linux:   element.ReadUint32(),
						}
					}
					r.res.received |= query1Received

					slog.Debug("sending result", "query", 1, "clientId", msg.GetClientID())
					data := []byte(fmt.Sprintf("%d,%d,%d", r.res.q1.windows, r.res.q1.mac, r.res.q1.linux))
					if err := cprotocol.SendResultMsg(r.client, cprotocol.Query1, uint64(msg.GetClientID()), uint64(msg.GetRequestID()), data); err != nil {
						return fmt.Errorf("couldn't send query 1 to client: %w", err)
					}
				case 2:
					slog.Debug("received result", "query", 2, "clientId", msg.GetClientID())
					for _, element := range elements.Iter() {
						r.res.q2 = append(r.res.q2, string(element.ReadBytes()))
					}
				case 3:
					slog.Debug("received result", "query", 3, "clientId", msg.GetClientID())
					for _, element := range elements.Iter() {
						r.res.q3 = append(r.res.q3, string(element.ReadBytes()))
					}
				case 4:
					slog.Debug("received result", "query", 4, "clientId", msg.GetClientID())
					for _, element := range elements.Iter() {
						r.res.q4 = append(r.res.q4, string(element.ReadBytes()))
					}
				case 5:
					slog.Debug("received result", "query", 5, "clientId", msg.GetClientID())
					for _, element := range elements.Iter() {
						r.res.q5 = append(r.res.q5, string(element.ReadBytes()))
					}
				default:
					utils.Assertf(false, "query number %d should not happen", queryNumber)
				}
			} else if msg.ExpectKind(protocol.End) {
				queryNumber := msg.GetQueryNumber()
				switch queryNumber {
				case 2:
					slog.Debug("sending result", "query", 2, "clientId", msg.GetClientID())
					r.res.received |= query2Received
					data := []byte(strings.Join(r.res.q2[:], "\n"))
					if err := cprotocol.SendResultMsg(r.client, cprotocol.Query2, uint64(msg.GetClientID()), uint64(msg.GetRequestID()), data); err != nil {
						return fmt.Errorf("couldn't send query 2 to client: %w", err)
					}
				case 3:
					slog.Debug("sending result", "query", 3, "clientId", msg.GetClientID())
					data := []byte(strings.Join(r.res.q3[:], "\n"))
					if err := cprotocol.SendResultMsg(r.client, cprotocol.Query3, uint64(msg.GetClientID()), uint64(msg.GetRequestID()), data); err != nil {
						return fmt.Errorf("couldn't send query 3 to client: %w", err)
					}
				case 4:
					slog.Debug("sending result", "query", 4, "clientId", msg.GetClientID())
					r.res.received |= query4Received
					slices.Sort(r.res.q4)
					data := []byte(strings.Join(r.res.q4, "\n"))
					if err := cprotocol.SendResultMsg(r.client, cprotocol.Query4, uint64(msg.GetClientID()), uint64(msg.GetRequestID()), data); err != nil {
						return fmt.Errorf("couldn't send query 4 to client: %w", err)
					}
				case 5:
					slog.Debug("sending result", "query", 5, "clientId", msg.GetClientID())
					r.res.received |= query5Received
					slices.Sort(r.res.q5)
					data := []byte(strings.Join(r.res.q5, "\n"))
					if err := cprotocol.SendResultMsg(r.client, cprotocol.Query5, uint64(msg.GetClientID()), uint64(msg.GetRequestID()), data); err != nil {
						return fmt.Errorf("couldn't send query 5 to client: %w", err)
					}
				default:
					utils.Assertf(false, "query number %d should not happen in end", queryNumber)
				}
			} else {
				return fmt.Errorf("unexpected message type: %s", msg.GetMessageType())
			}
			if err := delivery.Ack(false); err != nil {
				slog.Error("acknowledge error", "error", err)
			}
			if r.res.received == allQuerysReceived {
				slog.Debug("all querys received")
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
