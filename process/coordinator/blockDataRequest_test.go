package coordinator

import (
	"bytes"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
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

func TestNewBlockDataRequester(t *testing.T) {
	t.Parallel()

	t.Run("nil request handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.RequestHandler = nil

		blockDataRequester, err := NewBlockDataRequester(args)
		require.Nil(t, blockDataRequester)
		require.Equal(t, process.ErrNilRequestHandler, err)
	})

	t.Run("nil mini block pool should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.MiniBlockPool = nil

		blockDataRequester, err := NewBlockDataRequester(args)
		require.Nil(t, blockDataRequester)
		require.Equal(t, process.ErrNilMiniBlockPool, err)
	})

	t.Run("nil pre processors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.PreProcessors = nil

		blockDataRequester, err := NewBlockDataRequester(args)
		require.Nil(t, blockDataRequester)
		require.Equal(t, process.ErrNilPreProcessorsContainer, err)
	})

	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.ShardCoordinator = nil

		blockDataRequester, err := NewBlockDataRequester(args)
		require.Nil(t, blockDataRequester)
		require.Equal(t, process.ErrNilShardCoordinator, err)
	})

	t.Run("nil enable epochs handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.EnableEpochsHandler = nil

		blockDataRequester, err := NewBlockDataRequester(args)
		require.Nil(t, blockDataRequester)
		require.Equal(t, process.ErrNilEnableEpochsHandler, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()

		blockDataRequester, err := NewBlockDataRequester(args)
		require.False(t, blockDataRequester.IsInterfaceNil())
		require.Nil(t, err)
		require.NotNil(t, blockDataRequester.requestedTxs)
		require.NotNil(t, blockDataRequester.preProcessors)
		require.NotNil(t, blockDataRequester.miniBlockPool)
		require.NotNil(t, blockDataRequester.shardCoordinator)
		require.NotNil(t, blockDataRequester.enableEpochsHandler)
		require.NotNil(t, blockDataRequester.requestHandler)
		require.NotNil(t, blockDataRequester.requestedItemsHandler)
	})
}

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

