package poolsCleaner

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func createMockArgMiniBlocksPoolsCleaner() ArgMiniBlocksPoolsCleaner {
	return ArgMiniBlocksPoolsCleaner{
		ArgBasePoolsCleaner: ArgBasePoolsCleaner{
			RoundHandler:                   &mock.RoundHandlerMock{},
			ShardCoordinator:               &mock.CoordinatorStub{},
			MaxRoundsToKeepUnprocessedData: 1,
		},
		MiniblocksPool: testscommon.NewCacherStub(),
	}
}

func TestNewMiniBlocksPoolsCleaner_NilMiniblockPoolShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgMiniBlocksPoolsCleaner()
	args.MiniblocksPool = nil
	miniblockCleaner, err := NewMiniBlocksPoolsCleaner(args)

	assert.Equal(t, process.ErrNilMiniBlockPool, err)
	assert.Nil(t, miniblockCleaner)
}

func TestNewMiniBlocksPoolsCleaner_NilRoundHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgMiniBlocksPoolsCleaner()
	args.RoundHandler = nil
	miniblockCleaner, err := NewMiniBlocksPoolsCleaner(args)

	assert.Equal(t, process.ErrNilRoundHandler, err)
	assert.Nil(t, miniblockCleaner)
}

func TestNewMiniBlocksPoolsCleaner_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgMiniBlocksPoolsCleaner()
	args.ShardCoordinator = nil
	miniblockCleaner, err := NewMiniBlocksPoolsCleaner(args)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, miniblockCleaner)
}

func TestNewMiniBlocksPoolsCleaner_InvalidMaxRoundsToKeepUnprocessedDataShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgMiniBlocksPoolsCleaner()
	args.MaxRoundsToKeepUnprocessedData = 0
	miniblockCleaner, err := NewMiniBlocksPoolsCleaner(args)

	assert.True(t, errors.Is(err, process.ErrInvalidValue))
	assert.True(t, strings.Contains(err.Error(), "MaxRoundsToKeepUnprocessedData"))
	assert.Nil(t, miniblockCleaner)
}

func TestNewMiniBlocksPoolsCleaner_ShouldWork(t *testing.T) {
	t.Parallel()

	miniblockCleaner, err := NewMiniBlocksPoolsCleaner(createMockArgMiniBlocksPoolsCleaner())

	assert.Nil(t, err)
	assert.NotNil(t, miniblockCleaner)
}

func TestReceivedMiniBlock_WrongTypeShouldBeIgnored(t *testing.T) {
	t.Parallel()

	miniblockCleaner, _ := NewMiniBlocksPoolsCleaner(createMockArgMiniBlocksPoolsCleaner())

	key := []byte("mbKey")
	miniblock := &block.MiniBlockHeader{}
	miniblockCleaner.receivedMiniBlock(key, miniblock)
	assert.Nil(t, miniblockCleaner.mapMiniBlocksRounds[string(key)])
}

func TestReceivedMiniBlock_ShouldBeAddedInMap(t *testing.T) {
	t.Parallel()

	miniblockCleaner, _ := NewMiniBlocksPoolsCleaner(createMockArgMiniBlocksPoolsCleaner())

	key := []byte("mbKey")
	miniblock := &block.MiniBlock{}
	miniblockCleaner.receivedMiniBlock(key, miniblock)
	assert.NotNil(t, miniblockCleaner.mapMiniBlocksRounds[string(key)])
}

func TestCleanMiniblocksPoolsIfNeeded_MiniblockNotInPoolShouldBeRemovedFromMap(t *testing.T) {
	t.Parallel()

	args := createMockArgMiniBlocksPoolsCleaner()
	args.MiniblocksPool = &testscommon.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
	}
	miniblockCleaner, _ := NewMiniBlocksPoolsCleaner(args)

	key := []byte("mbKey")
	miniblock := &block.MiniBlock{}
	miniblockCleaner.receivedMiniBlock(key, miniblock)

	result := miniblockCleaner.cleanMiniblocksPoolsIfNeeded()
	assert.Equal(t, 0, result)
}

func TestCleanMiniblocksPoolsIfNeeded_RoundDiffTooSmallMiniblockShouldRemainInMap(t *testing.T) {
	t.Parallel()

	args := createMockArgMiniBlocksPoolsCleaner()
	args.MiniblocksPool = &testscommon.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, true
		},
	}
	miniblockCleaner, _ := NewMiniBlocksPoolsCleaner(args)

	key := []byte("mbKey")
	miniblock := &block.MiniBlock{}
	miniblockCleaner.receivedMiniBlock(key, miniblock)

	result := miniblockCleaner.cleanMiniblocksPoolsIfNeeded()
	assert.Equal(t, 1, result)
}

func TestCleanMiniblocksPoolsIfNeeded_MbShouldBeRemovedFromPoolAndMap(t *testing.T) {
	t.Parallel()

	args := createMockArgMiniBlocksPoolsCleaner()
	called := false
	args.MiniblocksPool = &testscommon.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, true
		},
		RemoveCalled: func(key []byte) {
			called = true
		},
	}
	roundHandler := &mock.RoundStub{
		IndexCalled: func() int64 {
			return 0
		},
	}
	args.RoundHandler = roundHandler
	miniblockCleaner, _ := NewMiniBlocksPoolsCleaner(args)

	key := []byte("mbKey")
	miniblock := &block.MiniBlock{}
	miniblockCleaner.receivedMiniBlock(key, miniblock)

	roundHandler.IndexCalled = func() int64 {
		return args.MaxRoundsToKeepUnprocessedData + 1
	}
	result := miniblockCleaner.cleanMiniblocksPoolsIfNeeded()
	assert.Equal(t, 0, result)
	assert.True(t, called)
}
