package block

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardInfo_NewShardInfoCreateData(t *testing.T) {
	t.Parallel()

	t.Run("nil enableEpochsHandler", func(t *testing.T) {
		t.Parallel()

		args := createDefaultShardInfoCreateDataArgs()
		sicd, err := NewShardInfoCreateData(
			nil,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		assert.Nil(t, sicd)
		assert.True(t, sicd.IsInterfaceNil())
		assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
	})

	t.Run("nil headersPool", func(t *testing.T) {
		t.Parallel()
		args := createDefaultShardInfoCreateDataArgs()
		sicd, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			nil,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		assert.Nil(t, sicd)
		assert.True(t, sicd.IsInterfaceNil())
		assert.Equal(t, process.ErrNilHeadersDataPool, err)
	})

	t.Run("nil proofsPool", func(t *testing.T) {
		t.Parallel()

		args := createDefaultShardInfoCreateDataArgs()

		sicd, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			nil,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		assert.Nil(t, sicd)
		assert.True(t, sicd.IsInterfaceNil())
		assert.Equal(t, process.ErrNilProofsPool, err)
	})

	t.Run("nil pendingMiniBlocksHandler", func(t *testing.T) {
		t.Parallel()

		args := createDefaultShardInfoCreateDataArgs()

		sicd, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			nil,
			args.blockTracker,
		)
		assert.Nil(t, sicd)
		assert.True(t, sicd.IsInterfaceNil())
		assert.Equal(t, process.ErrNilPendingMiniBlocksHandler, err)
	})

	t.Run("nil blockTracker", func(t *testing.T) {
		t.Parallel()

		args := createDefaultShardInfoCreateDataArgs()
		sicd, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			nil,
		)
		assert.Nil(t, sicd)
		assert.True(t, sicd.IsInterfaceNil())
		assert.Equal(t, process.ErrNilBlockTracker, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createDefaultShardInfoCreateDataArgs()

		sicd, err := NewShardInfoCreateData(
			args.enableEpochsHandler,
			args.headersPool,
			args.proofsPool,
			args.pendingMiniBlocksHandler,
			args.blockTracker,
		)
		assert.NotNil(t, sicd)
		assert.False(t, sicd.IsInterfaceNil())
		assert.Nil(t, err)
	})
}

func TestShardInfoCreateData_miniBlockHeaderFromMiniBlockHeader(t *testing.T) {
	t.Parallel()

	headerHandler := getShardHeaderForShard1()
	t.Run("ScheduledMiniBlocksFlag disabled", func(t *testing.T) {
		enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()

		miniblockHeaders := createShardMiniBlockHeaderFromHeader(headerHandler, enableEpochsHandler)
		require.NotNil(t, miniblockHeaders)
		require.Equal(t, 3, len(miniblockHeaders))
	})
	t.Run("ScheduledMiniBlocksFlag enabled, all miniblocks final", func(t *testing.T) {
		enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
		enableEpochsHandler.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
			return flag == common.ScheduledMiniBlocksFlag
		}

		miniblockHeaders := createShardMiniBlockHeaderFromHeader(headerHandler, enableEpochsHandler)
		require.NotNil(t, miniblockHeaders)
		require.Equal(t, 3, len(miniblockHeaders))
	})
	t.Run("ScheduledMiniBlocksFlag enabled, not all miniblocks final", func(t *testing.T) {
		enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
		enableEpochsHandler.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
			return flag == common.ScheduledMiniBlocksFlag
		}
		_ = headerHandler.GetMiniBlockHeaderHandlers()[1].SetConstructionState(int32(block.Proposed))
		require.False(t, headerHandler.GetMiniBlockHeaderHandlers()[1].IsFinal())

		miniblockHeaders := createShardMiniBlockHeaderFromHeader(headerHandler, enableEpochsHandler)
		require.NotNil(t, miniblockHeaders)
		require.Equal(t, 2, len(miniblockHeaders))
	})
	t.Run("ScheduledMiniBlocksFlag enabled, no final miniblocks", func(t *testing.T) {
		enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
		enableEpochsHandler.IsFlagEnabledCalled = func(flag core.EnableEpochFlag) bool {
			return flag == common.ScheduledMiniBlocksFlag
		}
		for i := 0; i < len(headerHandler.GetMiniBlockHeaderHandlers()); i++ {
			fmt.Printf("MiniBlockHeader %d: %v\n", i, headerHandler.GetMiniBlockHeaderHandlers()[i])
			_ = headerHandler.GetMiniBlockHeaderHandlers()[i].SetConstructionState(int32(block.Proposed))
			require.False(t, headerHandler.GetMiniBlockHeaderHandlers()[i].IsFinal())
		}

		miniblockHeaders := createShardMiniBlockHeaderFromHeader(headerHandler, enableEpochsHandler)
		require.NotNil(t, miniblockHeaders)
		require.Equal(t, 0, len(miniblockHeaders))
	})
}

