package dataprocessor_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/dataprocessor"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/mock"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/require"
)

func TestNewRatingsProcessor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argsFunc func() dataprocessor.RatingProcessorArgs
		exError  error
	}{
		{
			name: "NilShardCoordinator",
			argsFunc: func() dataprocessor.RatingProcessorArgs {
				args := getRatingsProcessorArgs()
				args.ShardCoordinator = nil
				return args
			},
			exError: dataprocessor.ErrNilShardCoordinator,
		},
		{
			name: "NilPubKeyConverter",
			argsFunc: func() dataprocessor.RatingProcessorArgs {
				args := getRatingsProcessorArgs()
				args.ValidatorPubKeyConverter = nil
				return args
			},
			exError: dataprocessor.ErrNilPubKeyConverter,
		},
		{
			name: "NilMarshalizer",
			argsFunc: func() dataprocessor.RatingProcessorArgs {
				args := getRatingsProcessorArgs()
				args.Marshalizer = nil
				return args
			},
			exError: dataprocessor.ErrNilMarshalizer,
		},
		{
			name: "NilHasher",
			argsFunc: func() dataprocessor.RatingProcessorArgs {
				args := getRatingsProcessorArgs()
				args.Hasher = nil
				return args
			},
			exError: dataprocessor.ErrNilHasher,
		},
		{
			name: "NilElasticIndexer",
			argsFunc: func() dataprocessor.RatingProcessorArgs {
				args := getRatingsProcessorArgs()
				args.ElasticIndexer = nil
				return args
			},
			exError: dataprocessor.ErrNilElasticIndexer,
		},
		{
			name: "All arguments ok",
			argsFunc: func() dataprocessor.RatingProcessorArgs {
				return getRatingsProcessorArgs()
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dataprocessor.NewRatingsProcessor(tt.argsFunc())
			require.Equal(t, err, tt.exError)
		})
	}
}

func TestRatingsProcessor_IndexRatingsForEpochStartMetaBlock_NothingToIndexShouldNotErr(t *testing.T) {
	t.Parallel()

	indexWasCalled := false
	args := getRatingsProcessorArgs()
	args.ElasticIndexer = &mock.ElasticIndexerStub{
		SaveValidatorsRatingCalled: func(indexID string, infoRating []workItems.ValidatorRatingInfo) {
			indexWasCalled = true
		},
	}
	rp, _ := dataprocessor.NewRatingsProcessor(args)

	metaBlock := &block.MetaBlock{Epoch: 5}

	err := rp.IndexRatingsForEpochStartMetaBlock(metaBlock)
	require.NoError(t, err)
	require.False(t, indexWasCalled)
}

func TestRatingsProcessor_IndexRatingsForEpochStartMetaBlock_UnmarshalPeerErrorShouldNotFail(t *testing.T) {
	t.Parallel()

	indexWasCalled := false

	args := getRatingsProcessorArgs()
	args.ElasticIndexer = &mock.ElasticIndexerStub{
		SaveValidatorsRatingCalled: func(indexID string, infoRating []workItems.ValidatorRatingInfo) {
			indexWasCalled = true
		},
	}
	rp, _ := dataprocessor.NewRatingsProcessor(args)

	metaBlock := &block.MetaBlock{Epoch: 5}

	peerAdapter := &mock.AccountsStub{
		GetAllLeavesCalled: func(_ []byte) (map[string][]byte, error) {
			return map[string][]byte{"": []byte("not peer accounts bytes")}, nil
		},
	}

	rp.SetPeerAdapter(peerAdapter)

	err := rp.IndexRatingsForEpochStartMetaBlock(metaBlock)
	require.NoError(t, err)
	require.False(t, indexWasCalled)
}

