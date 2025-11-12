package coordinator

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/state"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/factory/shard"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	commonMock "github.com/multiversx/mx-chain-go/testscommon/common"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/preprocMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
)

type testData struct {
	hdr            *block.Header
	mbHashes       [][]byte
	miniBlockInfos []*data.MiniBlockInfo
	tx1Hash        []byte
	tx2Hash        []byte
	tx3Hash        []byte
	mb1Info        block.MiniblockAndHash
	mb2Info        block.MiniblockAndHash
}

func TestTransactionCoordinator_CreateMbsCrossShardDstMe_NilHeader(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	miniBlocks, pendingMiniBlocks, numTxs, allAdded, err := tc.CreateMbsCrossShardDstMe(nil, nil)

	require.Equal(t, process.ErrNilHeaderHandler, err)
	require.Equal(t, 0, len(miniBlocks))
	require.Equal(t, 0, len(pendingMiniBlocks))
	require.Equal(t, uint32(0), numTxs)
	require.False(t, allAdded)
}

func TestTransactionCoordinator_CreateMbsCrossShardDstMe_UsesProposalContext(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	// Verify that proposal and execution contexts are separate instances
	require.False(t, reflect.DeepEqual(tc.preProcProposal, tc.preProcExecution))
	require.False(t, reflect.DeepEqual(tc.blockDataRequesterProposal, tc.blockDataRequester))

	td := createHeaderWithMiniBlocksAndTransactions()
	expectedMiniBlocks := []block.MiniblockAndHash{td.mb1Info, td.mb2Info}

	// Mock the block data requester proposal to return empty slice
	proposalBlockDataRequesterCalled := false
	tc.blockDataRequesterProposal = &preprocMocks.BlockDataRequesterStub{
		GetFinalCrossMiniBlockInfoAndRequestMissingCalled: func(header data.HeaderHandler) []*data.MiniBlockInfo {
			proposalBlockDataRequesterCalled = true
			return td.miniBlockInfos
		},
	}

	// Mock the execution block data requester to ensure it's not called
	executionBlockDataRequesterCalled := false
	tc.blockDataRequester = &preprocMocks.BlockDataRequesterStub{
		GetFinalCrossMiniBlockInfoAndRequestMissingCalled: func(header data.HeaderHandler) []*data.MiniBlockInfo {
			executionBlockDataRequesterCalled = true
			return []*data.MiniBlockInfo{}
		},
	}

	// make sure mini blocks are available in the pool
	_ = tc.miniBlockPool.Put(td.mbHashes[0], td.mb1Info.Miniblock, 100)
	_ = tc.miniBlockPool.Put(td.mbHashes[1], td.mb2Info.Miniblock, 100)

	// make sure transactions are available in the pool
	cacheId := process.ShardCacherIdentifier(td.mb1Info.Miniblock.SenderShardID, td.mb1Info.Miniblock.ReceiverShardID)
	ph.Transactions().AddData(td.tx1Hash, &transaction.Transaction{}, 100, cacheId)
	ph.Transactions().AddData(td.tx2Hash, &transaction.Transaction{}, 100, cacheId)
	cacheId = process.ShardCacherIdentifier(td.mb2Info.Miniblock.SenderShardID, td.mb2Info.Miniblock.ReceiverShardID)
	ph.Transactions().AddData(td.tx3Hash, &transaction.Transaction{}, 100, cacheId)

	miniBlocks, _, numTxs, allAdded, err := tc.CreateMbsCrossShardDstMe(td.hdr, nil)

	require.Nil(t, err)
	require.Equal(t, expectedMiniBlocks, miniBlocks)
	require.Equal(t, uint32(3), numTxs)
	require.True(t, allAdded)

	// Verify proposal context was used, not execution context
	require.True(t, proposalBlockDataRequesterCalled)
	require.False(t, executionBlockDataRequesterCalled)
}

func TestTransactionCoordinator_CreateMbsCrossShardDstMe_MaxBlockSizeReached(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	td := createHeaderWithMiniBlocksAndTransactions()

	// Mock block size computation to return max block size reached
	tc.blockSizeComputation = &testscommon.BlockSizeComputationStub{
		IsMaxBlockSizeReachedCalled: func(numNewMiniBlocks, numNewTxs int) bool {
			return true
		},
	}

	// Mock the block data requester to return some mini blocks
	tc.blockDataRequesterProposal = &preprocMocks.BlockDataRequesterStub{
		GetFinalCrossMiniBlockInfoAndRequestMissingCalled: func(header data.HeaderHandler) []*data.MiniBlockInfo {
			return td.miniBlockInfos
		},
	}

	miniBlocks, _, numTxs, allAdded, err := tc.CreateMbsCrossShardDstMe(td.hdr, nil)

	require.Nil(t, err)
	require.Equal(t, 0, len(miniBlocks))
	require.Equal(t, uint32(0), numTxs)
	require.False(t, allAdded) // Not all mini blocks were processed due to size limit
}

