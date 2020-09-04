package preprocess_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

const maxSizeInBytes = uint32(core.MegabyteSize * 90 / 100)

func TestNewBlockSizeComputation_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	bsc, err := preprocess.NewBlockSizeComputation(nil, &mock.BlockSizeThrottlerStub{}, maxSizeInBytes)

	assert.True(t, check.IfNil(bsc))
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewBlockSizeComputation_NilBlockSizeThrottlerShouldErr(t *testing.T) {
	t.Parallel()

	bsc, err := preprocess.NewBlockSizeComputation(&mock.ProtobufMarshalizerMock{}, nil, maxSizeInBytes)

	assert.True(t, check.IfNil(bsc))
	assert.Equal(t, process.ErrNilBlockSizeThrottler, err)
}

func TestNewBlockSizeComputation_WithMockMarshalizerShouldWorkAndComputeValues(t *testing.T) {
	t.Parallel()

	bsc, err := preprocess.NewBlockSizeComputation(&mock.ProtobufMarshalizerMock{}, &mock.BlockSizeThrottlerStub{}, maxSizeInBytes)

	assert.False(t, check.IfNil(bsc))
	assert.Nil(t, err)
	assert.Equal(t, uint32(9), bsc.MiniblockSize())
	assert.Equal(t, uint32(34), bsc.TxSize())
}

func TestNewBlockSizeComputation_MarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	numComputations := 4
	for i := 0; i < numComputations; i++ {
		testMarshalizerFailsShouldErr(t, i)
	}
}

func testMarshalizerFailsShouldErr(t *testing.T, idxCallMarshalFail int) {
	cnt := 0
	expectedErr := errors.New("expected error")
	bsc, err := preprocess.NewBlockSizeComputation(
		&mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) (bytes []byte, err error) {
				if cnt == idxCallMarshalFail {
					return nil, expectedErr
				}
				cnt++
				return []byte("dummy"), nil
			},
		},
		&mock.BlockSizeThrottlerStub{},
		maxSizeInBytes,
	)

	assert.True(t, check.IfNil(bsc))
	assert.Equal(t, expectedErr, err)
}

func TestBlockSizeComputation_AddNumMiniBlocks(t *testing.T) {
	t.Parallel()

	bsc, _ := preprocess.NewBlockSizeComputation(&mock.ProtobufMarshalizerMock{}, &mock.BlockSizeThrottlerStub{}, maxSizeInBytes)

	val := 56
	bsc.AddNumMiniBlocks(val)

	assert.Equal(t, uint32(val), bsc.NumMiniBlocks())
}

func TestBlockSizeComputation_AddNumTxs(t *testing.T) {
	t.Parallel()

	bsc, _ := preprocess.NewBlockSizeComputation(&mock.ProtobufMarshalizerMock{}, &mock.BlockSizeThrottlerStub{}, maxSizeInBytes)

	val := 57
	bsc.AddNumTxs(val)

	assert.Equal(t, uint32(val), bsc.NumTxs())
}

func TestBlockSizeComputation_Init(t *testing.T) {
	t.Parallel()

	bsc, _ := preprocess.NewBlockSizeComputation(&mock.ProtobufMarshalizerMock{}, &mock.BlockSizeThrottlerStub{}, maxSizeInBytes)

	numTxs := 57
	numMiniblocks := 23
	bsc.AddNumMiniBlocks(numMiniblocks)
	bsc.AddNumTxs(numTxs)

	bsc.Init()

	assert.Equal(t, uint32(0), bsc.NumTxs())
	assert.Equal(t, uint32(0), bsc.NumMiniBlocks())
}

func TestBlockSizeComputation_IsMaxBlockSizeReachedShouldWork(t *testing.T) {
	t.Parallel()

	bsc, _ := preprocess.NewBlockSizeComputation(
		&mock.ProtobufMarshalizerMock{},
		&mock.BlockSizeThrottlerStub{
			GetCurrentMaxSizeCalled: func() uint32 {
				return maxSizeInBytes
			},
		},
		maxSizeInBytes,
	)

	testData := []struct {
		numNewMiniBlocks int
		numNewTxs        int
		expected         bool
		name             string
	}{
		{numNewMiniBlocks: 0, numNewTxs: 0, expected: false, name: "with miniblocks 0 and txs 0"},
		{numNewMiniBlocks: 1000000, numNewTxs: 0, expected: true, name: "with miniblocks 1000000 and txs 0"},
		{numNewMiniBlocks: 0, numNewTxs: 1000000, expected: true, name: "with miniblocks 0 and txs 1000000"},
		{numNewMiniBlocks: 15, numNewTxs: 1000, expected: false, name: "with miniblocks 15 and txs 1000"},
		{numNewMiniBlocks: 1, numNewTxs: 27756, expected: false, name: "with miniblocks 1 and txs 27756"},
		{numNewMiniBlocks: 1, numNewTxs: 27757, expected: true, name: "with miniblocks 1 and txs 27757"},
		{numNewMiniBlocks: 2, numNewTxs: 27756, expected: true, name: "with miniblocks 2 and txs 27756"},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			assert.Equal(t, td.expected, bsc.IsMaxBlockSizeReached(td.numNewMiniBlocks, td.numNewTxs))
		})
	}
}

func TestBlockSizeComputation_IsMaxBlockSizeWithoutThrottleReachedShouldWork(t *testing.T) {
	t.Parallel()

	bsc, _ := preprocess.NewBlockSizeComputation(
		&mock.ProtobufMarshalizerMock{},
		&mock.BlockSizeThrottlerStub{
			GetCurrentMaxSizeCalled: func() uint32 {
				return 0
			},
		},
		maxSizeInBytes,
	)

	testData := []struct {
		numNewMiniBlocks int
		numNewTxs        int
		expected         bool
		name             string
	}{
		{numNewMiniBlocks: 0, numNewTxs: 0, expected: false, name: "with miniblocks 0 and txs 0"},
		{numNewMiniBlocks: 1000000, numNewTxs: 0, expected: true, name: "with miniblocks 1000000 and txs 0"},
		{numNewMiniBlocks: 0, numNewTxs: 1000000, expected: true, name: "with miniblocks 0 and txs 1000000"},
		{numNewMiniBlocks: 15, numNewTxs: 1000, expected: false, name: "with miniblocks 15 and txs 1000"},
		{numNewMiniBlocks: 1, numNewTxs: 27756, expected: false, name: "with miniblocks 1 and txs 27756"},
		{numNewMiniBlocks: 1, numNewTxs: 27757, expected: true, name: "with miniblocks 1 and txs 27757"},
		{numNewMiniBlocks: 2, numNewTxs: 27756, expected: true, name: "with miniblocks 2 and txs 27756"},
	}

	for _, td := range testData {
		t.Run(td.name, func(t *testing.T) {
			assert.Equal(t, td.expected, bsc.IsMaxBlockSizeWithoutThrottleReached(td.numNewMiniBlocks, td.numNewTxs))
		})
	}
}

func TestBlockSizeComputation_MaxTransactionsInOneMiniblock(t *testing.T) {
	t.Parallel()

	bsc, _ := preprocess.NewBlockSizeComputation(&mock.ProtobufMarshalizerMock{}, &mock.BlockSizeThrottlerStub{}, maxSizeInBytes)

	maxTxs := bsc.MaxTransactionsInOneMiniblock()

	assert.Equal(t, 27756, maxTxs)
}
