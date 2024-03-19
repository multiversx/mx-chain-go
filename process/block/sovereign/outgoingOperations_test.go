package sovereign

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	transactionData "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/errors"
	sovTests "github.com/multiversx/mx-chain-go/testscommon/sovereign"
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

func createArgs() ArgsOutgoingOperations {
	return ArgsOutgoingOperations{
		SubscribedEvents: createEvents(),
		DataCodec:        &sovTests.DataCodecMock{},
		TopicsChecker:    &sovTests.TopicsCheckerMock{},
	}
}

func TestNewOutgoingOperationsFormatter(t *testing.T) {
	t.Parallel()

	t.Run("no subscribed events, should return error", func(t *testing.T) {
		args := createArgs()
		args.SubscribedEvents = []SubscribedEvent{}
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, creator)
		require.Equal(t, errNoSubscribedEvent, err)
	})

	t.Run("nil data codec, should return error", func(t *testing.T) {
		args := createArgs()
		args.DataCodec = nil
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, creator)
		require.Equal(t, errors.ErrNilDataCodec, err)
	})

	t.Run("nil topics checker, should return error", func(t *testing.T) {
		args := createArgs()
		args.TopicsChecker = nil
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, creator)
		require.Equal(t, errors.ErrNilTopicsChecker, err)
	})

	t.Run("should work", func(t *testing.T) {
		args := createArgs()
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, err)
		require.False(t, creator.IsInterfaceNil())
	})
}

func createOutgoingOpsFormatter() *outgoingOperations {
	events := []SubscribedEvent{
		{
			Identifier: []byte("deposit"),
			Addresses: map[string]string{
				"addr1": "addr1",
				"addr2": "addr2",
			},
		},
	}

	args := ArgsOutgoingOperations{
		SubscribedEvents: events,
		DataCodec:        &sovTests.DataCodecMock{},
		TopicsChecker:    &sovTests.TopicsCheckerMock{},
	}
	opFormatter, _ := NewOutgoingOperationsFormatter(args)
	return opFormatter
}

func TestOutgoingOperations_CheckEvent(t *testing.T) {
	t.Parallel()

	t.Run("invalid identifier", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.SubscribedEvents = []SubscribedEvent{
			{
				Identifier: []byte(""),
				Addresses: map[string]string{
					"addr1": "addr1",
				},
			},
		}

		opFormatter, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, opFormatter)
		require.ErrorContains(t, err, "no subscribed identifier")
	})
	t.Run("no addresses", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.SubscribedEvents = []SubscribedEvent{
			{
				Identifier: []byte("identifier"),
				Addresses:  map[string]string{},
			},
		}

		opFormatter, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, opFormatter)
		require.ErrorContains(t, err, errNoSubscribedAddresses.Error())
	})
	t.Run("invalid address", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.SubscribedEvents = []SubscribedEvent{
			{
				Identifier: []byte("identifier"),
				Addresses: map[string]string{
					"addr1": "",
				},
			},
		}

		opFormatter, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, opFormatter)
		require.ErrorContains(t, err, errNoSubscribedAddresses.Error())
	})
}

