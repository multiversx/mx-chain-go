package storagerequesterscontainer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/factory/storageRequestersContainer"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errExpected = errors.New("expected error")

func createMessengerStubForShard(matchStrToErrOnCreate string, matchStrToErrOnRegister string) p2p.Messenger {
	stub := &p2pmocks.MessengerStub{}

	stub.CreateTopicCalled = func(name string, createChannelForTopic bool) error {
		if matchStrToErrOnCreate == "" {
			return nil
		}

		if strings.Contains(name, matchStrToErrOnCreate) {
			return errExpected
		}

		return nil
	}

	stub.RegisterMessageProcessorCalled = func(topic string, identifier string, handler p2p.MessageProcessor) error {
		if matchStrToErrOnRegister == "" {
			return nil
		}

		if strings.Contains(topic, matchStrToErrOnRegister) {
			return errExpected
		}

		return nil
	}

	return stub
}

func createStoreForShard() dataRetriever.StorageService {
	return &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{}, nil
		},
	}
}

func TestNewShardRequestersContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.ShardCoordinator = nil
	rcf, err := storagerequesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewShardRequestersContainerFactory_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Messenger = nil
	rcf, err := storagerequesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewShardRequestersContainerFactory_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Store = nil
	rcf, err := storagerequesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilStore, err)
}

func TestNewShardRequestersContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Marshalizer = nil
	rcf, err := storagerequesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewShardRequestersContainerFactory_NilUint64SliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.Uint64ByteSliceConverter = nil
	rcf, err := storagerequesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
}

func TestNewShardRequestersContainerFactory_NilDataPackerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	args.DataPacker = nil
	rcf, err := storagerequesterscontainer.NewShardRequestersContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilDataPacker, err)
}

func TestNewShardRequestersContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	rcf, err := storagerequesterscontainer.NewShardRequestersContainerFactory(args)

	assert.NotNil(t, rcf)
	assert.Nil(t, err)
	require.False(t, rcf.IsInterfaceNil())
}

func TestShardRequestersContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgumentsShard()
	rcf, _ := storagerequesterscontainer.NewShardRequestersContainerFactory(args)

	container, err := rcf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestShardRequestersContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	args := getArgumentsShard()
	args.ShardCoordinator = shardCoordinator
	rcf, _ := storagerequesterscontainer.NewShardRequestersContainerFactory(args)

	container, _ := rcf.Create()

	numRequesterSCRs := noOfShards + 1
	numRequesterTxs := noOfShards + 1
	numRequesterRewardTxs := 1
	numRequesterHeaders := 1
	numRequesterMiniBlocks := noOfShards + 2
	numRequesterMetaBlockHeaders := 1
	numRequesterTrieNodes := 1
	numPeerAuthentication := 1
	numValidatorInfo := 1
	totalRequesters := numRequesterTxs + numRequesterHeaders + numRequesterMiniBlocks +
		numRequesterMetaBlockHeaders + numRequesterSCRs + numRequesterRewardTxs + numRequesterTrieNodes +
		numPeerAuthentication + numValidatorInfo

	assert.Equal(t, totalRequesters, container.Len())
}

func getArgumentsShard() storagerequesterscontainer.FactoryArgs {
	return storagerequesterscontainer.FactoryArgs{
		GeneralConfig: config.Config{
			AccountsTrieStorage:     getMockStorageConfig(),
			PeerAccountsTrieStorage: getMockStorageConfig(),
			TrieStorageManagerConfig: config.TrieStorageManagerConfig{
				PruningBufferLen:      255,
				SnapshotsBufferLen:    255,
				SnapshotsGoroutineNum: 2,
			},
			StateTriesConfig: config.StateTriesConfig{
				AccountsStatePruningEnabled: false,
				PeerStatePruningEnabled:     false,
				MaxStateTrieLevelInMemory:   5,
				MaxPeerTrieLevelInMemory:    5,
			},
		},
		ShardIDForTries:          0,
		ChainID:                  "T",
		WorkingDirectory:         "",
		Hasher:                   &hashingMocks.HasherMock{},
		ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
		Messenger:                createMessengerStubForShard("", ""),
		Store:                    createStoreForShard(),
		Marshalizer:              &mock.MarshalizerMock{},
		Uint64ByteSliceConverter: &mock.Uint64ByteSliceConverterMock{},
		DataPacker:               &mock.DataPackerStub{},
		ManualEpochStartNotifier: &mock.ManualEpochStartNotifierStub{},
		ChanGracefullyClose:      make(chan endProcess.ArgEndProcess),
		EnableEpochsHandler:      &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
}
