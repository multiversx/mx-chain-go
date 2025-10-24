package block

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	"github.com/stretchr/testify/assert"
)

func TestShardInfo_NewShardInfoCreateData(t *testing.T) {
	t.Parallel()

	t.Run("nil enableEpochsHandler", func(t *testing.T) {
		t.Parallel()

		headersPool := &pool.HeadersPoolStub{}
		proofsPool := &dataRetriever.ProofsPoolMock{}
		pendingMiniBlocksHandler := &mock.PendingMiniBlocksHandlerStub{}
		blockTracker := &mock.BlockTrackerMock{}

		sicd, err := NewShardInfoCreateData(
			nil,
			headersPool,
			proofsPool,
			pendingMiniBlocksHandler,
			blockTracker,
		)
		assert.Nil(t, sicd)
		assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
	})

	t.Run("nil headersPool", func(t *testing.T) {
		t.Parallel()

		proofsPool := &dataRetriever.ProofsPoolMock{}
		pendingMiniBlocksHandler := &mock.PendingMiniBlocksHandlerStub{}
		blockTracker := &mock.BlockTrackerMock{}
		enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()

		sicd, err := NewShardInfoCreateData(
			enableEpochsHandler,
			nil,
			proofsPool,
			pendingMiniBlocksHandler,
			blockTracker,
		)
		assert.Nil(t, sicd)
		assert.Equal(t, process.ErrNilHeadersDataPool, err)
	})

	t.Run("nil proofsPool", func(t *testing.T) {
		t.Parallel()

		headersPool := &pool.HeadersPoolStub{}
		pendingMiniBlocksHandler := &mock.PendingMiniBlocksHandlerStub{}
		blockTracker := &mock.BlockTrackerMock{}
		enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()

		sicd, err := NewShardInfoCreateData(
			enableEpochsHandler,
			headersPool,
			nil,
			pendingMiniBlocksHandler,
			blockTracker,
		)
		assert.Nil(t, sicd)
		assert.Equal(t, process.ErrNilProofsPool, err)
	})

	t.Run("nil pendingMiniBlocksHandler", func(t *testing.T) {
		t.Parallel()

		headersPool := &pool.HeadersPoolStub{}
		proofsPool := &dataRetriever.ProofsPoolMock{}
		blockTracker := &mock.BlockTrackerMock{}
		enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()

		sicd, err := NewShardInfoCreateData(
			enableEpochsHandler,
			headersPool,
			proofsPool,
			nil,
			blockTracker,
		)
		assert.Nil(t, sicd)
		assert.Equal(t, process.ErrNilPendingMiniBlocksHandler, err)
	})

	t.Run("nil blockTracker", func(t *testing.T) {
		t.Parallel()

		headersPool := &pool.HeadersPoolStub{}
		proofsPool := &dataRetriever.ProofsPoolMock{}
		pendingMiniBlocksHandler := &mock.PendingMiniBlocksHandlerStub{}
		enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()

		sicd, err := NewShardInfoCreateData(
			enableEpochsHandler,
			headersPool,
			proofsPool,
			pendingMiniBlocksHandler,
			nil,
		)
		assert.Nil(t, sicd)
		assert.Equal(t, process.ErrNilBlockTracker, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		headersPool := &pool.HeadersPoolStub{}
		proofsPool := &dataRetriever.ProofsPoolMock{}
		pendingMiniBlocksHandler := &mock.PendingMiniBlocksHandlerStub{}
		blockTracker := &mock.BlockTrackerMock{}
		enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()

		sicd, err := NewShardInfoCreateData(
			enableEpochsHandler,
			headersPool,
			proofsPool,
			pendingMiniBlocksHandler,
			blockTracker,
		)
		assert.NotNil(t, sicd)
		assert.Nil(t, err)
	})
}
