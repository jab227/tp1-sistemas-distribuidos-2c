package protocol_test

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"testing"
)

func TestReadWritePayloadWithSingleElement(t *testing.T) {
	buffer := protocol.NewPayloadBuffer(1)

	const (
		byteToWrite    = 252
		bytesToWrite   = "hellope"
		uint32ToWrite  = 42
		float32ToWrite = 55.5
		expectedN      = 1
	)

	buffer.WriteByte(byteToWrite)
	buffer.WriteBytes([]byte(bytesToWrite))
	buffer.WriteUint32(uint32ToWrite)
	buffer.WriteFloat32(float32ToWrite)
	buffer.EndPayloadElement()

	payload := buffer.Bytes()
	p, n := protocol.NewPayload(payload)
	if n != expectedN {
		t.Errorf("expected %d got %d", expectedN, n)
	}

	element, ok := p.NextElement()
	if !ok {
		t.Errorf("expected ok")
	}
	b := element.ReadByte()
	if b != byteToWrite {
		t.Errorf("expected %d got %d", byteToWrite, b)
	}

	str := string(element.ReadBytes())
	if str != bytesToWrite {
		t.Errorf("expected %#v, got %#v", bytesToWrite, str)
	}

	u := element.ReadUint32()
	if u != uint32ToWrite {
		t.Errorf("expected %d got %d", uint32ToWrite, u)
	}

	f := element.ReadFloat32()
	if f != float32ToWrite {
		t.Errorf("expected %f got %f", float32ToWrite, f)
	}
	element, ok = p.NextElement()
	if ok {
		t.Errorf("expected not ok")
	}
}

func TestReadWritePayloadWithMultipleElements(t *testing.T) {

	const (
		byteToWrite    = 252
		bytesToWrite   = "hellope"
		uint32ToWrite  = 42
		float32ToWrite = 55.5
		expectedN      = 1
	)

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
		buffer.WriteByte(tt.b)
		buffer.WriteBytes(tt.bs)
		buffer.WriteUint32(tt.u)
		buffer.WriteFloat32(tt.f)
		buffer.EndPayloadElement()
	}

	payload := buffer.Bytes()
	p, n := protocol.NewPayload(payload)
	if n != len(tts) {
		t.Errorf("expected %d got %d", expectedN, n)
	}
	for i := 0; i < len(tts); i++ {
		element, ok := p.NextElement()
		if !ok {
			t.Errorf("expected ok")
		}
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
	_, ok := p.NextElement()
	if ok {
		t.Errorf("expected not ok")
	}
}

func TestReadWritePayloadWithMultipleElementsIter(t *testing.T) {

	const (
		byteToWrite    = 252
		bytesToWrite   = "hellope"
		uint32ToWrite  = 42
		float32ToWrite = 55.5
		expectedN      = 1
	)

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
		buffer.WriteByte(tt.b)
		buffer.WriteBytes(tt.bs)
		buffer.WriteUint32(tt.u)
		buffer.WriteFloat32(tt.f)
		buffer.EndPayloadElement()
	}

	payload := buffer.Bytes()
	p, n := protocol.NewPayload(payload)
	if n != len(tts) {
		t.Errorf("expected %d got %d", expectedN, n)
	}

	for i, element := range p.Elements() {
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
}
