package coordinator

import (
	"bytes"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process/factory/containers"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/preprocMocks"
)

func TestBlockDataRequest_getFinalCrossMiniBlockInfos(t *testing.T) {
	t.Parallel()

	hash1, hash2 := "hash1", "hash2"

	t.Run("scheduledMiniBlocks flag not set", func(t *testing.T) {
		t.Parallel()
		blockDataRequesterArgs := createMockArgs()
		blockDataRequester, _ := NewBlockDataRequester(blockDataRequesterArgs)

		var crossMiniBlockInfos []*data.MiniBlockInfo

		mbInfos := blockDataRequester.getFinalCrossMiniBlockInfos(crossMiniBlockInfos, &block.Header{})
		require.Equal(t, crossMiniBlockInfos, mbInfos)
	})

	t.Run("should work, miniblocks info found for final miniBlock header", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		enableEpochsHandlerStub := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
		args.EnableEpochsHandler = enableEpochsHandlerStub
		blockDataRequester, _ := NewBlockDataRequester(args)
		enableEpochsHandlerStub.AddActiveFlags(common.ScheduledMiniBlocksFlag)

		mbInfo1 := &data.MiniBlockInfo{Hash: []byte(hash1)}
		mbInfo2 := &data.MiniBlockInfo{Hash: []byte(hash2)}
		crossMiniBlockInfos := []*data.MiniBlockInfo{mbInfo1, mbInfo2}

		mbh1 := block.MiniBlockHeader{Hash: []byte(hash1)}
		mbhReserved1 := block.MiniBlockHeaderReserved{State: block.Proposed}
		mbh1.Reserved, _ = mbhReserved1.Marshal()

		mbh2 := block.MiniBlockHeader{Hash: []byte(hash2)}
		mbhReserved2 := block.MiniBlockHeaderReserved{State: block.Final}
		mbh2.Reserved, _ = mbhReserved2.Marshal()

		header := &block.MetaBlock{
			MiniBlockHeaders: []block.MiniBlockHeader{
				mbh1,
				mbh2,
			},
		}

		expectedMbInfos := []*data.MiniBlockInfo{mbInfo2}

		mbInfos := blockDataRequester.getFinalCrossMiniBlockInfos(crossMiniBlockInfos, header)
		require.Equal(t, expectedMbInfos, mbInfos)
	})
}

func TestTransactionCoordinator_requestMissingMiniBlocksAndTransactionsShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.MiniBlockPool = &cache.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(key, []byte("hash0")) || bytes.Equal(key, []byte("hash1")) || bytes.Equal(key, []byte("hash2")) {
				if bytes.Equal(key, []byte("hash0")) {
					return nil, true
				}

				if bytes.Equal(key, []byte("hash1")) {
					return &block.MiniBlock{
						Type: block.PeerBlock,
						TxHashes: [][]byte{
							[]byte("hash 1"),
							[]byte("hash 2"),
						},
					}, true
				}

				if bytes.Equal(key, []byte("hash2")) {
					return &block.MiniBlock{
						Type: block.TxBlock,
						TxHashes: [][]byte{
							[]byte("hash 3"),
							[]byte("hash 4"),
						},
					}, true
				}
			}
			return nil, false
		},
	}

	numTxsRequested := 0
	args.PreProcessors = containers.NewPreProcessorsContainer()
	err := args.PreProcessors.Add(block.TxBlock, &preprocMocks.PreProcessorMock{
		RequestTransactionsForMiniBlockCalled: func(miniBlock *block.MiniBlock) int {
			numTxsRequested += len(miniBlock.TxHashes)
			return len(miniBlock.TxHashes)
		},
	})
	require.Nil(t, err)

	wg := sync.WaitGroup{}
	wg.Add(3)
	mapRequestedMiniBlocksPerShard := make(map[uint32]int)
	mutMap := sync.RWMutex{}
	args.RequestHandler = &testscommon.RequestHandlerStub{
		RequestMiniBlocksHandlerCalled: func(destShardID uint32, miniblocksHashes [][]byte) {
			mutMap.Lock()
			mapRequestedMiniBlocksPerShard[destShardID] += len(miniblocksHashes)
			mutMap.Unlock()
			wg.Done()
		},
	}

	blockDataRequester, _ := NewBlockDataRequester(args)

	mbsInfo := []*data.MiniBlockInfo{
		{SenderShardID: 0},
		{SenderShardID: 1},
		{SenderShardID: 2},
		{SenderShardID: 0, Hash: []byte("hash0")},
		{SenderShardID: 1, Hash: []byte("hash1")},
		{SenderShardID: 2, Hash: []byte("hash2")},
		{SenderShardID: 0},
		{SenderShardID: 1},
		{SenderShardID: 0},
	}

	blockDataRequester.requestMissingMiniBlocksAndTransactions(mbsInfo)

	wg.Wait()

	mutMap.RLock()
	require.Equal(t, 3, mapRequestedMiniBlocksPerShard[0])
	require.Equal(t, 2, mapRequestedMiniBlocksPerShard[1])
	require.Equal(t, 1, mapRequestedMiniBlocksPerShard[2])
	require.Equal(t, 2, numTxsRequested)
	mutMap.RUnlock()
}

func createMockArgs() BlockDataRequestArgs {
	return BlockDataRequestArgs{
		RequestHandler:      &testscommon.RequestHandlerStub{},
		MiniBlockPool:       dataRetrieverMock.NewPoolsHolderMock().MiniBlocks(),
		PreProcessors:       &preprocMocks.PreProcessorContainerMock{},
		ShardCoordinator:    mock.NewMultiShardsCoordinatorMock(5),
		EnableEpochsHandler: enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
	}
}
