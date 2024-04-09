package encoding

import (
	"bytes"
	"fmt"
	"io"
	"math"

	"github.com/multiversx/mx-chain-go/abi/values"
)

// codec is the default codec for encoding and decoding values.
//
// See:
// - https://docs.multiversx.com/developers/data/simple-values
// - https://docs.multiversx.com/developers/data/composite-values
// - https://docs.multiversx.com/developers/data/custom-types
type codec struct {
}

// NewCodec creates a new default codec.
func NewCodec() *codec {
	return &codec{}
}

func (c *codec) EncodeNested(value any) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	err := c.doEncodeNested(buffer, value)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (c *codec) doEncodeNested(writer io.Writer, value any) error {
	switch value.(type) {
	case values.BoolValue:
		return c.encodeNestedBool(writer, value.(values.BoolValue))
	case values.U8Value:
		return c.encodeNestedNumber(writer, value.(values.U8Value).Value, 1)
	case values.U16Value:
		return c.encodeNestedNumber(writer, value.(values.U16Value).Value, 2)
	case values.U32Value:
		return c.encodeNestedNumber(writer, value.(values.U32Value).Value, 4)
	case values.U64Value:
		return c.encodeNestedNumber(writer, value.(values.U64Value).Value, 8)
	case values.I8Value:
		return c.encodeNestedNumber(writer, value.(values.I8Value).Value, 1)
	case values.I16Value:
		return c.encodeNestedNumber(writer, value.(values.I16Value).Value, 2)
	case values.I32Value:
		return c.encodeNestedNumber(writer, value.(values.I32Value).Value, 4)
	case values.I64Value:
		return c.encodeNestedNumber(writer, value.(values.I64Value).Value, 8)
	case values.BigIntValue:
		return c.encodeNestedBigNumber(writer, value.(values.BigIntValue).Value)
	case values.AddressValue:
		return c.encodeNestedAddress(writer, value.(values.AddressValue))
	case values.StringValue:
		return c.encodeNestedString(writer, value.(values.StringValue))
	case values.BytesValue:
		return c.encodeNestedBytes(writer, value.(values.BytesValue))
	case values.StructValue:
		return c.encodeNestedStruct(writer, value.(values.StructValue))
	case values.EnumValue:
		return c.encodeNestedEnum(writer, value.(values.EnumValue))
	case values.OptionValue:
		return c.encodeNestedOption(writer, value.(values.OptionValue))
	case values.InputListValue:
		return c.encodeNestedList(writer, value.(values.InputListValue))
	default:
		return fmt.Errorf("unsupported type for nested encoding: %T", value)
	}
}

func (c *codec) EncodeTopLevel(value any) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	err := c.doEncodeTopLevel(buffer, value)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (c *codec) doEncodeTopLevel(writer io.Writer, value any) error {
	switch value.(type) {
	case values.BoolValue:
		return c.encodeTopLevelBool(writer, value.(values.BoolValue))
	case values.U8Value:
		return c.encodeTopLevelUnsignedNumber(writer, uint64(value.(values.U8Value).Value))
	case values.U16Value:
		return c.encodeTopLevelUnsignedNumber(writer, uint64(value.(values.U16Value).Value))
	case values.U32Value:
		return c.encodeTopLevelUnsignedNumber(writer, uint64(value.(values.U32Value).Value))
	case values.U64Value:
		return c.encodeTopLevelUnsignedNumber(writer, value.(values.U64Value).Value)
	case values.I8Value:
		return c.encodeTopLevelSignedNumber(writer, int64(value.(values.I8Value).Value))
	case values.I16Value:
		return c.encodeTopLevelSignedNumber(writer, int64(value.(values.I16Value).Value))
	case values.I32Value:
		return c.encodeTopLevelSignedNumber(writer, int64(value.(values.I32Value).Value))
	case values.I64Value:
		return c.encodeTopLevelSignedNumber(writer, value.(values.I64Value).Value)
	case values.BigIntValue:
		return c.encodeTopLevelBigNumber(writer, value.(values.BigIntValue).Value)
	case values.AddressValue:
		return c.encodeTopLevelAddress(writer, value.(values.AddressValue))
	case values.StructValue:
		return c.encodeTopLevelStruct(writer, value.(values.StructValue))
	case values.EnumValue:
		return c.encodeTopLevelEnum(writer, value.(values.EnumValue))
	default:
		return fmt.Errorf("unsupported type for top-level encoding: %T", value)
	}
}

func (c *codec) DecodeNested(data []byte, value any) error {
	reader := bytes.NewReader(data)
	err := c.doDecodeNested(reader, value)
	if err != nil {
		return fmt.Errorf("cannot decode (nested) %T, because of: %w", value, err)
	}

	return nil
}

