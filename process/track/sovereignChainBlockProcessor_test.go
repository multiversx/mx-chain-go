package track_test

import (
	"errors"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	processBlock "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CreateSovereignChainBlockProcessorMockArguments -
func CreateSovereignChainBlockProcessorMockArguments() track.ArgBlockProcessor {
	blockProcessorArguments := CreateBlockProcessorMockArguments()

	blockProcessorArguments.RequestHandler = &testscommon.ExtendedShardHeaderRequestHandlerStub{}

	argsHeaderValidator := processBlock.ArgsHeaderValidator{
		Hasher:      &hashingMocks.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
	}
	headerValidator, _ := processBlock.NewHeaderValidator(argsHeaderValidator)
	sovereignChainHeaderValidator, _ := processBlock.NewSovereignChainHeaderValidator(headerValidator)
	blockProcessorArguments.HeaderValidator = sovereignChainHeaderValidator

	return blockProcessorArguments
}

func TestNewSovereignChainBlockProcessor_ShouldErrNilBlockProcessor(t *testing.T) {
	t.Parallel()

	scpb, err := track.NewSovereignChainBlockProcessor(nil)
	assert.Nil(t, scpb)
	assert.Equal(t, process.ErrNilBlockProcessor, err)
}

func TestNewSovereignChainBlockProcessor_ShouldErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	scpb, err := track.NewSovereignChainBlockProcessor(bp)
	assert.Nil(t, scpb)
	assert.ErrorIs(t, err, process.ErrWrongTypeAssertion)
}

func TestNewSovereignChainBlockProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateSovereignChainBlockProcessorMockArguments()
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	scpb, err := track.NewSovereignChainBlockProcessor(bp)
	assert.NotNil(t, scpb)
	assert.Nil(t, err)
}

func TestSovereignChainBlockProcessor_ShouldProcessReceivedHeaderShouldReturnFalseWhenGetLastNotarizedHeaderFails(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateSovereignChainBlockProcessorMockArguments()

	blockProcessorArguments.SelfNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return nil, nil, errors.New("error")
		},
	}

	blockProcessorArguments.CrossNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return nil, nil, errors.New("error")
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)
	scbp, _ := track.NewSovereignChainBlockProcessor(bp)

	assert.False(t, scbp.ShouldProcessReceivedHeader(&block.Header{}))
	assert.False(t, scbp.ShouldProcessReceivedHeader(&block.ShardHeaderExtended{}))
}

func TestSovereignChainBlockProcessor_ShouldProcessReceivedHeaderShouldWork(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateSovereignChainBlockProcessorMockArguments()

	blockProcessorArguments.SelfNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.Header{Nonce: 15}, []byte(""), nil
		},
	}

	blockProcessorArguments.CrossNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			shardHeaderExtended := &block.ShardHeaderExtended{
				Header: &block.HeaderV2{
					Header: &block.Header{Nonce: 10},
				},
			}
			return shardHeaderExtended, []byte(""), nil
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)
	scbp, _ := track.NewSovereignChainBlockProcessor(bp)

	assert.False(t, scbp.ShouldProcessReceivedHeader(&block.Header{Nonce: 14}))
	assert.False(t, scbp.ShouldProcessReceivedHeader(&block.Header{Nonce: 15}))
	assert.True(t, scbp.ShouldProcessReceivedHeader(&block.Header{Nonce: 16}))

	shardHeaderExtended := &block.ShardHeaderExtended{
		Header: &block.HeaderV2{
			Header: &block.Header{},
		},
	}

	_ = shardHeaderExtended.SetNonce(9)
	assert.False(t, scbp.ShouldProcessReceivedHeader(shardHeaderExtended))

	_ = shardHeaderExtended.SetNonce(10)
	assert.False(t, scbp.ShouldProcessReceivedHeader(shardHeaderExtended))

	_ = shardHeaderExtended.SetNonce(11)
	assert.True(t, scbp.ShouldProcessReceivedHeader(shardHeaderExtended))
}

func TestSovereignChainBlockProcessor_ProcessReceivedHeaderShouldWorkWhenHeaderIsFromSelfShard(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateSovereignChainBlockProcessorMockArguments()

	called := false
	blockProcessorArguments.BlockTracker = &mock.BlockTrackerHandlerMock{
		ComputeLongestSelfChainCalled: func() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
			called = true
			return nil, nil, nil, nil
		},
	}
	blockProcessorArguments.SelfNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.Header{}, nil, nil
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)
	scbp, _ := track.NewSovereignChainBlockProcessor(bp)

	scbp.ProcessReceivedHeader(&block.Header{Nonce: 1})

	assert.True(t, called)
}

