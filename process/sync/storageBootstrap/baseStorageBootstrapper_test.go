package storageBootstrap

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	epochNotifierMock "github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockShardStorageBootstrapperArgs() ArgsBaseStorageBootstrapper {
	argsBaseBootstrapper := ArgsBaseStorageBootstrapper{
		BootStorer:     &mock.BoostrapStorerMock{},
		ForkDetector:   &mock.ForkDetectorMock{},
		BlockProcessor: &testscommon.BlockProcessorStub{},
		ChainHandler:   &testscommon.ChainHandlerStub{},
		Marshalizer:    &mock.MarshalizerMock{},
		Store: &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{}, nil
			},
		},
		Uint64Converter:              &mock.Uint64ByteSliceConverterMock{},
		BootstrapRoundIndex:          1,
		ShardCoordinator:             &mock.ShardCoordinatorStub{},
		NodesCoordinator:             &shardingMocks.NodesCoordinatorMock{},
		EpochStartTrigger:            &mock.EpochStartTriggerStub{},
		BlockTracker:                 &mock.BlockTrackerMock{},
		ChainID:                      "1",
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		MiniblocksProvider:           &mock.MiniBlocksProviderStub{},
		EpochNotifier:                &epochNotifierMock.EpochNotifierStub{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
		AppStatusHandler:             &statusHandler.AppStatusHandlerMock{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ProofsPool:                   &dataRetrieverMocks.ProofsPoolMock{},
		ExecutionManager:             &processMocks.ExecutionManagerMock{},
	}

	return argsBaseBootstrapper
}

func TestBaseStorageBootstrapper_CheckBaseStorageBootstrapperArguments(t *testing.T) {
	t.Parallel()

	t.Run("nil bootstorer should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.BootStorer = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilBootStorer, err)
	})
	t.Run("nil fork detector should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.ForkDetector = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilForkDetector, err)
	})
	t.Run("nil block processor should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.BlockProcessor = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilBlockProcessor, err)
	})
	t.Run("nil chain handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.ChainHandler = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilBlockChain, err)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.Marshalizer = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilMarshalizer, err)
	})
	t.Run("nil store should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.Store = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilStore, err)
	})
	t.Run("nil uint64 converter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.Uint64Converter = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilUint64Converter, err)
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.ShardCoordinator = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilShardCoordinator, err)
	})
	t.Run("nil nodes coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.NodesCoordinator = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilNodesCoordinator, err)
	})
	t.Run("nil epoch start trigger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.EpochStartTrigger = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilEpochStartTrigger, err)
	})
	t.Run("nil block tracker should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.BlockTracker = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilBlockTracker, err)
	})
	t.Run("nil scheduled txs execution should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.ScheduledTxsExecutionHandler = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilScheduledTxsExecutionHandler, err)
	})
	t.Run("nil miniblocks provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.MiniblocksProvider = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilMiniBlocksProvider, err)
	})
	t.Run("nil epoch notifier should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.EpochNotifier = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilEpochNotifier, err)
	})
	t.Run("nil processed mini blocks tracker should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.ProcessedMiniBlocksTracker = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilProcessedMiniBlocksTracker, err)
	})
	t.Run("nil app status handler - should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.AppStatusHandler = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilAppStatusHandler, err)
	})
	t.Run("nil execution manager - should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBootstrapperArgs()
		args.ExecutionManager = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilExecutionManager, err)
	})
}

func TestBaseStorageBootstrapper_RestoreBlockBodyIntoPoolsShouldErrMissingHeader(t *testing.T) {
	t.Parallel()

	baseArgs := createMockShardStorageBootstrapperArgs()
	baseArgs.Store = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					return nil, errors.New("key not found")
				},
			}, nil
		},
	}
	args := ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper: baseArgs,
	}
	ssb, _ := NewShardStorageBootstrapper(args)

	hash := []byte("hash")
	err := ssb.restoreBlockBodyIntoPools(hash)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), process.ErrMissingHeader.Error()))
}

func TestBaseStorageBootstrapper_RestoreBlockBodyIntoPoolsShouldErrMissingBody(t *testing.T) {
	t.Parallel()

	headerHash := []byte("header_hash")
	header := &block.Header{}

	baseArgs := createMockShardStorageBootstrapperArgs()
	baseArgs.MiniblocksProvider = &mock.MiniBlocksProviderStub{
		GetMiniBlocksFromStorerCalled: func(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
			return nil, [][]byte{[]byte("missing_hash")}
		},
	}
	marshaledHeader, _ := baseArgs.Marshalizer.Marshal(header)
	storerMock := genericMocks.NewStorerMock()
	_ = storerMock.Put(headerHash, marshaledHeader)
	baseArgs.Store = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return storerMock, nil
		},
	}
	args := ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper: baseArgs,
	}
	ssb, _ := NewShardStorageBootstrapper(args)

	err := ssb.restoreBlockBodyIntoPools(headerHash)
	assert.Equal(t, process.ErrMissingBody, err)
}

