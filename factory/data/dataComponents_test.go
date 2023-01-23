package data_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/errors"
	dataComp "github.com/multiversx/mx-chain-go/factory/data"
	"github.com/multiversx/mx-chain-go/factory/mock"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

func TestNewDataComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetDataArgs(coreComponents, shardCoordinator)
	args.ShardCoordinator = nil

	dcf, err := dataComp.NewDataComponentsFactory(args)
	require.Nil(t, dcf)
	require.Equal(t, errors.ErrNilShardCoordinator, err)
}

func TestNewDataComponentsFactory_NilCoreComponentsShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetDataArgs(nil, shardCoordinator)
	args.Core = nil

	dcf, err := dataComp.NewDataComponentsFactory(args)
	require.Nil(t, dcf)
	require.Equal(t, errors.ErrNilCoreComponents, err)
}

func TestNewDataComponentsFactory_NilEpochStartNotifierShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetDataArgs(coreComponents, shardCoordinator)
	args.EpochStartNotifier = nil

	dcf, err := dataComp.NewDataComponentsFactory(args)
	require.Nil(t, dcf)
	require.Equal(t, errors.ErrNilEpochStartNotifier, err)
}

func TestNewDataComponentsFactory_OkValsShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetDataArgs(coreComponents, shardCoordinator)
	dcf, err := dataComp.NewDataComponentsFactory(args)
	require.NoError(t, err)
	require.NotNil(t, dcf)
}

func TestDataComponentsFactory_CreateShouldErrDueBadConfig(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	coreComponents := componentsMock.GetCoreComponents()
	args := componentsMock.GetDataArgs(coreComponents, shardCoordinator)
	args.Config.ShardHdrNonceHashStorage = config.StorageConfig{}
	dcf, err := dataComp.NewDataComponentsFactory(args)
	require.NoError(t, err)

	dc, err := dcf.Create()
	require.Error(t, err)
	require.Nil(t, dc)
}

func TestDataComponentsFactory_CreateForShardShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetDataArgs(coreComponents, shardCoordinator)
	dcf, err := dataComp.NewDataComponentsFactory(args)

	require.NoError(t, err)
	dc, err := dcf.Create()
	require.NoError(t, err)
	require.NotNil(t, dc)
}

func TestDataComponentsFactory_CreateForMetaShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	shardCoordinator.CurrentShard = core.MetachainShardId
	args := componentsMock.GetDataArgs(coreComponents, shardCoordinator)

	dcf, err := dataComp.NewDataComponentsFactory(args)
	require.NoError(t, err)
	dc, err := dcf.Create()
	require.NoError(t, err)
	require.NotNil(t, dc)
}

// ------------ Test DataComponents --------------------
func TestManagedDataComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetDataArgs(coreComponents, shardCoordinator)
	dcf, _ := dataComp.NewDataComponentsFactory(args)

	dc, _ := dcf.Create()

	err := dc.Close()
	require.NoError(t, err)
}
