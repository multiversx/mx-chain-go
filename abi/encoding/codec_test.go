package encoding

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/abi/values"
	"github.com/stretchr/testify/require"
)

func TestCodec_EncodeNested(t *testing.T) {
	codec := NewCodec()

	doTest := func(t *testing.T, value any, expected string) {
		encoded, err := codec.EncodeNested(value)
		require.NoError(t, err)
		require.Equal(t, expected, hex.EncodeToString(encoded))
	}

	t.Run("bool", func(t *testing.T) {
		doTest(t, values.BoolValue{Value: false}, "00")
		doTest(t, values.BoolValue{Value: true}, "01")
	})

	t.Run("u8, i8", func(t *testing.T) {
		doTest(t, values.U8Value{Value: 0x00}, "00")
		doTest(t, values.U8Value{Value: 0x01}, "01")
		doTest(t, values.U8Value{Value: 0x42}, "42")
		doTest(t, values.U8Value{Value: 0xff}, "ff")

		doTest(t, values.I8Value{Value: 0x00}, "00")
		doTest(t, values.I8Value{Value: 0x01}, "01")
		doTest(t, values.I8Value{Value: -1}, "ff")
		doTest(t, values.I8Value{Value: -128}, "80")
		doTest(t, values.I8Value{Value: 127}, "7f")
	})

	t.Run("u16", func(t *testing.T) {
		doTest(t, values.U16Value{Value: 0x00}, "0000")
		doTest(t, values.U16Value{Value: 0x11}, "0011")
		doTest(t, values.U16Value{Value: 0x1234}, "1234")
		doTest(t, values.U16Value{Value: 0xffff}, "ffff")
	})

	t.Run("u32", func(t *testing.T) {
		doTest(t, values.U32Value{Value: 0x00000000}, "00000000")
		doTest(t, values.U32Value{Value: 0x00000011}, "00000011")
		doTest(t, values.U32Value{Value: 0x00001122}, "00001122")
		doTest(t, values.U32Value{Value: 0x00112233}, "00112233")
		doTest(t, values.U32Value{Value: 0x11223344}, "11223344")
		doTest(t, values.U32Value{Value: 0xffffffff}, "ffffffff")
	})

	t.Run("u64", func(t *testing.T) {
		doTest(t, values.U64Value{Value: 0x0000000000000000}, "0000000000000000")
		doTest(t, values.U64Value{Value: 0x0000000000000011}, "0000000000000011")
		doTest(t, values.U64Value{Value: 0x0000000000001122}, "0000000000001122")
		doTest(t, values.U64Value{Value: 0x0000000000112233}, "0000000000112233")
		doTest(t, values.U64Value{Value: 0x0000000011223344}, "0000000011223344")
		doTest(t, values.U64Value{Value: 0x0000001122334455}, "0000001122334455")
		doTest(t, values.U64Value{Value: 0x0000112233445566}, "0000112233445566")
		doTest(t, values.U64Value{Value: 0x0011223344556677}, "0011223344556677")
		doTest(t, values.U64Value{Value: 0x1122334455667788}, "1122334455667788")
		doTest(t, values.U64Value{Value: 0xffffffffffffffff}, "ffffffffffffffff")
	})

	t.Run("bigInt", func(t *testing.T) {
		doTest(t, values.BigIntValue{Value: big.NewInt(0)}, "00000000")
		doTest(t, values.BigIntValue{Value: big.NewInt(1)}, "0000000101")
		doTest(t, values.BigIntValue{Value: big.NewInt(-1)}, "00000001ff")
	})

	t.Run("address", func(t *testing.T) {
		data, _ := hex.DecodeString("0139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1")
		doTest(t, values.AddressValue{Value: data}, "0139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1")
	})

	t.Run("address (bad)", func(t *testing.T) {
		data, _ := hex.DecodeString("0139472eff6886771a982f3083da5d42")
		_, err := codec.EncodeNested(values.AddressValue{Value: data})
		require.ErrorContains(t, err, "public key (address) has invalid length")
	})

	t.Run("string", func(t *testing.T) {
		doTest(t, values.StringValue{Value: ""}, "00000000")
		doTest(t, values.StringValue{Value: "abc"}, "00000003616263")
	})

	t.Run("bytes", func(t *testing.T) {
		doTest(t, values.BytesValue{Value: []byte{}}, "00000000")
		doTest(t, values.BytesValue{Value: []byte{'a', 'b', 'c'}}, "00000003616263")
	})

	t.Run("struct", func(t *testing.T) {
		fooStruct := values.StructValue{
			Fields: []values.Field{
				{
					Value: values.U8Value{Value: 0x01},
				},
				{
					Value: values.U16Value{Value: 0x4142},
				},
			},
		}

		doTest(t, fooStruct, "014142")
	})

	t.Run("enum (discriminant == 0)", func(t *testing.T) {
		fooEnum := values.EnumValue{
			Discriminant: 0,
		}

		doTest(t, fooEnum, "00")
	})

	t.Run("enum (discriminant != 0)", func(t *testing.T) {
		fooEnum := values.EnumValue{
			Discriminant: 42,
		}

		doTest(t, fooEnum, "2a")
	})

	t.Run("enum with values.Fields", func(t *testing.T) {
		fooEnum := values.EnumValue{
			Discriminant: 42,
			Fields: []values.Field{
				{
					Value: values.U8Value{Value: 0x01},
				},
				{
					Value: values.U16Value{Value: 0x4142},
				},
			},
		}

		doTest(t, fooEnum, "2a014142")
	})

	t.Run("option with value", func(t *testing.T) {
		fooOption := values.OptionValue{
			Value: values.U16Value{Value: 0x08},
		}

		doTest(t, fooOption, "010008")
	})

	t.Run("option without value", func(t *testing.T) {
		fooOption := values.OptionValue{
			Value: nil,
		}

		doTest(t, fooOption, "00")
	})

	t.Run("list", func(t *testing.T) {
		fooList := values.InputListValue{
			Items: []any{
				values.U16Value{Value: 1},
				values.U16Value{Value: 2},
				values.U16Value{Value: 3},
			},
		}

		doTest(t, fooList, "00000003000100020003")
	})
}

