package incomingHeader

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	sovereignTests "github.com/multiversx/mx-chain-go/sovereignnode/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	sovTests "github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/multiversx/mx-sdk-abi-incubator/golang/abi"
	"github.com/stretchr/testify/require"
)

func createArgs() ArgsIncomingHeaderProcessor {
	return ArgsIncomingHeaderProcessor{
		HeadersPool:            &mock.HeadersCacherStub{},
		TxPool:                 &testscommon.ShardedDataStub{},
		Marshaller:             &marshallerMock.MarshalizerMock{},
		Hasher:                 &hashingMocks.HasherMock{},
		OutGoingOperationsPool: &sovTests.OutGoingOperationsPoolMock{},
	}
}

func requireErrorIsInvalidNumTopics(t *testing.T, err error, idx int, numTopics int) {
	require.True(t, strings.Contains(err.Error(), errInvalidNumTopicsIncomingEvent.Error()))
	require.True(t, strings.Contains(err.Error(), fmt.Sprintf("%d", idx)))
	require.True(t, strings.Contains(err.Error(), fmt.Sprintf("%d", numTopics)))
}

func createIncomingHeadersWithIncrementalRound(numRounds uint64) []sovereign.IncomingHeaderHandler {
	ret := make([]sovereign.IncomingHeaderHandler, numRounds+1)

	for i := uint64(0); i <= numRounds; i++ {
		ret[i] = &sovereign.IncomingHeader{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Round: i,
				},
			},
			IncomingEvents: []*transaction.Event{
				{
					Topics:     [][]byte{[]byte("endpoint"), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), []byte("val1")},
					Data:       createEventData(),
					Identifier: []byte(topicIDDeposit),
				},
			},
		}
	}

	return ret
}

func createEventData() []byte {
	codec := abi.NewDefaultCodec()
	serializer := abi.NewSerializer(codec)

	transferData := abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "tx_nonce",
				Value: abi.U64Value{Value: 10},
			},
			{
				Name: "gas_limit",
				Value: abi.OptionValue{
					Value: abi.U64Value{Value: 255},
				},
			},
			{
				Name: "function",
				Value: abi.OptionValue{
					Value: abi.BytesValue{Value: []byte("func1")},
				},
			},
			{
				Name: "args",
				Value: abi.OptionValue{
					Value: abi.InputListValue{
						Items: []any{
							abi.BytesValue{Value: []byte("arg1")},
						},
					},
				},
			},
		},
	}

	encoded, _ := serializer.Serialize([]any{transferData})
	encodedBytes, _ := hex.DecodeString(encoded)
	return encodedBytes
}

func TestNewIncomingHeaderHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil headers pool, should return error", func(t *testing.T) {
		args := createArgs()
		args.HeadersPool = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, errNilHeadersPool, err)
		require.Nil(t, handler)
	})

	t.Run("nil tx pool, should return error", func(t *testing.T) {
		args := createArgs()
		args.TxPool = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, errNilTxPool, err)
		require.Nil(t, handler)
	})

	t.Run("nil marshaller, should return error", func(t *testing.T) {
		args := createArgs()
		args.Marshaller = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, core.ErrNilMarshalizer, err)
		require.Nil(t, handler)
	})

	t.Run("nil hasher, should return error", func(t *testing.T) {
		args := createArgs()
		args.Hasher = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, core.ErrNilHasher, err)
		require.Nil(t, handler)
	})

	t.Run("should work", func(t *testing.T) {
		args := createArgs()
		handler, err := NewIncomingHeaderProcessor(args)
		require.Nil(t, err)
		require.False(t, check.IfNil(handler))
	})
}