func TestOutgoingOperations_CreateOutgoingTxsDataErrorCases(t *testing.T) {
	t.Parallel()

	logs := []*data.LogData{
		{
			LogHandler: &transactionData.Log{
				Address: nil,
				Events: []*transactionData.Event{
					{
						Address:    []byte("addr1"),
						Identifier: []byte("deposit"),
						Topics:     [][]byte{[]byte("topic1"), []byte("topic1"), []byte("topic1"), []byte("topic1"), []byte("topic1")},
						Data:       []byte("data"),
					},
				},
			},
			TxHash: "",
		},
	}

	t.Run("nil logs", func(t *testing.T) {
		t.Parallel()

		outgoingOpsFormatter := createOutgoingOpsFormatter()
		outgoingTxData, err := outgoingOpsFormatter.CreateOutgoingTxsData(nil)
		require.NoError(t, err)
		require.Equal(t, 0, len(outgoingTxData))
	})
	t.Run("deserialize token error", func(t *testing.T) {
		t.Parallel()

		outgoingOpsFormatter := createOutgoingOpsFormatter()
		errDeserializeTokenData := fmt.Errorf("deserialize token data error")
		outgoingOpsFormatter.dataCodec = &sovTests.DataCodecMock{
			DeserializeTokenDataCalled: func(_ []byte) (*sovereign.EsdtTokenData, error) {
				return nil, errDeserializeTokenData
			},
		}

		outgoingTxData, err := outgoingOpsFormatter.CreateOutgoingTxsData(logs)
		require.Nil(t, outgoingTxData)
		require.Equal(t, errDeserializeTokenData, err)
	})
	t.Run("deserialize event error", func(t *testing.T) {
		t.Parallel()

		outgoingOpsFormatter := createOutgoingOpsFormatter()
		errDeserializeEventData := fmt.Errorf("deserialize event data error")
		outgoingOpsFormatter.dataCodec = &sovTests.DataCodecMock{
			DeserializeEventDataCalled: func(data []byte) (*sovereign.EventData, error) {
				return nil, errDeserializeEventData
			},
		}

		outgoingTxData, err := outgoingOpsFormatter.CreateOutgoingTxsData(logs)
		require.Nil(t, outgoingTxData)
		require.Equal(t, errDeserializeEventData, err)
	})
	t.Run("serialize operation error", func(t *testing.T) {
		t.Parallel()

		outgoingOpsFormatter := createOutgoingOpsFormatter()
		errSerializeOperation := fmt.Errorf("serialize operation error")
		outgoingOpsFormatter.dataCodec = &sovTests.DataCodecMock{
			SerializeOperationCalled: func(operation sovereign.Operation) ([]byte, error) {
				return nil, errSerializeOperation
			},
		}

		outgoingTxData, err := outgoingOpsFormatter.CreateOutgoingTxsData(logs)
		require.Nil(t, outgoingTxData)
		require.Equal(t, errSerializeOperation, err)
	})
	t.Run("check validity error", func(t *testing.T) {
		t.Parallel()

		outgoingOpsFormatter := createOutgoingOpsFormatter()
		errInvalidTopics := fmt.Errorf("check topics error")
		outgoingOpsFormatter.topicsChecker = &sovTests.TopicsCheckerMock{
			CheckValidityCalled: func(topics [][]byte) error {
				return errInvalidTopics
			},
		}

		outgoingTxData, err := outgoingOpsFormatter.CreateOutgoingTxsData(logs)
		require.Nil(t, outgoingTxData)
		require.Equal(t, errInvalidTopics, err)
	})
}

func TestOutgoingOperations_CreateOutgoingTxData(t *testing.T) {
	t.Parallel()

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")
	addr3 := []byte("addr3")

	identifier1 := []byte("deposit")
	identifier2 := []byte("send")

	tokenData1 := []byte("tokenData1")
	topic1 := [][]byte{
		[]byte("deposit"),
		[]byte("rcv1"),
		[]byte("token1"),
		[]byte("nonce1"),
		tokenData1,
	}
	data1 := []byte("data1")

	evData := &sovereign.EventData{
		Nonce: 1,
		TransferData: &sovereign.TransferData{
			GasLimit: 20000000,
			Function: []byte("add"),
			Args:     [][]byte{big.NewInt(20000000).Bytes()},
		},
	}

	amount := new(big.Int)
	amount.SetString("123000000000000000000", 10)
	tokenData := sovereign.EsdtTokenData{
		TokenType: core.Fungible,
		Amount:    amount,
	}

	operationBytes := []byte("operationBytes")

	dataCodec := &sovTests.DataCodecMock{
		DeserializeEventDataCalled: func(data []byte) (*sovereign.EventData, error) {
			require.Equal(t, data1, data)

			return evData, nil
		},
		DeserializeTokenDataCalled: func(data []byte) (*sovereign.EsdtTokenData, error) {
			require.Equal(t, tokenData1, data)

			return &tokenData, nil
		},
		SerializeOperationCalled: func(operation sovereign.Operation) ([]byte, error) {
			require.Equal(t, evData, operation.Data)
			require.Equal(t, tokenData, operation.Tokens[0].Data)

			return operationBytes, nil
		},
	}

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

	args := ArgsOutgoingOperations{
		SubscribedEvents: events,
		DataCodec:        dataCodec,
		TopicsChecker:    &sovTests.TopicsCheckerMock{},
	}
	opFormatter, _ := NewOutgoingOperationsFormatter(args)

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
						Address:    []byte("addr4"),
						Identifier: identifier2,
						Topics:     topic1,
						Data:       data1,
					},
				},
			},
			TxHash: "",
		},
	}

	outgoingTxData, err := opFormatter.CreateOutgoingTxsData(logs)
	require.Nil(t, err)
	require.Equal(t, [][]byte{operationBytes}, outgoingTxData)
}
