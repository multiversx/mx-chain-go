package block_test

import (
	"bytes"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	state2 "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	retriever "github.com/multiversx/mx-chain-go/dataRetriever"
	integrationTestsMock "github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
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

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ExecutionManager = &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return nil, nil
			},
		}

		metaBlockWithValidExecutionResult := validMetaHeaderV3
		metaBlockWithValidExecutionResult.GetExecutionResultsHandlersCalled = func() []data.BaseExecutionResultHandler {
			return []data.BaseExecutionResultHandler{
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

		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{
			EpochStartData: &block.EpochStart{
				LastFinalizedHeaders: make([]block.EpochStartShardData, 3),
				Economics:            block.Economics{},
			},
		})
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
	t.Run("no mini blocks added if isEpochStart", func(t *testing.T) {
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
			IsEpochStartCalled: func() bool {
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
			"shardCoordinator": mock.NewOneShardCoordinatorMock(),
			"blockTracker": &mock.BlockTrackerMock{
				GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
					return &testscommon.HeaderHandlerStub{}, nil, nil
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

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"shardCoordinator": mock.NewOneShardCoordinatorMock(),
			"blockTracker": &mock.BlockTrackerMock{
				GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
					return &testscommon.HeaderHandlerStub{}, nil, nil
				},
			},
			"dataPool": dataPoolMock,
			"proofsPool": &dataRetrieverMock.ProofsPoolMock{
				HasProofCalled: func(_ uint32, _ []byte) bool {
					return false
				},
			},
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

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"shardCoordinator": mock.NewOneShardCoordinatorMock(),
			"blockTracker": &mock.BlockTrackerMock{
				GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
					return &testscommon.HeaderHandlerStub{}, nil, nil
				},
			},
			"dataPool": dataPoolMock,
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
					return expectedErr
				},
			},
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
			"shardCoordinator": mock.NewOneShardCoordinatorMock(),
			"blockTracker": &mock.BlockTrackerMock{
				GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
					return &testscommon.HeaderHandlerStub{}, nil, nil
				},
			},
			"dataPool": dataPoolMock,
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

		err = mp.CreateProposalMiniBlocks(haveTimeFalse)
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

		err = mp.CreateProposalMiniBlocks(haveTimeTrue)
		require.Equal(t, expectedErr, err)
	})
	t.Run("with time and no error, no mini blocks/shard headers", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.MiniBlocksSelectionSession = miniblockSelectionSessionNoAdd
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		err = mp.CreateProposalMiniBlocks(haveTimeTrue)
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

		err = mp.SelectIncomingMiniBlocksForProposal(haveTimeTrue)
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

		err = mp.SelectIncomingMiniBlocksForProposal(haveTimeTrue)
		require.Equal(t, expectedErr, err)
	})
	t.Run("selection ok", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		err = mp.SelectIncomingMiniBlocksForProposal(haveTimeTrue)
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
		err = mp.SelectIncomingMiniBlocks(lastShardHeaders, orderedHeaders, orderedHeaderHashes, maxNumHeadersFromSameShard, haveTimeTrue)
		require.Nil(t, err)
	})

	t.Run("time is up before processing any header", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		// ensure proofs exist but haveTime will stop immediately
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() retriever.ProofsPool {
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

		err = mp.SelectIncomingMiniBlocks(lastShardHeaders, orderedHeaders, orderedHeaderHashes, 2, haveTimeFalse)
		require.Nil(t, err)
		require.Equal(t, 0, addRefCnt)
	})

	t.Run("maximum shard headers allowed in one meta block reached (max=0)", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() retriever.ProofsPool {
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
		err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{[]byte("h1")}, 0, haveTimeTrue)
		require.Nil(t, err)
		require.Equal(t, 0, called)
	})

	t.Run("skip header due to nonce gap", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() retriever.ProofsPool {
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
		err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{[]byte("h1")}, 2, haveTimeTrue)
		require.Nil(t, err)
		require.Equal(t, 0, cntAddRef)
	})

	t.Run("skip header due to per-shard limit", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() retriever.ProofsPool {
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
		err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h1, h2}, [][]byte{[]byte("h1"), []byte("h2")}, 1, haveTimeTrue)
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
			ph.ProofsCalled = func() retriever.ProofsPool {
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
		err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{[]byte("h1")}, 2, haveTimeTrue)
		require.Nil(t, err)
		require.Equal(t, 0, cntAddRef)
	})

	t.Run("no cross mini blocks with dst me -> add referenced header only", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() retriever.ProofsPool {
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
		err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{[]byte("h1")}, 2, haveTimeTrue)
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
			ph.ProofsCalled = func() retriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{HasProofCalled: func(shardID uint32, headerHash []byte) bool { return true }}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, []block.MiniblockAndHash, uint32, bool, error) {
				return nil, nil, 0, false, expectedErr
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
		err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{[]byte("h1")}, 2, haveTimeTrue)
		require.Equal(t, expectedErr, err)
	})

	t.Run("pending mini blocks returned -> break without adding header", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() retriever.ProofsPool {
				return &dataRetrieverMock.ProofsPoolMock{HasProofCalled: func(shardID uint32, headerHash []byte) bool { return true }}
			}
		}
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		cntAddRef := 0
		arguments.MiniBlocksSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{AddReferencedHeaderCalled: func(metaBlock data.HeaderHandler, metaBlockHash []byte) { cntAddRef++ }}
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, []block.MiniblockAndHash, uint32, bool, error) {
				return nil, []block.MiniblockAndHash{{}}, 0, false, nil
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		lastShardHeaders := createLastShardHeadersNotGenesis()
		h1 := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 0 }, GetNonceCalled: func() uint64 { return 11 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{"mb": 1} }}
		h2 := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 0 }, GetNonceCalled: func() uint64 { return 12 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{"mb": 1} }}
		err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h1, h2}, [][]byte{[]byte("h1"), []byte("h2")}, 2, haveTimeTrue)
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
			ph.ProofsCalled = func() retriever.ProofsPool {
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
			CreateMbsCrossShardDstMeCalled: func(header data.HeaderHandler, processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo) ([]block.MiniblockAndHash, []block.MiniblockAndHash, uint32, bool, error) {
				return []block.MiniblockAndHash{{}}, nil, 3, true, nil
			},
		}
		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		lastShardHeaders := createLastShardHeadersNotGenesis()
		h := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 0 }, GetNonceCalled: func() uint64 { return 11 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{"mb": 1} }}
		err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{[]byte("h1")}, 2, haveTimeTrue)
		require.Nil(t, err)
		require.Equal(t, 1, cntAddMbs)
		require.Equal(t, 1, cntAddRef)
		// last shard header updated and marked used
		require.True(t, lastShardHeaders[0].UsedInBlock)
		require.Equal(t, []byte("h1"), lastShardHeaders[0].Hash)
	})
}