func TestTransactionCoordinator_CreateMbsCrossShardDstMe_MiniBlockProcessing(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	td := createHeaderWithMiniBlocksAndTransactions()

	// Mock mini block pool to return the test mini block
	tc.miniBlockPool = &cache.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			if reflect.DeepEqual(key, td.mb1Info.Hash) {
				return td.mb1Info.Miniblock, true
			}
			return nil, false
		},
	}

	// Mock the block data requester to return the mini block info
	tc.blockDataRequesterProposal = &preprocMocks.BlockDataRequesterStub{
		GetFinalCrossMiniBlockInfoAndRequestMissingCalled: func(header data.HeaderHandler) []*data.MiniBlockInfo {
			return td.miniBlockInfos[:1] // Only first mini block
		},
	}

	// Mock preprocessor to simulate no missing transactions
	proposalPreprocessorCalled := false
	tc.preProcProposal.txPreProcessors[block.TxBlock] = &preprocMocks.PreProcessorMock{
		GetTransactionsAndRequestMissingForMiniBlockCalled: func(miniBlock *block.MiniBlock) ([]data.TransactionHandler, int) {
			proposalPreprocessorCalled = true
			require.Equal(t, td.mb1Info.Miniblock, miniBlock)
			return nil, 0 // No missing transactions
		},
	}

	// Mock execution preprocessor to ensure it's not called
	executionPreprocessorCalled := false
	tc.preProcExecution.txPreProcessors[block.TxBlock] = &preprocMocks.PreProcessorMock{
		GetTransactionsAndRequestMissingForMiniBlockCalled: func(miniBlock *block.MiniBlock) ([]data.TransactionHandler, int) {
			executionPreprocessorCalled = true
			return nil, 0
		},
	}

	miniBlocks, _, numTxs, allAdded, err := tc.CreateMbsCrossShardDstMe(td.hdr, nil)

	require.Nil(t, err)
	require.Equal(t, 1, len(miniBlocks))
	require.Equal(t, td.mb1Info.Miniblock, miniBlocks[0].Miniblock)
	require.Equal(t, td.mb1Info.Hash, miniBlocks[0].Hash)
	require.Equal(t, uint32(2), numTxs) // Two transactions in the mini block
	require.True(t, allAdded)

	// Verify proposal preprocessor was used, not execution
	require.True(t, proposalPreprocessorCalled)
	require.False(t, executionPreprocessorCalled)
}

func TestTransactionCoordinator_CreateMbsCrossShardDstMe_MiniBlockProcessing_WithGasComputationError(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	td := createHeaderWithMiniBlocksAndTransactions()

	// Mock mini block pool to return the test mini block
	tc.miniBlockPool = &cache.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			if reflect.DeepEqual(key, td.mb1Info.Hash) {
				return td.mb1Info.Miniblock, true
			}
			return nil, false
		},
	}

	// Mock the block data requester to return the mini block info
	tc.blockDataRequesterProposal = &preprocMocks.BlockDataRequesterStub{
		GetFinalCrossMiniBlockInfoAndRequestMissingCalled: func(header data.HeaderHandler) []*data.MiniBlockInfo {
			return td.miniBlockInfos[:1] // Only first mini block
		},
	}

	// Mock preprocessor to simulate no missing transactions
	proposalPreprocessorCalled := false
	tc.preProcProposal.txPreProcessors[block.TxBlock] = &preprocMocks.PreProcessorMock{
		GetTransactionsAndRequestMissingForMiniBlockCalled: func(miniBlock *block.MiniBlock) ([]data.TransactionHandler, int) {
			proposalPreprocessorCalled = true
			require.Equal(t, td.mb1Info.Miniblock, miniBlock)
			return nil, 0 // No missing transactions
		},
	}

	// Mock execution preprocessor to ensure it's not called
	executionPreprocessorCalled := false
	tc.preProcExecution.txPreProcessors[block.TxBlock] = &preprocMocks.PreProcessorMock{
		GetTransactionsAndRequestMissingForMiniBlockCalled: func(miniBlock *block.MiniBlock) ([]data.TransactionHandler, int) {
			executionPreprocessorCalled = true
			return nil, 0
		},
	}

	tc.gasComputation = &testscommon.GasComputationMock{
		AddIncomingMiniBlocksCalled: func(miniBlocks []data.MiniBlockHeaderHandler, transactions map[string][]data.TransactionHandler) (int, int, error) {
			return 0, 0, errors.New("gas computation error")
		},
	}

	miniBlocks, _, numTxs, allAdded, err := tc.CreateMbsCrossShardDstMe(td.hdr, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "gas computation error")
	require.Nil(t, miniBlocks)
	require.Equal(t, uint32(0), numTxs)
	require.False(t, allAdded) // No mini blocks were processed

	// Verify proposal preprocessor was used, not execution
	require.True(t, proposalPreprocessorCalled)
	require.False(t, executionPreprocessorCalled)
}

