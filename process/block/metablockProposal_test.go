package block_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"

	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	testscommonState "github.com/multiversx/mx-chain-go/testscommon/state"

	"github.com/multiversx/mx-chain-go/common"
	integrationTestsMock "github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/mbSelection"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
)

func TestMetaProcessor_CreateNewHeaderProposal(t *testing.T) {
	t.Parallel()

	defaultBootstrapComponents := &mock.BootstrapComponentsMock{
		Coordinator:          mock.NewOneShardCoordinatorMock(),
		HdrIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		VersionedHdrFactory: &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return &block.MetaBlock{}
			},
		},
	}

	validMetaHeaderV3 := testscommon.HeaderHandlerStub{
		IsHeaderV3Called: func() bool {
			return true
		},
		GetLastExecutionResultHandlerCalled: func() data.LastExecutionResultHandler {
			return &block.MetaExecutionResultInfo{
				ExecutionResult: &block.BaseMetaExecutionResult{
					BaseExecutionResult: &block.BaseExecutionResult{},
				},
			}
		},
	}

	prevValidMetaBlockV3 := testscommon.HeaderHandlerStub{
		IsHeaderV3Called: func() bool {
			return true
		},
		GetLastExecutionResultHandlerCalled: func() data.LastExecutionResultHandler {
			return &block.MetaExecutionResultInfo{
				ExecutionResult: &block.BaseMetaExecutionResult{},
			}
		},
	}
	validMetaExecutionResultsWithEpochChange := []data.BaseExecutionResultHandler{
		&block.MetaExecutionResult{
			ExecutionResult: &block.BaseMetaExecutionResult{},
			MiniBlockHeaders: []block.MiniBlockHeader{
				{
					Hash:          []byte("mb hash"),
					SenderShardID: core.MetachainShardId,
					Type:          block.RewardsBlock, // this miniBlock marks the epoch start
				},
			},
		},
	}
	validMetaExecutionResultsWithoutEpochChange := []data.BaseExecutionResultHandler{
		&block.MetaExecutionResult{
			ExecutionResult: &block.BaseMetaExecutionResult{},
			MiniBlockHeaders: []block.MiniBlockHeader{
				{
					Hash:            []byte("mb hash"),
					ReceiverShardID: core.MetachainShardId,
					SenderShardID:   0,
					Type:            block.TxBlock,
				},
			},
		},
	}

	t.Run("versioned header factory creates an invalid meta header, should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.EpochStartTrigger = &testscommon.EpochStartTriggerStub{
			EpochCalled: func() uint32 {
				return 1
			},
		}
		bc := *defaultBootstrapComponents
		bc.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return &block.Header{}
			},
		}

		arguments.BootstrapComponents = &bc

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Nil(t, header)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})
	t.Run("versioned header factory creates a metablock but with version < v3, should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.EpochStartTrigger = &testscommon.EpochStartTriggerStub{
			EpochCalled: func() uint32 {
				return 1
			},
		}
		bc := *defaultBootstrapComponents
		bc.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return &block.MetaBlock{}
			},
		}

		arguments.BootstrapComponents = &bc

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Nil(t, header)
		require.Equal(t, process.ErrInvalidHeader, err)
	})
	t.Run("correct meta header version, set round error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		bc := *defaultBootstrapComponents
		bc.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return &testscommon.HeaderHandlerStub{
					IsHeaderV3Called: func() bool {
						return true
					},
					SetRoundCalled: func(_ uint64) error {
						return expectedErr
					},
				}
			},
		}

		arguments.BootstrapComponents = &bc

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Nil(t, header)
		require.Equal(t, expectedErr, err)
	})
	t.Run("correct meta header version, set nonce error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		bc := *defaultBootstrapComponents
		versionedHeader := validMetaHeaderV3
		versionedHeader.SetNonceCalled = func(_ uint64) error {
			return expectedErr
		}
		bc.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return &versionedHeader
			},
		}

		arguments.BootstrapComponents = &bc

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Nil(t, header)
		require.Equal(t, expectedErr, err)
	})
	t.Run("correct meta header version, add execution result error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ExecutionManager = &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return nil, expectedErr
			},
		}
		bc := *defaultBootstrapComponents
		bc.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return &validMetaHeaderV3
			},
		}

		arguments.BootstrapComponents = &bc

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Nil(t, header)
		require.Equal(t, expectedErr, err)
	})
	t.Run("error checking epoch start data in execution results, should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ExecutionManager = &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return nil, nil
			},
		}

		metaBlockWithInvalidExecutionResult := validMetaHeaderV3
		metaBlockWithInvalidExecutionResult.GetExecutionResultsHandlersCalled = func() []data.BaseExecutionResultHandler {
			return []data.BaseExecutionResultHandler{
				&block.BaseExecutionResult{}, // invalid for meta block
			}
		}

		bc := *defaultBootstrapComponents
		bc.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return &metaBlockWithInvalidExecutionResult
			},
		}

		arguments.BootstrapComponents = &bc
		dataComponentsModified := *dataComponents
		dataComponentsModified.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &prevValidMetaBlockV3
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("prev header hash")
			},
		}
		arguments.DataComponents = &dataComponentsModified
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Nil(t, header)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
	})
	t.Run("with epoch start data in execution results, but missing epoch start data in meta block processor", func(t *testing.T) {
		t.Parallel()

		mapForMetaProcessor := createMetaProcessorMapForCreatingEpochStart()
		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(mapForMetaProcessor)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Equal(t, process.ErrNilEpochStartData, err)
		require.Nil(t, header)
	})
	t.Run("with epoch start data in execution results and in meta block processor, error on set epoch", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ExecutionManager = &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return nil, nil
			},
		}

		metaBlockWithValidExecutionResult := validMetaHeaderV3
		metaBlockWithValidExecutionResult.GetExecutionResultsHandlersCalled = func() []data.BaseExecutionResultHandler {
			return validMetaExecutionResultsWithEpochChange
		}
		metaBlockWithValidExecutionResult.SetEpochCalled = func(epoch uint32) error {
			require.Equal(t, uint32(1), epoch)
			return expectedErr
		}

		bc := *defaultBootstrapComponents
		bc.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return &metaBlockWithValidExecutionResult
			},
		}

		arguments.BootstrapComponents = &bc
		dataComponentsModified := *dataComponents
		dataComponentsModified.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &prevValidMetaBlockV3
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("prev header hash")
			},
		}
		arguments.DataComponents = &dataComponentsModified
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{
			EpochStartData: &block.EpochStart{
				LastFinalizedHeaders: make([]block.EpochStartShardData, 3),
				Economics:            block.Economics{},
			},
		})
		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Equal(t, expectedErr, err)
		require.Nil(t, header)
	})
	t.Run("with epoch start data in execution results and in meta block processor, error on set epoch start data", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ExecutionManager = &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return nil, nil
			},
		}

		metaBlockWithValidExecutionResult := validMetaHeaderV3
		metaBlockWithValidExecutionResult.GetExecutionResultsHandlersCalled = func() []data.BaseExecutionResultHandler {
			return validMetaExecutionResultsWithEpochChange
		}
		metaBlockWithValidExecutionResult.SetEpochStartHandlerCalled = func(_ data.EpochStartHandler) error {
			return expectedErr
		}

		bc := *defaultBootstrapComponents
		bc.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return &metaBlockWithValidExecutionResult
			},
		}

		arguments.BootstrapComponents = &bc
		dataComponentsModified := *dataComponents
		dataComponentsModified.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &prevValidMetaBlockV3
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("prev header hash")
			},
		}
		arguments.DataComponents = &dataComponentsModified
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{
			EpochStartData: &block.EpochStart{
				LastFinalizedHeaders: make([]block.EpochStartShardData, 3),
				Economics:            block.Economics{},
			},
		})
		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Equal(t, expectedErr, err)
		require.Nil(t, header)
	})
	t.Run("without epoch start data in execution results, should pass and not change epoch", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ExecutionManager = &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return nil, nil
			},
		}

		metaBlockWithValidExecutionResult := validMetaHeaderV3
		metaBlockWithValidExecutionResult.GetExecutionResultsHandlersCalled = func() []data.BaseExecutionResultHandler {
			return validMetaExecutionResultsWithoutEpochChange
		}
		metaBlockWithValidExecutionResult.SetEpochCalled = func(epoch uint32) error {
			require.Fail(t, "should not have been called")
			return nil
		}

		bc := *defaultBootstrapComponents
		bc.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return &metaBlockWithValidExecutionResult
			},
		}

		arguments.BootstrapComponents = &bc
		dataComponentsModified := *dataComponents
		dataComponentsModified.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &prevValidMetaBlockV3
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("prev header hash")
			},
		}
		arguments.DataComponents = &dataComponentsModified
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Nil(t, err)
		require.NotNil(t, header)
	})
	t.Run("with epoch start data in execution results and in meta block processor, should pass and change epoch", func(t *testing.T) {
		t.Parallel()

		headersPoolMock := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				return &block.MetaBlockV3{
					EpochChangeProposed: true,
				}, nil
			},
		}

		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return headersPoolMock
		}

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.DataPool = dataPool

		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlockV3{
					LastExecutionResult: &block.MetaExecutionResultInfo{
						ExecutionResult: &block.BaseMetaExecutionResult{},
					},
				}
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash1")
			},
		}

		bootstrapComponents.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return &block.MetaBlockV3{
					Epoch: 0,
				}
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{
			Epoch: 1,
			EpochStartData: &block.EpochStart{
				LastFinalizedHeaders: make([]block.EpochStartShardData, 3),
				Economics:            block.Economics{},
			},
		})

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Nil(t, err)
		require.NotNil(t, header)
	})

	t.Run("higher nonce in last execution result should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &prevValidMetaBlockV3
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash1")
			},
		}

		metaHeader := &testscommon.HeaderHandlerStub{
			IsHeaderV3Called: func() bool {
				return true
			},
			GetNonceCalled: func() uint64 {
				return 5
			},
			GetLastExecutionResultHandlerCalled: func() data.LastExecutionResultHandler {
				return &block.MetaExecutionResultInfo{
					ExecutionResult: &block.BaseMetaExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderNonce: 105,
						},
					},
				}
			},
		}

		bootstrapComponents.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return metaHeader
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 5)
		require.Nil(t, header)
		require.ErrorIs(t, err, process.ErrInvalidLastExecutionResult)
	})

	t.Run("nonce gap from last exec result exceeds maximum allowed, should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &testscommon.HeaderHandlerStub{
					IsHeaderV3Called: func() bool {
						return true
					},
					GetLastExecutionResultHandlerCalled: func() data.LastExecutionResultHandler {
						return &block.MetaExecutionResultInfo{
							ExecutionResult: &block.BaseMetaExecutionResult{
								BaseExecutionResult: &block.BaseExecutionResult{
									HeaderNonce: 5,
								},
							},
						}
					},
				}
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash1")
			},
		}

		metaHeader := &block.MetaBlockV3{
			Nonce: 105,
		}

		bootstrapComponents.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return metaHeader
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 105)
		require.Nil(t, header)
		require.ErrorIs(t, err, process.ErrNonceGapTooLarge)
		require.Contains(t, err.Error(), "from last execution")
		require.Contains(t, err.Error(), "gap of 100")
	})

	t.Run("nonce gap exceeds maximum allowed, should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &prevValidMetaBlockV3
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash1")
			},
		}

		metaHeader := &block.MetaBlockV3{
			ShardInfo: []block.ShardData{
				{
					ShardID: 0,
					Nonce:   100,
				},
			},
			ShardInfoProposal: []block.ShardDataProposal{
				{
					ShardID:    0,
					Nonce:      250, // 150 gap
					HeaderHash: []byte("hash"),
				},
			},
		}

		bootstrapComponents.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return metaHeader
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Nil(t, header)
		require.ErrorIs(t, err, process.ErrNonceGapTooLarge)
		require.Contains(t, err.Error(), "shard 0")
		require.Contains(t, err.Error(), "gap of 150")
	})

	t.Run("error on GetLastCrossNotarizedHeadersForAllShards should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &prevValidMetaBlockV3
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash1")
			},
		}

		metaHeader := &block.MetaBlockV3{
			ShardInfo: []block.ShardData{
				{
					ShardID: 0,
					Nonce:   100,
				},
			},
			ShardInfoProposal: []block.ShardDataProposal{
				{
					ShardID:    0,
					Nonce:      101,
					HeaderHash: []byte("hash"),
				},
			},
		}
		bootstrapComponents.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return metaHeader
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.BlockTracker = &integrationTestsMock.BlockTrackerStub{
			GetLastCrossNotarizedHeadersForAllShardsCalled: func() (map[uint32]data.HeaderHandler, error) {
				return nil, expectedError
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Equal(t, expectedError, err)
		require.Nil(t, header)
	})

	t.Run("missing last notarized in block tracker for included shard should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &prevValidMetaBlockV3
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash1")
			},
		}

		metaHeader := &block.MetaBlockV3{
			ShardInfo: []block.ShardData{
				{
					ShardID: 0,
					Nonce:   100,
				},
				{
					// this shard does not exist in block tracker
					ShardID:    1,
					Nonce:      250,
					HeaderHash: []byte("hash2"),
				},
			},
			ShardInfoProposal: []block.ShardDataProposal{
				{
					ShardID:    0,
					Nonce:      101,
					HeaderHash: []byte("hash"),
				},
			},
		}
		bootstrapComponents.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return metaHeader
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Equal(t, process.ErrMissingCrossNotarizedHeader, err)
		require.Nil(t, header)
	})

	t.Run("higher last notarized in block tracker should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &prevValidMetaBlockV3
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash1")
			},
		}

		metaHeader := &block.MetaBlockV3{
			ShardInfo: []block.ShardData{
				{
					ShardID: 0,
					Nonce:   100,
				},
			},
			ShardInfoProposal: []block.ShardDataProposal{
				{
					ShardID:    0,
					Nonce:      101,
					HeaderHash: []byte("hash"),
				},
			},
		}
		bootstrapComponents.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return metaHeader
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		crossNotarizedWithHigherNonce := &block.HeaderV3{Nonce: 102}
		arguments.BlockTracker.AddCrossNotarizedHeader(0, crossNotarizedWithHigherNonce, []byte("hash higher nonce"))
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Equal(t, process.ErrInvalidShardInfo, err)
		require.Nil(t, header)
	})

	t.Run("missing last notarized in block tracker for proposed shard should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &prevValidMetaBlockV3
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash1")
			},
		}

		metaHeader := &block.MetaBlockV3{
			ShardInfo: []block.ShardData{
				{
					ShardID: 0,
					Nonce:   100,
				},
			},
			ShardInfoProposal: []block.ShardDataProposal{
				{
					ShardID:    0,
					Nonce:      101,
					HeaderHash: []byte("hash"),
				},
				{
					// this shard does not exist in block tracker
					ShardID:    1,
					Nonce:      250,
					HeaderHash: []byte("hash2"),
				},
			},
		}
		bootstrapComponents.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return metaHeader
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Equal(t, process.ErrMissingCrossNotarizedHeader, err)
		require.Nil(t, header)
	})

	t.Run("lower proposed nonce should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &prevValidMetaBlockV3
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash1")
			},
		}

		metaHeader := &block.MetaBlockV3{
			ShardInfo: []block.ShardData{
				{
					ShardID: 0,
					Nonce:   100,
				},
			},
			ShardInfoProposal: []block.ShardDataProposal{
				{
					ShardID:    0,
					Nonce:      90, // lower
					HeaderHash: []byte("hash"),
				},
			},
		}
		bootstrapComponents.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return metaHeader
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.True(t, errors.Is(err, process.ErrInvalidProposedNonce))
		require.Contains(t, err.Error(), "proposed nonce 90")
		require.Nil(t, header)
	})

	t.Run("nonce gap within allowed limit, should succeed", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &prevValidMetaBlockV3
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("hash1")
			},
		}

		metaHeader := &block.MetaBlockV3{
			ShardInfo: []block.ShardData{
				{
					ShardID: 0,
					Nonce:   100,
				},
			},
			ShardInfoProposal: []block.ShardDataProposal{
				{
					ShardID:    0,
					Nonce:      101,
					HeaderHash: []byte("hash"),
				},
			},
		}
		bootstrapComponents.VersionedHdrFactory = &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return metaHeader
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Nil(t, err)
		require.NotNil(t, header)
	})
}

func TestMetaProcessor_CreateBlockProposal(t *testing.T) {
	t.Parallel()

	t.Run("nil header", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		checkCreateBlockProposalResult(t, mp, nil, haveTimeTrue, process.ErrNilBlockHeader)
	})
	t.Run("not header v3", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		notV3Header := &block.MetaBlock{}
		checkCreateBlockProposalResult(t, mp, notV3Header, haveTimeTrue, process.ErrInvalidHeader)
	})
	t.Run("shard header v3", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		shardHeaderV3 := &block.HeaderV3{}
		checkCreateBlockProposalResult(t, mp, shardHeaderV3, haveTimeTrue, process.ErrWrongTypeAssertion)
	})
	t.Run("createBlockBodyProposal error (ComputeLongestShardsChainsFromLastNotarized error)", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.BlockTracker = &mock.BlockTrackerMock{
			ComputeLongestShardsChainsFromLastNotarizedCalled: func() ([]data.HeaderHandler, [][]byte, map[uint32][]data.HeaderHandler, error) {
				return nil, nil, nil, expectedErr
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		validMetaHeaderV3 := &block.MetaBlockV3{}
		checkCreateBlockProposalResult(t, mp, validMetaHeaderV3, haveTimeTrue, expectedErr)
	})
	t.Run("createShardInfoV3 error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ShardInfoCreator = &processMocks.ShardInfoCreatorMock{
			CreateShardInfoV3Called: func(metaHeader data.MetaHeaderHandler, shardHeaders []data.HeaderHandler, shardHeaderHashes [][]byte) ([]data.ShardDataProposalHandler, []data.ShardDataHandler, error) {
				return nil, nil, expectedErr
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		validMetaHeaderV3 := &block.MetaBlockV3{}
		checkCreateBlockProposalResult(t, mp, validMetaHeaderV3, haveTimeTrue, expectedErr)
	})
	t.Run("set shard info error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			GetMiniBlocksCalled: func() block.MiniBlockSlice {
				return make([]*block.MiniBlock, 5) // coverage
			},
		}
		var invalidShardData data.ShardDataHandler
		arguments.ShardInfoCreator = &processMocks.ShardInfoCreatorMock{
			CreateShardInfoV3Called: func(metaHeader data.MetaHeaderHandler, shardHeaders []data.HeaderHandler, shardHeaderHashes [][]byte) ([]data.ShardDataProposalHandler, []data.ShardDataHandler, error) {
				return nil, []data.ShardDataHandler{invalidShardData}, nil
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		validMetaHeaderV3 := &block.MetaBlockV3{}
		checkCreateBlockProposalResult(t, mp, validMetaHeaderV3, haveTimeTrue, data.ErrInvalidTypeAssertion)
	})
	t.Run("set shard info proposal error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		var invalidShardDataProposal data.ShardDataProposalHandler
		arguments.ShardInfoCreator = &processMocks.ShardInfoCreatorMock{
			CreateShardInfoV3Called: func(metaHeader data.MetaHeaderHandler, shardHeaders []data.HeaderHandler, shardHeaderHashes [][]byte) ([]data.ShardDataProposalHandler, []data.ShardDataHandler, error) {
				return []data.ShardDataProposalHandler{invalidShardDataProposal}, []data.ShardDataHandler{}, nil
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		validMetaHeaderV3 := &block.MetaBlockV3{}
		checkCreateBlockProposalResult(t, mp, validMetaHeaderV3, haveTimeTrue, data.ErrInvalidTypeAssertion)
	})
	t.Run("set mini block header handlers error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		var invalidMiniBlockHeader data.MiniBlockHeaderHandler
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			GetMiniBlockHeaderHandlersCalled: func() []data.MiniBlockHeaderHandler {
				return []data.MiniBlockHeaderHandler{invalidMiniBlockHeader}
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		validMetaHeaderV3 := &block.MetaBlockV3{}
		checkCreateBlockProposalResult(t, mp, validMetaHeaderV3, haveTimeTrue, data.ErrInvalidTypeAssertion)
	})
	t.Run("marshall error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			GetMiniBlockHeaderHandlersCalled: func() []data.MiniBlockHeaderHandler { return nil },
		}
		cc := coreComponents
		cc.IntMarsh = &testscommon.MarshallerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		validMetaHeaderV3 := &block.MetaBlockV3{}
		checkCreateBlockProposalResult(t, mp, validMetaHeaderV3, haveTimeTrue, expectedErr)
	})
	t.Run("successful creation, non start of epoch block", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			GetMiniBlockHeaderHandlersCalled: func() []data.MiniBlockHeaderHandler { return nil },
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		validMetaHeaderV3 := &block.MetaBlockV3{}
		header, body, err := mp.CreateBlockProposal(validMetaHeaderV3, haveTimeTrue)
		require.Nil(t, err)
		require.NotNil(t, header)
		require.NotNil(t, body)
	})
	t.Run("no mini blocks added if epoch change propose set", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			GetMiniBlockHeaderHandlersCalled: func() []data.MiniBlockHeaderHandler {
				require.Fail(t, "should not be called")
				return nil
			},
		}
		arguments.EpochStartTrigger = &testscommon.EpochStartTriggerStub{
			GetEpochChangeProposedCalled: func() bool {
				return true
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		validMetaHeaderV3 := &block.MetaBlockV3{}
		header, body, err := mp.CreateBlockProposal(validMetaHeaderV3, haveTimeTrue)
		require.Nil(t, err)
		require.NotNil(t, header)
		require.NotNil(t, body)
	})
	t.Run("successful creation, start of epoch block with empy body", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			GetMiniBlockHeaderHandlersCalled: func() []data.MiniBlockHeaderHandler { return nil },
		}
		arguments.GasComputation = &testscommon.GasComputationMock{
			ResetCalled: func() {
				require.Fail(t, "should not be called")
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		// add epoch start data to the meta processor so that IsEpochStartBlock returns true
		validMetaHeaderV3 := &block.MetaBlockV3{
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: make([]block.EpochStartShardData, 3),
			},
		}
		header, body, err := mp.CreateBlockProposal(validMetaHeaderV3, haveTimeTrue)
		require.Nil(t, err)
		require.NotNil(t, header)
		require.NotNil(t, body)
		b := body.(*block.Body)
		// start of epoch block should have no mini blocks headers and no mini blocks in the body
		require.Len(t, header.GetMiniBlockHeaderHandlers(), 0)
		require.Len(t, b.MiniBlocks, 0)
	})
	t.Run("successful creation, epoch change proposal block with empy body", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			GetMiniBlockHeaderHandlersCalled: func() []data.MiniBlockHeaderHandler { return nil },
		}
		arguments.GasComputation = &testscommon.GasComputationMock{
			ResetCalled: func() {
				require.Fail(t, "should not be called")
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		// add epoch start data to the meta processor so that IsEpochStartBlock returns true
		validMetaHeaderV3 := &block.MetaBlockV3{
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: make([]block.EpochStartShardData, 3),
			},
		}
		header, body, err := mp.CreateBlockProposal(validMetaHeaderV3, haveTimeTrue)
		require.Nil(t, err)
		require.NotNil(t, header)
		require.NotNil(t, body)
		b := body.(*block.Body)
		// epoch change proposal should have no mini blocks headers and no mini blocks in the body
		require.Len(t, header.GetMiniBlockHeaderHandlers(), 0)
		require.Len(t, b.MiniBlocks, 0)
	})
}

