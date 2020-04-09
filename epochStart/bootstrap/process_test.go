package bootstrap

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	triesFactory "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

func createMockEpochStartBootstrapArgs() ArgsEpochStartBootstrap {
	return ArgsEpochStartBootstrap{
		PublicKey:                  &mock.PublicKeyStub{},
		Marshalizer:                &mock.MarshalizerMock{},
		TxSignMarshalizer:          &mock.MarshalizerMock{},
		Hasher:                     &mock.HasherMock{},
		Messenger:                  &mock.MessengerStub{},
		GeneralConfig:              config.Config{},
		EconomicsData:              &economics.EconomicsData{},
		SingleSigner:               &mock.SignerStub{},
		BlockSingleSigner:          &mock.SignerStub{},
		KeyGen:                     &mock.KeyGenMock{},
		BlockKeyGen:                &mock.KeyGenMock{},
		GenesisNodesConfig:         &mock.NodesSetupStub{},
		GenesisShardCoordinator:    mock.NewMultipleShardsCoordinatorMock(),
		PathManager:                &mock.PathManagerStub{},
		WorkingDir:                 "test_directory",
		DefaultDBPath:              "test_db",
		DefaultEpochString:         "test_epoch",
		DefaultShardString:         "test_shard",
		Rater:                      &mock.RaterStub{},
		DestinationShardAsObserver: 0,
		TrieContainer: &mock.TriesHolderMock{
			GetCalled: func(bytes []byte) data.Trie {
				return &mock.TrieStub{}
			},
		},
		TrieStorageManagers: map[string]data.StorageManager{
			triesFactory.UserAccountTrie: &mock.StorageManagerStub{},
		},
		Uint64Converter: &mock.Uint64ByteSliceConverterMock{},
		NodeShuffler:    &mock.NodeShufflerMock{},
	}
}

func TestNewEpochStartBootstrap(t *testing.T) {
	t.Parallel()

	args := createMockEpochStartBootstrapArgs()

	epochStartProvider, err := NewEpochStartBootstrap(args)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(epochStartProvider))
}

func TestComputedDurationOfEpoch(t *testing.T) {
	t.Parallel()

	roundsPerEpoch := int64(100)
	roundDuration := uint64(6)
	args := createMockEpochStartBootstrapArgs()
	args.GenesisNodesConfig = &mock.NodesSetupStub{
		GetRoundDurationCalled: func() uint64 {
			return roundDuration
		},
	}
	args.GeneralConfig = config.Config{
		EpochStartConfig: config.EpochStartConfig{
			RoundsPerEpoch: roundsPerEpoch,
		},
	}

	epochStartProvider, _ := NewEpochStartBootstrap(args)

	expectedDuration := time.Duration(roundDuration*uint64(roundsPerEpoch)) * time.Millisecond
	resultDuration := epochStartProvider.computedDurationOfEpoch()
	assert.Equal(t, expectedDuration, resultDuration)
}

func TestIsStartInEpochZero(t *testing.T) {
	t.Parallel()

	args := createMockEpochStartBootstrapArgs()
	args.GenesisNodesConfig = &mock.NodesSetupStub{
		GetStartTimeCalled: func() int64 {
			return 1000
		},
	}

	epochStartProvider, _ := NewEpochStartBootstrap(args)

	result := epochStartProvider.isStartInEpochZero()
	assert.False(t, result)
}

func TestEpochStartBootstrap_BootstrapStartInEpochNotEnable(t *testing.T) {
	args := createMockEpochStartBootstrapArgs()

	epochStartProvider, _ := NewEpochStartBootstrap(args)

	params, err := epochStartProvider.Bootstrap()
	assert.Nil(t, err)
	assert.NotNil(t, params)
}

func TestEpochStartBootstrap_Bootstrap(t *testing.T) {
	roundsPerEpoch := int64(100)
	roundDuration := uint64(60000)
	args := createMockEpochStartBootstrapArgs()
	args.GenesisNodesConfig = &mock.NodesSetupStub{
		GetRoundDurationCalled: func() uint64 {
			return roundDuration
		},
	}
	args.GeneralConfig = config.Config{
		EpochStartConfig: config.EpochStartConfig{
			RoundsPerEpoch: roundsPerEpoch,
		},
	}
	args.GeneralConfig = getGeneralConfig()

	epochStartProvider, _ := NewEpochStartBootstrap(args)

	done := make(chan bool, 1)

	go func() {
		_, _ = epochStartProvider.Bootstrap()
		<-done
	}()

	for {
		select {
		case <-done:
			assert.Fail(t, "should not be reach")
		case <-time.After(time.Second):
			assert.True(t, true, "pass with timeout")
			return
		}
	}
}

func TestPrepareForEpochZero(t *testing.T) {
	args := createMockEpochStartBootstrapArgs()

	epochStartProvider, _ := NewEpochStartBootstrap(args)

	params, err := epochStartProvider.prepareEpochZero()
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), params.Epoch)
}

func TestCreateSyncers(t *testing.T) {
	args := createMockEpochStartBootstrapArgs()

	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.shardCoordinator = mock.NewMultipleShardsCoordinatorMock()
	epochStartProvider.dataPool = &mock.PoolsHolderStub{
		HeadersCalled: func() dataRetriever.HeadersPool {
			return &mock.HeadersCacherStub{}
		},
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &mock.ShardedDataStub{}
		},
		MiniBlocksCalled: func() storage.Cacher {
			return &mock.CacherStub{}
		},
		TrieNodesCalled: func() storage.Cacher {
			return &mock.CacherStub{}
		},
	}
	epochStartProvider.whiteListHandler = &mock.WhiteListHandlerStub{}
	epochStartProvider.requestHandler = &mock.RequestHandlerStub{}

	err := epochStartProvider.createSyncers()
	assert.Nil(t, err)
}
