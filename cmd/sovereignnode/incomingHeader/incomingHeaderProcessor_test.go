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
	"github.com/multiversx/mx-chain-go/sovereignnode/dataCodec"
	sovereignTests "github.com/multiversx/mx-chain-go/sovereignnode/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	sovTests "github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/stretchr/testify/require"
)

func createArgs() ArgsIncomingHeaderProcessor {
	dataCodecMock, _ := dataCodec.NewDataCodec(&marshallerMock.MarshalizerMock{})

	return ArgsIncomingHeaderProcessor{
		HeadersPool:            &mock.HeadersCacherStub{},
		TxPool:                 &testscommon.ShardedDataStub{},
		Marshaller:             &marshallerMock.MarshalizerMock{},
		Hasher:                 &hashingMocks.HasherMock{},
		OutGoingOperationsPool: &sovTests.OutGoingOperationsPoolMock{},
		DataCodec:              dataCodecMock,
	}
}

func requireErrorIsInvalidNumTopics(t *testing.T, err error, idx int, numTopics int) {
	require.True(t, strings.Contains(err.Error(), errInvalidNumTopicsIncomingEvent.Error()))
	require.True(t, strings.Contains(err.Error(), fmt.Sprintf("%d", idx)))
	require.True(t, strings.Contains(err.Error(), fmt.Sprintf("%d", numTopics)))
}

func createIncomingHeadersWithIncrementalRound(numRounds uint64) []sovereign.IncomingHeaderHandler {
	ret := make([]sovereign.IncomingHeaderHandler, numRounds+1)

	tokenData := createTokenData(big.NewInt(100))

	for i := uint64(0); i <= numRounds; i++ {
		ret[i] = &sovereign.IncomingHeader{
			Header: &block.HeaderV2{
				Header: &block.Header{
					Round: i,
				},
			},
			IncomingEvents: []*transaction.Event{
				{
					Topics:     [][]byte{[]byte("endpoint"), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), tokenData},
					Data:       createEventData(),
					Identifier: []byte(topicIDDeposit),
				},
			},
		}
	}

	return ret
}

func createTokenData(amount *big.Int) []byte {
	args := createArgs()

	addr0, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	tokenData, _ := args.DataCodec.SerializeTokenData(sovereign.EsdtTokenData{
		TokenType:  0,
		Amount:     amount,
		Frozen:     false,
		Hash:       make([]byte, 0),
		Name:       []byte("token"),
		Attributes: make([]byte, 0),
		Creator:    addr0,
		Royalties:  big.NewInt(0),
		Uris:       make([][]byte, 0),
	})

	return tokenData
}

func createNftTokenData() []byte {
	args := createArgs()
	handler, _ := NewIncomingHeaderProcessor(args)

	addr0, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	tokenData, _ := handler.eventsProc.dataCodec.SerializeTokenData(sovereign.EsdtTokenData{
		TokenType:  1,
		Amount:     big.NewInt(1),
		Frozen:     false,
		Hash:       make([]byte, 0),
		Name:       []byte("nft"),
		Attributes: make([]byte, 0),
		Creator:    addr0,
		Royalties:  big.NewInt(250),
		Uris:       make([][]byte, 0),
	})

	return tokenData
}

func createEventData() []byte {
	args := createArgs()

	tdArgs := make([][]byte, 0)
	tdArgs = append(tdArgs, []byte("arg1"))
	evData, _ := args.DataCodec.SerializeEventData(sovereign.EventData{
		Nonce: 10,
		TransferData: &sovereign.TransferData{
			GasLimit: 255,
			Function: []byte("func1"),
			Args:     tdArgs,
		},
	})

	return evData
}

