package chat

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeFrame(t *testing.T) {
	original := Frame{Type: 0x0001, Value: []byte("hello")}
	data := EncodeFrame(original)

	decoded, err := DecodeFrame(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("DecodeFrame failed: %v", err)
	}
	if decoded.Type != original.Type {
		t.Errorf("type mismatch: %d != %d", decoded.Type, original.Type)
	}
	if string(decoded.Value) != string(original.Value) {
		t.Errorf("value mismatch: %s != %s", decoded.Value, original.Value)
	}
}

func TestDecodeFrame_EmptyValue(t *testing.T) {
	original := Frame{Type: 0x0010, Value: []byte{}}
	data := EncodeFrame(original)
	decoded, err := DecodeFrame(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("DecodeFrame failed: %v", err)
	}
	if decoded.Type != original.Type {
		t.Errorf("type mismatch")
	}
	if len(decoded.Value) != 0 {
		t.Errorf("expected empty value")
	}
}
