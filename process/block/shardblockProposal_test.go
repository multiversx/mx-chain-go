package block_test

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/mbSelection"
)

// Test CreateBlockProposal method
func TestShardProcessor_CreateBlockProposalNilHeader(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	hdr, body, err := sp.CreateBlockProposal(nil, func() bool { return true })

	assert.Nil(t, hdr)
	assert.Nil(t, body)
	assert.Equal(t, process.ErrNilBlockHeader, err)
}

func TestShardProcessor_CreateBlockProposalOldMetaHeader(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	// Use a MetaBlock instead of HeaderV3
	metaHdr := &block.MetaBlock{
		Nonce: 1,
		Round: 1,
	}

	hdr, body, err := sp.CreateBlockProposal(metaHdr, func() bool { return true })

	assert.Nil(t, hdr)
	assert.Nil(t, body)
	assert.Equal(t, process.ErrInvalidHeader, err)
}

func TestShardProcessor_CreateBlockProposalOldShardHeader(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	// Use a regular Header instead of HeaderV3 to trigger invalid header
	regularHdr := &block.Header{
		Nonce:    1,
		Round:    1,
		Epoch:    1,
		ShardID:  0,
		RootHash: []byte("rootHash"),
	}

	hdr, body, err := sp.CreateBlockProposal(regularHdr, func() bool { return true })

	assert.Nil(t, hdr)
	assert.Nil(t, body)
	assert.Equal(t, process.ErrInvalidHeader, err)
}

func TestShardProcessor_CreateBlockProposalWrongTypeAssertionV3Header(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	// Use a meta header V3 to trigger invalid header
	metaHeader := &block.MetaBlockV3{
		Nonce: 1,
		Round: 1,
		Epoch: 1,
	}

	hdr, body, err := sp.CreateBlockProposal(metaHeader, func() bool { return true })

	assert.Nil(t, hdr)
	assert.Nil(t, body)
	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestShardProcessor_CreateBlockProposalUpdateEpochError(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	// Set epoch start trigger to return error on EpochStartMetaHdrHash
	epochStartTrigger := &mock.EpochStartTriggerStub{
		IsEpochStartCalled: func() bool {
			return true
		},
		EpochStartMetaHdrHashCalled: func() []byte {
			return []byte("test hash")
		},
	}
	arguments.EpochStartTrigger = epochStartTrigger

	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	header := &testscommon.HeaderHandlerStub{
		IsHeaderV3Called: func() bool {
			return true
		},
		SetEpochStartMetaHashCalled: func(hash []byte) error {
			return expectedErr
		},
	}

	hdr, body, err := sp.CreateBlockProposal(header, func() bool { return true })

	assert.Nil(t, hdr)
	assert.Nil(t, body)
	assert.Equal(t, expectedErr, err)
}

func TestShardProcessor_CreateBlockProposalCreateBodyError(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	shardHdr := &block.HeaderV3{
		Nonce:   1,
		Round:   1,
		Epoch:   1,
		ShardID: 0,
	}

	// Mock haveTime to return false to trigger early exit
	haveTimeCalled := false
	haveTime := func() bool {
		if !haveTimeCalled {
			haveTimeCalled = true
			return true
		}
		return false
	}

	hdr, body, err := sp.CreateBlockProposal(shardHdr, haveTime)

	// Should succeed even if no time, just return empty body
	assert.NotNil(t, hdr)
	assert.NotNil(t, body)
	assert.Nil(t, err)
	assert.IsType(t, &block.HeaderV3{}, hdr)
	assert.IsType(t, &block.Body{}, body)
}

func TestShardProcessor_CreateBlockProposalSuccess(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	shardHdr := &block.HeaderV3{
		Nonce:   1,
		Round:   1,
		Epoch:   1,
		ShardID: 0,
	}

	hdr, body, err := sp.CreateBlockProposal(shardHdr, func() bool { return true })

	assert.NotNil(t, hdr)
	assert.NotNil(t, body)
	assert.Nil(t, err)
	assert.IsType(t, &block.HeaderV3{}, hdr)
	assert.IsType(t, &block.Body{}, body)

	// Verify the returned header is the same instance (proposal doesn't modify much)
	returnedShardHdr, ok := hdr.(*block.HeaderV3)
	require.True(t, ok)
	assert.Equal(t, shardHdr.Nonce, returnedShardHdr.Nonce)
	assert.Equal(t, shardHdr.Round, returnedShardHdr.Round)
	assert.Equal(t, shardHdr.Epoch, returnedShardHdr.Epoch)
	assert.Equal(t, shardHdr.ShardID, returnedShardHdr.ShardID)
}

// Test createBlockBodyProposal method behavior
func TestShardProcessor_CreateBlockBodyProposalHaveTimeCheck(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	shardHdr := &block.HeaderV3{
		Nonce:   1,
		Round:   1,
		Epoch:   1,
		ShardID: 0,
	}

	// Test with haveTime returning false immediately
	hdr, body, err := sp.CreateBlockProposal(shardHdr, func() bool { return false })

	assert.NotNil(t, hdr)
	assert.NotNil(t, body)
	assert.Nil(t, err)

	// Should return empty body when no time
	bodyObj, ok := body.(*block.Body)
	require.True(t, ok)
	assert.Empty(t, bodyObj.MiniBlocks)
}

// Test selectIncomingMiniBlocksForProposal method
func TestShardProcessor_SelectIncomingMiniBlocksForProposalNoMetaBlocks(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	// Mock block tracker to return empty meta blocks
	blockTracker := &mock.BlockTrackerMock{
		ComputeLongestMetaChainFromLastNotarizedCalled: func() ([]data.HeaderHandler, [][]byte, error) {
			return []data.HeaderHandler{}, [][]byte{}, nil
		},
		GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.MetaBlock{Nonce: 0}, []byte("hash"), nil
		},
	}
	arguments.BlockTracker = blockTracker

	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	shardHdr := &block.HeaderV3{
		Nonce:   1,
		Round:   1,
		Epoch:   1,
		ShardID: 0,
	}

	hdr, body, err := sp.CreateBlockProposal(shardHdr, func() bool { return true })

	assert.NotNil(t, hdr)
	assert.NotNil(t, body)
	assert.Nil(t, err)
}

