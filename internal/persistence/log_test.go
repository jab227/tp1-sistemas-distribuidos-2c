package persistence_test

import (
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/persistence"
)

func TestAppendAndCommitDataToWriteAheadLog(t *testing.T) {
	tlog := persistence.NewTransactionLog("wal.log")
	tlog.Append([]byte("hellope"), 1)
	tlog.Commit()
	tlog.Append([]byte("hellope!"), 2)
	tlog.Commit()
	tlog.Append([]byte("helloworld!"), 3)
	tlog.Commit()
	f, err := os.Open("wal.log")
	if err != nil {
		t.Fatal(err)
	}
	p, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	var got persistence.TransactionLog
	err = got.Unmarshal(p)
	if err != nil {
		t.Fatal(err)
	}

	want := tlog.GetLog()

	if !reflect.DeepEqual(got.GetLog(), want) {
		t.Errorf("got %v, want %v", got.GetLog(), want)
	}

	os.Remove("./wal.log")
}

func BenchmarkAppendAndCommit16KDataChunks(b *testing.B) {
	b.ReportAllocs()
	wal := persistence.NewTransactionLog("wal.log")
	data := make([]byte, 16*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wal.Append(data, 1)
		wal.Commit()
	}
	os.Remove("wal.log")
}

func BenchmarkAppendAndCommitAHundred16KDataChunks(b *testing.B) {
	b.ReportAllocs()
	wal := persistence.NewTransactionLog("wal.log")
	data := make([]byte, 16*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			wal.Append(data, 1)
			wal.Commit()
		}
	}
	os.Remove("wal.log")
}
