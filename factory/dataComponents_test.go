package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewDataComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	coreComponents := getCoreComponents()
	args := getDataArgs(coreComponents, shardCoordinator)
	args.ShardCoordinator = nil

	dcf, err := factory.NewDataComponentsFactory(args)
	require.Nil(t, dcf)
	require.Equal(t, errors.ErrNilShardCoordinator, err)
}

func TestNewDataComponentsFactory_NilCoreComponentsShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := getDataArgs(nil, shardCoordinator)
	args.Core = nil

	dcf, err := factory.NewDataComponentsFactory(args)
	require.Nil(t, dcf)
	require.Equal(t, errors.ErrNilCoreComponents, err)
}

func TestNewDataComponentsFactory_NilEpochStartNotifierShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	coreComponents := getCoreComponents()
	args := getDataArgs(coreComponents, shardCoordinator)
	args.EpochStartNotifier = nil

	dcf, err := factory.NewDataComponentsFactory(args)
	require.Nil(t, dcf)
	require.Equal(t, errors.ErrNilEpochStartNotifier, err)
}

func TestNewDataComponentsFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	coreComponents := getCoreComponents()
	args := getDataArgs(coreComponents, shardCoordinator)
	dcf, err := factory.NewDataComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, dcf)
}

func TestDataComponentsFactory_CreateShouldErrDueBadConfig(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	coreComponents := getCoreComponents()
	args := getDataArgs(coreComponents, shardCoordinator)
	args.Config.ShardHdrNonceHashStorage = config.StorageConfig{}
	dcf, err := factory.NewDataComponentsFactory(args)
	require.NoError(t, err)

	dc, err := dcf.Create()
	require.Error(t, err)
	require.Nil(t, dc)
}

func TestDataComponentsFactory_CreateForShardShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := getDataArgs(coreComponents, shardCoordinator)
	dcf, err := factory.NewDataComponentsFactory(args)

	require.NoError(t, err)
	dc, err := dcf.Create()
	require.NoError(t, err)
	require.NotNil(t, dc)
}

func TestDataComponentsFactory_CreateForMetaShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.CurrentShard = core.MetachainShardId
	args := getDataArgs(coreComponents, shardCoordinator)

	dcf, err := factory.NewDataComponentsFactory(args)
	require.NoError(t, err)
	dc, err := dcf.Create()
	require.NoError(t, err)
	require.NotNil(t, dc)
}

// ------------ Test DataComponents --------------------
func TestManagedDataComponents_Close_ShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := getDataArgs(coreComponents, shardCoordinator)
	dcf, _ := factory.NewDataComponentsFactory(args)

	dc, _ := dcf.Create()

	err := dc.Close()
	require.NoError(t, err)
}

func getDataArgs(coreComponents factory.CoreComponentsHolder, shardCoordinator sharding.Coordinator) factory.DataComponentsFactoryArgs {
	testEconomics := &economics.TestEconomicsData{EconomicsData: &economics.EconomicsData{}}
	testEconomics.SetMinGasPrice(200000000000)

	return factory.DataComponentsFactoryArgs{
		Config:                        testscommon.GetGeneralConfig(),
		ShardCoordinator:              shardCoordinator,
		Core:                          coreComponents,
		EpochStartNotifier:            &mock.EpochStartNotifierStub{},
		CurrentEpoch:                  0,
		CreateTrieEpochRootHashStorer: false,
	}
}