func TestMetaProcessor_VerifyBlockProposal(t *testing.T) {
	t.Run("invalid body handler, should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		err = mp.VerifyBlockProposal(&block.MetaBlockV3{}, nil, haveTime)
		require.ErrorIs(t, err, process.ErrNilBlockBody)
	})
	t.Run("block hash does not match, should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = createTestBlockchain()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header := &block.MetaBlockV3{
			PrevHash: []byte("prevHash"),
			Nonce:    1,
		}
		body := &block.Body{}
		err = mp.VerifyBlockProposal(header, body, haveTime)
		require.ErrorIs(t, err, process.ErrBlockHashDoesNotMatch)
	})
	t.Run("invalid header handler, should error", func(t *testing.T) {
		t.Parallel()

		prevBlockHash := []byte("prev header hash")
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents = &mock.DataComponentsMock{
			Storage:  dataComponents.Storage,
			DataPool: dataComponents.DataPool,
			BlockChain: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return prevBlockHash
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{}
				},
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header := &block.MetaBlock{
			PrevHash: prevBlockHash,
			Nonce:    1,
			Round:    1,
		}
		body := &block.Body{}
		err = mp.VerifyBlockProposal(header, body, haveTime)
		require.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})
	t.Run("header handler of type MetaBlock, should error", func(t *testing.T) {
		t.Parallel()

		prevBlockHash := []byte("prev header hash")
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents = &mock.DataComponentsMock{
			Storage:  dataComponents.Storage,
			DataPool: dataComponents.DataPool,
			BlockChain: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return prevBlockHash
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{}
				},
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header := &block.MetaBlock{
			PrevHash: prevBlockHash,
			Nonce:    1,
			Round:    1,
		}
		body := &block.Body{}
		err = mp.VerifyBlockProposal(header, body, haveTime)
		require.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})
	t.Run("body handler of type BodyV3, should error", func(t *testing.T) {
		t.Parallel()

		prevBlockHash := []byte("prev header hash")
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents = &mock.DataComponentsMock{
			Storage:  dataComponents.Storage,
			DataPool: dataComponents.DataPool,
			BlockChain: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return prevBlockHash
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{}
				},
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header := &block.MetaBlockV3{
			PrevHash: prevBlockHash,
			Nonce:    1,
			Round:    1,
		}
		body := &wrongBody{}
		err = mp.VerifyBlockProposal(header, body, haveTime)
		require.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})
	t.Run("epoch change proposed outside trigger window, should error", func(t *testing.T) {
		t.Parallel()

		prevBlockHash := []byte("prev header hash")
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents = &mock.DataComponentsMock{
			Storage:  dataComponents.Storage,
			DataPool: dataComponents.DataPool,
			BlockChain: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return prevBlockHash
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						Epoch: 1,
					}
				},
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.EpochStartTrigger = &testscommon.EpochStartTriggerStub{
			EpochCalled: func() uint32 {
				return 1
			},
			ShouldProposeEpochChangeCalled: func(round uint64, nonce uint64) bool {
				return false
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header := &block.MetaBlockV3{
			PrevHash:            prevBlockHash,
			Nonce:               1,
			Round:               1,
			Epoch:               1,
			EpochChangeProposed: true,
		}
		body := &block.Body{}
		err = mp.VerifyBlockProposal(header, body, haveTime)
		require.ErrorIs(t, err, process.ErrEpochChangeProposedOutsideTriggerWindow)
	})
	t.Run("epoch change should be proposed but header flag is missing, should error", func(t *testing.T) {
		t.Parallel()

		prevBlockHash := []byte("prev header hash")
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents = &mock.DataComponentsMock{
			Storage:  dataComponents.Storage,
			DataPool: dataComponents.DataPool,
			BlockChain: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return prevBlockHash
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						Epoch: 1,
					}
				},
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.EpochStartTrigger = &testscommon.EpochStartTriggerStub{
			EpochCalled: func() uint32 {
				return 1
			},
			ShouldProposeEpochChangeCalled: func(round uint64, nonce uint64) bool {
				return true
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header := &block.MetaBlockV3{
			PrevHash: prevBlockHash,
			Nonce:    1,
			Round:    1,
			Epoch:    1,
		}
		body := &block.Body{}
		err = mp.VerifyBlockProposal(header, body, haveTime)
		require.ErrorIs(t, err, process.ErrEpochChangeProposedOutsideTriggerWindow)
	})
	t.Run("body mismatch, should error", func(t *testing.T) {
		t.Parallel()

		prevBlockHash := []byte("prev header hash")
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents = &mock.DataComponentsMock{
			Storage:  dataComponents.Storage,
			DataPool: dataComponents.DataPool,
			BlockChain: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return prevBlockHash
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{}
				},
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header := &block.MetaBlockV3{
			PrevHash: prevBlockHash,
			Nonce:    1,
			Round:    1,
		}
		body := &block.Body{MiniBlocks: []*block.MiniBlock{
			{SenderShardID: 0},
		}}
		err = mp.VerifyBlockProposal(header, body, haveTime)
		require.ErrorIs(t, err, process.ErrHeaderBodyMismatch)
	})
	t.Run("invalid header execution results, should error", func(t *testing.T) {
		t.Parallel()

		prevBlockHash := []byte("prev header hash")
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents = &mock.DataComponentsMock{
			Storage:  dataComponents.Storage,
			DataPool: dataComponents.DataPool,
			BlockChain: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return prevBlockHash
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{}
				},
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ExecutionResultsVerifier = &processMocks.ExecutionResultsVerifierMock{
			VerifyHeaderExecutionResultsCalled: func(header data.HeaderHandler) error {
				return expectedErr
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header := &block.MetaBlockV3{
			PrevHash: prevBlockHash,
			Nonce:    1,
			Round:    1,
		}
		body := &block.Body{}
		err = mp.VerifyBlockProposal(header, body, haveTime)
		require.ErrorIs(t, err, expectedErr)
	})
	t.Run("error on nonce gap verification", func(t *testing.T) {
		t.Parallel()

		prevBlockHash := []byte("prev header hash")
		prevLastMetaExecutionResult := &block.MetaExecutionResultInfo{
			ExecutionResult: &block.BaseMetaExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{},
			},
		}
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents = &mock.DataComponentsMock{
			Storage:  dataComponents.Storage,
			DataPool: dataComponents.DataPool,
			BlockChain: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return prevBlockHash
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						LastExecutionResult: prevLastMetaExecutionResult,
					}
				},
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MissingDataResolver = &processMocks.MissingDataResolverMock{
			RequestMissingShardHeadersCalled: func(_ data.MetaHeaderHandler) error {
				require.Fail(t, "should have not been called")
				return nil
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header := &block.MetaBlockV3{
			PrevHash:            prevBlockHash,
			Nonce:               1,
			Round:               1,
			LastExecutionResult: prevLastMetaExecutionResult,
			ShardInfo: []block.ShardData{
				{
					ShardID: 0,
					Nonce:   100,
				},
			},
			ShardInfoProposal: []block.ShardDataProposal{
				{
					ShardID:    0,
					Nonce:      250, // 150 gap
					HeaderHash: []byte("hash"),
				},
			},
		}
		body := &block.Body{}
		err = mp.VerifyBlockProposal(header, body, haveTime)
		require.ErrorIs(t, err, process.ErrNonceGapTooLarge)
	})
	t.Run("error on request missing shard header", func(t *testing.T) {
		t.Parallel()

		prevBlockHash := []byte("prev header hash")
		prevLastMetaExecutionResult := &block.MetaExecutionResultInfo{
			ExecutionResult: &block.BaseMetaExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{},
			},
		}
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents = &mock.DataComponentsMock{
			Storage:  dataComponents.Storage,
			DataPool: dataComponents.DataPool,
			BlockChain: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return prevBlockHash
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						LastExecutionResult: prevLastMetaExecutionResult,
					}
				},
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MissingDataResolver = &processMocks.MissingDataResolverMock{
			RequestMissingShardHeadersCalled: func(_ data.MetaHeaderHandler) error {
				return expectedErr
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header := &block.MetaBlockV3{
			PrevHash:            prevBlockHash,
			Nonce:               1,
			Round:               1,
			LastExecutionResult: prevLastMetaExecutionResult,
		}
		body := &block.Body{}
		err = mp.VerifyBlockProposal(header, body, haveTime)
		require.ErrorIs(t, err, expectedErr)
	})
	t.Run("error on wait for missing data", func(t *testing.T) {
		t.Parallel()

		prevBlockHash := []byte("prev header hash")
		prevLastMetaExecutionResult := &block.MetaExecutionResultInfo{
			ExecutionResult: &block.BaseMetaExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{},
			},
		}
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents = &mock.DataComponentsMock{
			Storage:  dataComponents.Storage,
			DataPool: dataComponents.DataPool,
			BlockChain: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return prevBlockHash
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						LastExecutionResult: prevLastMetaExecutionResult,
					}
				},
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MissingDataResolver = &processMocks.MissingDataResolverMock{
			WaitForMissingDataCalled: func(_ time.Duration) error {
				return expectedErr
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header := &block.MetaBlockV3{
			PrevHash:            prevBlockHash,
			Nonce:               1,
			Round:               1,
			LastExecutionResult: prevLastMetaExecutionResult,
		}
		body := &block.Body{}
		err = mp.VerifyBlockProposal(header, body, haveTime)
		require.ErrorIs(t, err, expectedErr)
	})
	t.Run("error on check epoch correctness v3", func(t *testing.T) {
		t.Parallel()

		prevBlockHash := []byte("prev header hash")
		prevLastMetaExecutionResult := &block.MetaExecutionResultInfo{
			ExecutionResult: &block.BaseMetaExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{},
			},
		}
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents = &mock.DataComponentsMock{
			Storage:  dataComponents.Storage,
			DataPool: dataComponents.DataPool,
			BlockChain: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return prevBlockHash
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						LastExecutionResult: prevLastMetaExecutionResult,
					}
				},
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MissingDataResolver = &processMocks.MissingDataResolverMock{
			WaitForMissingDataCalled: func(_ time.Duration) error {
				return nil
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header := &block.MetaBlockV3{
			PrevHash:            prevBlockHash,
			Nonce:               1,
			Round:               1,
			LastExecutionResult: prevLastMetaExecutionResult,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{
					{}, {},
				},
			},
		}
		body := &block.Body{}
		err = mp.VerifyBlockProposal(header, body, haveTime)
		require.ErrorIs(t, err, process.ErrEpochDoesNotMatch)
	})
	t.Run("error on check shard headers validity and finality proposal", func(t *testing.T) {
		t.Parallel()

		prevBlockHash := []byte("prev header hash")
		prevLastMetaExecutionResult := &block.MetaExecutionResultInfo{
			ExecutionResult: &block.BaseMetaExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{},
			},
		}
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents = &mock.DataComponentsMock{
			Storage:  dataComponents.Storage,
			DataPool: dataComponents.DataPool,
			BlockChain: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return prevBlockHash
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						LastExecutionResult: prevLastMetaExecutionResult,
					}
				},
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.BlockTracker = &integrationTestsMock.BlockTrackerStub{
			GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
				return nil, make([]byte, 0), expectedErr
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header := &block.MetaBlockV3{
			PrevHash:            prevBlockHash,
			Nonce:               1,
			Round:               1,
			LastExecutionResult: prevLastMetaExecutionResult,
		}
		body := &block.Body{}
		err = mp.VerifyBlockProposal(header, body, haveTime)
		require.ErrorIs(t, err, expectedErr)
	})
	t.Run("verify block proposal, should work", func(t *testing.T) {
		t.Parallel()

		prevBlockHash := []byte("prev header hash")
		prevLastMetaExecutionResult := &block.MetaExecutionResultInfo{
			ExecutionResult: &block.BaseMetaExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{},
			},
		}
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents = &mock.DataComponentsMock{
			Storage:  dataComponents.Storage,
			DataPool: dataComponents.DataPool,
			BlockChain: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return prevBlockHash
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						LastExecutionResult: prevLastMetaExecutionResult,
					}
				},
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header := &block.MetaBlockV3{
			PrevHash:            prevBlockHash,
			Nonce:               1,
			Round:               1,
			LastExecutionResult: prevLastMetaExecutionResult,
		}
		body := &block.Body{}

		err = mp.VerifyBlockProposal(header, body, haveTime)
		require.NoError(t, err)
	})
	t.Run("epoch change proposed with miniblocks in body, should error", func(t *testing.T) {
		t.Parallel()

		prevBlockHash := []byte("prev header hash")
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents = &mock.DataComponentsMock{
			Storage:  dataComponents.Storage,
			DataPool: dataComponents.DataPool,
			BlockChain: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return prevBlockHash
				},
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{}
				},
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.EpochStartTrigger = &testscommon.EpochStartTriggerStub{
			ShouldProposeEpochChangeCalled: func(round uint64, nonce uint64) bool {
				return true
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		header := &block.MetaBlockV3{
			PrevHash:            prevBlockHash,
			Nonce:               1,
			Round:               1,
			EpochChangeProposed: true,
		}
		body := &block.Body{MiniBlocks: []*block.MiniBlock{
			{SenderShardID: 0},
		}}
		err = mp.VerifyBlockProposal(header, body, haveTime)
		require.ErrorIs(t, err, process.ErrEpochStartProposeBlockHasMiniBlocks)
	})
}

func Test_checkShardHeadersValidityAndFinalityProposal(t *testing.T) {
	t.Parallel()

	t.Run("error on getting last cross notarized header", func(t *testing.T) {
		t.Parallel()

		metaHeader := &block.MetaBlockV3{}

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"shardCoordinator": mock.NewOneShardCoordinatorMock(),
			"blockTracker": &mock.BlockTrackerMock{
				GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
					return nil, nil, expectedErr
				},
			},
			"epochStartTrigger": &testscommon.EpochStartTriggerStub{
				GetEpochChangeProposedCalled: func() bool {
					return false
				},
			},
		})
		require.Nil(t, err)

		err = mp.CheckShardHeadersValidityAndFinalityProposal(metaHeader)
		require.ErrorIs(t, err, expectedErr)
	})
	t.Run("error on getting shard headers from meta header", func(t *testing.T) {
		t.Parallel()

		metaHeader := &block.MetaBlockV3{
			ShardInfoProposal: []block.ShardDataProposal{
				{
					HeaderHash: []byte("hash"),
				},
			},
		}

		headersPoolMock := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(_ []byte) (data.HeaderHandler, error) {
				return nil, expectedErr
			},
		}
		dataPoolMock := &dataRetrieverMock.PoolsHolderMock{}
		dataPoolMock.SetHeadersPool(headersPoolMock)

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"shardCoordinator":    mock.NewOneShardCoordinatorMock(),
			"enableEpochsHandler": &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			"blockTracker": &mock.BlockTrackerMock{
				GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
					return &testscommon.HeaderHandlerStub{}, nil, nil
				},
			},
			"epochStartTrigger": &testscommon.EpochStartTriggerStub{
				GetEpochChangeProposedCalled: func() bool {
					return false
				},
			},
			"dataPool": dataPoolMock,
		})
		require.Nil(t, err)

		err = mp.CheckShardHeadersValidityAndFinalityProposal(metaHeader)
		require.ErrorIs(t, err, process.ErrMissingHeader)
	})
	t.Run("error on missing header proof", func(t *testing.T) {
		t.Parallel()

		metaHeader := &block.MetaBlockV3{
			ShardInfoProposal: []block.ShardDataProposal{
				{
					HeaderHash: []byte("hash"),
				},
			},
		}

		headersPoolMock := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(_ []byte) (data.HeaderHandler, error) {
				return &block.MetaBlockV3{}, nil
			},
		}
		dataPoolMock := &dataRetrieverMock.PoolsHolderMock{}
		dataPoolMock.SetHeadersPool(headersPoolMock)

		proofsPool := &dataRetrieverMock.ProofsPoolMock{
			HasProofCalled: func(_ uint32, _ []byte) bool {
				return false
			},
		}
		dataPoolMock.SetProofsPool(proofsPool)

		marshaller := &marshal.GogoProtoMarshalizer{}
		st := &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						blockBytes, _ := marshaller.Marshal(&block.HeaderV3{})
						return blockBytes, nil
					},
				}, nil
			},
		}

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"marshalizer":      marshaller,
			"shardCoordinator": mock.NewOneShardCoordinatorMock(),
			"blockTracker": &mock.BlockTrackerMock{
				GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
					return &testscommon.HeaderHandlerStub{}, nil, nil
				},
			},
			"epochStartTrigger": &testscommon.EpochStartTriggerStub{
				GetEpochChangeProposedCalled: func() bool {
					return false
				},
			},
			"dataPool":   dataPoolMock,
			"proofsPool": proofsPool,
			"store":      st,
		})
		require.Nil(t, err)

		err = mp.CheckShardHeadersValidityAndFinalityProposal(metaHeader)
		require.ErrorIs(t, err, process.ErrMissingHeaderProof)
	})
	t.Run("invalid used shard headers, should error", func(t *testing.T) {
		t.Parallel()

		metaHeader := &block.MetaBlockV3{
			ShardInfoProposal: []block.ShardDataProposal{
				{
					HeaderHash: []byte("hash"),
				},
			},
		}

		headersPoolMock := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(_ []byte) (data.HeaderHandler, error) {
				return &block.MetaBlockV3{}, nil
			},
		}
		dataPoolMock := &dataRetrieverMock.PoolsHolderMock{}
		dataPoolMock.SetHeadersPool(headersPoolMock)

		proofsPool := &dataRetrieverMock.ProofsPoolMock{
			HasProofCalled: func(_ uint32, _ []byte) bool {
				return true
			},
		}
		dataPoolMock.SetProofsPool(proofsPool)

		marshaller := &marshal.GogoProtoMarshalizer{}
		st := &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						blockBytes, _ := marshaller.Marshal(&block.HeaderV3{})
						return blockBytes, nil
					},
				}, nil
			},
		}

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"shardCoordinator":    mock.NewOneShardCoordinatorMock(),
			"enableEpochsHandler": &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			"blockTracker": &mock.BlockTrackerMock{
				GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
					return &testscommon.HeaderHandlerStub{}, nil, nil
				},
			},
			"epochStartTrigger": &testscommon.EpochStartTriggerStub{
				GetEpochChangeProposedCalled: func() bool {
					return false
				},
			},
			"dataPool":    dataPoolMock,
			"marshalizer": marshaller,
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{}
				},
			},
			"headerValidator": &integrationTestsMock.HeaderValidatorStub{
				IsHeaderConstructionValidCalled: func(_, _ data.HeaderHandler) error {
					return expectedErr
				},
			},
			"proofsPool": proofsPool,
			"store":      st,
		})
		require.Nil(t, err)

		err = mp.CheckShardHeadersValidityAndFinalityProposal(metaHeader)
		require.ErrorIs(t, err, expectedErr)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		metaHeader := &block.MetaBlockV3{}

		headersPoolMock := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(_ []byte) (data.HeaderHandler, error) {
				return &block.MetaBlockV3{}, nil
			},
		}
		dataPoolMock := &dataRetrieverMock.PoolsHolderMock{}
		dataPoolMock.SetHeadersPool(headersPoolMock)

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"shardCoordinator":    mock.NewOneShardCoordinatorMock(),
			"enableEpochsHandler": &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			"blockTracker": &mock.BlockTrackerMock{
				GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
					return &testscommon.HeaderHandlerStub{}, nil, nil
				},
			},
			"epochStartTrigger": &testscommon.EpochStartTriggerStub{
				GetEpochChangeProposedCalled: func() bool {
					return false
				},
			},
			"dataPool":    dataPoolMock,
			"marshalizer": &marshal.GogoProtoMarshalizer{},
			"proofsPool": &dataRetrieverMock.ProofsPoolMock{
				HasProofCalled: func(_ uint32, _ []byte) bool {
					return true
				},
			},
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{}
				},
			},
			"headerValidator": &integrationTestsMock.HeaderValidatorStub{
				IsHeaderConstructionValidCalled: func(_, _ data.HeaderHandler) error {
					return nil
				},
			},
			"shardInfoCreateData": &processMocks.ShardInfoCreatorMock{
				CreateShardInfoV3Called: func(_ data.MetaHeaderHandler, _ []data.HeaderHandler, _ [][]byte) ([]data.ShardDataProposalHandler, []data.ShardDataHandler, error) {
					return []data.ShardDataProposalHandler{}, []data.ShardDataHandler{}, nil
				},
			},
		})
		require.Nil(t, err)

		err = mp.CheckShardHeadersValidityAndFinalityProposal(metaHeader)
		require.Nil(t, err)
	})
}