func TestCodec_EncodeTopLevel(t *testing.T) {
	codec := NewCodec()

	doTest := func(t *testing.T, value any, expected string) {
		encoded, err := codec.EncodeTopLevel(value)
		require.NoError(t, err)
		require.Equal(t, expected, hex.EncodeToString(encoded))
	}

	t.Run("bool", func(t *testing.T) {
		doTest(t, values.BoolValue{Value: false}, "")
		doTest(t, values.BoolValue{Value: true}, "01")
	})

	t.Run("u8", func(t *testing.T) {
		doTest(t, values.U8Value{Value: 0x00}, "")
		doTest(t, values.U8Value{Value: 0x01}, "01")
	})

	t.Run("u16", func(t *testing.T) {
		doTest(t, values.U16Value{Value: 0x0042}, "42")
	})

	t.Run("u32", func(t *testing.T) {
		doTest(t, values.U32Value{Value: 0x00004242}, "4242")
	})

	t.Run("u64", func(t *testing.T) {
		doTest(t, values.U64Value{Value: 0x0042434445464748}, "42434445464748")
	})

	t.Run("bigInt", func(t *testing.T) {
		doTest(t, values.BigIntValue{Value: big.NewInt(0)}, "")
		doTest(t, values.BigIntValue{Value: big.NewInt(1)}, "01")
		doTest(t, values.BigIntValue{Value: big.NewInt(-1)}, "ff")
	})

	t.Run("address", func(t *testing.T) {
		data, _ := hex.DecodeString("0139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1")
		doTest(t, values.AddressValue{Value: data}, "0139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1")
	})

	t.Run("address (bad)", func(t *testing.T) {
		data, _ := hex.DecodeString("0139472eff6886771a982f3083da5d42")
		_, err := codec.EncodeTopLevel(values.AddressValue{Value: data})
		require.ErrorContains(t, err, "public key (address) has invalid length")
	})

	t.Run("struct", func(t *testing.T) {
		fooStruct := values.StructValue{
			Fields: []values.Field{
				{
					Value: values.U8Value{Value: 0x01},
				},
				{
					Value: values.U16Value{Value: 0x4142},
				},
			},
		}

		doTest(t, fooStruct, "014142")
	})

	t.Run("enum (discriminant == 0)", func(t *testing.T) {
		fooEnum := values.EnumValue{
			Discriminant: 0,
		}

		doTest(t, fooEnum, "")
	})

	t.Run("enum (discriminant != 0)", func(t *testing.T) {
		fooEnum := values.EnumValue{
			Discriminant: 42,
		}

		doTest(t, fooEnum, "2a")
	})

	t.Run("enum with values.Fields", func(t *testing.T) {
		fooEnum := values.EnumValue{
			Discriminant: 42,
			Fields: []values.Field{
				{
					Value: values.U8Value{Value: 0x01},
				},
				{
					Value: values.U16Value{Value: 0x4142},
				},
			},
		}

		doTest(t, fooEnum, "2a014142")
	})
}