func TestBaseStorageBootstrapper_RestoreBlockBodyIntoPoolsShouldErrWhenRestoreBlockBodyIntoPoolsFails(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("error")
	headerHash := []byte("header_hash")
	header := &block.Header{}

	baseArgs := createMockShardStorageBootstrapperArgs()
	baseArgs.MiniblocksProvider = &mock.MiniBlocksProviderStub{
		GetMiniBlocksFromStorerCalled: func(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
			return nil, nil
		},
	}
	baseArgs.BlockProcessor = &testscommon.BlockProcessorStub{
		RestoreBlockBodyIntoPoolsCalled: func(body data.BodyHandler) error {
			return expectedError
		},
	}
	marshaledHeader, _ := baseArgs.Marshalizer.Marshal(header)
	storerMock := genericMocks.NewStorerMock()
	_ = storerMock.Put(headerHash, marshaledHeader)
	baseArgs.Store = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return storerMock, nil
		},
	}
	args := ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper: baseArgs,
	}
	ssb, _ := NewShardStorageBootstrapper(args)

	err := ssb.restoreBlockBodyIntoPools(headerHash)
	assert.Equal(t, expectedError, err)
}

func TestBaseStorageBootstrapper_RestoreBlockBodyIntoPoolsShouldWork(t *testing.T) {
	t.Parallel()

	headerHash := []byte("header_hash")
	header := &block.Header{}

	baseArgs := createMockShardStorageBootstrapperArgs()
	baseArgs.MiniblocksProvider = &mock.MiniBlocksProviderStub{
		GetMiniBlocksFromStorerCalled: func(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
			return nil, nil
		},
	}
	baseArgs.BlockProcessor = &testscommon.BlockProcessorStub{
		RestoreBlockBodyIntoPoolsCalled: func(body data.BodyHandler) error {
			return nil
		},
	}
	marshaledHeader, _ := baseArgs.Marshalizer.Marshal(header)
	storerMock := genericMocks.NewStorerMock()
	_ = storerMock.Put(headerHash, marshaledHeader)
	baseArgs.Store = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return storerMock, nil
		},
	}
	args := ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper: baseArgs,
	}
	ssb, _ := NewShardStorageBootstrapper(args)

	err := ssb.restoreBlockBodyIntoPools(headerHash)
	assert.Nil(t, err)
}

func TestBaseStorageBootstrapper_GetBlockBodyShouldErrMissingBody(t *testing.T) {
	t.Parallel()

	header := &block.Header{}

	baseArgs := createMockShardStorageBootstrapperArgs()
	baseArgs.MiniblocksProvider = &mock.MiniBlocksProviderStub{
		GetMiniBlocksFromStorerCalled: func(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
			return nil, [][]byte{[]byte("missing_hash")}
		},
	}
	args := ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper: baseArgs,
	}
	ssb, _ := NewShardStorageBootstrapper(args)

	body, err := ssb.getBlockBody(header)
	assert.Nil(t, body)
	assert.Equal(t, process.ErrMissingBody, err)
}

func TestBaseStorageBootstrapper_GetBlockBodyShouldWork(t *testing.T) {
	t.Parallel()

	mb1 := &block.MiniBlock{
		SenderShardID:   0,
		ReceiverShardID: 1,
	}
	mb2 := &block.MiniBlock{
		SenderShardID:   0,
		ReceiverShardID: 2,
	}
	mbAndHashes := []*block.MiniblockAndHash{
		{
			Hash:      []byte("mbHash1"),
			Miniblock: mb1,
		},
		{
			Hash:      []byte("mbHash2"),
			Miniblock: mb2,
		},
	}
	expectedBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			mb1,
			mb2,
		},
	}
	header := &block.Header{}

	baseArgs := createMockShardStorageBootstrapperArgs()
	baseArgs.MiniblocksProvider = &mock.MiniBlocksProviderStub{
		GetMiniBlocksFromStorerCalled: func(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
			return mbAndHashes, nil
		},
	}
	args := ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper: baseArgs,
	}
	ssb, _ := NewShardStorageBootstrapper(args)

	body, err := ssb.getBlockBody(header)
	assert.Nil(t, err)
	assert.Equal(t, expectedBody, body)
}

func TestBaseStorageBootstrapper_setCurrentBlockInfoV3(t *testing.T) {
	t.Parallel()

	t.Run("in case of nil LastExecutionResult should fail", func(t *testing.T) {
		t.Parallel()

		baseArgs := createMockShardStorageBootstrapperArgs()
		args := ArgsShardStorageBootstrapper{
			ArgsBaseStorageBootstrapper: baseArgs,
		}

		ssb, _ := NewShardStorageBootstrapper(args)
		err := ssb.setCurrentBlockInfoV3(&block.HeaderV3{
			LastExecutionResult: nil,
		}, nil)

		require.Equal(t, process.ErrNilLastExecutionResultHandler, err)
	})

	t.Run("if getting the header fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		errExpected := errors.New("expected error")
		baseArgs := createMockShardStorageBootstrapperArgs()
		baseArgs.Marshalizer = &testscommon.MarshallerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errExpected
			},
		}
		args := ArgsShardStorageBootstrapper{
			ArgsBaseStorageBootstrapper: baseArgs,
		}

		ssb, _ := NewShardStorageBootstrapper(args)
		err := ssb.setCurrentBlockInfoV3(&block.HeaderV3{
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderHash: []byte("hashExecResult"),
				},
			},
		}, []byte("hash"))

		require.Equal(t, process.ErrUnmarshalWithoutSuccess, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		baseArgs := createMockShardStorageBootstrapperArgs()
		baseArgs.Marshalizer = &testscommon.MarshallerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return nil
			},
		}
		args := ArgsShardStorageBootstrapper{
			ArgsBaseStorageBootstrapper: baseArgs,
		}

		ssb, _ := NewShardStorageBootstrapper(args)
		err := ssb.setCurrentBlockInfoV3(&block.HeaderV3{
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderHash: []byte("hashExecResult"),
				},
			},
		}, []byte("hash"))

		require.Nil(t, err)
	})
}
