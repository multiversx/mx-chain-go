package abi

import (
	"bytes"
	"fmt"
	"io"
	"math"
)

type codec struct {
	pubKeyLength int
}

// argsNewCodec defines the arguments needed for a new codec
type argsNewCodec struct {
	pubKeyLength int
}

// newCodec creates a new default codec which follows the rules of the MultiversX Serialization format:
// https://docs.multiversx.com/developers/data/serialization-overview
func newCodec(args argsNewCodec) *codec {
	return &codec{
		pubKeyLength: args.pubKeyLength,
	}
}

// EncodeNested encodes the given value following the nested encoding rules
func (c *codec) EncodeNested(value any) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	err := c.doEncodeNested(buffer, value)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (c *codec) doEncodeNested(writer io.Writer, value any) error {
	switch value := value.(type) {
	case BoolValue:
		return c.encodeNestedBool(writer, value)
	case U8Value:
		return c.encodeNestedNumber(writer, value.Value, 1)
	case U16Value:
		return c.encodeNestedNumber(writer, value.Value, 2)
	case U32Value:
		return c.encodeNestedNumber(writer, value.Value, 4)
	case U64Value:
		return c.encodeNestedNumber(writer, value.Value, 8)
	case I8Value:
		return c.encodeNestedNumber(writer, value.Value, 1)
	case I16Value:
		return c.encodeNestedNumber(writer, value.Value, 2)
	case I32Value:
		return c.encodeNestedNumber(writer, value.Value, 4)
	case I64Value:
		return c.encodeNestedNumber(writer, value.Value, 8)
	case BigIntValue:
		return c.encodeNestedBigNumber(writer, value.Value)
	case AddressValue:
		return c.encodeNestedAddress(writer, value)
	case StringValue:
		return c.encodeNestedString(writer, value)
	case BytesValue:
		return c.encodeNestedBytes(writer, value)
	case StructValue:
		return c.encodeNestedStruct(writer, value)
	case EnumValue:
		return c.encodeNestedEnum(writer, value)
	case OptionValue:
		return c.encodeNestedOption(writer, value)
	case InputListValue:
		return c.encodeNestedList(writer, value)
	default:
		return fmt.Errorf("unsupported type for nested encoding: %T", value)
	}
}

// EncodeTopLevel encodes the given value following the top-level encoding rules
func (c *codec) EncodeTopLevel(value any) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	err := c.doEncodeTopLevel(buffer, value)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (c *codec) doEncodeTopLevel(writer io.Writer, value any) error {
	switch value := value.(type) {
	case BoolValue:
		return c.encodeTopLevelBool(writer, value)
	case U8Value:
		return c.encodeTopLevelUnsignedNumber(writer, uint64(value.Value))
	case U16Value:
		return c.encodeTopLevelUnsignedNumber(writer, uint64(value.Value))
	case U32Value:
		return c.encodeTopLevelUnsignedNumber(writer, uint64(value.Value))
	case U64Value:
		return c.encodeTopLevelUnsignedNumber(writer, value.Value)
	case I8Value:
		return c.encodeTopLevelSignedNumber(writer, int64(value.Value))
	case I16Value:
		return c.encodeTopLevelSignedNumber(writer, int64(value.Value))
	case I32Value:
		return c.encodeTopLevelSignedNumber(writer, int64(value.Value))
	case I64Value:
		return c.encodeTopLevelSignedNumber(writer, value.Value)
	case BigIntValue:
		return c.encodeTopLevelBigNumber(writer, value.Value)
	case AddressValue:
		return c.encodeTopLevelAddress(writer, value)
	case StructValue:
		return c.encodeTopLevelStruct(writer, value)
	case EnumValue:
		return c.encodeTopLevelEnum(writer, value)
	default:
		return fmt.Errorf("unsupported type for top-level encoding: %T", value)
	}
}

// DecodeNested decodes the given data into the provided object following the nested decoding rules
func (c *codec) DecodeNested(data []byte, value any) error {
	reader := bytes.NewReader(data)
	err := c.doDecodeNested(reader, value)
	if err != nil {
		return fmt.Errorf("cannot decode (nested) %T, because of: %w", value, err)
	}

	return nil
}

