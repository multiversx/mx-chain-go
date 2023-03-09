package status_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/errors"
	coreComp "github.com/multiversx/mx-chain-go/factory/core"
	"github.com/multiversx/mx-chain-go/factory/mock"
	statusComp "github.com/multiversx/mx-chain-go/factory/status"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatusComponentsFactory_NilCoreComponentsShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args, _ := componentsMock.GetStatusComponentsFactoryArgsAndProcessComponents(shardCoordinator)
	args.CoreComponents = nil
	scf, err := statusComp.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, errors.ErrNilCoreComponentsHolder, err)
}

func TestNewStatusComponentsFactory_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args, _ := componentsMock.GetStatusComponentsFactoryArgsAndProcessComponents(shardCoordinator)
	args.NodesCoordinator = nil
	scf, err := statusComp.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, errors.ErrNilNodesCoordinator, err)
}

func TestNewStatusComponentsFactory_NilEpochStartNotifierShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args, _ := componentsMock.GetStatusComponentsFactoryArgsAndProcessComponents(shardCoordinator)
	args.EpochStartNotifier = nil
	scf, err := statusComp.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, errors.ErrNilEpochStartNotifier, err)
}

func TestNewStatusComponentsFactory_NilNetworkComponentsShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args, _ := componentsMock.GetStatusComponentsFactoryArgsAndProcessComponents(shardCoordinator)
	args.NetworkComponents = nil
	scf, err := statusComp.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, errors.ErrNilNetworkComponentsHolder, err)
}

func TestNewStatusComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args, _ := componentsMock.GetStatusComponentsFactoryArgsAndProcessComponents(shardCoordinator)
	args.ShardCoordinator = nil
	scf, err := statusComp.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, errors.ErrNilShardCoordinator, err)
}

func TestNewStatusComponents_InvalidRoundDurationShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	coreArgs := componentsMock.GetCoreArgs()
	coreArgs.NodesFilename = "../mock/testdata/nodesSetupMockInvalidRound.json"
	coreComponentsFactory, _ := coreComp.NewCoreComponentsFactory(coreArgs)
	coreComponents, err := coreComp.NewManagedCoreComponents(coreComponentsFactory)
	require.Nil(t, err)
	require.NotNil(t, coreComponents)
	err = coreComponents.Create()
	require.Nil(t, err)
	networkComponents := componentsMock.GetNetworkComponents(componentsMock.GetCryptoComponents(coreComponents))
	dataComponents := componentsMock.GetDataComponents(coreComponents, shardCoordinator)
	stateComponents := componentsMock.GetStateComponents(coreComponents, shardCoordinator)

	statusArgs := statusComp.StatusComponentsFactoryArgs{
		Config:               testscommon.GetGeneralConfig(),
		ExternalConfig:       config.ExternalConfig{},
		ShardCoordinator:     shardCoordinator,
		NodesCoordinator:     &shardingMocks.NodesCoordinatorMock{},
		EpochStartNotifier:   &mock.EpochStartNotifierStub{},
		CoreComponents:       coreComponents,
		DataComponents:       dataComponents,
		NetworkComponents:    networkComponents,
		StateComponents:      stateComponents,
		IsInImportMode:       false,
		EconomicsConfig:      config.EconomicsConfig{},
		StatusCoreComponents: componentsMock.GetStatusCoreComponents(),
	}
	scf, err := statusComp.NewStatusComponentsFactory(statusArgs)
	assert.Nil(t, err)
	assert.NotNil(t, scf)

	statusComponents, err := scf.Create()
	assert.Nil(t, statusComponents)
	assert.Equal(t, errors.ErrInvalidRoundDuration, err)
}

func TestNewStatusComponentsFactory_ShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args, _ := componentsMock.GetStatusComponentsFactoryArgsAndProcessComponents(shardCoordinator)
	scf, err := statusComp.NewStatusComponentsFactory(args)
	require.NoError(t, err)
	require.False(t, check.IfNil(scf))
}

func TestStatusComponentsFactory_Create(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args, _ := componentsMock.GetStatusComponentsFactoryArgsAndProcessComponents(shardCoordinator)
	scf, err := statusComp.NewStatusComponentsFactory(args)
	require.Nil(t, err)

	res, err := scf.Create()
	require.NoError(t, err)
	require.NotNil(t, res)
}

// ------------ Test StatusComponents --------------------
func TestStatusComponents_CloseShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	statusArgs, _ := componentsMock.GetStatusComponentsFactoryArgsAndProcessComponents(shardCoordinator)
	scf, _ := statusComp.NewStatusComponentsFactory(statusArgs)
	cc, err := scf.Create()
	require.Nil(t, err)

	err = cc.Close()
	require.NoError(t, err)
}
