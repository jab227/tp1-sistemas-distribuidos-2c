package protocol

import (
	"bytes"
	"encoding/binary"
	"iter"
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

func (p *PayloadBuffer) BeginPayloadElement() {
	p.tmp.Reset()
}

func (p *PayloadBuffer) EndPayloadElement() {
	elementBytes := p.tmp.Bytes()
	binary.LittleEndian.PutUint32(p.fourBytesBuf[:], uint32(len(elementBytes)))
	p.buf.Write(p.fourBytesBuf[:])
	p.buf.Write(elementBytes)
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

// TODO(juan): make it a method of the message
type PayloadElements struct {
	payloads [][]byte
	pos      int
}

func newPayloadElements(p []byte) (*PayloadElements, int) {
	cnt := binary.LittleEndian.Uint32(p[:4])
	p = p[4:]
	payloads := make([][]byte, cnt)
	for i := 0; i < int(cnt); i++ {
		l := binary.LittleEndian.Uint32(p[:4])
		p = p[4:]
		payloads[i] = p[:l]
		p = p[l:]

	}
	return &PayloadElements{payloads, 0}, int(cnt)
}

type Element []byte

func (p *PayloadElements) Iter() iter.Seq2[int, Element] {
	return func(yield func(int, Element) bool) {
		for i, element := range p.payloads {
			if !yield(i, Element(element)) {
				return
			}
		}
	}
}

func (p *PayloadElements) NextElement() (Element, bool) {
	if p.pos == len(p.payloads) {
		return Element{}, false
	}
	element := p.payloads[p.pos]
	p.pos++
	return Element(element), true
}

func (p *Element) ReadUint32() uint32 {
	value := binary.LittleEndian.Uint32((*p)[:4])
	*p = (*p)[4:]
	return value
}

func (p *Element) ReadFloat32() float32 {
	value := binary.LittleEndian.Uint32((*p)[:4])
	(*p) = (*p)[4:]
	return math.Float32frombits(value)
}

func (p *Element) ReadByte() byte {
	b := (*p)[0]
	(*p) = (*p)[1:]
	return b
}

func (p *Element) ReadBytes() []byte {
	length := binary.LittleEndian.Uint32((*p)[:4])
	(*p) = (*p)[4:]
	data := (*p)[:length]
	(*p) = (*p)[length:]
	return data
}
