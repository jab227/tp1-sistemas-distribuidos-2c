package src

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/common/communication/message"
)

type BatchFileConfig struct {
	DataType       message.DataType
	Path           string
	NlinesFromDisk int
	BatchSize      int
	MaxBytes       int
}

type BatchFile struct {
	config       *BatchFileConfig
	reader       *FileLinesReader
	deleteReader func()
	taskQueue    *BlockingQueue[*message.DataMessageConfig]
}

func NewBatchFile(config *BatchFileConfig, taskQueue *BlockingQueue[*message.DataMessageConfig]) (*BatchFile, func(), error) {
	fileReader, deleteFileLinesReader, err := NewFileLinesReader(config.Path, config.NlinesFromDisk)
	if err != nil {
		return nil, nil, err
	}

	batchFile := &BatchFile{
		reader:       fileReader,
		deleteReader: deleteFileLinesReader,
		taskQueue:    taskQueue,
		config:       config,
	}
	cleanup := func() { deleteBatchFile(batchFile) }
	return batchFile, cleanup, nil
}

func deleteBatchFile(bf *BatchFile) {
	bf.deleteReader()
}

func (bf *BatchFile) Run(join chan error) {
	bf.pushStart()
	bf.pushDataMessages()
	bf.pushEnd()
	join <- nil
}

func (bf *BatchFile) pushStart() {
	dataConfig := &message.DataMessageConfig{
		Start:    true,
		DataType: bf.config.DataType,
	}
	bf.taskQueue.Push(dataConfig)
}

func (bf *BatchFile) pushDataMessages() {
	lines, more, _ := bf.reader.Read()
	batchLines := NewBatchLines(lines, bf.config.BatchSize, bf.config.MaxBytes)

	for {
		callback := func(data string) {
			bf.taskQueue.Push(&message.DataMessageConfig{
				DataType: bf.config.DataType,
				Data:     []byte(data),
			})
		}

		batchLines.Execute(callback)
		if !more {
			break
		}
		lines, more, _ = bf.reader.Read()
		batchLines.Reset(lines)
	}
}

func (bf *BatchFile) pushEnd() {
	dataConfig := &message.DataMessageConfig{
		End:      true,
		DataType: bf.config.DataType,
	}
	bf.taskQueue.Push(dataConfig)
}