func Test_getTxCountExecutionResults(t *testing.T) {
	t.Parallel()

	t.Run("nil meta block", func(t *testing.T) {
		t.Parallel()

		txCount, err := blproc.GetTxCountExecutionResults(nil)
		require.Nil(t, err)
		require.Equal(t, uint32(0), txCount)
	})
	t.Run("no execution results notarized", func(t *testing.T) {
		t.Parallel()

		metaBlock := &block.MetaBlockV3{}
		txCount, err := blproc.GetTxCountExecutionResults(metaBlock)
		require.Nil(t, err)
		require.Equal(t, uint32(0), txCount)
	})
	t.Run("empty execution results notarized", func(t *testing.T) {
		t.Parallel()

		metaBlock := &block.MetaBlockV3{
			ExecutionResults: []*block.MetaExecutionResult{{}, {}},
		}
		txCount, err := blproc.GetTxCountExecutionResults(metaBlock)
		require.Nil(t, err)
		require.Equal(t, uint32(0), txCount)
	})
	t.Run("invalid execution result in notarized list", func(t *testing.T) {
		t.Parallel()

		var metaExecutionResult *block.BaseExecutionResult
		metaBlock := &testscommon.HeaderHandlerStub{
			GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
				return []data.BaseExecutionResultHandler{
					metaExecutionResult,
				}
			},
		}

		txCount, err := blproc.GetTxCountExecutionResults(metaBlock)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
		require.Equal(t, uint32(0), txCount)
	})
	t.Run("execution results notarized", func(t *testing.T) {
		t.Parallel()

		metaBlock := &block.MetaBlockV3{
			ExecutionResults: []*block.MetaExecutionResult{
				{
					ExecutedTxCount: 5,
				},
				{
					ExecutedTxCount: 10,
				},
			},
		}
		txCount, err := blproc.GetTxCountExecutionResults(metaBlock)
		require.Nil(t, err)
		require.Equal(t, uint32(15), txCount)
	})
}

func TestMetaProcessor_hasStartOfEpochExecutionResults(t *testing.T) {
	t.Parallel()

	mbHeaderWithEpochStartData := block.MiniBlockHeader{
		Hash:          []byte("mb hash"),
		SenderShardID: core.MetachainShardId,
		Type:          block.RewardsBlock,
	}
	t.Run("nil meta block", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		hasEpochStartData, err := mp.HasStartOfEpochExecutionResults(nil)
		require.Equal(t, process.ErrNilHeaderHandler, err)
		require.False(t, hasEpochStartData)
	})
	t.Run("no executionResults", func(t *testing.T) {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		validMetaHeaderV3 := &block.MetaBlockV3{}
		hasEpochStartData, err := mp.HasStartOfEpochExecutionResults(validMetaHeaderV3)
		require.Nil(t, err)
		require.False(t, hasEpochStartData)
	})
	t.Run("executionResults with invalid data", func(t *testing.T) {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		validMetaHeaderV3 := &testscommon.HeaderHandlerStub{
			GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
				return []data.BaseExecutionResultHandler{
					&block.BaseExecutionResult{}, // invalid for meta block
				}
			},
		}
		hasEpochStartData, err := mp.HasStartOfEpochExecutionResults(validMetaHeaderV3)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
		require.False(t, hasEpochStartData)
	})
	t.Run("executionResults without epoch start data", func(t *testing.T) {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)
		mbHeader := mbHeaderWithEpochStartData
		mbHeader.Type = block.TxBlock
		validMetaHeaderV3 := &testscommon.HeaderHandlerStub{
			GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
				return []data.BaseExecutionResultHandler{
					&block.MetaExecutionResult{MiniBlockHeaders: []block.MiniBlockHeader{mbHeader}}}
			},
		}

		hasEpochStartData, err := mp.HasStartOfEpochExecutionResults(validMetaHeaderV3)
		require.Nil(t, err)
		require.False(t, hasEpochStartData)
	})
	t.Run("executionResults with reward miniBlocks epoch start data not from meta", func(t *testing.T) {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		mbHeader := mbHeaderWithEpochStartData
		mbHeader.SenderShardID = 0
		validMetaHeaderV3 := &testscommon.HeaderHandlerStub{
			GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
				return []data.BaseExecutionResultHandler{
					&block.MetaExecutionResult{MiniBlockHeaders: []block.MiniBlockHeader{mbHeader}}}
			},
		}

		hasEpochStartData, err := mp.HasStartOfEpochExecutionResults(validMetaHeaderV3)
		require.Nil(t, err)
		require.False(t, hasEpochStartData)
	})
	t.Run("executionResults with peer miniBlocks epoch start data not from meta", func(t *testing.T) {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		mbHeader := mbHeaderWithEpochStartData
		mbHeader.SenderShardID = 0
		mbHeader.Type = block.PeerBlock
		validMetaHeaderV3 := &testscommon.HeaderHandlerStub{
			GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
				return []data.BaseExecutionResultHandler{
					&block.MetaExecutionResult{MiniBlockHeaders: []block.MiniBlockHeader{mbHeader}}}
			},
		}

		hasEpochStartData, err := mp.HasStartOfEpochExecutionResults(validMetaHeaderV3)
		require.Nil(t, err)
		require.False(t, hasEpochStartData)
	})
	t.Run("executionResults with reward miniBlocks epoch start data from meta", func(t *testing.T) {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)
		mbHeader := mbHeaderWithEpochStartData
		mbHeader.Type = block.RewardsBlock
		validMetaHeaderV3 := &testscommon.HeaderHandlerStub{
			GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
				return []data.BaseExecutionResultHandler{
					&block.MetaExecutionResult{MiniBlockHeaders: []block.MiniBlockHeader{mbHeader}}}
			},
		}

		hasEpochStartData, err := mp.HasStartOfEpochExecutionResults(validMetaHeaderV3)
		require.Nil(t, err)
		require.True(t, hasEpochStartData)
	})
	t.Run("executionResults with peer miniBlocks epoch start data from meta", func(t *testing.T) {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)
		mbHeader := mbHeaderWithEpochStartData
		mbHeader.Type = block.PeerBlock
		validMetaHeaderV3 := &testscommon.HeaderHandlerStub{
			GetExecutionResultsHandlersCalled: func() []data.BaseExecutionResultHandler {
				return []data.BaseExecutionResultHandler{
					&block.MetaExecutionResult{MiniBlockHeaders: []block.MiniBlockHeader{mbHeader}}}
			},
		}

		hasEpochStartData, err := mp.HasStartOfEpochExecutionResults(validMetaHeaderV3)
		require.Nil(t, err)
		require.True(t, hasEpochStartData)
	})
}

func Test_hasRewardOrPeerMiniBlocksFromSelf(t *testing.T) {
	t.Parallel()

	t.Run("nil miniBlocks", func(t *testing.T) {
		t.Parallel()
		response := blproc.HasRewardOrPeerMiniBlocksFromMeta(nil)
		require.False(t, response)
	})
	t.Run("no miniBlocks", func(t *testing.T) {
		t.Parallel()
		response := blproc.HasRewardOrPeerMiniBlocksFromMeta([]data.MiniBlockHeaderHandler{})
		require.False(t, response)
	})
	t.Run("with reward miniBlocks from different shard", func(t *testing.T) {
		t.Parallel()
		miniBlocks := []data.MiniBlockHeaderHandler{
			&block.MiniBlockHeader{
				SenderShardID: 1,
				Type:          block.RewardsBlock,
			},
		}
		response := blproc.HasRewardOrPeerMiniBlocksFromMeta(miniBlocks)
		require.False(t, response)
	})
	t.Run("only tx miniBlocks", func(t *testing.T) {
		t.Parallel()
		miniBlocks := []data.MiniBlockHeaderHandler{
			&block.MiniBlockHeader{
				SenderShardID: common.MetachainShardId, // although not possible in combination with txblock
				Type:          block.TxBlock,
			},
		}
		response := blproc.HasRewardOrPeerMiniBlocksFromMeta(miniBlocks)
		require.False(t, response)
	})
	t.Run("with reward miniBlocks from meta shard", func(t *testing.T) {
		t.Parallel()
		miniBlocks := []data.MiniBlockHeaderHandler{
			&block.MiniBlockHeader{
				SenderShardID: common.MetachainShardId,
				Type:          block.RewardsBlock,
			},
		}
		response := blproc.HasRewardOrPeerMiniBlocksFromMeta(miniBlocks)
		require.True(t, response)
	})
	t.Run("with peer miniBlocks from meta shard", func(t *testing.T) {
		t.Parallel()
		miniBlocks := []data.MiniBlockHeaderHandler{
			&block.MiniBlockHeader{
				SenderShardID: common.MetachainShardId,
				Type:          block.PeerBlock,
			},
		}
		response := blproc.HasRewardOrPeerMiniBlocksFromMeta(miniBlocks)
		require.True(t, response)
	})
}

func TestMetaProcessor_createProposalMiniBlocks(t *testing.T) {
	t.Parallel()
	miniblockSelectionSessionNoAdd := &mbSelection.MiniBlockSelectionSessionStub{
		AddMiniBlocksAndHashesCalled: func(miniBlocksAndHashes []block.MiniblockAndHash) error {
			require.Fail(t, "miniBlocksAndHashes should not be called")
			return nil
		},
		AddReferencedHeaderCalled: func(metaBlock data.HeaderHandler, metaBlockHash []byte) {
			require.Fail(t, "AddReferencedHeader should not be called")
		},
		CreateAndAddMiniBlockFromTransactionsCalled: func(txHashes [][]byte) error {
			require.Fail(t, "CreateAndAddMiniBlockFromTransactions should not be called")
			return nil
		},
	}
	t.Run("no time", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = miniblockSelectionSessionNoAdd
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		err = mp.CreateProposalMiniBlocks(10, haveTimeFalse)
		require.Nil(t, err)
	})
	t.Run("with time and error returned by selectIncomingMiniBlocksForProposal", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.BlockTracker = &mock.BlockTrackerMock{
			ComputeLongestShardsChainsFromLastNotarizedCalled: func() ([]data.HeaderHandler, [][]byte, map[uint32][]data.HeaderHandler, error) {
				return nil, nil, nil, expectedErr
			},
		}
		arguments.MiniBlocksSelectionSession = miniblockSelectionSessionNoAdd
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		err = mp.CreateProposalMiniBlocks(10, haveTimeTrue)
		require.Equal(t, expectedErr, err)
	})
	t.Run("with time and no error, no mini blocks/shard headers", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = miniblockSelectionSessionNoAdd
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		err = mp.CreateProposalMiniBlocks(10, haveTimeTrue)
		require.Nil(t, err)
	})
}

func TestMetaProcessor_selectIncomingMiniBlocksForProposal(t *testing.T) {
	t.Parallel()

	t.Run("error from ComputeLongestShardsChainsFromLastNotarized", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.BlockTracker = &mock.BlockTrackerMock{
			ComputeLongestShardsChainsFromLastNotarizedCalled: func() ([]data.HeaderHandler, [][]byte, map[uint32][]data.HeaderHandler, error) {
				return nil, nil, nil, expectedErr
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		err = mp.SelectIncomingMiniBlocksForProposal(10, haveTimeTrue)
		require.Equal(t, expectedErr, err)
	})
	t.Run("error from getLastCrossNotarizedShardHeaders", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.BlockTracker = &mock.BlockTrackerMock{
			ComputeLongestShardsChainsFromLastNotarizedCalled: func() ([]data.HeaderHandler, [][]byte, map[uint32][]data.HeaderHandler, error) {
				return []data.HeaderHandler{}, [][]byte{}, nil, nil
			},
			GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
				return nil, nil, expectedErr
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		err = mp.SelectIncomingMiniBlocksForProposal(10, haveTimeTrue)
		require.Equal(t, expectedErr, err)
	})
	t.Run("selection ok", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		err = mp.SelectIncomingMiniBlocksForProposal(10, haveTimeTrue)
		require.Nil(t, err)
	})
}

func TestMetaProcessor_selectIncomingMiniBlocks(t *testing.T) {
	t.Parallel()

	t.Run("no ordered headers", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			AddMiniBlocksAndHashesCalled: func(miniBlocksAndHashes []block.MiniblockAndHash) error {
				require.Fail(t, "should not be called")
				return nil
			},
			AddReferencedHeaderCalled: func(metaBlock data.HeaderHandler, metaBlockHash []byte) {
				require.Fail(t, "should not be called")
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		lastShardHeaders := createLastShardHeadersNotGenesis()
		var orderedHeaders []data.HeaderHandler
		var orderedHeaderHashes [][]byte

		maxNumHeadersFromSameShard := uint32(2)
		_, err = mp.SelectIncomingMiniBlocks(lastShardHeaders, orderedHeaders, orderedHeaderHashes, maxNumHeadersFromSameShard, haveTimeTrue)
		require.Nil(t, err)
	})

	t.Run("time is up before processing any header", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		// ensure proofs exist but haveTime will stop immediately
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{HasProofCalled: func(shardID uint32, headerHash []byte) bool { return true }}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		addRefCnt := 0
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			AddReferencedHeaderCalled: func(metaBlock data.HeaderHandler, metaBlockHash []byte) {
				addRefCnt++
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		lastShardHeaders := createLastShardHeadersNotGenesis()
		h := &testscommon.HeaderHandlerStub{
			GetShardIDCalled:                 func() uint32 { return 0 },
			GetNonceCalled:                   func() uint64 { return 11 },
			GetMiniBlockHeadersWithDstCalled: func(destId uint32) map[string]uint32 { return map[string]uint32{"x": 1} },
		}
		orderedHeaders := []data.HeaderHandler{h}
		orderedHeaderHashes := [][]byte{[]byte("h1")}

		_, err = mp.SelectIncomingMiniBlocks(lastShardHeaders, orderedHeaders, orderedHeaderHashes, 2, haveTimeFalse)
		require.Nil(t, err)
		require.Equal(t, 0, addRefCnt)
	})

	t.Run("maximum shard headers allowed in one meta block reached (max=0)", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{HasProofCalled: func(shardID uint32, headerHash []byte) bool { return true }}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		called := 0
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			AddReferencedHeaderCalled: func(metaBlock data.HeaderHandler, metaBlockHash []byte) { called++ },
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		lastShardHeaders := createLastShardHeadersNotGenesis()
		h := &testscommon.HeaderHandlerStub{
			GetShardIDCalled:                 func() uint32 { return 0 },
			GetNonceCalled:                   func() uint64 { return 11 },
			GetMiniBlockHeadersWithDstCalled: func(destId uint32) map[string]uint32 { return map[string]uint32{"x": 1} },
		}
		_, err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{[]byte("h1")}, 0, haveTimeTrue)
		require.Nil(t, err)
		require.Equal(t, 0, called)
	})

	t.Run("skip header due to nonce gap", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{HasProofCalled: func(shardID uint32, headerHash []byte) bool { return true }}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		cntAddRef := 0
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			AddReferencedHeaderCalled: func(metaBlock data.HeaderHandler, metaBlockHash []byte) { cntAddRef++ },
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		lastShardHeaders := createLastShardHeadersNotGenesis()
		// last nonce for shard 0 is 10 -> header has 12 so gap > 1 triggers continue
		h := &testscommon.HeaderHandlerStub{
			GetShardIDCalled:                 func() uint32 { return 0 },
			GetNonceCalled:                   func() uint64 { return 12 },
			GetMiniBlockHeadersWithDstCalled: func(destId uint32) map[string]uint32 { return map[string]uint32{"x": 1} },
		}
		_, err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{[]byte("h1")}, 2, haveTimeTrue)
		require.Nil(t, err)
		require.Equal(t, 0, cntAddRef)
	})

	t.Run("skip header due to per-shard limit", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{HasProofCalled: func(shardID uint32, headerHash []byte) bool { return true }}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		cntAddRef := 0
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			AddReferencedHeaderCalled: func(metaBlock data.HeaderHandler, metaBlockHash []byte) { cntAddRef++ },
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		lastShardHeaders := createLastShardHeadersNotGenesis()
		h1 := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 0 }, GetNonceCalled: func() uint64 { return 11 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{} }}
		h2 := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 0 }, GetNonceCalled: func() uint64 { return 12 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{} }}
		_, err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h1, h2}, [][]byte{[]byte("h1"), []byte("h2")}, 1, haveTimeTrue)
		require.Nil(t, err)
		// only first header should be referenced
		require.Equal(t, 1, cntAddRef)
		// last shard header nonce for shard 0 should remain 11 due to per-shard limit preventing second update
		require.Equal(t, uint64(11), lastShardHeaders[0].Header.GetNonce())
	})

	t.Run("skip header due to missing proof", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{HasProofCalled: func(shardID uint32, headerHash []byte) bool { return false }}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		cntAddRef := 0
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			AddReferencedHeaderCalled: func(metaBlock data.HeaderHandler, metaBlockHash []byte) { cntAddRef++ },
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		lastShardHeaders := createLastShardHeadersNotGenesis()
		h := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 0 }, GetNonceCalled: func() uint64 { return 11 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{} }}
		_, err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{[]byte("h1")}, 2, haveTimeTrue)
		require.Nil(t, err)
		require.Equal(t, 0, cntAddRef)
	})

	t.Run("no cross mini blocks with dst me -> add referenced header only", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{HasProofCalled: func(shardID uint32, headerHash []byte) bool { return true }}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		cntAddRef := 0
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			AddReferencedHeaderCalled: func(metaBlock data.HeaderHandler, metaBlockHash []byte) { cntAddRef++ },
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		lastShardHeaders := createLastShardHeadersNotGenesis()
		h := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 0 }, GetNonceCalled: func() uint64 { return 11 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{} }}
		_, err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{[]byte("h1")}, 2, haveTimeTrue)
		require.Nil(t, err)
		require.Equal(t, 1, cntAddRef)
		// last shard header updated and marked used
		require.True(t, lastShardHeaders[0].UsedInBlock)
		require.Equal(t, []byte("h1"), lastShardHeaders[0].Hash)
	})

	t.Run("createMbsCrossShardDstMe returns error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{HasProofCalled: func(shardID uint32, headerHash []byte) bool { return true }}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, []block.MiniblockAndHash, uint32, bool, bool, error) {
				return nil, nil, 0, false, false, expectedErr
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		lastShardHeaders := createLastShardHeadersNotGenesis()
		h := &testscommon.HeaderHandlerStub{
			GetShardIDCalled:                 func() uint32 { return 0 },
			GetNonceCalled:                   func() uint64 { return 11 },
			GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{"mb": 1} },
		}
		_, err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{[]byte("h1")}, 2, haveTimeTrue)
		require.Equal(t, expectedErr, err)
	})

	t.Run("pending mini blocks returned -> break without adding header", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{HasProofCalled: func(shardID uint32, headerHash []byte) bool { return true }}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		cntAddRef := 0
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{AddReferencedHeaderCalled: func(metaBlock data.HeaderHandler, metaBlockHash []byte) { cntAddRef++ }}
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, []block.MiniblockAndHash, uint32, bool, bool, error) {
				return nil, []block.MiniblockAndHash{{}}, 0, false, false, nil
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		lastShardHeaders := createLastShardHeadersNotGenesis()
		h1 := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 0 }, GetNonceCalled: func() uint64 { return 11 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{"mb": 1} }}
		h2 := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 0 }, GetNonceCalled: func() uint64 { return 12 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{"mb": 1} }}
		_, err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h1, h2}, [][]byte{[]byte("h1"), []byte("h2")}, 2, haveTimeTrue)
		require.Nil(t, err)
		require.Equal(t, 0, cntAddRef)
		// ensure second header was not processed due to break after first
		require.Equal(t, uint64(10), lastShardHeaders[0].Header.GetNonce())
	})

	t.Run("success: miniblocks added and header referenced", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{HasProofCalled: func(shardID uint32, headerHash []byte) bool { return true }}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		cntAddRef := 0
		cntAddMbs := 0
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			AddReferencedHeaderCalled:    func(metaBlock data.HeaderHandler, metaBlockHash []byte) { cntAddRef++ },
			AddMiniBlocksAndHashesCalled: func(miniBlocksAndHashes []block.MiniblockAndHash) error { cntAddMbs++; return nil },
		}
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, []block.MiniblockAndHash, uint32, bool, bool, error) {
				return []block.MiniblockAndHash{{}}, nil, 3, true, false, nil
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		lastShardHeaders := createLastShardHeadersNotGenesis()
		h := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 0 }, GetNonceCalled: func() uint64 { return 11 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{"mb": 1} }}
		_, err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{[]byte("h1")}, 2, haveTimeTrue)
		require.Nil(t, err)
		require.Equal(t, 1, cntAddMbs)
		require.Equal(t, 1, cntAddRef)
		// last shard header updated and marked used
		require.True(t, lastShardHeaders[0].UsedInBlock)
		require.Equal(t, []byte("h1"), lastShardHeaders[0].Hash)
	})
}