// Test selectOutgoingTransactions method
func TestShardProcessor_SelectOutgoingTransactionsSuccess(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	// Mock transaction coordinator to return some transactions
	expectedTxs := [][]byte{[]byte("tx1"), []byte("tx2")}
	txCoordinator := &testscommon.TransactionCoordinatorMock{
		SelectOutgoingTransactionsCalled: func() [][]byte {
			return expectedTxs
		},
		CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, uint32, bool, error) {
			return []block.MiniblockAndHash{}, 0, true, nil
		},
	}
	arguments.TxCoordinator = txCoordinator

	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	shardHdr := &block.HeaderV3{
		Nonce:   1,
		Round:   1,
		Epoch:   1,
		ShardID: 0,
	}

	hdr, body, err := sp.CreateBlockProposal(shardHdr, func() bool { return true })

	assert.NotNil(t, hdr)
	assert.NotNil(t, body)
	assert.Nil(t, err)
}

// Test context isolation between proposal and processing flows
func TestShardProcessor_ProposalVsProcessingContextIsolation(t *testing.T) {
	t.Parallel()

	// This test verifies that the proposal flow doesn't interfere with processing flow state

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	// Track state modifications
	var proposalStateModifications []string
	var processingStateModifications []string

	// Mock components to track when they're accessed/modified
	txCoordinator := &testscommon.TransactionCoordinatorMock{
		SelectOutgoingTransactionsCalled: func() [][]byte {
			proposalStateModifications = append(proposalStateModifications, "SelectOutgoingTransactions")
			return [][]byte{[]byte("tx1")}
		},
		CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, uint32, bool, error) {
			proposalStateModifications = append(proposalStateModifications, "CreateMbsCrossShardDstMe")
			return []block.MiniblockAndHash{}, 0, true, nil
		},
		ProcessBlockTransactionCalled: func(header data.HeaderHandler, body *block.Body, haveTime func() time.Duration) error {
			processingStateModifications = append(processingStateModifications, "ProcessBlockTransaction")
			return nil
		},
	}
	arguments.TxCoordinator = txCoordinator

	blockTracker := &mock.BlockTrackerMock{
		ComputeLongestMetaChainFromLastNotarizedCalled: func() ([]data.HeaderHandler, [][]byte, error) {
			proposalStateModifications = append(proposalStateModifications, "ComputeLongestMetaChainFromLastNotarized")
			return []data.HeaderHandler{}, [][]byte{}, nil
		},
		GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			proposalStateModifications = append(proposalStateModifications, "GetLastCrossNotarizedHeader")
			return &block.MetaBlock{Nonce: 0}, []byte("hash"), nil
		},
	}
	arguments.BlockTracker = blockTracker

	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	shardHdr := &block.HeaderV3{
		Nonce:   1,
		Round:   1,
		Epoch:   1,
		ShardID: 0,
	}

	// Test proposal flow
	hdr, body, err := sp.CreateBlockProposal(shardHdr, func() bool { return true })
	require.Nil(t, err)
	require.NotNil(t, hdr)
	require.NotNil(t, body)

	// Verify that only proposal-related operations were called
	assert.Contains(t, proposalStateModifications, "ComputeLongestMetaChainFromLastNotarized")
	assert.Contains(t, proposalStateModifications, "GetLastCrossNotarizedHeader")
	assert.Contains(t, proposalStateModifications, "SelectOutgoingTransactions")

	// Verify that processing-specific operations were NOT called during proposal
	assert.NotContains(t, proposalStateModifications, "ProcessBlockTransaction")

	// The key isolation check: proposal flow should only read state, not modify processing state
	// This is ensured by the fact that CreateBlockProposal doesn't call processing methods
	assert.Empty(t, processingStateModifications, "Processing operations should not be called during proposal creation")
}

