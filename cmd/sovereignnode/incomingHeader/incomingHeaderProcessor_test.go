package incomingHeader

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/process/mock"
	sovereignTests "github.com/multiversx/mx-chain-go/sovereignnode/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	sovTests "github.com/multiversx/mx-chain-go/testscommon/sovereign"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func createArgs() ArgsIncomingHeaderProcessor {
	return ArgsIncomingHeaderProcessor{
		HeadersPool:            &mock.HeadersCacherStub{},
		TxPool:                 &testscommon.ShardedDataStub{},
		Marshaller:             &marshallerMock.MarshalizerMock{},
		Hasher:                 &hashingMocks.HasherMock{},
		OutGoingOperationsPool: &sovTests.OutGoingOperationsPoolMock{},
		DataCodec:              &mock.DataCodecMock{},
		TopicsChecker:          &mock.TopicsCheckerMock{},
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
					Topics:     [][]byte{[]byte("topicID"), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), []byte("0000000001010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")},
					Data:       []byte("0000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
					Identifier: []byte(eventIDDepositIncomingTransfer),
				},
			},
		}
	}

	return ret
}

func TestNewIncomingHeaderHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil headers pool, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.HeadersPool = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, errNilHeadersPool, err)
		require.Nil(t, handler)
	})

	t.Run("nil tx pool, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.TxPool = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, errNilTxPool, err)
		require.Nil(t, handler)
	})

	t.Run("nil marshaller, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.Marshaller = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, core.ErrNilMarshalizer, err)
		require.Nil(t, handler)
	})

	t.Run("nil hasher, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.Hasher = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, core.ErrNilHasher, err)
		require.Nil(t, handler)
	})

	t.Run("nil outgoing operations pool, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.OutGoingOperationsPool = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, errorsMx.ErrNilOutGoingOperationsPool, err)
		require.Nil(t, handler)
	})

	t.Run("nil data codec, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.DataCodec = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, errorsMx.ErrNilDataCodec, err)
		require.Nil(t, handler)
	})

	t.Run("nil topics checker, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.TopicsChecker = nil

		handler, err := NewIncomingHeaderProcessor(args)
		require.Equal(t, errorsMx.ErrNilTopicsChecker, err)
		require.Nil(t, handler)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		handler, err := NewIncomingHeaderProcessor(args)
		require.Nil(t, err)
		require.False(t, check.IfNil(handler))
	})
}

