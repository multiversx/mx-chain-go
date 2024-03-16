package sovereign

import (
	"encoding/hex"
	"fmt"
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

func createArgs() ArgsOutgoingOperations {
	return ArgsOutgoingOperations{
		SubscribedEvents: createEvents(),
		DataCodec:        &mock.DataCodecMock{},
		TopicsChecker:    &mock.TopicsCheckerMock{},
	}
}

func TestNewOutgoingOperationsFormatter(t *testing.T) {
	t.Parallel()

	t.Run("no subscribed events, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.SubscribedEvents = []SubscribedEvent{}
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, creator)
		require.Equal(t, errNoSubscribedEvent, err)
	})

	t.Run("nil data codec, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.DataCodec = nil
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, creator)
		require.Equal(t, errors.ErrNilDataCodec, err)
	})

	t.Run("nil topics checker, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.TopicsChecker = nil
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, creator)
		require.Equal(t, errors.ErrNilTopicsChecker, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, err)
		require.False(t, creator.IsInterfaceNil())
	})
}

func createValidArgs() *outgoingOperations {
	dataCodec := &mock.DataCodecMock{}
	topicsChecker := &mock.TopicsCheckerMock{
		CheckValidityCalled: func(topics [][]byte) error {
			return nil
		},
	}

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
		DataCodec:        dataCodec,
		TopicsChecker:    topicsChecker,
	}
	creator, _ := NewOutgoingOperationsFormatter(args)
	return creator
}

func TestOutgoingOperations_CheckEvent(t *testing.T) {
	t.Parallel()

	t.Run("invalid identifier", func(t *testing.T) {
		t.Parallel()

		events := []SubscribedEvent{
			{
				Identifier: []byte(""),
				Addresses: map[string]string{
					"addr1": "addr1",
				},
			},
		}

		args := ArgsOutgoingOperations{
			SubscribedEvents: events,
			DataCodec:        &mock.DataCodecMock{},
			TopicsChecker:    &mock.TopicsCheckerMock{},
		}
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, creator)
		require.ErrorContains(t, err, "no subscribed identifier")
	})
	t.Run("no addresses", func(t *testing.T) {
		t.Parallel()

		events := []SubscribedEvent{
			{
				Identifier: []byte("identifier"),
				Addresses:  map[string]string{},
			},
		}

		args := ArgsOutgoingOperations{
			SubscribedEvents: events,
			DataCodec:        &mock.DataCodecMock{},
			TopicsChecker:    &mock.TopicsCheckerMock{},
		}
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, creator)
		require.ErrorContains(t, err, errNoSubscribedAddresses.Error())
	})
	t.Run("invalid address", func(t *testing.T) {
		t.Parallel()

		events := []SubscribedEvent{
			{
				Identifier: []byte("identifier"),
				Addresses: map[string]string{
					"addr1": "",
				},
			},
		}

		args := ArgsOutgoingOperations{
			SubscribedEvents: events,
			DataCodec:        &mock.DataCodecMock{},
			TopicsChecker:    &mock.TopicsCheckerMock{},
		}
		creator, err := NewOutgoingOperationsFormatter(args)
		require.Nil(t, creator)
		require.ErrorContains(t, err, errNoSubscribedAddresses.Error())
	})
}

