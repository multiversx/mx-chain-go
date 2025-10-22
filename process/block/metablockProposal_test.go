package block_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/executionTrack"
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
