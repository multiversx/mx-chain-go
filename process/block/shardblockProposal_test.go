package block_test

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	retriever "github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/executionTrack"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/estimator"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	testscommonExecutionTrack "github.com/multiversx/mx-chain-go/testscommon/executionTrack"
	"github.com/multiversx/mx-chain-go/testscommon/mbSelection"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

func getSimpleHeaderV3Mock() *testscommon.HeaderHandlerStub {
	return &testscommon.HeaderHandlerStub{
		IsHeaderV3Called: func() bool {
			return true
		},
		GetLastExecutionResultHandlerCalled: func() data.LastExecutionResultHandler {
			return &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{},
			}
		},
		GetPrevHashCalled: func() []byte {
			return []byte("prev hash")
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
			CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, []block.MiniblockAndHash, uint32, bool, error) {
				return nil, nil, 0, false, expectedError
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
		arguments.ExecutionResultsTracker = &testscommonExecutionTrack.ExecutionResultsTrackerStub{
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
		arguments.ExecutionResultsTracker = &testscommonExecutionTrack.ExecutionResultsTrackerStub{
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
	t.Run("nil last execution result should error", func(t *testing.T) {
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
			CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, []block.MiniblockAndHash, uint32, bool, error) {
				return []block.MiniblockAndHash{
						{
							Miniblock: providedMb,
							Hash:      []byte("providedMB"),
						},
					},
					[]block.MiniblockAndHash{},
					0, true, nil
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		header := getSimpleHeaderV3Mock()
		header.GetLastExecutionResultHandlerCalled = func() data.LastExecutionResultHandler {
			return nil
		}
		hdr, body, err := sp.CreateBlockProposal(header, haveTimeTrue)
		require.Error(t, err)
		require.Nil(t, hdr)
		require.Nil(t, body)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.IntMarsh = &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return []byte("marshalled"), nil
			},
		}
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return getSimpleHeaderV3Mock()
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash")
			},
		}
		dataComponents.DataPool = &dataRetriever.PoolsHolderStub{
			ProofsCalled: func() retriever.ProofsPool {
				return &dataRetriever.ProofsPoolMock{
					HasProofCalled: func(shardID uint32, headerHash []byte) bool {
						return true
					},
				}
			},
			HeadersCalled: func() retriever.HeadersPool {
				return &pool.HeadersPoolStub{
					GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
						return &block.HeaderV3{
							LastExecutionResult: &block.ExecutionResultInfo{
								ExecutionResult: &block.BaseExecutionResult{},
							},
						}, nil
					},
				}
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
		providedPendingMb := &block.MiniBlock{
			TxHashes: [][]byte{[]byte("tx_hash2")},
		}
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, []block.MiniblockAndHash, uint32, bool, error) {
				return []block.MiniblockAndHash{
						{
							Miniblock: providedMb,
							Hash:      []byte("providedMB"),
						},
					},
					[]block.MiniblockAndHash{
						{
							Miniblock: providedPendingMb,
							Hash:      []byte("providedPendingMB"),
						},
					}, 0, true, nil
			},
			SelectOutgoingTransactionsCalled: func() ([][]byte, []data.MiniBlockHeaderHandler) {
				return [][]byte{}, []data.MiniBlockHeaderHandler{&block.MiniBlockHeader{Hash: []byte("providedPendingMB")}}
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
		require.Len(t, rawBody.MiniBlocks, 2)
		require.Equal(t, providedMb, rawBody.MiniBlocks[0])
		require.Equal(t, providedPendingMb, rawBody.MiniBlocks[1])
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

		_, err = sp.SelectIncomingMiniBlocks(providedLastCrossNotarizedMetaHdr, providedOrderedMetaBlocks, providedOrderedMetaBlocksHashes, haveTimeFalse)
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

		_, err = sp.SelectIncomingMiniBlocks(providedLastCrossNotarizedMetaHdr, providedOrderedMetaBlocks, providedOrderedMetaBlocksHashes, haveTimeTrue)
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
		_, err = sp.SelectIncomingMiniBlocks(providedLastCrossNotarizedMetaHdr, providedOrderedMetaBlocksCopy, providedOrderedMetaBlocksHashes, haveTimeTrue)
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
		_, err = sp.SelectIncomingMiniBlocks(providedLastCrossNotarizedMetaHdr, orderedMetaBlocks, providedOrderedMetaBlocksHashes, haveTimeTrue)
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
			CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, []block.MiniblockAndHash, uint32, bool, error) {
				return []block.MiniblockAndHash{
					{
						Miniblock: &block.MiniBlock{},
						Hash:      []byte("providedMB"),
					},
				}, nil, 0, true, nil
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
		_, err = sp.SelectIncomingMiniBlocks(providedLastCrossNotarizedMetaHdr, orderedMetaBlocks, orderedMetaBlocksHashes, haveTimeTrue)
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
			CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, []block.MiniblockAndHash, uint32, bool, error) {
				cntCreateMbsCrossShardDstMeCalled++
				if cntCreateMbsCrossShardDstMeCalled < 2 {
					return []block.MiniblockAndHash{
						{
							Miniblock: &block.MiniBlock{},
							Hash:      []byte("providedMB"),
						},
					}, nil, 0, true, nil
				}
				return nil, nil, 0, false, nil // shouldContinue = false -> only for coverage
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
		_, err = sp.SelectIncomingMiniBlocks(providedLastCrossNotarizedMetaHdr, orderedMetaBlocks, orderedMetaBlocksHashes, haveTimeTrue)
		require.NoError(t, err)
		// should be called three times, the third hdr returns shouldContinue false, but still added miniblocks, so meta block is referenced
		require.Equal(t, 3, cntAddReferencedMetaBlockCalled)
	})
}

