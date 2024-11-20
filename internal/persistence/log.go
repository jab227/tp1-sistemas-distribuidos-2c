package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Entry struct {
	data      []byte
	timestamp int64
	index     int64
	checksum  uint32
	kind      uint32
}

type TransactionLog struct {
	log      []Entry
	filename string
	lock     *sync.Mutex
}

func NewTransactionLog(filename string) *TransactionLog {
	var log TransactionLog
	log.filename = filename
	log.lock = new(sync.Mutex)
	return &log
}

func (t *TransactionLog) Append(data []byte, kind uint32) {
	crc := crc32.ChecksumIEEE(data)
	timestamp := time.Now().UnixNano()

	t.lock.Lock()
	defer t.lock.Unlock()

	index := len(t.log)
	t.log = append(t.log, Entry{
		data:      data,
		timestamp: timestamp,
		index:     int64(index),
		checksum:  crc,
		kind:      kind,
	})
}

func (t *TransactionLog) Unmarshal(p []byte) error {
	entriesCount := binary.LittleEndian.Uint64(p[0:8])
	log := make([]Entry, entriesCount)
	p = p[8:]
	i := 0
	for len(p) > 0 {
		entry, n, err := readEntry(p)
		if err != nil {
			return err
		}
		log[i] = entry
		p = p[n:]
		i += 1
	}
	t.log = log
	return nil
}

func readEntry(p []byte) (Entry, int, error) {
	crc := binary.LittleEndian.Uint32(p[0:4])
	kind := binary.LittleEndian.Uint32(p[4:8])
	index := binary.LittleEndian.Uint64(p[8:16])
	timestamp := binary.LittleEndian.Uint64(p[16:24])
	dataLen := binary.LittleEndian.Uint64(p[24:32])
	data := p[32:][:dataLen]
	if crc != crc32.ChecksumIEEE(data) {
		return Entry{}, -1, fmt.Errorf("couldn't unmarshal log: corrupted entry")
	}
	return Entry{
		data:      data,
		timestamp: int64(timestamp),
		index:     int64(index),
		checksum:  crc,
		kind:      kind,
	}, int(32 + dataLen), nil
}

func (t *TransactionLog) Marshal() []byte {
	buf4 := make([]byte, 4)
	buf8 := make([]byte, 8)
	var buf bytes.Buffer
	binary.LittleEndian.PutUint64(buf8, uint64(len(t.log)))
	buf.Write(buf8)
	for _, entry := range t.log {
		writeEntry(&buf, entry, buf4, buf8)
	}
	return buf.Bytes()
}

func writeEntry(w io.Writer, entry Entry, buf4 []byte, buf8 []byte) {
	binary.LittleEndian.PutUint32(buf4, entry.checksum)
	w.Write(buf4)
	binary.LittleEndian.PutUint32(buf4, entry.kind)
	w.Write(buf4)
	binary.LittleEndian.PutUint64(buf8, uint64(entry.index))
	w.Write(buf8)
	binary.LittleEndian.PutUint64(buf8, uint64(entry.timestamp))
	w.Write(buf8)
	binary.LittleEndian.PutUint64(buf8, uint64(len(entry.data)))
	w.Write(buf8)
	w.Write(entry.data)
}

// Recepcion de los datos
// Se decide y se registra la transaccion (APPEND)
// COMMIT
// Acknowledge
func (w *TransactionLog) Commit() error {
	w.lock.Lock()
	defer w.lock.Unlock()
	return atomicWriteFile(w.filename, w.Marshal())
}

func atomicWriteFile(filename string, data []byte) (err error) {
	f, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename)+".tmp")
	if err != nil {
		return
	}

	tempName := f.Name()
	defer func() {
		if err != nil {
			f.Close()
			os.Remove(tempName)
		}
	}()
	
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("couldn't write to temp file: %w", err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("couldn't sync file to disk: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("couldn't close file: %w", err)
	}

	if err := os.Rename(tempName, filename); err != nil {
		return fmt.Errorf("couldn't rename file: %w", err)
	}
	return nil
}

func (w *TransactionLog) GetLog() []Entry {
	return w.log
}
