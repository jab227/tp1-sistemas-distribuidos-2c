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

func (flr *FileLinesReader) Read() (string, bool, error) {
	sliceOfLines, more, err := flr.reader.Next()
	if len(sliceOfLines) == 0 {
		return "", more, err
	}

	separator := "\n"
	lines := strings.Join(sliceOfLines, separator) + separator
	return lines, more, err
}