func createCustomEventData(nonce uint64, gasLimit uint64, function []byte, arguments [][]byte) []byte {
	args := createArgs()

	evData, _ := args.DataCodec.SerializeEventData(sovereign.EventData{
		Nonce: nonce,
		TransferData: &sovereign.TransferData{
			GasLimit: gasLimit,
			Function: function,
			Args:     arguments,
		},
	})

	return evData
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
		tokenData := createTokenData(big.NewInt(100))

		err := handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 1)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte("endpoint"), []byte("addr"), []byte("tokenID1")}, Identifier: []byte(topicIDDeposit)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 3)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte("endpoint"), []byte("addr"), []byte("tokenID1"), []byte("nonce1")}, Identifier: []byte(topicIDDeposit)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 4)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte("endpoint"), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), tokenData, []byte("tokenID2")}, Identifier: []byte(topicIDDeposit)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 6)

		incomingHeader.IncomingEvents = []*transaction.Event{
			{
				Identifier: []byte(topicIDDeposit),
				Topics:     [][]byte{[]byte("endpoint"), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), tokenData},
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
					Identifier: []byte(topicIDExecuteBridgeOps),
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)

		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorIs(t, err, errInvalidNumTopicsIncomingEvent)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte(topicIDExecutedBridgeOp)}, Identifier: []byte(topicIDExecuteBridgeOps)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 1)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte(topicIDExecutedBridgeOp), []byte("hash")}, Identifier: []byte(topicIDExecuteBridgeOps)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 2)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte(topicIDExecutedBridgeOp), []byte("hash"), []byte("hash1"), []byte("hash2")}, Identifier: []byte(topicIDExecuteBridgeOps)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 4)

		require.Equal(t, 0, numConfirmedOperations)
	})

	t.Run("cannot compute scr hash, should return error", func(t *testing.T) {
		args := createArgs()
		tokenData := createTokenData(big.NewInt(100))

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
					Topics:     [][]byte{[]byte("endpoint"), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), tokenData},
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

	args := createArgs()
	handler, _ := NewIncomingHeaderProcessor(args)

	input := []byte("")
	ret, err := handler.eventsProc.getEventData(input)
	require.Nil(t, ret)
	require.Equal(t, errEmptyLogData, err)

	input = []byte("0a")
	ret, err = handler.eventsProc.getEventData(input)
	require.Nil(t, ret)
	require.True(t, strings.Contains(err.Error(), "cannot decode"))

	input = []byte("0a@ffaa@bb")
	ret, err = handler.eventsProc.getEventData(input)
	require.Nil(t, ret)
	require.True(t, strings.Contains(err.Error(), "cannot decode"))

	input, err = hex.DecodeString("000000000000000a000000")
	hexEventData, err := handler.eventsProc.dataCodec.SerializeEventData(sovereign.EventData{
		Nonce:        10,
		TransferData: nil,
	})
	require.Equal(t, input, hexEventData)
	ret, err = handler.eventsProc.getEventData(hexEventData)
	require.Nil(t, err)
	require.Equal(t, &eventData{
		nonce:                uint64(10),
		functionCallWithArgs: []byte(""),
		gasLimit:             uint64(0),
	}, ret)

	input, err = hex.DecodeString("000000000000000a01000000000000005e010000000566756e6331010000000200000004617267310000000461726732")
	hexEventData, err = handler.eventsProc.dataCodec.SerializeEventData(sovereign.EventData{
		Nonce: 10,
		TransferData: &sovereign.TransferData{
			GasLimit: 94,
			Function: []byte("func1"),
			Args:     [][]byte{[]byte("arg1"), []byte("arg2")},
		},
	})
	require.Equal(t, input, hexEventData)
	ret, err = handler.eventsProc.getEventData(input)
	require.Nil(t, err)
	require.Equal(t, &eventData{
		nonce:                uint64(10),
		functionCallWithArgs: []byte("@66756e6331@61726731@61726732"),
		gasLimit:             uint64(94),
	}, ret)

	input, err = hex.DecodeString("0000000000000002010000000001312d00010000000361646401000000010000000401312d00")
	hexEventData, err = handler.eventsProc.dataCodec.SerializeEventData(sovereign.EventData{
		Nonce: 2,
		TransferData: &sovereign.TransferData{
			GasLimit: 20000000,
			Function: []byte("add"),
			Args:     [][]byte{big.NewInt(20000000).Bytes()},
		},
	})
	require.Equal(t, input, hexEventData)
	ret, err = handler.eventsProc.getEventData(input)
	require.Nil(t, err)
	require.Equal(t, &eventData{
		nonce:                uint64(2),
		functionCallWithArgs: []byte("@616464@01312d00"),
		gasLimit:             uint64(20000000),
	}, ret)
}

func TestIncomingHeaderHandler_AddHeader(t *testing.T) {
	t.Parallel()

	args := createArgs()

	endpoint := []byte("endpoint")

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")

	token1 := []byte("token1")
	token2 := []byte("token2")

	token1Nonce := make([]byte, 0)
	token2Nonce := []byte{0x04}

	gasLimit1 := uint64(45100)
	gasLimit2 := uint64(54100)

	func1 := []byte("func1")
	func2 := []byte("func2")
	arg1 := []byte("arg1")
	arg2 := []byte("arg2")

	token1Data := createTokenData(big.NewInt(100))
	token1DataBytes, _ := args.DataCodec.GetTokenDataBytes(token1Nonce, token1Data)
	token2Data := createTokenData(big.NewInt(50))
	token2DataBytes, _ := args.DataCodec.GetTokenDataBytes(token1Nonce, token2Data)
	nftData := createNftTokenData()
	nftDataBytes, _ := args.DataCodec.GetTokenDataBytes(token2Nonce, nftData)

	scr1 := &smartContractResult.SmartContractResult{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  addr1,
		SndAddr:  core.ESDTSCAddress,
		Data:     []byte("MultiESDTNFTTransfer@02@" + hex.EncodeToString(token1) + "@" + hex.EncodeToString(token1Nonce) + "@" + hex.EncodeToString(token1DataBytes) + "@" + hex.EncodeToString(token2) + "@" + hex.EncodeToString(token2Nonce) + "@" + hex.EncodeToString(nftDataBytes) + "@" + hex.EncodeToString(func1) + "@" + hex.EncodeToString(arg1) + "@" + hex.EncodeToString(arg2)),
		GasLimit: gasLimit1,
	}
	scr2 := &smartContractResult.SmartContractResult{
		Nonce:    1,
		Value:    big.NewInt(0),
		RcvAddr:  addr2,
		SndAddr:  core.ESDTSCAddress,
		Data:     []byte("MultiESDTNFTTransfer@01@" + hex.EncodeToString(token1) + "@" + hex.EncodeToString(token1Nonce) + "@" + hex.EncodeToString(token2DataBytes) + "@" + hex.EncodeToString(func2) + "@" + hex.EncodeToString(arg1)),
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
		token1,
		token1Nonce,
		token1Data,
	}
	transfer2 := [][]byte{
		token2,
		token2Nonce,
		nftData,
	}
	topic1 := append([][]byte{endpoint}, [][]byte{addr1}...)
	topic1 = append(topic1, transfer1...)
	topic1 = append(topic1, transfer2...)
	eventData1 := createCustomEventData(uint64(0), gasLimit1, func1, [][]byte{arg1, arg2})

	transfer3 := [][]byte{
		token1,
		token1Nonce,
		token2Data,
	}
	topic2 := append([][]byte{endpoint}, [][]byte{addr2}...)
	topic2 = append(topic2, transfer3...)
	eventData2 := createCustomEventData(uint64(1), gasLimit2, func2, [][]byte{arg1})

	topic3 := [][]byte{
		[]byte("executedBridgeOp"),
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
			Identifier: []byte(topicIDExecuteBridgeOps),
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
