package src

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication/message"
)

type DataMessageSender struct {
	protocol  *communication.Protocol
	taskQueue *BlockingQueue[*message.DataMessageConfig]
}

func NewFileSender(protocol *communication.Protocol, taskQueue *BlockingQueue[*message.DataMessageConfig]) *DataMessageSender {
	return &DataMessageSender{
		protocol:  protocol,
		taskQueue: taskQueue,
	}
}

func (fs *DataMessageSender) Run(join chan error) {
	for dataConfig := range fs.taskQueue.Iter() {
		if err := fs.protocol.SendDataMessage(dataConfig); err != nil {
			join <- err
			return
		}
	}
	join <- nil
}
