package factory

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewDataPoolFromConfig(t *testing.T) {
	args := getGoodArgs()
	holder, err := NewDataPoolFromConfig(args)
	require.Nil(t, err)
	require.NotNil(t, holder)
}

func TestNewDataPoolFromConfig_MissingDependencyShouldErr(t *testing.T) {
	args := getGoodArgs()
	args.Config = nil
	holder, err := NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.Equal(t, dataRetriever.ErrNilConfig, err)

	args = getGoodArgs()
	args.EconomicsData = nil
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.Equal(t, dataRetriever.ErrNilEconomicsData, err)

	args = getGoodArgs()
	args.ShardCoordinator = nil
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.Equal(t, dataRetriever.ErrNilShardCoordinator, err)

	args = getGoodArgs()
	args.HealthService = nil
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.Equal(t, dataRetriever.ErrNilHealthService, err)
}

func TestNewDataPoolFromConfig_BadConfigShouldErr(t *testing.T) {
	// We test one (arbitrary and trivial) erroneous config for each component that needs to be created

	args := getGoodArgs()
	args.Config.TxDataPool.Capacity = 0
	holder, err := NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.NotNil(t, err)

	args = getGoodArgs()
	args.Config.UnsignedTransactionDataPool.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.NotNil(t, err)

	args = getGoodArgs()
	args.Config.RewardTransactionDataPool.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	require.NotNil(t, err)

	args = getGoodArgs()
	args.Config.HeadersPoolConfig.MaxHeadersPerShard = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.NotNil(t, err)

	args = getGoodArgs()
	args.Config.TxBlockBodyDataPool.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.NotNil(t, err)

	args = getGoodArgs()
	args.Config.PeerBlockBodyDataPool.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.NotNil(t, err)

	args = getGoodArgs()
	args.Config.TrieNodesDataPool.Capacity = 0
	holder, err = NewDataPoolFromConfig(args)
	require.Nil(t, holder)
	fmt.Println(err)
	require.NotNil(t, err)
}

func getGoodArgs() ArgsDataPool {
	testEconomics := &economics.TestEconomicsData{EconomicsData: &economics.EconomicsData{}}
	testEconomics.SetMinGasPrice(200000000000)

	return ArgsDataPool{
		Config:           getGeneralConfig(),
		EconomicsData:    testEconomics.EconomicsData,
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(2),
		HealthService:    testscommon.NewHealthServiceStub(),
	}
}

// move to tests common
func getGeneralConfig() *config.Config {
	storageCfg := config.StorageConfig{
		Cache: getCacheCfg(),
		DB:    getDBCfg(),
		Bloom: config.BloomFilterConfig{},
	}
	cacheCfg := getCacheCfg()
	return &config.Config{
		StoragePruning: config.StoragePruningConfig{
			Enabled:             false,
			FullArchive:         true,
			NumEpochsToKeep:     3,
			NumActivePersisters: 3,
		},
		TxDataPool: config.CacheConfig{
			Capacity:             10000,
			SizePerSender:        1000,
			SizeInBytes:          1000000000,
			SizeInBytesPerSender: 10000000,
			Shards:               1,
		},
		UnsignedTransactionDataPool: config.CacheConfig{
			Capacity:    10000,
			SizeInBytes: 1000000000,
			Shards:      1,
		},
		RewardTransactionDataPool: config.CacheConfig{
			Capacity:    10000,
			SizeInBytes: 1000000000,
			Shards:      1,
		},
		HeadersPoolConfig: config.HeadersPoolConfig{
			MaxHeadersPerShard:            100,
			NumElementsToRemoveOnEviction: 1,
		},
		TxBlockBodyDataPool:        cacheCfg,
		PeerBlockBodyDataPool:      cacheCfg,
		TrieNodesDataPool:          cacheCfg,
		TxStorage:                  storageCfg,
		MiniBlocksStorage:          storageCfg,
		ShardHdrNonceHashStorage:   storageCfg,
		MetaBlockStorage:           storageCfg,
		MetaHdrNonceHashStorage:    storageCfg,
		UnsignedTransactionStorage: storageCfg,
		RewardTxStorage:            storageCfg,
		BlockHeaderStorage:         storageCfg,
		Heartbeat: config.HeartbeatConfig{
			HeartbeatStorage: storageCfg,
		},
		StatusMetricsStorage: storageCfg,
		PeerBlockBodyStorage: storageCfg,
		BootstrapStorage:     storageCfg,
		TxLogsStorage:        storageCfg,
	}
}

func getCacheCfg() config.CacheConfig {
	return config.CacheConfig{
		Type:     "LRU",
		Capacity: 10,
		Shards:   1,
	}
}

func getDBCfg() config.DBConfig {
	return config.DBConfig{
		FilePath:          "",
		Type:              string(storageUnit.MemoryDB),
		BatchDelaySeconds: 10,
		MaxBatchSize:      10,
		MaxOpenFiles:      10,
	}
}