func TestIncomingHeaderHandler_AddHeaderErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("nil header, should return error", func(t *testing.T) {
		t.Parallel()

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
		t.Parallel()

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
		args.DataCodec = &mock.DataCodecMock{
			DeserializeTokenDataCalled: func(_ []byte) (*sovereign.EsdtTokenData, error) {
				addr, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
				return &sovereign.EsdtTokenData{
					TokenType:  0,
					Amount:     big.NewInt(0),
					Frozen:     false,
					Hash:       make([]byte, 0),
					Name:       []byte("token"),
					Attributes: make([]byte, 0),
					Creator:    addr,
					Royalties:  big.NewInt(0),
					Uris:       make([][]byte, 0),
				}, nil
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
		t.Parallel()

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
		t.Parallel()

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
		t.Parallel()

		args := createArgs()

		args.TopicsChecker = &mock.TopicsCheckerMock{
			CheckValidityCalled: func(topics [][]byte) error {
				return fmt.Errorf("invalid num topics")
			},
		}

		incomingHeader := &sovereign.IncomingHeader{
			Header: &block.HeaderV2{},
			IncomingEvents: []*transaction.Event{
				{
					Identifier: []byte(eventIDDepositIncomingTransfer),
					Topics:     [][]byte{},
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)
		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorContains(t, err, "invalid num topics")
	})

	t.Run("invalid num topics in executed ops event, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()

		args.TopicsChecker = &mock.TopicsCheckerMock{
			CheckValidityCalled: func(topics [][]byte) error {
				return fmt.Errorf("invalid num topics")
			},
		}

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
					Identifier: []byte(eventIDExecutedOutGoingBridgeOp),
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)

		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorContains(t, err, errInvalidNumTopicsIncomingEvent.Error())

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte(topicIDDepositIncomingTransfer)}, Identifier: []byte(eventIDExecutedOutGoingBridgeOp)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorContains(t, err, "invalid num topics")

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte(topicIDConfirmedOutGoingOperation)}, Identifier: []byte(eventIDExecutedOutGoingBridgeOp)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 1)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte(topicIDConfirmedOutGoingOperation), []byte("hash")}, Identifier: []byte(eventIDExecutedOutGoingBridgeOp)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 2)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte(topicIDConfirmedOutGoingOperation), []byte("hash"), []byte("hash1"), []byte("hash2")}, Identifier: []byte(eventIDExecutedOutGoingBridgeOp)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 4)

		require.Equal(t, 0, numConfirmedOperations)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte("topicID")}, Identifier: []byte(eventIDExecutedOutGoingBridgeOp)}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		require.Equal(t, errInvalidIncomingTopicIdentifier, err)
	})

	t.Run("invalid event id, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()

		incomingHeader := &sovereign.IncomingHeader{
			Header: &block.HeaderV2{},
			IncomingEvents: []*transaction.Event{
				{
					Identifier: []byte("eventID"),
					Topics:     [][]byte{},
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)
		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.Equal(t, errInvalidIncomingEventIdentifier, err)
	})

	t.Run("cannot compute scr hash, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()

		args.DataCodec = &mock.DataCodecMock{
			DeserializeTokenDataCalled: func(_ []byte) (*sovereign.EsdtTokenData, error) {
				addr, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
				return &sovereign.EsdtTokenData{
					TokenType:  0,
					Amount:     big.NewInt(0),
					Frozen:     false,
					Hash:       make([]byte, 0),
					Name:       []byte("token"),
					Attributes: make([]byte, 0),
					Creator:    addr,
					Royalties:  big.NewInt(0),
					Uris:       make([][]byte, 0),
				}, nil
			},
		}

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
					Identifier: []byte(eventIDDepositIncomingTransfer),
					Topics:     [][]byte{[]byte(topicIDDepositIncomingTransfer), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), []byte("0000000001010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")},
					Data:       []byte("0000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)
		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorIs(t, err, errMarshaller)
		require.Equal(t, 0, numSCRsAdded)
	})

	t.Run("cannot create event data, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()

		args.DataCodec = &mock.DataCodecMock{
			DeserializeEventDataCalled: func(_ []byte) (*sovereign.EventData, error) {
				return nil, fmt.Errorf("cannot deserialize event data")
			},
		}

		incomingHeader := &sovereign.IncomingHeader{
			Header: &block.HeaderV2{},
			IncomingEvents: []*transaction.Event{
				{
					Identifier: []byte(eventIDDepositIncomingTransfer),
					Topics:     [][]byte{[]byte(topicIDDepositIncomingTransfer), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), []byte("tokenData1")},
					Data:       []byte("data"),
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)
		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorContains(t, err, "cannot deserialize event data")
	})

	t.Run("cannot create token data, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()

		args.DataCodec = &mock.DataCodecMock{
			DeserializeTokenDataCalled: func(_ []byte) (*sovereign.EsdtTokenData, error) {
				return nil, fmt.Errorf("cannot deserialize token data")
			},
			DeserializeEventDataCalled: func(_ []byte) (*sovereign.EventData, error) {
				return &sovereign.EventData{}, nil
			},
		}

		incomingHeader := &sovereign.IncomingHeader{
			Header: &block.HeaderV2{},
			IncomingEvents: []*transaction.Event{
				{
					Identifier: []byte(eventIDDepositIncomingTransfer),
					Topics:     [][]byte{[]byte(topicIDDepositIncomingTransfer), []byte("addr"), []byte("tokenID1"), []byte("nonce1"), []byte("tokenData1")},
					Data:       []byte("data"),
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)
		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.ErrorContains(t, err, "cannot deserialize token data")
	})
}

