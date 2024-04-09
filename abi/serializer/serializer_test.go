package serializer

import (
	"testing"

	"github.com/multiversx/mx-chain-go/abi/encoding"
	"github.com/multiversx/mx-chain-go/abi/values"
	"github.com/stretchr/testify/require"
)

func TestSerializer_Serialize(t *testing.T) {
	serializer := NewSerializer(encoding.NewCodec())

	t.Run("u8", func(t *testing.T) {
		data, err := serializer.Serialize([]any{
			values.U8Value{Value: 0x42},
		})

		require.NoError(t, err)
		require.Equal(t, "42", data)
	})

	t.Run("u16", func(t *testing.T) {
		data, err := serializer.Serialize([]any{
			values.U16Value{Value: 0x4243},
		})

		require.NoError(t, err)
		require.Equal(t, "4243", data)
	})

	t.Run("u8, u16", func(t *testing.T) {
		data, err := serializer.Serialize([]any{
			values.U8Value{Value: 0x42},
			values.U16Value{Value: 0x4243},
		})

		require.NoError(t, err)
		require.Equal(t, "42@4243", data)
	})

	t.Run("multi<u8, u16, u32>", func(t *testing.T) {
		data, err := serializer.Serialize([]any{
			values.InputMultiValue{
				Items: []any{
					values.U8Value{Value: 0x42},
					values.U16Value{Value: 0x4243},
					values.U32Value{Value: 0x42434445},
				},
			},
		})

		require.NoError(t, err)
		require.Equal(t, "42@4243@42434445", data)
	})

	t.Run("u8, multi<u8, u16, u32>", func(t *testing.T) {
		data, err := serializer.Serialize([]any{
			values.U8Value{Value: 0x42},
			values.InputMultiValue{
				Items: []any{
					values.U8Value{Value: 0x42},
					values.U16Value{Value: 0x4243},
					values.U32Value{Value: 0x42434445},
				},
			},
		})

		require.NoError(t, err)
		require.Equal(t, "42@42@4243@42434445", data)
	})

	t.Run("multi<multi<u8, u16>, multi<u8, u16>>", func(t *testing.T) {
		data, err := serializer.Serialize([]any{
			values.InputMultiValue{
				Items: []any{
					values.InputMultiValue{
						Items: []any{
							values.U8Value{Value: 0x42},
							values.U16Value{Value: 0x4243},
						},
					},
					values.InputMultiValue{
						Items: []any{
							values.U8Value{Value: 0x44},
							values.U16Value{Value: 0x4445},
						},
					},
				},
			},
		})

		require.NoError(t, err)
		require.Equal(t, "42@4243@44@4445", data)
	})

	t.Run("variadic, of different types", func(t *testing.T) {
		data, err := serializer.Serialize([]any{
			values.InputVariadicValues{
				Items: []any{
					values.U8Value{Value: 0x42},
					values.U16Value{Value: 0x4243},
				},
			},
		})

		// For now, the serializer does not perform such a strict type check.
		// Although doable, it would be slightly complex and, if done, might be even dropped in the future
		// (with respect to the decoder that is embedded in Rust-based smart contracts).
		require.Nil(t, err)
		require.Equal(t, "42@4243", data)
	})

	t.Run("variadic<u8>, u8: should err because variadic must be last", func(t *testing.T) {
		_, err := serializer.Serialize([]any{
			values.InputVariadicValues{
				Items: []any{
					values.U8Value{Value: 0x42},
					values.U8Value{Value: 0x43},
				},
			},
			values.U8Value{Value: 0x44},
		})

		require.ErrorContains(t, err, "variadic values must be last among input values")
	})

	t.Run("u8, variadic<u8>", func(t *testing.T) {
		data, err := serializer.Serialize([]any{
			values.U8Value{Value: 0x41},
			values.InputVariadicValues{
				Items: []any{
					values.U8Value{Value: 0x42},
					values.U8Value{Value: 0x43},
				},
			},
		})

		require.Nil(t, err)
		require.Equal(t, "41@42@43", data)
	})
}

