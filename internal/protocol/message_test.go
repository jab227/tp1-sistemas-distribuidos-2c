package protocol_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
)

func TestCreatingAnEndMessasge(t *testing.T) {
	msg := protocol.NewEndMessage(protocol.MessageOptions{
		MessageID: 8,
		ClientID:  1,
		RequestID: 1,
	})
	if !msg.ExpectKind(protocol.End) {
		t.Error("expected message kind end")
	}
}

func TestCreatingADataMessage(t *testing.T) {
	t.Run("create games data message", func(t *testing.T) {
		msg := protocol.NewDataMessage(protocol.Games, []byte("elden ring"), protocol.MessageOptions{
			MessageID: 8,
			ClientID:  1,
			RequestID: 1,
		})

		if !msg.ExpectKind(protocol.Data) {
			t.Error("expected message kind data")
		}

		if !msg.HasGameData() || msg.HasReviewData() {
			t.Error("expected game data")
		}
	})

	t.Run("create reviews data message", func(t *testing.T) {
		msg := protocol.NewDataMessage(protocol.Reviews, []byte("elden ring"), protocol.MessageOptions{
			MessageID: 8,
			ClientID:  1,
			RequestID: 1,
		})

		if !msg.ExpectKind(protocol.Data) {
			t.Error("expected message kind data")
		}

		if !msg.HasReviewData() || msg.HasGameData() {
			t.Error("expected reviews data")
		}
	})
}

func TestCreatingAResultsessage(t *testing.T) {
	for i, query := range []protocol.QueryNumber{protocol.Query1, protocol.Query2, protocol.Query3, protocol.Query4, protocol.Query5} {
		testName := fmt.Sprintf("create games query %d", i+1)
		t.Run(testName, func(t *testing.T) {
			msg := protocol.NewResultsMessage(query, []byte("elden ring results"), protocol.MessageOptions{
				MessageID: 8,
				ClientID:  1,
				RequestID: 1,
			})

			if !msg.ExpectKind(protocol.Results) {
				t.Error("expected message kind results")
			}
			want := i + 1
			got := msg.GetQueryNumber()
			if got != want {
				t.Errorf("got query number %d, want %d", got, want)
			}
		})
	}
}

func TestMarshalAndUnmarshalOfMessage(t *testing.T) {
	msg := protocol.NewDataMessage(protocol.Games, []byte("elden ring"), protocol.MessageOptions{
		MessageID: 8,
		ClientID:  1,
		RequestID: 1,
	})

	data := msg.Marshal()
	var unmarshaledMsg protocol.Message
	err := unmarshaledMsg.Unmarshal(data)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if !reflect.DeepEqual(msg, unmarshaledMsg) {
		t.Errorf("got %#v, want %#v", unmarshaledMsg, msg)
	}
}