func TestCodec_DecodeNested(t *testing.T) {
	codec := NewCodec()

	t.Run("bool (true)", func(t *testing.T) {
		data, _ := hex.DecodeString("01")
		destination := &values.BoolValue{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.BoolValue{Value: true}, destination)
	})

	t.Run("bool (false)", func(t *testing.T) {
		data, _ := hex.DecodeString("00")
		destination := &values.BoolValue{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.BoolValue{Value: false}, destination)
	})

	t.Run("u8", func(t *testing.T) {
		data, _ := hex.DecodeString("01")
		destination := &values.U8Value{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.U8Value{Value: 0x01}, destination)
	})

	t.Run("u16", func(t *testing.T) {
		data, _ := hex.DecodeString("4142")
		destination := &values.U16Value{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.U16Value{Value: 0x4142}, destination)
	})

	t.Run("u32", func(t *testing.T) {
		data, _ := hex.DecodeString("41424344")
		destination := &values.U32Value{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.U32Value{Value: 0x41424344}, destination)
	})

	t.Run("u64", func(t *testing.T) {
		data, _ := hex.DecodeString("4142434445464748")
		destination := &values.U64Value{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.U64Value{Value: 0x4142434445464748}, destination)
	})

	t.Run("u16, should err because it cannot read 2 bytes", func(t *testing.T) {
		data, _ := hex.DecodeString("01")
		destination := &values.U16Value{}

		err := codec.DecodeNested(data, destination)
		require.ErrorContains(t, err, "cannot read exactly 2 bytes")
	})

	t.Run("u32, should err because it cannot read 4 bytes", func(t *testing.T) {
		data, _ := hex.DecodeString("4142")
		destination := &values.U32Value{}

		err := codec.DecodeNested(data, destination)
		require.ErrorContains(t, err, "cannot read exactly 4 bytes")
	})

	t.Run("u64, should err because it cannot read 8 bytes", func(t *testing.T) {
		data, _ := hex.DecodeString("41424344")
		destination := &values.U64Value{}

		err := codec.DecodeNested(data, destination)
		require.ErrorContains(t, err, "cannot read exactly 8 bytes")
	})

	t.Run("bigInt", func(t *testing.T) {
		data, _ := hex.DecodeString("00000000")
		destination := &values.BigIntValue{}
		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.BigIntValue{Value: big.NewInt(0)}, destination)

		data, _ = hex.DecodeString("0000000101")
		destination = &values.BigIntValue{}
		err = codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.BigIntValue{Value: big.NewInt(1)}, destination)

		data, _ = hex.DecodeString("00000001ff")
		destination = &values.BigIntValue{}
		err = codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.BigIntValue{Value: big.NewInt(-1)}, destination)
	})

	t.Run("address", func(t *testing.T) {
		data, _ := hex.DecodeString("0139472eff6886771a982f3083da5d421f24c29181e63888228dc81ca60d69e1")

		destination := &values.AddressValue{}
		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.AddressValue{Value: data}, destination)
	})

	t.Run("address (bad)", func(t *testing.T) {
		data, _ := hex.DecodeString("0139472eff6886771a982f3083da5d42")

		destination := &values.AddressValue{}
		err := codec.DecodeNested(data, destination)
		require.ErrorContains(t, err, "cannot read exactly 32 bytes")
	})

	t.Run("string", func(t *testing.T) {
		data, _ := hex.DecodeString("00000000")
		destination := &values.StringValue{}
		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.StringValue{}, destination)

		data, _ = hex.DecodeString("00000003616263")
		destination = &values.StringValue{}
		err = codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.StringValue{Value: "abc"}, destination)
	})

	t.Run("bytes", func(t *testing.T) {
		data, _ := hex.DecodeString("00000000")
		destination := &values.BytesValue{}
		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.BytesValue{Value: []byte{}}, destination)

		data, _ = hex.DecodeString("00000003616263")
		destination = &values.BytesValue{}
		err = codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.BytesValue{Value: []byte{'a', 'b', 'c'}}, destination)
	})

	t.Run("struct", func(t *testing.T) {
		data, _ := hex.DecodeString("014142")

		destination := &values.StructValue{
			Fields: []values.Field{
				{
					Value: &values.U8Value{},
				},
				{
					Value: &values.U16Value{},
				},
			},
		}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.StructValue{
			Fields: []values.Field{
				{
					Value: &values.U8Value{Value: 0x01},
				},
				{
					Value: &values.U16Value{Value: 0x4142},
				},
			},
		}, destination)
	})

	t.Run("enum (discriminant == 0)", func(t *testing.T) {
		data, _ := hex.DecodeString("00")
		destination := &values.EnumValue{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.EnumValue{
			Discriminant: 0x00,
		}, destination)
	})

	t.Run("enum (discriminant != 0)", func(t *testing.T) {
		data, _ := hex.DecodeString("01")
		destination := &values.EnumValue{}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.EnumValue{
			Discriminant: 0x01,
		}, destination)
	})

	t.Run("enum with values.Fields", func(t *testing.T) {
		data, _ := hex.DecodeString("01014142")

		destination := &values.EnumValue{
			Fields: []values.Field{
				{
					Value: &values.U8Value{},
				},
				{
					Value: &values.U16Value{},
				},
			},
		}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.EnumValue{
			Discriminant: 0x01,
			Fields: []values.Field{
				{
					Value: &values.U8Value{Value: 0x01},
				},
				{
					Value: &values.U16Value{Value: 0x4142},
				},
			},
		}, destination)
	})

	t.Run("option with value", func(t *testing.T) {
		data, _ := hex.DecodeString("010008")

		destination := &values.OptionValue{
			Value: &values.U16Value{},
		}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.OptionValue{
			Value: &values.U16Value{Value: 8},
		}, destination)
	})

	t.Run("option without value", func(t *testing.T) {
		data, _ := hex.DecodeString("00")

		destination := &values.OptionValue{
			Value: &values.U16Value{},
		}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.OptionValue{
			Value: nil,
		}, destination)
	})

	t.Run("list", func(t *testing.T) {
		data, _ := hex.DecodeString("00000003000100020003")

		destination := &values.OutputListValue{
			ItemCreator: func() any { return &values.U16Value{} },
			Items:       []any{},
		}

		err := codec.DecodeNested(data, destination)
		require.NoError(t, err)
		require.Equal(t,
			[]any{
				&values.U16Value{Value: 1},
				&values.U16Value{Value: 2},
				&values.U16Value{Value: 3},
			}, destination.Items)
	})
}

