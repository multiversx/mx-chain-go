package dataCodec

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/errors"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-go/abi"
	"github.com/stretchr/testify/require"
)

func createDataCodec() SovereignDataDecoder {
	codec := abi.NewCodec()
	args := ArgsDataCodec{
		Serializer: abi.NewSerializer(codec),
	}

	dtaCodec, _ := NewDataCodec(args)
	return dtaCodec
}

func TestNewDataCodec(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		codec := abi.NewCodec()
		args := ArgsDataCodec{
			Serializer: abi.NewSerializer(codec),
		}
		abiCodec, err := NewDataCodec(args)
		require.Nil(t, err)
		require.False(t, abiCodec.IsInterfaceNil())
	})
	t.Run("nil serializer should error", func(t *testing.T) {
		t.Parallel()

		args := ArgsDataCodec{
			Serializer: nil,
		}
		abiCodec, err := NewDataCodec(args)
		require.ErrorIs(t, errors.ErrNilSerializer, err)
		require.True(t, abiCodec.IsInterfaceNil())
	})
}

func TestDataCodec_EventDataCodec(t *testing.T) {
	t.Parallel()

	abiCodec := createDataCodec()

	t.Run("deserialize empty data should fail", func(t *testing.T) {
		t.Parallel()

		deserialized, err := abiCodec.DeserializeEventData(nil)
		require.Nil(t, deserialized)
		require.Equal(t, errEmptyData, err)
	})
	t.Run("nil transfer data should work", func(t *testing.T) {
		t.Parallel()

		sender, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
		eventData := sovereign.EventData{
			Nonce:        10,
			Sender:       sender,
			TransferData: nil,
		}

		serialized, err := abiCodec.SerializeEventData(eventData)
		require.Nil(t, err)
		require.Equal(t, "000000000000000a000000000000000000000000000000000000000000000000000000000000000000", hex.EncodeToString(serialized))

		deserialized, err := abiCodec.DeserializeEventData(serialized)
		require.Nil(t, err)
		require.Equal(t, &eventData, deserialized)
	})
	t.Run("defined transfer data should work", func(t *testing.T) {
		t.Parallel()

		sender, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
		eventData := sovereign.EventData{
			Nonce:  10,
			Sender: sender,
			TransferData: &sovereign.TransferData{
				GasLimit: 20000000,
				Function: []byte("add"),
				Args:     [][]byte{big.NewInt(20000000).Bytes()},
			},
		}

		serialized, err := abiCodec.SerializeEventData(eventData)
		require.Nil(t, err)
		require.Equal(t, "000000000000000a0000000000000000000000000000000000000000000000000000000000000000010000000001312d0000000003616464000000010000000401312d00", hex.EncodeToString(serialized))

		deserialized, err := abiCodec.DeserializeEventData(serialized)
		require.Nil(t, err)
		require.Equal(t, &eventData, deserialized)
	})
	t.Run("defined transfer data empty arg should work", func(t *testing.T) {
		t.Parallel()

		sender, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
		eventData := sovereign.EventData{
			Nonce:  10,
			Sender: sender,
			TransferData: &sovereign.TransferData{
				GasLimit: 20000000,
				Function: []byte("add"),
				Args:     [][]byte{big.NewInt(20000000).Bytes(), make([]byte, 0)},
			},
		}

		serialized, err := abiCodec.SerializeEventData(eventData)
		require.Nil(t, err)
		require.Equal(t, "000000000000000a0000000000000000000000000000000000000000000000000000000000000000010000000001312d0000000003616464000000020000000401312d0000000000", hex.EncodeToString(serialized))

		deserialized, err := abiCodec.DeserializeEventData(serialized)
		require.Nil(t, err)
		require.Equal(t, &eventData, deserialized)
	})
	t.Run("deserialize event data gasLimit 0 should work", func(t *testing.T) {
		t.Parallel()

		sender, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
		eventData := sovereign.EventData{
			Nonce:  10,
			Sender: sender,
			TransferData: &sovereign.TransferData{
				GasLimit: 0,
				Function: []byte("add"),
				Args:     [][]byte{big.NewInt(20000000).Bytes(), make([]byte, 0)},
			},
		}

		serialized, err := abiCodec.SerializeEventData(eventData)
		require.Nil(t, err)
		require.Equal(t, "000000000000000a000000000000000000000000000000000000000000000000000000000000000001000000000000000000000003616464000000020000000401312d0000000000", hex.EncodeToString(serialized))

		deserialized, err := abiCodec.DeserializeEventData(serialized)
		require.Nil(t, err)
		require.Equal(t, &eventData, deserialized)
	})
}

func TestDataCodec_SerializeTokenData(t *testing.T) {
	t.Parallel()

	abiCodec := createDataCodec()

	addr0, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	amount := new(big.Int)
	amount.SetString("123000000000000000000", 10)

	esdtTokenData := sovereign.EsdtTokenData{
		TokenType:  core.Fungible,
		Amount:     amount,
		Frozen:     false,
		Hash:       make([]byte, 0),
		Name:       []byte(""),
		Attributes: make([]byte, 0),
		Creator:    addr0,
		Royalties:  big.NewInt(0),
		Uris:       make([][]byte, 0),
	}

	serialized, err := abiCodec.SerializeTokenData(esdtTokenData)
	require.Nil(t, err)
	require.Equal(t, "000000000906aaf7c8516d0c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", hex.EncodeToString(serialized))
}

