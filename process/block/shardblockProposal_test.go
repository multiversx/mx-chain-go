package block_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	"github.com/stretchr/testify/require"

	retriever "github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/estimator"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/executionTrack"
	"github.com/multiversx/mx-chain-go/testscommon/mbSelection"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
)

func getSimpleHeaderV3Mock() *testscommon.HeaderHandlerStub {
	return &testscommon.HeaderHandlerStub{
		IsHeaderV3Called: func() bool {
			return true
		},
	}
}

func haveTimeTrue() bool {
	return true
}

func haveTimeFalse() bool {
	return false
}

type shardProcessorTest interface {
	CreateBlockProposal(
		initialHdr data.HeaderHandler,
		haveTime func() bool,
	) (data.HeaderHandler, data.BodyHandler, error)
}

func TestShardProcessor_CreateBlockProposal(t *testing.T) {
	t.Parallel()

	t.Run("nil header", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		checkCreateBlockProposalResult(t, sp, nil, haveTimeTrue, process.ErrNilBlockHeader)
	})
	t.Run("not header v3", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		metaHdr := &block.MetaBlock{
			Nonce: 1,
			Round: 1,
		}

		checkCreateBlockProposalResult(t, sp, metaHdr, haveTimeTrue, process.ErrInvalidHeader)
	})
	t.Run("meta header v3", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		metaHdr := &block.MetaBlockV3{
			Nonce: 1,
			Round: 1,
		}

		checkCreateBlockProposalResult(t, sp, metaHdr, haveTimeTrue, process.ErrWrongTypeAssertion)
	})
	t.Run("updateEpochIfNeeded fails due to error on SetEpochStartMetaHash", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
			IsEpochStartCalled: func() bool {
				return true
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		header := getSimpleHeaderV3Mock()
		header.SetEpochStartMetaHashCalled = func(hash []byte) error {
			return expectedErr
		}

		checkCreateBlockProposalResult(t, sp, header, haveTimeTrue, expectedErr)
	})
	t.Run("updateEpochIfNeeded fails due to error on SetEpoch", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
			IsEpochStartCalled: func() bool {
				return true
			},
			MetaEpochCalled: func() uint32 {
				return 1
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		header := getSimpleHeaderV3Mock()
		header.EpochField = 2 // different from the one from epochStartTrigger
		header.SetEpochCalled = func(epoch uint32) error {
			return expectedErr
		}

		checkCreateBlockProposalResult(t, sp, header, haveTimeTrue, expectedErr)
	})
	t.Run("selectIncomingMiniBlocksForProposal fails due to error on ComputeLongestMetaChainFromLastNotarized", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.BlockTracker = &mock.BlockTrackerMock{
			ComputeLongestMetaChainFromLastNotarizedCalled: func() ([]data.HeaderHandler, [][]byte, error) {
				return nil, nil, expectedErr
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		checkCreateBlockProposalResult(t, sp, getSimpleHeaderV3Mock(), haveTimeTrue, expectedErr)
	})
	t.Run("selectIncomingMiniBlocksForProposal fails due to error on GetLastCrossNotarizedHeader", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.BlockTracker = &mock.BlockTrackerMock{
			ComputeLongestMetaChainFromLastNotarizedCalled: func() ([]data.HeaderHandler, [][]byte, error) {
				return []data.HeaderHandler{&block.MetaBlockV3{
					ShardInfo: []block.ShardData{
						{
							ShardID: 1,
							ShardMiniBlockHeaders: []block.MiniBlockHeader{
								{
									SenderShardID:   1,
									ReceiverShardID: 0,
								},
							},
						},
					},
					MiniBlockHeaders: []block.MiniBlockHeader{
						{},
					},
				}}, [][]byte{[]byte("hash")}, nil
			},
			GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
				return nil, nil, expectedErr
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		checkCreateBlockProposalResult(t, sp, getSimpleHeaderV3Mock(), haveTimeTrue, expectedErr)
	})
	t.Run("selectIncomingMiniBlocksForProposal fails due to error on selectIncomingMiniBlocks", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		headers := dataComponents.DataPool.Headers()
		dataComponents.DataPool = &dataRetriever.PoolsHolderStub{
			ProofsCalled: func() retriever.ProofsPool {
				return &dataRetriever.ProofsPoolMock{
					HasProofCalled: func(shardID uint32, headerHash []byte) bool {
						return true
					},
				}
			},
			HeadersCalled: func() retriever.HeadersPool {
				return headers
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.BlockTracker = &mock.BlockTrackerMock{
			ComputeLongestMetaChainFromLastNotarizedCalled: func() ([]data.HeaderHandler, [][]byte, error) {
				return []data.HeaderHandler{&block.MetaBlockV3{
					ShardInfo: []block.ShardData{
						{
							ShardID: 1,
							ShardMiniBlockHeaders: []block.MiniBlockHeader{
								{
									SenderShardID:   1,
									ReceiverShardID: 0,
								},
							},
						},
					},
					MiniBlockHeaders: []block.MiniBlockHeader{
						{},
					},
				}}, [][]byte{[]byte("hash")}, nil
			},
			GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
				return &block.MetaBlockV3{}, []byte("hash"), nil // dummy
			},
		}
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, uint32, bool, error) {
				return nil, 0, false, expectedError
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		checkCreateBlockProposalResult(t, sp, getSimpleHeaderV3Mock(), haveTimeTrue, expectedErr)
	})
	t.Run("createProposalMiniBlocks fails due to error on CreateAndAddMiniBlockFromTransactions", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			CreateAndAddMiniBlockFromTransactionsCalled: func(txHashes [][]byte) error {
				return expectedErr
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		checkCreateBlockProposalResult(t, sp, getSimpleHeaderV3Mock(), haveTimeTrue, expectedErr)
	})
	t.Run("checkMiniBlocksAndMiniBlockHeadersConsistency fails due to different lengths", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
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
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		checkCreateBlockProposalResult(t, sp, getSimpleHeaderV3Mock(), haveTimeTrue, process.ErrNumOfMiniBlocksAndMiniBlocksHeadersMismatch)
	})
	t.Run("addExecutionResultsOnHeader fails due to error on GetPendingExecutionResults", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ExecutionResultsTracker = &executionTrack.ExecutionResultsTrackerStub{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return nil, expectedErr
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		checkCreateBlockProposalResult(t, sp, getSimpleHeaderV3Mock(), haveTimeTrue, expectedErr)
	})
	t.Run("addExecutionResultsOnHeader fails due to error on SetLastExecutionResultHandler", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.HeaderV2{} // using V2 for simplicity
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash")
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ExecutionResultsInclusionEstimator = &processMocks.InclusionEstimatorMock{
			DecideCalled: func(lastNotarised *estimator.LastExecutionResultForInclusion, pending []data.BaseExecutionResultHandler, currentHdrTsMs uint64) (allowed int) {
				return 1 // coverage only
			},
		}
		arguments.ExecutionResultsTracker = &executionTrack.ExecutionResultsTrackerStub{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return []data.BaseExecutionResultHandler{
					&block.ExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{},
					},
				}, nil
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		header := getSimpleHeaderV3Mock()
		header.SetLastExecutionResultHandlerCalled = func(resultHandler data.LastExecutionResultHandler) error {
			return expectedErr
		}
		checkCreateBlockProposalResult(t, sp, header, haveTimeTrue, expectedErr)
	})
	t.Run("SetMiniBlockHeaderHandlers failure", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		header := getSimpleHeaderV3Mock()
		header.SetMiniBlockHeaderHandlersCalled = func(mbsHandlers []data.MiniBlockHeaderHandler) error {
			return expectedErr
		}
		checkCreateBlockProposalResult(t, sp, header, haveTimeFalse, expectedErr) // using haveTimeFalse for extra coverage
	})
	t.Run("SetMetaBlockHashes failure", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		header := getSimpleHeaderV3Mock()
		header.SetMetaBlockHashesCalled = func(hashes [][]byte) error {
			return expectedErr
		}
		checkCreateBlockProposalResult(t, sp, header, haveTimeTrue, expectedErr)
	})
	t.Run("SetTxCount failure", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		header := getSimpleHeaderV3Mock()
		header.SetTxCountCalled = func(count uint32) error {
			return expectedErr
		}
		checkCreateBlockProposalResult(t, sp, header, haveTimeTrue, expectedErr)
	})
	t.Run("addExecutionResultsOnHeader fails due to error on CreateDataForInclusionEstimation", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		checkCreateBlockProposalResult(t, sp, getSimpleHeaderV3Mock(), haveTimeTrue, process.ErrNilHeaderHandler)
	})
	t.Run("Marshal failure", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.HeaderV2{} // using V2 for simplicity
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash")
			},
		}
		coreComponents.IntMarsh = &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		checkCreateBlockProposalResult(t, sp, getSimpleHeaderV3Mock(), haveTimeTrue, expectedErr)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.HeaderV2{} // using V2 for simplicity
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash")
			},
		}
		headers := dataComponents.DataPool.Headers()
		dataComponents.DataPool = &dataRetriever.PoolsHolderStub{
			ProofsCalled: func() retriever.ProofsPool {
				return &dataRetriever.ProofsPoolMock{
					HasProofCalled: func(shardID uint32, headerHash []byte) bool {
						return true
					},
				}
			},
			HeadersCalled: func() retriever.HeadersPool {
				return headers
			},
		}

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.BlockTracker = &mock.BlockTrackerMock{
			ComputeLongestMetaChainFromLastNotarizedCalled: func() ([]data.HeaderHandler, [][]byte, error) {
				return []data.HeaderHandler{&block.MetaBlockV3{
						ShardInfo: []block.ShardData{
							{
								ShardID: 1,
								ShardMiniBlockHeaders: []block.MiniBlockHeader{
									{
										SenderShardID:   1,
										ReceiverShardID: 0,
									},
								},
							},
							// for extra coverage, should be skipped as it is empty
							{
								ShardID:               1,
								ShardMiniBlockHeaders: []block.MiniBlockHeader{},
							},
						},
						MiniBlockHeaders: []block.MiniBlockHeader{
							{},
						},
					}},
					[][]byte{[]byte("hash_ok"), []byte("hash_empty")},
					nil
			},
			GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
				return &block.MetaBlockV3{}, []byte("hash"), nil // dummy
			},
		}
		providedMb := &block.MiniBlock{
			TxHashes: [][]byte{[]byte("tx_hash")},
		}
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, uint32, bool, error) {
				return []block.MiniblockAndHash{
					{
						Miniblock: providedMb,
						Hash:      []byte("providedMB"),
					},
				}, 0, true, nil
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		hdr, body, err := sp.CreateBlockProposal(getSimpleHeaderV3Mock(), haveTimeTrue)
		require.NoError(t, err)
		require.NotNil(t, hdr)
		require.NotNil(t, body)

		rawBody, ok := body.(*block.Body)
		require.True(t, ok)
		require.Len(t, rawBody.MiniBlocks, 1)
		require.Equal(t, providedMb, rawBody.MiniBlocks[0])
	})
}