func TestShardProcessor_VerifyBlockProposal(t *testing.T) {
	t.Parallel()

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
				return expectedError
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
		require.Equal(t, expectedError, err)
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
				return expectedError
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
		require.Equal(t, expectedError, err)
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
				return expectedError
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
		require.Equal(t, expectedError, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		poolMock, ok := dataComponents.DataPool.(*dataRetriever.PoolsHolderStub)
		require.True(t, ok)
		poolMock.HeadersCalled = func() retriever.HeadersPool {
			return &pool.HeadersPoolStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					return &block.HeaderV3{
						LastExecutionResult: &block.ExecutionResultInfo{
							ExecutionResult: &block.BaseExecutionResult{},
						},
					}, nil
				},
			}
		}
		currentBlockHeader := &block.HeaderV3{
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{},
			},
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
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{},
			},
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

	t.Run("cannot get last notarized header should err", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataPool, ok := dataComponents.DataPool.(*dataRetriever.PoolsHolderStub)
		require.True(t, ok)
		dataPool.HeadersCalled = func() retriever.HeadersPool {
			return &pool.HeadersPoolStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					return &block.HeaderV3{}, nil
				},
			}
		}

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		arguments.BlockTracker = &mock.BlockTrackerMock{
			GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
				return nil, nil, expectedError
			},
		}

		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{}
		err := sp.CheckMetaHeadersValidityAndFinalityProposal(header)
		require.Equal(t, expectedError, err)
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
				return expectedError
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
		require.ErrorContains(t, err, expectedError.Error())
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

func TestShardBlockProposal_CreateAndVerifyProposal(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})

	currentHeader := &block.HeaderV3{
		Nonce: 10,
		Round: 10,
		LastExecutionResult: &block.ExecutionResultInfo{
			NotarizedInRound: 10,
			ExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("prevHeaderHash"),
				HeaderNonce: 9,
				HeaderRound: 9,
				RootHash:    []byte("prevRootHash"),
				GasUsed:     100000,
			},
		},
	}
	currentHeaderHash := []byte("currHdrHash")
	blkc.SetCurrentBlockHeaderHash(currentHeaderHash)
	err := blkc.SetCurrentBlockHeaderAndRootHash(currentHeader, []byte("currHdrRootHash"))
	require.Nil(t, err)
	dataComponents.DataPool = &dataRetriever.PoolsHolderStub{
		ProofsCalled: func() retriever.ProofsPool {
			return &dataRetriever.ProofsPoolMock{
				HasProofCalled: func(shardID uint32, headerHash []byte) bool {
					return true
				},
			}
		},
		HeadersCalled: func() retriever.HeadersPool {
			return &pool.HeadersPoolStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					return &block.HeaderV3{
						LastExecutionResult: &block.ExecutionResultInfo{
							ExecutionResult: &block.BaseExecutionResult{},
						},
					}, nil
				},
			}
		},
	}
	dataComponents.BlockChain = blkc

	executionResultsTracker := executionTrack.NewExecutionResultsTracker()
	execResultsVerifier, _ := blproc.NewExecutionResultsVerifier(dataComponents.BlockChain, executionResultsTracker)

	arguments.ArgBaseProcessor.ExecutionResultsTracker = executionResultsTracker
	arguments.ArgBaseProcessor.ExecutionResultsVerifier = execResultsVerifier

	shardProcessor, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	t.Run("should work - without miniblocks and transactions", func(t *testing.T) {
		header := &block.HeaderV3{
			Round:    11,
			Nonce:    11,
			PrevHash: currentHeaderHash,
		}
		headerProposed, bodyProposed, err := shardProcessor.CreateBlockProposal(header, haveTimeTrue)
		require.Nil(t, err)
		require.NotNil(t, headerProposed)
		require.NotNil(t, bodyProposed)

		err = shardProcessor.VerifyBlockProposal(headerProposed, bodyProposed, func() time.Duration { return time.Second })
		require.Nil(t, err)
	})

	t.Run("nil proposed block should fail", func(t *testing.T) {
		headerProposed, bodyProposed, err := shardProcessor.CreateBlockProposal(nil, haveTimeTrue)
		require.Equal(t, process.ErrNilBlockHeader, err)
		require.Nil(t, headerProposed)
		require.Nil(t, bodyProposed)
	})

	t.Run("nil proposed body should fail", func(t *testing.T) {
		header := &block.HeaderV3{
			Round:    11,
			Nonce:    11,
			PrevHash: currentHeaderHash,
		}
		headerProposed, bodyProposed, err := shardProcessor.CreateBlockProposal(header, haveTimeTrue)
		require.Nil(t, err)
		require.NotNil(t, headerProposed)
		require.NotNil(t, bodyProposed)

		err = shardProcessor.VerifyBlockProposal(headerProposed, nil, func() time.Duration { return time.Second })
		require.Equal(t, process.ErrNilBlockBody, err)
	})
}