func (c *codec) doDecodeNested(reader io.Reader, value any) error {
	switch value.(type) {
	case *values.BoolValue:
		return c.decodeNestedBool(reader, value.(*values.BoolValue))
	case *values.U8Value:
		return c.decodeNestedNumber(reader, &value.(*values.U8Value).Value, 1)
	case *values.U16Value:
		return c.decodeNestedNumber(reader, &value.(*values.U16Value).Value, 2)
	case *values.U32Value:
		return c.decodeNestedNumber(reader, &value.(*values.U32Value).Value, 4)
	case *values.U64Value:
		return c.decodeNestedNumber(reader, &value.(*values.U64Value).Value, 8)
	case *values.I8Value:
		return c.decodeNestedNumber(reader, &value.(*values.I8Value).Value, 1)
	case *values.I16Value:
		return c.decodeNestedNumber(reader, &value.(*values.I16Value).Value, 2)
	case *values.I32Value:
		return c.decodeNestedNumber(reader, &value.(*values.I32Value).Value, 4)
	case *values.I64Value:
		return c.decodeNestedNumber(reader, &value.(*values.I64Value).Value, 8)
	case *values.BigIntValue:
		n, err := c.decodeNestedBigNumber(reader)
		if err != nil {
			return err
		}

		value.(*values.BigIntValue).Value = n
		return nil
	case *values.AddressValue:
		return c.decodeNestedAddress(reader, value.(*values.AddressValue))
	case *values.StringValue:
		return c.decodeNestedString(reader, value.(*values.StringValue))
	case *values.BytesValue:
		return c.decodeNestedBytes(reader, value.(*values.BytesValue))
	case *values.StructValue:
		return c.decodeNestedStruct(reader, value.(*values.StructValue))
	case *values.EnumValue:
		return c.decodeNestedEnum(reader, value.(*values.EnumValue))
	case *values.OptionValue:
		return c.decodeNestedOption(reader, value.(*values.OptionValue))
	case *values.OutputListValue:
		return c.decodeNestedList(reader, value.(*values.OutputListValue))
	default:
		return fmt.Errorf("unsupported type for nested decoding: %T", value)
	}
}

func (c *codec) DecodeTopLevel(data []byte, value any) error {
	err := c.doDecodeTopLevel(data, value)
	if err != nil {
		return fmt.Errorf("cannot decode (top-level) %T, because of: %w", value, err)
	}

	return nil
}

func (c *codec) doDecodeTopLevel(data []byte, value any) error {
	switch value.(type) {
	case *values.BoolValue:
		return c.decodeTopLevelBool(data, value.(*values.BoolValue))
	case *values.U8Value:
		n, err := c.decodeTopLevelUnsignedNumber(data, math.MaxUint8)
		if err != nil {
			return err
		}

		value.(*values.U8Value).Value = uint8(n)
	case *values.U16Value:
		n, err := c.decodeTopLevelUnsignedNumber(data, math.MaxUint16)
		if err != nil {
			return err
		}

		value.(*values.U16Value).Value = uint16(n)
	case *values.U32Value:
		n, err := c.decodeTopLevelUnsignedNumber(data, math.MaxUint32)
		if err != nil {
			return err
		}

		value.(*values.U32Value).Value = uint32(n)
	case *values.U64Value:
		n, err := c.decodeTopLevelUnsignedNumber(data, math.MaxUint64)
		if err != nil {
			return err
		}

		value.(*values.U64Value).Value = uint64(n)
	case *values.I8Value:
		n, err := c.decodeTopLevelSignedNumber(data, math.MaxInt8)
		if err != nil {
			return err
		}

		value.(*values.I8Value).Value = int8(n)
	case *values.I16Value:
		n, err := c.decodeTopLevelSignedNumber(data, math.MaxInt16)
		if err != nil {
			return err
		}

		value.(*values.I16Value).Value = int16(n)
	case *values.I32Value:
		n, err := c.decodeTopLevelSignedNumber(data, math.MaxInt32)
		if err != nil {
			return err
		}

		value.(*values.I32Value).Value = int32(n)

	case *values.I64Value:
		n, err := c.decodeTopLevelSignedNumber(data, math.MaxInt64)
		if err != nil {
			return err
		}

		value.(*values.I64Value).Value = int64(n)
	case *values.BigIntValue:
		n := c.decodeTopLevelBigNumber(data)
		value.(*values.BigIntValue).Value = n
	case *values.AddressValue:
		return c.decodeTopLevelAddress(data, value.(*values.AddressValue))
	case *values.StructValue:
		return c.decodeTopLevelStruct(data, value.(*values.StructValue))
	case *values.EnumValue:
		return c.decodeTopLevelEnum(data, value.(*values.EnumValue))
	default:
		return fmt.Errorf("unsupported type for top-level decoding: %T", value)
	}

	return nil
}
