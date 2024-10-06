package utils

import (
	"bufio"
	"io"
	"iter"
)

type LineReader struct {
	scanner *bufio.Scanner
	chunk   []string
	pos     int
	size    int
}

func NewLinesReader(r io.Reader, chunkSize int) *LineReader {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	return &LineReader{
		scanner: scanner,
		chunk:   make([]string, chunkSize),
		size:    chunkSize,
	}
}

func (c *LineReader) Text() string {
	c.scanner.Scan()
	return c.scanner.Text()
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

func (l *LineReader) Lines() iter.Seq2[[]string, error] {
	return func(yield func([]string, error) bool) {
		for {
			lines, more, err := l.Next()
			if !yield(lines, err) {
				return
			}
			if !more {
				return
			}
		}
	}
}
