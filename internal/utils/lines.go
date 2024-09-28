package utils

import (
	"bufio"
	"io"
)

type LineReader struct {
	scanner *bufio.Scanner
	chunk   []string
	pos     int
	size    int
}

func NewLinesReader(r io.Reader, chunkSize int) *LineReader {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 24*1024)
	scanner.Buffer(buf, 128*1024)
	return &LineReader{
		scanner: scanner,
		chunk:   make([]string, chunkSize),
		size:    chunkSize,
	}
}

func (c *LineReader) Next() ([]string, bool, error) {
	c.pos = 0	
	more := true
	for c.pos < c.size {
		ok := c.scanner.Scan()
		if !ok {
			more = false
			break
		}
		c.chunk[c.pos] = c.scanner.Text()
		c.pos++
	}
	return c.chunk[:c.pos], more, c.scanner.Err()	
}
