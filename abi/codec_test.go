package abi

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCodec(t *testing.T) {
	t.Run("should work", func(t *testing.T) {
		codec, err := newCodec(argsNewCodec{
			pubKeyLength: 32,
		})

		require.NoError(t, err)
		require.NotNil(t, codec)
	})

	t.Run("should err if bad public key length", func(t *testing.T) {
		_, err := newCodec(argsNewCodec{
			pubKeyLength: 0,
		})

		require.ErrorContains(t, err, "bad public key length")
	})
}

func TestCodec_EncodeNested(t *testing.T) {
	codec, _ := newCodec(argsNewCodec{
		pubKeyLength: 32,
	})

	doTest := func(t *testing.T, value any, expected string) {
		encoded, err := codec.EncodeNested(value)
		require.NoError(t, err)
		require.Equal(t, expected, hex.EncodeToString(encoded))
	}

	t.Run("bool", func(t *testing.T) {
		doTest(t, BoolValue{Value: false}, "00")
		doTest(t, BoolValue{Value: true}, "01")
	})

	t.Run("u8, i8", func(t *testing.T) {
		doTest(t, U8Value{Value: 0x00}, "00")
		doTest(t, U8Value{Value: 0x01}, "01")
		doTest(t, U8Value{Value: 0x42}, "42")
		doTest(t, U8Value{Value: 0xff}, "ff")

		doTest(t, I8Value{Value: 0x00}, "00")
		doTest(t, I8Value{Value: 0x01}, "01")
		doTest(t, I8Value{Value: -1}, "ff")
		doTest(t, I8Value{Value: -128}, "80")
		doTest(t, I8Value{Value: 127}, "7f")
	})

	t.Run("u16, i16", func(t *testing.T) {
		doTest(t, U16Value{Value: 0x00}, "0000")
		doTest(t, U16Value{Value: 0x11}, "0011")
		doTest(t, U16Value{Value: 0x1234}, "1234")
		doTest(t, U16Value{Value: 0xffff}, "ffff")

		doTest(t, I16Value{Value: 0x0000}, "0000")
		doTest(t, I16Value{Value: 0x0011}, "0011")
		doTest(t, I16Value{Value: -1}, "ffff")
		doTest(t, I16Value{Value: -32768}, "8000")
	})

	t.Run("u32, i32", func(t *testing.T) {
		doTest(t, U32Value{Value: 0x00000000}, "00000000")
		doTest(t, U32Value{Value: 0x00000011}, "00000011")
		doTest(t, U32Value{Value: 0x00001122}, "00001122")
		doTest(t, U32Value{Value: 0x00112233}, "00112233")
		doTest(t, U32Value{Value: 0x11223344}, "11223344")
		doTest(t, U32Value{Value: 0xffffffff}, "ffffffff")

		doTest(t, I32Value{Value: 0x00000000}, "00000000")
		doTest(t, I32Value{Value: 0x00000011}, "00000011")
		doTest(t, I32Value{Value: -1}, "ffffffff")
		doTest(t, I32Value{Value: -2147483648}, "80000000")
	})

	t.Run("u64, i64", func(t *testing.T) {
		doTest(t, U64Value{Value: 0x0000000000000000}, "0000000000000000")
		doTest(t, U64Value{Value: 0x0000000000000011}, "0000000000000011")
		doTest(t, U64Value{Value: 0x0000000000001122}, "0000000000001122")
		doTest(t, U64Value{Value: 0x0000000000112233}, "0000000000112233")
		doTest(t, U64Value{Value: 0x0000000011223344}, "0000000011223344")
		doTest(t, U64Value{Value: 0x0000001122334455}, "0000001122334455")
		doTest(t, U64Value{Value: 0x0000112233445566}, "0000112233445566")
		doTest(t, U64Value{Value: 0x0011223344556677}, "0011223344556677")
		doTest(t, U64Value{Value: 0x1122334455667788}, "1122334455667788")
		doTest(t, U64Value{Value: 0xffffffffffffffff}, "ffffffffffffffff")

		doTest(t, I64Value{Value: 0x0000000000000000}, "0000000000000000")
		doTest(t, I64Value{Value: 0x0000000000000011}, "0000000000000011")
		doTest(t, I64Value{Value: -1}, "ffffffffffffffff")
	})

	t.Run("bigInt", func(t *testing.T) {
		doTest(t, BigIntValue{Value: big.NewInt(0)}, "00000000")
		doTest(t, BigIntValue{Value: big.NewInt(1)}, "0000000101")
		doTest(t, BigIntValue{Value: big.NewInt(-1)}, "00000001ff")
	})

	t.Run("address", func(t *testing.T) {
		data, _ := hex.DecodeString("0139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1")
		doTest(t, AddressValue{Value: data}, "0139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1")
	})

	t.Run("address (bad)", func(t *testing.T) {
		data, _ := hex.DecodeString("0139472eff6886771a982f3083da5d42")
		_, err := codec.EncodeNested(AddressValue{Value: data})
		require.ErrorContains(t, err, "public key (address) has invalid length")
	})

	t.Run("string", func(t *testing.T) {
		doTest(t, StringValue{Value: ""}, "00000000")
		doTest(t, StringValue{Value: "abc"}, "00000003616263")
	})

	t.Run("bytes", func(t *testing.T) {
		doTest(t, BytesValue{Value: []byte{}}, "00000000")
		doTest(t, BytesValue{Value: []byte{'a', 'b', 'c'}}, "00000003616263")
	})

	t.Run("struct", func(t *testing.T) {
		fooStruct := StructValue{
			Fields: []Field{
				{
					Value: U8Value{Value: 0x01},
				},
				{
					Value: U16Value{Value: 0x4142},
				},
			},
		}

		doTest(t, fooStruct, "014142")
	})

	t.Run("enum (discriminant == 0)", func(t *testing.T) {
		fooEnum := EnumValue{
			Discriminant: 0,
		}

		doTest(t, fooEnum, "00")
	})

	t.Run("enum (discriminant != 0)", func(t *testing.T) {
		fooEnum := EnumValue{
			Discriminant: 42,
		}

		doTest(t, fooEnum, "2a")
	})

	t.Run("enum with Fields", func(t *testing.T) {
		fooEnum := EnumValue{
			Discriminant: 42,
			Fields: []Field{
				{
					Value: U8Value{Value: 0x01},
				},
				{
					Value: U16Value{Value: 0x4142},
				},
			},
		}

		doTest(t, fooEnum, "2a014142")
	})

	t.Run("option with value", func(t *testing.T) {
		fooOption := OptionValue{
			Value: U16Value{Value: 0x08},
		}

		doTest(t, fooOption, "010008")
	})

	t.Run("option without value", func(t *testing.T) {
		fooOption := OptionValue{
			Value: nil,
		}

		doTest(t, fooOption, "00")
	})

	t.Run("list", func(t *testing.T) {
		fooList := InputListValue{
			Items: []any{
				U16Value{Value: 1},
				U16Value{Value: 2},
				U16Value{Value: 3},
			},
		}

		doTest(t, fooList, "00000003000100020003")
	})

	t.Run("should err when unknown type", func(t *testing.T) {
		type dummy struct {
			foobar string
		}

		encoded, err := codec.EncodeNested(&dummy{foobar: "hello"})
		require.ErrorContains(t, err, "unsupported type for nested encoding: *abi.dummy")
		require.Nil(t, encoded)
	})
}