func TestSovereignChainBlockProcessor_ProcessReceivedHeaderShouldWorkWhenHeaderIsFromCrossShard(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateSovereignChainBlockProcessorMockArguments()

	called := false
	blockProcessorArguments.BlockTracker = &mock.BlockTrackerHandlerMock{
		SortHeadersFromNonceCalled: func(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
			called = true
			return nil, nil
		},
	}
	blockProcessorArguments.CrossNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			shardHeaderExtended := &block.ShardHeaderExtended{
				Header: &block.HeaderV2{
					Header: &block.Header{},
				},
			}
			return shardHeaderExtended, []byte(""), nil
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)
	scbp, _ := track.NewSovereignChainBlockProcessor(bp)

	shardHeaderExtended := &block.ShardHeaderExtended{
		Header: &block.HeaderV2{
			Header: &block.Header{Nonce: 1},
		},
	}

	scbp.ProcessReceivedHeader(shardHeaderExtended)

	assert.True(t, called)
}

func TestSovereignChainBlockProcessor_DoJobOnReceivedCrossNotarizedHeaderShouldWork(t *testing.T) {
	t.Parallel()

	hasherMock := &hashingMocks.HasherMock{}
	marshalizerMock := &mock.MarshalizerMock{}

	blockProcessorArguments := CreateSovereignChainBlockProcessorMockArguments()

	shardHeaderExtended1 := &block.ShardHeaderExtended{
		Header: &block.HeaderV2{
			Header: &block.Header{
				Round: 1,
				Nonce: 1,
			},
		},
	}

	header1Marshalled, _ := marshalizerMock.Marshal(shardHeaderExtended1.Header)
	headerHash1 := hasherMock.Compute(string(header1Marshalled))

	shardHeaderExtended2 := &block.ShardHeaderExtended{
		Header: &block.HeaderV2{
			Header: &block.Header{
				Round:    2,
				Nonce:    2,
				PrevHash: headerHash1,
			},
		},
	}

	header2Marshalled, _ := marshalizerMock.Marshal(shardHeaderExtended2.Header)
	headerHash2 := hasherMock.Compute(string(header2Marshalled))

	shardHeaderExtended2Marshalled, _ := marshalizerMock.Marshal(shardHeaderExtended2)
	shardHeaderExtendedHash2 := hasherMock.Compute(string(shardHeaderExtended2Marshalled))

	shardHeaderExtended3 := &block.ShardHeaderExtended{
		Header: &block.HeaderV2{
			Header: &block.Header{
				Round:    3,
				Nonce:    3,
				PrevHash: headerHash2,
			},
		},
	}

	shardHeaderExtended3Marshalled, _ := marshalizerMock.Marshal(shardHeaderExtended3)
	shardHeaderExtendedHash3 := hasherMock.Compute(string(shardHeaderExtended3Marshalled))

	blockProcessorArguments.CrossNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return shardHeaderExtended1, headerHash1, nil
		},
	}

	blockProcessorArguments.BlockTracker = &mock.BlockTrackerHandlerMock{
		SortHeadersFromNonceCalled: func(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
			return []data.HeaderHandler{shardHeaderExtended2, shardHeaderExtended3}, [][]byte{shardHeaderExtendedHash2, shardHeaderExtendedHash3}
		},
	}

	wasCalled := false
	blockProcessorArguments.CrossNotarizedHeadersNotifier = &mock.BlockNotifierHandlerStub{
		CallHandlersCalled: func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
			wasCalled = true
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)
	scbp, _ := track.NewSovereignChainBlockProcessor(bp)

	scbp.DoJobOnReceivedCrossNotarizedHeader(core.SovereignChainShardId)

	assert.True(t, wasCalled)
}

