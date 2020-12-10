package storageResolversContainers_test

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/storageResolversContainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createStubTopicMessageHandlerForMeta(matchStrToErrOnCreate string, matchStrToErrOnRegister string) dataRetriever.TopicMessageHandler {
	tmhs := mock.NewTopicMessageHandlerStub()

	tmhs.CreateTopicCalled = func(name string, createChannelForTopic bool) error {
		if matchStrToErrOnCreate == "" {
			return nil
		}
		if strings.Contains(name, matchStrToErrOnCreate) {
			return errExpected
		}

		return nil
	}

	tmhs.RegisterMessageProcessorCalled = func(topic string, identifier string, handler p2p.MessageProcessor) error {
		if matchStrToErrOnRegister == "" {
			return nil
		}
		if strings.Contains(topic, matchStrToErrOnRegister) {
			return errExpected
		}

		return nil
	}

	return tmhs
}

func createStoreForMeta() dataRetriever.StorageService {
	return &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{}
		},
	}
}

//------- NewResolversContainerFactory

func TestNewMetaResolversContainerFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.ShardCoordinator = nil
	rcf, err := storageResolversContainers.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
}

func TestNewMetaResolversContainerFactory_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.Messenger = nil
	rcf, err := storageResolversContainers.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewMetaResolversContainerFactory_NilStoreShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.Store = nil
	rcf, err := storageResolversContainers.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilStore, err)
}

func TestNewMetaResolversContainerFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.Marshalizer = nil
	rcf, err := storageResolversContainers.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewMetaResolversContainerFactory_NilUint64SliceConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.Uint64ByteSliceConverter = nil
	rcf, err := storageResolversContainers.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
}

func TestNewMetaResolversContainerFactory_NilDataPackerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	args.DataPacker = nil
	rcf, err := storageResolversContainers.NewMetaResolversContainerFactory(args)

	assert.Nil(t, rcf)
	assert.Equal(t, dataRetriever.ErrNilDataPacker, err)
}

func TestNewMetaResolversContainerFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	rcf, err := storageResolversContainers.NewMetaResolversContainerFactory(args)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(rcf))
}

//------- Create

func TestMetaResolversContainerFactory_CreateShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgumentsMeta()
	rcf, _ := storageResolversContainers.NewMetaResolversContainerFactory(args)

	container, err := rcf.Create()

	assert.NotNil(t, container)
	assert.Nil(t, err)
}

func TestMetaResolversContainerFactory_With4ShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfShards := 4
	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.SetNoShards(uint32(noOfShards))
	shardCoordinator.CurrentShard = 1

	args := getArgumentsMeta()
	args.ShardCoordinator = shardCoordinator
	rcf, err := storageResolversContainers.NewMetaResolversContainerFactory(args)
	require.Nil(t, err)

	container, err := rcf.Create()
	require.Nil(t, err)

	numResolversShardHeadersForMetachain := noOfShards
	numResolverMetablocks := 1
	numResolversMiniBlocks := noOfShards + 2
	numResolversUnsigned := noOfShards + 1
	numResolversRewards := noOfShards
	numResolversTxs := noOfShards + 1
	numResolversTrieNodes := 2
	totalResolvers := numResolversShardHeadersForMetachain + numResolverMetablocks + numResolversMiniBlocks +
		numResolversUnsigned + numResolversTxs + numResolversTrieNodes + numResolversRewards

	assert.Equal(t, totalResolvers, container.Len())
	assert.Equal(t, totalResolvers, container.Len())
}

func getMockStorageConfig() config.StorageConfig {
	return config.StorageConfig{
		Cache: config.CacheConfig{
			Name:     "mock",
			Type:     "LRU",
			Capacity: 1000,
			Shards:   1,
		},
		DB: config.DBConfig{
			FilePath:          "",
			Type:              "MemoryDB",
			BatchDelaySeconds: 1,
			MaxBatchSize:      1,
			MaxOpenFiles:      10,
		},
	}
}

func getArgumentsMeta() storageResolversContainers.FactoryArgs {
	return storageResolversContainers.FactoryArgs{
		GeneralConfig: config.Config{
			AccountsTrieStorage:     getMockStorageConfig(),
			PeerAccountsTrieStorage: getMockStorageConfig(),
			TrieSnapshotDB: config.DBConfig{
				FilePath:          "",
				Type:              "MemoryDB",
				BatchDelaySeconds: 1,
				MaxBatchSize:      1,
				MaxOpenFiles:      10,
			},
			TrieStorageManagerConfig: config.TrieStorageManagerConfig{
				PruningBufferLen:   255,
				SnapshotsBufferLen: 255,
				MaxSnapshots:       255,
			},
			StateTriesConfig: config.StateTriesConfig{
				CheckpointRoundsModulus:     100,
				AccountsStatePruningEnabled: false,
				PeerStatePruningEnabled:     false,
				MaxStateTrieLevelInMemory:   5,
				MaxPeerTrieLevelInMemory:    5,
			},
		},
		ShardIDForTries:          0,
		ChainID:                  "T",
		WorkingDirectory:         "",
		Hasher:                   mock.HasherMock{},
		ShardCoordinator:         mock.NewOneShardCoordinatorMock(),
		Messenger:                createStubTopicMessageHandlerForMeta("", ""),
		Store:                    createStoreForMeta(),
		Marshalizer:              &mock.MarshalizerMock{},
		Uint64ByteSliceConverter: &mock.Uint64ByteSliceConverterMock{},
		DataPacker:               &mock.DataPackerStub{},
		ManualEpochStartNotifier: &mock.ManualEpochStartNotifierStub{},
		ChanGracefullyClose:      make(chan endProcess.ArgEndProcess),
	}
}
