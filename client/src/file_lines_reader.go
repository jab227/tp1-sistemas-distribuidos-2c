package src

import (
	"os"
	"strings"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type FileLinesReader struct {
	file   *os.File
	reader *utils.LineReader
}

func NewFileLinesReader(filename string) (*FileLinesReader, func(), error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}

	linesToRead := 1
	reader := utils.NewLinesReader(file, linesToRead)
	fileReader := &FileLinesReader{file: file, reader: reader}
	cleanup := func() {
		fileReader.deleteFileLinesReader()
	}
	return fileReader, cleanup, nil
}

func (flr *FileLinesReader) deleteFileLinesReader() error {
	return flr.file.Close()
}

func (flr *FileLinesReader) Read() (string, bool, error) {
	separator := "\n"
	sliceOfLines, more, err := flr.reader.Next()
	lines := strings.Join(sliceOfLines, separator) + separator
	return lines, more, err
}