func TestTransactionCoordinator_CreateMbsCrossShardDstMe_MiniBlockProcessing_WithPendingMiniBlocks(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	td := createHeaderWithMiniBlocksAndTransactions()

	// Mock mini block pool to return the test mini block
	tc.miniBlockPool = &cache.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			if reflect.DeepEqual(key, td.mb1Info.Hash) {
				return td.mb1Info.Miniblock, true
			}
			return td.mb2Info.Miniblock, true
		},
	}

	// Mock the block data requester to return the mini block info
	tc.blockDataRequesterProposal = &preprocMocks.BlockDataRequesterStub{
		GetFinalCrossMiniBlockInfoAndRequestMissingCalled: func(header data.HeaderHandler) []*data.MiniBlockInfo {
			return td.miniBlockInfos
		},
	}

	// Mock preprocessor to simulate no missing transactions
	proposalPreprocessorCalled := false
	tc.preProcProposal.txPreProcessors[block.TxBlock] = &preprocMocks.PreProcessorMock{
		GetTransactionsAndRequestMissingForMiniBlockCalled: func(miniBlock *block.MiniBlock) ([]data.TransactionHandler, int) {
			proposalPreprocessorCalled = true
			return nil, 0 // No missing transactions
		},
	}

	// Mock execution preprocessor to ensure it's not called
	executionPreprocessorCalled := false
	tc.preProcExecution.txPreProcessors[block.TxBlock] = &preprocMocks.PreProcessorMock{
		GetTransactionsAndRequestMissingForMiniBlockCalled: func(miniBlock *block.MiniBlock) ([]data.TransactionHandler, int) {
			executionPreprocessorCalled = true
			return nil, 0
		},
	}

	tc.gasComputation = &testscommon.GasComputationMock{
		AddIncomingMiniBlocksCalled: func(miniBlocks []data.MiniBlockHeaderHandler, transactions map[string][]data.TransactionHandler) (int, int, error) {
			return 0, 1, nil // last mb added index is 0, so only first mini block is added, num pendings miniblocks is 1, so the second is pending
		},
	}

	miniBlocks, pendingMiniBlocks, numTxs, allAdded, err := tc.CreateMbsCrossShardDstMe(td.hdr, nil)

	require.Nil(t, err)

	require.Equal(t, 1, len(miniBlocks))
	require.Equal(t, td.mb1Info.Miniblock, miniBlocks[0].Miniblock)
	require.Equal(t, td.mb1Info.Hash, miniBlocks[0].Hash)

	require.Equal(t, 1, len(pendingMiniBlocks))
	require.Equal(t, td.mb2Info.Miniblock, pendingMiniBlocks[0].Miniblock)
	require.Equal(t, td.mb2Info.Hash, pendingMiniBlocks[0].Hash)

	require.Equal(t, uint32(3), numTxs)
	require.False(t, allAdded)

	// Verify proposal preprocessor was used, not execution
	require.True(t, proposalPreprocessorCalled)
	require.False(t, executionPreprocessorCalled)
}

func TestTransactionCoordinator_CreateMbsCrossShardDstMe_SkipShardOnMissingMiniBlock(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	td := createHeaderWithMiniBlocksAndTransactions()
	// this should be the first mini block info, which will be missing from shard 2, then the next one will be skipped
	missingMiniBlockInfo := &data.MiniBlockInfo{
		Hash:          []byte("missing_mb_hash"),
		SenderShardID: 2,
		Round:         0,
	}

	// Mock mini block pool to return nil (mini block not found)
	tc.miniBlockPool = &cache.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			if reflect.DeepEqual(key, missingMiniBlockInfo.Hash) {
				return nil, false
			}
			if reflect.DeepEqual(key, td.mb1Info.Hash) {
				return td.mb1Info.Miniblock, true // although the second mini block is available, it should be skipped
			}
			return nil, false // other mini blocks not found as well
		},
	}

	// Mock the block data requester to return multiple mini blocks from same shard
	tc.blockDataRequesterProposal = &preprocMocks.BlockDataRequesterStub{
		GetFinalCrossMiniBlockInfoAndRequestMissingCalled: func(header data.HeaderHandler) []*data.MiniBlockInfo {
			return []*data.MiniBlockInfo{missingMiniBlockInfo, td.miniBlockInfos[0], td.miniBlockInfos[1]}
		},
	}

	miniBlocks, _, numTxs, allAdded, err := tc.CreateMbsCrossShardDstMe(td.hdr, nil)

	require.Nil(t, err)
	require.Equal(t, 0, len(miniBlocks))
	require.Equal(t, uint32(0), numTxs)
	require.False(t, allAdded) // No mini blocks were processed
}

