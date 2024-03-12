package sovereign

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	transactionData "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/stretchr/testify/require"
)

func createEvents() []SubscribedEvent {
	return []SubscribedEvent{
		{
			Identifier: []byte("id"),
			Addresses: map[string]string{
				"decodedAddr": "encodedAddr",
			},
		},
	}
}

func TestNewOutgoingOperationsFormatter(t *testing.T) {
	t.Parallel()

	t.Run("no subscribed events, should return error", func(t *testing.T) {
		creator, err := NewOutgoingOperationsFormatter([]SubscribedEvent{}, &mock.DataCodecMock{})
		require.Nil(t, creator)
		require.Equal(t, errNoSubscribedEvent, err)
	})

	t.Run("nil data codec, should return error", func(t *testing.T) {
		events := createEvents()
		creator, err := NewOutgoingOperationsFormatter(events, nil)
		require.Nil(t, creator)
		require.Equal(t, errors.ErrNilDataCodec, err)
	})

	t.Run("should work", func(t *testing.T) {
		events := createEvents()
		creator, err := NewOutgoingOperationsFormatter(events, &mock.DataCodecMock{})
		require.Nil(t, err)
		require.False(t, creator.IsInterfaceNil())
	})
}

func TestOutgoingOperations_CreateOutgoingTxData(t *testing.T) {
	t.Parallel()

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")
	addr3 := []byte("addr3")

	identifier1 := []byte("deposit")
	identifier2 := []byte("send")

	events := []SubscribedEvent{
		{
			Identifier: identifier1,
			Addresses: map[string]string{
				string(addr1): string(addr1),
				string(addr2): string(addr2),
			},
		},
		{
			Identifier: identifier2,
			Addresses: map[string]string{
				string(addr3): string(addr3),
			},
		},
	}

	dataCodec := &mock.DataCodecMock{
		DeserializeEventDataCalled: func(data []byte) (*sovereign.EventData, error) {
			return &sovereign.EventData{
				Nonce: 1,
				TransferData: &sovereign.TransferData{
					GasLimit: 20000000,
					Function: []byte("add"),
					Args:     [][]byte{big.NewInt(20000000).Bytes()},
				},
			}, nil
		},
		DeserializeTokenDataCalled: func(data []byte) (*sovereign.EsdtTokenData, error) {
			addr0, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
			amount := new(big.Int)
			amount.SetString("123000000000000000000", 10)

			return &sovereign.EsdtTokenData{
				TokenType:  0,
				Amount:     amount,
				Frozen:     false,
				Hash:       make([]byte, 0),
				Name:       make([]byte, 0),
				Attributes: make([]byte, 0),
				Creator:    addr0,
				Royalties:  big.NewInt(0),
				Uris:       make([][]byte, 0),
			}, nil
		},
		SerializeOperationCalled: func(operation sovereign.Operation) ([]byte, error) {
			operationBytes, _ := hex.DecodeString("c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f60000000100000006746f6b656e310000000000000000000000000906aaf7c8516d0c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000001312d0000000003616464000000010000000401312d00")
			return operationBytes, nil
		},
	}

	creator, _ := NewOutgoingOperationsFormatter(events, dataCodec)

	addr, _ := hex.DecodeString("c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f6")
	tokenData, _ := hex.DecodeString("000000000906aaf7c8516d0c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	topic1 := [][]byte{
		[]byte("deposit"),
		addr,
		[]byte("token1"),
		make([]byte, 0),
		tokenData,
	}
	data1, _ := hex.DecodeString("000000000000000a010000000001312d00010000000361646401000000010000000401312d00")

	logs := []*data.LogData{
		{
			LogHandler: &transactionData.Log{
				Address: nil,
				Events: []*transactionData.Event{
					{
						Address:    addr1,
						Identifier: identifier1,
						Topics:     topic1,
						Data:       data1,
					},
				},
			},
			TxHash: "",
		},
	}

	outgoingTxData, err := creator.CreateOutgoingTxsData(logs)
	require.Nil(t, err)
	expectedTxData, _ := hex.DecodeString("c0c0739e0cf6232a934d2e56cfcd10881eb1c7336f128fc155a4a84292cfe7f60000000100000006746f6b656e310000000000000000000000000906aaf7c8516d0c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000001312d0000000003616464000000010000000401312d00")
	require.Equal(t, [][]byte{expectedTxData}, outgoingTxData)
}
