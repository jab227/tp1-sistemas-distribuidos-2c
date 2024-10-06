package communication

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication/message"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication/payload"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication/utils"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/network"
)

type Protocol struct {
	socket         *network.SocketTcp
	syncAckMsgConf *message.SyncAckMessageConfig
}

func NewProtocol(socket *network.SocketTcp) *Protocol {
	return &Protocol{
		socket: socket,
	}
}

func (p *Protocol) Sync() error {
	if err := p.sendSyncMessage(); err != nil {
		return err
	}

	syncAckMessage, err := p.recvSyncAckMessage()
	if err != nil {
		return err
	}

	p.syncAckMsgConf = &message.SyncAckMessageConfig{
		ClientId:  syncAckMessage.Header.ClientId,
		RequestId: syncAckMessage.Header.RequestId,
	}
	return nil
}

func (p *Protocol) SyncAck(clientId uint32) error {
	_, err := p.recvSyncMessage()
	if err != nil {
		return err
	}

	syncAckMsgConf := &message.SyncAckMessageConfig{
		ClientId: clientId,
	}
	if err := p.sendSyncAckMessage(syncAckMsgConf); err != nil {
		return err
	}

	p.syncAckMsgConf = syncAckMsgConf
	return nil
}

func (p *Protocol) SendDataMessage(dataMsgConf *message.DataMessageConfig) error {
	dataMessage := message.NewDataMessage(dataMsgConf)
	dataMessage.Header.ClientId = p.syncAckMsgConf.ClientId
	err := sendMessage(p, dataMessage)
	return err
}

func (p *Protocol) RecvDataMessage() (*message.Message[*payload.Data], error) {
	dataMessage := message.NewDefaultDataMessage()
	if err := recvMessage(p, dataMessage); err != nil {
		return nil, err
	}
	return dataMessage, nil
}

func (p *Protocol) SendResultMessage(resultMsgConf *message.ResultMessageConfig) error {
	resultMessage := message.NewResultMessage(resultMsgConf)
	resultMessage.Header.ClientId = p.syncAckMsgConf.ClientId
	err := sendMessage(p, resultMessage)
	return err
}

func (p *Protocol) RecvResultMessage() (*message.Message[*payload.Result], error) {
	resultMessage := message.NewDefaultResultMessage()
	if err := recvMessage(p, resultMessage); err != nil {
		return nil, err
	}
	return resultMessage, nil
}

func (p *Protocol) sendSyncMessage() error {
	syncMessage := message.NewSyncMessage()
	err := sendMessage(p, syncMessage)
	return err
}

func (p *Protocol) recvSyncMessage() (*message.Message[*payload.Empty], error) {
	syncMessage := message.NewDefaultSyncMessage()
	if err := recvMessage(p, syncMessage); err != nil {
		return nil, err
	}
	return syncMessage, nil
}

func (p *Protocol) sendSyncAckMessage(syncAckMsgConf *message.SyncAckMessageConfig) error {
	syncAckMessage := message.NewSyncAckMessage(syncAckMsgConf)
	err := sendMessage(p, syncAckMessage)
	return err
}

func (p *Protocol) recvSyncAckMessage() (*message.Message[*payload.Empty], error) {
	syncAckMessage := message.NewDefaultSyncAckMessage()
	if err := recvMessage(p, syncAckMessage); err != nil {
		return nil, err
	}
	return syncAckMessage, nil
}

func sendMessage[T utils.Marshallable](context *Protocol, message *message.Message[T]) error {
	data := message.Marshall()
	return context.socket.Send(data)
}

func recvMessage[T utils.Marshallable](context *Protocol, message *message.Message[T]) error {
	if err := recvHeader(context, message); err != nil {
		return err
	}
	return recvPayload(context, message)
}

func recvHeader[T utils.Marshallable](context *Protocol, message *message.Message[T]) error {
	sizeOfHeader := message.SizeofHeader()
	headerData := make([]byte, sizeOfHeader)
	if err := context.socket.Receive(headerData); err != nil {
		return err
	}
	message.UnmarshallHeader(headerData)
	return nil
}

func recvPayload[T utils.Marshallable](context *Protocol, message *message.Message[T]) error {
	payloadSize := message.Header.PayloadSize
	payloadData := make([]byte, payloadSize)
	if err := context.socket.Receive(payloadData); err != nil {
		return err
	}
	message.UnmarshallPayload(payloadData)
	return nil
}