func TestCodec_EncodeTopLevel(t *testing.T) {
	codec, _ := newCodec(argsNewCodec{
		pubKeyLength: 32,
	})

	doTest := func(t *testing.T, value any, expected string) {
		encoded, err := codec.EncodeTopLevel(value)
		require.NoError(t, err)
		require.Equal(t, expected, hex.EncodeToString(encoded))
	}

	t.Run("bool", func(t *testing.T) {
		doTest(t, BoolValue{Value: false}, "")
		doTest(t, BoolValue{Value: true}, "01")
	})

	t.Run("u8, i8", func(t *testing.T) {
		doTest(t, U8Value{Value: 0x00}, "")
		doTest(t, U8Value{Value: 0x01}, "01")

		doTest(t, I8Value{Value: 0x00}, "")
		doTest(t, I8Value{Value: 0x01}, "01")
		doTest(t, I8Value{Value: -1}, "ff")
	})

	t.Run("u16, i16", func(t *testing.T) {
		doTest(t, U16Value{Value: 0x0042}, "42")

		doTest(t, I16Value{Value: 0x0000}, "")
		doTest(t, I16Value{Value: 0x0011}, "11")
		doTest(t, I16Value{Value: -1}, "ff")
	})

	t.Run("u32, i32", func(t *testing.T) {
		doTest(t, U32Value{Value: 0x00004242}, "4242")

		doTest(t, I32Value{Value: 0x00000000}, "")
		doTest(t, I32Value{Value: 0x00000011}, "11")
		doTest(t, I32Value{Value: -1}, "ff")
	})

	t.Run("u64, i64", func(t *testing.T) {
		doTest(t, U64Value{Value: 0x0042434445464748}, "42434445464748")

		doTest(t, I64Value{Value: 0x0000000000000000}, "")
		doTest(t, I64Value{Value: 0x0000000000000011}, "11")
		doTest(t, I64Value{Value: -1}, "ff")
	})

	t.Run("bigInt", func(t *testing.T) {
		doTest(t, BigIntValue{Value: big.NewInt(0)}, "")
		doTest(t, BigIntValue{Value: big.NewInt(1)}, "01")
		doTest(t, BigIntValue{Value: big.NewInt(-1)}, "ff")
	})

	t.Run("address", func(t *testing.T) {
		data, _ := hex.DecodeString("0139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1")
		doTest(t, AddressValue{Value: data}, "0139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1")
	})

	t.Run("address (bad)", func(t *testing.T) {
		data, _ := hex.DecodeString("0139472eff6886771a982f3083da5d42")
		_, err := codec.EncodeTopLevel(AddressValue{Value: data})
		require.ErrorContains(t, err, "public key (address) has invalid length")
	})

	t.Run("struct", func(t *testing.T) {
		fooStruct := StructValue{
			Fields: []Field{
				{
					Value: U8Value{Value: 0x01},
				},
				{
					Value: U16Value{Value: 0x4142},
				},
			},
		}

		doTest(t, fooStruct, "014142")
	})

	t.Run("enum (discriminant == 0)", func(t *testing.T) {
		fooEnum := EnumValue{
			Discriminant: 0,
		}

		doTest(t, fooEnum, "")
	})

	t.Run("enum (discriminant != 0)", func(t *testing.T) {
		fooEnum := EnumValue{
			Discriminant: 42,
		}

		doTest(t, fooEnum, "2a")
	})

	t.Run("enum with Fields", func(t *testing.T) {
		fooEnum := EnumValue{
			Discriminant: 42,
			Fields: []Field{
				{
					Value: U8Value{Value: 0x01},
				},
				{
					Value: U16Value{Value: 0x4142},
				},
			},
		}

		doTest(t, fooEnum, "2a014142")
	})

	t.Run("should err when unknown type", func(t *testing.T) {
		type dummy struct {
			foobar string
		}

		encoded, err := codec.EncodeTopLevel(&dummy{foobar: "hello"})
		require.ErrorContains(t, err, "unsupported type for top-level encoding: *abi.dummy")
		require.Nil(t, encoded)
	})
}