func TestBlockDataRequest_RequestBlockTransactions(t *testing.T) {
	t.Parallel()

	t.Run("nil body should return early", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()
		blockDataRequester, _ := NewBlockDataRequester(args)

		// Should not panic
		blockDataRequester.RequestBlockTransactions(nil)
	})

	t.Run("empty body should work", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()
		blockDataRequester, _ := NewBlockDataRequester(args)

		emptyBody := &block.Body{}
		blockDataRequester.RequestBlockTransactions(emptyBody)

		// Should initialize requestedTxs map
		blockDataRequester.mutRequestedTxs.RLock()
		require.Equal(t, 0, len(blockDataRequester.requestedTxs))
		blockDataRequester.mutRequestedTxs.RUnlock()
	})

	t.Run("should request transactions for different block types", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()

		// Create preprocessors container with mocks
		preprocContainer := containers.NewPreProcessorsContainer()
		txPreproc := &preprocMocks.PreProcessorMock{
			RequestBlockTransactionsCalled: func(body *block.Body) int {
				totalTxs := 0
				for _, mb := range body.MiniBlocks {
					if mb.Type == block.TxBlock {
						totalTxs += len(mb.TxHashes)
					}
				}
				return totalTxs
			},
		}
		peerPreproc := &preprocMocks.PreProcessorMock{
			RequestBlockTransactionsCalled: func(body *block.Body) int {
				totalTxs := 0
				for _, mb := range body.MiniBlocks {
					if mb.Type == block.PeerBlock {
						totalTxs += len(mb.TxHashes)
					}
				}
				return totalTxs
			},
		}

		_ = preprocContainer.Add(block.TxBlock, txPreproc)
		_ = preprocContainer.Add(block.PeerBlock, peerPreproc)
		args.PreProcessors = preprocContainer

		blockDataRequester, _ := NewBlockDataRequester(args)

		// Create body with different block types
		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					Type:     block.TxBlock,
					TxHashes: [][]byte{[]byte("tx1"), []byte("tx2")},
				},
				{
					Type:     block.PeerBlock,
					TxHashes: [][]byte{[]byte("peer1"), []byte("peer2"), []byte("peer3")},
				},
			},
		}

		blockDataRequester.RequestBlockTransactions(body)

		// Check that requestedTxs was updated
		blockDataRequester.mutRequestedTxs.RLock()
		require.Equal(t, 2, len(blockDataRequester.requestedTxs))
		require.Equal(t, 2, blockDataRequester.requestedTxs[block.TxBlock])
		require.Equal(t, 3, blockDataRequester.requestedTxs[block.PeerBlock])
		blockDataRequester.mutRequestedTxs.RUnlock()
	})

	t.Run("should handle preprocessor not found gracefully", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()
		blockDataRequester, _ := NewBlockDataRequester(args)
		blockDataRequester.preProcessors = &preprocMocks.PreProcessorContainerMock{
			GetCalled: func(key block.Type) (process.PreProcessor, error) {
				return nil, errors.New("not found")
			},
		}

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					Type:     block.TxBlock,
					TxHashes: [][]byte{[]byte("tx1")},
				},
			},
		}

		// Should not panic even if preprocessor is not found
		blockDataRequester.RequestBlockTransactions(body)

		// Should still initialize the map
		blockDataRequester.mutRequestedTxs.RLock()
		require.Equal(t, 0, len(blockDataRequester.requestedTxs))
		blockDataRequester.mutRequestedTxs.RUnlock()
	})

	t.Run("should handle concurrent requests safely", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()
		blockDataRequester, _ := NewBlockDataRequester(args)

		// Create a preprocessor that simulates some processing time
		preprocContainer := containers.NewPreProcessorsContainer()
		txPreproc := &preprocMocks.PreProcessorMock{
			RequestBlockTransactionsCalled: func(body *block.Body) int {
				time.Sleep(10 * time.Millisecond) // Simulate processing time
				if body == nil || len(body.MiniBlocks) == 0 || body.MiniBlocks[0] == nil {
					return 0
				}
				return len(body.MiniBlocks[0].TxHashes)
			},
		}
		_ = preprocContainer.Add(block.TxBlock, txPreproc)
		args.PreProcessors = preprocContainer

		blockDataRequester, _ = NewBlockDataRequester(args)

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					Type:     block.TxBlock,
					TxHashes: [][]byte{[]byte("tx1")},
				},
			},
		}

		// Make concurrent calls
		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				blockDataRequester.RequestBlockTransactions(body)
			}()
		}

		wg.Wait()

		// Should not have race conditions
		blockDataRequester.mutRequestedTxs.RLock()
		require.Equal(t, 1, len(blockDataRequester.requestedTxs))
		blockDataRequester.mutRequestedTxs.RUnlock()
	})
}

