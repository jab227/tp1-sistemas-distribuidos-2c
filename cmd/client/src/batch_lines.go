package src

import (
	"strings"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

type Callback func(string)

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

func (bl *BatchLines) Execute(callback Callback) error {
	bl.callback = callback

	for lines, err := range bl.lineReader.Lines() {
		if err != nil {
			return err
		}
		bl.processLines(lines)
	}
	bl.Flush()
	return nil
}

func (bl *BatchLines) Flush() {
	awaitingLines := bl.awaitingLines
	if len(awaitingLines) > 0 {
		bl.processData(awaitingLines)
		bl.resetAwaitingState()
	}
}

func (bl *BatchLines) processData(data string) {
	bl.callback(data)
	bl.processedBytes += len(data)
}

func (bl *BatchLines) resetAwaitingState() {
	bl.awaitingLines = ""
	bl.awaitingBytesCount = 0
	bl.awaitingLinesCount = 0
}

func (bl *BatchLines) processLines(sliceOfLines []string) {
	currentLines := bl.newLinesJoin(sliceOfLines)
	currentBytesCount := len(currentLines)
	currentLinesCount := len(sliceOfLines)

	newCurrentBytesCount := bl.awaitingBytesCount + currentBytesCount
	newCurrentLinesCount := bl.awaitingLinesCount + currentLinesCount
	if newCurrentBytesCount > bl.maxBytes || newCurrentLinesCount > bl.maxLines {
		bl.Flush()
	}

	if currentBytesCount > bl.maxBytes {
		bl.splitLinesProcessing(sliceOfLines, currentLines)
		return
	}

	bl.awaitingLines += currentLines
	bl.awaitingBytesCount += currentBytesCount
	bl.awaitingLinesCount += currentLinesCount
}

func (bl *BatchLines) splitLinesProcessing(sliceOfLines []string, currentLines string) {
	currentLinesCount := len(sliceOfLines)
	if currentLinesCount == 1 {
		bl.processData(currentLines)
		return
	}

	middle := currentLinesCount / 2
	bl.processLines(sliceOfLines[:middle])
	bl.processLines(sliceOfLines[middle:])
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
