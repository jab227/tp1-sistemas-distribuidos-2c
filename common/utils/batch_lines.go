package utils

import (
	"strings"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type Callback func(string) error

type BatchLines struct {
	lines              string
	processedBytes     int
	reader             *strings.Reader
	lineReader         *utils.LineReader
	maxBytes           int
	maxLines           int
	awaitingLines      string
	awaitingBytesCount int
	awaitingLinesCount int
	callback           Callback
}

func NewBatchLines(lines string, size int, maxBytes int) *BatchLines {
	stringReader := strings.NewReader(lines)
	batch := &BatchLines{
		lines:      lines,
		reader:     stringReader,
		lineReader: utils.NewLinesReader(stringReader, size),
		maxBytes:   maxBytes,
		maxLines:   size,
	}
	return batch
}

func (bl *BatchLines) Run(callback Callback) error {
	bl.callback = callback
	for {
		sliceOfLines, more, err := bl.lineReader.Next()
		if err != nil {
			return err
		}
		if err := bl.processLines(sliceOfLines); err != nil {
			return err
		}

		if !more {
			return bl.Flush()
		}
	}
}

func (bl *BatchLines) Flush() error {
	awaitingLines := bl.awaitingLines
	if len(awaitingLines) > 0 {
		if err := bl.processData(awaitingLines); err != nil {
			return err
		}
		bl.resetAwaitingState()
	}
	return nil
}

func (bl *BatchLines) processData(data string) error {
	if err := bl.callback(data); err != nil {
		return err
	}
	bl.processedBytes += len(data)
	return nil
}

func (bl *BatchLines) resetAwaitingState() {
	bl.awaitingLines = ""
	bl.awaitingBytesCount = 0
	bl.awaitingLinesCount = 0
}

func (bl *BatchLines) processLines(sliceOfLines []string) error {
	currentLines := bl.newLinesJoin(sliceOfLines)
	currentBytesCount := len(currentLines)
	currentLinesCount := len(sliceOfLines)
	newCurrentBytesCount := bl.awaitingBytesCount + currentBytesCount
	newCurrentLinesCount := bl.awaitingLinesCount + currentLinesCount
	if newCurrentBytesCount > bl.maxBytes || newCurrentLinesCount > bl.maxLines {
		if err := bl.Flush(); err != nil {
			return err
		}
	}

	if currentBytesCount > bl.maxBytes {
		return bl.splitLinesProcessing(sliceOfLines, currentLines)
	}

	bl.awaitingLines += currentLines
	bl.awaitingBytesCount += currentBytesCount
	bl.awaitingLinesCount += currentLinesCount
	return nil
}

func (bl *BatchLines) splitLinesProcessing(sliceOfLines []string, currentLines string) error {
	currentLinesCount := len(sliceOfLines)
	if currentLinesCount == 1 {
		return bl.processData(currentLines)
	}

	middle := currentLinesCount / 2
	if err := bl.processLines(sliceOfLines[:middle]); err != nil {
		return err
	}
	return bl.processLines(sliceOfLines[middle:])
}

func (bl *BatchLines) newLinesJoin(sliceOfLines []string) string {
	if len(sliceOfLines) == 0 {
		return ""
	}

	separator := "\n"
	lines := strings.Join(sliceOfLines, separator) + separator
	return lines
}

func (bl *BatchLines) Reset(s string) {
	bl.lines = s
	bl.reader.Reset(s)
	bl.lineReader = utils.NewLinesReader(bl.reader, bl.maxLines)
	bl.processedBytes = 0
	bl.resetAwaitingState()
}

func (bl *BatchLines) ProcessedBytes() int {
	return bl.processedBytes
}