func TestOutgoingOperations_OutgoingEvents(t *testing.T) {
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

		outgoingOps := createValidArgs()

		outgoingTxData, err := outgoingOps.CreateOutgoingTxsData(nil)
		require.NoError(t, err)
		require.Equal(t, 0, len(outgoingTxData))
	})
	t.Run("deserialize token error", func(t *testing.T) {
		t.Parallel()

		outgoingOps := createValidArgs()

		outgoingOps.dataCodec = &mock.DataCodecMock{
			DeserializeTokenDataCalled: func(_ []byte) (*sovereign.EsdtTokenData, error) {
				return nil, fmt.Errorf("deserialize token data error")
			},
		}

		outgoingTxData, err := outgoingOps.CreateOutgoingTxsData(logs)
		require.Nil(t, outgoingTxData)
		require.Equal(t, "deserialize token data error", err.Error())
	})
	t.Run("deserialize event error", func(t *testing.T) {
		t.Parallel()

		outgoingOps := createValidArgs()

		outgoingOps.dataCodec = &mock.DataCodecMock{
			DeserializeEventDataCalled: func(data []byte) (*sovereign.EventData, error) {
				return nil, fmt.Errorf("deserialize event data error")
			},
			DeserializeTokenDataCalled: func(_ []byte) (*sovereign.EsdtTokenData, error) {
				return &sovereign.EsdtTokenData{}, nil
			},
		}

		outgoingTxData, err := outgoingOps.CreateOutgoingTxsData(logs)
		require.Nil(t, outgoingTxData)
		require.Equal(t, "deserialize event data error", err.Error())
	})
	t.Run("serialize operation error", func(t *testing.T) {
		t.Parallel()

		outgoingOps := createValidArgs()

		outgoingOps.dataCodec = &mock.DataCodecMock{
			DeserializeEventDataCalled: func(data []byte) (*sovereign.EventData, error) {
				return &sovereign.EventData{}, nil
			},
			DeserializeTokenDataCalled: func(_ []byte) (*sovereign.EsdtTokenData, error) {
				return &sovereign.EsdtTokenData{}, nil
			},
			SerializeOperationCalled: func(operation sovereign.Operation) ([]byte, error) {
				return nil, fmt.Errorf("serialize operation error")
			},
		}

		outgoingTxData, err := outgoingOps.CreateOutgoingTxsData(logs)
		require.Nil(t, outgoingTxData)
		require.Equal(t, "serialize operation error", err.Error())
	})
}

func TestOutgoingOperations_CheckValidity(t *testing.T) {
	t.Parallel()

	dataCodec := &mock.DataCodecMock{}
	topicsChecker := &mock.TopicsCheckerMock{
		CheckValidityCalled: func(topics [][]byte) error {
			return fmt.Errorf("invalid")
		},
	}

	events := []SubscribedEvent{
		{
			Identifier: []byte("deposit"),
			Addresses: map[string]string{
				"addr1": "addr1",
			},
		},
	}

	args := ArgsOutgoingOperations{
		SubscribedEvents: events,
		DataCodec:        dataCodec,
		TopicsChecker:    topicsChecker,
	}
	creator, _ := NewOutgoingOperationsFormatter(args)

	logs := []*data.LogData{
		{
			LogHandler: &transactionData.Log{
				Address: nil,
				Events: []*transactionData.Event{
					{
						Address:    []byte("addr1"),
						Identifier: []byte("deposit"),
						Topics:     [][]byte{[]byte("topic1")},
						Data:       []byte("data"),
					},
				},
			},
			TxHash: "",
		},
	}

	outgoingTxData, err := creator.CreateOutgoingTxsData(logs)
	require.Nil(t, outgoingTxData)
	require.Error(t, err, "invalid")
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

	addr0, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	amount := new(big.Int)
	amount.SetString("123000000000000000000", 10)
	tokenData := sovereign.EsdtTokenData{
		TokenType:  0,
		Amount:     amount,
		Frozen:     false,
		Hash:       make([]byte, 0),
		Name:       make([]byte, 0),
		Attributes: make([]byte, 0),
		Creator:    addr0,
		Royalties:  big.NewInt(0),
		Uris:       make([][]byte, 0),
	}

	operationBytes := []byte("operationBytes")

	dataCodec := &mock.DataCodecMock{
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

	topicsChecker := &mock.TopicsCheckerMock{
		CheckValidityCalled: func(topics [][]byte) error {
			return nil
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
		TopicsChecker:    topicsChecker,
	}
	creator, _ := NewOutgoingOperationsFormatter(args)

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

	outgoingTxData, err := creator.CreateOutgoingTxsData(logs)
	require.Nil(t, err)
	require.Equal(t, [][]byte{operationBytes}, outgoingTxData)
}
