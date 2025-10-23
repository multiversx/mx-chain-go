package block_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/graceperiod"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/state"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	testscommonExecutionTrack "github.com/multiversx/mx-chain-go/testscommon/executionTrack"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mbSelection"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
	"github.com/multiversx/mx-chain-go/testscommon/round"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
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
		expectedError := errors.New("expected error")
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

func Test_addExecutionResultsOnHeader(t *testing.T) {
	t.Parallel()
	t.Run("executionResultsTracker returns error should error", func(t *testing.T) {
		t.Parallel()
		expectedErr := errors.New("expected error")
		sp, _ := blproc.ConstructPartialShardBlockProcessorForTest(map[string]interface{}{
			"executionResultsTracker": &testscommonExecutionTrack.ExecutionResultsTrackerStub{
				GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
					return nil, expectedErr
				},
			},
		})
		err := sp.AddExecutionResultsOnHeader(&block.HeaderV3{})
		require.Error(t, err)
		require.Equal(t, expectedErr, err)
	})
	t.Run("GetPrevBlockLastExecutionResult returns error should error", func(t *testing.T) {
		t.Parallel()
		sp, _ := blproc.ConstructPartialShardBlockProcessorForTest(map[string]interface{}{
			"executionResultsTracker": &testscommonExecutionTrack.ExecutionResultsTrackerStub{
				GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
					return []data.BaseExecutionResultHandler{
						&block.ExecutionResult{
							BaseExecutionResult: &block.BaseExecutionResult{},
						},
					}, nil
				},
			},
		})
		err := sp.AddExecutionResultsOnHeader(&block.HeaderV3{})
		require.Error(t, err)
		require.Equal(t, process.ErrNilBlockChain, err)
	})
	t.Run("CreateDataForInclusionEstimation returns error should error", func(t *testing.T) {
		t.Parallel()
		sp, _ := blproc.ConstructPartialShardBlockProcessorForTest(map[string]interface{}{
			"executionResultsTracker": &testscommonExecutionTrack.ExecutionResultsTrackerStub{
				GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
					return []data.BaseExecutionResultHandler{
						&block.ExecutionResult{
							BaseExecutionResult: &block.BaseExecutionResult{},
						},
					}, nil
				},
			},
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return []byte("hash")
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.HeaderV3{
						PrevHash: []byte("prev_hash"),
					}
				},
			},
		})
		err := sp.AddExecutionResultsOnHeader(&block.HeaderV3{})
		require.Error(t, err)
		require.Equal(t, process.ErrNilLastExecutionResultHandler, err)
	})

	t.Run("CreateLastExecutionResultInfoFromExecutionResult returns error should error", func(t *testing.T) {
		t.Parallel()

		baseExecutionResults := &block.BaseExecutionResult{
			HeaderHash:  []byte("hash"),
			HeaderNonce: 100,
			HeaderRound: 1,
			RootHash:    []byte("rootHash"),
		}
		header := &block.HeaderV3{
			PrevHash: []byte("prev_hash"),

			LastExecutionResult: &block.ExecutionResultInfo{
				NotarizedInRound: 1,
				ExecutionResult:  baseExecutionResults,
			},
		}

		sp, _ := blproc.ConstructPartialShardBlockProcessorForTest(map[string]interface{}{
			"executionResultsTracker": &testscommonExecutionTrack.ExecutionResultsTrackerStub{
				GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
					// return one meta execution result (so Decide can include it)
					meta := &block.MetaExecutionResult{
						ExecutionResult: &block.BaseMetaExecutionResult{
							BaseExecutionResult: &block.BaseExecutionResult{
								HeaderHash:  []byte("hdr-hash"),
								HeaderNonce: 1,
								HeaderRound: 2,
								RootHash:    []byte("root"),
								GasUsed:     100000,
							},
							ValidatorStatsRootHash: []byte("vstats"),
							AccumulatedFeesInEpoch: big.NewInt(123),
							DevFeesInEpoch:         big.NewInt(45),
						},
						ReceiptsHash:     []byte{},
						MiniBlockHeaders: nil,
						ExecutedTxCount:  0,
					}
					return []data.BaseExecutionResultHandler{meta}, nil
				},
			},
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return []byte("hash")
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return header
				},
			},
			"executionResultsInclusionEstimator": &processMocks.InclusionEstimatorMock{
				DecideCalled: func(lastNotarised *estimator.LastExecutionResultForInclusion, pending []data.BaseExecutionResultHandler, currentHdrTsMs uint64) (allowed int) {
					return 1
				},
			},
			"shardCoordinator": &mock.ShardCoordinatorStub{
				SelfIdCalled: func() uint32 {
					return 0
				},
			},
		})
		err := sp.AddExecutionResultsOnHeader(&block.HeaderV3{Round: 3})
		require.Error(t, err)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})
	t.Run("will work with valid data", func(t *testing.T) {
		t.Parallel()

		baseExecutionResults := &block.BaseExecutionResult{
			HeaderHash:  []byte("hash"),
			HeaderNonce: 3,
			HeaderRound: 1,
			RootHash:    []byte("rootHash"),
		}
		header := &block.HeaderV3{
			PrevHash: []byte("prev_hash"),

			LastExecutionResult: &block.ExecutionResultInfo{
				NotarizedInRound: 1,
				ExecutionResult:  baseExecutionResults,
			},
		}

		genesisTimeStampMs := uint64(1000)
		roundTime := uint64(100)
		roundHandler := &round.RoundHandlerMock{
			GetTimeStampForRoundCalled: func(round uint64) uint64 {
				return genesisTimeStampMs + round*roundTime
			},
		}
		defaultCfg := config.ExecutionResultInclusionEstimatorConfig{
			SafetyMargin:       110,
			MaxResultsPerBlock: 10,
		}

		executionResult1 := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{HeaderHash: []byte("hash1"), HeaderNonce: 1, HeaderRound: 2, GasUsed: 100_000_000},
		}
		executionResult2 := &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{HeaderHash: []byte("hash2"), HeaderNonce: 2, HeaderRound: 2, GasUsed: 999_000_000},
		}
		sp, _ := blproc.ConstructPartialShardBlockProcessorForTest(map[string]interface{}{
			"executionResultsTracker": &testscommonExecutionTrack.ExecutionResultsTrackerStub{
				GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
					return []data.BaseExecutionResultHandler{
						executionResult1,
						executionResult2,
					}, nil
				},
			},
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return []byte("hash")
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return header
				},
			},
			"executionResultsInclusionEstimator": estimator.NewExecutionResultInclusionEstimator(defaultCfg, roundHandler),
			"shardCoordinator": &mock.ShardCoordinatorStub{
				SelfIdCalled: func() uint32 {
					return 0
				},
			},
		})

		proposalHeader := &block.HeaderV3{Round: 3}
		err := sp.AddExecutionResultsOnHeader(proposalHeader)

		// expected only first pending execution result to be added
		require.NoError(t, err)
		actualLastExecutionResult := proposalHeader.GetLastExecutionResultHandler().(*block.ExecutionResultInfo)
		assert.NotNil(t, actualLastExecutionResult)
		assert.Equal(t, actualLastExecutionResult.ExecutionResult, executionResult1.BaseExecutionResult)
		assert.NotNil(t, proposalHeader.ExecutionResults)
		assert.Len(t, proposalHeader.ExecutionResults, 1)
		assert.Equal(t, proposalHeader.ExecutionResults[0], executionResult1)
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
		require.Equal(t, 2, cntAddReferencedMetaBlockCalled) // should be called twice, the third hdr returns shouldContinue false
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

	t.Run("checkEpochCorectnessCrossChain fails should error", func(t *testing.T) {
		t.Parallel()

		subcomponents := createSubComponentsForVerifyProposalTest()
		subcomponents["epochChangeGracePeriodHandler"] = &gracePeriodErrStub{}
		subcomponents["epochStartTrigger"] = &mock.EpochStartTriggerStub{
			EpochStartRoundCalled: func() uint64 {
				return 10
			},
			EpochFinalityAttestingRoundCalled: func() uint64 {
				return 15
			},
		}
		genesisNonce := uint64(0)
		subcomponents["genesisNonce"] = genesisNonce
		subcomponents["forkDetector"] = &mock.ForkDetectorMock{
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return genesisNonce
			},
		}
		sp, err := blproc.ConstructPartialShardBlockProcessorForTest(subcomponents)
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
		require.Error(t, err)
		assert.Equal(t, "epochChangeGracePeriodHandler forced error", err.Error())
	})
	t.Run("checkEpochCorectness fails should error", func(t *testing.T) {
		t.Parallel()

		subcomponents := createSubComponentsForVerifyProposalTest()

		sp, err := blproc.ConstructPartialShardBlockProcessorForTest(subcomponents)
		require.Nil(t, err)

		body := &block.Body{}

		header := &block.HeaderV3{
			PrevHash:         []byte("hash"),
			Nonce:            1,
			Round:            2,
			Epoch:            5,
			MiniBlockHeaders: []block.MiniBlockHeader{},
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{},
			},
		}

		err = sp.VerifyBlockProposal(header, body, haveTime)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "epoch does not match")
	})
	t.Run("checkMetaHeadersValidityAndFinalityProposal fails should error", func(t *testing.T) {
		t.Parallel()

		expError := errors.New("expected error from checkMetaHeadersValidityAndFinalityProposal")
		subcomponents := createSubComponentsForVerifyProposalTest()
		subcomponents["blockTracker"] = &mock.BlockTrackerMock{
			GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
				return nil, nil, expError
			},
		}

		sp, err := blproc.ConstructPartialShardBlockProcessorForTest(subcomponents)
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
		require.Error(t, err)
		assert.Equal(t, expError, err)
	})
	t.Run("verifyCrossShardMiniBlockDstMe fails should error", func(t *testing.T) {
		t.Parallel()

		headerValidator := &processMocks.HeaderValidatorMock{
			IsHeaderConstructionValidCalled: func(currHdr, prevHdr data.HeaderHandler) error {
				return nil
			},
		}
		subcomponents := createSubComponentsForVerifyProposalTest()
		subcomponents["headerValidator"] = headerValidator
		subcomponents["hasher"] = &hashingMocks.HasherMock{}
		subcomponents["proofsPool"] = &dataRetriever.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				return true
			},
		}

		sp, err := blproc.ConstructPartialShardBlockProcessorForTest(subcomponents)
		require.Nil(t, err)

		body := &block.Body{}

		metablockHashes := [][]byte{
			[]byte("hash1"),
			[]byte("hash2"),
		}
		header := &block.HeaderV3{
			PrevHash:         []byte("hash"),
			Nonce:            1,
			Round:            2,
			MiniBlockHeaders: []block.MiniBlockHeader{},
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{},
			},
			MetaBlockHashes: metablockHashes,
		}

		err = sp.VerifyBlockProposal(header, body, haveTime)

		require.Error(t, err)
		// stub will not return meta headers, causing type assertion to fail
		expError := errors.New("wrong type assertion")
		assert.Equal(t, expError, err)
	})
	t.Run("verifyGasLimit fails should error", func(t *testing.T) {
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
		arguments.GasComputation = &testscommon.GasComputationMock{
			CheckIncomingMiniBlocksCalled: func(miniBlocks []data.MiniBlockHeaderHandler, transactions map[string][]data.TransactionHandler) (int, int, error) {
				return 0, 0, expectedError
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
		require.Error(t, err)
		require.Equal(t, expectedErr, err)
	})
	t.Run("getHeaderHash fails should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		coreComponents.IntMarsh = &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				t.Log("called IntMarsh.Marshal on ", obj)
				return nil, expectedErr
			},
		}

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
		require.Error(t, err)
		require.Equal(t, expectedErr, err)
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

	t.Run("GetHeaderByHash error should propagate", func(t *testing.T) {
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

		expectedError := errors.New("expected error from GetHeaderByHash")
		dataPool.HeadersCalled = func() retriever.HeadersPool {
			return &pool.HeadersPoolStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					return nil, expectedError
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
		require.Error(t, err)
		require.ErrorIs(t, err, expectedError)
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

func TestShardProcessor_collectExecutionResults(t *testing.T) {
	t.Parallel()

	t.Run("with CreateReceiptsHash error should return error", func(t *testing.T) {
		t.Parallel()

		txCoordinator := &testscommon.TransactionCoordinatorMock{
			CreateReceiptsHashCalled: func() ([]byte, error) {
				return nil, expectedErr
			},
		}
		sp, err := blproc.ConstructPartialShardBlockProcessorForTest(map[string]interface{}{
			"txCoordinator": txCoordinator,
			"shardCoordinator": &mock.ShardCoordinatorStub{
				SelfIdCalled: func() uint32 {
					return 0
				},
			},
		})
		require.Nil(t, err)

		header := &block.HeaderV3{}
		_, err = sp.CollectExecutionResults(make([]byte, 0), header, &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("with gas used exceeds gas provided should return error", func(t *testing.T) {
		t.Parallel()

		subComponents, header, body := createSubComponentsForCollectExecutionResultsTest()
		gasProvider := subComponents["gasConsumedProvider"].(*testscommon.GasHandlerStub)
		gasProvider.TotalGasProvidedCalled = func() uint64 {
			return 10 // less than gas used in test
		}
		sp, err := blproc.ConstructPartialShardBlockProcessorForTest(subComponents)
		require.Nil(t, err)
		headerHash := []byte("header hash to be tested")
		_, err = sp.CollectExecutionResults(headerHash, header, body)
		assert.Equal(t, process.ErrGasUsedExceedsGasProvided, err)
	})

	t.Run("with cacheExecutedMiniBlocks error should return error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("MarshalizerMock generic error")
		subComponents, header, body := createSubComponentsForCollectExecutionResultsTest()

		expected_blocks := 10
		marshallerCalled := 0
		marshallerMock := subComponents["marshalizer"].(*mock.MarshalizerMock)
		marshaller := &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				// Marshaller is first called for mini block headers, then for executed mini blocks
				// We want to fail on the executed mini blocks marshalling

				if marshallerCalled == expected_blocks {
					marshallerMock.Fail = true
				}
				marshallerCalled++
				return marshallerMock.Marshal(obj)
			},
		}
		subComponents["marshalizer"] = marshaller
		sp, err := blproc.ConstructPartialShardBlockProcessorForTest(subComponents)
		require.Nil(t, err)
		headerHash := []byte("header hash to be tested")
		_, err = sp.CollectExecutionResults(headerHash, header, body)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("with createMiniBlockHeaderHandlers error should return error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("MarshalizerMock generic error")
		subComponents, header, body := createSubComponentsForCollectExecutionResultsTest()

		marshaller := subComponents["marshalizer"].(*mock.MarshalizerMock)
		marshaller.Fail = true
		sp, err := blproc.ConstructPartialShardBlockProcessorForTest(subComponents)
		require.Nil(t, err)
		headerHash := []byte("header hash to be tested")
		_, err = sp.CollectExecutionResults(headerHash, header, body)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("with cacheIntermediateTxsForHeader error should return error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("cacheIntermediateTxsForHeader error")
		subComponents, header, body := createSubComponentsForCollectExecutionResultsTest()

		marshallerMock := subComponents["marshalizer"].(*mock.MarshalizerMock)
		marshaller := &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				if _, ok := obj.(map[block.Type]map[string]data.TransactionHandler); ok {
					return nil, expectedErr
				} else {
					return marshallerMock.Marshal(obj)
				}
			}}
		subComponents["marshalizer"] = marshaller
		sp, err := blproc.ConstructPartialShardBlockProcessorForTest(subComponents)
		require.Nil(t, err)
		headerHash := []byte("header hash to be tested")
		_, err = sp.CollectExecutionResults(headerHash, header, body)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		subComponents, header, body := createSubComponentsForCollectExecutionResultsTest()

		sp, err := blproc.ConstructPartialShardBlockProcessorForTest(subComponents)
		require.Nil(t, err)

		headerHash := []byte("header hash to be tested")
		result, err := sp.CollectExecutionResults(headerHash, header, body)

		require.Nil(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uint64(1350), result.GetGasUsed(), "gas used should be set correctly")
		assert.Equal(t, uint32(10), result.GetHeaderEpoch(), "epoch should be 10 as per mock header")
		assert.Equal(t, headerHash, result.GetHeaderHash(), "header hash should match input")
		assert.Equal(t, uint64(155), result.GetHeaderNonce(), "nonce should be 155 as per mock header")
		assert.Equal(t, uint64(2067), result.GetHeaderRound(), "round should be 2067 as per mock header")
		assert.Equal(t, []byte("root hash to be tested"), result.GetRootHash(), "root hash should match mock")

		realResult := result.(*block.ExecutionResult)
		realResult.GetMiniBlockHeaders()
		actual_miniblocks := map[block.Type]int{}
		expected_miniblocks := map[block.Type]int{
			block.TxBlock:                  4,
			block.SmartContractResultBlock: 2,
			block.ReceiptBlock:             1,
			block.RewardsBlock:             1,
			block.InvalidBlock:             1,
			block.PeerBlock:                1,
		}
		for i, mbResult := range realResult.GetMiniBlockHeaders() {
			switch mbResult.GetType() {
			case block.TxBlock:
				actual_miniblocks[block.TxBlock]++
				if mbResult.GetSenderShardID() == uint32(0) && mbResult.GetReceiverShardID() == uint32(0) {
					assert.Equal(t, uint32(2), mbResult.GetTxCount(), "same-shard tx miniblock should have 2 txs as per mock")
				} else if mbResult.GetSenderShardID() == uint32(0) && mbResult.GetReceiverShardID() == uint32(1) {
					assert.Equal(t, uint32(2), mbResult.GetTxCount(), "cross-shard tx miniblock should have 2 txs as per mock")
				} else if mbResult.GetSenderShardID() == uint32(1) && mbResult.GetReceiverShardID() == uint32(0) {
					assert.Equal(t, uint32(1), mbResult.GetTxCount(), "cross-shard tx miniblock should have 1 tx as per mock")
				} else if mbResult.GetSenderShardID() == uint32(2) && mbResult.GetReceiverShardID() == uint32(0) {
					assert.Equal(t, uint32(1), mbResult.GetTxCount(), "cross-shard tx miniblock should have 1 tx as per mock")
				} else {
					require.Fail(t, "unexpected tx miniblock shard IDs")
				}
			case block.SmartContractResultBlock:
				actual_miniblocks[block.SmartContractResultBlock]++
				if mbResult.GetSenderShardID() == uint32(0) && mbResult.GetReceiverShardID() == uint32(1) {
					assert.Equal(t, uint32(1), mbResult.GetTxCount(), "outgoing SCR miniblock should have 1 tx as per mock")
				} else if mbResult.GetSenderShardID() == uint32(2) && mbResult.GetReceiverShardID() == uint32(0) {
					assert.Equal(t, uint32(2), mbResult.GetTxCount(), "incoming SCR miniblock should have 2 txs as per mock")
				} else {
					require.Fail(t, "unexpected SCR miniblock shard IDs")
				}
			case block.ReceiptBlock:
				actual_miniblocks[block.ReceiptBlock]++
				assert.NotEqual(t, uint32(0), mbResult.GetSenderShardID(), "receipt miniblock should not be self-shard")
				assert.Equal(t, uint32(1), mbResult.GetTxCount(), "receipt miniblock should have 1 tx as per mock")
			case block.RewardsBlock:
				actual_miniblocks[block.RewardsBlock]++
				assert.Equal(t, uint32(1), mbResult.GetTxCount(), "rewards miniblock should have 1 tx as per mock")
			case block.InvalidBlock:
				actual_miniblocks[block.InvalidBlock]++
				assert.Equal(t, uint32(1), mbResult.GetTxCount(), "invalid miniblock should have 1 tx as per mock")
			case block.PeerBlock:
				actual_miniblocks[block.PeerBlock]++
				assert.Equal(t, uint32(1), mbResult.GetTxCount(), "peer miniblock should have 1 tx as per mock")
			default:
				require.Fail(t, "unexpected miniblock type")
			}

			fmt.Println("MiniBlockHeader Result ", i, ": type ", mbResult.GetType(), ", sender shard ", mbResult.GetSenderShardID(), ", receiver shard ", mbResult.GetReceiverShardID(), ", tx count ", mbResult.GetTxCount())
		}
		assert.Equal(t, expected_miniblocks, actual_miniblocks, "miniblock counts by type should match expected")
		assert.Equal(t, big.NewInt(1000), realResult.GetAccumulatedFees(), "accumulated fees should match mock")
		assert.Equal(t, big.NewInt(100), realResult.GetDeveloperFees(), "developer fees should match mock")
	})
}

func createSubComponentsForCollectExecutionResultsTest() (map[string]interface{}, data.HeaderHandler, *block.Body) {

	txCoordinator := &testscommon.TransactionCoordinatorMock{
		GetCreatedMiniBlocksFromMeCalled: func() block.MiniBlockSlice {
			return block.MiniBlockSlice{
				// same-shard regular transactions
				{
					SenderShardID:   0,
					ReceiverShardID: 0,
					Type:            block.TxBlock,
					TxHashes: [][]byte{
						[]byte("tx_self_3"),
						[]byte("tx_self_4"),
					},
				},
				// cross-shard outgoing transactions
				{
					SenderShardID:   0,
					ReceiverShardID: 1,
					Type:            block.TxBlock,
					TxHashes: [][]byte{
						[]byte("tx_cross_3"),
						[]byte("tx_cross_4"),
					},
				},
				// invalid transactions
				{
					SenderShardID:   0,
					ReceiverShardID: 0,
					Type:            block.InvalidBlock,
					TxHashes: [][]byte{
						[]byte("tx_invalid_1"),
					},
				},
			}
		},

		CreatePostProcessMiniBlocksCalled: func() block.MiniBlockSlice {
			return block.MiniBlockSlice{
				// self shard SCRs - should be sanitized out
				{
					SenderShardID:   0,
					ReceiverShardID: 0,
					Type:            block.SmartContractResultBlock,
					TxHashes: [][]byte{
						[]byte("scr_self_1"),
						[]byte("scr_self_2"),
					},
				},
				// outgoing SCRs
				{
					SenderShardID:   0,
					ReceiverShardID: 1,
					Type:            block.SmartContractResultBlock,
					TxHashes: [][]byte{
						[]byte("scr_out_1"),
					},
				},
				// self shard receipts - should be sanitized out
				{
					SenderShardID:   0,
					ReceiverShardID: 0,
					Type:            block.ReceiptBlock,
					TxHashes: [][]byte{
						[]byte("rcpt_0"),
					},
				},
			}
		},

		CreateReceiptsHashCalled: func() ([]byte, error) {
			return []byte("mock_receipts_hash"), nil
		},
		GetAllIntermediateTxsCalled: func() map[block.Type]map[string]data.TransactionHandler {
			return map[block.Type]map[string]data.TransactionHandler{}
		},
	}

	rootHash := []byte("root hash to be tested")
	accounts := map[state.AccountsDbIdentifier]state.AccountsAdapter{
		0: &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return rootHash, nil
			},
		},
	}
	feeHandler := &mock.FeeAccumulatorStub{
		GetAccumulatedFeesCalled: func() *big.Int {
			return big.NewInt(1000)
		},
		GetDeveloperFeesCalled: func() *big.Int {
			return big.NewInt(100)
		},
	}
	gasConsumedProvider := &testscommon.GasHandlerStub{
		TotalGasProvidedCalled: func() uint64 {
			return 1500
		},
		TotalGasRefundedCalled: func() uint64 {
			return 100
		},
		TotalGasPenalizedCalled: func() uint64 {
			return 50
		},
	}
	subComponents := map[string]interface{}{
		"txCoordinator": txCoordinator,
		"shardCoordinator": &mock.ShardCoordinatorStub{
			SelfIdCalled: func() uint32 {
				return 0
			},
		},
		"feeHandler":          feeHandler,
		"gasConsumedProvider": gasConsumedProvider,
		"marshalizer":         &mock.MarshalizerMock{},
		"hasher":              &hashingMocks.HasherMock{},
		"enableEpochsHandler": enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
		"dataPool":            initDataPool(),
		"accountsDB":          accounts,
	}

	header, body := createHeaderAndBodyForTestingProcessBlockProposal()

	return subComponents, header, body
}

func createHeaderAndBodyForTestingProcessBlockProposal() (data.HeaderHandler, *block.Body) {
	header := &block.HeaderV3{
		PrevHash: []byte("previousHash"),
		Epoch:    10,
		Nonce:    155,
		Round:    2067,
		TxCount:  13,
	}
	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			//Outgoing miniblock from shard
			{
				SenderShardID: 0,
				TxHashes: [][]byte{
					[]byte("tx_self_1"),
					[]byte("tx_self_2"),
					[]byte("tx_cross_1"),
					[]byte("tx_cross_2"),
					[]byte("tx_invalid_1"),
					[]byte("tx_unexecutable_1"),
				},
			},
			// Incoming miniblocks to shard
			{
				SenderShardID:   1,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{[]byte("tx_incoming_1")},
			},
			{
				SenderShardID:   2,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{[]byte("tx_incoming_2")},
			},
			//Incoming SCR miniblock
			{
				SenderShardID:   2,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{[]byte("tx_scr_in_1"), []byte("tx_scr_in_2")},
				Type:            block.SmartContractResultBlock,
			},
			// Other types of miniblocks
			{
				Type:            block.RewardsBlock,
				SenderShardID:   core.MetachainShardId,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{[]byte("reward1")},
			},
			{
				Type:            block.PeerBlock,
				SenderShardID:   core.MetachainShardId,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{[]byte("peer1")},
			},
			{
				Type:            block.ReceiptBlock,
				SenderShardID:   2,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{[]byte("rcpt_1")},
			},
		},
	}
	return header, body
}

