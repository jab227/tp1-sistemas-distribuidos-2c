package src

import (
	"iter"
	"os"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type FileLinesReader struct {
	file   *os.File
	reader *utils.LineReader
}

func NewFileLinesReader(filename string, nlines int) (*FileLinesReader, func(), error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}

	reader := utils.NewLinesReader(file, nlines)
	fileReader := &FileLinesReader{file: file, reader: reader}
	cleanup := func() {
		fileReader.deleteFileLinesReader()
	}
	return fileReader, cleanup, nil
}

func (flr *FileLinesReader) deleteFileLinesReader() error {
	return flr.file.Close()
}

func (flr *FileLinesReader) Text() string {
	return flr.reader.Text()
}

func (flr *FileLinesReader) Lines() iter.Seq2[[]string, error] {
	return flr.reader.Lines()
}