func TestShardProcessor_SelectIncomingMiniBlocks(t *testing.T) {
	t.Parallel()

	providedLastCrossNotarizedMetaHdr := &block.MetaBlockV3{
		Nonce: 1,
	}
	providedOrderedMetaBlocks := []data.HeaderHandler{
		&block.MetaBlockV3{
			Nonce: 2,
		},
		&block.MetaBlockV3{
			Nonce: 3,
		},
	}
	providedOrderedMetaBlocksHashes := [][]byte{
		[]byte("hash2"),
		[]byte("hash3"),
	}
	t.Run("no time left should break and return nil", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			GetReferencedMetaBlocksCalled: func() []data.HeaderHandler {
				require.Fail(t, "should have not been called")
				return nil
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		err = sp.SelectIncomingMiniBlocks(providedLastCrossNotarizedMetaHdr, providedOrderedMetaBlocks, providedOrderedMetaBlocksHashes, haveTimeFalse)
		require.NoError(t, err)
	})
	t.Run("too many referenced blocks should break and return nil", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		headers := dataComponents.DataPool.Headers()
		dataComponents.DataPool = &dataRetriever.PoolsHolderStub{
			HeadersCalled: func() retriever.HeadersPool {
				return headers
			},
			ProofsCalled: func() retriever.ProofsPool {
				return &dataRetriever.ProofsPoolMock{
					HasProofCalled: func(shardID uint32, headerHash []byte) bool {
						require.Fail(t, "should have not been called")
						return true
					},
				}
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			GetReferencedMetaBlocksCalled: func() []data.HeaderHandler {
				return make([]data.HeaderHandler, process.MaxMetaHeadersAllowedInOneShardBlock)
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		err = sp.SelectIncomingMiniBlocks(providedLastCrossNotarizedMetaHdr, providedOrderedMetaBlocks, providedOrderedMetaBlocksHashes, haveTimeTrue)
		require.NoError(t, err)
	})
	t.Run("nonce too high for one meta header should break", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		cntHasProof := 0
		headers := dataComponents.DataPool.Headers()
		dataComponents.DataPool = &dataRetriever.PoolsHolderStub{
			HeadersCalled: func() retriever.HeadersPool {
				return headers
			},
			ProofsCalled: func() retriever.ProofsPool {
				return &dataRetriever.ProofsPoolMock{
					HasProofCalled: func(shardID uint32, headerHash []byte) bool {
						cntHasProof++
						return true
					},
				}
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		providedOrderedMetaBlocksCopy := make([]data.HeaderHandler, len(providedOrderedMetaBlocks))
		copy(providedOrderedMetaBlocksCopy, providedOrderedMetaBlocks)
		_ = providedOrderedMetaBlocksCopy[1].SetNonce(10)
		err = sp.SelectIncomingMiniBlocks(providedLastCrossNotarizedMetaHdr, providedOrderedMetaBlocksCopy, providedOrderedMetaBlocksHashes, haveTimeTrue)
		require.NoError(t, err)
		require.Equal(t, 1, cntHasProof)
	})
	t.Run("missing proof for one meta header should break", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		headers := dataComponents.DataPool.Headers()
		dataComponents.DataPool = &dataRetriever.PoolsHolderStub{
			HeadersCalled: func() retriever.HeadersPool {
				return headers
			},
			ProofsCalled: func() retriever.ProofsPool {
				return &dataRetriever.ProofsPoolMock{
					HasProofCalled: func(shardID uint32, headerHash []byte) bool {
						return false
					},
				}
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		orderedMetaBlocks := []data.HeaderHandler{
			&testscommon.HeaderHandlerStub{
				GetMiniBlockHeadersWithDstCalled: func(destId uint32) map[string]uint32 {
					require.Fail(t, "should have not been called")
					return nil
				},
				GetNonceCalled: func() uint64 {
					return 2
				},
			},
		}
		err = sp.SelectIncomingMiniBlocks(providedLastCrossNotarizedMetaHdr, orderedMetaBlocks, providedOrderedMetaBlocksHashes, haveTimeTrue)
		require.NoError(t, err)
	})
	t.Run("createMbsCrossShardDstMe fails due to error on AddMiniBlocksAndHashes", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		headers := dataComponents.DataPool.Headers()
		dataComponents.DataPool = &dataRetriever.PoolsHolderStub{
			HeadersCalled: func() retriever.HeadersPool {
				return headers
			},
			ProofsCalled: func() retriever.ProofsPool {
				return &dataRetriever.ProofsPoolMock{
					HasProofCalled: func(shardID uint32, headerHash []byte) bool {
						return true
					},
				}
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			AddMiniBlocksAndHashesCalled: func(miniBlocksAndHashes []block.MiniblockAndHash) error {
				return expectedErr
			},
		}
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, uint32, bool, error) {
				return []block.MiniblockAndHash{
					{
						Miniblock: &block.MiniBlock{},
						Hash:      []byte("providedMB"),
					},
				}, 0, true, nil
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		orderedMetaBlocks := []data.HeaderHandler{
			&testscommon.HeaderHandlerStub{
				GetMiniBlockHeadersWithDstCalled: func(destId uint32) map[string]uint32 {
					return map[string]uint32{
						"hash2_0": 1,
						"hash2_1": 2,
					}
				},
				GetNonceCalled: func() uint64 {
					return 2
				},
			},
		}
		orderedMetaBlocksHashes := providedOrderedMetaBlocksHashes
		orderedMetaBlocksHashes = orderedMetaBlocksHashes[:1]
		err = sp.SelectIncomingMiniBlocks(providedLastCrossNotarizedMetaHdr, orderedMetaBlocks, orderedMetaBlocksHashes, haveTimeTrue)
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		headers := dataComponents.DataPool.Headers()
		dataComponents.DataPool = &dataRetriever.PoolsHolderStub{
			HeadersCalled: func() retriever.HeadersPool {
				return headers
			},
			ProofsCalled: func() retriever.ProofsPool {
				return &dataRetriever.ProofsPoolMock{
					HasProofCalled: func(shardID uint32, headerHash []byte) bool {
						return true
					},
				}
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		cntAddReferencedMetaBlockCalled := 0
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			AddReferencedMetaBlockCalled: func(metaBlock data.HeaderHandler, metaBlockHash []byte) {
				cntAddReferencedMetaBlockCalled++
			},
		}
		cntCreateMbsCrossShardDstMeCalled := 0
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, uint32, bool, error) {
				cntCreateMbsCrossShardDstMeCalled++
				if cntCreateMbsCrossShardDstMeCalled < 2 {
					return []block.MiniblockAndHash{
						{
							Miniblock: &block.MiniBlock{},
							Hash:      []byte("providedMB"),
						},
					}, 0, true, nil
				}
				return nil, 0, false, nil // shouldContinue = false -> only for coverage
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		orderedMetaBlocks := []data.HeaderHandler{
			&testscommon.HeaderHandlerStub{
				GetMiniBlockHeadersWithDstCalled: func(destId uint32) map[string]uint32 {
					return map[string]uint32{
						"hash2_0": 1,
						"hash2_1": 2,
					}
				},
				GetNonceCalled: func() uint64 {
					return 2
				},
			},
			&testscommon.HeaderHandlerStub{
				GetMiniBlockHeadersWithDstCalled: func(destId uint32) map[string]uint32 {
					return make(map[string]uint32) // empty
				},
				GetNonceCalled: func() uint64 {
					return 2 // same nonce
				},
			},
			&testscommon.HeaderHandlerStub{
				GetMiniBlockHeadersWithDstCalled: func(destId uint32) map[string]uint32 {
					return map[string]uint32{
						"hash2_2": 1,
						"hash2_3": 2,
					}
				},
				GetNonceCalled: func() uint64 {
					return 2 // same nonce
				},
			},
		}
		orderedMetaBlocksHashes := providedOrderedMetaBlocksHashes
		orderedMetaBlocksHashes = append(orderedMetaBlocksHashes, []byte("hash4"))
		err = sp.SelectIncomingMiniBlocks(providedLastCrossNotarizedMetaHdr, orderedMetaBlocks, orderedMetaBlocksHashes, haveTimeTrue)
		require.NoError(t, err)
		require.Equal(t, 2, cntAddReferencedMetaBlockCalled) // should be called twice, the third hdr returns shouldContinue false
	})
}

func TestShardProcessor_VerifyBlockProposal(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local error")

	t.Run("nil header should error", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		body := &block.Body{}
		err = sp.VerifyBlockProposal(nil, body, haveTime)
		require.Equal(t, process.ErrNilBlockHeader, err)
	})

	t.Run("block hash does not match should request prev header hash", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		currentBlockHeader := &block.Header{}
		_ = dataComponents.BlockChain.SetCurrentBlockHeaderAndRootHash(currentBlockHeader, []byte("root"))
		dataComponents.BlockChain.SetCurrentBlockHeaderHash([]byte("wrong"))

		called := false
		wg := &sync.WaitGroup{}
		wg.Add(1)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.RequestHandler = &testscommon.RequestHandlerStub{
			RequestShardHeaderForEpochCalled: func(shardID uint32, hash []byte, epoch uint32) {
				called = true
				require.Equal(t, "prevHash", string(hash))
				wg.Done()
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		body := &block.Body{}
		header := &block.Header{
			Nonce:    1,
			Round:    2,
			Epoch:    1,
			PrevHash: []byte("prevHash"),
		}
		err = sp.VerifyBlockProposal(header, body, haveTime)
		require.Equal(t, process.ErrBlockHashDoesNotMatch, err)

		wg.Wait()
		require.True(t, called)
	})

	t.Run("wrong header type should error", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		body := &block.Body{}
		header := &block.MetaBlock{
			Nonce: 1,
		}
		err = sp.VerifyBlockProposal(header, body, haveTime)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})

	t.Run("wrong header version should error", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		body := &block.Body{}
		header := &block.Header{
			Nonce: 1,
		}
		err = sp.VerifyBlockProposal(header, body, haveTime)
		require.Equal(t, process.ErrInvalidHeader, err)
	})

	t.Run("wrong body should error", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		body := &wrongBody{}
		header := &block.HeaderV3{
			Nonce: 1,
		}
		err = sp.VerifyBlockProposal(header, body, haveTime)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})
	t.Run("different mbs header from body vs from header should error", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{nil},
		}

		header := &block.HeaderV3{
			Nonce: 1,
			MiniBlockHeaders: []block.MiniBlockHeader{
				{},
			},
		}
		err = sp.VerifyBlockProposal(header, body, haveTime)
		require.Equal(t, process.ErrNilMiniBlock, err)
	})

	t.Run("header execution results verification fails should error", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.ExecutionResultsVerifier = &processMocks.ExecutionResultsVerifierMock{
			VerifyHeaderExecutionResultsCalled: func(header data.HeaderHandler) error {
				return localErr
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		body := &block.Body{}

		header := &block.HeaderV3{
			Nonce:            1,
			MiniBlockHeaders: []block.MiniBlockHeader{},
		}
		err = sp.VerifyBlockProposal(header, body, haveTime)
		require.Equal(t, localErr, err)
	})

	t.Run("check inclusion estimation fails should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		currentBlockHeader := &block.HeaderV2{
			Header: &block.Header{},
		}
		_ = dataComponents.BlockChain.SetCurrentBlockHeaderAndRootHash(currentBlockHeader, []byte("root"))
		dataComponents.BlockChain.SetCurrentBlockHeaderHash([]byte("hash"))
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ExecutionResultsVerifier = &processMocks.ExecutionResultsVerifierMock{
			VerifyHeaderExecutionResultsCalled: func(header data.HeaderHandler) error {
				return nil
			},
		}
		arguments.ExecutionResultsInclusionEstimator = &processMocks.InclusionEstimatorMock{
			DecideCalled: func(lastNotarised *estimator.LastExecutionResultForInclusion, pending []data.BaseExecutionResultHandler, currentHdrTsMs uint64) (allowed int) {
				return 10
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		body := &block.Body{}

		header := &block.HeaderV3{
			PrevHash:         []byte("hash"),
			Nonce:            1,
			Round:            2,
			MiniBlockHeaders: []block.MiniBlockHeader{},
		}
		err = sp.VerifyBlockProposal(header, body, haveTime)
		require.Equal(t, process.ErrInvalidNumberOfExecutionResultsInHeader, err)
	})

	t.Run("request missing meta headers fails should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		currentBlockHeader := &block.HeaderV2{
			Header: &block.Header{},
		}
		_ = dataComponents.BlockChain.SetCurrentBlockHeaderAndRootHash(currentBlockHeader, []byte("root"))
		dataComponents.BlockChain.SetCurrentBlockHeaderHash([]byte("hash"))
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ExecutionResultsVerifier = &processMocks.ExecutionResultsVerifierMock{
			VerifyHeaderExecutionResultsCalled: func(header data.HeaderHandler) error {
				return nil
			},
		}
		arguments.ExecutionResultsInclusionEstimator = &processMocks.InclusionEstimatorMock{
			DecideCalled: func(lastNotarised *estimator.LastExecutionResultForInclusion, pending []data.BaseExecutionResultHandler, currentHdrTsMs uint64) (allowed int) {
				return 0
			},
		}

		arguments.MissingDataResolver = &processMocks.MissingDataResolverMock{
			RequestMissingMetaHeadersCalled: func(shardHeader data.ShardHeaderHandler) error {
				return localErr
			},
		}

		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		body := &block.Body{}

		header := &block.HeaderV3{
			PrevHash:         []byte("hash"),
			Nonce:            1,
			Round:            2,
			MiniBlockHeaders: []block.MiniBlockHeader{},
		}
		err = sp.VerifyBlockProposal(header, body, haveTime)
		require.Equal(t, localErr, err)
	})

	t.Run("wait for missing data fails should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		currentBlockHeader := &block.HeaderV2{
			Header: &block.Header{},
		}
		_ = dataComponents.BlockChain.SetCurrentBlockHeaderAndRootHash(currentBlockHeader, []byte("root"))
		dataComponents.BlockChain.SetCurrentBlockHeaderHash([]byte("hash"))
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ExecutionResultsVerifier = &processMocks.ExecutionResultsVerifierMock{
			VerifyHeaderExecutionResultsCalled: func(header data.HeaderHandler) error {
				return nil
			},
		}
		arguments.ExecutionResultsInclusionEstimator = &processMocks.InclusionEstimatorMock{
			DecideCalled: func(lastNotarised *estimator.LastExecutionResultForInclusion, pending []data.BaseExecutionResultHandler, currentHdrTsMs uint64) (allowed int) {
				return 0
			},
		}

		arguments.MissingDataResolver = &processMocks.MissingDataResolverMock{
			RequestMissingMetaHeadersCalled: func(shardHeader data.ShardHeaderHandler) error {
				return nil
			},
			WaitForMissingDataCalled: func(timeout time.Duration) error {
				return localErr
			},
		}

		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		body := &block.Body{}

		header := &block.HeaderV3{
			PrevHash:         []byte("hash"),
			Nonce:            1,
			Round:            2,
			MiniBlockHeaders: []block.MiniBlockHeader{},
		}
		err = sp.VerifyBlockProposal(header, body, haveTime)
		require.Equal(t, localErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		currentBlockHeader := &block.HeaderV2{
			Header: &block.Header{},
		}
		_ = dataComponents.BlockChain.SetCurrentBlockHeaderAndRootHash(currentBlockHeader, []byte("root"))
		dataComponents.BlockChain.SetCurrentBlockHeaderHash([]byte("hash"))
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ExecutionResultsVerifier = &processMocks.ExecutionResultsVerifierMock{
			VerifyHeaderExecutionResultsCalled: func(header data.HeaderHandler) error {
				return nil
			},
		}
		arguments.ExecutionResultsInclusionEstimator = &processMocks.InclusionEstimatorMock{
			DecideCalled: func(lastNotarised *estimator.LastExecutionResultForInclusion, pending []data.BaseExecutionResultHandler, currentHdrTsMs uint64) (allowed int) {
				return 0
			},
		}

		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		body := &block.Body{}

		header := &block.HeaderV3{
			PrevHash:         []byte("hash"),
			Nonce:            1,
			Round:            2,
			MiniBlockHeaders: []block.MiniBlockHeader{},
		}
		err = sp.VerifyBlockProposal(header, body, haveTime)
		require.NoError(t, err)
	})
}

