package utils

import (
	"errors"
	"io"
)

// Reads at least len(p) bytes
func ReadAtLeast(r io.Reader, p []byte) error {
	read := 0
	size := len(p)
	for read < size {
		n, err := r.Read(p[read:])
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			if read < len(p) {
				return io.ErrUnexpectedEOF
			}
			break
		}
		read += n
	}
	return nil
}

// Writes at least len(p) bytes
func WriteAll(w io.Writer, p []byte) error {
	written := 0
	for written < len(p) {
		n, err := w.Write(p[written:])
		if err != nil {
			if !errors.Is(err, io.ErrShortWrite) {
				return err
			}
		}
		written += n
	}
	return nil
}