func TestSovereignChainBlockProcessor_RequestHeadersShouldAddAndRequestForShardHeaders(t *testing.T) {
	t.Parallel()

	var mutRequest sync.Mutex

	blockProcessorArguments := CreateSovereignChainBlockProcessorMockArguments()

	shardIDAddCalled := make([]uint32, 0)
	nonceAddCalled := make([]uint64, 0)

	blockProcessorArguments.BlockTracker = &mock.BlockTrackerHandlerMock{
		AddHeaderFromPoolCalled: func(shardID uint32, nonce uint64) {
			shardIDAddCalled = append(shardIDAddCalled, shardID)
			nonceAddCalled = append(nonceAddCalled, nonce)
		},
	}

	shardIDRequestCalled := make([]uint32, 0)
	nonceRequestCalled := make([]uint64, 0)
	blockProcessorArguments.RequestHandler = &testscommon.ExtendedShardHeaderRequestHandlerStub{
		RequestHandlerStub: testscommon.RequestHandlerStub{
			RequestShardHeaderByNonceCalled: func(shardId uint32, nonce uint64) {
				mutRequest.Lock()
				shardIDRequestCalled = append(shardIDRequestCalled, shardId)
				nonceRequestCalled = append(nonceRequestCalled, nonce)
				mutRequest.Unlock()
			},
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)
	scbp, _ := track.NewSovereignChainBlockProcessor(bp)

	shardID := uint32(0)
	fromNonce := uint64(1)

	scbp.RequestHeaders(shardID, fromNonce)

	time.Sleep(100 * time.Millisecond)

	mutRequest.Lock()
	sort.Slice(nonceRequestCalled, func(i, j int) bool {
		return nonceRequestCalled[i] < nonceRequestCalled[j]
	})
	mutRequest.Unlock()

	require.Equal(t, 2, len(shardIDAddCalled))
	require.Equal(t, 2, len(nonceAddCalled))
	require.Equal(t, 2, len(shardIDRequestCalled))
	require.Equal(t, 2, len(nonceRequestCalled))

	assert.Equal(t, []uint32{shardID, shardID}, shardIDAddCalled)
	assert.Equal(t, []uint64{fromNonce, fromNonce + 1}, nonceAddCalled)
	assert.Equal(t, []uint32{shardID, shardID}, shardIDRequestCalled)
	assert.Equal(t, []uint64{fromNonce, fromNonce + 1}, nonceRequestCalled)
}

func TestSovereignChainBlockProcessor_RequestHeadersShouldAddAndRequestForExtendedShardHeaders(t *testing.T) {
	t.Parallel()

	var mutRequest sync.Mutex

	blockProcessorArguments := CreateSovereignChainBlockProcessorMockArguments()

	shardIDAddCalled := make([]uint32, 0)
	nonceAddCalled := make([]uint64, 0)

	blockProcessorArguments.BlockTracker = &mock.BlockTrackerHandlerMock{
		AddHeaderFromPoolCalled: func(shardID uint32, nonce uint64) {
			shardIDAddCalled = append(shardIDAddCalled, shardID)
			nonceAddCalled = append(nonceAddCalled, nonce)
		},
	}

	shardIDRequestCalled := make([]uint32, 0)
	nonceRequestCalled := make([]uint64, 0)
	blockProcessorArguments.RequestHandler = &testscommon.ExtendedShardHeaderRequestHandlerStub{
		RequestExtendedShardHeaderByNonceCalled: func(nonce uint64) {
			mutRequest.Lock()
			shardIDRequestCalled = append(shardIDRequestCalled, core.SovereignChainShardId)
			nonceRequestCalled = append(nonceRequestCalled, nonce)
			mutRequest.Unlock()
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)
	scbp, _ := track.NewSovereignChainBlockProcessor(bp)

	shardID := core.SovereignChainShardId
	fromNonce := uint64(1)

	scbp.RequestHeaders(shardID, fromNonce)

	time.Sleep(100 * time.Millisecond)

	mutRequest.Lock()
	sort.Slice(nonceRequestCalled, func(i, j int) bool {
		return nonceRequestCalled[i] < nonceRequestCalled[j]
	})
	mutRequest.Unlock()

	require.Equal(t, 2, len(shardIDAddCalled))
	require.Equal(t, 2, len(nonceAddCalled))
	require.Equal(t, 2, len(shardIDRequestCalled))
	require.Equal(t, 2, len(nonceRequestCalled))

	assert.Equal(t, []uint32{shardID, shardID}, shardIDAddCalled)
	assert.Equal(t, []uint64{fromNonce, fromNonce + 1}, nonceAddCalled)
	assert.Equal(t, []uint32{shardID, shardID}, shardIDRequestCalled)
	assert.Equal(t, []uint64{fromNonce, fromNonce + 1}, nonceRequestCalled)
}
