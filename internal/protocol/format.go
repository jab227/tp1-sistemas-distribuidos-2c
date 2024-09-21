package protocol

import (
	"bytes"
	"encoding/binary"
	"math"
)

type PayloadBuffer struct {
	buf          *bytes.Buffer
	tmp          *bytes.Buffer
	elementCount int
	fourBytesBuf [4]byte
}

func NewPayloadBuffer(elementCount int) *PayloadBuffer {
	var buf bytes.Buffer
	elementCountBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(elementCountBuf, uint32(elementCount))
	buf.Write(elementCountBuf)
	return &PayloadBuffer{
		buf:          &buf,
		tmp:          &bytes.Buffer{},
		elementCount: elementCount,
	}
}

func (p *PayloadBuffer) EndPayloadElement() {
	elementBytes := p.tmp.Bytes()
	binary.LittleEndian.PutUint32(p.fourBytesBuf[:], uint32(len(elementBytes)))
	p.buf.Write(p.fourBytesBuf[:])
	p.buf.Write(elementBytes)
	p.tmp.Reset()
}

func (p *PayloadBuffer) WriteByte(b byte) {
	p.tmp.WriteByte(b)
}

func (p *PayloadBuffer) WriteBytes(bs []byte) {
	binary.LittleEndian.PutUint32(p.fourBytesBuf[:], uint32(len(bs)))
	p.tmp.Write(p.fourBytesBuf[:])
	p.tmp.Write(bs)
}

func (p *PayloadBuffer) WriteUint32(v uint32) {
	binary.LittleEndian.PutUint32(p.fourBytesBuf[:], v)
	p.tmp.Write(p.fourBytesBuf[:])
}

func (p *PayloadBuffer) WriteFloat32(v float32) {
	binary.LittleEndian.PutUint32(p.fourBytesBuf[:], math.Float32bits(v))
	p.tmp.Write(p.fourBytesBuf[:])
}

func (p *PayloadBuffer) Bytes() []byte {
	return p.buf.Bytes()
}

type Payload struct {
	payloads [][]byte
	pos      int
}

func NewPayload(p []byte) (*Payload, int) {
	cnt := binary.LittleEndian.Uint32(p[:4])
	p = p[4:]
	payloads := make([][]byte, cnt)
	for i := 0; i < int(cnt); i++ {
		l := binary.LittleEndian.Uint32(p[:4])
		p = p[4:]
		payloads[i] = p[:l]
		p = p[l:]

	}
	return &Payload{payloads, 0}, int(cnt)
}



func (p *Payload) ReadUint32() uint32 {
	value := binary.LittleEndian.Uint32(p.payloads[0][:4])
	p.payloads[0] = p.payloads[0][4:]
	return value
}

func (p *Payload) ReadFloat32() float32 {
	value := binary.LittleEndian.Uint32(p.payloads[0][:4])
	p.payloads[0] = p.payloads[0][4:]
	return math.Float32frombits(value)
}

func (p *Payload) ReadByte() byte {
	b := p.payloads[0][0]
	p.payloads[0] = p.payloads[0][1:]
	return b
}

func (p *Payload) ReadBytes() []byte {
	length := binary.LittleEndian.Uint32(p.payloads[0][:4])
	p.payloads[0] = p.payloads[0][4:]
	data := p.payloads[0][:length]
	p.payloads[0] = p.payloads[0][length:]
	return data
}
