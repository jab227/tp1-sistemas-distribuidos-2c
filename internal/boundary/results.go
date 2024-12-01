package boundary

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/cprotocol"

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

type Results struct {
	q1       query1
	q2       query2
	q3       query3
	q4       query4
	q5       query5
	received receivedQuerys
}

type ResultsState struct {
	res map[uint32]Results
}

func (rs *ResultsState) GetClientResults(client uint32) Results {
	return rs.res[client]
}

func (rs *ResultsState) hasClientAllResults(client uint32) bool {
	return rs.res[client].received == allQuerysReceived
}

func (rs *ResultsState) AddReceived(clientId uint32, query receivedQuerys) {
	clientResults := rs.res[clientId]
	clientResults.received |= query
	rs.res[clientId] = clientResults
}

func (rs *ResultsState) AddQuery1(clientId uint32, query query1) {
	clientResults := rs.res[clientId]
	clientResults.q1 = query
	rs.res[clientId] = clientResults
}

func (rs *ResultsState) AddQuery2(clientId uint32, query string) {
	clientResults := rs.res[clientId]
	clientResults.q2 = append(clientResults.q2, query)
	rs.res[clientId] = clientResults
}

func (rs *ResultsState) AddQuery3(clientId uint32, query string) {
	clientResults := rs.res[clientId]
	clientResults.q3 = append(clientResults.q3, query)
	rs.res[clientId] = clientResults
}

func (rs *ResultsState) AddQuery4(clientId uint32, query string) {
	clientResults := rs.res[clientId]
	clientResults.q4 = append(clientResults.q4, query)
	rs.res[clientId] = clientResults
}

func (rs *ResultsState) AddQuery5(clientId uint32, query string) {
	clientResults := rs.res[clientId]
	clientResults.q5 = append(clientResults.q5, query)
	rs.res[clientId] = clientResults
}

type ResultsService struct {
	io            *client.IOManager
	done          chan struct{}
	res           *ResultsState
	boundaryState *State
}