func TestBlockDataRequest_RequestMiniBlocksAndTransactions(t *testing.T) {
	t.Parallel()

	t.Run("nil header should return early", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()
		blockDataRequester, _ := NewBlockDataRequester(args)

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("expected no panic, but got: %v", r)
			}
		}()

		blockDataRequester.RequestMiniBlocksAndTransactions(nil)
	})

	t.Run("should work with scheduled mini blocks flag disabled", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()
		enableEpochsHandlerStub := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
		args.EnableEpochsHandler = enableEpochsHandlerStub

		calledMiniblock := 0
		wg := &sync.WaitGroup{}
		wg.Add(1)
		args.RequestHandler = &testscommon.RequestHandlerStub{
			RequestMiniBlocksHandlerCalled: func(destShardID uint32, miniblocksHashes [][]byte) {
				calledMiniblock++
				wg.Done()
			},
		}
		blockDataRequester, _ := NewBlockDataRequester(args)

		// Mock the mini block pool to return some mini blocks
		miniBlockPool := &cache.CacherStub{
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal(key, []byte("hash1")) {
					return &block.MiniBlock{
						Type:     block.TxBlock,
						TxHashes: [][]byte{[]byte("tx1"), []byte("tx2")},
					}, true
				}
				return nil, false
			},
		}
		blockDataRequester.miniBlockPool = miniBlockPool

		// Mock preprocessors
		calledCount := 0
		preprocContainer := containers.NewPreProcessorsContainer()
		txPreproc := &preprocMocks.PreProcessorMock{
			RequestTransactionsForMiniBlockCalled: func(miniBlock *block.MiniBlock) int {
				calledCount++
				return len(miniBlock.TxHashes)
			},
		}
		_ = preprocContainer.Add(block.TxBlock, txPreproc)
		blockDataRequester.preProcessors = preprocContainer

		// Mock the shard coordinator to return self ID
		shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
		shardCoordinator.CurrentShard = 1
		blockDataRequester.shardCoordinator = shardCoordinator

		// Mock the header to return cross mini blocks
		headerWithCrossMbs := &testscommon.HeaderHandlerStub{
			RoundField: 100,
			GetMiniBlockHeadersWithDstCalled: func(destShardID uint32) map[string]uint32 {
				return map[string]uint32{
					"hash1": 0, // from shard 0 to shard 1
					"hash2": 2, // from shard 2 to shard 1
				}
			},
		}

		blockDataRequester.RequestMiniBlocksAndTransactions(headerWithCrossMbs)

		wg.Wait()

		require.Equal(t, 1, calledCount)
		require.Equal(t, 1, calledMiniblock)
	})

	t.Run("should work with scheduled mini blocks flag enabled", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()
		enableEpochsHandlerStub := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
		enableEpochsHandlerStub.AddActiveFlags(common.ScheduledMiniBlocksFlag)
		args.EnableEpochsHandler = enableEpochsHandlerStub

		blockDataRequester, _ := NewBlockDataRequester(args)

		// Mock the mini block pool to return mini blocks not found
		miniBlockPool := &cache.CacherStub{
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false // Mini blocks not found, so they will be requested
			},
		}
		blockDataRequester.miniBlockPool = miniBlockPool

		// Mock the request handler to track requests
		var requestedMiniBlocksCount int
		mutRequested := sync.RWMutex{}

		wg := &sync.WaitGroup{}
		wg.Add(2)
		requestHandler := &testscommon.RequestHandlerStub{
			RequestMiniBlocksHandlerCalled: func(destShardID uint32, miniblocksHashes [][]byte) {
				mutRequested.Lock()
				requestedMiniBlocksCount += len(miniblocksHashes)
				mutRequested.Unlock()
				wg.Done()
			},
		}
		blockDataRequester.requestHandler = requestHandler

		// Mock the shard coordinator
		shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
		shardCoordinator.CurrentShard = 1
		blockDataRequester.shardCoordinator = shardCoordinator

		// Create header with cross mini blocks
		header := &testscommon.HeaderHandlerStub{
			RoundField: 100,
			GetMiniBlockHeadersWithDstCalled: func(destShardID uint32) map[string]uint32 {
				return map[string]uint32{
					"hash1": 0, // from shard 0 to shard 1
					"hash2": 2, // from shard 2 to shard 1
				}
			},
			GetMiniBlockHeaderHandlersCalled: func() []data.MiniBlockHeaderHandler {
				// Create final mini block headers
				mbh1 := &block.MiniBlockHeader{Hash: []byte("hash1")}
				mbhReserved1 := block.MiniBlockHeaderReserved{State: block.Final}
				mbh1.Reserved, _ = mbhReserved1.Marshal()

				mbh2 := &block.MiniBlockHeader{Hash: []byte("hash2")}
				mbhReserved2 := block.MiniBlockHeaderReserved{State: block.Final}
				mbh2.Reserved, _ = mbhReserved2.Marshal()

				return []data.MiniBlockHeaderHandler{mbh1, mbh2}
			},
		}

		blockDataRequester.RequestMiniBlocksAndTransactions(header)

		wg.Wait()

		// Verify that mini blocks were requested
		mutRequested.RLock()
		require.Equal(t, 2, requestedMiniBlocksCount)
		mutRequested.RUnlock()
	})

	t.Run("should handle empty cross mini blocks", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()
		blockDataRequester, _ := NewBlockDataRequester(args)

		// Mock the shard coordinator
		shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
		shardCoordinator.CurrentShard = 1
		blockDataRequester.shardCoordinator = shardCoordinator

		// Create header with no cross mini blocks
		header := &testscommon.HeaderHandlerStub{
			RoundField: 100,
			GetMiniBlockHeadersWithDstCalled: func(destShardID uint32) map[string]uint32 {
				return map[string]uint32{} // Empty map
			},
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("expected no panic, but got: %v", r)
			}
		}()

		blockDataRequester.RequestMiniBlocksAndTransactions(header)
	})
}