func TestRatingsProcessor_IndexRatingsForGenesisMetaBlock_ShouldWork(t *testing.T) {
	t.Parallel()

	indexWasCalled := false

	args := getRatingsProcessorArgs()
	args.ElasticIndexer = &mock.ElasticIndexerStub{
		SaveValidatorsRatingCalled: func(indexID string, infoRating []workItems.ValidatorRatingInfo) {
			indexWasCalled = true
		},
	}
	args.GenesisNodesConfig = &mock.GenesisNodesSetupHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]sharding.GenesisNodeInfoHandler, map[uint32][]sharding.GenesisNodeInfoHandler) {
			nodeEligible := mock.NewNodeInfo([]byte("addr1"), []byte("pubKey1"), 0, 10)
			return map[uint32][]sharding.GenesisNodeInfoHandler{
				0: {nodeEligible},
			}, nil
		},
	}
	rp, _ := dataprocessor.NewRatingsProcessor(args)

	metaBlock := &block.MetaBlock{Nonce: 0, Epoch: 0}

	acc, _ := state.NewPeerAccount([]byte("peer account"))
	accBytes, _ := args.Marshalizer.Marshal(acc)
	peerAdapter := &mock.AccountsStub{
		GetAllLeavesCalled: func(_ []byte) (map[string][]byte, error) {
			return map[string][]byte{"": accBytes}, nil
		},
	}

	rp.SetPeerAdapter(peerAdapter)

	err := rp.IndexRatingsForEpochStartMetaBlock(metaBlock)
	require.NoError(t, err)
	require.True(t, indexWasCalled)
}

func TestRatingsProcessor_IndexRatingsForEpochStartMetaBlock_ShouldWork(t *testing.T) {
	t.Parallel()

	indexWasCalled := false

	args := getRatingsProcessorArgs()
	args.ElasticIndexer = &mock.ElasticIndexerStub{
		SaveValidatorsRatingCalled: func(indexID string, infoRating []workItems.ValidatorRatingInfo) {
			indexWasCalled = true
		},
	}
	rp, _ := dataprocessor.NewRatingsProcessor(args)

	metaBlock := &block.MetaBlock{Nonce: 5, Epoch: 5}

	acc, _ := state.NewPeerAccount([]byte("peer account"))
	accBytes, _ := args.Marshalizer.Marshal(acc)
	peerAdapter := &mock.AccountsStub{
		GetAllLeavesCalled: func(_ []byte) (map[string][]byte, error) {
			return map[string][]byte{"": accBytes}, nil
		},
	}

	rp.SetPeerAdapter(peerAdapter)

	err := rp.IndexRatingsForEpochStartMetaBlock(metaBlock)
	require.NoError(t, err)
	require.True(t, indexWasCalled)
}

func getRatingsProcessorArgs() dataprocessor.RatingProcessorArgs {
	return dataprocessor.RatingProcessorArgs{
		ValidatorPubKeyConverter: mock.NewPubkeyConverterMock(96),
		ShardCoordinator:         &mock.ShardCoordinatorMock{NumOfShards: 2},
		GenesisNodesConfig:       &mock.GenesisNodesSetupHandlerStub{},
		DbPathWithChainID:        "path",
		GeneralConfig: config.Config{
			EvictionWaitingList: config.EvictionWaitingListConfig{
				Size: 10,
				DB:   getDBConfig(),
			},
			TrieSnapshotDB: getDBConfig(),
			TrieStorageManagerConfig: config.TrieStorageManagerConfig{
				PruningBufferLen:   10,
				SnapshotsBufferLen: 10,
				MaxSnapshots:       10,
			},
			PeerAccountsTrieStorage: config.StorageConfig{
				Cache: getCacheConfig(),
				DB:    getDBConfig(),
				Bloom: config.BloomFilterConfig{},
			},
			StateTriesConfig: config.StateTriesConfig{
				CheckpointRoundsModulus:     10,
				AccountsStatePruningEnabled: false,
				PeerStatePruningEnabled:     false,
				MaxStateTrieLevelInMemory:   10,
				MaxPeerTrieLevelInMemory:    10,
			},
		},
		Marshalizer:    &mock.MarshalizerMock{},
		Hasher:         &mock.HasherMock{},
		ElasticIndexer: &mock.ElasticIndexerStub{},
	}
}

func getDBConfig() config.DBConfig {
	return config.DBConfig{
		FilePath:          "path",
		Type:              "MemoryDB",
		BatchDelaySeconds: 30,
		MaxBatchSize:      6,
		MaxOpenFiles:      10,
	}
}

func getCacheConfig() config.CacheConfig {
	return config.CacheConfig{
		Capacity: 10000,
		Type:     "LRU",
		Shards:   1,
	}
}