func TestMetaProcessor_SelectContendedShardHeaders(t *testing.T) {
	t.Parallel()

	parentHash := []byte("parentHash")
	hashLow, hashHigh := []byte("hashLow"), []byte("hashHigh")
	parent := &block.Header{ShardID: 0, Nonce: 10, Round: 10}

	buildProcessorWithProofs := func(supernovaEnabled bool, candidateRounds []uint64, referenced *[][]byte, hasProof func(headerHash []byte) bool) interface {
		SelectContendedShardHeaders(round uint64, lastShardHdrs map[uint32]blproc.ShardHeaderInfo, hdrsAddedForShard map[uint32]uint32, haveTime func() bool) error
	} {
		candidates := make([]data.HeaderHandler, 0, len(candidateRounds))
		candidateHashes := [][]byte{hashLow, hashHigh}[:len(candidateRounds)]
		for _, round := range candidateRounds {
			candidates = append(candidates, &block.Header{ShardID: 0, Nonce: 11, Round: round, PrevHash: parentHash})
		}

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		coreComponents.EnableEpochsHandlerField = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return supernovaEnabled && flag == common.SupernovaFlag
			},
		}
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.HeadersCalled = func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardID uint32) ([]data.HeaderHandler, [][]byte, error) {
						return candidates, candidateHashes, nil
					},
				}
			}
			ph.ProofsCalled = func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{
					HasProofCalled: func(shardID uint32, headerHash []byte) bool { return hasProof(headerHash) },
				}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.HeaderValidator = &processMocks.HeaderValidatorMock{
			IsHeaderConstructionValidCalled: func(currHdr, prevHdr data.HeaderHandler) error { return nil },
		}
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			AddReferencedHeaderCalled: func(header data.HeaderHandler, headerHash []byte) {
				*referenced = append(*referenced, headerHash)
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		return mp
	}

	buildProcessor := func(supernovaEnabled bool, candidateRounds []uint64, referenced *[][]byte) interface {
		SelectContendedShardHeaders(round uint64, lastShardHdrs map[uint32]blproc.ShardHeaderInfo, hdrsAddedForShard map[uint32]uint32, haveTime func() bool) error
	} {
		return buildProcessorWithProofs(supernovaEnabled, candidateRounds, referenced, func(_ []byte) bool { return true })
	}

	newLastShardHdrs := func() map[uint32]blproc.ShardHeaderInfo {
		return map[uint32]blproc.ShardHeaderInfo{0: {Header: parent, Hash: parentHash}}
	}

	t.Run("holds within the discovery window", func(t *testing.T) {
		t.Parallel()

		referenced := make([][]byte, 0)
		mp := buildProcessor(true, []uint64{14, 16}, &referenced)

		// candidate round 14, window 3: rounds 15-16 still hold
		err := mp.SelectContendedShardHeaders(16, newLastShardHdrs(), map[uint32]uint32{}, haveTimeTrue)
		require.Nil(t, err)
		require.Empty(t, referenced)
	})

	t.Run("includes the lowest-round proofed candidate after the window", func(t *testing.T) {
		t.Parallel()

		referenced := make([][]byte, 0)
		mp := buildProcessor(true, []uint64{14, 16}, &referenced)

		lastShardHdrs := newLastShardHdrs()
		err := mp.SelectContendedShardHeaders(17, lastShardHdrs, map[uint32]uint32{}, haveTimeTrue)
		require.Nil(t, err)
		require.Equal(t, [][]byte{hashLow}, referenced)
		require.Equal(t, hashLow, lastShardHdrs[0].Hash)
	})

	t.Run("skips a shard that already progressed", func(t *testing.T) {
		t.Parallel()

		referenced := make([][]byte, 0)
		mp := buildProcessor(true, []uint64{14, 16}, &referenced)

		err := mp.SelectContendedShardHeaders(17, newLastShardHdrs(), map[uint32]uint32{0: 1}, haveTimeTrue)
		require.Nil(t, err)
		require.Empty(t, referenced)
	})

	t.Run("skips an unproofed lower-round candidate", func(t *testing.T) {
		t.Parallel()

		referenced := make([][]byte, 0)
		mp := buildProcessorWithProofs(true, []uint64{14, 16}, &referenced, func(headerHash []byte) bool {
			return !bytes.Equal(headerHash, hashLow)
		})

		lastShardHdrs := newLastShardHdrs()
		err := mp.SelectContendedShardHeaders(19, lastShardHdrs, map[uint32]uint32{}, haveTimeTrue)
		require.Nil(t, err)
		require.Equal(t, [][]byte{hashHigh}, referenced)
	})

	t.Run("skips non-contended candidates", func(t *testing.T) {
		t.Parallel()

		referenced := make([][]byte, 0)
		mp := buildProcessor(true, []uint64{11}, &referenced)

		err := mp.SelectContendedShardHeaders(17, newLastShardHdrs(), map[uint32]uint32{}, haveTimeTrue)
		require.Nil(t, err)
		require.Empty(t, referenced)
	})

	t.Run("does nothing without supernova", func(t *testing.T) {
		t.Parallel()

		referenced := make([][]byte, 0)
		mp := buildProcessor(false, []uint64{14}, &referenced)

		err := mp.SelectContendedShardHeaders(17, newLastShardHdrs(), map[uint32]uint32{}, haveTimeTrue)
		require.Nil(t, err)
		require.Empty(t, referenced)
	})
}

func TestMetaProcessor_selectIncomingMiniBlocks_GapsAndDuplicates(t *testing.T) {
	t.Parallel()

	// helper to build a MetaProcessor with proofs pool behavior
	type metaSel interface {
		SelectIncomingMiniBlocks(lastShardHdr map[uint32]blproc.ShardHeaderInfo, orderedHdrs []data.HeaderHandler, orderedHdrsHashes [][]byte, maxNumHeadersFromSameShard uint32, haveTime func() bool) (map[uint32]uint32, error)
	}
	buildMp := func(hasProofFn func(shardID uint32, headerHash []byte) bool) metaSel {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{HasProofCalled: hasProofFn}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)
		return mp
	}

	t.Run("inconsistent ordered headers and hashes lengths -> error", func(t *testing.T) {
		t.Parallel()

		mp := buildMp(func(uint32, []byte) bool { return true })
		lastShardHeaders := createLastShardHeadersNotGenesis()
		h := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 0 }, GetNonceCalled: func() uint64 { return 11 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{} }}
		_, err := mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{}, 2, haveTimeTrue)
		require.Equal(t, process.ErrInconsistentShardHeadersAndHashes, err)
	})

	t.Run("missing last shard header for ordered header -> error", func(t *testing.T) {
		t.Parallel()

		mp := buildMp(func(uint32, []byte) bool { return true })
		lastShardHeaders := createLastShardHeadersNotGenesis()
		// header from shard 99, not present in lastShardHeaders map
		h := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 99 }, GetNonceCalled: func() uint64 { return 1 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{} }}
		_, err := mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{[]byte("h1")}, 2, haveTimeTrue)
		require.Equal(t, process.ErrMissingHeader, err)
	})

	t.Run("duplicate nonce: first has proof accepted, second skipped", func(t *testing.T) {
		t.Parallel()

		cntAddRef := 0
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{HasProofCalled: func(uint32, []byte) bool { return true }}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			AddReferencedHeaderCalled: func(metaBlock data.HeaderHandler, metaBlockHash []byte) { cntAddRef++ },
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		lastShardHeaders := createLastShardHeadersNotGenesis()
		h1 := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 0 }, GetNonceCalled: func() uint64 { return 11 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{} }}
		h2 := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 0 }, GetNonceCalled: func() uint64 { return 11 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{} }}
		_, err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h1, h2}, [][]byte{[]byte("h1"), []byte("h2")}, 2, haveTimeTrue)
		require.Nil(t, err)
		require.Equal(t, 1, cntAddRef)
		// last shard header updated to first hash and used
		require.True(t, lastShardHeaders[0].UsedInBlock)
		require.Equal(t, []byte("h1"), lastShardHeaders[0].Hash)
	})

	t.Run("duplicate nonce: first missing proof skipped, second with proof accepted", func(t *testing.T) {
		t.Parallel()

		cntAddRef := 0
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{HasProofCalled: func(_ uint32, hash []byte) bool { return string(hash) == "h2" }}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			AddReferencedHeaderCalled: func(metaBlock data.HeaderHandler, metaBlockHash []byte) { cntAddRef++ },
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		lastShardHeaders := createLastShardHeadersNotGenesis()
		h1 := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 0 }, GetNonceCalled: func() uint64 { return 11 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{} }}
		h2 := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 0 }, GetNonceCalled: func() uint64 { return 11 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{} }}
		_, err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h1, h2}, [][]byte{[]byte("h1"), []byte("h2")}, 2, haveTimeTrue)
		require.Nil(t, err)
		require.Equal(t, 1, cntAddRef)
		// last shard header updated to second hash and used
		require.True(t, lastShardHeaders[0].UsedInBlock)
		require.Equal(t, []byte("h2"), lastShardHeaders[0].Hash)
	})
}

func TestMetaProcessor_hasExecutionResultsForProposedEpochChange(t *testing.T) {
	t.Parallel()

	t.Run("should error because of GetHeaderByHash", func(t *testing.T) {
		t.Parallel()

		metaHeader := &block.MetaBlockV3{
			ExecutionResults: []*block.MetaExecutionResult{
				{
					ExecutionResult: &block.BaseMetaExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderHash: []byte("headerHash1"),
						},
					},
				},
			},
		}

		headersPoolMock := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				return nil, expectedErr
			},
		}

		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return headersPoolMock
		}

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.DataPool = dataPool

		dataComponents.Storage = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						return nil, expectedErr
					},
				}, nil
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.HasExecutionResultsForProposedEpochChange(metaHeader)
		require.ErrorIs(t, err, process.ErrMissingHeader)
	})

	t.Run("should return ErrStartOfEpochExecutionResultsDoNotExist", func(t *testing.T) {
		t.Parallel()

		metaHeader := &block.MetaBlockV3{
			ExecutionResults: []*block.MetaExecutionResult{
				{
					ExecutionResult: &block.BaseMetaExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderHash: []byte("headerHash0"),
						},
					},
				},
				{
					ExecutionResult: &block.BaseMetaExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderHash: []byte("headerHash1"),
						},
					},
				},
			},
		}

		headersPoolMock := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, []byte("headerHash1")) {
					return &block.MetaBlockV3{
						EpochChangeProposed: true,
					}, nil
				}
				return &block.MetaBlockV3{
					EpochChangeProposed: false,
				}, nil
			},
		}

		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return headersPoolMock
		}

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.DataPool = dataPool

		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.HasExecutionResultsForProposedEpochChange(metaHeader)
		require.Equal(t, process.ErrStartOfEpochExecutionResultsDoNotExist, err)
	})

	t.Run("should find header with epoch change proposal", func(t *testing.T) {
		t.Parallel()

		metaHeader := &block.MetaBlockV3{
			ExecutionResults: []*block.MetaExecutionResult{
				{
					ExecutionResult: &block.BaseMetaExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderHash: []byte("headerHash0"),
						},
					},
				},
				{
					ExecutionResult: &block.BaseMetaExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderHash: []byte("headerHash1"),
						},
					},
					MiniBlockHeaders: []block.MiniBlockHeader{
						{
							SenderShardID: common.MetachainShardId,
							Type:          block.RewardsBlock,
						},
					},
				},
			},
		}

		headersPoolMock := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, []byte("headerHash1")) {
					return &block.MetaBlockV3{
						EpochChangeProposed: true,
					}, nil
				}
				return &block.MetaBlockV3{
					EpochChangeProposed: false,
				}, nil
			},
		}

		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return headersPoolMock
		}

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.DataPool = dataPool

		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		proposedChange, err := mp.HasExecutionResultsForProposedEpochChange(metaHeader)
		require.Nil(t, err)
		require.True(t, proposedChange)
	})
}

