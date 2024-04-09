package encoding

import (
	"io"

	"github.com/multiversx/mx-chain-go/abi/values"
)

func (c *codec) encodeNestedBytes(writer io.Writer, value values.BytesValue) error {
	err := encodeLength(writer, uint32(len(value.Value)))
	if err != nil {
		return err
	}

	_, err = writer.Write(value.Value)
	return err
}

func (c *codec) decodeNestedBytes(reader io.Reader, value *values.BytesValue) error {
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
