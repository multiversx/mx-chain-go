package encoding

import (
	"io"

	"github.com/multiversx/mx-chain-go/abi/values"
)

func (c *codec) encodeNestedString(writer io.Writer, value values.StringValue) error {
	data := []byte(value.Value)
	err := encodeLength(writer, uint32(len(data)))
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	return err
}

func (c *codec) decodeNestedString(reader io.Reader, value *values.StringValue) error {
	length, err := decodeLength(reader)
	if err != nil {
		return err
	}

	data, err := readBytesExactly(reader, int(length))
	if err != nil {
		return err
	}

	value.Value = string(data)
	return nil
}
