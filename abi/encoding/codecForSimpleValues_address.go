package encoding

import (
	"io"

	"github.com/multiversx/mx-chain-go/abi/values"
)

func (c *codec) encodeNestedAddress(writer io.Writer, value values.AddressValue) error {
	return c.encodeTopLevelAddress(writer, value)
}

func (c *codec) encodeTopLevelAddress(writer io.Writer, value values.AddressValue) error {
	err := checkPubKeyLength(value.Value)
	if err != nil {
		return err
	}

	_, err = writer.Write(value.Value)
	return err
}

func (c *codec) decodeNestedAddress(reader io.Reader, value *values.AddressValue) error {
	data, err := readBytesExactly(reader, pubKeyLength)
	if err != nil {
		return err
	}

	value.Value = data
	return nil
}

func (c *codec) decodeTopLevelAddress(data []byte, value *values.AddressValue) error {
	err := checkPubKeyLength(data)
	if err != nil {
		return err
	}

	value.Value = data
	return nil
}