func TestCodec_DecodeTopLevel(t *testing.T) {
	codec := NewCodec()

	t.Run("bool (true)", func(t *testing.T) {
		data, _ := hex.DecodeString("01")
		destination := &values.BoolValue{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.BoolValue{Value: true}, destination)
	})

	t.Run("bool (false)", func(t *testing.T) {
		data, _ := hex.DecodeString("")
		destination := &values.BoolValue{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.BoolValue{Value: false}, destination)
	})

	t.Run("u8", func(t *testing.T) {
		data, _ := hex.DecodeString("01")
		destination := &values.U8Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.U8Value{Value: 0x01}, destination)
	})

	t.Run("u16", func(t *testing.T) {
		data, _ := hex.DecodeString("02")
		destination := &values.U16Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.U16Value{Value: 0x0002}, destination)
	})

	t.Run("u32", func(t *testing.T) {
		data, _ := hex.DecodeString("03")
		destination := &values.U32Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.U32Value{Value: 0x00000003}, destination)
	})

	t.Run("u64", func(t *testing.T) {
		data, _ := hex.DecodeString("04")
		destination := &values.U64Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.U64Value{Value: 0x0000000000000004}, destination)
	})

	t.Run("u8, should err because decoded value is too large", func(t *testing.T) {
		data, _ := hex.DecodeString("4142")
		destination := &values.U8Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.ErrorContains(t, err, "decoded value is too large")
	})

	t.Run("u16, should err because decoded value is too large", func(t *testing.T) {
		data, _ := hex.DecodeString("41424344")
		destination := &values.U16Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.ErrorContains(t, err, "decoded value is too large")
	})

	t.Run("u32, should err because decoded value is too large", func(t *testing.T) {
		data, _ := hex.DecodeString("4142434445464748")
		destination := &values.U32Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.ErrorContains(t, err, "decoded value is too large")
	})

	t.Run("u64, should err because decoded value is too large", func(t *testing.T) {
		data, _ := hex.DecodeString("41424344454647489876")
		destination := &values.U64Value{}

		err := codec.DecodeTopLevel(data, destination)
		require.ErrorContains(t, err, "decoded value is too large")
	})

	t.Run("bigInt", func(t *testing.T) {
		data, _ := hex.DecodeString("")
		destination := &values.BigIntValue{}
		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.BigIntValue{Value: big.NewInt(0)}, destination)

		data, _ = hex.DecodeString("01")
		destination = &values.BigIntValue{}
		err = codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.BigIntValue{Value: big.NewInt(1)}, destination)

		data, _ = hex.DecodeString("ff")
		destination = &values.BigIntValue{}
		err = codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.BigIntValue{Value: big.NewInt(-1)}, destination)
	})

	t.Run("struct", func(t *testing.T) {
		data, _ := hex.DecodeString("014142")

		destination := &values.StructValue{
			Fields: []values.Field{
				{
					Value: &values.U8Value{},
				},
				{
					Value: &values.U16Value{},
				},
			},
		}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.StructValue{
			Fields: []values.Field{
				{
					Value: &values.U8Value{Value: 0x01},
				},
				{
					Value: &values.U16Value{Value: 0x4142},
				},
			},
		}, destination)
	})

	t.Run("enum (discriminant == 0)", func(t *testing.T) {
		data, _ := hex.DecodeString("")
		destination := &values.EnumValue{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.EnumValue{
			Discriminant: 0x00,
		}, destination)
	})

	t.Run("enum (discriminant != 0)", func(t *testing.T) {
		data, _ := hex.DecodeString("01")
		destination := &values.EnumValue{}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.EnumValue{
			Discriminant: 0x01,
		}, destination)
	})

	t.Run("enum with values.Fields", func(t *testing.T) {
		data, _ := hex.DecodeString("01014142")

		destination := &values.EnumValue{
			Fields: []values.Field{
				{
					Value: &values.U8Value{},
				},
				{
					Value: &values.U16Value{},
				},
			},
		}

		err := codec.DecodeTopLevel(data, destination)
		require.NoError(t, err)
		require.Equal(t, &values.EnumValue{
			Discriminant: 0x01,
			Fields: []values.Field{
				{
					Value: &values.U8Value{Value: 0x01},
				},
				{
					Value: &values.U16Value{Value: 0x4142},
				},
			},
		}, destination)
	})
}