func TestShardProcessor_CheckInclusionEstimationForExecutionResults(t *testing.T) {
	t.Parallel()

	t.Run("cannot get prev block last execution results should error", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{}
		err := sp.CheckInclusionEstimationForExecutionResults(header)
		require.Equal(t, process.ErrNilHeaderHandler, err)
	})

	t.Run("invalid number of execution results", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		currentBlockHeader := &block.HeaderV2{
			Header: &block.Header{},
		}
		_ = dataComponents.BlockChain.SetCurrentBlockHeaderAndRootHash(currentBlockHeader, []byte("root"))
		dataComponents.BlockChain.SetCurrentBlockHeaderHash([]byte("hash"))
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		arguments.ExecutionResultsInclusionEstimator = &processMocks.InclusionEstimatorMock{
			DecideCalled: func(lastNotarised *estimator.LastExecutionResultForInclusion, pending []data.BaseExecutionResultHandler, currentHdrTsMs uint64) (allowed int) {
				return 1
			},
		}

		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{}
		err := sp.CheckInclusionEstimationForExecutionResults(header)
		require.Equal(t, process.ErrInvalidNumberOfExecutionResultsInHeader, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		currentBlockHeader := &block.HeaderV2{
			Header: &block.Header{},
		}
		_ = dataComponents.BlockChain.SetCurrentBlockHeaderAndRootHash(currentBlockHeader, []byte("root"))
		dataComponents.BlockChain.SetCurrentBlockHeaderHash([]byte("hash"))
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{}
		err := sp.CheckInclusionEstimationForExecutionResults(header)
		require.Equal(t, nil, err)
	})
}

