package src

import (
	"fmt"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication"
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
	// TODO: Aplicar logica cuando se obtengan resultados reales si es necesario
	for {
		result, err := r.protocol.RecvResultMessage()
		if err != nil {
			return err
		}
		fmt.Println(string(result.Payload.Payload.Data))
	}
}