func TestShardInfoCreateData_createShardMiniBlockHeaderFromExecutionResultHandler(t *testing.T) {
	t.Parallel()

	execResultHandler := getExecutionResultForShard1()
	shardMiniBlockHeaders := createShardMiniBlockHeaderFromExecutionResultHandler(execResultHandler)
	require.NotNil(t, shardMiniBlockHeaders)
	require.Equal(t, 3, len(shardMiniBlockHeaders))
	for i := 0; i < len(shardMiniBlockHeaders); i++ {
		assert.Equal(t, execResultHandler.MiniBlockHeaders[i].Hash, shardMiniBlockHeaders[i].Hash)
		assert.Equal(t, execResultHandler.MiniBlockHeaders[i].Type, shardMiniBlockHeaders[i].Type)
		assert.Equal(t, execResultHandler.MiniBlockHeaders[i].TxCount, shardMiniBlockHeaders[i].TxCount)
		assert.Equal(t, execResultHandler.MiniBlockHeaders[i].SenderShardID, shardMiniBlockHeaders[i].SenderShardID)
		assert.Equal(t, execResultHandler.MiniBlockHeaders[i].ReceiverShardID, shardMiniBlockHeaders[i].ReceiverShardID)
	}
}

type shardInfoCreateDataTestArgs struct {
	headersPool              *pool.HeadersPoolStub
	proofsPool               *dataRetriever.ProofsPoolMock
	pendingMiniBlocksHandler *mock.PendingMiniBlocksHandlerStub
	blockTracker             *mock.BlockTrackerMock
	enableEpochsHandler      *enableEpochsHandlerMock.EnableEpochsHandlerStub
}

func createDefaultShardInfoCreateDataArgs() *shardInfoCreateDataTestArgs {
	return &shardInfoCreateDataTestArgs{
		headersPool:              &pool.HeadersPoolStub{},
		proofsPool:               &dataRetriever.ProofsPoolMock{},
		pendingMiniBlocksHandler: &mock.PendingMiniBlocksHandlerStub{},
		blockTracker:             &mock.BlockTrackerMock{},
		enableEpochsHandler:      enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
	}
}

func getMiniBlockHeadersForShard1() []block.MiniBlockHeader {
	mbHash1 := []byte("mb hash 1")
	mbHash2 := []byte("mb hash 2")
	mbHash3 := []byte("mb hash 3")

	miniBlockHeader1 := block.MiniBlockHeader{
		Hash:            mbHash1,
		Type:            block.TxBlock,
		TxCount:         10,
		SenderShardID:   1,
		ReceiverShardID: 2,
	}
	miniBlockHeader2 := block.MiniBlockHeader{
		Hash:            mbHash2,
		Type:            block.InvalidBlock,
		TxCount:         1,
		SenderShardID:   1,
		ReceiverShardID: 2,
	}
	miniBlockHeader3 := block.MiniBlockHeader{
		Hash:            mbHash3,
		Type:            block.SmartContractResultBlock,
		TxCount:         5,
		SenderShardID:   1,
		ReceiverShardID: 0,
	}

	miniBlockHeaders := make([]block.MiniBlockHeader, 0)
	miniBlockHeaders = append(miniBlockHeaders, miniBlockHeader1)
	miniBlockHeaders = append(miniBlockHeaders, miniBlockHeader2)
	miniBlockHeaders = append(miniBlockHeaders, miniBlockHeader3)
	return miniBlockHeaders
}

func getShardHeaderForShard1() data.HeaderHandler {
	prevHash := []byte("prevHash")
	prevRandSeed := []byte("prevRandSeed")
	currRandSeed := []byte("currRandSeed")
	return &block.HeaderV2{
		Header: &block.Header{
			Round:            10,
			Nonce:            45,
			ShardID:          0,
			PrevRandSeed:     prevRandSeed,
			RandSeed:         currRandSeed,
			PrevHash:         prevHash,
			MiniBlockHeaders: getMiniBlockHeadersForShard1(),
		},
	}
}

func getExecutionResultForShard1() *block.ExecutionResult {

	return &block.ExecutionResult{
		ExecutedTxCount:  100,
		MiniBlockHeaders: getMiniBlockHeadersForShard1(),
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("header hash"),
			HeaderNonce: 12345,
			HeaderEpoch: 7,
			HeaderRound: 15,
			RootHash:    []byte("root hash"),
			GasUsed:     5000,
		},
	}
}