func TestTransactionCoordinator_CreateMbsCrossShardDstMe_SkipShardOnMissingTransactions(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	td := createHeaderWithMiniBlocksAndTransactions()

	tc.miniBlockPool.Put(td.mbHashes[0], td.mb1Info.Miniblock, 100)
	tc.miniBlockPool.Put(td.mbHashes[1], td.mb2Info.Miniblock, 100)

	// Mock the block data requester
	tc.blockDataRequesterProposal = &preprocMocks.BlockDataRequesterStub{
		GetFinalCrossMiniBlockInfoAndRequestMissingCalled: func(header data.HeaderHandler) []*data.MiniBlockInfo {
			return td.miniBlockInfos
		},
	}

	// Make only the second mini block transactions available
	cacheId := process.ShardCacherIdentifier(td.mb2Info.Miniblock.SenderShardID, td.mb2Info.Miniblock.ReceiverShardID)
	ph.Transactions().AddData(td.tx3Hash, &transaction.Transaction{}, 100, cacheId)

	miniBlocks, _, numTxs, allAdded, err := tc.CreateMbsCrossShardDstMe(td.hdr, nil)

	require.Nil(t, err)
	require.Equal(t, 1, len(miniBlocks))
	require.Equal(t, td.mb2Info.Miniblock, miniBlocks[0].Miniblock)
	require.Equal(t, uint32(len(td.mb2Info.Miniblock.TxHashes)), numTxs)
	require.False(t, allAdded)
}

func TestTransactionCoordinator_CreateMbsCrossShardDstMe_ErrorOnUnknownBlockType(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	hdr := createTestHeaderForProposal()
	mbHash := []byte("mb_hash_1")

	// Create test mini block with unknown type
	testMiniBlock := &block.MiniBlock{
		SenderShardID:   1,
		ReceiverShardID: 0,
		Type:            block.Type(99), // Unknown block type
		TxHashes:        [][]byte{[]byte("tx_hash_1")},
	}

	// Mock mini block pool to return the test mini block
	tc.miniBlockPool = &cache.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			if reflect.DeepEqual(key, mbHash) {
				return testMiniBlock, true
			}
			return nil, false
		},
	}

	// Mock the block data requester
	tc.blockDataRequesterProposal = &preprocMocks.BlockDataRequesterStub{
		GetFinalCrossMiniBlockInfoAndRequestMissingCalled: func(header data.HeaderHandler) []*data.MiniBlockInfo {
			return []*data.MiniBlockInfo{
				{
					Hash:          mbHash,
					SenderShardID: 1,
					Round:         1,
				},
			}
		},
	}

	miniBlocks, _, numTxs, allAdded, err := tc.CreateMbsCrossShardDstMe(hdr, nil)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "unknown block type")
	require.Nil(t, miniBlocks)
	require.Equal(t, uint32(0), numTxs)
	require.False(t, allAdded)
}

func TestTransactionCoordinator_SelectOutgoingTransactions_EmptyResult(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	// Mock proposal preprocessors to return empty transactions
	proposalPreprocessorCalled := false
	tc.preProcProposal.txPreProcessors[block.TxBlock] = &preprocMocks.PreProcessorMock{
		SelectOutgoingTransactionsCalled: func(_ uint64) ([][]byte, []data.TransactionHandler, error) {
			proposalPreprocessorCalled = true
			return [][]byte{}, []data.TransactionHandler{}, nil
		},
	}

	// Mock execution preprocessor to ensure it's not called
	executionPreprocessorCalled := false
	tc.preProcExecution.txPreProcessors[block.TxBlock] = &preprocMocks.PreProcessorMock{
		SelectOutgoingTransactionsCalled: func(_ uint64) ([][]byte, []data.TransactionHandler, error) {
			executionPreprocessorCalled = true
			return [][]byte{}, []data.TransactionHandler{}, nil
		},
	}

	txHashes, _ := tc.SelectOutgoingTransactions()

	require.Equal(t, 0, len(txHashes))

	// Verify proposal preprocessor was used, not execution
	require.True(t, proposalPreprocessorCalled)
	require.False(t, executionPreprocessorCalled)
}

func TestTransactionCoordinator_SelectOutgoingTransactions_ReturnsTransactions(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	expectedTxHashes := [][]byte{[]byte("tx_hash_1"), []byte("tx_hash_2")}
	expectedTxs := []data.TransactionHandler{&transaction.Transaction{}, &transaction.Transaction{}}

	// Mock proposal preprocessor to return transactions
	proposalPreprocessorCalled := false
	tc.preProcProposal.txPreProcessors[block.TxBlock] = &preprocMocks.PreProcessorMock{
		SelectOutgoingTransactionsCalled: func(_ uint64) ([][]byte, []data.TransactionHandler, error) {
			proposalPreprocessorCalled = true
			return expectedTxHashes, expectedTxs, nil
		},
	}

	// Mock execution preprocessor to ensure it's not called
	executionPreprocessorCalled := false
	tc.preProcExecution.txPreProcessors[block.TxBlock] = &preprocMocks.PreProcessorMock{
		SelectOutgoingTransactionsCalled: func(_ uint64) ([][]byte, []data.TransactionHandler, error) {
			executionPreprocessorCalled = true
			return [][]byte{}, []data.TransactionHandler{}, nil
		},
	}

	txHashes, _ := tc.SelectOutgoingTransactions()

	require.Equal(t, len(expectedTxHashes), len(txHashes))
	for i, expectedHash := range expectedTxHashes {
		require.Equal(t, expectedHash, txHashes[i])
	}

	// Verify proposal preprocessor was used, not execution
	require.True(t, proposalPreprocessorCalled)
	require.False(t, executionPreprocessorCalled)
}