func TestCodec_DecodeNested(t *testing.T) {
	codec, _ := newCodec(argsNewCodec{
		pubKeyLength: 32,
	})

	t.Run("bool (true)", func(t *testing.T) {
		data, _ := hex.DecodeString("01")
		destination := &BoolValue{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &BoolValue{Value: true}, destination)
	})

	t.Run("bool (false)", func(t *testing.T) {
		data, _ := hex.DecodeString("00")
		destination := &BoolValue{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &BoolValue{Value: false}, destination)
	})

	t.Run("u8", func(t *testing.T) {
		data, _ := hex.DecodeString("01")
		destination := &U8Value{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &U8Value{Value: 0x01}, destination)
	})

	t.Run("i8", func(t *testing.T) {
		data, _ := hex.DecodeString("ff")
		destination := &I8Value{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &I8Value{Value: -1}, destination)
	})

	t.Run("u16", func(t *testing.T) {
		data, _ := hex.DecodeString("4142")
		destination := &U16Value{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &U16Value{Value: 0x4142}, destination)
	})

	t.Run("i16", func(t *testing.T) {
		data, _ := hex.DecodeString("ffff")
		destination := &I16Value{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &I16Value{Value: -1}, destination)
	})

	t.Run("u32", func(t *testing.T) {
		data, _ := hex.DecodeString("41424344")
		destination := &U32Value{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &U32Value{Value: 0x41424344}, destination)
	})

	t.Run("i32", func(t *testing.T) {
		data, _ := hex.DecodeString("ffffffff")
		destination := &I32Value{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &I32Value{Value: -1}, destination)
	})

	t.Run("u64", func(t *testing.T) {
		data, _ := hex.DecodeString("4142434445464748")
		destination := &U64Value{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &U64Value{Value: 0x4142434445464748}, destination)
	})

	t.Run("i64", func(t *testing.T) {
		data, _ := hex.DecodeString("ffffffffffffffff")
		destination := &I64Value{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &I64Value{Value: -1}, destination)
	})

	t.Run("u16, should err because it cannot read 2 bytes", func(t *testing.T) {
		data, _ := hex.DecodeString("01")
		destination := &U16Value{}

		err := codec.DecodeNested(data, destination)
		require.ErrorContains(t, err, "cannot read exactly 2 bytes")
	})

	t.Run("u32, should err because it cannot read 4 bytes", func(t *testing.T) {
		data, _ := hex.DecodeString("4142")
		destination := &U32Value{}

		err := codec.DecodeNested(data, destination)
		require.ErrorContains(t, err, "cannot read exactly 4 bytes")
	})

	t.Run("u64, should err because it cannot read 8 bytes", func(t *testing.T) {
		data, _ := hex.DecodeString("41424344")
		destination := &U64Value{}

		err := codec.DecodeNested(data, destination)
		require.ErrorContains(t, err, "cannot read exactly 8 bytes")
	})

	t.Run("bigInt", func(t *testing.T) {
		data, _ := hex.DecodeString("00000000")
		destination := &BigIntValue{}
		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &BigIntValue{Value: big.NewInt(0)}, destination)

		data, _ = hex.DecodeString("0000000101")
		destination = &BigIntValue{}
		err = codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &BigIntValue{Value: big.NewInt(1)}, destination)

		data, _ = hex.DecodeString("00000001ff")
		destination = &BigIntValue{}
		err = codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &BigIntValue{Value: big.NewInt(-1)}, destination)
	})

	t.Run("bigInt: should err when bad data", func(t *testing.T) {
		data, _ := hex.DecodeString("0000000301")
		destination := &BigIntValue{}
		err := codec.DecodeNested(data, destination)
		require.ErrorContains(t, err, "cannot decode (nested) *abi.BigIntValue, because of: cannot read exactly 3 bytes")
	})

	t.Run("address", func(t *testing.T) {
		data, _ := hex.DecodeString("0139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1")

		destination := &AddressValue{}
		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &AddressValue{Value: data}, destination)
	})

	t.Run("address (bad)", func(t *testing.T) {
		data, _ := hex.DecodeString("0139472eff6886771a982f3083da5d42")

		destination := &AddressValue{}
		err := codec.DecodeNested(data, destination)
		require.ErrorContains(t, err, "cannot read exactly 32 bytes")
	})

	t.Run("string", func(t *testing.T) {
		data, _ := hex.DecodeString("00000000")
		destination := &StringValue{}
		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &StringValue{}, destination)

		data, _ = hex.DecodeString("00000003616263")
		destination = &StringValue{}
		err = codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &StringValue{Value: "abc"}, destination)
	})

	t.Run("bytes", func(t *testing.T) {
		data, _ := hex.DecodeString("00000000")
		destination := &BytesValue{}
		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &BytesValue{Value: []byte{}}, destination)

		data, _ = hex.DecodeString("00000003616263")
		destination = &BytesValue{}
		err = codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &BytesValue{Value: []byte{'a', 'b', 'c'}}, destination)
	})

	t.Run("struct", func(t *testing.T) {
		data, _ := hex.DecodeString("014142")

		destination := &StructValue{
			Fields: []Field{
				{
					Value: &U8Value{},
				},
				{
					Value: &U16Value{},
				},
			},
		}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &StructValue{
			Fields: []Field{
				{
					Value: &U8Value{Value: 0x01},
				},
				{
					Value: &U16Value{Value: 0x4142},
				},
			},
		}, destination)
	})

	t.Run("enum (discriminant == 0)", func(t *testing.T) {
		data, _ := hex.DecodeString("00")
		destination := &EnumValue{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &EnumValue{
			Discriminant: 0x00,
		}, destination)
	})

	t.Run("enum (discriminant != 0)", func(t *testing.T) {
		data, _ := hex.DecodeString("01")
		destination := &EnumValue{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &EnumValue{
			Discriminant: 0x01,
		}, destination)
	})

	t.Run("enum with Fields", func(t *testing.T) {
		data, _ := hex.DecodeString("01014142")

		destination := &EnumValue{
			Fields: []Field{
				{
					Value: &U8Value{},
				},
				{
					Value: &U16Value{},
				},
			},
		}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &EnumValue{
			Discriminant: 0x01,
			Fields: []Field{
				{
					Value: &U8Value{Value: 0x01},
				},
				{
					Value: &U16Value{Value: 0x4142},
				},
			},
		}, destination)
	})

	t.Run("option with value", func(t *testing.T) {
		data, _ := hex.DecodeString("010008")

		destination := &OptionValue{
			Value: &U16Value{},
		}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &OptionValue{
			Value: &U16Value{Value: 8},
		}, destination)
	})

	t.Run("option without value", func(t *testing.T) {
		data, _ := hex.DecodeString("00")

		destination := &OptionValue{
			Value: &U16Value{},
		}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &OptionValue{
			Value: nil,
		}, destination)
	})

	t.Run("list", func(t *testing.T) {
		data, _ := hex.DecodeString("00000003000100020003")

		destination := &OutputListValue{
			ItemCreator: func() any { return &U16Value{} },
			Items:       []any{},
		}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t,
			[]any{
				&U16Value{Value: 1},
				&U16Value{Value: 2},
				&U16Value{Value: 3},
			}, destination.Items)
	})

	t.Run("should err when unknown type", func(t *testing.T) {
		type dummy struct {
			foobar string
		}

		err := codec.DecodeNested([]byte{0x00}, &dummy{foobar: "hello"})
		require.ErrorContains(t, err, "unsupported type for nested decoding: *abi.dummy")
	})
}

