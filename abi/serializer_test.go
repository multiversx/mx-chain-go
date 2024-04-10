package abi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSerializer_Serialize(t *testing.T) {
	serializer, err := NewSerializer(ArgsNewSerializer{
		PartsSeparator: "@",
		PubKeyLength:   32,
	})
	require.NoError(t, err)

	t.Run("u8", func(t *testing.T) {
		data, err := serializer.Serialize([]any{
			U8Value{Value: 0x42},
		})

		require.NoError(t, err)
		require.Equal(t, "42", data)
	})

	t.Run("u16", func(t *testing.T) {
		data, err := serializer.Serialize([]any{
			U16Value{Value: 0x4243},
		})

		require.NoError(t, err)
		require.Equal(t, "4243", data)
	})

	t.Run("u8, u16", func(t *testing.T) {
		data, err := serializer.Serialize([]any{
			U8Value{Value: 0x42},
			U16Value{Value: 0x4243},
		})

		require.NoError(t, err)
		require.Equal(t, "42@4243", data)
	})

	t.Run("multi<u8, u16, u32>", func(t *testing.T) {
		data, err := serializer.Serialize([]any{
			InputMultiValue{
				Items: []any{
					U8Value{Value: 0x42},
					U16Value{Value: 0x4243},
					U32Value{Value: 0x42434445},
				},
			},
		})

		require.NoError(t, err)
		require.Equal(t, "42@4243@42434445", data)
	})

	t.Run("u8, multi<u8, u16, u32>", func(t *testing.T) {
		data, err := serializer.Serialize([]any{
			U8Value{Value: 0x42},
			InputMultiValue{
				Items: []any{
					U8Value{Value: 0x42},
					U16Value{Value: 0x4243},
					U32Value{Value: 0x42434445},
				},
			},
		})

		require.NoError(t, err)
		require.Equal(t, "42@42@4243@42434445", data)
	})

	t.Run("multi<multi<u8, u16>, multi<u8, u16>>", func(t *testing.T) {
		data, err := serializer.Serialize([]any{
			InputMultiValue{
				Items: []any{
					InputMultiValue{
						Items: []any{
							U8Value{Value: 0x42},
							U16Value{Value: 0x4243},
						},
					},
					InputMultiValue{
						Items: []any{
							U8Value{Value: 0x44},
							U16Value{Value: 0x4445},
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
			InputVariadicValues{
				Items: []any{
					U8Value{Value: 0x42},
					U16Value{Value: 0x4243},
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
			InputVariadicValues{
				Items: []any{
					U8Value{Value: 0x42},
					U8Value{Value: 0x43},
				},
			},
			U8Value{Value: 0x44},
		})

		require.ErrorContains(t, err, "variadic values must be last among input values")
	})

	t.Run("u8, variadic<u8>", func(t *testing.T) {
		data, err := serializer.Serialize([]any{
			U8Value{Value: 0x41},
			InputVariadicValues{
				Items: []any{
					U8Value{Value: 0x42},
					U8Value{Value: 0x43},
				},
			},
		})

		require.Nil(t, err)
		require.Equal(t, "41@42@43", data)
	})
}

func TestSerializer_Deserialize(t *testing.T) {
	serializer, err := NewSerializer(ArgsNewSerializer{
		PartsSeparator: "@",
		PubKeyLength:   32,
	})
	require.NoError(t, err)

	t.Run("nil destination", func(t *testing.T) {
		err := serializer.Deserialize("", []any{nil})
		require.ErrorContains(t, err, "cannot deserialize into nil value")
	})

	t.Run("u8", func(t *testing.T) {
		outputValues := []any{
			&U8Value{},
		}

		err := serializer.Deserialize("42", outputValues)

		require.Nil(t, err)
		require.Equal(t, []any{
			&U8Value{Value: 0x42},
		}, outputValues)
	})

	t.Run("u16", func(t *testing.T) {
		outputValues := []any{
			&U16Value{},
		}

		err := serializer.Deserialize("4243", outputValues)

		require.Nil(t, err)
		require.Equal(t, []any{
			&U16Value{Value: 0x4243},
		}, outputValues)
	})

	t.Run("u8, u16", func(t *testing.T) {
		outputValues := []any{
			&U8Value{},
			&U16Value{},
		}

		err := serializer.Deserialize("42@4243", outputValues)

		require.Nil(t, err)
		require.Equal(t, []any{
			&U8Value{Value: 0x42},
			&U16Value{Value: 0x4243},
		}, outputValues)
	})

	t.Run("multi<u8, u16, u32>", func(t *testing.T) {
		outputValues := []any{
			&OutputMultiValue{
				Items: []any{
					&U8Value{},
					&U16Value{},
					&U32Value{},
				},
			},
		}

		err := serializer.Deserialize("42@4243@42434445", outputValues)

		require.Nil(t, err)
		require.Equal(t, []any{
			&OutputMultiValue{
				Items: []any{
					&U8Value{Value: 0x42},
					&U16Value{Value: 0x4243},
					&U32Value{Value: 0x42434445},
				},
			},
		}, outputValues)
	})

	t.Run("u8, multi<u8, u16, u32>", func(t *testing.T) {
		outputValues := []any{
			&U8Value{},
			&OutputMultiValue{
				Items: []any{
					&U8Value{},
					&U16Value{},
					&U32Value{},
				},
			},
		}

		err := serializer.Deserialize("42@42@4243@42434445", outputValues)

		require.Nil(t, err)
		require.Equal(t, []any{
			&U8Value{Value: 0x42},
			&OutputMultiValue{
				Items: []any{
					&U8Value{Value: 0x42},
					&U16Value{Value: 0x4243},
					&U32Value{Value: 0x42434445},
				},
			},
		}, outputValues)
	})

	t.Run("variadic, should err because of nil item creator", func(t *testing.T) {
		destination := &OutputVariadicValues{
			Items: []any{},
		}

		err := serializer.Deserialize("", []any{destination})
		require.ErrorContains(t, err, "cannot deserialize variadic values: item creator is nil")
	})

	t.Run("empty: u8", func(t *testing.T) {
		destination := &OutputVariadicValues{
			Items:       []any{},
			ItemCreator: func() any { return &U8Value{} },
		}

		err := serializer.Deserialize("", []any{destination})
		require.NoError(t, err)
		require.Equal(t, []any{&U8Value{Value: 0}}, destination.Items)
	})

	t.Run("variadic<u8>", func(t *testing.T) {
		destination := &OutputVariadicValues{
			Items:       []any{},
			ItemCreator: func() any { return &U8Value{} },
		}

		err := serializer.Deserialize("2A@2B@2C", []any{destination})
		require.NoError(t, err)

		require.Equal(t, []any{
			&U8Value{Value: 42},
			&U8Value{Value: 43},
			&U8Value{Value: 44},
		}, destination.Items)
	})

	t.Run("varidic<u8>, with empty items", func(t *testing.T) {
		destination := &OutputVariadicValues{
			Items:       []any{},
			ItemCreator: func() any { return &U8Value{} },
		}

		err := serializer.Deserialize("@01@", []any{destination})
		require.NoError(t, err)

		require.Equal(t, []any{
			&U8Value{Value: 0},
			&U8Value{Value: 1},
			&U8Value{Value: 0},
		}, destination.Items)
	})

	t.Run("varidic<u32>", func(t *testing.T) {
		destination := &OutputVariadicValues{
			Items:       []any{},
			ItemCreator: func() any { return &U32Value{} },
		}

		err := serializer.Deserialize("AABBCCDD@DDCCBBAA", []any{destination})
		require.NoError(t, err)

		require.Equal(t, []any{
			&U32Value{Value: 0xAABBCCDD},
			&U32Value{Value: 0xDDCCBBAA},
		}, destination.Items)
	})

	t.Run("varidic<u8>, should err because decoded value is too large", func(t *testing.T) {
		destination := &OutputVariadicValues{
			Items:       []any{},
			ItemCreator: func() any { return &U8Value{} },
		}

		err := serializer.Deserialize("0100", []any{destination})
		require.ErrorContains(t, err, "cannot decode (top-level) *abi.U8Value, because of: decoded value is too large: 256 > 255")
	})
}
