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
	epochNotifierMock "github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

func createMockShardStorageBoostrapperArgs() ArgsBaseStorageBootstrapper {
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
	}

	return argsBaseBootstrapper
}

func TestBaseStorageBootstrapper_CheckBaseStorageBootstrapperArguments(t *testing.T) {
	t.Parallel()

	t.Run("nil bootstorer should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.BootStorer = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilBootStorer, err)
	})
	t.Run("nil fork detector should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.ForkDetector = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilForkDetector, err)
	})
	t.Run("nil block processor should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.BlockProcessor = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilBlockProcessor, err)
	})
	t.Run("nil chain handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.ChainHandler = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilBlockChain, err)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.Marshalizer = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilMarshalizer, err)
	})
	t.Run("nil store should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.Store = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilStore, err)
	})
	t.Run("nil uint64 converter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.Uint64Converter = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilUint64Converter, err)
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.ShardCoordinator = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilShardCoordinator, err)
	})
	t.Run("nil nodes coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.NodesCoordinator = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilNodesCoordinator, err)
	})
	t.Run("nil epoch start trigger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.EpochStartTrigger = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilEpochStartTrigger, err)
	})
	t.Run("nil block tracker should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.BlockTracker = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilBlockTracker, err)
	})
	t.Run("nil scheduled txs execution should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.ScheduledTxsExecutionHandler = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilScheduledTxsExecutionHandler, err)
	})
	t.Run("nil miniblocks provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.MiniblocksProvider = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilMiniBlocksProvider, err)
	})
	t.Run("nil epoch notifier should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.EpochNotifier = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilEpochNotifier, err)
	})
	t.Run("nil processed mini blocks tracker should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.ProcessedMiniBlocksTracker = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilProcessedMiniBlocksTracker, err)
	})
	t.Run("nil app status handler - should error", func(t *testing.T) {
		t.Parallel()

		args := createMockShardStorageBoostrapperArgs()
		args.AppStatusHandler = nil

		err := checkBaseStorageBootstrapperArguments(args)
		assert.Equal(t, process.ErrNilAppStatusHandler, err)
	})
}

func TestBaseStorageBootstrapper_RestoreBlockBodyIntoPoolsShouldErrMissingHeader(t *testing.T) {
	t.Parallel()

	baseArgs := createMockShardStorageBoostrapperArgs()
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

	baseArgs := createMockShardStorageBoostrapperArgs()
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

	baseArgs := createMockShardStorageBoostrapperArgs()
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

	baseArgs := createMockShardStorageBoostrapperArgs()
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

	baseArgs := createMockShardStorageBoostrapperArgs()
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

	baseArgs := createMockShardStorageBoostrapperArgs()
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