func TestMetaProcessor_checkEpochCorrectnessV3(t *testing.T) {
	t.Parallel()

	executionResults := []*block.MetaExecutionResult{
		{
			ExecutionResult: &block.BaseMetaExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash: []byte("headerHash0"),
				},
			},
		},
		{
			ExecutionResult: &block.BaseMetaExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash: []byte("headerHash1"),
				},
			},
			MiniBlockHeaders: []block.MiniBlockHeader{
				{
					SenderShardID: common.MetachainShardId,
					Type:          block.RewardsBlock,
				},
			},
		},
	}

	t.Run("should return nil current header", func(t *testing.T) {
		t.Parallel()

		metaHeader := &block.MetaBlockV3{
			ExecutionResults: []*block.MetaExecutionResult{
				{
					ExecutionResult: &block.BaseMetaExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderHash: []byte("headerHash1"),
						},
					},
				},
			},
		}

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return nil
				},
			},
		})
		require.Nil(t, err)

		err = mp.CheckEpochCorrectnessV3(metaHeader)
		require.Nil(t, err)
	})

	t.Run("should error ErrNilHeaderHandler", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{}
				},
			},
		})
		require.Nil(t, err)

		err = mp.CheckEpochCorrectnessV3(nil)
		require.Equal(t, process.ErrNilHeaderHandler, err)
	})

	t.Run("should return error hasExecutionResultsForProposedEpochChange", func(t *testing.T) {
		t.Parallel()

		metaHeader := &block.MetaBlockV3{
			Epoch: 2,
			ExecutionResults: []*block.MetaExecutionResult{
				{
					ExecutionResult: &block.BaseMetaExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderHash: []byte("headerHash0"),
						},
					},
				},
				{
					ExecutionResult: &block.BaseMetaExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderHash: []byte("headerHash1"),
						},
					},
					MiniBlockHeaders: []block.MiniBlockHeader{
						{
							SenderShardID: common.MetachainShardId,
							Type:          block.RewardsBlock,
						},
					},
				},
			},
		}

		headersPoolMock := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				return nil, expectedErr
			},
		}
		dataPoolMock := &dataRetrieverMock.PoolsHolderMock{}
		dataPoolMock.SetHeadersPool(headersPoolMock)

		marshaller := &marshal.GogoProtoMarshalizer{}
		st := &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						blockBytes, _ := marshaller.Marshal(&block.MetaBlockV3{})
						return blockBytes, expectedErr
					},
				}, nil
			},
		}

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"dataPool": dataPoolMock,
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						Epoch: 1,
					}
				},
			},
			"marshalizer": &marshal.GogoProtoMarshalizer{},
			"store":       st,
			"epochStartTrigger": &testscommon.EpochStartTriggerStub{
				EpochCalled: func() uint32 {
					return 1
				},
				ShouldProposeEpochChangeCalled: func(round uint64, nonce uint64) bool {
					return false
				},
			},
		})
		require.Nil(t, err)

		err = mp.CheckEpochCorrectnessV3(metaHeader)
		require.ErrorIs(t, err, process.ErrMissingHeader)
	})

	t.Run("should return error ErrEpochDoesNotMatch because of incomplete data", func(t *testing.T) {
		t.Parallel()

		metaHeader := &block.MetaBlockV3{
			ExecutionResults: executionResults,
		}

		headersPoolMock := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, []byte("headerHash1")) {
					return &block.MetaBlockV3{
						EpochChangeProposed: true,
					}, nil
				}
				return &block.MetaBlockV3{
					EpochChangeProposed: false,
				}, nil
			},
		}
		dataPoolMock := &dataRetrieverMock.PoolsHolderMock{}
		dataPoolMock.SetHeadersPool(headersPoolMock)

		marshaller := &marshal.GogoProtoMarshalizer{}
		st := &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						blockBytes, _ := marshaller.Marshal(&block.MetaBlockV3{})
						return blockBytes, nil
					},
				}, nil
			},
		}

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"dataPool": dataPoolMock,
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{}
				},
			},
			"epochStartTrigger": &testscommon.EpochStartTriggerStub{EpochCalled: func() uint32 {
				return 1
			}},
			"marshalizer": marshaller,
			"store":       st,
		})
		require.Nil(t, err)

		err = mp.CheckEpochCorrectnessV3(metaHeader)
		require.Equal(t, process.ErrEpochDoesNotMatch, err)
	})

	t.Run("should return error ErrEpochDoesNotMatch because of no epoch start results", func(t *testing.T) {
		t.Parallel()

		metaHeader := &block.MetaBlockV3{
			Epoch:      2,
			EpochStart: block.EpochStart{},
			ExecutionResults: []*block.MetaExecutionResult{
				{
					ExecutionResult: &block.BaseMetaExecutionResult{
						BaseExecutionResult: &block.BaseExecutionResult{
							HeaderHash: []byte("headerHash0"),
						},
					},
				},
			},
		}

		headersPoolMock := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, []byte("headerHash1")) {
					return &block.MetaBlockV3{
						EpochChangeProposed: true,
					}, nil
				}
				return &block.MetaBlockV3{
					EpochChangeProposed: false,
				}, nil
			},
		}
		dataPoolMock := &dataRetrieverMock.PoolsHolderMock{}
		dataPoolMock.SetHeadersPool(headersPoolMock)

		marshaller := &marshal.GogoProtoMarshalizer{}
		st := &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						blockBytes, _ := marshaller.Marshal(&block.MetaBlockV3{})
						return blockBytes, nil
					},
				}, nil
			},
		}

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"dataPool": dataPoolMock,
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						Epoch: 1,
					}
				},
			},
			"epochStartTrigger": &testscommon.EpochStartTriggerStub{EpochCalled: func() uint32 {
				return 1
			}},
			"marshalizer": marshaller,
			"store":       st,
		})
		require.Nil(t, err)

		err = mp.CheckEpochCorrectnessV3(metaHeader)
		require.Equal(t, process.ErrEpochDoesNotMatch, err)
	})

	t.Run("should return error ErrEpochDoesNotMatch because of epoch not changed", func(t *testing.T) {
		t.Parallel()

		epochStartData := block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{}, {},
			},
		}

		metaHeader := &block.MetaBlockV3{
			Epoch:            1,
			EpochStart:       epochStartData,
			ExecutionResults: executionResults,
		}

		headersPoolMock := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, []byte("headerHash1")) {
					return &block.MetaBlockV3{
						EpochChangeProposed: true,
					}, nil
				}
				return &block.MetaBlockV3{
					EpochChangeProposed: false,
				}, nil
			},
		}
		dataPoolMock := &dataRetrieverMock.PoolsHolderMock{}
		dataPoolMock.SetHeadersPool(headersPoolMock)

		marshaller := &marshal.GogoProtoMarshalizer{}
		st := &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						blockBytes, _ := marshaller.Marshal(&block.MetaBlockV3{})
						return blockBytes, nil
					},
				}, nil
			},
		}

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"dataPool": dataPoolMock,
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						Epoch: 1,
					}
				},
			},
			"epochStartTrigger": &testscommon.EpochStartTriggerStub{EpochCalled: func() uint32 {
				return 1
			}},
			"store":       st,
			"marshalizer": marshaller,
		})
		require.Nil(t, err)
		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{
			EpochStartData: &epochStartData,
		})

		err = mp.CheckEpochCorrectnessV3(metaHeader)
		require.Equal(t, process.ErrEpochDoesNotMatch, err)
	})

	t.Run("should return error ErrEpochDoesNotMatch because of epoch is discontinuous", func(t *testing.T) {
		t.Parallel()

		epochStartData := block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{}, {},
			},
		}

		metaHeader := &block.MetaBlockV3{
			Epoch:            3,
			EpochStart:       epochStartData,
			ExecutionResults: executionResults,
		}

		headersPoolMock := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, []byte("headerHash1")) {
					return &block.MetaBlockV3{
						EpochChangeProposed: true,
					}, nil
				}
				return &block.MetaBlockV3{
					EpochChangeProposed: false,
				}, nil
			},
		}
		dataPoolMock := &dataRetrieverMock.PoolsHolderMock{}
		dataPoolMock.SetHeadersPool(headersPoolMock)

		marshaller := &marshal.GogoProtoMarshalizer{}
		st := &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						blockBytes, _ := marshaller.Marshal(&block.MetaBlockV3{})
						return blockBytes, nil
					},
				}, nil
			},
		}

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"dataPool": dataPoolMock,
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						Epoch: 1,
					}
				},
			},
			"epochStartTrigger": &testscommon.EpochStartTriggerStub{EpochCalled: func() uint32 {
				return 1
			}},
			"marshalizer": &marshal.GogoProtoMarshalizer{},
			"store":       st,
		})
		require.Nil(t, err)
		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{
			EpochStartData: &epochStartData,
		})

		err = mp.CheckEpochCorrectnessV3(metaHeader)
		require.Equal(t, process.ErrEpochDoesNotMatch, err)
	})
	t.Run("not equal epoch start data should error", func(t *testing.T) {
		t.Parallel()

		epochStartData := block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{}, {},
			},
		}
		epochStartDataFromMetaProcessor := block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{}, {}, {},
			},
		}

		metaHeader := &block.MetaBlockV3{
			Epoch:            2,
			EpochStart:       epochStartData,
			ExecutionResults: executionResults,
		}

		headersPoolMock := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, []byte("headerHash1")) {
					return &block.MetaBlockV3{
						EpochChangeProposed: true,
					}, nil
				}
				return &block.MetaBlockV3{
					EpochChangeProposed: false,
				}, nil
			},
		}
		dataPoolMock := &dataRetrieverMock.PoolsHolderMock{}
		dataPoolMock.SetHeadersPool(headersPoolMock)

		marshaller := &marshal.GogoProtoMarshalizer{}
		st := &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						blockBytes, _ := marshaller.Marshal(&block.MetaBlockV3{})
						return blockBytes, nil
					},
				}, nil
			},
		}

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"dataPool": dataPoolMock,
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						Epoch: 1,
					}
				},
			},
			"epochStartTrigger": &testscommon.EpochStartTriggerStub{EpochCalled: func() uint32 {
				return 1
			}},
			"store":       st,
			"marshalizer": &marshal.GogoProtoMarshalizer{},
		})
		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{
			EpochStartData: &epochStartDataFromMetaProcessor,
		})
		require.Nil(t, err)

		err = mp.CheckEpochCorrectnessV3(metaHeader)
		require.Equal(t, process.ErrEpochDoesNotMatch, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		epochStartData := block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{}, {},
			},
		}
		metaHeader := &block.MetaBlockV3{
			Epoch:            2,
			EpochStart:       epochStartData,
			ExecutionResults: executionResults,
		}

		headersPoolMock := &pool.HeadersPoolStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, []byte("headerHash1")) {
					return &block.MetaBlockV3{
						EpochChangeProposed: true,
					}, nil
				}
				return &block.MetaBlockV3{
					EpochChangeProposed: false,
				}, nil
			},
		}
		dataPoolMock := &dataRetrieverMock.PoolsHolderMock{}
		dataPoolMock.SetHeadersPool(headersPoolMock)

		marshaller := &marshal.GogoProtoMarshalizer{}
		st := &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						blockBytes, _ := marshaller.Marshal(&block.MetaBlockV3{})
						return blockBytes, nil
					},
				}, nil
			},
		}

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"dataPool":    dataPoolMock,
			"marshalizer": &marshal.GogoProtoMarshalizer{},
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						Epoch: 1,
					}
				},
			},
			"epochStartTrigger": &testscommon.EpochStartTriggerStub{EpochCalled: func() uint32 {
				return 1
			}},
			"store": st,
		})
		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{
			Epoch: 2,
			EpochStartData: &block.EpochStart{
				LastFinalizedHeaders: epochStartData.LastFinalizedHeaders,
			},
		})
		require.Nil(t, err)

		err = mp.CheckEpochCorrectnessV3(metaHeader)
		require.Nil(t, err)
	})
}

func TestMetaProcessor_checkShardInfoValidity(t *testing.T) {
	t.Parallel()

	t.Run("should return error from CreateShardInfoV3", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"shardInfoCreateData": &processMocks.ShardInfoCreatorMock{
				CreateShardInfoV3Called: func(metaHeader data.MetaHeaderHandler, shardHeaders []data.HeaderHandler, shardHeaderHashes [][]byte) ([]data.ShardDataProposalHandler, []data.ShardDataHandler, error) {
					return nil, nil, expectedErr
				},
			},
		})
		require.Nil(t, err)

		err = mp.CheckShardInfoValidity(nil, &blproc.UsedShardHeadersInfo{})
		require.ErrorContains(t, err, expectedErr.Error())
	})

	t.Run("should return ErrHeaderShardDataMismatch error", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"shardInfoCreateData": &processMocks.ShardInfoCreatorMock{
				CreateShardInfoV3Called: func(metaHeader data.MetaHeaderHandler, shardHeaders []data.HeaderHandler, shardHeaderHashes [][]byte) ([]data.ShardDataProposalHandler, []data.ShardDataHandler, error) {
					return nil, []data.ShardDataHandler{
						&block.ShardData{},
					}, nil
				},
			},
		})
		require.Nil(t, err)

		err = mp.CheckShardInfoValidity(&block.MetaBlockV3{
			ShardInfo: []block.ShardData{},
		}, &blproc.UsedShardHeadersInfo{})

		require.Equal(t, process.ErrHeaderShardDataMismatch, err)
	})

	t.Run("should return ErrHeaderShardDataMismatch error because of createdShardInfo", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"shardInfoCreateData": &processMocks.ShardInfoCreatorMock{
				CreateShardInfoV3Called: func(metaHeader data.MetaHeaderHandler, shardHeaders []data.HeaderHandler, shardHeaderHashes [][]byte) ([]data.ShardDataProposalHandler, []data.ShardDataHandler, error) {
					return nil, []data.ShardDataHandler{
						&block.ShardData{
							Nonce: 0,
						},
					}, nil
				},
			},
		})
		require.Nil(t, err)

		err = mp.CheckShardInfoValidity(&block.MetaBlockV3{
			ShardInfo: []block.ShardData{
				{
					Nonce: 2,
				},
			},
		}, &blproc.UsedShardHeadersInfo{})

		require.ErrorContains(t, err, process.ErrHeaderShardDataMismatch.Error())
	})

	t.Run("should return ErrHeaderShardDataMismatch error because of createdShardInfoProposal", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"shardInfoCreateData": &processMocks.ShardInfoCreatorMock{
				CreateShardInfoV3Called: func(metaHeader data.MetaHeaderHandler, shardHeaders []data.HeaderHandler, shardHeaderHashes [][]byte) ([]data.ShardDataProposalHandler, []data.ShardDataHandler, error) {
					return []data.ShardDataProposalHandler{
							&block.ShardDataProposal{
								Nonce: 0,
							},
						}, []data.ShardDataHandler{
							&block.ShardData{
								Nonce: 0,
							},
						}, nil
				},
			},
		})
		require.Nil(t, err)

		err = mp.CheckShardInfoValidity(&block.MetaBlockV3{
			ShardInfo: []block.ShardData{
				{
					Nonce: 0,
				},
			},
			ShardInfoProposal: []block.ShardDataProposal{
				{
					Nonce: 2,
				},
			},
		}, &blproc.UsedShardHeadersInfo{})

		require.ErrorContains(t, err, process.ErrHeaderShardDataMismatch.Error())
	})
}

func TestMetaProcessor_checkHeadersSequenceCorrectness(t *testing.T) {
	t.Parallel()

	t.Run("should return error from IsHeaderConstructionValid", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"headerValidator": &processMocks.HeaderValidatorMock{
				IsHeaderConstructionValidCalled: func(currHdr, prevHdr data.HeaderHandler) error {
					return expectedErr
				},
			},
			"enableEpochsHandler": &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			"blockTracker":        &integrationTestsMock.BlockTrackerStub{},
		})
		require.Nil(t, err)

		err = mp.CheckHeadersSequenceCorrectness([]blproc.ShardHeaderInfo{
			{
				Header: &block.Header{Nonce: 2},
			},
		}, blproc.ShardHeaderInfo{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work for contended unsettled header without a better competitor", func(t *testing.T) {
		t.Parallel()

		contendedHash := []byte("contended hash")
		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"headerValidator": &processMocks.HeaderValidatorMock{
				IsHeaderConstructionValidCalled: func(currHdr, prevHdr data.HeaderHandler) error {
					return nil
				},
			},
			"enableEpochsHandler": &enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					return flag == common.SupernovaFlag
				},
			},
			"blockTracker": &integrationTestsMock.BlockTrackerStub{
				IsSettledCrossHeaderCalled: func(header data.HeaderHandler, headerHash []byte) bool {
					return false
				},
			},
			"dataPool": &dataRetrieverMock.PoolsHolderStub{
				ProofsCalled: func() dataRetriever.ProofsPool {
					return &dataRetrieverMock.ProofsPoolMock{}
				},
				HeadersCalled: func() dataRetriever.HeadersPool {
					return &mock.HeadersCacherStub{}
				},
			},
		})
		require.Nil(t, err)

		err = mp.CheckHeadersSequenceCorrectness([]blproc.ShardHeaderInfo{
			{
				// rounds 2-4 skipped after the last notarized shard header
				Header: &block.Header{Nonce: 2, Round: 5},
				Hash:   contendedHash,
			},
		}, blproc.ShardHeaderInfo{
			Header: &block.Header{Nonce: 1, Round: 1},
		})
		require.Nil(t, err)
	})

	t.Run("should return error for contended unsettled header with a better proofed competitor", func(t *testing.T) {
		t.Parallel()

		contendedHash := []byte("contended hash")
		parentHash := []byte("parent hash")
		competitorHash := []byte("competitor hash")
		competitorHdr := &block.Header{Nonce: 2, Round: 2, PrevHash: parentHash}
		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"headerValidator": &processMocks.HeaderValidatorMock{
				IsHeaderConstructionValidCalled: func(currHdr, prevHdr data.HeaderHandler) error {
					return nil
				},
			},
			"enableEpochsHandler": &enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					return flag == common.SupernovaFlag
				},
			},
			"blockTracker": &integrationTestsMock.BlockTrackerStub{
				IsSettledCrossHeaderCalled: func(header data.HeaderHandler, headerHash []byte) bool {
					return false
				},
			},
			"dataPool": &dataRetrieverMock.PoolsHolderStub{
				ProofsCalled: func() dataRetriever.ProofsPool {
					return &dataRetrieverMock.ProofsPoolMock{}
				},
				HeadersCalled: func() dataRetriever.HeadersPool {
					return &mock.HeadersCacherStub{
						GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardID uint32) ([]data.HeaderHandler, [][]byte, error) {
							return []data.HeaderHandler{competitorHdr}, [][]byte{competitorHash}, nil
						},
					}
				},
			},
			"proofsPool": &dataRetrieverMock.ProofsPoolMock{
				HasProofCalled: func(shardID uint32, headerHash []byte) bool {
					return string(headerHash) == string(competitorHash)
				},
			},
		})
		require.Nil(t, err)

		err = mp.CheckHeadersSequenceCorrectness([]blproc.ShardHeaderInfo{
			{
				// rounds 2-4 skipped after the last notarized shard header
				Header: &block.Header{Nonce: 2, Round: 5, PrevHash: parentHash},
				Hash:   contendedHash,
			},
		}, blproc.ShardHeaderInfo{
			Header: &block.Header{Nonce: 1, Round: 1},
			Hash:   parentHash,
		})
		require.ErrorContains(t, err, "better proofed competitor")
	})

	t.Run("should work for contended settled header", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"headerValidator": &processMocks.HeaderValidatorMock{
				IsHeaderConstructionValidCalled: func(currHdr, prevHdr data.HeaderHandler) error {
					return nil
				},
			},
			"enableEpochsHandler": &enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					return flag == common.SupernovaFlag
				},
			},
			"blockTracker": &integrationTestsMock.BlockTrackerStub{
				IsSettledCrossHeaderCalled: func(header data.HeaderHandler, headerHash []byte) bool {
					return true
				},
			},
		})
		require.Nil(t, err)

		err = mp.CheckHeadersSequenceCorrectness([]blproc.ShardHeaderInfo{
			{
				Header: &block.Header{Nonce: 2, Round: 5},
				Hash:   []byte("contended settled hash"),
			},
		}, blproc.ShardHeaderInfo{
			Header: &block.Header{Nonce: 1, Round: 1},
		})
		require.Nil(t, err)
	})

	t.Run("should work for non-contended header without settlement lookup", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"headerValidator": &processMocks.HeaderValidatorMock{
				IsHeaderConstructionValidCalled: func(currHdr, prevHdr data.HeaderHandler) error {
					return nil
				},
			},
			"enableEpochsHandler": &enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
					return flag == common.SupernovaFlag
				},
			},
			"blockTracker": &integrationTestsMock.BlockTrackerStub{
				IsSettledCrossHeaderCalled: func(header data.HeaderHandler, headerHash []byte) bool {
					require.Fail(t, "settlement must not be checked on the clean path")
					return false
				},
			},
		})
		require.Nil(t, err)

		err = mp.CheckHeadersSequenceCorrectness([]blproc.ShardHeaderInfo{
			{
				Header: &block.Header{Nonce: 2, Round: 2},
				Hash:   []byte("clean hash"),
			},
		}, blproc.ShardHeaderInfo{
			Header: &block.Header{Nonce: 1, Round: 1},
		})
		require.Nil(t, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"headerValidator": &processMocks.HeaderValidatorMock{
				IsHeaderConstructionValidCalled: func(currHdr, prevHdr data.HeaderHandler) error {
					return nil
				},
			},
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return nil
				},
			},
			"enableEpochsHandler": &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			"blockTracker":        &integrationTestsMock.BlockTrackerStub{},
		})
		require.Nil(t, err)

		err = mp.CheckHeadersSequenceCorrectness([]blproc.ShardHeaderInfo{
			{
				Header: &block.Header{Nonce: 0},
			},
			{
				Header: &block.Header{Nonce: 1},
			},
		}, blproc.ShardHeaderInfo{
			Header: &block.Header{Nonce: 0},
		})
		require.Nil(t, err)
	})
}

func TestMetaProcessor_VerifyEpochStartData(t *testing.T) {
	t.Parallel()

	t.Run("same epoch start data, should return true", func(t *testing.T) {
		t.Parallel()

		lastFinalizedData := []block.EpochStartShardData{
			{
				ShardID:    1,
				Epoch:      1,
				Nonce:      1,
				HeaderHash: []byte("headerHash1"),
			},
		}

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.EpochStartDataCreator = &mock.EpochStartDataCreatorStub{
			CreateEpochStartShardDataMetablockV3Called: func(metablock data.MetaHeaderHandler) ([]block.EpochStartShardData, error) {
				return lastFinalizedData, nil
			},
		}

		mp, _ := blproc.NewMetaProcessor(arguments)
		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{
			Epoch: 1,
			EpochStartData: &block.EpochStart{
				LastFinalizedHeaders: lastFinalizedData,
				Economics:            block.Economics{},
			},
		})

		epochStartData := &block.EpochStart{
			LastFinalizedHeaders: lastFinalizedData,
		}
		metaHeader := &block.MetaBlockV3{
			Epoch:      1,
			EpochStart: *epochStartData,
		}

		ok := mp.VerifyEpochStartData(metaHeader)
		require.True(t, ok)
	})

	t.Run("different epoch start data, should return false", func(t *testing.T) {
		t.Parallel()

		lastFinalizedData := []block.EpochStartShardData{
			{
				ShardID:    1,
				Epoch:      1,
				Nonce:      1,
				HeaderHash: []byte("headerHash1"),
			},
		}

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.EpochStartDataCreator = &mock.EpochStartDataCreatorStub{
			CreateEpochStartShardDataMetablockV3Called: func(metablock data.MetaHeaderHandler) ([]block.EpochStartShardData, error) {
				return lastFinalizedData, nil
			},
		}

		mp, _ := blproc.NewMetaProcessor(arguments)

		lastFinalizedData2 := []block.EpochStartShardData{
			{
				ShardID:    2,
				Epoch:      2,
				Nonce:      2,
				HeaderHash: []byte("headerHash2"),
			},
		}
		epochStartData := &block.EpochStart{
			LastFinalizedHeaders: lastFinalizedData2,
		}
		metaHeader := &block.MetaBlockV3{
			Epoch:      3,
			EpochStart: *epochStartData,
		}

		ok := mp.VerifyEpochStartData(metaHeader)
		require.False(t, ok)
	})
}

