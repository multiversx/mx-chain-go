package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatusComponentsFactory_NilCoreComponentsShouldErr(t *testing.T) {
	t.Parallel()

	args := getStatusComponentsFactoryArgs()
	args.CoreComponents = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, factory.ErrNilCoreComponentsHolder, err)
}

func TestNewStatusComponentsFactory_NilProcessComponentsShouldErr(t *testing.T) {
	t.Parallel()

	args := getStatusComponentsFactoryArgs()
	args.ProcessComponents = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, factory.ErrNilProcessComponentsHolder, err)
}

func TestNewStatusComponentsFactory_NilNetworkComponentsShouldErr(t *testing.T) {
	t.Parallel()

	args := getStatusComponentsFactoryArgs()
	args.NetworkComponents = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, factory.ErrNilNetworkComponentsHolder, err)
}

func TestNewStatusComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := getStatusComponentsFactoryArgs()
	args.ShardCoordinator = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, factory.ErrNilShardCoordinator, err)
}

func TestNewStatusComponentsFactory_InvalidRoundDurationShouldErr(t *testing.T) {
	t.Parallel()

	args := getStatusComponentsFactoryArgs()
	args.RoundDurationSec = 0
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, factory.ErrInvalidRoundDuration, err)
}

func TestNewStatusComponentsFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	scf, err := factory.NewStatusComponentsFactory(getStatusComponentsFactoryArgs())
	require.NoError(t, err)
	require.False(t, check.IfNil(scf))
}

func TestStatusComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	args := getStatusComponentsFactoryArgs()
	scf, err := factory.NewStatusComponentsFactory(args)
	require.Nil(t, err)

	res, err := scf.Create()
	require.NoError(t, err)
	require.NotNil(t, res)
}

func getStatusComponentsFactoryArgs() factory.StatusComponentsFactoryArgs {
	return factory.StatusComponentsFactoryArgs{
		Config:            config.Config{},
		ExternalConfig:    config.ExternalConfig{},
		RoundDurationSec:  4,
		ElasticOptions:    &indexer.Options{},
		ShardCoordinator:  mock.NewMultiShardsCoordinatorMock(2),
		CoreComponents:    getCoreComponents(),
		NetworkComponents: getNetworkComponents(),
		ProcessComponents: getProcessComponents(),
	}
}

func getNetworkComponents() factory.NetworkComponentsHolder {
	networkArgs := getNetworkArgs()
	networkComponents, _ := factory.NewManagedNetworkComponents(networkArgs)
	_ = networkComponents.Create()
	return networkComponents
}

func getProcessComponents() factory.ProcessComponentsHolder {
	processArgs := getProcessArgs()
	processComponents, _ := factory.NewManagedProcessComponents(processArgs)
	_ = processComponents.Create()
	return processComponents
}
