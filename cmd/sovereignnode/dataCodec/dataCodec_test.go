package dataCodec

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"

	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-sdk-abi-incubator/golang/abi"
	"github.com/stretchr/testify/require"
)

func createDataCodec() DataCodecProcessor {
	codec := abi.NewDefaultCodec()
	args := DataCodec{
		Serializer: abi.NewSerializer(codec),
		Marshaller: &marshallerMock.MarshalizerMock{},
	}

	dataCodecMock, _ := NewDataCodec(args)

	return dataCodecMock
}

func TestNewDataCodec(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		codec := abi.NewDefaultCodec()
		args := DataCodec{
			Serializer: abi.NewSerializer(codec),
			Marshaller: &marshallerMock.MarshalizerMock{},
		}
		dataCodec, err := NewDataCodec(args)
		require.Nil(t, err)
		require.False(t, dataCodec.IsInterfaceNil())
	})
	t.Run("nil serializer should error", func(t *testing.T) {
		t.Parallel()

		args := DataCodec{
			Serializer: nil,
			Marshaller: &marshallerMock.MarshalizerMock{},
		}
		dataCodec, err := NewDataCodec(args)
		require.ErrorIs(t, errors.ErrNilSerializer, err)
		require.True(t, dataCodec.IsInterfaceNil())
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		codec := abi.NewDefaultCodec()
		args := DataCodec{
			Serializer: abi.NewSerializer(codec),
			Marshaller: nil,
		}
		dataCodec, err := NewDataCodec(args)
		require.ErrorIs(t, errors.ErrNilMarshalizer, err)
		require.True(t, dataCodec.IsInterfaceNil())
	})
}

func TestDataCodec_SerializeEventData(t *testing.T) {
	t.Parallel()

	dataCodec := createDataCodec()

	t.Run("nil transfer data should work", func(t *testing.T) {
		t.Parallel()

		eventData := sovereign.EventData{
			Nonce:        10,
			TransferData: nil,
		}

		serialized, err := dataCodec.SerializeEventData(eventData)
		require.Nil(t, err)
		require.Equal(t, "000000000000000a000000", hex.EncodeToString(serialized))
	})
	t.Run("defined transfer data should work", func(t *testing.T) {
		t.Parallel()

		eventData := sovereign.EventData{
			Nonce: 10,
			TransferData: &sovereign.TransferData{
				GasLimit: 20000000,
				Function: []byte("add"),
				Args:     [][]byte{big.NewInt(20000000).Bytes()},
			},
		}

		serialized, err := dataCodec.SerializeEventData(eventData)
		require.Nil(t, err)
		require.Equal(t, "000000000000000a010000000001312d00010000000361646401000000010000000401312d00", hex.EncodeToString(serialized))
	})

}

func TestDataCodec_DeserializeEventData(t *testing.T) {
	t.Parallel()

	dataCodec := createDataCodec()

	t.Run("empty data should fail", func(t *testing.T) {
		t.Parallel()

		deserialized, err := dataCodec.DeserializeEventData(nil)
		require.Nil(t, deserialized)
		require.Equal(t, errEmptyData, err)
	})
	t.Run("deserialize event data should work", func(t *testing.T) {
		t.Parallel()

		expectedEventData := &sovereign.EventData{
			Nonce: 10,
			TransferData: &sovereign.TransferData{
				GasLimit: 20000000,
				Function: []byte("add"),
				Args:     [][]byte{big.NewInt(20000000).Bytes()},
			},
		}
		eventData, err := hex.DecodeString("000000000000000a010000000001312d00010000000361646401000000010000000401312d00")
		require.Nil(t, err)

		deserialized, err := dataCodec.DeserializeEventData(eventData)
		require.Nil(t, err)
		require.Equal(t, expectedEventData, deserialized)
	})
}

