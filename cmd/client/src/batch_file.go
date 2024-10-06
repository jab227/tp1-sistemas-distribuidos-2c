package src

import (
	"strings"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/communication/message"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type BatchFileConfig struct {
	DataType       message.DataType `json:"DataType"`
	Path           string           `json:"Path"`
	NlinesFromDisk int              `json:"NlinesFromDisk"`
	BatchSize      int              `json:"BatchSize"`
	MaxBytes       int              `json:"MaxBytes"`
}

type BatchFile struct {
	config       *BatchFileConfig
	fileReader   *FileLinesReader
	deleteReader func()
	taskQueue    *utils.BlockingQueue[*message.DataMessageConfig]
}

func NewBatchFile(config *BatchFileConfig, taskQueue *utils.BlockingQueue[*message.DataMessageConfig]) (*BatchFile, func(), error) {
	fileReader, deleteFileLinesReader, err := NewFileLinesReader(config.Path, config.NlinesFromDisk)
	if err != nil {
		return nil, nil, err
	}

	batchFile := &BatchFile{
		fileReader:   fileReader,
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
	if err := bf.pushDataMessages(); err != nil {
		join <- err
		return
	}
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

func (bf *BatchFile) pushDataMessages() error {
	batchLines := NewBatchLines("", bf.config.BatchSize, bf.config.MaxBytes)
	callback := func(data string) {
		bf.taskQueue.Push(&message.DataMessageConfig{
			DataType: bf.config.DataType,
			Data:     []byte(data),
		})
	}

	bf.fileReader.Text()
	for sliceOfLines, err := range bf.fileReader.Lines() {
		if err != nil {
			return err
		}

		lines := strings.Join(sliceOfLines, "\n") + "\n"
		batchLines.Reset(lines)
		if err := batchLines.Execute(callback); err != nil {
			return err
		}
	}
	return nil
}

func (bf *BatchFile) pushEnd() {
	dataConfig := &message.DataMessageConfig{
		End:      true,
		DataType: bf.config.DataType,
	}
	bf.taskQueue.Push(dataConfig)
}