func TestDataCodec_DeserializeTokenData(t *testing.T) {
	t.Parallel()

	abiCodec := createDataCodec()

	t.Run("empty data should fail", func(t *testing.T) {
		t.Parallel()

		deserialized, err := abiCodec.DeserializeTokenData(nil)
		require.Nil(t, deserialized)
		require.Equal(t, errEmptyTokenData, err)
	})
	t.Run("token data should work", func(t *testing.T) {
		t.Parallel()

		addr0, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
		amount := new(big.Int)
		amount.SetString("123000000000000000000", 10)

		expectedEsdtTokenData := &sovereign.EsdtTokenData{
			TokenType:  core.Fungible,
			Amount:     amount,
			Frozen:     false,
			Hash:       make([]byte, 0),
			Name:       []byte(""),
			Attributes: make([]byte, 0),
			Creator:    addr0,
			Royalties:  big.NewInt(0),
			Uris:       make([][]byte, 0),
		}

		tokenData, err := hex.DecodeString("000000000906aaf7c8516d0c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		require.Nil(t, err)

		deserialized, err := abiCodec.DeserializeTokenData(tokenData)
		require.Nil(t, err)
		require.Equal(t, expectedEsdtTokenData, deserialized)
	})
}

func TestDataCodec_SerializeOperation(t *testing.T) {
	t.Parallel()

	abiCodec := createDataCodec()

	t.Run("full operation", func(t *testing.T) {
		t.Parallel()

		receiver, _ := hex.DecodeString("c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6")
		creator, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
		amount := new(big.Int)
		amount.SetString("123000000000000000000", 10)

		token1 := sovereign.EsdtToken{
			Identifier: []byte("SVN-123456"),
			Nonce:      0,
			Data: sovereign.EsdtTokenData{
				TokenType:  core.Fungible,
				Amount:     amount,
				Frozen:     false,
				Hash:       []byte("hash"),
				Name:       []byte("SVN"),
				Attributes: []byte("attr"),
				Creator:    creator,
				Royalties:  big.NewInt(10000),
				Uris:       [][]byte{[]byte("url1")},
			},
		}
		token2 := sovereign.EsdtToken{
			Identifier: []byte("SVN-654321"),
			Nonce:      0,
			Data: sovereign.EsdtTokenData{
				TokenType:  core.Fungible,
				Amount:     amount,
				Frozen:     false,
				Hash:       []byte(""),
				Name:       []byte(""),
				Attributes: []byte(""),
				Creator:    creator,
				Royalties:  big.NewInt(0),
				Uris:       [][]byte{},
			},
		}
		tokens := make([]sovereign.EsdtToken, 0)
		tokens = append(tokens, token1)
		tokens = append(tokens, token2)

		sender, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
		operationData := &sovereign.EventData{
			Nonce:  100,
			Sender: sender,
			TransferData: &sovereign.TransferData{
				GasLimit: 20000000,
				Function: []byte("add"),
				Args:     [][]byte{big.NewInt(20000000).Bytes()},
			},
		}

		operation := sovereign.Operation{
			Address: receiver,
			Tokens:  tokens,
			Data:    operationData,
		}

		serialized, err := abiCodec.SerializeOperation(operation)
		require.Nil(t, err)
		require.Equal(t, "c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6000000020000000a53564e2d3132333435360000000000000000000000000906aaf7c8516d0c00000000000004686173680000000353564e00000004617474720000000000000000000000000000000000000000000000000000000000000000000000022710000000010000000475726c310000000a53564e2d3635343332310000000000000000000000000906aaf7c8516d0c0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000640000000000000000000000000000000000000000000000000000000000000000010000000001312d0000000003616464000000010000000401312d00", hex.EncodeToString(serialized))
	})
	t.Run("operation with nil transfer data", func(t *testing.T) {
		t.Parallel()

		receiver, _ := hex.DecodeString("c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6")
		addr, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
		amount := new(big.Int)
		amount.SetString("123000000000000000000", 10)

		token1 := sovereign.EsdtToken{
			Identifier: []byte("SVN-123456"),
			Nonce:      0,
			Data: sovereign.EsdtTokenData{
				TokenType:  core.Fungible,
				Amount:     amount,
				Frozen:     false,
				Hash:       []byte("hash"),
				Name:       []byte("SVN1"),
				Attributes: []byte("attr"),
				Creator:    addr,
				Royalties:  big.NewInt(7000),
				Uris:       [][]byte{[]byte("url1")},
			},
		}
		tokens := make([]sovereign.EsdtToken, 0)
		tokens = append(tokens, token1)

		sender, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
		operationData := &sovereign.EventData{
			Nonce:        10,
			Sender:       sender,
			TransferData: nil,
		}

		operation := sovereign.Operation{
			Address: receiver,
			Tokens:  tokens,
			Data:    operationData,
		}

		serialized, err := abiCodec.SerializeOperation(operation)
		require.Nil(t, err)
		require.Equal(t, "c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6000000010000000a53564e2d3132333435360000000000000000000000000906aaf7c8516d0c00000000000004686173680000000453564e3100000004617474720000000000000000000000000000000000000000000000000000000000000000000000021b58000000010000000475726c31000000000000000a000000000000000000000000000000000000000000000000000000000000000000", hex.EncodeToString(serialized))
	})
}