func TestSerializer_Deserialize(t *testing.T) {
	serializer := NewSerializer(encoding.NewCodec())

	t.Run("nil destination", func(t *testing.T) {
		err := serializer.Deserialize("", []any{nil})
		require.ErrorContains(t, err, "cannot deserialize into nil value")
	})

	t.Run("u8", func(t *testing.T) {
		outputValues := []any{
			&values.U8Value{},
		}

		err := serializer.Deserialize("42", outputValues)

		require.Nil(t, err)
		require.Equal(t, []any{
			&values.U8Value{Value: 0x42},
		}, outputValues)
	})

	t.Run("u16", func(t *testing.T) {
		outputValues := []any{
			&values.U16Value{},
		}

		err := serializer.Deserialize("4243", outputValues)

		require.Nil(t, err)
		require.Equal(t, []any{
			&values.U16Value{Value: 0x4243},
		}, outputValues)
	})

	t.Run("u8, u16", func(t *testing.T) {
		outputValues := []any{
			&values.U8Value{},
			&values.U16Value{},
		}

		err := serializer.Deserialize("42@4243", outputValues)

		require.Nil(t, err)
		require.Equal(t, []any{
			&values.U8Value{Value: 0x42},
			&values.U16Value{Value: 0x4243},
		}, outputValues)
	})

	t.Run("multi<u8, u16, u32>", func(t *testing.T) {
		outputValues := []any{
			&values.OutputMultiValue{
				Items: []any{
					&values.U8Value{},
					&values.U16Value{},
					&values.U32Value{},
				},
			},
		}

		err := serializer.Deserialize("42@4243@42434445", outputValues)

		require.Nil(t, err)
		require.Equal(t, []any{
			&values.OutputMultiValue{
				Items: []any{
					&values.U8Value{Value: 0x42},
					&values.U16Value{Value: 0x4243},
					&values.U32Value{Value: 0x42434445},
				},
			},
		}, outputValues)
	})

	t.Run("u8, multi<u8, u16, u32>", func(t *testing.T) {
		outputValues := []any{
			&values.U8Value{},
			&values.OutputMultiValue{
				Items: []any{
					&values.U8Value{},
					&values.U16Value{},
					&values.U32Value{},
				},
			},
		}

		err := serializer.Deserialize("42@42@4243@42434445", outputValues)

		require.Nil(t, err)
		require.Equal(t, []any{
			&values.U8Value{Value: 0x42},
			&values.OutputMultiValue{
				Items: []any{
					&values.U8Value{Value: 0x42},
					&values.U16Value{Value: 0x4243},
					&values.U32Value{Value: 0x42434445},
				},
			},
		}, outputValues)
	})

	t.Run("variadic, should err because of nil item creator", func(t *testing.T) {
		destination := &values.OutputVariadicValues{
			Items: []any{},
		}

		err := serializer.Deserialize("", []any{destination})
		require.ErrorContains(t, err, "cannot deserialize variadic values: item creator is nil")
	})

	t.Run("empty: u8", func(t *testing.T) {
		destination := &values.OutputVariadicValues{
			Items:       []any{},
			ItemCreator: func() any { return &values.U8Value{} },
		}

		err := serializer.Deserialize("", []any{destination})
		require.NoError(t, err)
		require.Equal(t, []any{&values.U8Value{Value: 0}}, destination.Items)
	})

	t.Run("variadic<u8>", func(t *testing.T) {
		destination := &values.OutputVariadicValues{
			Items:       []any{},
			ItemCreator: func() any { return &values.U8Value{} },
		}

		err := serializer.Deserialize("2A@2B@2C", []any{destination})
		require.NoError(t, err)

		require.Equal(t, []any{
			&values.U8Value{Value: 42},
			&values.U8Value{Value: 43},
			&values.U8Value{Value: 44},
		}, destination.Items)
	})

	t.Run("varidic<u8>, with empty items", func(t *testing.T) {
		destination := &values.OutputVariadicValues{
			Items:       []any{},
			ItemCreator: func() any { return &values.U8Value{} },
		}

		err := serializer.Deserialize("@01@", []any{destination})
		require.NoError(t, err)

		require.Equal(t, []any{
			&values.U8Value{Value: 0},
			&values.U8Value{Value: 1},
			&values.U8Value{Value: 0},
		}, destination.Items)
	})

	t.Run("varidic<u32>", func(t *testing.T) {
		destination := &values.OutputVariadicValues{
			Items:       []any{},
			ItemCreator: func() any { return &values.U32Value{} },
		}

		err := serializer.Deserialize("AABBCCDD@DDCCBBAA", []any{destination})
		require.NoError(t, err)

		require.Equal(t, []any{
			&values.U32Value{Value: 0xAABBCCDD},
			&values.U32Value{Value: 0xDDCCBBAA},
		}, destination.Items)
	})

	t.Run("varidic<u8>, should err because decoded value is too large", func(t *testing.T) {
		destination := &values.OutputVariadicValues{
			Items:       []any{},
			ItemCreator: func() any { return &values.U8Value{} },
		}

		err := serializer.Deserialize("0100", []any{destination})
		require.ErrorContains(t, err, "cannot decode (top-level) *values.U8Value, because of: decoded value is too large: 256 > 255")
	})
}