func createSubComponentsForVerifyProposalTest() map[string]interface{} {
	poolMock := initDataPool()
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
		PrevHash: []byte("hash"),
		Nonce:    0,
		Round:    1,
		ShardID:  0,
		LastExecutionResult: &block.ExecutionResultInfo{
			ExecutionResult: &block.BaseExecutionResult{},
		},
	}

	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = blkc.SetCurrentBlockHeaderAndRootHash(currentBlockHeader, []byte("root"))
	blkc.SetCurrentBlockHeaderHash([]byte("hash"))
	gracePeriod, _ := graceperiod.NewEpochChangeGracePeriod([]config.EpochChangeGracePeriodByEpoch{{EnableEpoch: 0, GracePeriodInRounds: 1}})

	return map[string]interface{}{
		"blockChain": blkc,
		"dataPool":   poolMock,
		"shardCoordinator": &mock.ShardCoordinatorStub{
			SelfIdCalled: func() uint32 {
				return 0
			},
		},
		"executionResultsVerifier": &processMocks.ExecutionResultsVerifierMock{
			VerifyHeaderExecutionResultsCalled: func(header data.HeaderHandler) error {
				return nil
			},
		},
		"executionResultsInclusionEstimator": &processMocks.InclusionEstimatorMock{
			DecideCalled: func(lastNotarised *estimator.LastExecutionResultForInclusion, pending []data.BaseExecutionResultHandler, currentHdrTsMs uint64) (allowed int) {
				return 0
			},
		},
		"missingDataResolver": &processMocks.MissingDataResolverMock{
			WaitForMissingDataCalled: func(timeout time.Duration) error {
				return nil
			},
		},
		"epochStartTrigger": &mock.EpochStartTriggerStub{
			EpochStartRoundCalled: func() uint64 {
				return 10
			},
			EpochFinalityAttestingRoundCalled: func() uint64 {
				return 5
			},
		},
		"enableEpochsHandler":           enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
		"epochChangeGracePeriodHandler": gracePeriod,
		"blockTracker": &mock.BlockTrackerMock{
			GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
				return &block.MetaBlockV3{}, []byte("h"), nil
			},
		},
		"gasComputation": &testscommon.GasComputationMock{
			CheckIncomingMiniBlocksCalled: func(miniBlocks []data.MiniBlockHeaderHandler, transactions map[string][]data.TransactionHandler) (int, int, error) {
				return len(miniBlocks) - 1, 0, nil // no pending mini blocks left
			},
			CheckOutgoingTransactionsCalled: func(txHashes [][]byte, transactions []data.TransactionHandler) ([][]byte, []data.MiniBlockHeaderHandler, error) {
				return txHashes, []data.MiniBlockHeaderHandler{&block.MiniBlockHeader{}}, nil // one pending mini block added
			},
		},
		"appStatusHandler": &statusHandlerMock.AppStatusHandlerStub{},
		"marshalizer":      &mock.MarshalizerMock{},
		"roundHandler":     &mock.RoundHandlerMock{},
	}

}