func TestTransactionCoordinator_SelectOutgoingTransactions_MultipleBlockTypes(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	txHashesType1 := [][]byte{[]byte("tx_hash_1"), []byte("tx_hash_2")}
	txHashesType2 := [][]byte{[]byte("tx_hash_3"), []byte("tx_hash_4")}

	// add transactions to the transactions pool
	cacheId := process.ShardCacherIdentifier(0, 0)
	ph.Transactions().AddData(txHashesType1[0], &transaction.Transaction{SndAddr: []byte("sender1"), Nonce: 0}, 100, cacheId)
	ph.Transactions().AddData(txHashesType1[1], &transaction.Transaction{SndAddr: []byte("sender1"), Nonce: 1}, 100, cacheId)

	// add transactions to the unsigned transactions pool
	ph.UnsignedTransactions().AddData(txHashesType2[0], &transaction.Transaction{SndAddr: []byte("sender2"), Nonce: 0}, 100, cacheId)
	ph.UnsignedTransactions().AddData(txHashesType2[1], &transaction.Transaction{SndAddr: []byte("sender3"), Nonce: 0}, 100, cacheId)

	// Add both block types to the keys
	tc.preProcProposal.keysTxPreProcs = []block.Type{block.TxBlock, block.SmartContractResultBlock}

	txHashes, _ := tc.SelectOutgoingTransactions()

	// Should contain hashes from TxBlock type, for SmartContractsResultsBlock type the selection returns empty
	expectedTotal := len(txHashesType1)
	require.Equal(t, expectedTotal, len(txHashes))

	// Verify all expected hashes are present
	for _, expectedHash := range txHashesType1 {
		found := false
		for _, actualHash := range txHashes {
			if reflect.DeepEqual(expectedHash, actualHash) {
				found = true
				break
			}
		}
		require.True(t, found, "Expected hash %s not found in result", string(expectedHash))
	}
}

func TestTransactionCoordinator_SelectOutgoingTransactions_HandlesErrors(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	// Mock proposal preprocessor to return error
	tc.preProcProposal.txPreProcessors[block.TxBlock] = &preprocMocks.PreProcessorMock{
		SelectOutgoingTransactionsCalled: func(_ uint64) ([][]byte, []data.TransactionHandler, error) {
			return nil, nil, errors.New("test error")
		},
	}

	txHashes, _ := tc.SelectOutgoingTransactions()

	// Function should continue and return empty slice despite error
	require.Equal(t, 0, len(txHashes))
}

func TestTransactionCoordinator_SelectOutgoingTransactions_HandlesNilPreprocessor(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	// Set proposal preprocessor to nil
	tc.preProcProposal.txPreProcessors[block.TxBlock] = nil

	txHashes, _ := tc.SelectOutgoingTransactions()

	// Function should handle nil preprocessor gracefully
	require.Equal(t, 0, len(txHashes))
}

func TestTransactionCoordinator_SelectOutgoingTransactions_AddOutgoingTransactionsError(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)
	txHashesType1 := [][]byte{[]byte("tx_hash_1"), []byte("tx_hash_2")}
	txHashesType2 := [][]byte{[]byte("tx_hash_3"), []byte("tx_hash_4")}

	// add transactions to the transactions pool
	cacheId := process.ShardCacherIdentifier(0, 0)
	ph.Transactions().AddData(txHashesType1[0], &transaction.Transaction{SndAddr: []byte("sender1"), Nonce: 0}, 100, cacheId)
	ph.Transactions().AddData(txHashesType1[1], &transaction.Transaction{SndAddr: []byte("sender1"), Nonce: 1}, 100, cacheId)

	// add transactions to the unsigned transactions pool
	ph.UnsignedTransactions().AddData(txHashesType2[0], &transaction.Transaction{SndAddr: []byte("sender2"), Nonce: 0}, 100, cacheId)
	ph.UnsignedTransactions().AddData(txHashesType2[1], &transaction.Transaction{SndAddr: []byte("sender3"), Nonce: 0}, 100, cacheId)

	// Add both block types to the keys
	tc.preProcProposal.keysTxPreProcs = []block.Type{block.TxBlock, block.SmartContractResultBlock}
	tc.gasComputation = &testscommon.GasComputationMock{
		AddOutgoingTransactionsCalled: func(txHashes [][]byte, transactions []data.TransactionHandler) ([][]byte, []data.MiniBlockHeaderHandler, error) {
			return nil, nil, errors.New("test error in AddOutgoingTransactions")
		},
	}

	require.NotPanics(t, func() {
		selectedTxHashes, selectedPendingIncomingMiniBlocks := tc.SelectOutgoingTransactions()
		// Function should continue and return empty slice despite error
		require.Nil(t, selectedTxHashes)
		require.Nil(t, selectedPendingIncomingMiniBlocks)
	})
}