func TestIncomingHeaderHandler_AddHeaderErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("nil header, should return error", func(t *testing.T) {
		args := createArgs()
		handler, _ := NewIncomingHeaderProcessor(args)

		err := handler.AddHeader([]byte("hash"), nil)
		require.Equal(t, data.ErrNilHeader, err)

		incomingHeader := &sovereignTests.IncomingHeaderStub{
			GetHeaderHandlerCalled: func() data.HeaderHandler {
				return nil
			},
		}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		require.Equal(t, data.ErrNilHeader, err)
	})

	t.Run("should not add header before start round", func(t *testing.T) {
		startRound := uint64(11)

		args := createArgs()
		args.MainChainNotarizationStartRound = startRound
		wasHeaderAdded := false
		args.HeadersPool = &mock.HeadersCacherStub{
			AddHeaderInShardCalled: func(headerHash []byte, header data.HeaderHandler, shardID uint32) {
				wasHeaderAdded = true
				require.Equal(t, header.GetRound(), startRound)
			},
		}
		handler, _ := NewIncomingHeaderProcessor(args)
		headers := createIncomingHeadersWithIncrementalRound(startRound)

		for i := 0; i < len(headers)-1; i++ {
			err := handler.AddHeader([]byte("hash"), headers[i])
			require.Nil(t, err)
			require.False(t, wasHeaderAdded)
		}

		err := handler.AddHeader([]byte("hash"), headers[startRound])
		require.Nil(t, err)
		require.True(t, wasHeaderAdded)
	})

	t.Run("invalid header type, should return error", func(t *testing.T) {
		args := createArgs()
		handler, _ := NewIncomingHeaderProcessor(args)

		incomingHeader := &sovereignTests.IncomingHeaderStub{
			GetHeaderHandlerCalled: func() data.HeaderHandler {
				return &block.MetaBlock{}
			},
		}
		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.Equal(t, errInvalidHeaderType, err)
	})

	t.Run("cannot compute extended header hash, should return error", func(t *testing.T) {
		args := createArgs()

		errMarshaller := errors.New("cannot marshal")
		args.Marshaller = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, errMarshaller
			},
		}
		handler, _ := NewIncomingHeaderProcessor(args)

		err := handler.AddHeader([]byte("hash"), &sovereign.IncomingHeader{Header: &block.HeaderV2{}})
		require.Equal(t, errMarshaller, err)
	})

	t.Run("invalid num topics in deposit event, should return error", func(t *testing.T) {
		args := createArgs()

		numSCRsAdded := 0
		args.TxPool = &testscommon.ShardedDataStub{
			AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheID string) {
				numSCRsAdded++
			},
		}

		incomingHeader := &sovereign.IncomingHeader{
			Header: &block.HeaderV2{},
			IncomingEvents: []*transaction.Event{
				{
					Topics:     [][]byte{[]byte("addr")},
					Identifier: []byte(topicIDDeposit),
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)

		err := handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 1)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte("deposit"), []byte("addr"), []byte("tokenID1")}, Identifier: []byte(topicIDDeposit)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 3)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte("deposit"), []byte("addr"), []byte("tokenID1"), []byte("nonce1")}, Identifier: []byte(topicIDDeposit)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 4)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte("deposit"), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), []byte("val1"), []byte("tokenID2")}, Identifier: []byte(topicIDDeposit)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 6)

		incomingHeader.IncomingEvents = []*transaction.Event{
			{
				Identifier: []byte(topicIDDeposit),
				Topics:     [][]byte{[]byte("deposit"), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), []byte("val1")},
				Data:       createEventData(),
			},
			{
				Identifier: []byte(topicIDDeposit),
				Topics:     [][]byte{[]byte("addr")},
				Data:       createEventData(),
			},
		}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 1, 1)

		require.Equal(t, 0, numSCRsAdded)
	})

	t.Run("invalid num topics in confirm bridge op event, should return error", func(t *testing.T) {
		args := createArgs()

		numConfirmedOperations := 0
		args.OutGoingOperationsPool = &sovTests.OutGoingOperationsPoolMock{
			ConfirmOperationCalled: func(hashOfHashes []byte, hash []byte) error {
				numConfirmedOperations++
				return nil
			},
		}

		incomingHeader := &sovereign.IncomingHeader{
			Header: &block.HeaderV2{},
			IncomingEvents: []*transaction.Event{
				{
					Topics:     [][]byte{},
					Identifier: []byte(topicIDExecutedBridgeOp),
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)

		err := handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 0)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte("hash")}, Identifier: []byte(topicIDDeposit)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 1)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte("hash"), []byte("hash1"), []byte("hash2")}, Identifier: []byte(topicIDDeposit)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 3)

		require.Equal(t, 0, numConfirmedOperations)
	})

	t.Run("cannot compute scr hash, should return error", func(t *testing.T) {
		args := createArgs()

		numSCRsAdded := 0
		args.TxPool = &testscommon.ShardedDataStub{
			AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheID string) {
				numSCRsAdded++
			},
		}

		errMarshaller := errors.New("cannot marshal")
		args.Marshaller = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				_, isSCR := obj.(*smartContractResult.SmartContractResult)
				if isSCR {
					return nil, errMarshaller
				}

				return json.Marshal(obj)
			},
		}

		incomingHeader := &sovereign.IncomingHeader{
			Header: &block.HeaderV2{},
			IncomingEvents: []*transaction.Event{
				{
					Identifier: []byte(topicIDDeposit),
					Topics:     [][]byte{[]byte("deposit"), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), []byte("val1")},
					Data:       createEventData(),
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)
		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorIs(t, err, errMarshaller)
		require.Equal(t, 0, numSCRsAdded)
	})
}