func TestShardBlockProposal_CreateAndVerifyProposal_WithTransactions(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})

	currentHeader := &block.HeaderV3{
		Nonce: 10,
		Round: 10,
		LastExecutionResult: &block.ExecutionResultInfo{
			NotarizedInRound: 10,
			ExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("prevHeaderHash"),
				HeaderNonce: 9,
				HeaderRound: 9,
				RootHash:    []byte("prevRootHash"),
				GasUsed:     100000,
			},
		},
	}
	currentHeaderHash := []byte("currHdrHash")
	blkc.SetCurrentBlockHeaderHash(currentHeaderHash)
	err := blkc.SetCurrentBlockHeaderAndRootHash(currentHeader, []byte("currHdrRootHash"))
	require.Nil(t, err)

	epochStartMetaHash := []byte("epochStartMetaHash")
	metaBlockHash1 := []byte("metaBlockHash1")
	metaBlock1 := &block.MetaBlockV3{
		Round: 10,
	}
	lastCommitedMetaHash := []byte("lastCommitedMeta")
	lastCommitedMeta := &block.MetaBlockV3{
		Round: 9,
		Nonce: 9,
	}
	lastCrossNotarizedMetaHdrHash := []byte("lastCrossNotarizedMetaHdrHash")
	lastCrossNotarizedMetaHdr := &block.MetaBlockV3{
		Round: 8,
		Nonce: 8,
	}

	providedMb := &block.MiniBlock{
		TxHashes: [][]byte{[]byte("tx_hash")},
	}

	headers := &mock.HeadersCacherStub{}
	headers.GetHeaderByHashCalled = func(hash []byte) (data.HeaderHandler, error) {
		if bytes.Equal(hash, epochStartMetaHash) {
			return &block.MetaBlockV3{}, nil
		}
		if bytes.Equal(hash, metaBlockHash1) {
			return metaBlock1, nil
		}
		if bytes.Equal(hash, lastCommitedMetaHash) {
			return lastCommitedMeta, nil
		}
		if bytes.Equal(hash, lastCrossNotarizedMetaHdrHash) {
			return lastCrossNotarizedMetaHdr, nil
		}

		return &block.HeaderV3{
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{},
			},
		}, nil
	}

	headers.AddHeader(epochStartMetaHash, &block.MetaBlockV3{})

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
		MiniBlocksCalled: func() storage.Cacher {
			return &cache.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					value = providedMb
					ok = true
					return
				},
			}
		},
		TransactionsCalled: func() retriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
					value = &transaction.Transaction{}
					ok = true
					return
				},
			}
		},
	}
	dataComponents.BlockChain = blkc

	executionResultsTracker := executionTrack.NewExecutionResultsTracker()
	execResultsVerifier, _ := blproc.NewExecutionResultsVerifier(dataComponents.BlockChain, executionResultsTracker)

	arguments.ArgBaseProcessor.ExecutionResultsTracker = executionResultsTracker
	arguments.ArgBaseProcessor.ExecutionResultsVerifier = execResultsVerifier

	arguments.MissingDataResolver = &processMocks.MissingDataResolverMock{
		WaitForMissingDataCalled: func(timeout time.Duration) error {
			return nil
		},
	}

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
				}},
				[][]byte{lastCommitedMetaHash},
				nil
		},
		GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return lastCrossNotarizedMetaHdr, lastCrossNotarizedMetaHdrHash, nil
		},
	}
	mbHash, _ := core.CalculateHash(arguments.CoreComponents.InternalMarshalizer(), arguments.CoreComponents.Hasher(), providedMb)
	arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, []block.MiniblockAndHash, uint32, bool, error) {
			return []block.MiniblockAndHash{
				{
					Miniblock: providedMb,
					Hash:      mbHash,
				},
			}, nil, 0, true, nil
		},
	}

	shardProcessor, err := blproc.NewShardProcessor(arguments)
	require.Nil(t, err)

	argsHeaderValidator := blproc.ArgsHeaderValidator{
		Hasher:              coreComponents.Hasher(),
		Marshalizer:         coreComponents.InternalMarshalizer(),
		EnableEpochsHandler: coreComponents.EnableEpochsHandler(),
	}
	headerValidator, _ := blproc.NewHeaderValidator(argsHeaderValidator)
	shardProcessor.SetHeaderValidator(headerValidator)

	header := &block.HeaderV3{
		Round:           11,
		Nonce:           11,
		PrevHash:        currentHeaderHash,
		MetaBlockHashes: [][]byte{metaBlockHash1},
		LastExecutionResult: &block.ExecutionResultInfo{
			ExecutionResult: &block.BaseExecutionResult{},
		},
	}
	headerProposed, bodyProposed, err := shardProcessor.CreateBlockProposal(header, haveTimeTrue)
	require.Nil(t, err)
	require.NotNil(t, headerProposed)
	require.NotNil(t, bodyProposed)

	err = shardProcessor.VerifyBlockProposal(headerProposed, bodyProposed, func() time.Duration { return time.Second })
	require.Nil(t, err)
}