func TestTransactionCoordinator_CreateMbsCrossShardDstMe_ProcessedMiniBlocksInfo(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	hdr := createTestHeaderForProposal()
	mbHash := []byte("mb_hash_1")

	// Create processed mini blocks info with already processed mini block
	processedMiniBlocksInfo := map[string]*processedMb.ProcessedMiniBlockInfo{
		string(mbHash): {
			FullyProcessed:         true,
			IndexOfLastTxProcessed: 1,
		},
	}

	// Mock the block data requester
	tc.blockDataRequesterProposal = &preprocMocks.BlockDataRequesterStub{
		GetFinalCrossMiniBlockInfoAndRequestMissingCalled: func(header data.HeaderHandler) []*data.MiniBlockInfo {
			return []*data.MiniBlockInfo{
				{
					Hash:          mbHash,
					SenderShardID: 1,
					Round:         1,
				},
			}
		},
	}

	miniBlocks, _, numTxs, allAdded, err := tc.CreateMbsCrossShardDstMe(hdr, processedMiniBlocksInfo)

	require.Nil(t, err)
	require.Equal(t, 0, len(miniBlocks)) // Should skip already processed mini block
	require.Equal(t, uint32(0), numTxs)
	require.True(t, allAdded)
}

func TestTransactionCoordinator_CreateMbsCrossShardDstMe_TypeAssertion(t *testing.T) {
	t.Parallel()

	ph := dataRetrieverMock.NewPoolsHolderMock()
	tc, err := createMockTransactionCoordinatorForProposalTests(ph)
	require.Nil(t, err)
	require.NotNil(t, tc)

	hdr := createTestHeaderForProposal()
	mbHash := []byte("mb_hash_1")

	// Mock mini block pool to return wrong type
	tc.miniBlockPool = &cache.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			if reflect.DeepEqual(key, mbHash) {
				return "invalid_type", true // Wrong type
			}
			return nil, false
		},
	}

	// Mock the block data requester
	tc.blockDataRequesterProposal = &preprocMocks.BlockDataRequesterStub{
		GetFinalCrossMiniBlockInfoAndRequestMissingCalled: func(header data.HeaderHandler) []*data.MiniBlockInfo {
			return []*data.MiniBlockInfo{
				{
					Hash:          mbHash,
					SenderShardID: 1,
					Round:         1,
				},
			}
		},
	}

	miniBlocks, _, numTxs, allAdded, err := tc.CreateMbsCrossShardDstMe(hdr, nil)

	require.Nil(t, err)
	require.Equal(t, 0, len(miniBlocks)) // Should skip due to type assertion failure
	require.Equal(t, uint32(0), numTxs)
	require.False(t, allAdded) // Not all mini blocks were processed
}