func TestIncomingHeaderProcessor_getEventData(t *testing.T) {
	t.Parallel()

	input := []byte("")
	ret, err := getEventData(input)
	require.Nil(t, ret)
	require.Equal(t, errEmptyLogData, err)

	input = []byte("0a")
	ret, err = getEventData(input)
	require.Nil(t, ret)
	require.True(t, strings.Contains(err.Error(), "cannot decode"))

	input = []byte("0a@ffaa@bb")
	ret, err = getEventData(input)
	require.Nil(t, ret)
	require.True(t, strings.Contains(err.Error(), "cannot decode"))

	input, err = hex.DecodeString("000000000000000a000000")
	ret, err = getEventData(input)
	require.Nil(t, err)
	require.Equal(t, &eventData{
		nonce:                uint64(10),
		functionCallWithArgs: []byte(""),
		gasLimit:             uint64(0),
	}, ret)

	input, err = hex.DecodeString("000000000000000a01000000000000005e010000000566756e633101000000010000000461726731")
	ret, err = getEventData(input)
	require.Nil(t, err)
	require.Equal(t, &eventData{
		nonce:                uint64(10),
		functionCallWithArgs: []byte("@66756e6331@61726731"),
		gasLimit:             uint64(94),
	}, ret)

	input, err = hex.DecodeString("0000000000000002010000000001312d000100000003616464010000000200000004002d310100000004002d3101")
	ret, err = getEventData(input)
	require.Nil(t, err)
	require.Equal(t, &eventData{
		nonce:                uint64(2),
		functionCallWithArgs: []byte("@616464@002d3101@002d3101"),
		gasLimit:             uint64(20000000),
	}, ret)
}