func TestShardProcessor_CheckMetaHeadersValidityAndFinalityProposal(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local error")

	t.Run("cannot get last notarized header should err", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		arguments.BlockTracker = &mock.BlockTrackerMock{
			GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
				return nil, nil, localErr
			},
		}

		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{}
		err := sp.CheckMetaHeadersValidityAndFinalityProposal(header)
		require.Equal(t, localErr, err)
	})

	t.Run("cannot find used meta header should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		metaHeader := &block.MetaBlockV3{}
		arguments.BlockTracker = &mock.BlockTrackerMock{
			GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
				return metaHeader, []byte("h"), nil
			},
		}

		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{
			MetaBlockHashes: [][]byte{[]byte("hh")},
		}
		err := sp.CheckMetaHeadersValidityAndFinalityProposal(header)
		require.NotNil(t, err)
		require.ErrorContains(t, err, process.ErrMissingHeader.Error())
	})

	t.Run("invalid header should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		metaHeader := &block.MetaBlockV3{}
		arguments.BlockTracker = &mock.BlockTrackerMock{
			GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
				return metaHeader, []byte("h"), nil
			},
		}
		arguments.HeaderValidator = &processMocks.HeaderValidatorMock{
			IsHeaderConstructionValidCalled: func(currHdr, prevHdr data.HeaderHandler) error {
				return localErr
			},
		}

		dataPool, ok := dataComponents.Datapool().(*dataRetriever.PoolsHolderStub)
		require.True(t, ok)

		dataPool.HeadersCalled = func() retriever.HeadersPool {
			return &pool.HeadersPoolStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					return &block.Header{}, nil
				},
			}
		}

		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{
			MetaBlockHashes: [][]byte{[]byte("hh")},
		}
		err := sp.CheckMetaHeadersValidityAndFinalityProposal(header)
		require.NotNil(t, err)
		require.ErrorContains(t, err, localErr.Error())
	})

	t.Run("missing proof should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		metaHeader := &block.MetaBlockV3{}

		arguments.HeaderValidator = &processMocks.HeaderValidatorMock{
			IsHeaderConstructionValidCalled: func(currHdr, prevHdr data.HeaderHandler) error {
				return nil
			},
		}

		arguments.BlockTracker = &mock.BlockTrackerMock{
			GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
				return metaHeader, []byte("h"), nil
			},
		}

		dataPool, ok := dataComponents.Datapool().(*dataRetriever.PoolsHolderStub)
		require.True(t, ok)

		dataPool.HeadersCalled = func() retriever.HeadersPool {
			return &pool.HeadersPoolStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					return &block.Header{}, nil
				},
			}
		}

		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{
			MetaBlockHashes: [][]byte{[]byte("hh")},
		}
		err := sp.CheckMetaHeadersValidityAndFinalityProposal(header)
		require.NotNil(t, err)
		require.ErrorContains(t, err, process.ErrHeaderNotFinal.Error())
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		metaHeader := &block.MetaBlockV3{}

		arguments.HeaderValidator = &processMocks.HeaderValidatorMock{
			IsHeaderConstructionValidCalled: func(currHdr, prevHdr data.HeaderHandler) error {
				return nil
			},
		}

		arguments.BlockTracker = &mock.BlockTrackerMock{
			GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
				return metaHeader, []byte("h"), nil
			},
		}

		dataPool, ok := dataComponents.Datapool().(*dataRetriever.PoolsHolderStub)
		require.True(t, ok)

		dataPool.HeadersCalled = func() retriever.HeadersPool {
			return &pool.HeadersPoolStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					return &block.Header{}, nil
				},
			}
		}
		dataPool.ProofsCalled = func() retriever.ProofsPool {
			return &dataRetriever.ProofsPoolMock{
				HasProofCalled: func(shardID uint32, headerHash []byte) bool {
					return true
				},
			}
		}

		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{
			MetaBlockHashes: [][]byte{[]byte("hh")},
		}
		err := sp.CheckMetaHeadersValidityAndFinalityProposal(header)
		require.Nil(t, err)
	})
}

func checkCreateBlockProposalResult(
	t *testing.T,
	sp shardProcessorTest,
	header data.HeaderHandler,
	haveTime func() bool,
	expectedError error,
) {
	hdr, body, err := sp.CreateBlockProposal(header, haveTime)
	require.Equal(t, expectedError, err)
	require.Nil(t, hdr)
	require.Nil(t, body)
}
