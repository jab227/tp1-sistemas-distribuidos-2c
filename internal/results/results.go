package results

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"slices"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/middlewares/client"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type query1 struct {
	windows uint32
	mac     uint32
	linux   uint32
}

func (q query1) Marshal() []byte {
	buffer := protocol.NewPayloadBuffer(1)
	buffer.BeginPayloadElement()
	buffer.WriteByte(1)
	buffer.WriteUint32(q.windows)
	buffer.WriteUint32(q.mac)
	buffer.WriteUint32(q.linux)
	buffer.EndPayloadElement()
	return buffer.Bytes()
}

type query2 [10]string

func (q query2) Marshal() []byte {
	buffer := protocol.NewPayloadBuffer(1)
	buffer.BeginPayloadElement()
	buffer.WriteByte(2)
	for _, s := range q {
		buffer.WriteBytes([]byte(s))
	}
	buffer.EndPayloadElement()
	return buffer.Bytes()
}

type query3 [5]string

func (q query3) Marshal() []byte {
	buffer := protocol.NewPayloadBuffer(1)
	buffer.BeginPayloadElement()
	buffer.WriteByte(3)
	for _, s := range q {
		buffer.WriteBytes([]byte(s))
	}
	buffer.EndPayloadElement()
	return buffer.Bytes()
}

type query4 []string

func (q query4) Marshal() []byte {
	buffer := protocol.NewPayloadBuffer(1)
	buffer.BeginPayloadElement()
	buffer.WriteByte(4)
	for _, s := range q {
		buffer.WriteBytes([]byte(s))
	}
	buffer.EndPayloadElement()
	return buffer.Bytes()
}

type query5 []string

func (q query5) Marshal() []byte {
	buffer := protocol.NewPayloadBuffer(1)
	buffer.BeginPayloadElement()
	buffer.WriteByte(5)
	for _, s := range q {
		buffer.WriteBytes([]byte(s))
	}
	buffer.EndPayloadElement()
	return buffer.Bytes()
}

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
	io     client.IOManager
	done   chan struct{}
	res    *results
}

// I don't own the connection
func NewResultsService(conn net.Conn, io client.IOManager) *ResultsService {
	return &ResultsService{
		client: conn,
		io:     io,
		done:   make(chan struct{}),
		res:    &results{},
	}
}

func (r *ResultsService) Destroy() {
	// r.io.Close()
}

func (r *ResultsService) Done() <-chan struct{} {
	return r.done
}

func (r *ResultsService) Run(ctx context.Context) error {
	consumerCh := r.io.Input.GetConsumer()
	defer func() {
		r.done <- struct{}{}
	}()

	writer := bufio.NewWriter(r.client)
	defer writer.Flush()
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
					for _, element := range elements.Iter() {
						r.res.q1 = query1{
							windows: element.ReadUint32(),
							mac:     element.ReadUint32(),
							linux:   element.ReadUint32(),
						}
					}
					r.res.received |= query1Received
					_, err := writer.Write(r.res.q1.Marshal())
					if err != nil {
						return fmt.Errorf("couldn't write query 1: %w", err)
					}
				case 2:
					for _, element := range elements.Iter() {
						var q2 [10]string
						for i := 0; i < len(q2); i++ {
							q2[i] = string(element.ReadBytes())
						}
						r.res.q2 = q2
					}
					r.res.received |= query2Received
					_, err := writer.Write(r.res.q2.Marshal())
					if err != nil {
						return fmt.Errorf("couldn't write query 2: %w", err)
					}
				case 3:
					for _, element := range elements.Iter() {
						var q3 [5]string
						for i := 0; i < len(q3); i++ {
							q3[i] = string(element.ReadBytes())
						}
						r.res.q3 = q3
					}
					r.res.received |= query3Received
					_, err := writer.Write(r.res.q3.Marshal())
					if err != nil {
						return fmt.Errorf("couldn't write query 3: %w", err)
					}
				case 4:
					for _, element := range elements.Iter() {
						r.res.q4 = append(r.res.q4, string(element.ReadBytes()))
					}
				case 5:
					for _, element := range elements.Iter() {
						r.res.q5 = append(r.res.q5, string(element.ReadBytes()))
					}
				default:
					utils.Assertf(false, "query number %d should not happen", queryNumber)
				}
			} else if msg.ExpectKind(protocol.End) {
				queryNumber := msg.GetQueryNumber()
				switch queryNumber {
				case 4:
					r.res.received |= query4Received
					slices.Sort(r.res.q4)
					_, err := writer.Write(r.res.q4.Marshal())
					if err != nil {
						return fmt.Errorf("couldn't write query 4: %w", err)
					}
				case 5:
					r.res.received |= query5Received
					slices.Sort(r.res.q5)
					_, err := writer.Write(r.res.q5.Marshal())
					if err != nil {
						return fmt.Errorf("couldn't write query 5: %w", err)
					}
				default:
					utils.Assertf(false, "query number %d should not happen in end", queryNumber)
				}
			} else {
				return fmt.Errorf("unexpected message type: %s", msg.GetMessageType())
			}

			if r.res.received == allQuerysReceived {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
