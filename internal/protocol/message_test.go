package protocol_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
)

func TestCreatingAnEndMessasge(t *testing.T) {
	msg := protocol.NewEndMessage(protocol.Games, protocol.MessageOptions{
		MessageID: 8,
		ClientID:  1,
		RequestID: 1,
	})
	if !msg.ExpectKind(protocol.End) {
		t.Error("expected message kind end")
	}

	if !msg.HasGameData() || msg.HasReviewData() {
		t.Error("expected message has game data")
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

	t.Run("create data message with multiple elements and iterate over fields", func(t *testing.T) {
		type payloadValues struct {
			b  byte
			bs []byte
			u  uint32
			f  float32
		}

		tts := []payloadValues{
			{252, []byte("hellope"), 424, 55.5},
			{69, []byte("elden ring"), 218, 4200.5},
			{42, []byte("borderlands 3"), 1024, -21.5},
		}
		buffer := protocol.NewPayloadBuffer(len(tts))
		for _, tt := range tts {
			buffer.BeginPayloadElement()
			{
				buffer.WriteByte(tt.b)
				buffer.WriteBytes(tt.bs)
				buffer.WriteUint32(tt.u)
				buffer.WriteFloat32(tt.f)
			}
			buffer.EndPayloadElement()
		}

		payload := buffer.Bytes()

		msg := protocol.NewDataMessage(protocol.Games, payload, protocol.MessageOptions{
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
		elements := msg.Elements()
		for i, element := range elements.Iter() {
			b := element.ReadByte()
			if b != tts[i].b {
				t.Errorf("expected %d got %d", tts[i].b, b)
			}

			str := string(element.ReadBytes())
			if str != string(tts[i].bs) {
				t.Errorf("expected %#v, got %#v", string(tts[i].bs), str)
			}

			u := element.ReadUint32()
			if u != tts[i].u {
				t.Errorf("expected %d got %d", tts[i].u, u)
			}

			f := element.ReadFloat32()
			if f != tts[i].f {
				t.Errorf("expected %f got %f", tts[i].f, f)
			}
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
