package abi

import (
	"io"
)

func (c *codec) encodeNestedBytes(writer io.Writer, value BytesValue) error {
	err := encodeLength(writer, uint32(len(value.Value)))
	if err != nil {
		return err
	}

	_, err = writer.Write(value.Value)
	return err
}

func (c *codec) decodeNestedBytes(reader io.Reader, value *BytesValue) error {
	length, err := decodeLength(reader)
	if err != nil {
		return err
	}

	data, err := readBytesExactly(reader, int(length))
	if err != nil {
		return err
	}

	value.Value = data
	return nil
}
