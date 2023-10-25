package sovereign

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	transactionData "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

var identifier = []byte("deposit")

func TestOutgoingOperations_CreateOutgoingTxData(t *testing.T) {
	addr1 := []byte("addr1")
	addr2 := []byte("addr2")
	addr3 := []byte("addr3")

	identifier2 := []byte("send")

	events := []SubscribedEvent{
		{
			Identifier: identifier,
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

	creator, _ := NewOutgoingOperationsFormatter(events)
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

	logs := []*data.LogData{
		{
			LogHandler: &transactionData.Log{
				Address: nil,
				Events: []*transactionData.Event{
					{
						Address:    addr1,
						Identifier: identifier,
						Topics:     topic1,
						Data:       data1,
					},
					{
						Address:    addr2,
						Identifier: identifier,
						Topics:     topic2,
						Data:       data2,
					},
					{
						Address:    []byte("addr4"),
						Identifier: identifier,
						Topics:     topic1,
						Data:       data2,
					},
				},
			},
			TxHash: "",
		},
	}

	outgoingTxData := creator.CreateOutgoingTxData(logs)
	expectedTxData := []byte("bridgeOps@rcv1@token1@nonce1@functionToCall1@arg1@arg2@50000@rcv2@token2@nonce2@value2@token3@nonce3@functionToCall2@arg2@40000")
	require.Equal(t, expectedTxData, outgoingTxData)
}
