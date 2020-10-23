package track_test

import (
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"

	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	processBlock "github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func CreateBlockProcessorMockArguments() track.ArgBlockProcessor {
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	argsHeaderValidator := processBlock.ArgsHeaderValidator{
		Hasher:      &mock.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
	}
	headerValidator, _ := processBlock.NewHeaderValidator(argsHeaderValidator)

	arguments := track.ArgBlockProcessor{
		HeaderValidator:  headerValidator,
		RequestHandler:   &mock.RequestHandlerStub{},
		ShardCoordinator: shardCoordinatorMock,
		BlockTracker:     &mock.BlockTrackerHandlerMock{},
		CrossNotarizer:   &mock.BlockNotarizerHandlerMock{},
		SelfNotarizer:    &mock.BlockNotarizerHandlerMock{},
		CrossNotarizedHeadersNotifier: &mock.BlockNotifierHandlerStub{
			GetNumRegisteredHandlersCalled: func() int {
				return 1
			},
		},
		SelfNotarizedFromCrossHeadersNotifier: &mock.BlockNotifierHandlerStub{
			GetNumRegisteredHandlersCalled: func() int {
				return 1
			},
		},
		SelfNotarizedHeadersNotifier: &mock.BlockNotifierHandlerStub{
			GetNumRegisteredHandlersCalled: func() int {
				return 1
			},
		},
		FinalMetachainHeadersNotifier: &mock.BlockNotifierHandlerStub{
			GetNumRegisteredHandlersCalled: func() int {
				return 1
			},
		},
		Rounder: &mock.RounderMock{},
	}

	return arguments
}

