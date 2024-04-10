package abi

import (
	"encoding/binary"
	"fmt"
	"io"
)

func encodeLength(writer io.Writer, length uint32) error {
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, length)

	_, err := writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func decodeLength(reader io.Reader) (uint32, error) {
	bytes, err := readBytesExactly(reader, 4)
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(bytes), nil
}

func readBytesExactly(reader io.Reader, numBytes int) ([]byte, error) {
	if numBytes == 0 {
		return []byte{}, nil
	}

	data := make([]byte, numBytes)
	n, err := reader.Read(data)
	if err != nil {
		return nil, err
	}
	if n != numBytes {
		return nil, fmt.Errorf("cannot read exactly %d bytes", numBytes)
	}

	return data, err
}

func checkPubKeyLength(pubkey []byte) error {
	if len(pubkey) != pubKeyLength {
		return fmt.Errorf("public key (address) has invalid length: %d", len(pubkey))
	}

	return nil
}
