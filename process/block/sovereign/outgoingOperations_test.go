package sovereign

import (
	"encoding/hex"
	"github.com/multiversx/mx-chain-core-go/data"
	transactionData "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-sdk-abi-incubator/golang/abi"
	"testing"

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

	codec := abi.NewDefaultCodec()
	serializer := abi.NewSerializer(codec)

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")
	addr3 := []byte("addr3")
	addr4 := []byte("addr4")

	identifier1 := []byte("deposit")
	identifier2 := []byte("send")

	topic1 := [][]byte{
		[]byte("endpoint1"),
		[]byte("rcv1"),
		[]byte("token1"),
		[]byte{0x00},
		[]byte{0x64},
	}
	dataStruct1 := abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "tx_nonce",
				Value: abi.U64Value{Value: 1},
			},
			{
				Name: "gas_limit",
				Value: abi.OptionValue{
					Value: abi.U64Value{Value: 50000},
				},
			},
			{
				Name: "function",
				Value: abi.OptionValue{
					Value: abi.BytesValue{Value: []byte("functionToCall1")},
				},
			},
			{
				Name: "args",
				Value: abi.OptionValue{
					Value: abi.InputListValue{
						Items: []any{
							abi.BytesValue{Value: []byte("arg1")},
							abi.BytesValue{Value: []byte("arg2")},
						},
					},
				},
			},
		},
	}
	encodedData1, err := serializer.Serialize([]any{dataStruct1})
	require.Nil(t, err)
	data1, err := hex.DecodeString(encodedData1)
	require.Nil(t, err)

	topic2 := [][]byte{
		[]byte("endpoint2"),
		[]byte("rcv2"),
		[]byte("token2"),
		[]byte{0x00},
		[]byte{0x32},
		[]byte("token3"),
		[]byte{0x00},
		[]byte{0x80},
	}
	dataStruct2 := abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "tx_nonce",
				Value: abi.U64Value{Value: 2},
			},
			{
				Name: "gas_limit",
				Value: abi.OptionValue{
					Value: abi.U64Value{Value: 40000},
				},
			},
			{
				Name: "function",
				Value: abi.OptionValue{
					Value: abi.BytesValue{Value: []byte("functionToCall2")},
				},
			},
			{
				Name: "args",
				Value: abi.OptionValue{
					Value: abi.InputListValue{
						Items: []any{
							abi.BytesValue{Value: []byte("arg2")},
						},
					},
				},
			},
		},
	}
	encodedData2, err := serializer.Serialize([]any{dataStruct2})
	require.Nil(t, err)
	data2, err := hex.DecodeString(encodedData2)
	require.Nil(t, err)

	topic3 := [][]byte{
		[]byte("endpoint3"),
		[]byte("rcv3"),
		[]byte("token4"),
		[]byte{0x00},
		[]byte{0x01},
	}
	dataStruct3 := abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "tx_nonce",
				Value: abi.U64Value{Value: 3},
			},
			{
				Name: "gas_limit",
				Value: abi.OptionValue{
					Value: abi.U64Value{Value: 55000},
				},
			},
			{
				Name: "function",
				Value: abi.OptionValue{
					Value: abi.BytesValue{Value: []byte("functionToCall3")},
				},
			},
			{
				Name: "args",
				Value: abi.OptionValue{
					Value: abi.InputListValue{
						Items: []any{
							abi.BytesValue{Value: []byte("arg3")},
							abi.BytesValue{Value: []byte("arg4")},
						},
					},
				},
			},
		},
	}
	encodedData3, err := serializer.Serialize([]any{dataStruct3})
	require.Nil(t, err)
	data3, err := hex.DecodeString(encodedData3)
	require.Nil(t, err)

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
						Address:    addr4,
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

	outgoingTxData := creator.CreateOutgoingTxsData(logs)

	operation1 := NewOperationData(topic1)
	evData1, _ := GetEventData(data1)
	operation1.transferData.gasLimit = evData1.gasLimit
	operation1.transferData.function = evData1.function
	operation1.transferData.args = evData1.args
	serializedOp1 := SerializeOperationData(operation1)

	operation2 := NewOperationData(topic2)
	evData2, _ := GetEventData(data2)
	operation2.transferData.gasLimit = evData2.gasLimit
	operation2.transferData.function = evData2.function
	operation2.transferData.args = evData2.args
	serializedOp2 := SerializeOperationData(operation2)

	operation3 := NewOperationData(topic3)
	evData3, _ := GetEventData(data3)
	operation3.transferData.gasLimit = evData3.gasLimit
	operation3.transferData.function = evData3.function
	operation3.transferData.args = evData3.args
	serializedOp3 := SerializeOperationData(operation3)

	txsData := make([]byte, 0)
	txsData = append(txsData, serializedOp1...)
	txsData = append(txsData, serializedOp2...)
	txsData = append(txsData, serializedOp3...)

	require.Equal(t, [][]byte{txsData}, outgoingTxData)
}