func TestNewBlockProcessor_ShouldErrNilHeaderValidator(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.HeaderValidator = nil
	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, process.ErrNilHeaderValidator, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrNilRequestHandler(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.RequestHandler = nil
	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, process.ErrNilRequestHandler, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrNilShardCoordinator(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.ShardCoordinator = nil
	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrNilBlockTrackerHandler(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.BlockTracker = nil
	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrNilBlockTrackerHandler, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrNilCrossNotarizer(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.CrossNotarizer = nil
	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrNilCrossNotarizer, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrNilSelfNotarizer(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.SelfNotarizer = nil
	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrNilSelfNotarizer, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrCrossNotarizedHeadersNotifier(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.CrossNotarizedHeadersNotifier = nil
	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrNilCrossNotarizedHeadersNotifier, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrSelfNotarizedFromCrossHeadersNotifier(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.SelfNotarizedFromCrossHeadersNotifier = nil
	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrNilSelfNotarizedFromCrossHeadersNotifier, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrSelfNotarizedHeadersNotifier(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.SelfNotarizedHeadersNotifier = nil
	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrNilSelfNotarizedHeadersNotifier, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrFinalMetachainHeadersNotifier(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.FinalMetachainHeadersNotifier = nil
	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrNilFinalMetachainHeadersNotifier, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrNilRounder(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.Rounder = nil
	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrNilRounder, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Nil(t, err)
	assert.NotNil(t, bp)
}

func TestProcessReceivedHeader_ShouldWorkWhenHeaderIsFromSelfShard(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()

	called := false
	blockProcessorArguments.BlockTracker = &mock.BlockTrackerHandlerMock{
		ComputeLongestSelfChainCalled: func() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
			called = true
			return nil, nil, nil, nil
		},
	}
	blockProcessorArguments.SelfNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &dataBlock.Header{}, nil, nil
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	bp.ProcessReceivedHeader(&dataBlock.Header{Nonce: 1})

	assert.True(t, called)
}

func TestProcessReceivedHeader_ShouldWorkWhenHeaderIsFromCrossShard(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()

	called := false
	blockProcessorArguments.CrossNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			called = true
			return &dataBlock.MetaBlock{}, []byte(""), nil
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	bp.ProcessReceivedHeader(&dataBlock.MetaBlock{Nonce: 1})

	assert.True(t, called)
}

func TestDoJobOnReceivedHeader_ShouldWork(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()

	header := &dataBlock.Header{
		ShardID: blockProcessorArguments.ShardCoordinator.SelfId(),
	}

	blockProcessorArguments.BlockTracker = &mock.BlockTrackerHandlerMock{
		ComputeLongestSelfChainCalled: func() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
			return nil, nil, []data.HeaderHandler{header}, nil
		},
	}

	called := false
	blockProcessorArguments.SelfNotarizedHeadersNotifier = &mock.BlockNotifierHandlerStub{
		CallHandlersCalled: func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
			if shardID == blockProcessorArguments.ShardCoordinator.SelfId() {
				called = true
			}
		},
		GetNumRegisteredHandlersCalled: func() int {
			return 1
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	bp.DoJobOnReceivedHeader(blockProcessorArguments.ShardCoordinator.SelfId())

	assert.True(t, called)
}

func TestDoJobOnReceivedCrossNotarizedHeader_ShouldWork(t *testing.T) {
	t.Parallel()

	hasherMock := &mock.HasherMock{}
	marshalizerMock := &mock.MarshalizerMock{}

	blockProcessorArguments := CreateBlockProcessorMockArguments()

	header := &dataBlock.Header{
		Round:   1,
		Nonce:   1,
		ShardID: blockProcessorArguments.ShardCoordinator.SelfId(),
	}
	headerMarshalized, _ := marshalizerMock.Marshal(header)
	headerHash := hasherMock.Compute(string(headerMarshalized))
	headerInfo := track.HeaderInfo{Hash: headerHash, Header: header}

	metaBlock1 := &dataBlock.MetaBlock{
		Round: 1,
		Nonce: 1,
	}
	metaBlock1Marshalized, _ := marshalizerMock.Marshal(metaBlock1)
	metaBlockHash1 := hasherMock.Compute(string(metaBlock1Marshalized))

	metaBlock2 := &dataBlock.MetaBlock{
		Round:    2,
		Nonce:    2,
		PrevHash: metaBlockHash1,
	}
	metaBlock2Marshalized, _ := marshalizerMock.Marshal(metaBlock2)
	metaBlockHash2 := hasherMock.Compute(string(metaBlock2Marshalized))

	metaBlock3 := &dataBlock.MetaBlock{
		Round:    3,
		Nonce:    3,
		PrevHash: metaBlockHash2,
	}
	metaBlock3Marshalized, _ := marshalizerMock.Marshal(metaBlock3)
	metaBlockHash3 := hasherMock.Compute(string(metaBlock3Marshalized))

	blockProcessorArguments.CrossNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return metaBlock1, metaBlockHash1, nil
		},
	}

	blockProcessorArguments.BlockTracker = &mock.BlockTrackerHandlerMock{
		SortHeadersFromNonceCalled: func(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
			return []data.HeaderHandler{metaBlock2, metaBlock3}, [][]byte{metaBlockHash2, metaBlockHash3}
		},
		GetSelfHeadersCalled: func(headerHandler data.HeaderHandler) []*track.HeaderInfo {
			return []*track.HeaderInfo{&headerInfo}
		},
	}

	called := 0

	blockProcessorArguments.SelfNotarizedHeadersNotifier = &mock.BlockNotifierHandlerStub{
		CallHandlersCalled: func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
			called++
		},
		GetNumRegisteredHandlersCalled: func() int {
			return 1
		},
	}

	blockProcessorArguments.CrossNotarizedHeadersNotifier = &mock.BlockNotifierHandlerStub{
		CallHandlersCalled: func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
			called++
		},
		GetNumRegisteredHandlersCalled: func() int {
			return 1
		},
	}

	blockProcessorArguments.SelfNotarizedFromCrossHeadersNotifier = &mock.BlockNotifierHandlerStub{
		CallHandlersCalled: func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
			called++
		},
		GetNumRegisteredHandlersCalled: func() int {
			return 1
		},
	}

	blockProcessorArguments.FinalMetachainHeadersNotifier = &mock.BlockNotifierHandlerStub{
		CallHandlersCalled: func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
			called++
		},
		GetNumRegisteredHandlersCalled: func() int {
			return 1
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	bp.DoJobOnReceivedCrossNotarizedHeader(core.MetachainShardId)

	assert.Equal(t, 2, called)
}

func TestComputeLongestChainFromLastCrossNotarized_ShouldReturnNil(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.CrossNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return nil, nil, errors.New("error")
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, headers, hashes :=
		bp.ComputeLongestChainFromLastCrossNotarized(blockProcessorArguments.ShardCoordinator.SelfId())

	assert.Nil(t, lastCrossNotarizedHeader)
	assert.Nil(t, lastCrossNotarizedHeaderHash)
	assert.Nil(t, headers)
	assert.Nil(t, hashes)
}

func TestComputeLongestChainFromLastCrossNotarized_ShouldWork(t *testing.T) {
	t.Parallel()

	hasherMock := &mock.HasherMock{}
	marshalizerMock := &mock.MarshalizerMock{}

	blockProcessorArguments := CreateBlockProcessorMockArguments()

	metaBlock1 := &dataBlock.MetaBlock{
		Round: 1,
		Nonce: 1,
	}
	metaBlock1Marshalized, _ := marshalizerMock.Marshal(metaBlock1)
	metaBlockHash1 := hasherMock.Compute(string(metaBlock1Marshalized))

	metaBlock2 := &dataBlock.MetaBlock{
		Round:    2,
		Nonce:    2,
		PrevHash: metaBlockHash1,
	}
	metaBlock2Marshalized, _ := marshalizerMock.Marshal(metaBlock2)
	metaBlockHash2 := hasherMock.Compute(string(metaBlock2Marshalized))

	metaBlock3 := &dataBlock.MetaBlock{
		Round:    3,
		Nonce:    3,
		PrevHash: metaBlockHash2,
	}
	metaBlock3Marshalized, _ := marshalizerMock.Marshal(metaBlock3)
	metaBlockHash3 := hasherMock.Compute(string(metaBlock3Marshalized))

	blockProcessorArguments.CrossNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return metaBlock1, metaBlockHash1, nil
		},
	}

	blockProcessorArguments.BlockTracker = &mock.BlockTrackerHandlerMock{
		SortHeadersFromNonceCalled: func(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
			return []data.HeaderHandler{metaBlock2, metaBlock3}, [][]byte{metaBlockHash2, metaBlockHash3}
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, headers, hashes :=
		bp.ComputeLongestChainFromLastCrossNotarized(blockProcessorArguments.ShardCoordinator.SelfId())

	assert.Equal(t, metaBlock1, lastCrossNotarizedHeader)
	assert.Equal(t, metaBlockHash1, lastCrossNotarizedHeaderHash)
	assert.Equal(t, metaBlock2, headers[0])
	assert.Equal(t, metaBlockHash2, hashes[0])
}

func TestComputeSelfNotarizedHeaders_ShouldWork(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()

	header1 := &dataBlock.Header{Nonce: 1}
	hash1 := []byte("hash1")
	header2 := &dataBlock.Header{Nonce: 2}
	hash2 := []byte("hash2")
	headerInfo1 := track.HeaderInfo{Hash: hash1, Header: header1}
	headerInfo2 := track.HeaderInfo{Hash: hash2, Header: header2}

	blockProcessorArguments.BlockTracker = &mock.BlockTrackerHandlerMock{
		GetSelfHeadersCalled: func(headerHandler data.HeaderHandler) []*track.HeaderInfo {
			return []*track.HeaderInfo{&headerInfo2, &headerInfo1}
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	headers, hashes := bp.ComputeSelfNotarizedHeaders([]data.HeaderHandler{&dataBlock.MetaBlock{}})

	require.Equal(t, 2, len(headers))
	assert.Equal(t, header1, headers[0])
	assert.Equal(t, hash1, hashes[0])
	assert.Equal(t, header2, headers[1])
	assert.Equal(t, hash2, hashes[1])
}

func TestComputeSelfNotarizedHeaders_ShouldReturnEmptySliceWhenHeaderIsNil(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	headers, _ := bp.ComputeLongestChain(blockProcessorArguments.ShardCoordinator.SelfId(), nil)

	assert.Equal(t, 0, len(headers))
}

func TestComputeSelfNotarizedHeaders_ShouldReturnEmptySliceWhenSortHeadersFromNonceReturnEmptySlice(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	headers, _ := bp.ComputeLongestChain(blockProcessorArguments.ShardCoordinator.SelfId(), &dataBlock.Header{})

	assert.Equal(t, 0, len(headers))
}

func TestBlockProcessorComputeLongestChain_ShouldWork(t *testing.T) {
	t.Parallel()

	hasherMock := &mock.HasherMock{}
	marshalizerMock := &mock.MarshalizerMock{}

	blockProcessorArguments := CreateBlockProcessorMockArguments()

	header1 := &dataBlock.Header{
		Round: 1,
		Nonce: 1,
	}
	header1Marshalized, _ := marshalizerMock.Marshal(header1)
	headerHash1 := hasherMock.Compute(string(header1Marshalized))

	header2 := &dataBlock.Header{
		Round:    2,
		Nonce:    2,
		PrevHash: headerHash1,
	}
	header2Marshalized, _ := marshalizerMock.Marshal(header2)
	headerHash2 := hasherMock.Compute(string(header2Marshalized))

	header3 := &dataBlock.Header{
		Round:    3,
		Nonce:    3,
		PrevHash: headerHash2,
	}
	header3Marshalized, _ := marshalizerMock.Marshal(header3)
	headerHash3 := hasherMock.Compute(string(header3Marshalized))

	blockProcessorArguments.BlockTracker = &mock.BlockTrackerHandlerMock{
		SortHeadersFromNonceCalled: func(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
			return []data.HeaderHandler{header2, header3}, [][]byte{headerHash2, headerHash3}
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	headers, hashes := bp.ComputeLongestChain(blockProcessorArguments.ShardCoordinator.SelfId(), header1)

	require.Equal(t, 1, len(headers))
	assert.Equal(t, header2, headers[0])
	assert.Equal(t, headerHash2, hashes[0])
}

func TestGetNextHeader_ShouldReturnEmptySliceWhenPrevHeaderIsNil(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	longestChainHeadersIndexes := make([]int, 0)
	headersIndexes := make([]int, 0)
	sortedHeaders := []data.HeaderHandler{&dataBlock.Header{Nonce: 1}}
	bp.GetNextHeader(&longestChainHeadersIndexes, headersIndexes, nil, sortedHeaders, 0)

	assert.Equal(t, 0, len(longestChainHeadersIndexes))
}

func TestGetNextHeader_ShouldReturnEmptySliceWhenSortedHeadersHaveHigherNonces(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	longestChainHeadersIndexes := make([]int, 0)
	headersIndexes := make([]int, 0)
	prevHeader := &dataBlock.Header{}
	sortedHeaders := []data.HeaderHandler{&dataBlock.Header{Nonce: 2}}
	bp.GetNextHeader(&longestChainHeadersIndexes, headersIndexes, prevHeader, sortedHeaders, 0)

	assert.Equal(t, 0, len(longestChainHeadersIndexes))
}

func TestGetNextHeader_ShouldReturnEmptySliceWhenHeaderConstructionIsNotValid(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	longestChainHeadersIndexes := make([]int, 0)
	headersIndexes := make([]int, 0)
	prevHeader := &dataBlock.Header{}
	sortedHeaders := []data.HeaderHandler{&dataBlock.Header{Nonce: 1}}
	bp.GetNextHeader(&longestChainHeadersIndexes, headersIndexes, prevHeader, sortedHeaders, 0)

	assert.Equal(t, 0, len(longestChainHeadersIndexes))
}

func TestGetNextHeader_ShouldReturnEmptySliceWhenHeaderFinalityIsNotChecked(t *testing.T) {
	t.Parallel()

	hasherMock := &mock.HasherMock{}
	marshalizerMock := &mock.MarshalizerMock{}

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	longestChainHeadersIndexes := make([]int, 0)
	headersIndexes := make([]int, 0)

	header1 := &dataBlock.Header{
		Round: 1,
		Nonce: 1,
	}
	header1Marshalized, _ := marshalizerMock.Marshal(header1)
	headerHash1 := hasherMock.Compute(string(header1Marshalized))

	header2 := &dataBlock.Header{
		Round:    2,
		Nonce:    2,
		PrevHash: headerHash1,
	}

	sortedHeaders := []data.HeaderHandler{header2}
	bp.GetNextHeader(&longestChainHeadersIndexes, headersIndexes, header1, sortedHeaders, 0)

	assert.Equal(t, 0, len(longestChainHeadersIndexes))
}

func TestGetNextHeader_ShouldWork(t *testing.T) {
	t.Parallel()

	hasherMock := &mock.HasherMock{}
	marshalizerMock := &mock.MarshalizerMock{}

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	longestChainHeadersIndexes := make([]int, 0)
	headersIndexes := make([]int, 0)

	header1 := &dataBlock.Header{
		Round: 1,
		Nonce: 1,
	}
	header1Marshalized, _ := marshalizerMock.Marshal(header1)
	headerHash1 := hasherMock.Compute(string(header1Marshalized))

	header2 := &dataBlock.Header{
		Round:    2,
		Nonce:    2,
		PrevHash: headerHash1,
	}
	header2Marshalized, _ := marshalizerMock.Marshal(header2)
	headerHash2 := hasherMock.Compute(string(header2Marshalized))

	header3 := &dataBlock.Header{
		Round:    3,
		Nonce:    3,
		PrevHash: headerHash2,
	}

	sortedHeaders := []data.HeaderHandler{header2, header3}
	bp.GetNextHeader(&longestChainHeadersIndexes, headersIndexes, header1, sortedHeaders, 0)

	require.Equal(t, 1, len(longestChainHeadersIndexes))
	assert.Equal(t, 0, longestChainHeadersIndexes[0])
}

func TestCheckHeaderFinality_ShouldErrNilBlockHeader(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	sortedHeaders := []data.HeaderHandler{&dataBlock.Header{Nonce: 1}}
	err := bp.CheckHeaderFinality(nil, sortedHeaders, 0)

	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestCheckHeaderFinality_ShouldErrHeaderNotFinal(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	header := &dataBlock.Header{}
	sortedHeaders := []data.HeaderHandler{&dataBlock.Header{Nonce: 1}}
	err := bp.CheckHeaderFinality(header, sortedHeaders, 0)

	assert.Equal(t, process.ErrHeaderNotFinal, err)
}

func TestCheckHeaderFinality_ShouldWork(t *testing.T) {
	t.Parallel()

	hasherMock := &mock.HasherMock{}
	marshalizerMock := &mock.MarshalizerMock{}

	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	header1 := &dataBlock.Header{
		Round: 1,
		Nonce: 1,
	}
	header1Marshalized, _ := marshalizerMock.Marshal(header1)
	headerHash1 := hasherMock.Compute(string(header1Marshalized))

	header2 := &dataBlock.Header{
		Round:    2,
		Nonce:    2,
		PrevHash: headerHash1,
	}

	sortedHeaders := []data.HeaderHandler{header2}
	err := bp.CheckHeaderFinality(header1, sortedHeaders, 0)

	assert.Nil(t, err)
}

func TestRequestHeadersIfNeeded_ShouldNotRequestIfHeaderIsNil(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()

	called := false
	blockProcessorArguments.RequestHandler = &mock.RequestHandlerStub{
		RequestMetaHeaderByNonceCalled: func(nonce uint64) {
			called = true
		},
		RequestShardHeaderByNonceCalled: func(shardId uint32, nonce uint64) {
			called = true
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	sortedHeaders := []data.HeaderHandler{&dataBlock.Header{}}
	longestChainHeaders := []data.HeaderHandler{&dataBlock.Header{}}
	bp.RequestHeadersIfNeeded(nil, sortedHeaders, longestChainHeaders)
	time.Sleep(50 * time.Millisecond)

	assert.False(t, called)
}

func TestRequestHeadersIfNeeded_ShouldNotRequestIfSortedHeadersAreEmpty(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()

	called := false
	blockProcessorArguments.RequestHandler = &mock.RequestHandlerStub{
		RequestMetaHeaderByNonceCalled: func(nonce uint64) {
			called = true
		},
		RequestShardHeaderByNonceCalled: func(shardId uint32, nonce uint64) {
			called = true
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	lastNotarizedHeader := &dataBlock.Header{}
	longestChainHeaders := []data.HeaderHandler{&dataBlock.Header{}}
	bp.RequestHeadersIfNeeded(lastNotarizedHeader, nil, longestChainHeaders)
	time.Sleep(50 * time.Millisecond)

	assert.False(t, called)
}

func TestRequestHeadersIfNeeded_ShouldNotRequestIfNodeIsSync(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()

	called := false
	blockProcessorArguments.RequestHandler = &mock.RequestHandlerStub{
		RequestMetaHeaderByNonceCalled: func(nonce uint64) {
			called = true
		},
		RequestShardHeaderByNonceCalled: func(shardId uint32, nonce uint64) {
			called = true
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	lastNotarizedHeader := &dataBlock.Header{}
	sortedHeaders := []data.HeaderHandler{&dataBlock.Header{}}
	longestChainHeaders := []data.HeaderHandler{&dataBlock.Header{}}
	bp.RequestHeadersIfNeeded(lastNotarizedHeader, sortedHeaders, longestChainHeaders)
	time.Sleep(50 * time.Millisecond)

	assert.False(t, called)
}

func TestRequestHeadersIfNeeded_ShouldNotRequestIfLongestChainHasAdvanced(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()

	called := false
	blockProcessorArguments.RequestHandler = &mock.RequestHandlerStub{
		RequestMetaHeaderByNonceCalled: func(nonce uint64) {
			called = true
		},
		RequestShardHeaderByNonceCalled: func(shardId uint32, nonce uint64) {
			called = true
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	lastNotarizedHeader := &dataBlock.Header{}
	longestChainHeaders := []data.HeaderHandler{&dataBlock.Header{Nonce: 1}}
	sortedHeaders := []data.HeaderHandler{&dataBlock.Header{Nonce: 3}}
	bp.RequestHeadersIfNeeded(lastNotarizedHeader, sortedHeaders, longestChainHeaders)
	time.Sleep(50 * time.Millisecond)

	assert.False(t, called)
}

func TestRequestHeadersIfNeeded_ShouldRequestIfLongestChainHasNotAdvanced(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()

	var mutCalled sync.RWMutex
	var wg sync.WaitGroup

	wg.Add(2)

	mutCalled.Lock()
	calledMeta := false
	calledShard := false
	mutCalled.Unlock()

	blockProcessorArguments.RequestHandler = &mock.RequestHandlerStub{
		RequestMetaHeaderByNonceCalled: func(nonce uint64) {
			wg.Done()
			mutCalled.Lock()
			calledMeta = true
			mutCalled.Unlock()
		},
		RequestShardHeaderByNonceCalled: func(shardId uint32, nonce uint64) {
			wg.Done()
			mutCalled.Lock()
			calledShard = true
			mutCalled.Unlock()
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	lastNotarizedHeader := &dataBlock.Header{}
	sortedHeaders := []data.HeaderHandler{&dataBlock.Header{Nonce: 2}}
	bp.RequestHeadersIfNeeded(lastNotarizedHeader, sortedHeaders, nil)
	wg.Wait()

	mutCalled.RLock()
	assert.False(t, calledMeta)
	assert.True(t, calledShard)
	mutCalled.RUnlock()

	wg.Add(2)

	mutCalled.Lock()
	calledMeta = false
	calledShard = false
	mutCalled.Unlock()

	lastNotarizedHeader2 := &dataBlock.MetaBlock{}
	sortedHeaders2 := []data.HeaderHandler{&dataBlock.MetaBlock{Nonce: 2}}
	bp.RequestHeadersIfNeeded(lastNotarizedHeader2, sortedHeaders2, nil)
	wg.Wait()

	mutCalled.RLock()
	assert.True(t, calledMeta)
	assert.False(t, calledShard)
	mutCalled.RUnlock()
}

func TestRequestHeadersIfNothingNewIsReceived_ShouldNotRequestIfHighestRoundFromReceivedHeadersIsNearToChronologyRound(t *testing.T) {
	t.Parallel()

	testRequestHeaders(t, 3, 3, 3)
}

func TestRequestHeadersIfNothingNewIsReceived_ShouldNotRequestIfLastNotarizedHeaderNonceIsFarFromLatestValidHeaderNonce(t *testing.T) {
	t.Parallel()

	testRequestHeaders(
		t,
		process.MaxHeadersToRequestInAdvance+4,
		process.MaxHeadersToRequestInAdvance+3,
		process.MaxHeadersToRequestInAdvance+3,
	)
}

func testRequestHeaders(t *testing.T, roundIndex uint64, round uint64, nonce uint64) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()

	called := false
	blockProcessorArguments.RequestHandler = &mock.RequestHandlerStub{
		RequestMetaHeaderByNonceCalled: func(nonce uint64) {
			called = true
		},
		RequestShardHeaderByNonceCalled: func(shardId uint32, nonce uint64) {
			called = true
		},
	}

	blockProcessorArguments.Rounder = &mock.RounderMock{
		RoundIndex: process.MaxRoundsWithoutNewBlockReceived + int64(roundIndex),
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	lastNotarizedHeader := &dataBlock.Header{Nonce: 1, Round: 1}
	sortedReceivedHeaders := []data.HeaderHandler{&dataBlock.Header{Nonce: nonce, Round: round}}
	longestChainHeaders := []data.HeaderHandler{&dataBlock.Header{Nonce: nonce - 1, Round: round - 1}}
	latestValidHeader := bp.GetLatestValidHeader(lastNotarizedHeader, longestChainHeaders)
	highestRound := bp.GetHighestRoundInReceivedHeaders(latestValidHeader, sortedReceivedHeaders)
	bp.RequestHeadersIfNothingNewIsReceived(lastNotarizedHeader.GetNonce(), latestValidHeader, highestRound)
	time.Sleep(50 * time.Millisecond)

	assert.False(t, called)
}

func TestRequestHeadersIfNothingNewIsReceived_ShouldRequestIfHighestRoundFromReceivedHeadersIsFarFromChronologyRound(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()

	var mutCalled sync.RWMutex
	var wg sync.WaitGroup

	wg.Add(2)

	mutCalled.Lock()
	called := false
	mutCalled.Unlock()

	blockProcessorArguments.RequestHandler = &mock.RequestHandlerStub{
		RequestMetaHeaderByNonceCalled: func(nonce uint64) {
			wg.Done()
			mutCalled.Lock()
			called = true
			mutCalled.Unlock()
		},
		RequestShardHeaderByNonceCalled: func(shardId uint32, nonce uint64) {
			wg.Done()
			mutCalled.Lock()
			called = true
			mutCalled.Unlock()
		},
	}

	blockProcessorArguments.Rounder = &mock.RounderMock{
		RoundIndex: process.MaxRoundsWithoutNewBlockReceived + 4,
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	lastNotarizedHeader := &dataBlock.Header{Nonce: 1, Round: 1}
	sortedReceivedHeaders := []data.HeaderHandler{&dataBlock.Header{Nonce: 3, Round: 3}}
	longestChainHeaders := []data.HeaderHandler{&dataBlock.Header{Nonce: 2, Round: 2}}
	latestValidHeader := bp.GetLatestValidHeader(lastNotarizedHeader, longestChainHeaders)
	highestRound := bp.GetHighestRoundInReceivedHeaders(latestValidHeader, sortedReceivedHeaders)
	bp.RequestHeadersIfNothingNewIsReceived(lastNotarizedHeader.GetNonce(), latestValidHeader, highestRound)
	wg.Wait()

	mutCalled.RLock()
	assert.True(t, called)
	mutCalled.RUnlock()
}

func TestRequestHeaders_ShouldAddAndRequestForShardHeaders(t *testing.T) {
	t.Parallel()

	var mutRequest sync.Mutex

	blockProcessorArguments := CreateBlockProcessorMockArguments()

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
	blockProcessorArguments.RequestHandler = &mock.RequestHandlerStub{
		RequestShardHeaderByNonceCalled: func(shardId uint32, nonce uint64) {
			mutRequest.Lock()
			shardIDRequestCalled = append(shardIDRequestCalled, shardId)
			nonceRequestCalled = append(nonceRequestCalled, nonce)
			mutRequest.Unlock()
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	shardID := uint32(1)
	fromNonce := uint64(1)

	bp.RequestHeaders(shardID, fromNonce)

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

func TestRequestHeaders_ShouldAddAndRequestForMetaHeaders(t *testing.T) {
	t.Parallel()

	var mutRequest sync.Mutex

	blockProcessorArguments := CreateBlockProcessorMockArguments()

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
	blockProcessorArguments.RequestHandler = &mock.RequestHandlerStub{
		RequestMetaHeaderByNonceCalled: func(nonce uint64) {
			mutRequest.Lock()
			shardIDRequestCalled = append(shardIDRequestCalled, core.MetachainShardId)
			nonceRequestCalled = append(nonceRequestCalled, nonce)
			mutRequest.Unlock()
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	shardID := core.MetachainShardId
	fromNonce := uint64(1)

	bp.RequestHeaders(shardID, fromNonce)

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

func TestShouldProcessReceivedHeader_ShouldReturnFalseWhenGetLastNotarizedHeaderFails(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()

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

	assert.False(t, bp.ShouldProcessReceivedHeader(&dataBlock.Header{}))
	assert.False(t, bp.ShouldProcessReceivedHeader(&dataBlock.MetaBlock{}))
}

func TestShouldProcessReceivedHeader_ShouldWork(t *testing.T) {
	t.Parallel()

	blockProcessorArguments := CreateBlockProcessorMockArguments()

	blockProcessorArguments.SelfNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &dataBlock.Header{Nonce: 15}, []byte(""), nil
		},
	}

	blockProcessorArguments.CrossNotarizer = &mock.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &dataBlock.MetaBlock{Nonce: 10}, []byte(""), nil
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	assert.False(t, bp.ShouldProcessReceivedHeader(&dataBlock.Header{Nonce: 14}))
	assert.False(t, bp.ShouldProcessReceivedHeader(&dataBlock.Header{Nonce: 15}))
	assert.True(t, bp.ShouldProcessReceivedHeader(&dataBlock.Header{Nonce: 16}))

	assert.False(t, bp.ShouldProcessReceivedHeader(&dataBlock.MetaBlock{Nonce: 9}))
	assert.False(t, bp.ShouldProcessReceivedHeader(&dataBlock.MetaBlock{Nonce: 10}))
	assert.True(t, bp.ShouldProcessReceivedHeader(&dataBlock.MetaBlock{Nonce: 11}))
}
