package block_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/executionTrack"
	"github.com/multiversx/mx-chain-go/testscommon/mbSelection"
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
			return &block.MetaExecutionResultInfo{}
		},
	}
	validMetaExecutionResultsWithEpochChange := []data.BaseExecutionResultHandler{
		&block.MetaExecutionResult{
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
		arguments.ExecutionResultsTracker = &executionTrack.ExecutionResultsTrackerStub{
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
		arguments.ExecutionResultsTracker = &executionTrack.ExecutionResultsTrackerStub{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return nil, nil
			},
		}

		metaBlockWithInvalidExecutionResult := validMetaHeaderV3
		metaBlockWithInvalidExecutionResult.GetExecutionResultsHandlersCalled = func() []data.BaseExecutionResultHandler {
			return []data.BaseExecutionResultHandler{
				&block.ExecutionResult{}, // invalid for meta block
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
		arguments.ExecutionResultsTracker = &executionTrack.ExecutionResultsTrackerStub{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return nil, nil
			},
		}

		metaBlockWithValidExecutionResult := validMetaHeaderV3
		metaBlockWithValidExecutionResult.GetExecutionResultsHandlersCalled = func() []data.BaseExecutionResultHandler {
			return []data.BaseExecutionResultHandler{
				&block.MetaExecutionResult{
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
		arguments.ExecutionResultsTracker = &executionTrack.ExecutionResultsTrackerStub{
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

		mp.SetEpochStartData(&block.EpochStart{
			LastFinalizedHeaders: make([]block.EpochStartShardData, 3),
			Economics:            block.Economics{},
		})
		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Equal(t, expectedErr, err)
		require.Nil(t, header)
	})
	t.Run("with epoch start data in execution results and in meta block processor, error on set epoch start data", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ExecutionResultsTracker = &executionTrack.ExecutionResultsTrackerStub{
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

		mp.SetEpochStartData(&block.EpochStart{
			LastFinalizedHeaders: make([]block.EpochStartShardData, 3),
			Economics:            block.Economics{},
		})
		header, err := mp.CreateNewHeaderProposal(1, 1)
		require.Equal(t, expectedErr, err)
		require.Nil(t, header)
	})
	t.Run("without epoch start data in execution results, should pass and not change epoch", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ExecutionResultsTracker = &executionTrack.ExecutionResultsTrackerStub{
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
		arguments.ExecutionResultsTracker = &executionTrack.ExecutionResultsTrackerStub{
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

		mp.SetEpochStartData(&block.EpochStart{
			LastFinalizedHeaders: make([]block.EpochStartShardData, 3),
			Economics:            block.Economics{},
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
			CreateShardInfoV3Called: func(metaHeader data.MetaHeaderHandler, shardHeaders []data.HeaderHandler, shardHeaderHashes [][]byte) ([]data.ShardDataHandler, error) {
				return nil, expectedErr
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
			CreateShardInfoV3Called: func(metaHeader data.MetaHeaderHandler, shardHeaders []data.HeaderHandler, shardHeaderHashes [][]byte) ([]data.ShardDataHandler, error) {
				return []data.ShardDataHandler{invalidShardData}, nil
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
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
	arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	mp, err := blproc.NewMetaProcessor(arguments)
	require.Nil(t, err)

	header := &block.MetaBlock{
		Nonce: 1,
		Round: 1,
	}
	body := &block.Body{}

	err = mp.VerifyBlockProposal(header, body, haveTime)
	require.NoError(t, err)
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
					&block.ExecutionResult{},
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
