package sovereign

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	transactionData "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon"
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
		creator, err := NewOutgoingOperationsFormatter([]SubscribedEvent{}, &testscommon.RoundHandlerMock{})
		require.Nil(t, creator)
		require.Equal(t, errNoSubscribedEvent, err)
	})

	t.Run("nil round handler, should return error", func(t *testing.T) {
		events := createEvents()
		creator, err := NewOutgoingOperationsFormatter(events, nil)
		require.Nil(t, creator)
		require.Equal(t, errors.ErrNilRoundHandler, err)
	})

	t.Run("should work", func(t *testing.T) {
		events := createEvents()
		creator, err := NewOutgoingOperationsFormatter(events, &testscommon.RoundHandlerMock{})
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

	roundHandler := &testscommon.RoundHandlerMock{
		IndexCalled: func() int64 {
			return 123
		},
	}

	creator, _ := NewOutgoingOperationsFormatter(events, roundHandler)
	topic1 := [][]byte{
		[]byte("rcv1"),
		[]byte("token1"),
		[]byte("nonce1"),
		[]byte("value1"),
	}
	data1 := []byte("functionToCall1@arg1@arg2@50000")

	topic2 := [][]byte{
		[]byte("rcv2"),
		[]byte("token2"),
		[]byte("nonce2"),
		[]byte("value2"),

		[]byte("token3"),
		[]byte("nonce3"),
		[]byte("value3"),
	}
	data2 := []byte("functionToCall2@arg2@40000")

	topic3 := [][]byte{
		[]byte("rcv3"),
		[]byte("token4"),
		[]byte("nonce4"),
		[]byte("value4"),
	}
	data3 := []byte("functionToCall3@arg3@arg4@55000")

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
					{
						Address:    addr3,
						Identifier: identifier2,
						Topics:     topic2,
						Data:       data2,
					},
					{
						Address:    []byte("addr4"),
						Identifier: identifier2,
						Topics:     topic1,
						Data:       data2,
					},
					{
						Address:    addr2,
						Identifier: identifier1,
						Topics:     topic3,
						Data:       data3,
					},
				},
			},
			TxHash: "",
		},
	}

	outgoingTxData := creator.CreateOutgoingTxData(logs)
	expectedTxData := []byte("bridgeOps@123@rcv1@token1@nonce1@functionToCall1@arg1@arg2@50000@rcv2@token2@nonce2@value2@token3@nonce3@functionToCall2@arg2@40000@rcv3@token4@nonce4@functionToCall3@arg3@arg4@55000")
	require.Equal(t, [][]byte{expectedTxData}, outgoingTxData)
}