func TestDataCodec_SerializeTokenData(t *testing.T) {
	t.Parallel()

	dataCodec := createDataCodec()

	addr0, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	amount := new(big.Int)
	amount.SetString("123000000000000000000", 10)

	esdtTokenData := sovereign.EsdtTokenData{
		TokenType:  0,
		Amount:     amount,
		Frozen:     false,
		Hash:       make([]byte, 0),
		Name:       []byte(""),
		Attributes: make([]byte, 0),
		Creator:    addr0,
		Royalties:  big.NewInt(0),
		Uris:       make([][]byte, 0),
	}

	serialized, err := dataCodec.SerializeTokenData(esdtTokenData)
	require.Nil(t, err)
	require.Equal(t, "000000000906aaf7c8516d0c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", hex.EncodeToString(serialized))
}

func TestDataCodec_DeserializeTokenData(t *testing.T) {
	t.Parallel()

	dataCodec := createDataCodec()

	t.Run("empty data should fail", func(t *testing.T) {
		t.Parallel()

		deserialized, err := dataCodec.DeserializeTokenData(nil)
		require.Nil(t, deserialized)
		require.Equal(t, errEmptyTokenData, err)
	})
	t.Run("empty data should fail", func(t *testing.T) {
		t.Parallel()

		addr0, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
		amount := new(big.Int)
		amount.SetString("123000000000000000000", 10)

		expectedEsdtTokenData := &sovereign.EsdtTokenData{
			TokenType:  0,
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

		deserialized, err := dataCodec.DeserializeTokenData(tokenData)
		require.Nil(t, err)
		require.Equal(t, expectedEsdtTokenData, deserialized)
	})
}

func TestDataCodec_GetTokenDataBytes(t *testing.T) {
	t.Parallel()

	dataCodec := createDataCodec()
	marshaller := marshallerMock.MarshalizerMock{}

	t.Run("empty token data should fail", func(t *testing.T) {
		t.Parallel()

		tokenDataBytes, err := dataCodec.GetTokenDataBytes(nil, nil)
		require.Nil(t, tokenDataBytes)
		require.Equal(t, errEmptyTokenData, err)
	})
	t.Run("fungible token should return only amount", func(t *testing.T) {
		t.Parallel()

		addr0, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
		amount := new(big.Int)
		amount.SetString("123000000000000000000", 10)
		tokenNonce := make([]byte, 0)

		esdtTokenData := sovereign.EsdtTokenData{
			TokenType:  0,
			Amount:     amount,
			Frozen:     false,
			Hash:       make([]byte, 0),
			Name:       []byte(""),
			Attributes: make([]byte, 0),
			Creator:    addr0,
			Royalties:  big.NewInt(0),
			Uris:       make([][]byte, 0),
		}

		serialized, err := dataCodec.SerializeTokenData(esdtTokenData)
		require.Nil(t, err)

		tokenDataBytes, err := dataCodec.GetTokenDataBytes(tokenNonce, serialized)
		require.Nil(t, err)
		require.Equal(t, amount.Bytes(), tokenDataBytes)
	})
	t.Run("non-fungible token should return marshalled digital token", func(t *testing.T) {
		t.Parallel()

		creator, _ := hex.DecodeString("c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6")
		hash, _ := hex.DecodeString("516d515a534548567a6d444674325463685341486b6d7661435565754e61705171687a76444852746b4a514e4d4c")
		name, _ := hex.DecodeString("534e4654202331")
		nonce := []byte{0x01}
		attributes, _ := hex.DecodeString("746167733a736e66742c6e6674313b6d657461646174613a516d4e70726973543437567636787854455a3445444b385151695056706d6e333351774867613657776e456a4d66")
		uri, _ := hex.DecodeString("68747470733a2f2f697066732e696f2f697066732f516d515a534548567a6d444674325463685341486b6d7661435565754e61705171687a76444852746b4a514e4d4c")

		esdtData := sovereign.EsdtTokenData{
			TokenType:  1,
			Amount:     big.NewInt(1),
			Frozen:     false,
			Hash:       hash,
			Name:       name,
			Attributes: attributes,
			Creator:    creator,
			Royalties:  big.NewInt(2500),
			Uris:       [][]byte{uri},
		}

		serialized, err := dataCodec.SerializeTokenData(esdtData)
		require.Nil(t, err)

		tokenDataBytes, err := dataCodec.GetTokenDataBytes(nonce, serialized)
		require.Nil(t, err)

		digitalToken := &esdt.ESDigitalToken{
			Type:  uint32(esdtData.TokenType),
			Value: esdtData.Amount,
			TokenMetaData: &esdt.MetaData{
				Nonce:      bytesToUint64(nonce),
				Name:       esdtData.Name,
				Creator:    esdtData.Creator,
				Royalties:  uint32(esdtData.Royalties.Uint64()),
				Hash:       esdtData.Hash,
				URIs:       esdtData.Uris,
				Attributes: esdtData.Attributes,
			},
		}
		digitalTokenMarshalled, _ := marshaller.Marshal(digitalToken)

		require.Equal(t, digitalTokenMarshalled, tokenDataBytes)
	})
}

func TestDataCodec_SerializeOperation(t *testing.T) {
	t.Parallel()

	dataCodec := createDataCodec()

	t.Run("full operation", func(t *testing.T) {
		t.Parallel()

		addr, _ := hex.DecodeString("c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6")
		amount := new(big.Int)
		amount.SetString("123000000000000000000", 10)

		token1 := sovereign.EsdtToken{
			Identifier: []byte("SVN-123456"),
			Nonce:      0,
			Data: sovereign.EsdtTokenData{
				TokenType:  0,
				Amount:     amount,
				Frozen:     false,
				Hash:       []byte("hash"),
				Name:       []byte("SVN1"),
				Attributes: []byte("attr"),
				Creator:    addr,
				Royalties:  big.NewInt(10000),
				Uris:       [][]byte{[]byte("url1")},
			},
		}
		token2 := sovereign.EsdtToken{
			Identifier: []byte("SVN-654321"),
			Nonce:      0,
			Data: sovereign.EsdtTokenData{
				TokenType:  0,
				Amount:     amount,
				Frozen:     false,
				Hash:       []byte("hash"),
				Name:       []byte("SVN2"),
				Attributes: []byte("attr"),
				Creator:    addr,
				Royalties:  big.NewInt(5000),
				Uris:       [][]byte{[]byte("url2")},
			},
		}
		tokens := make([]sovereign.EsdtToken, 0)
		tokens = append(tokens, token1)
		tokens = append(tokens, token2)

		transferData := &sovereign.TransferData{
			GasLimit: 20000000,
			Function: []byte("add"),
			Args:     [][]byte{big.NewInt(20000000).Bytes()},
		}

		operation := sovereign.Operation{
			Address:      addr,
			Tokens:       tokens,
			TransferData: transferData,
		}

		serialized, err := dataCodec.SerializeOperation(operation)
		require.Nil(t, err)
		require.Equal(t, "c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6000000020000000a53564e2d3132333435360000000000000000000000000906aaf7c8516d0c00000000000004686173680000000453564e310000000461747472c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6000000022710000000010000000475726c310000000a53564e2d3635343332310000000000000000000000000906aaf7c8516d0c00000000000004686173680000000453564e320000000461747472c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6000000021388000000010000000475726c32010000000001312d0000000003616464000000010000000401312d00", hex.EncodeToString(serialized))
	})
	t.Run("operation with nil transfer data", func(t *testing.T) {
		t.Parallel()

		addr, _ := hex.DecodeString("c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6")
		amount := new(big.Int)
		amount.SetString("123000000000000000000", 10)

		token1 := sovereign.EsdtToken{
			Identifier: []byte("SVN-123456"),
			Nonce:      0,
			Data: sovereign.EsdtTokenData{
				TokenType:  0,
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

		operation := sovereign.Operation{
			Address:      addr,
			Tokens:       tokens,
			TransferData: nil,
		}

		serialized, err := dataCodec.SerializeOperation(operation)
		require.Nil(t, err)
		require.Equal(t, "c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6000000010000000a53564e2d3132333435360000000000000000000000000906aaf7c8516d0c00000000000004686173680000000453564e310000000461747472c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6000000021b58000000010000000475726c3100", hex.EncodeToString(serialized))
	})
}