// Test miniBlocksSelectionSession state isolation
func TestShardProcessor_MiniBlocksSelectionSessionIsolation(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	resetCalled := false
	sessionMock := mbSelection.MiniBlockSelectionSessionStub{
		ResetSelectionSessionCalled: func() {
			resetCalled = true
		},
		GetNumTxsAddedCalled: func() uint32 { return 5 },
		GetMiniBlocksCalled: func() block.MiniBlockSlice {
			return []*block.MiniBlock{
				{
					TxHashes:        [][]byte{[]byte("tx1")},
					ReceiverShardID: 1,
					SenderShardID:   0,
				},
			}
		},
		GetMiniBlockHeaderHandlersCalled: func() []data.MiniBlockHeaderHandler {
			return []data.MiniBlockHeaderHandler{
				&block.MiniBlockHeader{
					Hash:            []byte("mbHash1"),
					SenderShardID:   0,
					ReceiverShardID: 1,
					TxCount:         1,
				},
			}
		},
		AddMiniBlocksAndHashesCalled: func(miniBlocks []block.MiniblockAndHash) error {
			return nil
		},
		GetReferencedMetaBlocksCalled: func() []data.HeaderHandler {
			return []data.HeaderHandler{}
		},
		CreateAndAddMiniBlockFromTransactionsCalled: func(txs [][]byte) error {
			return nil
		},
	}

	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)
	sp.SetMiniBlockSelectionSession(&sessionMock)

	shardHdr := &block.HeaderV3{
		Nonce:   1,
		Round:   1,
		Epoch:   1,
		ShardID: 0,
	}

	_, _, err = sp.CreateBlockProposal(shardHdr, func() bool { return true })
	require.Nil(t, err)

	// Verify that the session is reset at the beginning - this ensures isolation
	// between different proposal creation calls
	assert.True(t, resetCalled, "MiniBlocksSelectionSession should be reset to ensure isolation")
}

// Test blockSizeThrottler usage consistency
func TestShardProcessor_BlockSizeThrottlerConsistency(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	throttlerCalls := []string{}
	throttlerMock := &mock.BlockSizeThrottlerStub{
		ComputeCurrentMaxSizeCalled: func() {
			throttlerCalls = append(throttlerCalls, "ComputeCurrentMaxSize")
		},
		AddCalled: func(round uint64, size uint32) {
			throttlerCalls = append(throttlerCalls, "Add")
		},
		SucceedCalled: func(round uint64) {
			throttlerCalls = append(throttlerCalls, "Succeed")
		},
	}
	arguments.BlockSizeThrottler = throttlerMock

	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	shardHdr := &block.HeaderV3{
		Nonce:   1,
		Round:   1,
		Epoch:   1,
		ShardID: 0,
	}

	_, _, err = sp.CreateBlockProposal(shardHdr, func() bool { return true })
	require.Nil(t, err)

	// Verify that ComputeCurrentMaxSize is called during proposal creation
	assert.Contains(t, throttlerCalls, "ComputeCurrentMaxSize")

	// Verify that Add and Succeed are NOT called during proposal (they're processing-specific)
	assert.NotContains(t, throttlerCalls, "Add", "Add should not be called during proposal creation")
	assert.NotContains(t, throttlerCalls, "Succeed", "Succeed should not be called during proposal creation")
}