// I don't own the connection
func NewResultsService(io *client.IOManager, boundaryState *State) *ResultsService {
	return &ResultsService{
		io:            io,
		done:          make(chan struct{}),
		res:           &ResultsState{res: make(map[uint32]Results)},
		boundaryState: boundaryState,
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
			slog.Debug("received message", "clientID", msg.GetClientID())
			if msg.ExpectKind(protocol.Results) {
				queryNumber := msg.GetQueryNumber()
				slog.Debug("received query result", "query number", queryNumber)
				elements := msg.Elements()
				switch queryNumber {
				case 1:
					slog.Debug("received result", "query", 1, "clientId", msg.GetClientID())
					for _, element := range elements.Iter() {
						r.res.AddQuery1(msg.GetClientID(), query1{
							windows: element.ReadUint32(),
							mac:     element.ReadUint32(),
							linux:   element.ReadUint32(),
						})
					}
					r.res.AddReceived(msg.GetClientID(), query1Received)

					// Creates string representation
					slog.Debug("sending result", "query", 1, "clientId", msg.GetClientID())
					clientResults := r.res.GetClientResults(msg.GetClientID())
					data := []byte(fmt.Sprintf("%d,%d,%d", clientResults.q1.windows, clientResults.q1.mac, clientResults.q1.linux))

					// Sends message
					clientMsg := *cprotocol.NewResultMessage(cprotocol.Query1, uint64(msg.GetClientID()), uint64(msg.GetRequestID()), data)
					clientCh := r.boundaryState.GetClientCh(uint64(msg.GetClientID()))
					if clientCh == nil {
						return fmt.Errorf("chan is nil for client %d", msg.GetClientID())
					}
					safeChannelWrite(clientCh, clientMsg)
				case 2:
					slog.Debug("received result", "query", 2, "clientId", msg.GetClientID())
					for _, element := range elements.Iter() {
						r.res.AddQuery2(msg.GetClientID(), string(element.ReadBytes()))
					}
				case 3:
					slog.Debug("received result", "query", 3, "clientId", msg.GetClientID())
					for _, element := range elements.Iter() {
						r.res.AddQuery3(msg.GetClientID(), string(element.ReadBytes()))
					}
				case 4:
					slog.Debug("received result", "query", 4, "clientId", msg.GetClientID())
					for _, element := range elements.Iter() {
						r.res.AddQuery4(msg.GetClientID(), string(element.ReadBytes()))
					}
				case 5:
					slog.Debug("received result", "query", 5, "clientId", msg.GetClientID())
					for _, element := range elements.Iter() {
						r.res.AddQuery5(msg.GetClientID(), string(element.ReadBytes()))
					}
				default:
					utils.Assertf(false, "query number %d should not happen", queryNumber)
				}
			} else if msg.ExpectKind(protocol.End) {
				queryNumber := msg.GetQueryNumber()
				switch queryNumber {
				case 2:
					// Update results state
					slog.Debug("sending result", "query", 2, "clientId", msg.GetClientID())
					r.res.AddReceived(msg.GetClientID(), query2Received)

					// Get string representation
					clientResults := r.res.GetClientResults(msg.GetClientID())
					data := []byte(strings.Join(clientResults.q2[:], "\n"))

					// Sends results
					clientMsg := *cprotocol.NewResultMessage(cprotocol.Query2, uint64(msg.GetClientID()), uint64(msg.GetRequestID()), data)
					clientCh := r.boundaryState.GetClientCh(uint64(msg.GetClientID()))
					if clientCh == nil {
						return fmt.Errorf("chan is nil for client %d", msg.GetClientID())
					}
					safeChannelWrite(clientCh, clientMsg)
				case 3:
					// Update results state
					slog.Debug("sending result", "query", 3, "clientId", msg.GetClientID())
					r.res.AddReceived(msg.GetClientID(), query3Received)

					// Get string representation
					clientResults := r.res.GetClientResults(msg.GetClientID())
					data := []byte(strings.Join(clientResults.q3[:], "\n"))

					// Send results
					clientMsg := *cprotocol.NewResultMessage(cprotocol.Query3, uint64(msg.GetClientID()), uint64(msg.GetRequestID()), data)
					clientCh := r.boundaryState.GetClientCh(uint64(msg.GetClientID()))
					if clientCh == nil {
						return fmt.Errorf("chan is nil for client %d", msg.GetClientID())
					}
					safeChannelWrite(clientCh, clientMsg)
				case 4:
					// Update results state
					slog.Debug("sending result", "query", 4, "clientId", msg.GetClientID())
					r.res.AddReceived(msg.GetClientID(), query4Received)

					// Get string representation
					clientResults := r.res.GetClientResults(msg.GetClientID())
					slices.Sort(clientResults.q4)
					data := []byte(strings.Join(clientResults.q4, "\n"))

					// Send results
					clientMsg := *cprotocol.NewResultMessage(cprotocol.Query4, uint64(msg.GetClientID()), uint64(msg.GetRequestID()), data)
					clientCh := r.boundaryState.GetClientCh(uint64(msg.GetClientID()))
					if clientCh == nil {
						return fmt.Errorf("chan is nil for client %d", msg.GetClientID())
					}
					safeChannelWrite(clientCh, clientMsg)
				case 5:
					// Update results state
					slog.Debug("sending result", "query", 5, "clientId", msg.GetClientID())
					r.res.AddReceived(msg.GetClientID(), query5Received)

					// Get string representation
					clientResults := r.res.GetClientResults(msg.GetClientID())
					slices.Sort(clientResults.q5)
					data := []byte(strings.Join(clientResults.q5, "\n"))

					// Send results
					clientMsg := *cprotocol.NewResultMessage(cprotocol.Query5, uint64(msg.GetClientID()), uint64(msg.GetRequestID()), data)
					clientCh := r.boundaryState.GetClientCh(uint64(msg.GetClientID()))
					if clientCh == nil {
						return fmt.Errorf("chan is nil for client %d", msg.GetClientID())
					}
					safeChannelWrite(clientCh, clientMsg)
				default:
					utils.Assertf(false, "query number %d should not happen in end", queryNumber)
				}
			} else {
				return fmt.Errorf("unexpected message type: %s", msg.GetMessageType())
			}
			if err := delivery.Ack(false); err != nil {
				slog.Error("acknowledge error", "error", err)
			}
			if r.res.hasClientAllResults(msg.GetClientID()) {
				slog.Debug("all querys received")
				close(r.boundaryState.GetClientCh(uint64(msg.GetClientID())))
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func safeChannelWrite(ch chan cprotocol.Message, value cprotocol.Message) {
	defer func() {
		if r := recover(); r != nil {
			slog.Debug("Writing to a closed channel. Client has already received all results")
		}
	}()
	ch <- value
}
