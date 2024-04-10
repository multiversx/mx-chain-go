package abi

import (
	"fmt"
	"io"
)

func (c *codec) encodeNestedAddress(writer io.Writer, value AddressValue) error {
	return c.encodeTopLevelAddress(writer, value)
}

func (c *codec) encodeTopLevelAddress(writer io.Writer, value AddressValue) error {
	err := c.checkPubKeyLength(value.Value)
	if err != nil {
		return err
	}

	_, err = writer.Write(value.Value)
	return err
}

func (c *codec) decodeNestedAddress(reader io.Reader, value *AddressValue) error {
	data, err := readBytesExactly(reader, c.pubKeyLength)
	if err != nil {
		return err
	}

	value.Value = data
	return nil
}

func (c *codec) decodeTopLevelAddress(data []byte, value *AddressValue) error {
	err := c.checkPubKeyLength(data)
	if err != nil {
		return err
	}

	value.Value = data
	return nil
}

func (c *codec) checkPubKeyLength(pubkey []byte) error {
	if len(pubkey) != c.pubKeyLength {
		return fmt.Errorf("public key (address) has invalid length: %d", len(pubkey))
	}

	return nil
}