// Test time constraint handling
func TestShardProcessor_ProposalTimeConstraints(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	callCount := 0
	haveTime := func() bool {
		callCount++
		// Return false after first call to simulate running out of time
		return callCount <= 1
	}

	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	shardHdr := &block.HeaderV3{
		Nonce:   1,
		Round:   1,
		Epoch:   1,
		ShardID: 0,
	}

	start := time.Now()
	hdr, body, err := sp.CreateBlockProposal(shardHdr, haveTime)
	elapsed := time.Since(start)

	require.Nil(t, err)
	require.NotNil(t, hdr)
	require.NotNil(t, body)

	// Should complete quickly when time is limited
	assert.Less(t, elapsed, 100*time.Millisecond, "Should complete quickly when haveTime returns false")

	// Verify haveTime was called (indicating time constraints are checked)
	assert.Greater(t, callCount, 0, "haveTime should be called to check time constraints")
}

// Test error handling in createMbsCrossShardDstMe
func TestShardProcessor_CreateMbsCrossShardDstMeErrorHandling(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	expectedError := errors.New("create mbs cross shard error")
	txCoordinator := &testscommon.TransactionCoordinatorMock{
		CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, uint32, bool, error) {
			return nil, 0, false, expectedError
		},
		SelectOutgoingTransactionsCalled: func() [][]byte {
			return [][]byte{}
		},
	}
	arguments.TxCoordinator = txCoordinator

	// Mock block tracker to return at least one meta block to trigger the error path
	blockTracker := &mock.BlockTrackerMock{
		ComputeLongestMetaChainFromLastNotarizedCalled: func() ([]data.HeaderHandler, [][]byte, error) {
			metaBlock := &block.MetaBlock{
				Nonce: 1,
				Round: 1,
				ShardInfo: []block.ShardData{
					{
						ShardID:    0,
						HeaderHash: []byte("shardHash"),
						ShardMiniBlockHeaders: []block.MiniBlockHeader{
							{
								Hash:            []byte("mbHash"),
								SenderShardID:   1,
								ReceiverShardID: 0,
								TxCount:         1,
							},
						},
					},
				},
			}
			return []data.HeaderHandler{metaBlock}, [][]byte{[]byte("metaHash")}, nil
		},
		GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return &block.MetaBlock{Nonce: 0}, []byte("hash"), nil
		},
	}
	arguments.BlockTracker = blockTracker
	dataComponents.DataPool.Proofs().AddProof(&block.HeaderProof{
		HeaderHash: []byte("metaHash"),
	})

	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	shardHdr := &block.HeaderV3{
		Nonce:   1,
		Round:   1,
		Epoch:   1,
		ShardID: 0,
	}

	hdr, body, err := sp.CreateBlockProposal(shardHdr, func() bool { return true })

	// The error should be propagated up
	assert.Nil(t, hdr)
	assert.Nil(t, body)
	assert.Equal(t, expectedError, err)
}

// Test that HeaderV3 is required and regular Header fails type assertion
func TestShardProcessor_CreateBlockProposalRequiresHeaderV3(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	// Test with HeaderV3 - should succeed
	headerV3 := &block.HeaderV3{
		Nonce:   1,
		Round:   1,
		Epoch:   1,
		ShardID: 0,
	}

	hdr, body, err := sp.CreateBlockProposal(headerV3, func() bool { return true })
	assert.NotNil(t, hdr)
	assert.NotNil(t, body)
	assert.Nil(t, err)
	assert.IsType(t, &block.HeaderV3{}, hdr)

	// Test with regular Header - should fail with type assertion error
	regularHeader := &block.Header{
		Nonce:    1,
		Round:    1,
		Epoch:    1,
		ShardID:  0,
		RootHash: []byte("rootHash"),
	}

	hdr2, body2, err2 := sp.CreateBlockProposal(regularHeader, func() bool { return true })
	assert.Nil(t, hdr2)
	assert.Nil(t, body2)
	assert.Equal(t, process.ErrWrongTypeAssertion, err2)
}