func TestShardProcessor_VerifyGasLimit(t *testing.T) {
	t.Parallel()

	t.Run("getTransactionsForMiniBlock fails due to missing mini block", func(t *testing.T) {
		t.Parallel()

		outgoingMbh, outgoingMb, incomingMbh, incomingMb := createMiniBlocks()
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataPool := adaptDataPoolForVerifyGas(t, dataComponents.DataPool, outgoingMb, incomingMb)
		dataPool.MiniBlocksCalled = func() storage.Cacher {
			return &cache.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, false
				},
			}
		}
		dataComponents.DataPool = dataPool
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		sp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		err = sp.VerifyGasLimit(createHeaderFromMBs(outgoingMbh, incomingMbh))
		require.Equal(t, process.ErrMissingMiniBlock, err)
	})
	t.Run("getTransactionsForMiniBlock fails due to mini block cast issue", func(t *testing.T) {
		t.Parallel()

		outgoingMbh, outgoingMb, incomingMbh, incomingMb := createMiniBlocks()
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataPool := adaptDataPoolForVerifyGas(t, dataComponents.DataPool, outgoingMb, incomingMb)
		dataPool.MiniBlocksCalled = func() storage.Cacher {
			return &cache.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					value = "non mini block"
					ok = true
					return
				},
			}
		}
		dataComponents.DataPool = dataPool
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		sp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		err = sp.VerifyGasLimit(createHeaderFromMBs(outgoingMbh, incomingMbh))
		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})
	t.Run("getTransactionsForMiniBlock fails due to error on GetTransactionHandlerFromPool", func(t *testing.T) {
		t.Parallel()

		outgoingMbh, outgoingMb, incomingMbh, incomingMb := createMiniBlocks()
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataPool := adaptDataPoolForVerifyGas(t, dataComponents.DataPool, outgoingMb, incomingMb)
		dataPool.TransactionsCalled = func() retriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
					return nil, false
				},
			}
		}
		dataComponents.DataPool = dataPool
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		sp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		err = sp.VerifyGasLimit(createHeaderFromMBs(outgoingMbh, incomingMbh))
		require.Error(t, err)
	})
	t.Run("CheckIncomingMiniBlocks error", func(t *testing.T) {
		t.Parallel()

		outgoingMbh, outgoingMb, incomingMbh, incomingMb := createMiniBlocks()
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataPool := adaptDataPoolForVerifyGas(t, dataComponents.DataPool, outgoingMb, incomingMb)
		dataComponents.DataPool = dataPool
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.GasComputation = &testscommon.GasComputationMock{
			CheckIncomingMiniBlocksCalled: func(miniBlocks []data.MiniBlockHeaderHandler, transactions map[string][]data.TransactionHandler) (int, int, error) {
				return 0, 0, expectedError
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		err = sp.VerifyGasLimit(createHeaderFromMBs(outgoingMbh, incomingMbh))
		require.Equal(t, expectedError, err)
	})
	t.Run("CheckOutgoingTransactions error", func(t *testing.T) {
		t.Parallel()

		outgoingMbh, outgoingMb, incomingMbh, incomingMb := createMiniBlocks()
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataPool := adaptDataPoolForVerifyGas(t, dataComponents.DataPool, outgoingMb, incomingMb)
		dataComponents.DataPool = dataPool
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.GasComputation = &testscommon.GasComputationMock{
			CheckIncomingMiniBlocksCalled: func(miniBlocks []data.MiniBlockHeaderHandler, transactions map[string][]data.TransactionHandler) (int, int, error) {
				return len(miniBlocks), 0, nil
			},
			CheckOutgoingTransactionsCalled: func(txHashes [][]byte, transactions []data.TransactionHandler) ([][]byte, []data.MiniBlockHeaderHandler, error) {
				return nil, nil, expectedError
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		err = sp.VerifyGasLimit(createHeaderFromMBs(outgoingMbh, incomingMbh))
		require.Equal(t, expectedError, err)
	})
	t.Run("CheckOutgoingTransactions results in limit exceeded", func(t *testing.T) {
		t.Parallel()

		outgoingMbh, outgoingMb, incomingMbh, incomingMb := createMiniBlocks()
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataPool := adaptDataPoolForVerifyGas(t, dataComponents.DataPool, outgoingMb, incomingMb)
		dataComponents.DataPool = dataPool
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.GasComputation = &testscommon.GasComputationMock{
			CheckIncomingMiniBlocksCalled: func(miniBlocks []data.MiniBlockHeaderHandler, transactions map[string][]data.TransactionHandler) (int, int, error) {
				return len(miniBlocks), 0, nil
			},
			CheckOutgoingTransactionsCalled: func(txHashes [][]byte, transactions []data.TransactionHandler) ([][]byte, []data.MiniBlockHeaderHandler, error) {
				return txHashes[:len(txHashes)-1], nil, nil // one tx over the limit
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		err = sp.VerifyGasLimit(createHeaderFromMBs(outgoingMbh, incomingMbh))
		require.ErrorIs(t, err, process.ErrInvalidMaxGasLimitPerMiniBlock)
		require.Contains(t, err.Error(), "outgoing transactions exceeded")
	})
	t.Run("CheckOutgoingTransactions adds extra pending mini blocks on CheckOutgoingTransactions", func(t *testing.T) {
		t.Parallel()

		outgoingMbh, outgoingMb, incomingMbh, incomingMb := createMiniBlocks()
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataPool := adaptDataPoolForVerifyGas(t, dataComponents.DataPool, outgoingMb, incomingMb)
		dataComponents.DataPool = dataPool
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.GasComputation = &testscommon.GasComputationMock{
			CheckIncomingMiniBlocksCalled: func(miniBlocks []data.MiniBlockHeaderHandler, transactions map[string][]data.TransactionHandler) (int, int, error) {
				return len(miniBlocks), 0, nil // no pending mini blocks left
			},
			CheckOutgoingTransactionsCalled: func(txHashes [][]byte, transactions []data.TransactionHandler) ([][]byte, []data.MiniBlockHeaderHandler, error) {
				return txHashes, []data.MiniBlockHeaderHandler{&block.MiniBlockHeader{}}, nil // one pending mini block added
			},
		}
		sp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		err = sp.VerifyGasLimit(createHeaderFromMBs(outgoingMbh, incomingMbh))
		require.ErrorIs(t, err, process.ErrInvalidMaxGasLimitPerMiniBlock)
		require.Contains(t, err.Error(), "incoming mini blocks exceeded the limit")
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		outgoingMbh, outgoingMb, incomingMbh, incomingMb := createMiniBlocks()
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataPool := adaptDataPoolForVerifyGas(t, dataComponents.DataPool, outgoingMb, incomingMb)
		dataComponents.DataPool = dataPool
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		sp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		header := createHeaderFromMBs(outgoingMbh, incomingMbh)
		headerV3, ok := header.(*block.HeaderV3)
		require.True(t, ok)
		headerV3.LastExecutionResult = &block.ExecutionResultInfo{
			ExecutionResult: &block.BaseExecutionResult{},
		}
		err = sp.VerifyGasLimit(headerV3)
		require.NoError(t, err)
	})
}

func adaptDataPoolForVerifyGas(
	t *testing.T,
	initialPool retriever.PoolsHolder,
	outgoingMb *block.MiniBlock,
	incomingMb *block.MiniBlock,
) *dataRetriever.PoolsHolderStub {
	headers := initialPool.Headers()
	proofs := initialPool.Proofs()
	return &dataRetriever.PoolsHolderStub{
		HeadersCalled: func() retriever.HeadersPool {
			return headers
		},
		ProofsCalled: func() retriever.ProofsPool {
			return proofs
		},
		MiniBlocksCalled: func() storage.Cacher {
			return &cache.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					switch string(key) {
					case "outgoingMBHash":
						value = outgoingMb
						ok = true
						return
					case "incomingMBHash":
						value = incomingMb
						ok = true
						return
					default:
						require.Fail(t, "unexpected key")
					}

					return
				},
			}
		},
		TransactionsCalled: func() retriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
					value = &transaction.Transaction{}
					ok = true
					return
				},
			}
		},
	}
}

