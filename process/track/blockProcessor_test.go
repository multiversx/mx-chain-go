package track_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	block2 "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func CreateBlockProcessorMockArguments() track.ArgBlockProcessor {
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:      &mock.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
	}
	headerValidator, _ := block.NewHeaderValidator(argsHeaderValidator)

	arguments := track.ArgBlockProcessor{
		HeaderValidator:               headerValidator,
		RequestHandler:                &mock.RequestHandlerStub{},
		ShardCoordinator:              shardCoordinatorMock,
		BlockTracker:                  &track.BlockTrackerHandlerMock{},
		CrossNotarizer:                &track.BlockNotarizerHandlerMock{},
		CrossNotarizedHeadersNotifier: &track.BlockNotifierHandlerMock{},
		SelfNotarizedHeadersNotifier:  &track.BlockNotifierHandlerMock{},
	}

	return arguments
}

func TestNewBlockProcessor_ShouldErrNilHeaderValidator(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.HeaderValidator = nil

	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, process.ErrNilHeaderValidator, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrNilRequestHandler(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.RequestHandler = nil

	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, process.ErrNilRequestHandler, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrNilShardCoordinator(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.ShardCoordinator = nil

	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrNilBlockTrackerHandler(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.BlockTracker = nil

	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrNilBlockTrackerHandler, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrNilCrossNotarizer(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.CrossNotarizer = nil

	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrNilCrossNotarizer, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrCrossNotarizedHeadersNotifier(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.CrossNotarizedHeadersNotifier = nil

	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrCrossNotarizedHeadersNotifier, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrSelfNotarizedHeadersNotifier(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.SelfNotarizedHeadersNotifier = nil

	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrSelfNotarizedHeadersNotifier, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldWork(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Nil(t, err)
	assert.NotNil(t, bp)
}

func TestProcessReceivedHeader_ShouldWorkWhenHeaderIsFromSelfShard(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	called := false
	blockProcessorArguments.BlockTracker = &track.BlockTrackerHandlerMock{
		ComputeLongestSelfChainCalled: func() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
			called = true
			return nil, nil, nil, nil
		},
	}
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	bp.ProcessReceivedHeader(&block2.Header{})

	assert.True(t, called)
}

func TestProcessReceivedHeader_ShouldWorkWhenHeaderIsFromCrossShard(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	called := false
	blockProcessorArguments.CrossNotarizer = &track.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			called = true
			return nil, nil, nil
		},
	}
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	bp.ProcessReceivedHeader(&block2.MetaBlock{})

	assert.True(t, called)
}

func TestDoJobOnReceivedHeader_ShouldWork(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()

	header := &block2.Header{
		ShardId: blockProcessorArguments.ShardCoordinator.SelfId(),
	}

	blockProcessorArguments.BlockTracker = &track.BlockTrackerHandlerMock{
		ComputeLongestSelfChainCalled: func() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
			return nil, nil, []data.HeaderHandler{header}, nil
		},
	}
	called := false
	blockProcessorArguments.SelfNotarizedHeadersNotifier = &track.BlockNotifierHandlerMock{
		CallHandlersCalled: func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
			if shardID == blockProcessorArguments.ShardCoordinator.SelfId() {
				called = true
			}
		},
	}
	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	bp.DoJobOnReceivedHeader(blockProcessorArguments.ShardCoordinator.SelfId())

	assert.True(t, called)
}

func TestDoJobOnReceivedCrossNotarizedHeader_ShouldWork(t *testing.T) {
	hasherMock := &mock.HasherMock{}
	marshalizerMock := &mock.MarshalizerMock{}

	blockProcessorArguments := CreateBlockProcessorMockArguments()

	header := &block2.Header{
		Round:   1,
		Nonce:   1,
		ShardId: blockProcessorArguments.ShardCoordinator.SelfId(),
	}
	headerMarshalized, _ := marshalizerMock.Marshal(header)
	headerHash := hasherMock.Compute(string(headerMarshalized))
	headerInfo := track.HeaderInfo{Hash: headerHash, Header: header}

	metaBlock1 := &block2.MetaBlock{
		Round: 1,
		Nonce: 1,
	}
	metaBlock1Marshalized, _ := marshalizerMock.Marshal(metaBlock1)
	metaBlockHash1 := hasherMock.Compute(string(metaBlock1Marshalized))

	metaBlock2 := &block2.MetaBlock{
		Round:    2,
		Nonce:    2,
		PrevHash: metaBlockHash1,
	}
	metaBlock2Marshalized, _ := marshalizerMock.Marshal(metaBlock2)
	metaBlockHash2 := hasherMock.Compute(string(metaBlock2Marshalized))

	metaBlock3 := &block2.MetaBlock{
		Round:    3,
		Nonce:    3,
		PrevHash: metaBlockHash2,
	}
	metaBlock3Marshalized, _ := marshalizerMock.Marshal(metaBlock3)
	metaBlockHash3 := hasherMock.Compute(string(metaBlock3Marshalized))

	blockProcessorArguments.CrossNotarizer = &track.BlockNotarizerHandlerMock{
		GetLastNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return metaBlock1, metaBlockHash1, nil
		},
	}

	blockProcessorArguments.BlockTracker = &track.BlockTrackerHandlerMock{
		SortHeadersFromNonceCalled: func(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
			return []data.HeaderHandler{metaBlock2, metaBlock3}, [][]byte{metaBlockHash2, metaBlockHash3}
		},
		GetSelfHeadersCalled: func(headerHandler data.HeaderHandler) []*track.HeaderInfo {
			return []*track.HeaderInfo{&headerInfo}
		},
	}

	called := 0

	blockProcessorArguments.SelfNotarizedHeadersNotifier = &track.BlockNotifierHandlerMock{
		CallHandlersCalled: func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
			called++
		},
	}

	blockProcessorArguments.CrossNotarizedHeadersNotifier = &track.BlockNotifierHandlerMock{
		CallHandlersCalled: func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
			called++
		},
	}

	bp, _ := track.NewBlockProcessor(blockProcessorArguments)

	bp.DoJobOnReceivedCrossNotarizedHeader(sharding.MetachainShardId)

	assert.Equal(t, 2, called)
}