func TestCodec_DecodeTopLevel(t *testing.T) {
	codec, _ := newCodec(argsNewCodec{
		pubKeyLength: 32,
	})

	t.Run("bool (true)", func(t *testing.T) {
		data, _ := hex.DecodeString("01")
		destination := &BoolValue{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &BoolValue{Value: true}, destination)
	})

	t.Run("bool (false)", func(t *testing.T) {
		data, _ := hex.DecodeString("")
		destination := &BoolValue{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &BoolValue{Value: false}, destination)
	})

	t.Run("u8", func(t *testing.T) {
		data, _ := hex.DecodeString("01")
		destination := &U8Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &U8Value{Value: 0x01}, destination)
	})

	t.Run("i8", func(t *testing.T) {
		data, _ := hex.DecodeString("ff")
		destination := &I8Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &I8Value{Value: -1}, destination)
	})

	t.Run("u16", func(t *testing.T) {
		data, _ := hex.DecodeString("02")
		destination := &U16Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &U16Value{Value: 0x0002}, destination)
	})

	t.Run("i16", func(t *testing.T) {
		data, _ := hex.DecodeString("ffff")
		destination := &I16Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &I16Value{Value: -1}, destination)
	})

	t.Run("u32", func(t *testing.T) {
		data, _ := hex.DecodeString("03")
		destination := &U32Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &U32Value{Value: 0x00000003}, destination)
	})

	t.Run("i32", func(t *testing.T) {
		data, _ := hex.DecodeString("ffffffff")
		destination := &I32Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &I32Value{Value: -1}, destination)
	})

	t.Run("u64", func(t *testing.T) {
		data, _ := hex.DecodeString("04")
		destination := &U64Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &U64Value{Value: 0x0000000000000004}, destination)
	})

	t.Run("i64", func(t *testing.T) {
		data, _ := hex.DecodeString("ffffffffffffffff")
		destination := &I64Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &I64Value{Value: -1}, destination)
	})

	t.Run("u8, i8: should err because decoded value is too large", func(t *testing.T) {
		data, _ := hex.DecodeString("4142")

		err := codec.DecodeTopLevel(data, &U8Value{})
		require.ErrorContains(t, err, "decoded value is too large")

		err = codec.DecodeTopLevel(data, &I8Value{})
		require.ErrorContains(t, err, "decoded value is too large")
	})

	t.Run("u16, i16: should err because decoded value is too large", func(t *testing.T) {
		data, _ := hex.DecodeString("41424344")

		err := codec.DecodeTopLevel(data, &U16Value{})
		require.ErrorContains(t, err, "decoded value is too large")

		err = codec.DecodeTopLevel(data, &I16Value{})
		require.ErrorContains(t, err, "decoded value is too large")
	})

	t.Run("u32, i32: should err because decoded value is too large", func(t *testing.T) {
		data, _ := hex.DecodeString("4142434445464748")

		err := codec.DecodeTopLevel(data, &U32Value{})
		require.ErrorContains(t, err, "decoded value is too large")

		err = codec.DecodeTopLevel(data, &I32Value{})
		require.ErrorContains(t, err, "decoded value is too large")
	})

	t.Run("u64, i64: should err because decoded value is too large", func(t *testing.T) {
		data, _ := hex.DecodeString("41424344454647489876")

		err := codec.DecodeTopLevel(data, &U64Value{})
		require.ErrorContains(t, err, "decoded value is too large")

		err = codec.DecodeTopLevel(data, &I64Value{})
		require.ErrorContains(t, err, "decoded value is too large")
	})

	t.Run("bigInt", func(t *testing.T) {
		data, _ := hex.DecodeString("")
		destination := &BigIntValue{}
		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &BigIntValue{Value: big.NewInt(0)}, destination)

		data, _ = hex.DecodeString("01")
		destination = &BigIntValue{}
		err = codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &BigIntValue{Value: big.NewInt(1)}, destination)

		data, _ = hex.DecodeString("ff")
		destination = &BigIntValue{}
		err = codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &BigIntValue{Value: big.NewInt(-1)}, destination)
	})

	t.Run("struct", func(t *testing.T) {
		data, _ := hex.DecodeString("014142")

		destination := &StructValue{
			Fields: []Field{
				{
					Value: &U8Value{},
				},
				{
					Value: &U16Value{},
				},
			},
		}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &StructValue{
			Fields: []Field{
				{
					Value: &U8Value{Value: 0x01},
				},
				{
					Value: &U16Value{Value: 0x4142},
				},
			},
		}, destination)
	})

	t.Run("enum (discriminant == 0)", func(t *testing.T) {
		data, _ := hex.DecodeString("")
		destination := &EnumValue{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &EnumValue{
			Discriminant: 0x00,
		}, destination)
	})

	t.Run("enum (discriminant != 0)", func(t *testing.T) {
		data, _ := hex.DecodeString("01")
		destination := &EnumValue{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &EnumValue{
			Discriminant: 0x01,
		}, destination)
	})

	t.Run("enum with Fields", func(t *testing.T) {
		data, _ := hex.DecodeString("01014142")

		destination := &EnumValue{
			Fields: []Field{
				{
					Value: &U8Value{},
				},
				{
					Value: &U16Value{},
				},
			},
		}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &EnumValue{
			Discriminant: 0x01,
			Fields: []Field{
				{
					Value: &U8Value{Value: 0x01},
				},
				{
					Value: &U16Value{Value: 0x4142},
				},
			},
		}, destination)
	})

	t.Run("should err when unknown type", func(t *testing.T) {
		type dummy struct {
			foobar string
		}

		err := codec.DecodeTopLevel([]byte{0x00}, &dummy{foobar: "hello"})
		require.ErrorContains(t, err, "unsupported type for top-level decoding: *abi.dummy")
	})
}