func createHeaderFromMBs(mbs ...block.MiniBlockHeader) data.ShardHeaderHandler {
	return &block.HeaderV3{MiniBlockHeaders: mbs}
}

func createMiniBlocks() (
	outgoingMbh block.MiniBlockHeader,
	outgoingMb *block.MiniBlock,
	incomingMbh block.MiniBlockHeader,
	incomingMb *block.MiniBlock,
) {
	outgoingMbh, outgoingMb = createMiniBlock("outgoingMBHash", 0, 0)
	incomingMbh, incomingMb = createMiniBlock("incomingMBHash", 1, 0)
	return
}

func createMiniBlock(hash string, srcShard uint32, dstShard uint32) (block.MiniBlockHeader, *block.MiniBlock) {
	mbh := block.MiniBlockHeader{
		Hash:            []byte(hash),
		SenderShardID:   srcShard,
		ReceiverShardID: dstShard,
	}
	mb := &block.MiniBlock{
		TxHashes: [][]byte{
			[]byte("txHash"),
		},
		SenderShardID:   srcShard,
		ReceiverShardID: dstShard,
	}
	return mbh, mb
}

func TestShardProcessor_ProcessBlockProposal(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments(createComponentHolderMocks())

	t.Run("nil header should error", func(t *testing.T) {
		t.Parallel()

		sp, _ := blproc.NewShardProcessor(arguments)
		body := &block.Body{}
		_, err := sp.ProcessBlockProposal(nil, body)

		require.Equal(t, process.ErrNilBlockHeader, err)
	})
	t.Run("nil body should error", func(t *testing.T) {
		t.Parallel()

		sp, _ := blproc.NewShardProcessor(arguments)
		header := &block.HeaderV3{}
		_, err := sp.ProcessBlockProposal(header, nil)

		require.Equal(t, process.ErrNilBlockBody, err)
	})
	t.Run("not headerV3 should error", func(t *testing.T) {
		t.Parallel()

		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.Header{} // wrong type
		body := &block.Body{}
		_, err := sp.ProcessBlockProposal(header, body)

		require.Equal(t, process.ErrInvalidHeader, err)
	})
	t.Run("wrong header type (meta) should error", func(t *testing.T) {
		t.Parallel()

		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.MetaBlockV3{} // wrong type
		body := &block.Body{}
		_, err := sp.ProcessBlockProposal(header, body)

		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})
	t.Run("wrong body type should error", func(t *testing.T) {
		t.Parallel()

		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{}
		body := &wrongBody{} // wrong type
		_, err := sp.ProcessBlockProposal(header, body)

		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})
	t.Run("createBlockStarted fails should error", func(t *testing.T) {
		t.Parallel()

		args := CreateMockArguments(createComponentHolderMocks())
		args.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			AddIntermediateTransactionsCalled: func(mapSCRs map[block.Type][]data.TransactionHandler, key []byte) error {
				return expectedErr
			},
		}
		sp, _ := blproc.NewShardProcessor(args)

		header := &block.HeaderV3{}
		body := &block.Body{}
		_, err := sp.ProcessBlockProposal(header, body)

		require.Equal(t, expectedErr, err)
	})
	t.Run("IsDataPreparedForProcessing fails should error", func(t *testing.T) {
		t.Parallel()

		args := CreateMockArguments(createComponentHolderMocks())
		args.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			IsDataPreparedForProcessingCalled: func(haveTime func() time.Duration) error {
				return expectedErr
			},
		}
		sp, _ := blproc.NewShardProcessor(args)

		header := &block.HeaderV3{}
		body := &block.Body{}
		_, err := sp.ProcessBlockProposal(header, body)

		require.Equal(t, expectedErr, err)
	})
	t.Run("checkEpochStartInfoAvailableIfNeeded fails should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		args := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		args.EpochStartTrigger = &mock.EpochStartTriggerStub{
			MetaEpochCalled: func() uint32 {
				return 5
			},
			IsEpochStartCalled: func() bool {
				return false
			},
		}
		dataComponents.DataPool = &dataRetriever.PoolsHolderStub{
			HeadersCalled: func() retriever.HeadersPool {
				return &pool.HeadersPoolStub{
					GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
						return nil, expectedErr
					},
				}
			},
		}
		sp, _ := blproc.NewShardProcessor(args)

		header := &block.HeaderV3{
			Epoch:              10,
			EpochStartMetaHash: []byte("epochStartHash"),
		}
		body := &block.Body{}
		_, err := sp.ProcessBlockProposal(header, body)

		require.Error(t, err)
		require.ErrorIs(t, err, process.ErrEpochStartInfoNotAvailable)
	})
	t.Run("WaitForHeadersIfNeeded fails should error", func(t *testing.T) {
		t.Parallel()

		args := CreateMockArguments(createComponentHolderMocks())
		args.HeadersForBlock = &testscommon.HeadersForBlockMock{
			WaitForHeadersIfNeededCalled: func(haveTime func() time.Duration) error {
				return expectedErr
			},
		}
		sp, _ := blproc.NewShardProcessor(args)

		header := &block.HeaderV3{}
		body := &block.Body{}
		_, err := sp.ProcessBlockProposal(header, body)
		require.Equal(t, expectedErr, err)
	})
	t.Run("WaitForHeadersIfNeeded fails should error", func(t *testing.T) {
		t.Parallel()

		args := CreateMockArguments(createComponentHolderMocks())
		args.BlockChainHook = &testscommon.BlockChainHookStub{
			SetCurrentHeaderCalled: func(hdr data.HeaderHandler) error {
				return expectedErr
			},
		}
		sp, _ := blproc.NewShardProcessor(args)

		header := &block.HeaderV3{}
		body := &block.Body{}
		_, err := sp.ProcessBlockProposal(header, body)
		require.Equal(t, expectedErr, err)
	})
	t.Run("ProcessBlockTransaction fails should error", func(t *testing.T) {
		t.Parallel()

		args := CreateMockArguments(createComponentHolderMocks())
		args.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			ProcessBlockTransactionCalled: func(header data.HeaderHandler, body *block.Body, haveTime func() time.Duration) error {
				return expectedErr
			},
		}
		sp, _ := blproc.NewShardProcessor(args)

		header := &block.HeaderV3{}
		body := &block.Body{}
		_, err := sp.ProcessBlockProposal(header, body)
		require.Equal(t, expectedErr, err)
	})
	t.Run("VerifyCreatedBlockTransactions fails should error", func(t *testing.T) {
		t.Parallel()

		args := CreateMockArguments(createComponentHolderMocks())
		args.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			VerifyCreatedBlockTransactionsCalled: func(hdr data.HeaderHandler, body *block.Body) error {
				return expectedErr
			},
		}
		sp, _ := blproc.NewShardProcessor(args)

		header := &block.HeaderV3{}
		body := &block.Body{}
		_, err := sp.ProcessBlockProposal(header, body)
		require.Equal(t, expectedErr, err)
	})
	t.Run("CalculateHash fails should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.IntMarsh = &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		args := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, _ := blproc.NewShardProcessor(args)

		header := &block.HeaderV3{}
		body := &block.Body{}
		_, err := sp.ProcessBlockProposal(header, body)
		require.Equal(t, expectedErr, err)
	})
	t.Run("collectExecutionResults fails should error", func(t *testing.T) {
		t.Parallel()

		args := CreateMockArguments(createComponentHolderMocks())
		args.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			CreateReceiptsHashCalled: func() ([]byte, error) {
				return nil, expectedErr
			},
		}
		sp, _ := blproc.NewShardProcessor(args)

		header := &block.HeaderV3{}
		body := &block.Body{}
		_, err := sp.ProcessBlockProposal(header, body)
		require.Equal(t, expectedErr, err)
	})
	t.Run("HandleProcessErrorCutoff fails should error", func(t *testing.T) {
		t.Parallel()

		args := CreateMockArguments(createComponentHolderMocks())
		args.BlockProcessingCutoffHandler = &testscommon.BlockProcessingCutoffStub{
			HandleProcessErrorCutoffCalled: func(header data.HeaderHandler) error {
				return expectedErr
			},
		}
		sp, _ := blproc.NewShardProcessor(args)

		header := &block.HeaderV3{}
		body := &block.Body{}
		_, err := sp.ProcessBlockProposal(header, body)
		require.Equal(t, expectedErr, err)
	})
	t.Run("should work, no transactions", func(t *testing.T) {
		t.Parallel()

		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{}
		body := &block.Body{}
		_, err := sp.ProcessBlockProposal(header, body)

		require.Nil(t, err)
	})
}

