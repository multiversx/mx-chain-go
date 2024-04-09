package abi

import (
	"fmt"
	"io"
)

func (c *codec) encodeNestedBool(writer io.Writer, value BoolValue) error {
	if value.Value {
		_, err := writer.Write([]byte{trueAsByte})
		return err
	}

	_, err := writer.Write([]byte{falseAsByte})
	return err
}

func (c *codec) decodeNestedBool(reader io.Reader, value *BoolValue) error {
	data, err := readBytesExactly(reader, 1)
	if err != nil {
		return err
	}

	value.Value, err = c.byteToBool(data[0])
	if err != nil {
		return err
	}

	return nil
}

func (c *codec) encodeTopLevelBool(writer io.Writer, value BoolValue) error {
	if !value.Value {
		// For "false", write nothing.
		return nil
	}

	_, err := writer.Write([]byte{trueAsByte})
	return err
}

func (c *codec) decodeTopLevelBool(data []byte, value *BoolValue) error {
	if len(data) == 0 {
		value.Value = false
		return nil
	}

	if len(data) == 1 {
		boolValue, err := c.byteToBool(data[0])
		if err != nil {
			return err
		}

		value.Value = boolValue
		return nil
	}

	return fmt.Errorf("unexpected boolean value: %v", data)
}

func (c *codec) byteToBool(data uint8) (bool, error) {
	switch data {
	case trueAsByte:
		return true, nil
	case falseAsByte:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected boolean value: %d", data)
	}
}
