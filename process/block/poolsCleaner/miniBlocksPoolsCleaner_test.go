package poolsCleaner

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewMiniBlocksPoolsCleaner_NilMiniblockPoolShouldErr(t *testing.T) {
	t.Parallel()

	miniblockCleaner, err := NewMiniBlocksPoolsCleaner(nil, &mock.RounderMock{}, &mock.CoordinatorStub{})

	assert.Equal(t, process.ErrNilMiniBlockPool, err)
	assert.Nil(t, miniblockCleaner)
}

func TestNewMiniBlocksPoolsCleaner_NilRounderShouldErr(t *testing.T) {
	t.Parallel()

	miniblockCleaner, err := NewMiniBlocksPoolsCleaner(testscommon.NewCacherStub(), nil, &mock.CoordinatorStub{})

	assert.Equal(t, process.ErrNilRounder, err)
	assert.Nil(t, miniblockCleaner)
}

func TestNewMiniBlocksPoolsCleaner_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	miniblockCleaner, err := NewMiniBlocksPoolsCleaner(testscommon.NewCacherStub(), &mock.RoundStub{}, nil)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, miniblockCleaner)
}

func TestNewMiniBlocksPoolsCleaner_ShouldWork(t *testing.T) {
	t.Parallel()

	miniblockCleaner, err := NewMiniBlocksPoolsCleaner(testscommon.NewCacherStub(), &mock.RoundStub{}, &mock.CoordinatorStub{})

	assert.Nil(t, err)
	assert.NotNil(t, miniblockCleaner)
}

func TestReceivedMiniBlock_WrongTypeShouldBeIgnored(t *testing.T) {
	t.Parallel()

	miniblockCleaner, _ := NewMiniBlocksPoolsCleaner(testscommon.NewCacherStub(), &mock.RounderMock{}, &mock.CoordinatorStub{})

	key := []byte("mbKey")
	miniblock := &block.MiniBlockHeader{}
	miniblockCleaner.receivedMiniBlock(key, miniblock)
	assert.Nil(t, miniblockCleaner.mapMiniBlocksRounds[string(key)])
}

func TestReceivedMiniBlock_ShouldBeAddedInMap(t *testing.T) {
	t.Parallel()

	miniblockCleaner, _ := NewMiniBlocksPoolsCleaner(testscommon.NewCacherStub(), &mock.RounderMock{}, &mock.CoordinatorStub{})

	key := []byte("mbKey")
	miniblock := &block.MiniBlock{}
	miniblockCleaner.receivedMiniBlock(key, miniblock)
	assert.NotNil(t, miniblockCleaner.mapMiniBlocksRounds[string(key)])
}

func TestCleanMiniblocksPoolsIfNeeded_MiniblockNotInPoolShouldBeRemovedFromMap(t *testing.T) {
	t.Parallel()

	miniblockCleaner, _ := NewMiniBlocksPoolsCleaner(
		&testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		},
		&mock.RounderMock{},
		&mock.CoordinatorStub{},
	)

	key := []byte("mbKey")
	miniblock := &block.MiniBlock{}
	miniblockCleaner.receivedMiniBlock(key, miniblock)

	result := miniblockCleaner.cleanMiniblocksPoolsIfNeeded()
	assert.Equal(t, 0, result)
}

func TestCleanMiniblocksPoolsIfNeeded_RoundDiffTooSmallMiniblockShouldRemainInMap(t *testing.T) {
	t.Parallel()

	miniblockCleaner, _ := NewMiniBlocksPoolsCleaner(
		&testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, true
			},
		},
		&mock.RounderMock{},
		&mock.CoordinatorStub{},
	)

	key := []byte("mbKey")
	miniblock := &block.MiniBlock{}
	miniblockCleaner.receivedMiniBlock(key, miniblock)

	result := miniblockCleaner.cleanMiniblocksPoolsIfNeeded()
	assert.Equal(t, 1, result)
}

func TestCleanMiniblocksPoolsIfNeeded_MbShouldBeRemovedFromPoolAndMap(t *testing.T) {
	t.Parallel()

	called := false
	rounder := &mock.RoundStub{
		IndexCalled: func() int64 {
			return 0
		},
	}
	miniblockCleaner, _ := NewMiniBlocksPoolsCleaner(
		&testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, true
			},
			RemoveCalled: func(key []byte) {
				called = true
			},
		},
		rounder,
		&mock.CoordinatorStub{},
	)

	key := []byte("mbKey")
	miniblock := &block.MiniBlock{}
	miniblockCleaner.receivedMiniBlock(key, miniblock)

	rounder.IndexCalled = func() int64 {
		return process.MaxRoundsToKeepUnprocessedMiniBlocks + 1
	}
	result := miniblockCleaner.cleanMiniblocksPoolsIfNeeded()
	assert.Equal(t, 0, result)
	assert.True(t, called)
}