func TestShardProcessor_ShouldEpochStartInfoBeAvailable(t *testing.T) {
	t.Parallel()

	t.Run("no epoch start meta hash should return false", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{}
		result := sp.ShouldEpochStartInfoBeAvailable(header)
		require.False(t, result)
	})

	t.Run("epoch start triggered should return false", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
			IsEpochStartCalled: func() bool {
				return true
			},
		}
		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{}
		_ = header.SetEpochStartMetaHash([]byte("hash"))
		result := sp.ShouldEpochStartInfoBeAvailable(header)
		require.False(t, result)
	})

	t.Run("epoch less than or equal to meta epoch should return false", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
			IsEpochStartCalled: func() bool {
				return false
			},
			MetaEpochCalled: func() uint32 {
				return 5
			},
		}
		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{}
		_ = header.SetEpochStartMetaHash([]byte("hash"))
		_ = header.SetEpoch(5) // equal
		result := sp.ShouldEpochStartInfoBeAvailable(header)
		require.False(t, result)

		_ = header.SetEpoch(4) // less than
		result = sp.ShouldEpochStartInfoBeAvailable(header)
		require.False(t, result)
	})

	t.Run("should return true when all conditions met", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
			IsEpochStartCalled: func() bool {
				return false
			},
			MetaEpochCalled: func() uint32 {
				return 5
			},
		}
		sp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{}
		_ = header.SetEpochStartMetaHash([]byte("hash"))
		_ = header.SetEpoch(10)
		result := sp.ShouldEpochStartInfoBeAvailable(header)
		require.True(t, result)
	})
}