func TestTransactionCoordinator_verifyCreatedMiniBlocksSanity(t *testing.T) {
	t.Parallel()

	t.Run("duplicated transactions should error", func(t *testing.T) {
		t.Parallel()

		ph := dataRetrieverMock.NewPoolsHolderMock()
		tc, err := createMockTransactionCoordinatorForProposalTests(ph)
		require.Nil(t, err)
		require.NotNil(t, tc)

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("hash"),
					},
				},
				{
					TxHashes: [][]byte{
						[]byte("hash"),
					},
				},
			},
		}
		err = tc.verifyCreatedMiniBlocksSanity(body)
		require.True(t, errors.Is(err, process.ErrDuplicatedTransaction))
	})

	t.Run("error adding un-executable transactions to the collected transactions - should error", func(t *testing.T) {
		t.Parallel()

		ph := dataRetrieverMock.NewPoolsHolderMock()
		tc, err := createMockTransactionCoordinatorForProposalTests(ph)

		tc.preProcExecution.txPreProcessors[block.TxBlock] = &preprocMocks.PreProcessorMock{
			GetUnExecutableTransactionsCalled: func() map[string]struct{} {
				return map[string]struct{}{
					string([]byte("tx_unexec_hash")): {},
				}
			},
			// make the same preprocessor also report the created miniblocks from me
			GetCreatedMiniBlocksFromMeCalled: func() block.MiniBlockSlice {
				return block.MiniBlockSlice{
					&block.MiniBlock{
						SenderShardID:   0,
						ReceiverShardID: 0,
						TxHashes: [][]byte{
							[]byte("tx_unexec_hash"),
						},
					},
				}
			},
		}
		require.Nil(t, err)
		require.NotNil(t, tc)

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					SenderShardID:   0,
					ReceiverShardID: 0,
					TxHashes: [][]byte{
						[]byte("tx_unexec_hash"),
					},
				},
			},
		}

		err = tc.verifyCreatedMiniBlocksSanity(body)
		require.True(t, errors.Is(err, process.ErrDuplicatedTransaction))
	})

	t.Run("transactions mismatch should error", func(t *testing.T) {
		t.Parallel()

		ph := dataRetrieverMock.NewPoolsHolderMock()
		tc, err := createMockTransactionCoordinatorForProposalTests(ph)
		require.Nil(t, err)
		require.NotNil(t, tc)

		// preprocessor says we created tx 'a'
		tc.preProcExecution.txPreProcessors[block.TxBlock] = &preprocMocks.PreProcessorMock{
			GetUnExecutableTransactionsCalled: func() map[string]struct{} {
				return map[string]struct{}{} // no duplicates this time
			},
			GetCreatedMiniBlocksFromMeCalled: func() block.MiniBlockSlice {
				return block.MiniBlockSlice{
					&block.MiniBlock{
						SenderShardID:   0,
						ReceiverShardID: 0,
						TxHashes: [][]byte{
							[]byte("tx_a"),
						},
					},
				}
			},
		}

		// but the block body includes tx 'b'
		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					SenderShardID:   0,
					ReceiverShardID: 0,
					TxHashes: [][]byte{
						[]byte("tx_b"),
					},
				},
			},
		}

		err = tc.verifyCreatedMiniBlocksSanity(body)
		require.True(t, errors.Is(err, process.ErrTransactionsMismatch))
	})

	t.Run("verifyCreatedMiniBlocksSanity should error on collectTransactionsFromMiniBlocks", func(t *testing.T) {
		t.Parallel()

		ph := dataRetrieverMock.NewPoolsHolderMock()
		tc, err := createMockTransactionCoordinatorForProposalTests(ph)
		require.NoError(t, err)
		require.NotNil(t, tc)

		tc.preProcExecution.txPreProcessors[block.TxBlock] = &preprocMocks.PreProcessorMock{
			GetCreatedMiniBlocksFromMeCalled: func() block.MiniBlockSlice {
				return block.MiniBlockSlice{
					{
						SenderShardID:   0,
						ReceiverShardID: 0,
						TxHashes: [][]byte{
							[]byte("tx_dup"),
							[]byte("tx_dup"), // duplicate triggers collectTransactionsFromMiniBlocks error
						},
					},
				}
			},
			GetUnExecutableTransactionsCalled: func() map[string]struct{} {
				return nil
			},
		}

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					SenderShardID:   0,
					ReceiverShardID: 0,
					TxHashes: [][]byte{
						[]byte("tx_dup"),
					},
				},
			},
		}

		err = tc.verifyCreatedMiniBlocksSanity(body)
		require.Error(t, err)
		require.Contains(t, err.Error(), "for created miniBlocks")
		require.True(t, errors.Is(err, process.ErrDuplicatedTransaction))
	})

	t.Run("should work empty body", func(t *testing.T) {
		t.Parallel()

		ph := dataRetrieverMock.NewPoolsHolderMock()
		tc, err := createMockTransactionCoordinatorForProposalTests(ph)
		require.Nil(t, err)
		require.NotNil(t, tc)

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{},
			},
		}
		err = tc.verifyCreatedMiniBlocksSanity(body)
		require.NoError(t, err)
	})
	t.Run("should work with unexecutable transactions", func(t *testing.T) {
		t.Parallel()

		ph := dataRetrieverMock.NewPoolsHolderMock()
		tc, err := createMockTransactionCoordinatorForProposalTests(ph)
		require.NoError(t, err)
		require.NotNil(t, tc)

		// first mini-block tx3 is un-executable
		mb1 := &block.MiniBlock{
			SenderShardID: 0,
			Type:          block.TxBlock,
			TxHashes:      [][]byte{[]byte("tx1"), []byte("tx2"), []byte("tx_unexec")},
		}
		// second mini-block incomming - will be ignored for created txs verification
		mb2 := &block.MiniBlock{
			SenderShardID:   1,
			ReceiverShardID: 0,
			Type:            block.TxBlock,
			TxHashes:        [][]byte{[]byte("tx4")},
		}

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{mb1, mb2},
		}

		// Preprocessor still returns all created mini-blocks
		tc.preProcExecution.txPreProcessors[block.TxBlock] = &preprocMocks.PreProcessorMock{
			GetCreatedMiniBlocksFromMeCalled: func() block.MiniBlockSlice {
				return block.MiniBlockSlice{
					{
						SenderShardID:   0,
						ReceiverShardID: 0,
						Type:            block.TxBlock,
						TxHashes:        [][]byte{[]byte("tx1"), []byte("tx2")},
					},
				}
			},
			GetUnExecutableTransactionsCalled: func() map[string]struct{} {
				return map[string]struct{}{
					string([]byte("tx_unexec")): {},
				}
			},
		}

		err = tc.verifyCreatedMiniBlocksSanity(body)
		require.NoError(t, err)
	})

}