func TestMetaProcessor_processIfFirstBlockAfterEpochStartBlockV3(t *testing.T) {
	t.Parallel()

	t.Run("should return ErrWrongTypeAssertion error because of nil previous executed block", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain.SetLastExecutedBlockHeaderAndRootHash(&block.HeaderV3{}, nil, nil)
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		err = mp.ProcessIfFirstBlockAfterEpochStartBlockV3()
		require.Equal(t, common.ErrWrongTypeAssertion, err)
	})

	t.Run("should return nil because it is not start of epoch block", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"blockChain": &testscommon.ChainHandlerStub{
				GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						EpochStart: block.EpochStart{
							LastFinalizedHeaders: nil,
						},
					}
				},
			},
		})
		require.Nil(t, err)

		err = mp.ProcessIfFirstBlockAfterEpochStartBlockV3()
		require.Nil(t, err)
	})

	t.Run("if SaveNodesCoordinatorUpdates fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"blockChain": &testscommon.ChainHandlerStub{
				GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						Epoch: 2,
						EpochStart: block.EpochStart{
							LastFinalizedHeaders: []block.EpochStartShardData{
								{}, {}, {},
							},
						},
					}
				},
			},
			"validatorStatisticsProcessor": &testscommon.ValidatorStatisticsProcessorStub{
				SaveNodesCoordinatorUpdatesCalled: func(epoch uint32) (bool, error) {
					return false, expectedErr
				},
			},
		})
		require.Nil(t, err)

		err = mp.ProcessIfFirstBlockAfterEpochStartBlockV3()
		require.Equal(t, expectedErr, err)
	})

	t.Run("if ToggleUnStakeUnBondCalled fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"blockChain": &testscommon.ChainHandlerStub{
				GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						Epoch: 2,
						EpochStart: block.EpochStart{
							LastFinalizedHeaders: []block.EpochStartShardData{
								{}, {}, {},
							},
						},
					}
				},
			},
			"validatorStatisticsProcessor": &testscommon.ValidatorStatisticsProcessorStub{
				SaveNodesCoordinatorUpdatesCalled: func(epoch uint32) (bool, error) {
					return true, nil
				},
			},
			"epochSystemSCProcessor": &testscommon.EpochStartSystemSCStub{
				ToggleUnStakeUnBondCalled: func(value bool) error {
					return expectedErr
				},
			},
		})
		require.Nil(t, err)

		err = mp.ProcessIfFirstBlockAfterEpochStartBlockV3()
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"blockChain": &testscommon.ChainHandlerStub{
				GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{
						Epoch: 2,
						EpochStart: block.EpochStart{
							LastFinalizedHeaders: []block.EpochStartShardData{
								{}, {}, {},
							},
						},
					}
				},
			},
			"validatorStatisticsProcessor": &testscommon.ValidatorStatisticsProcessorStub{
				SaveNodesCoordinatorUpdatesCalled: func(epoch uint32) (bool, error) {
					return true, nil
				},
			},
			"epochSystemSCProcessor": &testscommon.EpochStartSystemSCStub{
				ToggleUnStakeUnBondCalled: func(value bool) error {
					return nil
				},
			},
		})
		require.Nil(t, err)

		err = mp.ProcessIfFirstBlockAfterEpochStartBlockV3()
		require.Nil(t, err)
	})
}

func TestMetaProcessor_processEpochStartProposeBlock(t *testing.T) {
	t.Parallel()

	defaultMetaBlockV3 := block.MetaBlockV3{
		LastExecutionResult: &block.MetaExecutionResultInfo{
			ExecutionResult: &block.BaseMetaExecutionResult{
				AccumulatedFeesInEpoch: big.NewInt(0),
				DevFeesInEpoch:         big.NewInt(0),
			},
		},
	}

	t.Run("should return ErrNilBlockHeader because of nil metaHeader argument", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessEpochStartProposeBlock(nil, nil)
		require.Equal(t, process.ErrNilBlockHeader, err)
	})

	t.Run("should return ErrNilBody because of nil body argument", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessEpochStartProposeBlock(&defaultMetaBlockV3, nil)
		require.Equal(t, process.ErrNilBlockBody, err)
	})

	t.Run("should return ErrEpochStartProposeBlockHasMiniBlocks because the body has mini blocks", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessEpochStartProposeBlock(&defaultMetaBlockV3, &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{},
			},
		})
		require.Equal(t, process.ErrEpochStartProposeBlockHasMiniBlocks, err)
	})

	t.Run("if processEconomicsDataForEpochStartProposeBlock fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"blockChain": &testscommon.ChainHandlerStub{
				GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
					return &block.MetaExecutionResult{}
				},
			},
			"epochStartDataCreator": &mock.EpochStartDataCreatorStub{
				CreateEpochStartShardDataMetablockV3Called: func(metaBlock data.MetaHeaderHandler) ([]block.EpochStartShardData, error) {
					return nil, expectedErr
				},
			},
		})
		require.Nil(t, err)

		_, err = mp.ProcessEpochStartProposeBlock(&defaultMetaBlockV3, &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("if processing epoch start mini blocks fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		blockchainMock := &testscommon.ChainHandlerMock{}
		err := blockchainMock.SetGenesisHeader(&block.Header{})
		require.Nil(t, err)
		blockchainMock.SetLastExecutionInfo(&block.MetaBlockV3{}, &block.MetaExecutionResult{
			ExecutionResult: &block.BaseMetaExecutionResult{},
		})
		dataComponents.BlockChain = blockchainMock

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.ValidatorStatisticsProcessor = &testscommon.ValidatorStatisticsProcessorStub{
			RootHashCalled: func() ([]byte, error) {
				return nil, expectedErr
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{})
		_, err = mp.ProcessEpochStartProposeBlock(&defaultMetaBlockV3, &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("if updating validator statistics fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		blockchainMock := &testscommon.ChainHandlerMock{}
		err := blockchainMock.SetGenesisHeader(&block.Header{})
		require.Nil(t, err)
		blockchainMock.SetLastExecutionInfo(&block.MetaBlockV3{}, &block.MetaExecutionResult{
			ExecutionResult: &block.BaseMetaExecutionResult{},
		})
		dataComponents.BlockChain = blockchainMock

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.EpochEconomics = &mock.EpochEconomicsStub{
			ComputeEndOfEpochEconomicsV3Called: func(metaBlock data.MetaHeaderHandler, prevBlockExecutionResults data.BaseMetaExecutionResultHandler, epochStartHandler data.EpochStartHandler) (*block.Economics, error) {
				return &block.Economics{
					RewardsForProtocolSustainability: big.NewInt(0),
					PrevEpochStartRound:              1,
				}, nil
			},
		}

		arguments.ValidatorStatisticsProcessor = &testscommon.ValidatorStatisticsProcessorStub{
			UpdatePeerStateV3Called: func(header data.MetaHeaderHandler, metaExecutionResult data.MetaExecutionResultHandler) ([]byte, error) {
				return nil, expectedErr
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{})

		_, err = mp.ProcessEpochStartProposeBlock(&defaultMetaBlockV3, &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("commit state is not called by processEpochStartProposeBlock", func(t *testing.T) {
		t.Parallel()

		commitCalled := false

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		blockchainMock := &testscommon.ChainHandlerMock{}
		err := blockchainMock.SetGenesisHeader(&block.Header{})
		require.Nil(t, err)
		blockchainMock.SetLastExecutionInfo(&block.MetaBlockV3{}, &block.MetaExecutionResult{
			ExecutionResult: &block.BaseMetaExecutionResult{
				AccumulatedFeesInEpoch: big.NewInt(10),
				DevFeesInEpoch:         big.NewInt(10),
			},
		})
		dataComponents.BlockChain = blockchainMock

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.EpochEconomics = &mock.EpochEconomicsStub{
			ComputeEndOfEpochEconomicsV3Called: func(metaBlock data.MetaHeaderHandler, prevBlockExecutionResults data.BaseMetaExecutionResultHandler, epochStartHandler data.EpochStartHandler) (*block.Economics, error) {
				return &block.Economics{
					RewardsForProtocolSustainability: big.NewInt(0),
					PrevEpochStartRound:              1,
				}, nil
			},
		}

		accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
		accounts := &testscommonState.AccountsStub{
			CommitCalled: func() ([]byte, error) {
				commitCalled = true
				return []byte("stateRoot"), nil
			},
		}
		accountsDb[state.UserAccountsState] = accounts
		accountsDb[state.PeerAccountsState] = accounts
		arguments.AccountsDB = accountsDb

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{})

		_, err = mp.ProcessEpochStartProposeBlock(&defaultMetaBlockV3, &block.Body{})
		require.Nil(t, err)
		require.False(t, commitCalled)
	})

	t.Run("if HandleProcessErrorCutoff fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		blockchainMock := &testscommon.ChainHandlerMock{}
		err := blockchainMock.SetGenesisHeader(&block.Header{})
		require.Nil(t, err)
		blockchainMock.SetLastExecutionInfo(&block.MetaBlockV3{}, &block.MetaExecutionResult{
			ExecutionResult: &block.BaseMetaExecutionResult{
				AccumulatedFeesInEpoch: big.NewInt(10),
				DevFeesInEpoch:         big.NewInt(10),
			},
		})
		dataComponents.BlockChain = blockchainMock

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.EpochEconomics = &mock.EpochEconomicsStub{
			ComputeEndOfEpochEconomicsV3Called: func(metaBlock data.MetaHeaderHandler, prevBlockExecutionResults data.BaseMetaExecutionResultHandler, epochStartHandler data.EpochStartHandler) (*block.Economics, error) {
				return &block.Economics{
					RewardsForProtocolSustainability: big.NewInt(0),
					PrevEpochStartRound:              1,
				}, nil
			},
		}

		arguments.BlockProcessingCutoffHandler = &testscommon.BlockProcessingCutoffStub{
			HandleProcessErrorCutoffCalled: func(header data.HeaderHandler) error {
				return expectedErr
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{})

		_, err = mp.ProcessEpochStartProposeBlock(&defaultMetaBlockV3, &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("if calculating the hash fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		err := coreComponents.SetInternalMarshalizer(&marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		})
		require.Nil(t, err)

		blockchainMock := &testscommon.ChainHandlerMock{}
		err = blockchainMock.SetGenesisHeader(&block.Header{})
		require.Nil(t, err)
		blockchainMock.SetLastExecutionInfo(&block.MetaBlockV3{}, &block.MetaExecutionResult{
			ExecutionResult: &block.BaseMetaExecutionResult{},
		})
		dataComponents.BlockChain = blockchainMock

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.EpochEconomics = &mock.EpochEconomicsStub{
			ComputeEndOfEpochEconomicsV3Called: func(metaBlock data.MetaHeaderHandler, prevBlockExecutionResults data.BaseMetaExecutionResultHandler, epochStartHandler data.EpochStartHandler) (*block.Economics, error) {
				return &block.Economics{
					RewardsForProtocolSustainability: big.NewInt(0),
					PrevEpochStartRound:              1,
				}, nil
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{})

		_, err = mp.ProcessEpochStartProposeBlock(&defaultMetaBlockV3, &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		blockchainMock := &testscommon.ChainHandlerMock{}
		err := blockchainMock.SetGenesisHeader(&block.Header{})
		require.Nil(t, err)
		blockchainMock.SetLastExecutionInfo(&block.MetaBlockV3{}, &block.MetaExecutionResult{
			ExecutionResult: &block.BaseMetaExecutionResult{
				AccumulatedFeesInEpoch: big.NewInt(10),
				DevFeesInEpoch:         big.NewInt(10),
			},
		})
		dataComponents.BlockChain = blockchainMock

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.EpochEconomics = &mock.EpochEconomicsStub{
			ComputeEndOfEpochEconomicsV3Called: func(metaBlock data.MetaHeaderHandler, prevBlockExecutionResults data.BaseMetaExecutionResultHandler, epochStartHandler data.EpochStartHandler) (*block.Economics, error) {
				return &block.Economics{
					RewardsForProtocolSustainability: big.NewInt(0),
					PrevEpochStartRound:              1,
				}, nil
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{
			Epoch: 1,
			EpochStartData: &block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{},
				Economics:            block.Economics{},
			},
		})

		_, err = mp.ProcessEpochStartProposeBlock(&defaultMetaBlockV3, &block.Body{})
		require.Nil(t, err)
	})
}

func TestMetaProcessor_processEconomicsDataForEpochStartProposeBlock(t *testing.T) {
	t.Parallel()

	defaultMetaBlockV3 := block.MetaBlockV3{
		LastExecutionResult: &block.MetaExecutionResultInfo{
			ExecutionResult: &block.BaseMetaExecutionResult{
				AccumulatedFeesInEpoch: big.NewInt(0),
				DevFeesInEpoch:         big.NewInt(0),
			},
		},
	}

	t.Run("should return ErrNilBaseExecutionResult error on nil last execution result", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"blockChain": &testscommon.ChainHandlerStub{
				GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
					return nil
				},
			},
		})
		require.Nil(t, err)

		err = mp.ProcessEconomicsDataForEpochStartProposeBlock(&defaultMetaBlockV3)
		require.ErrorContains(t, err, process.ErrNilBaseExecutionResult.Error())
	})

	t.Run("should return ErrWrongTypeAssertion error on wrong type of last execution result", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"blockChain": &testscommon.ChainHandlerStub{
				GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
					return &block.ExecutionResult{}
				},
			},
		})
		require.Nil(t, err)

		err = mp.ProcessEconomicsDataForEpochStartProposeBlock(&defaultMetaBlockV3)
		require.Equal(t, common.ErrWrongTypeAssertion, err)
	})

	t.Run("if CreateEpochStartShardDataMetablockV3 fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"blockChain": &testscommon.ChainHandlerStub{
				GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
					return &block.MetaExecutionResult{}
				},
			},
			"epochStartDataCreator": &mock.EpochStartDataCreatorStub{
				CreateEpochStartShardDataMetablockV3Called: func(metaBlock data.MetaHeaderHandler) ([]block.EpochStartShardData, error) {
					return nil, expectedErr
				},
			},
		})
		require.Nil(t, err)

		err = mp.ProcessEconomicsDataForEpochStartProposeBlock(&defaultMetaBlockV3)
		require.Equal(t, expectedErr, err)
	})

	t.Run("if ComputeEndOfEpochEconomicsV3 fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"blockChain": &testscommon.ChainHandlerStub{
				GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
					return &block.MetaExecutionResult{}
				},
			},
			"epochStartDataCreator": &mock.EpochStartDataCreatorStub{
				CreateEpochStartShardDataMetablockV3Called: func(metaBlock data.MetaHeaderHandler) ([]block.EpochStartShardData, error) {
					return []block.EpochStartShardData{
						{},
					}, nil
				},
			},
			"epochEconomics": &mock.EpochEconomicsStub{
				ComputeEndOfEpochEconomicsV3Called: func(metaBlock data.MetaHeaderHandler, prevBlockExecutionResults data.BaseMetaExecutionResultHandler, epochStartHandler data.EpochStartHandler) (*block.Economics, error) {
					return nil, expectedErr
				},
			},
		})
		require.Nil(t, err)

		err = mp.ProcessEconomicsDataForEpochStartProposeBlock(&defaultMetaBlockV3)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"blockChain": &testscommon.ChainHandlerStub{
				GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
					return &block.MetaExecutionResult{}
				},
			},
			"epochStartDataCreator": &mock.EpochStartDataCreatorStub{
				CreateEpochStartShardDataMetablockV3Called: func(metaBlock data.MetaHeaderHandler) ([]block.EpochStartShardData, error) {
					return []block.EpochStartShardData{
						{},
					}, nil
				},
			},
			"epochEconomics": &mock.EpochEconomicsStub{
				ComputeEndOfEpochEconomicsV3Called: func(metaBlock data.MetaHeaderHandler, prevBlockExecutionResults data.BaseMetaExecutionResultHandler, epochStartHandler data.EpochStartHandler) (*block.Economics, error) {
					return &block.Economics{}, nil
				},
			},
		})
		require.Nil(t, err)

		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{})
		err = mp.ProcessEconomicsDataForEpochStartProposeBlock(&defaultMetaBlockV3)
		require.Nil(t, err)
	})
}

func TestMetaProcessor_createExecutionResult(t *testing.T) {
	t.Parallel()

	t.Run("if computing the gas used fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"feeHandler": &mock.FeeAccumulatorStub{
				GetAccumulatedFeesCalled: func() *big.Int {
					return big.NewInt(5)
				},
				GetDeveloperFeesCalled: func() *big.Int {
					return big.NewInt(5)
				},
			},
			"gasConsumedProvider": &testscommon.GasHandlerStub{
				TotalGasPenalizedCalled: func() uint64 {
					return 10
				},
				TotalGasRefundedCalled: func() uint64 {
					return 10
				},
			},
		})
		require.Nil(t, err)

		mbh := []data.MiniBlockHeaderHandler{
			&block.MiniBlockHeader{},
		}
		_, err = mp.CreateExecutionResult(mbh, &block.MetaBlockV3{
			EpochChangeProposed: true,
		}, []byte("headerHash"), []byte("receiptHash"), []byte("valStatRootHash"), 5)
		require.Equal(t, process.ErrGasUsedExceedsGasProvided, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		arguments.FeeHandler = &mock.FeeAccumulatorStub{
			GetAccumulatedFeesCalled: func() *big.Int {
				return big.NewInt(5)
			},
			GetDeveloperFeesCalled: func() *big.Int {
				return big.NewInt(5)
			},
		}
		arguments.GasHandler = &testscommon.GasHandlerStub{
			TotalGasProvidedCalled: func() uint64 {
				return 10
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		mbh := []data.MiniBlockHeaderHandler{
			&block.MiniBlockHeader{},
		}
		execResult, err := mp.CreateExecutionResult(mbh, &block.MetaBlockV3{
			EpochChangeProposed: true,
		}, []byte("headerHash"), []byte("receiptHash"), []byte("valStatRootHash"), 5)
		require.Nil(t, err)

		metaExecResult, ok := execResult.(*block.MetaExecutionResult)
		require.True(t, ok)
		require.Equal(t, metaExecResult.ExecutedTxCount, uint64(5))
		require.Equal(t, metaExecResult.ReceiptsHash, []byte("receiptHash"))
		require.Equal(t, metaExecResult.GetValidatorStatsRootHash(), []byte("valStatRootHash"))
	})
}

func TestMetaProcessor_collectExecutionResults(t *testing.T) {
	t.Parallel()

	t.Run("if CreateReceiptsHash fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		txCoordinatorMock := createTxCoordinatorMock()
		txCoordinatorMock.CreateReceiptsHashCalled = func() ([]byte, error) {
			return nil, expectedErr
		}
		arguments.TxCoordinator = &txCoordinatorMock

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.CollectExecutionResults([]byte("headerHash"), &block.MetaBlockV3{
			LastExecutionResult: &block.MetaExecutionResultInfo{
				ExecutionResult: &block.BaseMetaExecutionResult{
					AccumulatedFeesInEpoch: big.NewInt(0),
					DevFeesInEpoch:         big.NewInt(0),
				},
			},
		}, &block.Body{}, []byte("valStatRootHash"))
		require.Equal(t, expectedErr, err)
	})

	t.Run("if marshal fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		err := coreComponents.SetInternalMarshalizer(&marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		})
		require.Nil(t, err)

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		txCoordinatorMock := createTxCoordinatorMock()
		arguments.TxCoordinator = &txCoordinatorMock

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.CollectExecutionResults([]byte("headerHash"), &block.MetaBlockV3{
			LastExecutionResult: &block.MetaExecutionResultInfo{
				ExecutionResult: &block.BaseMetaExecutionResult{
					AccumulatedFeesInEpoch: big.NewInt(0),
					DevFeesInEpoch:         big.NewInt(0),
				},
			},
		}, &block.Body{}, []byte("valStatRootHash"))
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.BaseMetaExecutionResult{
					AccumulatedFeesInEpoch: big.NewInt(10),
					DevFeesInEpoch:         big.NewInt(10),
				}
			},
		}

		txCoordinatorMock := createTxCoordinatorMock()
		arguments.TxCoordinator = &txCoordinatorMock

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		execResult, err := mp.CollectExecutionResults([]byte("headerHash"), &block.MetaBlockV3{
			LastExecutionResult: &block.MetaExecutionResultInfo{
				ExecutionResult: &block.BaseMetaExecutionResult{
					AccumulatedFeesInEpoch: big.NewInt(0),
					DevFeesInEpoch:         big.NewInt(0),
				},
			},
		}, &block.Body{}, []byte("valStatRootHash"))
		require.Nil(t, err)

		metaExecResult, ok := execResult.(*block.MetaExecutionResult)
		require.True(t, ok)
		require.Equal(t, metaExecResult.ExecutedTxCount, uint64(4))
		require.Equal(t, metaExecResult.ReceiptsHash, []byte("receiptHash"))
		require.Equal(t, metaExecResult.GetValidatorStatsRootHash(), []byte("valStatRootHash"))
	})
}

