package datastore

import (
	"bufio"
	"encoding/binary"
	"fmt"
)

type Entry struct {
	key, value, checksum string
}

func (e *Entry) Encode() []byte {
	kl := len(e.key)
	vl := len(e.value)
	cl := len(e.checksum)
	size := kl + vl + cl + 12
	res := make([]byte, size)
	binary.LittleEndian.PutUint32(res, uint32(size))
	binary.LittleEndian.PutUint32(res[4:], uint32(kl))
	copy(res[8:], e.key)
	binary.LittleEndian.PutUint32(res[kl+8:], uint32(vl))
	copy(res[kl+12:], e.value)
	copy(res[kl+12+vl:], e.checksum)
	return res
}

func (e *Entry) Decode(input []byte) {
	kl := binary.LittleEndian.Uint32(input[4:])
	keyBuf := make([]byte, kl)
	copy(keyBuf, input[8:kl+8])
	e.key = string(keyBuf)

	vl := binary.LittleEndian.Uint32(input[kl+8:])
	valBuf := make([]byte, vl)
	copy(valBuf, input[kl+12:kl+12+vl])
	e.value = string(valBuf)

	e.checksum = string(input[kl+12+vl:])
}


func readValue(in *bufio.Reader) (string, string, error) {
	// Зчитування заголовка для визначення розміру ключа
	header, err := in.Peek(8)
	if err != nil {
		return "", "", err
	}
	keySize := int(binary.LittleEndian.Uint32(header[4:]))
	_, err = in.Discard(keySize + 8)
	if err != nil {
		return "", "", err
	}

	// Зчитування заголовка для визначення розміру значення
	header, err = in.Peek(4)
	if err != nil {
		return "", "", err
	}
	valSize := int(binary.LittleEndian.Uint32(header))
	_, err = in.Discard(4)
	if err != nil {
		return "", "", err
	}

	// Зчитування значення
	data := make([]byte, valSize)
	n, err := in.Read(data)
	if err != nil {
		return "", "", err
	}
	if n != valSize {
		return "", "", fmt.Errorf("can't read value bytes (read %d, expected %d)", n, valSize)
	}
	value := string(data)

	// Зчитування контрольної суми
	checksumSize := 40 // Розмір SHA-1 контрольної суми у вигляді рядка (40 символів)
	checksumData := make([]byte, checksumSize)
	n, err = in.Read(checksumData)
	if err != nil {
		return "", "", err
	}
	if n != checksumSize {
		return "", "", fmt.Errorf("can't read checksum bytes (read %d, expected %d)", n, checksumSize)
	}
	checksum := string(checksumData)

	return value, checksum, nil
}



