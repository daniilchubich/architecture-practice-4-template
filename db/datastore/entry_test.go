package datastore

import (
	"bufio"
	"bytes"
	"testing"
)

func TestEntry_Encode(t *testing.T) {
	e := Entry{"key", "value", calculateChecksum("key+value")}
	e.Decode(e.Encode())
	if e.key != "key" {
		t.Error("incorrect key")
	}
	if e.value != "value" {
		t.Error("incorrect value")
	}
}

func TestReadValue(t *testing.T) {
	ch := calculateChecksum("test-value")
	e := Entry{"key", "test-value", ch}
	data := e.Encode()
	v, chRead, err := readValue(bufio.NewReader(bytes.NewReader(data)))
	if err != nil {
		t.Fatal(err)
	}
	if v != "test-value" {
		t.Errorf("Got bad value [%s]", v)
	}
	if ch != chRead {
		t.Errorf("Got bad checksum [%s]", ch)
	}
}
