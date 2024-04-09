package abi

import (
	"bytes"
	"fmt"
	"io"
)

func (c *codec) encodeNestedStruct(writer io.Writer, value StructValue) error {
	for _, field := range value.Fields {
		err := c.doEncodeNested(writer, field.Value)
		if err != nil {
			return fmt.Errorf("cannot encode field '%s' of struct, because of: %w", field.Name, err)
		}
	}

	return nil
}

func (c *codec) encodeTopLevelStruct(writer io.Writer, value StructValue) error {
	return c.encodeNestedStruct(writer, value)
}

func (c *codec) decodeNestedStruct(reader io.Reader, value *StructValue) error {
	for _, field := range value.Fields {
		err := c.doDecodeNested(reader, field.Value)
		if err != nil {
			return fmt.Errorf("cannot decode field '%s' of struct, because of: %w", field.Name, err)
		}
	}

	return nil
}

func (c *codec) decodeTopLevelStruct(data []byte, value *StructValue) error {
	reader := bytes.NewReader(data)
	return c.decodeNestedStruct(reader, value)
}