func TestMetaProcessor_collectExecutionResultsEpochStartProposal(t *testing.T) {
	t.Parallel()

	t.Run("should fail because of error on CreateReceiptsHash", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			CreateReceiptsHashCalled: func() ([]byte, error) {
				return nil, expectedErr
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.CollectExecutionResultsEpochStartProposal([]byte("headerHash"), &block.MetaBlockV3{
			LastExecutionResult: &block.MetaExecutionResultInfo{
				ExecutionResult: &block.BaseMetaExecutionResult{
					AccumulatedFeesInEpoch: big.NewInt(0),
					DevFeesInEpoch:         big.NewInt(0),
				},
			},
		}, &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("hash1"),
						[]byte("hash2"),
					},
				},
			},
		}, []byte("valStatRootHash"))
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.BaseMetaExecutionResult{
					AccumulatedFeesInEpoch: big.NewInt(10),
					DevFeesInEpoch:         big.NewInt(10),
				}
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		execResult, err := mp.CollectExecutionResultsEpochStartProposal([]byte("headerHash"), &block.MetaBlockV3{
			LastExecutionResult: &block.MetaExecutionResultInfo{
				ExecutionResult: &block.BaseMetaExecutionResult{
					AccumulatedFeesInEpoch: big.NewInt(0),
					DevFeesInEpoch:         big.NewInt(0),
				},
			},
		}, &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("hash1"),
						[]byte("hash2"),
					},
				},
			},
		}, []byte("valStatRootHash"))
		require.Nil(t, err)

		metaExecResult, ok := execResult.(*block.MetaExecutionResult)
		require.True(t, ok)
		require.Equal(t, metaExecResult.ExecutedTxCount, uint64(2))
		require.Equal(t, metaExecResult.ReceiptsHash, []byte("receiptHash"))
		require.Equal(t, metaExecResult.GetValidatorStatsRootHash(), []byte("valStatRootHash"))
	})
}

func TestMetaProcessor_ProcessBlockProposal(t *testing.T) {
	t.Parallel()

	defaultMetaBlockV3 := block.MetaBlockV3{
		LastExecutionResult: &block.MetaExecutionResultInfo{
			ExecutionResult: &block.BaseMetaExecutionResult{
				AccumulatedFeesInEpoch: big.NewInt(0),
				DevFeesInEpoch:         big.NewInt(0),
			},
		},
	}
	t.Run("should return ErrNilBlockHeader because of nil argument", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(nil, []byte("headerHash"), &block.Body{})
		require.Equal(t, process.ErrNilBlockHeader, err)
	})

	t.Run("should return ErrNilBlockBody because of nil argument", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.MetaBlockV3{}, []byte("headerHash"), nil)
		require.Equal(t, process.ErrNilBlockBody, err)
	})

	t.Run("should return ErrInvalidHeader because of nil argument", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.MetaBlock{}, []byte("headerHash"), &block.Body{})
		require.Equal(t, process.ErrInvalidHeader, err)
	})

	t.Run("should return ErrWrongTypeAssertion in case of wrong header", func(t *testing.T) {
		t.Parallel()

		checkEpochCounter := 0
		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		coreComponents.EpochNotifierField = &epochNotifier.EpochNotifierStub{
			CheckEpochCalled: func(header data.HeaderHandler) {
				checkEpochCounter += 1
			},
		}

		checkRoundCounter := 0
		coreComponents.RoundNotifierField = &epochNotifier.RoundNotifierStub{
			CheckRoundCalled: func(header data.HeaderHandler) {
				checkRoundCounter += 1
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.HeaderV3{
			Round: 2,
			Epoch: 2,
		}, []byte("headerHash"), &block.Body{})
		require.Equal(t, process.ErrWrongTypeAssertion, err)
		require.Equal(t, 1, checkEpochCounter)
	})

	t.Run("should return ErrAccountStateDirty in case of dirty state", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
		accounts := &testscommonState.AccountsStub{
			CommitCalled: func() ([]byte, error) {
				return nil, nil
			},
			RootHashCalled: func() ([]byte, error) {
				return nil, nil
			},
			RecreateTrieIfNeededCalled: func(options common.RootHashHolder) error {
				return nil
			},
			JournalLenCalled: func() int {
				return 1
			},
		}

		accountsDb[state.UserAccountsState] = accounts
		accountsDb[state.PeerAccountsState] = accounts

		arguments.AccountsDB = accountsDb
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.MetaBlockV3{}, []byte("headerHash"), &block.Body{})
		require.True(t, errors.Is(err, process.ErrAccountStateDirty))
	})

	t.Run("if checking context fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		previousHash := []byte("hash")
		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutedBlockInfoCalled: func() (uint64, []byte, []byte) {
				return 0, previousHash, nil
			},
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlockV3{}
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.MetaBlockV3{
			PrevHash: []byte("wrongHash"),
		}, []byte("headerHash"), &block.Body{})
		require.Equal(t, process.ErrBlockHashDoesNotMatch, err)
	})

	t.Run("if creating block fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()

		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlockV3{
					EpochStart: block.EpochStart{
						LastFinalizedHeaders: []block.EpochStartShardData{
							{},
						},
					},
				}
			},
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.MetaExecutionResult{
					ExecutionResult: &block.BaseMetaExecutionResult{},
				}
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			AddIntermediateTransactionsCalled: func(mapSCRs map[block.Type][]data.TransactionHandler, key []byte) error {
				return expectedErr
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.MetaBlockV3{
			Nonce: 1,
		}, []byte("headerHash"), &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("if setting the current header fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		arguments.BlockChainHook = &testscommon.BlockChainHookStub{
			SetCurrentHeaderCalled: func(hdr data.HeaderHandler) error {
				return expectedErr
			},
		}

		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlockV3{}
			},
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.MetaExecutionResult{
					ExecutionResult: &block.BaseMetaExecutionResult{},
				}
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.MetaBlockV3{
			Nonce: 1,
		}, []byte("headerHash"), &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("if processing first block after epoch start fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlockV3{
					EpochStart: block.EpochStart{
						LastFinalizedHeaders: []block.EpochStartShardData{
							{},
						},
					},
				}
			},
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.MetaExecutionResult{
					ExecutionResult: &block.BaseMetaExecutionResult{},
				}
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.EpochSystemSCProcessor = &testscommon.EpochStartSystemSCStub{
			ToggleUnStakeUnBondCalled: func(value bool) error {
				return expectedErr
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.MetaBlockV3{
			Nonce: 1,
		}, []byte("headerHash"), &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("if processing epoch start proposal block fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlockV3{}
			},
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.MetaExecutionResult{
					ExecutionResult: &block.BaseMetaExecutionResult{},
				}
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.MetaBlockV3{
			Nonce:               1,
			EpochChangeProposed: true,
		}, []byte("headerHash"), &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{}, {}, {},
			},
		})
		require.Equal(t, process.ErrEpochStartProposeBlockHasMiniBlocks, err)
	})

	t.Run("if checking the data prepared for processing fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlockV3{}
			},
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.MetaExecutionResult{
					ExecutionResult: &block.BaseMetaExecutionResult{},
				}
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			IsDataPreparedForProcessingCalled: func(haveTime func() time.Duration) error {
				return expectedErr
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.MetaBlockV3{
			Nonce: 1,
		}, []byte("headerHash"), &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("if waiting for failing headers fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlockV3{}
			},
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.MetaExecutionResult{
					ExecutionResult: &block.BaseMetaExecutionResult{},
				}
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.HeadersForBlock = &testscommon.HeadersForBlockMock{
			WaitForHeadersIfNeededCalled: func(haveTime func() time.Duration) error {
				return expectedErr
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.MetaBlockV3{
			Nonce: 1,
		}, []byte("headerHash"), &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("if processing the transaction block fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlockV3{}
			},
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.MetaExecutionResult{
					ExecutionResult: &block.BaseMetaExecutionResult{},
				}
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			ProcessBlockTransactionCalled: func(header data.HeaderHandler, body *block.Body, haveTime func() time.Duration) error {
				return expectedErr
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.MetaBlockV3{
			Nonce: 1,
		}, []byte("headerHash"), &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("if verifying created block fails, the error should be returned", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlockV3{}
			},
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.MetaExecutionResult{
					ExecutionResult: &block.BaseMetaExecutionResult{},
				}
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			VerifyCreatedBlockTransactionsCalled: func(hdr data.HeaderHandler, body *block.Body) error {
				return expectedErr
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.MetaBlockV3{
			Nonce: 1,
		}, []byte("headerHash"), &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("if updating protocol fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlockV3{}
			},
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.MetaExecutionResult{
					ExecutionResult: &block.BaseMetaExecutionResult{},
				}
			},
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.SCToProtocol = &mock.SCToProtocolStub{
			UpdateProtocolCalled: func(body *block.Body, nonce uint64) error {
				return expectedErr
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.MetaBlockV3{
			Nonce: 1,
		}, []byte("headerHash"), &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("if updating validator statistics fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.MetaExecutionResult{
					ExecutionResult: &block.BaseMetaExecutionResult{},
				}
			},
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlockV3{}
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.ValidatorStatisticsProcessor = &testscommon.ValidatorStatisticsProcessorStub{
			UpdatePeerStateV3Called: func(header data.MetaHeaderHandler, metaExecutionResult data.MetaExecutionResultHandler) ([]byte, error) {
				return nil, expectedErr
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.MetaBlockV3{
			Nonce: 1,
		}, []byte("headerHash"), &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("commit state is not called by ProcessBlockProposal", func(t *testing.T) {
		t.Parallel()

		commitCalled := false

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.MetaExecutionResult{
					ExecutionResult: &block.BaseMetaExecutionResult{
						AccumulatedFeesInEpoch: big.NewInt(10),
						DevFeesInEpoch:         big.NewInt(10),
					},
				}
			},
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &defaultMetaBlockV3
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
		accounts := &testscommonState.AccountsStub{
			CommitCalled: func() ([]byte, error) {
				commitCalled = true
				return []byte("stateRoot"), nil
			},
			RootHashCalled: func() ([]byte, error) {
				return nil, nil
			},
			RecreateTrieIfNeededCalled: func(options common.RootHashHolder) error {
				return nil
			},
		}

		accountsDb[state.UserAccountsState] = accounts
		accountsDb[state.PeerAccountsState] = accounts

		arguments.AccountsDB = accountsDb
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		newBlock := defaultMetaBlockV3
		newBlock.Nonce = 1
		_, err = mp.ProcessBlockProposal(&newBlock, []byte("headerHash"), &block.Body{})
		require.Nil(t, err)
		require.False(t, commitCalled)

		err = mp.CommitBlockProposalState(&newBlock)
		require.Nil(t, err)
		require.True(t, commitCalled)
	})

	t.Run("if HandleProcessErrorCutoff fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.MetaExecutionResult{
					ExecutionResult: &block.BaseMetaExecutionResult{
						AccumulatedFeesInEpoch: big.NewInt(10),
						DevFeesInEpoch:         big.NewInt(10),
					},
				}
			},
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &defaultMetaBlockV3
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.BlockProcessingCutoffHandler = &testscommon.BlockProcessingCutoffStub{
			HandleProcessErrorCutoffCalled: func(header data.HeaderHandler) error {
				return expectedErr
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		newBlock := defaultMetaBlockV3
		newBlock.Nonce = 1
		_, err = mp.ProcessBlockProposal(&newBlock, []byte("headerHash"), &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("if creating the execution result fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()

		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.MetaExecutionResult{
					ExecutionResult: &block.BaseMetaExecutionResult{
						AccumulatedFeesInEpoch: big.NewInt(10),
						DevFeesInEpoch:         big.NewInt(10),
					},
				}
			},
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlockV3{}
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			CreateReceiptsHashCalled: func() ([]byte, error) {
				return nil, expectedErr
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessBlockProposal(&block.MetaBlockV3{
			Nonce: 1,
		}, []byte("headerHash"), &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
				return &block.MetaExecutionResult{
					ExecutionResult: &block.BaseMetaExecutionResult{
						AccumulatedFeesInEpoch: big.NewInt(10),
						DevFeesInEpoch:         big.NewInt(10),
					},
				}
			},
			GetLastExecutedBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlockV3{}
			},
		}

		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)
		arguments.ValidatorStatisticsProcessor = &testscommon.ValidatorStatisticsProcessorStub{
			RootHashCalled: func() ([]byte, error) {
				return nil, expectedErr
			},
		}

		receiptHash := []byte("receiptHash")
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			CreateReceiptsHashCalled: func() ([]byte, error) {
				return receiptHash, nil
			},
		}

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		executionResult, err := mp.ProcessBlockProposal(&block.MetaBlockV3{
			Nonce: 1,
			LastExecutionResult: &block.MetaExecutionResultInfo{
				ExecutionResult: &block.BaseMetaExecutionResult{
					DevFeesInEpoch:         big.NewInt(1), // fees not taken from header
					AccumulatedFeesInEpoch: big.NewInt(1), // fees not taken from header
				},
			},
		}, []byte("headerHash"), &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					Type: block.ReceiptBlock,
					TxHashes: [][]byte{
						[]byte("txHash1"),
					},
				},
			},
		})
		require.Nil(t, err)

		metaExecutionResult, ok := executionResult.(*block.MetaExecutionResult)
		require.True(t, ok)

		require.Equal(t, receiptHash, metaExecutionResult.ReceiptsHash)
		require.Equal(t, big.NewInt(10), metaExecutionResult.ExecutionResult.DevFeesInEpoch)
		require.Equal(t, big.NewInt(10), metaExecutionResult.ExecutionResult.AccumulatedFeesInEpoch)
		require.Equal(t, 0, len(metaExecutionResult.MiniBlockHeaders))
		require.Equal(t, uint64(0), metaExecutionResult.GetExecutedTxCount())
	})
}

func createTxCoordinatorMock() testscommon.TransactionCoordinatorMock {
	return testscommon.TransactionCoordinatorMock{
		GetCreatedMiniBlocksFromMeCalled: func() block.MiniBlockSlice {
			return []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("hash1"),
						[]byte("hash2"),
					},
				},
			}
		},
		CreatePostProcessMiniBlocksCalled: func() block.MiniBlockSlice {
			return []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("hash3"),
						[]byte("hash4"),
					},
				},
			}
		},
		GetCreatedInShardMiniBlocksCalled: func() []*block.MiniBlock {
			return []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("hash5"),
						[]byte("hash6"),
					},
				},
			}
		},
	}
}

func createLastShardHeadersNotGenesis() map[uint32]blproc.ShardHeaderInfo {
	shard0 := uint32(0)
	shard1 := uint32(1)
	shard2 := uint32(2)

	return map[uint32]blproc.ShardHeaderInfo{
		shard0: {
			Header: &block.Header{
				ShardID: shard0,
				Nonce:   10,
				Round:   10,
			},
			Hash: []byte("hash1"),
		},
		shard1: {
			Header: &block.Header{
				ShardID: shard1,
				Nonce:   10,
				Round:   10,
			},
			Hash: []byte("hash2"),
		},
		shard2: {
			Header: &block.Header{
				ShardID: shard2,
				Nonce:   10,
				Round:   10,
			},
			Hash: []byte("hash3"),
		},
	}
}

func createMetaProcessorMapForCreatingEpochStart() map[string]interface{} {
	executionResultHeaderHash := []byte("exec result header hash")
	executionResultsForEpochStart := block.MetaExecutionResult{
		ExecutionResult: &block.BaseMetaExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash: executionResultHeaderHash,
			},
		},
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				Hash:          []byte("mb hash"),
				SenderShardID: core.MetachainShardId,
				Type:          block.RewardsBlock, // this miniBlock marks the epoch start
			},
		},
	}
	prevValidMetaBlockV3 := testscommon.HeaderHandlerStub{
		IsHeaderV3Called: func() bool {
			return true
		},
		GetLastExecutionResultHandlerCalled: func() data.LastExecutionResultHandler {
			return &block.MetaExecutionResultInfo{
				ExecutionResult: &block.BaseMetaExecutionResult{},
			}
		},
	}
	blockTracker := integrationTestsMock.BlockTrackerStub{
		GetLastCrossNotarizedHeadersForAllShardsCalled: func() (map[uint32]data.HeaderHandler, error) {
			return map[uint32]data.HeaderHandler{
				0:                       &block.HeaderV3{},
				1:                       &block.HeaderV3{},
				common.MetachainShardId: &block.MetaBlockV3{},
			}, nil
		},
	}

	return map[string]interface{}{
		"shardCoordinator": &mock.ShardCoordinatorStub{
			SelfIdCalled: func() uint32 {
				return common.MetachainShardId
			},
		},
		"epochStartTrigger": &testscommon.EpochStartTriggerStub{
			EpochCalled: func() uint32 {
				return 0
			},
			ShouldProposeEpochChangeCalled: func(round uint64, nonce uint64) bool {
				return false
			},
		},
		"versionedHeaderFactory": &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return &block.MetaBlockV3{}
			},
		},
		"executionManager": &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return []data.BaseExecutionResultHandler{&executionResultsForEpochStart}, nil
			},
		},
		"blockChain": &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &prevValidMetaBlockV3
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("prev header hash")
			},
		},
		"executionResultsInclusionEstimator": &processMocks.InclusionEstimatorMock{
			DecideCalled: func(lastNotarised *common.LastExecutionResultForInclusion, pending []data.BaseExecutionResultHandler, currentHdrTsMs uint64) (allowed int) {
				return 1 // allow the inclusion of the first execution result
			},
		},
		"appStatusHandler":    &statusHandlerMock.AppStatusHandlerStub{},
		"maxProposalNonceGap": uint64(10),
		"blockTracker":        &blockTracker,
	}
}

type ancestryGateProcessor interface {
	SelectIncomingMiniBlocks(lastShardHdrs map[uint32]blproc.ShardHeaderInfo, orderedHdrs []data.HeaderHandler, orderedHdrsHashes [][]byte, maxNumHeadersFromSameShard uint32, haveTime func() bool) (map[uint32]uint32, error)
	SelectContendedShardHeaders(round uint64, lastShardHdrs map[uint32]blproc.ShardHeaderInfo, hdrsAddedForShard map[uint32]uint32, haveTime func() bool) error
	CheckHeadersSequenceCorrectness(hdrsForShard []blproc.ShardHeaderInfo, lastNotarizedHeaderInfoForShard blproc.ShardHeaderInfo) error
}