func TestMetaProcessor_selectIncomingMiniBlocks_GapsAndDuplicates(t *testing.T) {
	t.Parallel()

	// helper to build a MetaProcessor with proofs pool behavior
	type metaSel interface {
		SelectIncomingMiniBlocks(lastShardHdr map[uint32]blproc.ShardHeaderInfo, orderedHdrs []data.HeaderHandler, orderedHdrsHashes [][]byte, maxNumHeadersFromSameShard uint32, haveTime func() bool) error
	}
	buildMp := func(hasProofFn func(shardID uint32, headerHash []byte) bool) metaSel {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() retriever.ProofsPool {
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
		err := mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{}, 2, haveTimeTrue)
		require.Equal(t, process.ErrInconsistentShardHeadersAndHashes, err)
	})

	t.Run("missing last shard header for ordered header -> error", func(t *testing.T) {
		t.Parallel()

		mp := buildMp(func(uint32, []byte) bool { return true })
		lastShardHeaders := createLastShardHeadersNotGenesis()
		// header from shard 99, not present in lastShardHeaders map
		h := &testscommon.HeaderHandlerStub{GetShardIDCalled: func() uint32 { return 99 }, GetNonceCalled: func() uint64 { return 1 }, GetMiniBlockHeadersWithDstCalled: func(uint32) map[string]uint32 { return map[string]uint32{} }}
		err := mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h}, [][]byte{[]byte("h1")}, 2, haveTimeTrue)
		require.Equal(t, process.ErrMissingHeader, err)
	})

	t.Run("duplicate nonce: first has proof accepted, second skipped", func(t *testing.T) {
		t.Parallel()

		cntAddRef := 0
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		pools := dataComponents.DataPool
		if ph, ok := pools.(*dataRetrieverMock.PoolsHolderStub); ok {
			ph.ProofsCalled = func() retriever.ProofsPool {
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
		err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h1, h2}, [][]byte{[]byte("h1"), []byte("h2")}, 2, haveTimeTrue)
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
			ph.ProofsCalled = func() retriever.ProofsPool {
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
		err = mp.SelectIncomingMiniBlocks(lastShardHeaders, []data.HeaderHandler{h1, h2}, [][]byte{[]byte("h1"), []byte("h2")}, 2, haveTimeTrue)
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
		dataPoolMock := &dataRetrieverMock.PoolsHolderMock{}
		dataPoolMock.SetHeadersPool(headersPoolMock)

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"dataPool": dataPoolMock,
		})
		require.Nil(t, err)

		_, err = mp.HasExecutionResultsForProposedEpochChange(metaHeader)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should error ErrWrongTypeAssertion", func(t *testing.T) {
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
				return nil, nil
			},
		}
		dataPoolMock := &dataRetrieverMock.PoolsHolderMock{}
		dataPoolMock.SetHeadersPool(headersPoolMock)

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"dataPool": dataPoolMock,
		})
		require.Nil(t, err)

		_, err = mp.HasExecutionResultsForProposedEpochChange(metaHeader)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
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
		dataPoolMock := &dataRetrieverMock.PoolsHolderMock{}
		dataPoolMock.SetHeadersPool(headersPoolMock)

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"dataPool": dataPoolMock,
		})
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
		dataPoolMock := &dataRetrieverMock.PoolsHolderMock{}
		dataPoolMock.SetHeadersPool(headersPoolMock)

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"dataPool": dataPoolMock,
		})
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

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"dataPool": dataPoolMock,
			"blockChain": &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return &block.MetaBlockV3{}
				},
			},
		})
		require.Nil(t, err)

		err = mp.CheckEpochCorrectnessV3(metaHeader)
		require.ErrorIs(t, err, expectedErr)
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
		})
		require.Nil(t, err)

		err = mp.CheckHeadersSequenceCorrectness([]blproc.ShardHeaderInfo{
			{
				Header: &block.Header{Nonce: 2},
			},
		}, blproc.ShardHeaderInfo{})
		require.Equal(t, expectedErr, err)
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

	t.Run("should return err from SaveNodesCoordinatorUpdates", func(t *testing.T) {
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

	t.Run("should return err from ToggleUnStakeUnBond", func(t *testing.T) {
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

		_, err = mp.ProcessEpochStartProposeBlock(&block.MetaBlockV3{}, nil)
		require.Equal(t, process.ErrNilBlockBody, err)
	})

	t.Run("should return ErrEpochStartProposeBlockHasMiniBlocks because the body has mini blocks", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, boostrapComponents, statusComponents)

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		_, err = mp.ProcessEpochStartProposeBlock(&block.MetaBlockV3{}, &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{},
			},
		})
		require.Equal(t, process.ErrEpochStartProposeBlockHasMiniBlocks, err)
	})

	t.Run("should return error on processEconomicsDataForEpochStartProposeBlock from CreateEpochStartShardDataMetablockV3", func(t *testing.T) {
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

		_, err = mp.ProcessEpochStartProposeBlock(&block.MetaBlockV3{}, &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("should return error on processing epoch start mini blocks, when getting the validator statistics root hash", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		blockchainMock := &testscommon.ChainHandlerMock{}
		err := blockchainMock.SetGenesisHeader(&block.Header{})
		require.Nil(t, err)
		blockchainMock.SetLastExecutionResult(&block.MetaExecutionResult{})
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
		_, err = mp.ProcessEpochStartProposeBlock(&block.MetaBlockV3{}, &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("should return error when updating validator statistics", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		blockchainMock := &testscommon.ChainHandlerMock{}
		err := blockchainMock.SetGenesisHeader(&block.Header{})
		require.Nil(t, err)
		blockchainMock.SetLastExecutionResult(&block.MetaExecutionResult{})
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

		_, err = mp.ProcessEpochStartProposeBlock(&block.MetaBlockV3{}, &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("should return error when committing the state", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		blockchainMock := &testscommon.ChainHandlerMock{}
		err := blockchainMock.SetGenesisHeader(&block.Header{})
		require.Nil(t, err)
		blockchainMock.SetLastExecutionResult(&block.MetaExecutionResult{})
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
		accounts := &state2.AccountsStub{
			CommitCalled: func() ([]byte, error) {
				return nil, expectedErr
			},
		}
		accountsDb[state.UserAccountsState] = accounts
		accountsDb[state.PeerAccountsState] = accounts
		arguments.AccountsDB = accountsDb

		mp, err := blproc.NewMetaProcessor(arguments)
		require.Nil(t, err)

		mp.SetEpochStartData(&blproc.EpochStartDataWrapper{})

		_, err = mp.ProcessEpochStartProposeBlock(&block.MetaBlockV3{}, &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("should return error from HandleProcessErrorCutoff", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		blockchainMock := &testscommon.ChainHandlerMock{}
		err := blockchainMock.SetGenesisHeader(&block.Header{})
		require.Nil(t, err)
		blockchainMock.SetLastExecutionResult(&block.MetaExecutionResult{})
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

		_, err = mp.ProcessEpochStartProposeBlock(&block.MetaBlockV3{}, &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("should return error when calculating the hash", func(t *testing.T) {
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
		blockchainMock.SetLastExecutionResult(&block.MetaExecutionResult{})
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

		_, err = mp.ProcessEpochStartProposeBlock(&block.MetaBlockV3{}, &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, boostrapComponents, statusComponents := createMockComponentHolders()
		blockchainMock := &testscommon.ChainHandlerMock{}
		err := blockchainMock.SetGenesisHeader(&block.Header{})
		require.Nil(t, err)
		blockchainMock.SetLastExecutionResult(&block.MetaExecutionResult{})
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

		_, err = mp.ProcessEpochStartProposeBlock(&block.MetaBlockV3{
			LastExecutionResult: &block.MetaExecutionResultInfo{
				ExecutionResult: &block.BaseMetaExecutionResult{
					AccumulatedFeesInEpoch: big.NewInt(0),
					DevFeesInEpoch:         big.NewInt(0),
				},
			},
		}, &block.Body{})
		require.Nil(t, err)
	})
}

func TestMetaProcessor_processEconomicsDataForEpochStartProposeBlock(t *testing.T) {
	t.Parallel()

	t.Run("should return ErrNilBaseExecutionResult error", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"blockChain": &testscommon.ChainHandlerStub{
				GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
					return nil
				},
			},
		})
		require.Nil(t, err)

		err = mp.ProcessEconomicsDataForEpochStartProposeBlock(&block.MetaBlockV3{})
		require.ErrorContains(t, err, process.ErrNilBaseExecutionResult.Error())
	})

	t.Run("should return ErrWrongTypeAssertion error because of wrong type of last execution result", func(t *testing.T) {
		t.Parallel()

		mp, err := blproc.ConstructPartialMetaBlockProcessorForTest(map[string]interface{}{
			"blockChain": &testscommon.ChainHandlerStub{
				GetLastExecutionResultCalled: func() data.BaseExecutionResultHandler {
					return &block.ExecutionResult{}
				},
			},
		})
		require.Nil(t, err)

		err = mp.ProcessEconomicsDataForEpochStartProposeBlock(&block.MetaBlockV3{})
		require.Equal(t, common.ErrWrongTypeAssertion, err)
	})

	t.Run("should return error from CreateEpochStartShardDataMetablockV3", func(t *testing.T) {
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

		err = mp.ProcessEconomicsDataForEpochStartProposeBlock(&block.MetaBlockV3{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("should return error from ComputeEndOfEpochEconomicsV3", func(t *testing.T) {
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

		err = mp.ProcessEconomicsDataForEpochStartProposeBlock(&block.MetaBlockV3{})
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
		err = mp.ProcessEconomicsDataForEpochStartProposeBlock(&block.MetaBlockV3{})
		require.Nil(t, err)
	})
}

func TestMetaProcessor_createExecutionResult(t *testing.T) {
	t.Parallel()

	t.Run("should return ErrGasUsedExceedsGasProvided error", func(t *testing.T) {
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

	t.Run("should fail because of error on CreateReceiptsHash", func(t *testing.T) {
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

	t.Run("should fail because of marshal error", func(t *testing.T) {
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

		metaExecResult, ok := execResult.(*block.MetaExecutionResult)
		require.True(t, ok)
		require.Equal(t, metaExecResult.ExecutedTxCount, uint64(2))
		require.Equal(t, metaExecResult.ReceiptsHash, []byte("receiptHash"))
		require.Equal(t, metaExecResult.GetValidatorStatsRootHash(), []byte("valStatRootHash"))
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
