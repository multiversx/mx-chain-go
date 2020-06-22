package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/stretchr/testify/require"
)

func TestNewDataComponentsFactory_NilEconomicsDataShouldErr(t *testing.T) {
	t.Parallel()

	args := getDataArgs()
	args.EconomicsData = nil

	dcf, err := factory.NewDataComponentsFactory(args)
	require.Nil(t, dcf)
	require.Equal(t, factory.ErrNilEconomicsData, err)
}

func TestNewDataComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getDataArgs()
	args.ShardCoordinator = nil

	dcf, err := factory.NewDataComponentsFactory(args)
	require.Nil(t, dcf)
	require.Equal(t, factory.ErrNilShardCoordinator, err)
}

func TestNewDataComponentsFactory_NilCoreComponentsShouldErr(t *testing.T) {
	t.Parallel()

	args := getDataArgs()
	args.Core = nil

	dcf, err := factory.NewDataComponentsFactory(args)
	require.Nil(t, dcf)
	require.Equal(t, factory.ErrNilCoreComponents, err)
}

func TestNewDataComponentsFactory_NilEpochStartNotifierShouldErr(t *testing.T) {
	t.Parallel()

	args := getDataArgs()
	args.EpochStartNotifier = nil

	dcf, err := factory.NewDataComponentsFactory(args)
	require.Nil(t, dcf)
	require.Equal(t, factory.ErrNilEpochStartNotifier, err)
}

func TestNewDataComponentsFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	args := getDataArgs()

	dcf, err := factory.NewDataComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, dcf)
}

func TestDataComponentsFactory_CreateShouldErrDueBadConfig(t *testing.T) {
	t.Parallel()

	args := getDataArgs()
	args.Config.ShardHdrNonceHashStorage = config.StorageConfig{}
	dcf, err := factory.NewDataComponentsFactory(args)
	require.NoError(t, err)

	dc, err := dcf.Create()
	require.Error(t, err)
	require.Nil(t, dc)
}

func TestDataComponentsFactory_CreateForShardShouldWork(t *testing.T) {
	t.Parallel()

	args := getDataArgs()
	dcf, err := factory.NewDataComponentsFactory(args)

	require.NoError(t, err)
	dc, err := dcf.Create()
	require.NoError(t, err)
	require.NotNil(t, dc)
}

func TestDataComponentsFactory_CreateForMetaShouldWork(t *testing.T) {
	t.Parallel()

	args := getDataArgs()
	multiShrdCoord := mock.NewMultiShardsCoordinatorMock(3)
	multiShrdCoord.CurrentShard = core.MetachainShardId
	args.ShardCoordinator = multiShrdCoord
	dcf, err := factory.NewDataComponentsFactory(args)
	require.NoError(t, err)
	dc, err := dcf.Create()
	require.NoError(t, err)
	require.NotNil(t, dc)
}

func getDataArgs() factory.DataComponentsFactoryArgs {
	coreArgs := getCoreArgs()
	coreComponents, _ := factory.NewManagedCoreComponents(factory.CoreComponentsHandlerArgs(coreArgs))
	_ = coreComponents.Create()

	testEconomics := &economics.TestEconomicsData{EconomicsData: &economics.EconomicsData{}}
	testEconomics.SetMinGasPrice(200000000000)

	return factory.DataComponentsFactoryArgs{
		Config:             getGeneralConfig(),
		EconomicsData:      testEconomics.EconomicsData,
		ShardCoordinator:   mock.NewMultiShardsCoordinatorMock(2),
		Core:               coreComponents,
		EpochStartNotifier: &mock.EpochStartNotifierStub{},
		CurrentEpoch:       0,
	}
}

func getGeneralConfig() config.Config {
	storageCfg := config.StorageConfig{
		Cache: getCacheCfg(),
		DB:    getDBCfg(),
		Bloom: config.BloomFilterConfig{},
	}
	cacheCfg := getCacheCfg()
	return config.Config{
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