func TestIncomingHeaderProcessor_createEventData(t *testing.T) {
	t.Parallel()

	t.Run("empty transfer data", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.DataCodec = &mock.DataCodecMock{
			DeserializeEventDataCalled: func(_ []byte) (*sovereign.EventData, error) {
				return &sovereign.EventData{
					TransferData: nil,
				}, nil
			},
		}
		handler, _ := NewIncomingHeaderProcessor(args)

		ret, err := handler.eventsProc.createEventData([]byte(""))
		require.Nil(t, err)
		require.Equal(t, &eventData{
			functionCallWithArgs: make([]byte, 0),
		}, ret)
	})

	t.Run("transfer data with function no args", func(t *testing.T) {
		t.Parallel()

		func1 := []byte("func1")

		args := createArgs()
		args.DataCodec = &mock.DataCodecMock{
			DeserializeEventDataCalled: func(_ []byte) (*sovereign.EventData, error) {
				return &sovereign.EventData{
					TransferData: &sovereign.TransferData{
						Function: func1,
					},
				}, nil
			},
		}
		handler, _ := NewIncomingHeaderProcessor(args)

		ret, err := handler.eventsProc.createEventData([]byte(""))
		require.Nil(t, err)
		expectedArgs := append([]byte("@"), hex.EncodeToString(func1)...)
		require.Equal(t, &eventData{
			functionCallWithArgs: expectedArgs,
		}, ret)
	})

	t.Run("transfer data with function and args", func(t *testing.T) {
		t.Parallel()

		func1 := []byte("func1")
		arg1 := []byte("arg1")
		arg2 := []byte("arg2")

		args := createArgs()
		args.DataCodec = &mock.DataCodecMock{
			DeserializeEventDataCalled: func(_ []byte) (*sovereign.EventData, error) {
				return &sovereign.EventData{
					TransferData: &sovereign.TransferData{
						Function: func1,
						Args:     [][]byte{arg1, arg2},
					},
				}, nil
			},
		}
		handler, _ := NewIncomingHeaderProcessor(args)

		ret, err := handler.eventsProc.createEventData([]byte(""))
		require.Nil(t, err)
		expectedArgs := append([]byte("@"), hex.EncodeToString(func1)...)
		expectedArgs = append(expectedArgs, "@"+hex.EncodeToString(arg1)...)
		expectedArgs = append(expectedArgs, "@"+hex.EncodeToString(arg2)...)
		require.Equal(t, &eventData{
			functionCallWithArgs: expectedArgs,
		}, ret)
	})
}

func TestIncomingHeaderProcessor_createSCRData(t *testing.T) {
	t.Parallel()

	topicID := []byte("topicID")
	receiver := []byte("rcv")
	nft := []byte("nft")
	nonce := []byte("nonce")
	nftData := []byte("nftData")

	topics := [][]byte{
		topicID,
		receiver,
		nft,
		nonce,
		nftData,
	}

	args := createArgs()
	args.DataCodec = &mock.DataCodecMock{
		DeserializeTokenDataCalled: func(_ []byte) (*sovereign.EsdtTokenData, error) {
			return &sovereign.EsdtTokenData{
				TokenType: core.NonFungible,
				Royalties: big.NewInt(0),
			}, nil
		},
	}
	args.Marshaller = &marshallerMock.MarshalizerStub{
		MarshalCalled: func(_ interface{}) ([]byte, error) {
			return nftData, nil
		},
	}
	handler, _ := NewIncomingHeaderProcessor(args)

	ret, err := handler.eventsProc.createSCRData(topics)
	require.Nil(t, err)

	expectedSCR := []byte(core.BuiltInFunctionMultiESDTNFTTransfer + "@01")
	expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(nft)...)
	expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(nonce)...)
	expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(nftData)...)
	require.Equal(t, expectedSCR, ret)
}