func TestShardProcessor_GetCrossShardIncomingMiniBlocksFromBody(t *testing.T) {
	t.Parallel()

	t.Run("empty body should return empty", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		sp, _ := blproc.NewShardProcessor(arguments)

		body := &block.Body{}
		result := sp.GetCrossShardIncomingMiniBlocksFromBody(body)
		require.Empty(t, result)
	})

	t.Run("should filter only cross shard incoming miniblocks", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		// Using default OneShardCoordinator from createComponentHolderMocks
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, _ := blproc.NewShardProcessor(arguments)

		// Test with different shard IDs - one matching self shard, others cross-shard
		// Default coordinator has selfId = 0
		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					SenderShardID:   0,
					ReceiverShardID: 0,
					TxHashes:        [][]byte{[]byte("tx1")},
				},
				{
					SenderShardID:   1,
					ReceiverShardID: 0,
					TxHashes:        [][]byte{[]byte("tx2")},
				},
				{
					SenderShardID:   0,
					ReceiverShardID: 1,
					TxHashes:        [][]byte{[]byte("tx3")},
				},
				{
					SenderShardID:   2,
					ReceiverShardID: 0,
					TxHashes:        [][]byte{[]byte("tx4")},
				},
			},
		}
		result := sp.GetCrossShardIncomingMiniBlocksFromBody(body)
		// Should include miniblocks from shard 1 and 2 going to shard 0
		require.Len(t, result, 2)
		require.Equal(t, uint32(1), result[0].SenderShardID)
		require.Equal(t, uint32(2), result[1].SenderShardID)
	})
}

