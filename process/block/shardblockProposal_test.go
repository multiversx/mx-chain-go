package block_test

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
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
	"github.com/stretchr/testify/require"
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
			GetPendingExecutionResultsCalled: func() ([]data.ExecutionResultHandler, error) {
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
			DecideCalled: func(lastNotarised *estimator.LastExecutionResultForInclusion, pending []data.ExecutionResultHandler, currentHdrTsMs uint64) (allowed int) {
				return 1 // coverage only
			},
		}
		arguments.ExecutionResultsTracker = &executionTrack.ExecutionResultsTrackerStub{
			GetPendingExecutionResultsCalled: func() ([]data.ExecutionResultHandler, error) {
				return []data.ExecutionResultHandler{
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

		providedOrderedMetaBlocksCopy := providedOrderedMetaBlocks
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

// Test context isolation between proposal and processing flows
func TestShardProcessor_ProposalVsProcessingContextIsolation(t *testing.T) {
	t.Parallel()

	// This test verifies that the proposal flow doesn't interfere with processing flow state

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	dataComponents.BlockChain = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.HeaderV2{} // using V2 for simplicity
		},
		GetCurrentBlockHeaderHashCalled: func() []byte {
			return []byte("hash")
		},
	}

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
	hdr, body, err := sp.CreateBlockProposal(shardHdr, haveTimeTrue)
	require.Nil(t, err)
	require.NotNil(t, hdr)
	require.NotNil(t, body)

	// Verify that only proposal-related operations were called
	require.Contains(t, proposalStateModifications, "ComputeLongestMetaChainFromLastNotarized")
	require.Contains(t, proposalStateModifications, "GetLastCrossNotarizedHeader")
	require.Contains(t, proposalStateModifications, "SelectOutgoingTransactions")

	// Verify that processing-specific operations were NOT called during proposal
	require.NotContains(t, proposalStateModifications, "ProcessBlockTransaction")

	// The key isolation check: proposal flow should only read state, not modify processing state
	// This is ensured by the fact that CreateBlockProposal doesn't call processing methods
	require.Empty(t, processingStateModifications, "Processing operations should not be called during proposal creation")
}

// Test miniBlocksSelectionSession state isolation
func TestShardProcessor_MiniBlocksSelectionSessionIsolation(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	dataComponents.BlockChain = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.HeaderV2{} // using V2 for simplicity
		},
		GetCurrentBlockHeaderHashCalled: func() []byte {
			return []byte("hash")
		},
	}

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

	_, _, err = sp.CreateBlockProposal(shardHdr, haveTimeTrue)
	require.Nil(t, err)

	// Verify that the session is reset at the beginning - this ensures isolation
	// between different proposal creation calls
	require.True(t, resetCalled, "MiniBlocksSelectionSession should be reset to ensure isolation")
}

// Test time constraint handling
func TestShardProcessor_ProposalTimeConstraints(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	dataComponents.BlockChain = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.HeaderV2{} // using V2 for simplicity
		},
		GetCurrentBlockHeaderHashCalled: func() []byte {
			return []byte("hash")
		},
	}

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
	require.Less(t, elapsed, 100*time.Millisecond, "Should complete quickly when haveTime returns false")

	// Verify haveTime was called (indicating time constraints are checked)
	require.Greater(t, callCount, 0, "haveTime should be called to check time constraints")
}
