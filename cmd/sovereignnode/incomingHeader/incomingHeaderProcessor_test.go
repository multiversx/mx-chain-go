package incomingHeader

import (
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
	"github.com/stretchr/testify/require"
)

func createArgs() ArgsIncomingHeaderProcessor {
	return ArgsIncomingHeaderProcessor{
		HeadersPool: &mock.HeadersCacherStub{},
		TxPool:      &testscommon.ShardedDataStub{},
		Marshaller:  &marshallerMock.MarshalizerMock{},
		Hasher:      &hashingMocks.HasherMock{},
	}
}

func requireErrorIsInvalidNumTopics(t *testing.T, err error, idx int, numTopics int) {
	require.True(t, strings.Contains(err.Error(), errInvalidNumTopicsIncomingEvent.Error()))
	require.True(t, strings.Contains(err.Error(), fmt.Sprintf("%d", idx)))
	require.True(t, strings.Contains(err.Error(), fmt.Sprintf("%d", numTopics)))
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

	t.Run("invalid num topics in event, should return error", func(t *testing.T) {
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
					Topics: [][]byte{[]byte("addr")},
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)

		err := handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 1)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte("addr"), []byte("tokenID1")}}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 2)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte("addr"), []byte("tokenID1"), []byte("nonce1")}}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 3)

		incomingHeader.IncomingEvents[0] = &transaction.Event{Topics: [][]byte{[]byte("addr"), []byte("tokenID1"), []byte("nonce1"), []byte("val1"), []byte("tokenID2")}}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 0, 5)

		incomingHeader.IncomingEvents = []*transaction.Event{
			{
				Topics: [][]byte{[]byte("addr"), []byte("tokenID1"), []byte("nonce1"), []byte("val1")},
			},
			{
				Topics: [][]byte{[]byte("addr")},
			},
		}
		err = handler.AddHeader([]byte("hash"), incomingHeader)
		requireErrorIsInvalidNumTopics(t, err, 1, 1)

		require.Equal(t, 0, numSCRsAdded)
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
					Topics: [][]byte{[]byte("addr"), []byte("tokenID1"), []byte("nonce1"), []byte("val1")},
				},
			},
		}

		handler, _ := NewIncomingHeaderProcessor(args)
		err := handler.AddHeader([]byte("hash"), incomingHeader)
		require.Equal(t, errMarshaller, err)
		require.Equal(t, 0, numSCRsAdded)
	})
}

func TestIncomingHeaderHandler_AddHeader(t *testing.T) {
	t.Parallel()

	args := createArgs()

	addr1 := []byte("addr1")
	addr2 := []byte("addr2")

	scr1 := &smartContractResult.SmartContractResult{
		Nonce:   0,
		Value:   big.NewInt(0),
		RcvAddr: addr1,
		SndAddr: core.ESDTSCAddress,
		Data:    []byte("MultiESDTNFTTransfer@02@746f6b656e31@04@64@746f6b656e32@@32"),
	}
	scr2 := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(0),
		RcvAddr: addr2,
		SndAddr: core.ESDTSCAddress,
		Data:    []byte("MultiESDTNFTTransfer@01@746f6b656e31@01@96"),
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
	topic1 := append([][]byte{addr1}, transfer1...)
	topic1 = append(topic1, transfer2...)

	transfer3 := [][]byte{
		[]byte("token1"),
		big.NewInt(1).Bytes(),
		big.NewInt(150).Bytes(),
	}
	topic2 := append([][]byte{addr2}, transfer3...)

	incomingEvents := []*transaction.Event{
		{
			Identifier: []byte("deposit"),
			Topics:     topic1,
		},
		{
			Identifier: []byte("deposit"),
			Topics:     topic2,
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
		AddCalled: func(headerHash []byte, header data.HeaderHandler) {
			require.Equal(t, extendedHeaderHash, headerHash)
			require.Equal(t, extendedHeader, header)

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

	handler, _ := NewIncomingHeaderProcessor(args)
	incomingHeader := &sovereign.IncomingHeader{
		Header:         headerV2,
		IncomingEvents: incomingEvents,
	}
	err = handler.AddHeader([]byte("hash"), incomingHeader)
	require.Nil(t, err)
	require.True(t, wasAddedInHeaderPool)
	require.True(t, wasAddedInTxPool)
}

func TestIncomingHeaderHandler_AddHeaderExpectNoncesIncremented(t *testing.T) {
	t.Parallel()

	args := createArgs()

	addr1 := []byte("addr1")
	addInTxPoolCt := uint64(0)
	args.TxPool = &testscommon.ShardedDataStub{
		AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheID string) {
			expectedSCR := &smartContractResult.SmartContractResult{
				Nonce:   addInTxPoolCt,
				Value:   big.NewInt(0),
				RcvAddr: addr1,
				SndAddr: core.ESDTSCAddress,
				Data:    []byte("MultiESDTNFTTransfer@01@746f6b656e31@@64"),
			}

			expectedScrHash, err := core.CalculateHash(args.Marshaller, args.Hasher, expectedSCR)
			require.Nil(t, err)
			require.Equal(t, expectedScrHash, key)

			addInTxPoolCt++
		},
	}

	errCt := uint64(4)
	errUnmarshall := errors.New("error unmarshall")
	args.Marshaller = &marshallerMock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			if addInTxPoolCt == errCt {
				return nil, errUnmarshall
			}

			return json.Marshal(obj)
		},
	}

	handler, _ := NewIncomingHeaderProcessor(args)
	incomingHeader := &sovereign.IncomingHeader{
		Header: &block.HeaderV2{
			ScheduledRootHash: []byte("root hash"),
		},
		IncomingEvents: []*transaction.Event{
			{
				Identifier: []byte("deposit"),
				Topics: [][]byte{
					addr1,
					[]byte("token1"),
					big.NewInt(0).Bytes(),
					big.NewInt(100).Bytes(),
				},
			},
		},
	}

	i := uint64(0)
	for ; i < 4; i++ {
		err := handler.AddHeader([]byte(fmt.Sprintf("hash%d", i)), incomingHeader)
		require.Nil(t, err)
		require.Equal(t, i+1, addInTxPoolCt)
	}
	require.Equal(t, uint64(4), handler.scrProc.nonce)

	err := handler.AddHeader([]byte(fmt.Sprintf("hash%d", i)), incomingHeader)
	require.Equal(t, errUnmarshall, err)
	require.Equal(t, i, addInTxPoolCt)
	require.Equal(t, uint64(4), handler.scrProc.nonce)

	errCt = 0
	for ; i < 10; i++ {
		err = handler.AddHeader([]byte(fmt.Sprintf("hash%d", i)), incomingHeader)
		require.Nil(t, err)
		require.Equal(t, i+1, addInTxPoolCt)
	}

	require.Equal(t, uint64(10), handler.scrProc.nonce)
}
