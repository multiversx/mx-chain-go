package storageBootstrap

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func createMockShardStorageBoostrapperArgs() ArgsBaseStorageBootstrapper {
	argsBaseBootstrapper := ArgsBaseStorageBootstrapper{
		BootStorer:                   &mock.BoostrapStorerMock{},
		ForkDetector:                 &mock.ForkDetectorMock{},
		BlockProcessor:               &mock.BlockProcessorMock{},
		ChainHandler:                 &testscommon.ChainHandlerStub{},
		Marshalizer:                  &mock.MarshalizerMock{},
		Store:                        &mock.ChainStorerMock{},
		Uint64Converter:              &mock.Uint64ByteSliceConverterMock{},
		BootstrapRoundIndex:          1,
		ShardCoordinator:             &mock.ShardCoordinatorStub{},
		NodesCoordinator:             &mock.NodesCoordinatorMock{},
		EpochStartTrigger:            &mock.EpochStartTriggerStub{},
		BlockTracker:                 &mock.BlockTrackerMock{},
		ChainID:                      "1",
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		MiniblocksProvider:           &mock.MiniBlocksProviderStub{},
	}

	return argsBaseBootstrapper
}

func TestBaseStorageBootstrapper_CheckBlockBodyIntegrityShouldErrMissingHeader(t *testing.T) {
	t.Parallel()

	baseArgs := createMockShardStorageBoostrapperArgs()
	ssb, _ := NewShardStorageBootstrapper(ArgsShardStorageBootstrapper{ArgsBaseStorageBootstrapper: baseArgs})

	hash := []byte("hash")
	err := ssb.checkBlockBodyIntegrity(hash)
	assert.True(t, strings.Contains(err.Error(), process.ErrMissingHeader.Error()))
}

func TestBaseStorageBootstrapper_CheckBlockBodyIntegrityShouldErrMissingBody(t *testing.T) {
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
	storerMock := mock.NewStorerMock()
	_ = storerMock.Put(headerHash, marshaledHeader)
	baseArgs.Store = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return storerMock
		},
	}
	ssb, _ := NewShardStorageBootstrapper(ArgsShardStorageBootstrapper{ArgsBaseStorageBootstrapper: baseArgs})

	err := ssb.checkBlockBodyIntegrity(headerHash)
	assert.Equal(t, process.ErrMissingBody, err)
}

func TestBaseStorageBootstrapper_CheckBlockBodyIntegrityShouldErrWhenRestoreBlockBodyIntoPoolsFails(t *testing.T) {
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
	baseArgs.BlockProcessor = &mock.BlockProcessorMock{
		RestoreBlockBodyIntoPoolsCalled: func(body data.BodyHandler) error {
			return expectedError
		},
	}
	marshaledHeader, _ := baseArgs.Marshalizer.Marshal(header)
	storerMock := mock.NewStorerMock()
	_ = storerMock.Put(headerHash, marshaledHeader)
	baseArgs.Store = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return storerMock
		},
	}
	ssb, _ := NewShardStorageBootstrapper(ArgsShardStorageBootstrapper{ArgsBaseStorageBootstrapper: baseArgs})

	err := ssb.checkBlockBodyIntegrity(headerHash)
	assert.Equal(t, expectedError, err)
}

func TestBaseStorageBootstrapper_CheckBlockBodyIntegrityShouldWork(t *testing.T) {
	t.Parallel()

	headerHash := []byte("header_hash")
	header := &block.Header{}
	baseArgs := createMockShardStorageBoostrapperArgs()
	baseArgs.MiniblocksProvider = &mock.MiniBlocksProviderStub{
		GetMiniBlocksFromStorerCalled: func(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
			return nil, nil
		},
	}
	baseArgs.BlockProcessor = &mock.BlockProcessorMock{
		RestoreBlockBodyIntoPoolsCalled: func(body data.BodyHandler) error {
			return nil
		},
	}
	marshaledHeader, _ := baseArgs.Marshalizer.Marshal(header)
	storerMock := mock.NewStorerMock()
	_ = storerMock.Put(headerHash, marshaledHeader)
	baseArgs.Store = &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return storerMock
		},
	}
	ssb, _ := NewShardStorageBootstrapper(ArgsShardStorageBootstrapper{ArgsBaseStorageBootstrapper: baseArgs})

	err := ssb.checkBlockBodyIntegrity(headerHash)
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
	ssb, _ := NewShardStorageBootstrapper(ArgsShardStorageBootstrapper{ArgsBaseStorageBootstrapper: baseArgs})

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

	header := &block.Header{}
	baseArgs := createMockShardStorageBoostrapperArgs()
	baseArgs.MiniblocksProvider = &mock.MiniBlocksProviderStub{
		GetMiniBlocksFromStorerCalled: func(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
			return mbAndHashes, nil
		},
	}
	expectedBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			mb1,
			mb2,
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