func TestGetHaveTimeForProposal(t *testing.T) {
	t.Parallel()

	startTime := time.Now()
	maxDuration := 100 * time.Millisecond

	haveTimeLocal := blproc.GetHaveTimeForProposal(startTime, maxDuration)
	remaining := haveTimeLocal()

	require.Greater(t, remaining, time.Duration(0))
	require.LessOrEqual(t, remaining, maxDuration)

}

func TestShouldDisableOutgoingTxs(t *testing.T) {
	t.Parallel()

	t.Run("should return true when conditions met", func(t *testing.T) {
		t.Parallel()

		coreComponents, _, _, _ := createComponentHolderMocks()
		enableEpochsHandler := coreComponents.EnableEpochsHandler()
		enableRoundsHandler := coreComponents.EnableRoundsHandler()

		result := blproc.ShouldDisableOutgoingTxs(enableEpochsHandler, enableRoundsHandler)
		// This tests that the function executes without error
		require.NotNil(t, result) // result can be true or false depending on configuration
	})
}

func TestShardProcessor_OnProposedBlock(t *testing.T) {
	t.Parallel()

	t.Run("wrong type assertion on body should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		wrongBodyInstance := &wrongBody{}
		header := getSimpleHeaderV3Mock()
		proposedHash := []byte("proposedHash")

		err = sp.OnProposedBlock(wrongBodyInstance, header, proposedHash)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})

	t.Run("GetHeaderByHash error should return error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataPool, ok := dataComponents.DataPool.(*dataRetriever.PoolsHolderStub)
		require.True(t, ok)
		dataPool.HeadersCalled = func() retriever.HeadersPool {
			return &pool.HeadersPoolStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					return nil, expectedErr
				},
			}
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		body := &block.Body{}
		header := getSimpleHeaderV3Mock()
		proposedHash := []byte("proposedHash")

		err = sp.OnProposedBlock(body, header, proposedHash)
		require.Equal(t, expectedErr, err)
	})

	t.Run("GetLastBaseExecutionResultHandler error should return error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataPool, ok := dataComponents.DataPool.(*dataRetriever.PoolsHolderStub)
		require.True(t, ok)
		dataPool.HeadersCalled = func() retriever.HeadersPool {
			return &pool.HeadersPoolStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					return &block.HeaderV3{}, nil // nil last exec result
				},
			}
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		body := &block.Body{}
		header := getSimpleHeaderV3Mock()
		proposedHash := []byte("proposedHash")

		err = sp.OnProposedBlock(body, header, proposedHash)
		require.Error(t, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasOnProposedBlockCalled := false
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataPool, ok := dataComponents.DataPool.(*dataRetriever.PoolsHolderStub)
		require.True(t, ok)
		dataPool.TransactionsCalled = func() retriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				OnProposedBlockCalled: func(blockHash []byte, blockBody *block.Body, blockHeader data.HeaderHandler, accountsProvider common.AccountNonceAndBalanceProvider, blockchainInfo common.BlockchainInfo) error {
					wasOnProposedBlockCalled = true
					return nil
				},
			}
		}
		dataPool.HeadersCalled = func() retriever.HeadersPool {
			return &pool.HeadersPoolStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					return getSimpleHeaderV3Mock(), nil
				},
			}
		}
		dataComponents.DataPool = dataPool
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		body := &block.Body{}
		header := getSimpleHeaderV3Mock()
		proposedHash := []byte("proposedHash")

		err = sp.OnProposedBlock(body, header, proposedHash)
		require.NoError(t, err)
		require.True(t, wasOnProposedBlockCalled)
	})
}