func TestIncomingHeaderHandler_AddHeader(t *testing.T) {
	t.Parallel()

	args := createArgs()

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")

	token1 := []byte("token1")
	token2 := []byte("token2")

	token1Nonce := make([]byte, 0)
	token2Nonce := []byte{0x01}

	amount1 := big.NewInt(100)
	token1Data := amount1.Bytes()
	amount2 := big.NewInt(50)
	token2Data := amount2.Bytes()

	scr1Nonce := uint64(0)
	scr2Nonce := uint64(1)
	gasLimit1 := uint64(45100)
	gasLimit2 := uint64(54100)
	func1 := []byte("func1")
	func2 := []byte("func2")
	arg1 := []byte("arg1")
	arg2 := []byte("arg2")

	scr1 := &smartContractResult.SmartContractResult{
		Nonce:   scr1Nonce,
		Value:   big.NewInt(0),
		RcvAddr: addr1,
		SndAddr: core.ESDTSCAddress,
		Data: []byte(core.BuiltInFunctionMultiESDTNFTTransfer + "@02" +
			"@" + hex.EncodeToString(token1) +
			"@" + hex.EncodeToString(token1Nonce) +
			"@" + hex.EncodeToString(token1Data) +
			"@" + hex.EncodeToString(token2) +
			"@" + hex.EncodeToString(token2Nonce) +
			"@" + hex.EncodeToString(token2Data) +
			"@" + hex.EncodeToString(func1) +
			"@" + hex.EncodeToString(arg1) +
			"@" + hex.EncodeToString(arg2)),
		GasLimit: gasLimit1,
	}
	scr2 := &smartContractResult.SmartContractResult{
		Nonce:   scr2Nonce,
		Value:   big.NewInt(0),
		RcvAddr: addr2,
		SndAddr: core.ESDTSCAddress,
		Data: []byte(core.BuiltInFunctionMultiESDTNFTTransfer + "@01" +
			"@" + hex.EncodeToString(token1) +
			"@" + hex.EncodeToString(token1Nonce) +
			"@" + hex.EncodeToString(token2Data) +
			"@" + hex.EncodeToString(func2) +
			"@" + hex.EncodeToString(arg1)),
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
		token2Data,
	}
	topic1 := [][]byte{
		[]byte(topicIDDepositIncomingTransfer),
		addr1,
	}
	topic1 = append(topic1, transfer1...)
	topic1 = append(topic1, transfer2...)
	eventData1 := []byte("eventData1")

	transfer3 := [][]byte{
		token1,
		token1Nonce,
		token2Data,
	}
	topic2 := [][]byte{
		[]byte(topicIDDepositIncomingTransfer),
		addr2,
	}
	topic2 = append(topic2, transfer3...)
	eventData2 := []byte("eventData2")

	topic3 := [][]byte{
		[]byte(topicIDConfirmedOutGoingOperation),
		[]byte("hashOfHashes"),
		[]byte("hashOfBridgeOp"),
	}

	incomingEvents := []*transaction.Event{
		{
			Identifier: []byte(eventIDDepositIncomingTransfer),
			Topics:     topic1,
			Data:       eventData1,
		},
		{
			Identifier: []byte(eventIDDepositIncomingTransfer),
			Topics:     topic2,
			Data:       eventData2,
		},
		{
			Identifier: []byte(eventIDExecutedOutGoingBridgeOp),
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
			require.Equal(t, topic3[0], []byte(topicIDConfirmedOutGoingOperation))
			require.Equal(t, topic3[1], hashOfHashes)
			require.Equal(t, topic3[2], hash)

			wasOutGoingOpConfirmed = true
			return nil
		},
	}

	tcCalled := -1
	args.TopicsChecker = &mock.TopicsCheckerMock{
		CheckValidityCalled: func(topics [][]byte) error {
			tcCalled++

			if tcCalled == 0 {
				require.Equal(t, topic1, topics)
			} else {
				require.Equal(t, topic2, topics)
			}

			return nil
		},
	}

	edCalled := -1
	tdCalled := -1
	args.DataCodec = &mock.DataCodecMock{
		DeserializeEventDataCalled: func(data []byte) (*sovereign.EventData, error) {
			edCalled++

			if edCalled == 0 {
				require.Equal(t, eventData1, data)

				return &sovereign.EventData{
					Nonce: scr1Nonce,
					TransferData: &sovereign.TransferData{
						Function: func1,
						Args:     [][]byte{arg1, arg2},
						GasLimit: gasLimit1,
					},
				}, nil
			} else {
				require.Equal(t, eventData2, data)

				return &sovereign.EventData{
					Nonce: scr2Nonce,
					TransferData: &sovereign.TransferData{
						Function: func2,
						Args:     [][]byte{arg1},
						GasLimit: gasLimit2,
					},
				}, nil
			}
		},

		DeserializeTokenDataCalled: func(data []byte) (*sovereign.EsdtTokenData, error) {
			tdCalled++

			if tdCalled == 0 {
				require.Equal(t, token1Data, data)

				return &sovereign.EsdtTokenData{
					TokenType: core.Fungible,
					Amount:    amount1,
				}, nil
			} else {
				require.Equal(t, token2Data, data)

				return &sovereign.EsdtTokenData{
					TokenType: core.Fungible,
					Amount:    amount2,
				}, nil
			}

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