func TestBlockDataRequest_IsDataPreparedForProcessing(t *testing.T) {
	t.Parallel()

	t.Run("no requested transactions should return nil", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()
		blockDataRequester, _ := NewBlockDataRequester(args)

		// No transactions requested yet
		blockDataRequester.mutRequestedTxs.Lock()
		blockDataRequester.requestedTxs = make(map[block.Type]int)
		blockDataRequester.mutRequestedTxs.Unlock()

		haveTime := func() time.Duration { return time.Second }
		err := blockDataRequester.IsDataPreparedForProcessing(haveTime)
		require.Nil(t, err)
	})

	t.Run("all preprocessors succeed should return nil", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()

		// Create preprocessors container with mocks that always succeed
		preprocContainer := containers.NewPreProcessorsContainer()
		txPreproc := &preprocMocks.PreProcessorMock{
			IsDataPreparedCalled: func(requestedTxs int, haveTime func() time.Duration) error {
				return nil
			},
		}
		peerPreproc := &preprocMocks.PreProcessorMock{
			IsDataPreparedCalled: func(requestedTxs int, haveTime func() time.Duration) error {
				return nil
			},
		}

		_ = preprocContainer.Add(block.TxBlock, txPreproc)
		_ = preprocContainer.Add(block.PeerBlock, peerPreproc)
		args.PreProcessors = preprocContainer

		blockDataRequester, _ := NewBlockDataRequester(args)

		// Set some requested transactions
		blockDataRequester.mutRequestedTxs.Lock()
		blockDataRequester.requestedTxs[block.TxBlock] = 5
		blockDataRequester.requestedTxs[block.PeerBlock] = 3
		blockDataRequester.mutRequestedTxs.Unlock()

		haveTime := func() time.Duration { return time.Second }
		err := blockDataRequester.IsDataPreparedForProcessing(haveTime)
		require.Nil(t, err)
	})

	t.Run("one preprocessor fails should return error", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()

		expectedErr := errors.New("data not prepared")
		preprocContainer := containers.NewPreProcessorsContainer()
		txPreproc := &preprocMocks.PreProcessorMock{
			IsDataPreparedCalled: func(requestedTxs int, haveTime func() time.Duration) error {
				return expectedErr
			},
		}
		peerPreproc := &preprocMocks.PreProcessorMock{
			IsDataPreparedCalled: func(requestedTxs int, haveTime func() time.Duration) error {
				return nil
			},
		}

		_ = preprocContainer.Add(block.TxBlock, txPreproc)
		_ = preprocContainer.Add(block.PeerBlock, peerPreproc)
		args.PreProcessors = preprocContainer

		blockDataRequester, _ := NewBlockDataRequester(args)

		// Set some requested transactions
		blockDataRequester.mutRequestedTxs.Lock()
		blockDataRequester.requestedTxs[block.TxBlock] = 5
		blockDataRequester.requestedTxs[block.PeerBlock] = 3
		blockDataRequester.mutRequestedTxs.Unlock()

		haveTime := func() time.Duration { return time.Second }
		err := blockDataRequester.IsDataPreparedForProcessing(haveTime)
		require.Equal(t, expectedErr, err)
	})

	t.Run("preprocessor not found should continue gracefully", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()

		// Create preprocessors container with only one preprocessor
		preprocContainer := containers.NewPreProcessorsContainer()
		txPreproc := &preprocMocks.PreProcessorMock{
			IsDataPreparedCalled: func(requestedTxs int, haveTime func() time.Duration) error {
				return nil
			},
		}
		_ = preprocContainer.Add(block.TxBlock, txPreproc)
		args.PreProcessors = preprocContainer

		blockDataRequester, _ := NewBlockDataRequester(args)

		// Set requested transactions for both types, but only one preprocessor exists
		blockDataRequester.mutRequestedTxs.Lock()
		blockDataRequester.requestedTxs[block.TxBlock] = 5
		blockDataRequester.requestedTxs[block.PeerBlock] = 3
		blockDataRequester.mutRequestedTxs.Unlock()

		haveTime := func() time.Duration { return time.Second }
		err := blockDataRequester.IsDataPreparedForProcessing(haveTime)
		require.Nil(t, err) // Should not panic and should return nil
	})

	t.Run("concurrent processing should work correctly", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()

		// Create preprocessors that simulate some processing time
		preprocContainer := containers.NewPreProcessorsContainer()
		txPreproc := &preprocMocks.PreProcessorMock{
			IsDataPreparedCalled: func(requestedTxs int, haveTime func() time.Duration) error {
				time.Sleep(10 * time.Millisecond) // Simulate processing time
				return nil
			},
		}
		peerPreproc := &preprocMocks.PreProcessorMock{
			IsDataPreparedCalled: func(requestedTxs int, haveTime func() time.Duration) error {
				time.Sleep(15 * time.Millisecond) // Different processing time
				return nil
			},
		}

		_ = preprocContainer.Add(block.TxBlock, txPreproc)
		_ = preprocContainer.Add(block.PeerBlock, peerPreproc)
		args.PreProcessors = preprocContainer

		blockDataRequester, _ := NewBlockDataRequester(args)

		// Set multiple requested transactions
		blockDataRequester.mutRequestedTxs.Lock()
		blockDataRequester.requestedTxs[block.TxBlock] = 5
		blockDataRequester.requestedTxs[block.PeerBlock] = 3
		blockDataRequester.mutRequestedTxs.Unlock()

		haveTime := func() time.Duration { return time.Second }
		err := blockDataRequester.IsDataPreparedForProcessing(haveTime)
		require.Nil(t, err)
	})

	t.Run("haveTime function should be called by preprocessors", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()

		timeFunctionCalled := false
		haveTime := func() time.Duration {
			timeFunctionCalled = true
			return time.Second
		}

		preprocContainer := containers.NewPreProcessorsContainer()
		txPreproc := &preprocMocks.PreProcessorMock{
			IsDataPreparedCalled: func(requestedTxs int, haveTime func() time.Duration) error {
				haveTime() // Call the function to verify it's passed correctly
				return nil
			},
		}

		_ = preprocContainer.Add(block.TxBlock, txPreproc)
		args.PreProcessors = preprocContainer

		blockDataRequester, _ := NewBlockDataRequester(args)

		// Set requested transactions
		blockDataRequester.mutRequestedTxs.Lock()
		blockDataRequester.requestedTxs[block.TxBlock] = 5
		blockDataRequester.mutRequestedTxs.Unlock()

		err := blockDataRequester.IsDataPreparedForProcessing(haveTime)
		require.Nil(t, err)
		require.True(t, timeFunctionCalled)
	})
}