func createPreProcessorContainerWithPoolsHolder(poolsHolder dataRetriever.PoolsHolder) process.PreProcessorsContainer {
	preProcessorsFactoryArgs := shard.ArgsPreProcessorsContainerFactory{
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(5),
		Store:            initStore(),
		Marshalizer:      &mock.MarshalizerMock{},
		Hasher:           &hashingMocks.HasherMock{},
		DataPool:         poolsHolder,
		PubkeyConverter:  createMockPubkeyConverter(),
		Accounts:         &stateMock.AccountsStub{},
		AccountsProposal: &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return nil, nil
			},
			GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
				return nil, state.ErrAccNotFound // only new accounts
			},
		},
		RequestHandler:               &testscommon.RequestHandlerStub{},
		TxProcessor:                  &testscommon.TxProcessorMock{},
		ScProcessor:                  &testscommon.SCProcessorMock{},
		ScResultProcessor:            &testscommon.SmartContractResultsProcessorMock{},
		RewardsTxProcessor:           &testscommon.RewardTxProcessorMock{},
		EconomicsFee:                 FeeHandlerMock(),
		GasHandler:                   &testscommon.GasHandlerStub{},
		BlockTracker:                 &mock.BlockTrackerMock{},
		BlockSizeComputation:         &testscommon.BlockSizeComputationStub{},
		BalanceComputation:           &testscommon.BalanceComputationStub{},
		EnableEpochsHandler:          enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
		EnableRoundsHandler:          &testscommon.EnableRoundsHandlerStub{},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
		TxExecutionOrderHandler:      &commonMock.TxExecutionOrderHandlerStub{},
		TxCacheSelectionConfig:       createMockTxCacheSelectionConfig(),
	}

	preFactory, _ := shard.NewPreProcessorsContainerFactory(preProcessorsFactoryArgs)
	container, _ := preFactory.Create()

	return container
}

func createMockTransactionCoordinatorForProposalTests(poolsHolder dataRetriever.PoolsHolder) (*transactionCoordinator, error) {
	args := createMockTransactionCoordinatorArguments()

	// Create separate preprocessor containers for execution and proposal to ensure context isolation
	args.PreProcessors = createPreProcessorContainerWithPoolsHolder(poolsHolder)
	args.PreProcessorsProposal = createPreProcessorContainerWithPoolsHolder(poolsHolder)

	createAndAddBlockDataRequesters(
		&args,
		5,
		poolsHolder,
		&testscommon.RequestHandlerStub{},
		args.PreProcessors,
		args.PreProcessorsProposal,
	)

	return NewTransactionCoordinator(args)
}

func createTestHeaderForProposal() *block.Header {
	return &block.Header{
		Round:     1,
		Nonce:     2,
		ShardID:   0,
		TimeStamp: uint64(time.Now().Unix()),
	}
}

func createHeaderWithMiniBlocksAndTransactions() testData {
	hdr := createTestHeaderForProposal()
	mbHashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	miniBlockInfos := []*data.MiniBlockInfo{
		{
			Hash:          mbHashes[0],
			SenderShardID: 2,
			Round:         0,
		},
		{
			Hash:          mbHashes[1],
			SenderShardID: 1,
			Round:         0,
		},
	}

	tx1Hash := []byte("tx1")
	tx2Hash := []byte("tx2")
	tx3Hash := []byte("tx3")
	mb1Info := block.MiniblockAndHash{
		Miniblock: &block.MiniBlock{
			SenderShardID:   2,
			ReceiverShardID: 0,
			Type:            block.TxBlock,
			TxHashes:        [][]byte{tx1Hash, tx2Hash},
		},
		Hash: mbHashes[0],
	}

	mb2Info := block.MiniblockAndHash{
		Miniblock: &block.MiniBlock{
			SenderShardID:   1,
			ReceiverShardID: 0,
			Type:            block.TxBlock,
			TxHashes:        [][]byte{tx3Hash},
		},
		Hash: mbHashes[1],
	}
	hdr.MiniBlockHeaders = []block.MiniBlockHeader{
		{
			ReceiverShardID: 0,
			Hash:            mbHashes[0],
			Type:            block.TxBlock,
		},
		{
			ReceiverShardID: 0,
			Hash:            mbHashes[1],
			Type:            block.TxBlock,
		},
	}

	return testData{
		hdr:            hdr,
		mbHashes:       mbHashes,
		miniBlockInfos: miniBlockInfos,
		tx1Hash:        tx1Hash,
		tx2Hash:        tx2Hash,
		tx3Hash:        tx3Hash,
		mb1Info:        mb1Info,
		mb2Info:        mb2Info,
	}
}