func TestMetaProcessor_ReferencedMetaAncestryGate(t *testing.T) {
	t.Parallel()

	haveTime := func() bool { return true }

	metaHash97 := []byte("metaHash97")
	metaHash98 := []byte("metaHash98")
	metaHash99 := []byte("metaHash99")
	headHash := []byte("metaHash100")
	deadMetaHash := []byte("deadMetaHash99")

	meta97 := &block.MetaBlock{Nonce: 97, PrevHash: []byte("metaHash96")}
	meta98 := &block.MetaBlock{Nonce: 98, PrevHash: metaHash97}
	meta99 := &block.MetaBlock{Nonce: 99, PrevHash: metaHash98}
	headMeta := &block.MetaBlock{Nonce: 100, PrevHash: metaHash99}
	deadMeta := &block.MetaBlock{Nonce: 99, PrevHash: metaHash98, Round: 1000}

	parentShardHash := []byte("parentShardHash")
	parentShard := &block.Header{ShardID: 0, Nonce: 10, Round: 10}

	newLastShardHdrs := func() map[uint32]blproc.ShardHeaderInfo {
		return map[uint32]blproc.ShardHeaderInfo{0: {Header: parentShard, Hash: parentShardHash}}
	}
	newShardCandidate := func(round uint64, metaHashes [][]byte) *block.Header {
		return &block.Header{ShardID: 0, Nonce: 11, Round: round, PrevHash: parentShardHash, MetaBlockHashes: metaHashes}
	}

	buildProcessor := func(
		supernovaEnabled bool,
		poolHeaders map[string]data.HeaderHandler,
		nonceCandidates []data.HeaderHandler,
		nonceCandidateHashes [][]byte,
		nonceHashStorer *storageStubs.StorerStub,
		referenced *[][]byte,
	) ancestryGateProcessor {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		coreComponents.EnableEpochsHandlerField = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return supernovaEnabled && flag == common.SupernovaFlag
			},
		}
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled:     func() data.HeaderHandler { return headMeta },
			GetCurrentBlockHeaderHashCalled: func() []byte { return headHash },
		}
		dataComponents.Storage = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				if unitType == dataRetriever.MetaHdrNonceHashDataUnit && nonceHashStorer != nil {
					return nonceHashStorer, nil
				}
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						return nil, errors.New("not found")
					},
					SearchFirstCalled: func(key []byte) ([]byte, error) {
						return nil, errors.New("not found")
					},
				}, nil
			},
		}
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.HeadersCalled = func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
						header, found := poolHeaders[string(hash)]
						if !found {
							return nil, errors.New("header not found")
						}
						return header, nil
					},
					GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardID uint32) ([]data.HeaderHandler, [][]byte, error) {
						if len(nonceCandidates) == 0 {
							return nil, nil, errors.New("no headers")
						}
						return nonceCandidates, nonceCandidateHashes, nil
					},
				}
			}
			ph.ProofsCalled = func() dataRetriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{
					HasProofCalled: func(shardID uint32, headerHash []byte) bool { return true },
				}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.HeaderValidator = &processMocks.HeaderValidatorMock{
			IsHeaderConstructionValidCalled: func(currHdr, prevHdr data.HeaderHandler) error { return nil },
		}
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{
			AddReferencedHeaderCalled: func(header data.HeaderHandler, headerHash []byte) {
				*referenced = append(*referenced, headerHash)
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		return mp
	}

	canonicalPool := func() map[string]data.HeaderHandler {
		return map[string]data.HeaderHandler{
			string(metaHash98):   meta98,
			string(metaHash99):   meta99,
			string(headHash):     headMeta,
			string(deadMetaHash): deadMeta,
		}
	}

	t.Run("ordinary selection keeps ancestor references and pays no storage reads", func(t *testing.T) {
		t.Parallel()

		iopsGuard := &storageStubs.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				require.Fail(t, "nonce hash storer must not be read for pooled ancestors")
				return nil, nil
			},
		}
		referenced := make([][]byte, 0)
		candidate := newShardCandidate(11, [][]byte{metaHash99, headHash})
		candidateHash := []byte("candidateHash")
		mp := buildProcessor(true, canonicalPool(), nil, nil, iopsGuard, &referenced)

		_, err := mp.SelectIncomingMiniBlocks(newLastShardHdrs(), []data.HeaderHandler{candidate}, [][]byte{candidateHash}, 10, haveTime)
		require.Nil(t, err)
		require.Equal(t, [][]byte{candidateHash}, referenced)
	})

	t.Run("ordinary selection skips a dead branch reference", func(t *testing.T) {
		t.Parallel()

		referenced := make([][]byte, 0)
		candidate := newShardCandidate(11, [][]byte{deadMetaHash})
		mp := buildProcessor(true, canonicalPool(), nil, nil, nil, &referenced)

		_, err := mp.SelectIncomingMiniBlocks(newLastShardHdrs(), []data.HeaderHandler{candidate}, [][]byte{[]byte("candidateHash")}, 10, haveTime)
		require.Nil(t, err)
		require.Empty(t, referenced)
	})

	t.Run("ordinary selection skips an unresolvable reference", func(t *testing.T) {
		t.Parallel()

		referenced := make([][]byte, 0)
		candidate := newShardCandidate(11, [][]byte{[]byte("unknownMetaHash")})
		mp := buildProcessor(true, canonicalPool(), nil, nil, nil, &referenced)

		_, err := mp.SelectIncomingMiniBlocks(newLastShardHdrs(), []data.HeaderHandler{candidate}, [][]byte{[]byte("candidateHash")}, 10, haveTime)
		require.Nil(t, err)
		require.Empty(t, referenced)
	})

	t.Run("pre Supernova selection ignores references", func(t *testing.T) {
		t.Parallel()

		referenced := make([][]byte, 0)
		candidate := newShardCandidate(11, [][]byte{deadMetaHash})
		candidateHash := []byte("candidateHash")
		mp := buildProcessor(false, canonicalPool(), nil, nil, nil, &referenced)

		_, err := mp.SelectIncomingMiniBlocks(newLastShardHdrs(), []data.HeaderHandler{candidate}, [][]byte{candidateHash}, 10, haveTime)
		require.Nil(t, err)
		require.Equal(t, [][]byte{candidateHash}, referenced)
	})

	t.Run("arbitration skips a dead branch reference", func(t *testing.T) {
		t.Parallel()

		referenced := make([][]byte, 0)
		candidate := newShardCandidate(14, [][]byte{deadMetaHash})
		mp := buildProcessor(true, canonicalPool(), []data.HeaderHandler{candidate}, [][]byte{[]byte("candidateHash")}, nil, &referenced)

		err := mp.SelectContendedShardHeaders(18, newLastShardHdrs(), map[uint32]uint32{}, haveTime)
		require.Nil(t, err)
		require.Empty(t, referenced)
	})

	t.Run("arbitration keeps an ancestor reference", func(t *testing.T) {
		t.Parallel()

		referenced := make([][]byte, 0)
		candidate := newShardCandidate(14, [][]byte{metaHash99})
		candidateHash := []byte("candidateHash")
		mp := buildProcessor(true, canonicalPool(), []data.HeaderHandler{candidate}, [][]byte{candidateHash}, nil, &referenced)

		err := mp.SelectContendedShardHeaders(18, newLastShardHdrs(), map[uint32]uint32{}, haveTime)
		require.Nil(t, err)
		require.Equal(t, [][]byte{candidateHash}, referenced)
	})

	t.Run("validator rejects a dead branch reference with the dedicated error", func(t *testing.T) {
		t.Parallel()

		referenced := make([][]byte, 0)
		candidate := newShardCandidate(11, [][]byte{deadMetaHash})
		mp := buildProcessor(true, canonicalPool(), nil, nil, nil, &referenced)

		err := mp.CheckHeadersSequenceCorrectness(
			[]blproc.ShardHeaderInfo{{Header: candidate, Hash: []byte("candidateHash")}},
			blproc.ShardHeaderInfo{Header: parentShard, Hash: parentShardHash},
		)
		require.ErrorIs(t, err, blproc.ErrReferencedNonAncestorMetaHeader)
	})

	t.Run("validator accepts ancestor references", func(t *testing.T) {
		t.Parallel()

		referenced := make([][]byte, 0)
		candidate := newShardCandidate(11, [][]byte{metaHash99, headHash})
		mp := buildProcessor(true, canonicalPool(), nil, nil, nil, &referenced)

		err := mp.CheckHeadersSequenceCorrectness(
			[]blproc.ShardHeaderInfo{{Header: candidate, Hash: []byte("candidateHash")}},
			blproc.ShardHeaderInfo{Header: parentShard, Hash: parentShardHash},
		)
		require.Nil(t, err)
	})

	t.Run("reference below the pooled walk falls back to the canonical nonce hash storer", func(t *testing.T) {
		t.Parallel()

		// meta98 is absent from the pool, so the walk stops at nonce 99 and nonce 97 needs the storer
		poolWithGap := map[string]data.HeaderHandler{
			string(metaHash99): meta99,
			string(headHash):   headMeta,
			string(metaHash97): meta97,
		}

		matchingStorer := &storageStubs.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) { return metaHash97, nil },
		}
		referenced := make([][]byte, 0)
		candidate := newShardCandidate(11, [][]byte{metaHash97})
		candidateHash := []byte("candidateHash")
		mp := buildProcessor(true, poolWithGap, nil, nil, matchingStorer, &referenced)

		_, err := mp.SelectIncomingMiniBlocks(newLastShardHdrs(), []data.HeaderHandler{candidate}, [][]byte{candidateHash}, 10, haveTime)
		require.Nil(t, err)
		require.Equal(t, [][]byte{candidateHash}, referenced)

		mismatchStorer := &storageStubs.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) { return []byte("otherCanonicalHash"), nil },
		}
		referenced = make([][]byte, 0)
		mp = buildProcessor(true, poolWithGap, nil, nil, mismatchStorer, &referenced)

		_, err = mp.SelectIncomingMiniBlocks(newLastShardHdrs(), []data.HeaderHandler{candidate}, [][]byte{candidateHash}, 10, haveTime)
		require.Nil(t, err)
		require.Empty(t, referenced)
	})
}

type ancestryCacheProcessor interface {
	CheckReferencedMetaAncestryForProposal(headers []data.HeaderHandler) error
}

type ancestryReadCounters struct {
	poolReads   int
	storerReads map[string]int
}

func (counters *ancestryReadCounters) totalStorerReads() int {
	total := 0
	for _, reads := range counters.storerReads {
		total += reads
	}

	return total
}

func TestMetaProcessor_AncestryCanonicalCache(t *testing.T) {
	t.Parallel()

	metaHash99 := []byte("metaHash99")
	headHash := []byte("metaHash100")
	meta99 := &block.MetaBlock{Nonce: 99, PrevHash: []byte("metaHash98")}
	headMeta := &block.MetaBlock{Nonce: 100, PrevHash: metaHash99}

	canonicalHash := func(nonce uint64) []byte { return []byte(fmt.Sprintf("canonical%d", nonce)) }
	nonceKey := func(nonce uint64) string { return fmt.Sprintf("%d", nonce) }

	// canonical run headers far below the pool walk horizon at nonce 99
	runPool := func(fromNonce uint64, toNonce uint64) map[string]data.HeaderHandler {
		poolHeaders := map[string]data.HeaderHandler{
			string(metaHash99): meta99,
			string(headHash):   headMeta,
		}
		for nonce := fromNonce; nonce <= toNonce; nonce++ {
			poolHeaders[string(canonicalHash(nonce))] = &block.MetaBlock{Nonce: nonce}
		}

		return poolHeaders
	}
	runStorer := func(fromNonce uint64, toNonce uint64) map[string][]byte {
		hashes := make(map[string][]byte)
		for nonce := fromNonce; nonce <= toNonce; nonce++ {
			hashes[nonceKey(nonce)] = canonicalHash(nonce)
		}

		return hashes
	}
	shardCandidate := func(metaHashes [][]byte) *block.Header {
		return &block.Header{ShardID: 0, Nonce: 11, MetaBlockHashes: metaHashes}
	}

	buildProcessor := func(
		poolHeaders map[string]data.HeaderHandler,
		storerHashByKey map[string][]byte,
		counters *ancestryReadCounters,
		poolMissesFirstRequest map[string]bool,
	) ancestryCacheProcessor {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		coreComponents.EnableEpochsHandlerField = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.SupernovaFlag
			},
		}
		coreComponents.UInt64ByteSliceConv = &mock.Uint64ByteSliceConverterMock{
			ToByteSliceCalled: func(nonce uint64) []byte {
				return []byte(nonceKey(nonce))
			},
		}
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled:     func() data.HeaderHandler { return headMeta },
			GetCurrentBlockHeaderHashCalled: func() []byte { return headHash },
		}
		dataComponents.Storage = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				if unitType == dataRetriever.MetaHdrNonceHashDataUnit {
					return &storageStubs.StorerStub{
						GetCalled: func(key []byte) ([]byte, error) {
							counters.storerReads[string(key)]++
							hash, found := storerHashByKey[string(key)]
							if !found {
								return nil, errors.New("nonce hash not found")
							}
							return hash, nil
						},
					}, nil
				}
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						return nil, errors.New("not found")
					},
					SearchFirstCalled: func(key []byte) ([]byte, error) {
						return nil, errors.New("not found")
					},
				}, nil
			},
		}
		if ph, ok := dataComponents.DataPool.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.HeadersCalled = func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
						counters.poolReads++
						if poolMissesFirstRequest[string(hash)] {
							poolMissesFirstRequest[string(hash)] = false
							return nil, errors.New("header not found")
						}
						header, found := poolHeaders[string(hash)]
						if !found {
							return nil, errors.New("header not found")
						}
						return header, nil
					},
				}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		return mp
	}

	t.Run("a below horizon run resolves once per header and reads each nonce once", func(t *testing.T) {
		t.Parallel()

		counters := &ancestryReadCounters{storerReads: make(map[string]int)}
		mp := buildProcessor(runPool(20, 27), runStorer(20, 27), counters, nil)
		firstHeader := shardCandidate([][]byte{canonicalHash(20), canonicalHash(21), canonicalHash(22), canonicalHash(23)})
		secondHeader := shardCandidate([][]byte{canonicalHash(24), canonicalHash(25), canonicalHash(26), canonicalHash(27)})

		err := mp.CheckReferencedMetaAncestryForProposal([]data.HeaderHandler{firstHeader, secondHeader})
		require.Nil(t, err)

		// per header: one reference resolution; plus the pool walk reads of the first probe
		require.Equal(t, 4, counters.poolReads)
		require.Equal(t, 8, counters.totalStorerReads())
		for nonce := uint64(20); nonce <= 27; nonce++ {
			require.Equal(t, 1, counters.storerReads[nonceKey(nonce)])
		}
	})

	t.Run("a second pass over the same run answers from the cache with zero reads", func(t *testing.T) {
		t.Parallel()

		counters := &ancestryReadCounters{storerReads: make(map[string]int)}
		mp := buildProcessor(runPool(20, 27), runStorer(20, 27), counters, nil)
		firstHeader := shardCandidate([][]byte{canonicalHash(20), canonicalHash(21), canonicalHash(22), canonicalHash(23)})
		secondHeader := shardCandidate([][]byte{canonicalHash(24), canonicalHash(25), canonicalHash(26), canonicalHash(27)})

		err := mp.CheckReferencedMetaAncestryForProposal([]data.HeaderHandler{firstHeader, secondHeader, firstHeader, secondHeader})
		require.Nil(t, err)

		require.Equal(t, 4, counters.poolReads)
		require.Equal(t, 8, counters.totalStorerReads())
	})

	t.Run("a dead reference at a covered nonce is rejected without another storer read", func(t *testing.T) {
		t.Parallel()

		deadHash := []byte("deadHash25")
		poolHeaders := runPool(24, 26)
		poolHeaders[string(deadHash)] = &block.MetaBlock{Nonce: 25, Round: 1000}

		counters := &ancestryReadCounters{storerReads: make(map[string]int)}
		mp := buildProcessor(poolHeaders, runStorer(24, 26), counters, nil)
		candidate := shardCandidate([][]byte{canonicalHash(24), canonicalHash(25), canonicalHash(26), deadHash})

		err := mp.CheckReferencedMetaAncestryForProposal([]data.HeaderHandler{candidate})
		require.ErrorIs(t, err, blproc.ErrReferencedNonAncestorMetaHeader)
		require.Equal(t, 1, counters.storerReads[nonceKey(25)])
	})

	t.Run("a pruned storer nonce stays fail closed and is read only once", func(t *testing.T) {
		t.Parallel()

		storerHashes := runStorer(20, 21)
		delete(storerHashes, nonceKey(21))

		counters := &ancestryReadCounters{storerReads: make(map[string]int)}
		mp := buildProcessor(runPool(20, 21), storerHashes, counters, nil)
		candidate := shardCandidate([][]byte{canonicalHash(20), canonicalHash(21)})

		err := mp.CheckReferencedMetaAncestryForProposal([]data.HeaderHandler{candidate})
		require.ErrorIs(t, err, blproc.ErrReferencedNonAncestorMetaHeader)
		require.Equal(t, 1, counters.storerReads[nonceKey(21)])
	})

	t.Run("the pool walk freezes once the canonical region activates", func(t *testing.T) {
		t.Parallel()

		// meta98 reaches the pool only after the walk stopped there; the canonical storer holds a
		// divergent hash at 98, so accepting the late header through the walk would flip the verdict
		metaHash98 := []byte("metaHash98")
		poolHeaders := runPool(20, 20)
		poolHeaders[string(metaHash98)] = &block.MetaBlock{Nonce: 98, PrevHash: []byte("metaHash97")}
		storerHashes := runStorer(20, 20)
		storerHashes[nonceKey(98)] = []byte("otherLocalHash98")

		counters := &ancestryReadCounters{storerReads: make(map[string]int)}
		mp := buildProcessor(poolHeaders, storerHashes, counters, map[string]bool{string(metaHash98): true})
		firstHeader := shardCandidate([][]byte{canonicalHash(20)})
		secondHeader := shardCandidate([][]byte{metaHash98})

		err := mp.CheckReferencedMetaAncestryForProposal([]data.HeaderHandler{firstHeader, secondHeader})
		require.ErrorIs(t, err, blproc.ErrReferencedNonAncestorMetaHeader)
	})

	t.Run("walked window references answer from the walk set without re resolution", func(t *testing.T) {
		t.Parallel()

		counters := &ancestryReadCounters{storerReads: make(map[string]int)}
		mp := buildProcessor(runPool(0, 0), nil, counters, nil)
		candidate := shardCandidate([][]byte{metaHash99, metaHash99, headHash})

		err := mp.CheckReferencedMetaAncestryForProposal([]data.HeaderHandler{candidate})
		require.Nil(t, err)

		// one resolution plus one walk step; the repeat and the parent hash answer from the set
		require.Equal(t, 2, counters.poolReads)
		require.Equal(t, 0, counters.totalStorerReads())
	})
}

type epochStartDataProcessor interface {
	SetComputedEpochStartData(epoch uint32, epochStartData *block.EpochStart)
	GetComputedEpochStartData(epoch uint32) (*block.EpochStart, error)
	VerifyEpochStartData(header data.MetaHeaderHandler) bool
}

// the computed epoch start data must survive the epoch start block commit, so that after an epoch
// boundary rollback the canonical sibling still verifies; the epoch guard confines it to its epoch
func TestMetaProcessor_ComputedEpochStartDataEpochGuard(t *testing.T) {
	t.Parallel()

	epochStartData := &block.EpochStart{
		LastFinalizedHeaders: []block.EpochStartShardData{
			{ShardID: 0, HeaderHash: []byte("shard header hash")},
		},
		Economics: block.Economics{
			PrevEpochStartRound: 100,
		},
	}

	buildProcessor := func(t *testing.T) epochStartDataProcessor {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		return mp
	}

	t.Run("data is served only for its own epoch", func(t *testing.T) {
		t.Parallel()

		mp := buildProcessor(t)
		mp.SetComputedEpochStartData(7, epochStartData)

		computed, err := mp.GetComputedEpochStartData(7)
		require.Nil(t, err)
		require.True(t, computed.Equal(epochStartData))

		_, err = mp.GetComputedEpochStartData(8)
		require.ErrorIs(t, err, process.ErrNilEpochStartData)
		_, err = mp.GetComputedEpochStartData(6)
		require.ErrorIs(t, err, process.ErrNilEpochStartData)
	})

	t.Run("a same parent sibling verifies against the retained data", func(t *testing.T) {
		t.Parallel()

		mp := buildProcessor(t)
		mp.SetComputedEpochStartData(7, epochStartData)

		sibling := &block.MetaBlockV3{Epoch: 7, EpochStart: *epochStartData}
		require.True(t, mp.VerifyEpochStartData(sibling))
	})

	t.Run("an emptied wrapper fails sibling verification", func(t *testing.T) {
		t.Parallel()

		// the old commit time reset produced exactly this state and stalled the boundary
		mp := buildProcessor(t)
		mp.SetComputedEpochStartData(7, &block.EpochStart{})

		sibling := &block.MetaBlockV3{Epoch: 7, EpochStart: *epochStartData}
		require.False(t, mp.VerifyEpochStartData(sibling))
	})

	t.Run("data from another epoch fails verification", func(t *testing.T) {
		t.Parallel()

		mp := buildProcessor(t)
		mp.SetComputedEpochStartData(6, epochStartData)

		sibling := &block.MetaBlockV3{Epoch: 7, EpochStart: *epochStartData}
		require.False(t, mp.VerifyEpochStartData(sibling))
	})
}