func TestBlockDataRequest_receivedMiniBlock(t *testing.T) {
	t.Parallel()

	t.Run("nil key should return early", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()
		blockDataRequester, _ := NewBlockDataRequester(args)

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("expected no panic, but got: %v", r)
			}
		}()
		blockDataRequester.receivedMiniBlock(nil, &block.MiniBlock{})
	})

	t.Run("unrequested mini block should return early", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()
		blockDataRequester, _ := NewBlockDataRequester(args)

		// Mini block was not requested
		key := []byte("hash1")
		miniBlock := &block.MiniBlock{
			Type:     block.TxBlock,
			TxHashes: [][]byte{[]byte("tx1"), []byte("tx2")},
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("expected no panic, but got: %v", r)
			}
		}()
		blockDataRequester.receivedMiniBlock(key, miniBlock)
	})

	t.Run("wrong type assertion should log warning and return", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()
		blockDataRequester, _ := NewBlockDataRequester(args)

		// Add the mini block to requested items
		key := []byte("hash1")
		_ = blockDataRequester.requestedItemsHandler.Add(string(key))

		// Pass wrong type (not a mini block)
		wrongValue := "not a mini block"

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("expected no panic, but got: %v", r)
			}
		}()
		blockDataRequester.receivedMiniBlock(key, wrongValue)
	})

	t.Run("preprocessor not found should log warning and return", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()
		blockDataRequester, _ := NewBlockDataRequester(args)

		// Add the mini block to requested items
		key := []byte("hash1")
		_ = blockDataRequester.requestedItemsHandler.Add(string(key))

		// Create mini block with type that has no preprocessor
		miniBlock := &block.MiniBlock{
			Type:     block.InvalidBlock, // This type won't have a preprocessor
			TxHashes: [][]byte{[]byte("tx1")},
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("expected no panic, but got: %v", r)
			}
		}()
		blockDataRequester.receivedMiniBlock(key, miniBlock)
	})

	t.Run("should successfully request transactions for mini block", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()

		// Create preprocessors container with mock
		countCalled := 0
		preprocContainer := containers.NewPreProcessorsContainer()
		txPreproc := &preprocMocks.PreProcessorMock{
			RequestTransactionsForMiniBlockCalled: func(miniBlock *block.MiniBlock) int {
				countCalled++
				return len(miniBlock.TxHashes)
			},
		}
		_ = preprocContainer.Add(block.TxBlock, txPreproc)
		args.PreProcessors = preprocContainer

		blockDataRequester, _ := NewBlockDataRequester(args)

		// Add the mini block to requested items
		key := []byte("hash1")
		_ = blockDataRequester.requestedItemsHandler.Add(string(key))

		// Create mini block
		miniBlock := &block.MiniBlock{
			Type:     block.TxBlock,
			TxHashes: [][]byte{[]byte("tx1"), []byte("tx2"), []byte("tx3")},
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("expected no panic, but got: %v", r)
			}
		}()
		blockDataRequester.receivedMiniBlock(key, miniBlock)
		require.Equal(t, 1, countCalled)
	})

	t.Run("should handle mini block with no transactions", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()

		// Create preprocessors container with mock
		preprocContainer := containers.NewPreProcessorsContainer()
		txPreproc := &preprocMocks.PreProcessorMock{
			RequestTransactionsForMiniBlockCalled: func(miniBlock *block.MiniBlock) int {
				return len(miniBlock.TxHashes)
			},
		}
		_ = preprocContainer.Add(block.TxBlock, txPreproc)
		args.PreProcessors = preprocContainer

		blockDataRequester, _ := NewBlockDataRequester(args)

		// Add the mini block to requested items
		key := []byte("hash1")
		_ = blockDataRequester.requestedItemsHandler.Add(string(key))

		// Create mini block with no transactions
		miniBlock := &block.MiniBlock{
			Type:     block.TxBlock,
			TxHashes: [][]byte{}, // Empty transactions
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("expected no panic, but got: %v", r)
			}
		}()
		blockDataRequester.receivedMiniBlock(key, miniBlock)
	})

	t.Run("should handle multiple mini block types", func(t *testing.T) {
		t.Parallel()
		args := createMockArgs()

		// Create preprocessors container with multiple mocks
		numRequestedTxs := 0
		numRequestedPeer := 0
		preprocContainer := containers.NewPreProcessorsContainer()
		txPreproc := &preprocMocks.PreProcessorMock{
			RequestTransactionsForMiniBlockCalled: func(miniBlock *block.MiniBlock) int {
				require.Equal(t, block.TxBlock, miniBlock.Type)
				numRequestedTxs += len(miniBlock.TxHashes)
				return len(miniBlock.TxHashes)
			},
		}
		peerPreproc := &preprocMocks.PreProcessorMock{
			RequestTransactionsForMiniBlockCalled: func(miniBlock *block.MiniBlock) int {
				require.Equal(t, block.PeerBlock, miniBlock.Type)
				numRequestedPeer += len(miniBlock.TxHashes)
				return len(miniBlock.TxHashes)
			},
		}
		_ = preprocContainer.Add(block.TxBlock, txPreproc)
		_ = preprocContainer.Add(block.PeerBlock, peerPreproc)
		args.PreProcessors = preprocContainer

		blockDataRequester, _ := NewBlockDataRequester(args)

		// Test TxBlock type
		key1 := []byte("hash1")
		_ = blockDataRequester.requestedItemsHandler.Add(string(key1))
		txMiniBlock := &block.MiniBlock{
			Type:     block.TxBlock,
			TxHashes: [][]byte{[]byte("tx1"), []byte("tx2")},
		}
		blockDataRequester.receivedMiniBlock(key1, txMiniBlock)

		// Test PeerBlock type
		key2 := []byte("hash2")
		_ = blockDataRequester.requestedItemsHandler.Add(string(key2))
		peerMiniBlock := &block.MiniBlock{
			Type:     block.PeerBlock,
			TxHashes: [][]byte{[]byte("peer1"), []byte("peer2"), []byte("peer3")},
		}
		blockDataRequester.receivedMiniBlock(key2, peerMiniBlock)
		require.Equal(t, 2, numRequestedTxs)
		require.Equal(t, 3, numRequestedPeer)
	})
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
