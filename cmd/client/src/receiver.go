package src

import (
	"fmt"
	"log/slog"
	"os"
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
	var received int
	for received < 5 {
		result, err := r.protocol.RecvResultMessage()
		if err != nil {
			return err
		}

		if result.Payload.Header.Type == uint8(message.Query1) {
			query1 := strings.Split(string(result.Payload.Payload.Data), ",")
			windows := query1[0]
			mac := query1[1]
			linux := query1[2]
			fmt.Fprintf(os.Stdout, "Query 1:\n")
			fmt.Fprintf(os.Stdout, "===========\n")
			fmt.Fprintf(os.Stdout, "windows: %s\n", windows)
			fmt.Fprintf(os.Stdout, "mac: %s\n", mac)
			fmt.Fprintf(os.Stdout, "linux: %s\n", linux)
			received += 1
		} else if result.Payload.Header.Type == uint8(message.Query2) {
			query2 := strings.Split(string(result.Payload.Payload.Data), "\n")
			fmt.Fprintf(os.Stdout, "Query 2:\n")
			fmt.Fprintf(os.Stdout, "===========\n")
			for i, s := range query2 {
				fmt.Fprintf(os.Stdout, "%d: %s\n", i, s)
			}
			received += 1
		} else if result.Payload.Header.Type == uint8(message.Query3) {
			query3 := strings.Split(string(result.Payload.Payload.Data), "\n")
			fmt.Fprintf(os.Stdout, "Query 3:\n")
			fmt.Fprintf(os.Stdout, "===========\n")
			for i, s := range query3 {
				fmt.Fprintf(os.Stdout, "%d: %s\n", i, s)
			}
			received += 1
		} else if result.Payload.Header.Type == uint8(message.Query4) {
			query4 := strings.Split(string(result.Payload.Payload.Data), "\n")
			fmt.Fprintf(os.Stdout, "Query 4:\n")
			fmt.Fprintf(os.Stdout, "===========\n")
			for i, s := range query4 {
				fmt.Fprintf(os.Stdout, "%d: %s\n", i, s)
			}
			received += 1
		} else if result.Payload.Header.Type == uint8(message.Query5) {
			query5 := strings.Split(string(result.Payload.Payload.Data), "\n")
			fmt.Fprintf(os.Stdout, "Query 5:\n")
			fmt.Fprintf(os.Stdout, "===========\n")
			for i, s := range query5 {
				fmt.Fprintf(os.Stdout, "%d: %s\n", i, s)
			}
			received += 1
		} else {
			slog.Debug(fmt.Sprintf("Unknown query type: %d\n", result.Payload.Header.Type))
		}
	}
	return nil
}
