package chat

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	Magic   uint16 = 0xCA7E
	HeadLen int    = 2 + 4 // Type(2) + Length(4)
)

type Frame struct {
	Type   uint16
	Length uint32
	Value  []byte
}

func EncodeFrame(f Frame) []byte {
	buf := make([]byte, HeadLen+len(f.Value))
	binary.BigEndian.PutUint16(buf[0:2], f.Type)
	binary.BigEndian.PutUint32(buf[2:6], uint32(len(f.Value)))
	copy(buf[6:], f.Value)
	return buf
}

func DecodeFrame(r io.Reader) (*Frame, error) {
	header := make([]byte, HeadLen)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	f := &Frame{
		Type:   binary.BigEndian.Uint16(header[0:2]),
		Length: binary.BigEndian.Uint32(header[2:6]),
	}
	if f.Length > 1<<20 { // 1MB max
		return nil, fmt.Errorf("frame too large: %d", f.Length)
	}
	if f.Length > 0 {
		f.Value = make([]byte, f.Length)
		if _, err := io.ReadFull(r, f.Value); err != nil {
			return nil, fmt.Errorf("read value: %w", err)
		}
	}
	return f, nil
}