func TestIncomingHeaderHandler_AddHeader(t *testing.T) {
	t.Parallel()

	args := createArgs()

	codec := abi.NewDefaultCodec()
	serializer := abi.NewSerializer(codec)

	endpoint := []byte("endpoint")

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")

	gasLimit1 := uint64(45100)
	gasLimit2 := uint64(54100)

	scr1 := &smartContractResult.SmartContractResult{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  addr1,
		SndAddr:  core.ESDTSCAddress,
		Data:     []byte("MultiESDTNFTTransfer@02@746f6b656e31@04@64@746f6b656e32@@32@66756e6331@61726731@61726732"),
		GasLimit: gasLimit1,
	}
	scr2 := &smartContractResult.SmartContractResult{
		Nonce:    1,
		Value:    big.NewInt(0),
		RcvAddr:  addr2,
		SndAddr:  core.ESDTSCAddress,
		Data:     []byte("MultiESDTNFTTransfer@01@746f6b656e31@01@96@66756e6332@61726731"),
		GasLimit: gasLimit2,
	}

	scrHash1, err := core.CalculateHash(args.Marshaller, args.Hasher, scr1)
	require.Nil(t, err)
	scrHash2, err := core.CalculateHash(args.Marshaller, args.Hasher, scr2)
	require.Nil(t, err)

	cacheID := process.ShardCacherIdentifier(core.MainChainShardId, core.SovereignChainShardId)

	type scrInPool struct {
		data        *smartContractResult.SmartContractResult
		sizeInBytes int
		cacheID     string
	}
	expectedSCRsInPool := map[string]*scrInPool{
		string(scrHash1): {
			data:        scr1,
			sizeInBytes: scr1.Size(),
			cacheID:     cacheID,
		},
		string(scrHash2): {
			data:        scr2,
			sizeInBytes: scr2.Size(),
			cacheID:     cacheID,
		},
	}

	headerV2 := &block.HeaderV2{ScheduledRootHash: []byte("root hash")}

	transfer1 := [][]byte{
		[]byte("token1"),
		big.NewInt(4).Bytes(),
		big.NewInt(100).Bytes(),
	}
	transfer2 := [][]byte{
		[]byte("token2"),
		big.NewInt(0).Bytes(),
		big.NewInt(50).Bytes(),
	}
	topic1 := append([][]byte{endpoint}, [][]byte{addr1}...)
	topic1 = append(topic1, transfer1...)
	topic1 = append(topic1, transfer2...)

	transfer3 := [][]byte{
		[]byte("token1"),
		big.NewInt(1).Bytes(),
		big.NewInt(150).Bytes(),
	}
	topic2 := append([][]byte{endpoint}, [][]byte{addr2}...)
	topic2 = append(topic2, transfer3...)

	eventDataStruct1 := abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "tx_nonce",
				Value: abi.U64Value{Value: 0},
			},
			{
				Name: "gas_limit",
				Value: abi.OptionValue{
					Value: abi.U64Value{Value: gasLimit1},
				},
			},
			{
				Name: "function",
				Value: abi.OptionValue{
					Value: abi.BytesValue{Value: []byte("func1")},
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
	encodedEventData1, err := serializer.Serialize([]any{eventDataStruct1})
	require.Nil(t, err)
	eventData1, err := hex.DecodeString(encodedEventData1)
	require.Nil(t, err)

	eventDataStruct2 := abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "tx_nonce",
				Value: abi.U64Value{Value: 1},
			},
			{
				Name: "gas_limit",
				Value: abi.OptionValue{
					Value: abi.U64Value{Value: gasLimit2},
				},
			},
			{
				Name: "function",
				Value: abi.OptionValue{
					Value: abi.BytesValue{Value: []byte("func2")},
				},
			},
			{
				Name: "args",
				Value: abi.OptionValue{
					Value: abi.InputListValue{
						Items: []any{
							abi.BytesValue{Value: []byte("arg1")},
						},
					},
				},
			},
		},
	}
	encodedEventData2, err := serializer.Serialize([]any{eventDataStruct2})
	require.Nil(t, err)
	eventData2, err := hex.DecodeString(encodedEventData2)
	require.Nil(t, err)

	topic3 := [][]byte{
		[]byte("hashOfHashes"),
		[]byte("hashOfBridgeOp"),
	}

	incomingEvents := []*transaction.Event{
		{
			Identifier: []byte(topicIDDeposit),
			Topics:     topic1,
			Data:       eventData1,
		},
		{
			Identifier: []byte(topicIDDeposit),
			Topics:     topic2,
			Data:       eventData2,
		},
		{
			Identifier: []byte(topicIDExecutedBridgeOp),
			Topics:     topic3,
		},
	}

	extendedHeader := &block.ShardHeaderExtended{
		Header: headerV2,
		IncomingMiniBlocks: []*block.MiniBlock{
			{
				TxHashes:        [][]byte{scrHash1, scrHash2},
				ReceiverShardID: core.SovereignChainShardId,
				SenderShardID:   core.MainChainShardId,
				Type:            block.SmartContractResultBlock,
			},
		},
		IncomingEvents: incomingEvents,
	}
	extendedHeaderHash, err := core.CalculateHash(args.Marshaller, args.Hasher, extendedHeader)
	require.Nil(t, err)

	wasAddedInHeaderPool := false
	args.HeadersPool = &mock.HeadersCacherStub{
		AddHeaderInShardCalled: func(headerHash []byte, header data.HeaderHandler, shardID uint32) {
			require.Equal(t, extendedHeaderHash, headerHash)
			require.Equal(t, extendedHeader, header)
			require.Equal(t, core.MainChainShardId, shardID)

			wasAddedInHeaderPool = true
		},
	}

	wasAddedInTxPool := false
	args.TxPool = &testscommon.ShardedDataStub{
		AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheID string) {
			expectedSCR, found := expectedSCRsInPool[string(key)]
			require.True(t, found)

			require.Equal(t, expectedSCR.data, data)
			require.Equal(t, expectedSCR.sizeInBytes, sizeInBytes)
			require.Equal(t, expectedSCR.cacheID, cacheID)

			wasAddedInTxPool = true
		},
	}

	wasOutGoingOpConfirmed := false
	args.OutGoingOperationsPool = &sovTests.OutGoingOperationsPoolMock{
		ConfirmOperationCalled: func(hashOfHashes []byte, hash []byte) error {
			require.Equal(t, topic3[0], hashOfHashes)
			require.Equal(t, topic3[1], hash)

			wasOutGoingOpConfirmed = true
			return nil
		},
	}

	handler, _ := NewIncomingHeaderProcessor(args)
	incomingHeader := &sovereign.IncomingHeader{
		Header:         headerV2,
		IncomingEvents: incomingEvents,
	}
	err = handler.AddHeader([]byte("hash"), incomingHeader)
	require.Nil(t, err)
	require.True(t, wasAddedInHeaderPool)
	require.True(t, wasAddedInTxPool)
	require.True(t, wasOutGoingOpConfirmed)
}
