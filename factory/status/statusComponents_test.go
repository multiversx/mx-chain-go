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
	coreArgs.Config.GeneralSettings.ChainParametersByEpoch[0].RoundDuration = 500
	coreArgs.NodesConfig = config.NodesConfig{
		StartTime: 0,
		InitialNodes: []*config.InitialNodeConfig{
			{
				PubKey:  "227a5a5ec0c58171b7f4ee9ecc304ea7b176fb626741a25c967add76d6cd361d6995929f9b60a96237381091cefb1b061225e5bb930b40494a5ac9d7524fd67dfe478e5ccd80f17b093cff5722025761fb0217c39dbd5ae45e01eb5a3113be93",
				Address: "erd1ulhw20j7jvgfgak5p05kv667k5k9f320sgef5ayxkt9784ql0zssrzyhjp",
			},
			{
				PubKey:  "ef9522d654bc08ebf2725468f41a693aa7f3cf1cb93922cff1c8c81fba78274016010916f4a7e5b0855c430a724a2d0b3acd1fe8e61e37273a17d58faa8c0d3ef6b883a33ec648950469a1e9757b978d9ae662a019068a401cff56eea059fd08",
				Address: "erd17c4fs6mz2aa2hcvva2jfxdsrdknu4220496jmswer9njznt22eds0rxlr4",
			},
			{
				PubKey:  "e91ab494cedd4da346f47aaa1a3e792bea24fb9f6cc40d3546bc4ca36749b8bfb0164e40dbad2195a76ee0fd7fb7da075ecbf1b35a2ac20638d53ea5520644f8c16952225c48304bb202867e2d71d396bff5a5971f345bcfe32c7b6b0ca34c84",
				Address: "erd10d2gufxesrp8g409tzxljlaefhs0rsgjle3l7nq38de59txxt8csj54cd3",
			},
		},
	}
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