func (c *codec) doDecodeNested(reader io.Reader, value any) error {
	switch value := value.(type) {
	case *BoolValue:
		return c.decodeNestedBool(reader, value)
	case *U8Value:
		return c.decodeNestedNumber(reader, &value.Value, 1)
	case *U16Value:
		return c.decodeNestedNumber(reader, &value.Value, 2)
	case *U32Value:
		return c.decodeNestedNumber(reader, &value.Value, 4)
	case *U64Value:
		return c.decodeNestedNumber(reader, &value.Value, 8)
	case *I8Value:
		return c.decodeNestedNumber(reader, &value.Value, 1)
	case *I16Value:
		return c.decodeNestedNumber(reader, &value.Value, 2)
	case *I32Value:
		return c.decodeNestedNumber(reader, &value.Value, 4)
	case *I64Value:
		return c.decodeNestedNumber(reader, &value.Value, 8)
	case *BigIntValue:
		n, err := c.decodeNestedBigNumber(reader)
		if err != nil {
			return err
		}

		value.Value = n
		return nil
	case *AddressValue:
		return c.decodeNestedAddress(reader, value)
	case *StringValue:
		return c.decodeNestedString(reader, value)
	case *BytesValue:
		return c.decodeNestedBytes(reader, value)
	case *StructValue:
		return c.decodeNestedStruct(reader, value)
	case *EnumValue:
		return c.decodeNestedEnum(reader, value)
	case *OptionValue:
		return c.decodeNestedOption(reader, value)
	case *OutputListValue:
		return c.decodeNestedList(reader, value)
	default:
		return fmt.Errorf("unsupported type for nested decoding: %T", value)
	}
}

// DecodeTopLevel decodes the given data into the provided object following the top-level decoding rules
func (c *codec) DecodeTopLevel(data []byte, value any) error {
	err := c.doDecodeTopLevel(data, value)
	if err != nil {
		return fmt.Errorf("cannot decode (top-level) %T, because of: %w", value, err)
	}

	return nil
}

func (c *codec) doDecodeTopLevel(data []byte, value any) error {
	switch value := value.(type) {
	case *BoolValue:
		return c.decodeTopLevelBool(data, value)
	case *U8Value:
		n, err := c.decodeTopLevelUnsignedNumber(data, math.MaxUint8)
		if err != nil {
			return err
		}

		value.Value = uint8(n)
	case *U16Value:
		n, err := c.decodeTopLevelUnsignedNumber(data, math.MaxUint16)
		if err != nil {
			return err
		}

		value.Value = uint16(n)
	case *U32Value:
		n, err := c.decodeTopLevelUnsignedNumber(data, math.MaxUint32)
		if err != nil {
			return err
		}

		value.Value = uint32(n)
	case *U64Value:
		n, err := c.decodeTopLevelUnsignedNumber(data, math.MaxUint64)
		if err != nil {
			return err
		}

		value.Value = uint64(n)
	case *I8Value:
		n, err := c.decodeTopLevelSignedNumber(data, math.MaxInt8)
		if err != nil {
			return err
		}

		value.Value = int8(n)
	case *I16Value:
		n, err := c.decodeTopLevelSignedNumber(data, math.MaxInt16)
		if err != nil {
			return err
		}

		value.Value = int16(n)
	case *I32Value:
		n, err := c.decodeTopLevelSignedNumber(data, math.MaxInt32)
		if err != nil {
			return err
		}

		value.Value = int32(n)

	case *I64Value:
		n, err := c.decodeTopLevelSignedNumber(data, math.MaxInt64)
		if err != nil {
			return err
		}

		value.Value = int64(n)
	case *BigIntValue:
		n := c.decodeTopLevelBigNumber(data)
		value.Value = n
	case *AddressValue:
		return c.decodeTopLevelAddress(data, value)
	case *StructValue:
		return c.decodeTopLevelStruct(data, value)
	case *EnumValue:
		return c.decodeTopLevelEnum(data, value)
	default:
		return fmt.Errorf("unsupported type for top-level decoding: %T", value)
	}

	return nil
}
