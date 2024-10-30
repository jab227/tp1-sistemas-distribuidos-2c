package client

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"iter"
	"os"
)

type FileBatcher struct {
	file       *os.File
	lineReader *utils.LineReader
	config     FileBatcherConfig
}

func NewFileBatcher(config FileBatcherConfig) (*FileBatcher, error) {
	file, err := os.Open(config.Path)
	if err != nil {
		return nil, err
	}

	reader := utils.NewLinesReader(file, config.NLinesFromDisk)
	reader.Text()

	return &FileBatcher{
		file:       file,
		lineReader: reader,
		config:     config,
	}, err
}

func (f *FileBatcher) Contents() iter.Seq[[]string] {
	return func(yield func([]string) bool) {
		var buffer []string
		var lastRead []string
		bytesCounted := 0

		for {
			line := f.lineReader.Text()
			if line == "" {
				if len(buffer) >= 1 {
					yield(buffer)
					return
				} else if len(lastRead) <= 0 {
					return
				} else {
					yield(lastRead)
					lastRead = nil
					return
				}
			}

			if len(lastRead) > 0 {
				buffer = append(buffer, lastRead...)
				bytesCounted += len(lastRead)
				lastRead = nil
			}

			if len(buffer) >= f.config.MaxElements || bytesCounted >= f.config.MaxBytes {
				if !yield(buffer) {
					lastRead = nil
					buffer = nil
					return
				}

				buffer = nil
				bytesCounted = 0
				lastRead = append(lastRead, line)
			} else {
				buffer = append(buffer, line)
				bytesCounted += len(line)
			}
		}
	}
}

func (f *FileBatcher) Close() error {
	return f.file.Close()
}
