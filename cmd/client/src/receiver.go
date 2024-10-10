package src

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication/message"
)

type Receiver struct {
	clientConfig *ClientConfig
	protocol     *communication.Protocol
}

func NewReceiver(clientConfig *ClientConfig, protocol *communication.Protocol) *Receiver {
	return &Receiver{clientConfig: clientConfig, protocol: protocol}
}

func (r *Receiver) Run(join chan error) {
	if err := r.receive(); err != nil {
		join <- err
	}
	join <- nil
}

func (r *Receiver) receive() error {
	for {
		result, err := r.protocol.RecvResultMessage()
		if err != nil {
			return err
		}

		if result.Payload.Header.Type == uint8(message.Query1) {
			query1 := strings.Split(string(result.Payload.Payload.Data), ",")
			windows := query1[0]
			mac := query1[1]
			linux := query1[2]
			slog.Info(fmt.Sprintf("Query1: Windows %s, Mac %s, Linux %s\n", windows, linux, mac))
		} else if result.Payload.Header.Type == uint8(message.Query2) {
			query2 := strings.Split(string(result.Payload.Payload.Data), "\n")
			slog.Info(fmt.Sprintf("Query2: %s\n", query2))
		} else if result.Payload.Header.Type == uint8(message.Query3) {
			query3 := strings.Split(string(result.Payload.Payload.Data), "\n")
			slog.Info(fmt.Sprintf("Query3: %s\n", query3))
		} else if result.Payload.Header.Type == uint8(message.Query4) {
			query4 := strings.Split(string(result.Payload.Payload.Data), "\n")
			slog.Info(fmt.Sprintf("Query4: %s\n", query4))
		} else if result.Payload.Header.Type == uint8(message.Query5) {
			query5 := strings.Split(string(result.Payload.Payload.Data), "\n")
			slog.Info(fmt.Sprintf("Query5: %s\n", query5))
		} else {
			slog.Debug(fmt.Sprintf("Unknown query type: %d\n", result.Payload.Header.Type))
		}
	}
}
