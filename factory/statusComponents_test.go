package factory_test

import (
	"fmt"
	"sync"
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

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	args.CoreComponents = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, factory.ErrNilCoreComponentsHolder, err)
}

func TestNewStatusComponentsFactory_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	args.NodesCoordinator = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, factory.ErrNilNodesCoordinator, err)
}

func TestNewStatusComponentsFactory_NilEpochStartNotifierShouldErr(t *testing.T) {
	t.Parallel()

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	args.EpochStartNotifier = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, factory.ErrNilEpochStartNotifier, err)
}

func TestNewStatusComponentsFactory_NilStatusHandlerErr(t *testing.T) {
	t.Parallel()

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	args.StatusUtils = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, factory.ErrNilStatusHandlersUtils, err)
}

func TestNewStatusComponentsFactory_NilNetworkComponentsShouldErr(t *testing.T) {
	t.Parallel()

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	args.NetworkComponents = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, factory.ErrNilNetworkComponentsHolder, err)
}

func TestNewStatusComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	args.ShardCoordinator = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, factory.ErrNilShardCoordinator, err)
}

func TestNewStatusComponentsFactory_InvalidRoundDurationShouldErr(t *testing.T) {
	t.Parallel()

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	args.RoundDurationSec = 0
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, factory.ErrInvalidRoundDuration, err)
}

func TestNewStatusComponentsFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	scf, err := factory.NewStatusComponentsFactory(args)
	require.NoError(t, err)
	require.False(t, check.IfNil(scf))
}

func TestStatusComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	scf, err := factory.NewStatusComponentsFactory(args)
	require.Nil(t, err)

	res, err := scf.Create()
	require.NoError(t, err)
	require.NotNil(t, res)
}

// ------------ Test ManagedStatusComponents --------------------
func TestManagedStatusComponents_CreateWithInvalidArgs_ShouldErr(t *testing.T) {
	t.Skip("Should be fixed")
	statusArgs, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	managedStatusComponents, err := factory.NewManagedStatusComponents(statusArgs)
	require.NoError(t, err)
	err = managedStatusComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedStatusComponents.StatusHandler())
}

func TestManagedStatusComponents_Create_ShouldWork(t *testing.T) {
	statusArgs, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	managedStatusComponents, err := factory.NewManagedStatusComponents(statusArgs)
	require.NoError(t, err)
	require.Nil(t, managedStatusComponents.StatusHandler())
	require.Nil(t, managedStatusComponents.ElasticIndexer())
	require.Nil(t, managedStatusComponents.SoftwareVersionChecker())
	require.Nil(t, managedStatusComponents.TpsBenchmark())

	err = managedStatusComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedStatusComponents.StatusHandler())
	require.NotNil(t, managedStatusComponents.ElasticIndexer())
	require.NotNil(t, managedStatusComponents.SoftwareVersionChecker())
	require.NotNil(t, managedStatusComponents.TpsBenchmark())
}

func TestManagedStatusComponents_Close(t *testing.T) {
	statusArgs, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	managedStatusComponents, _ := factory.NewManagedStatusComponents(statusArgs)
	err := managedStatusComponents.Create()
	require.NoError(t, err)

	err = managedStatusComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedStatusComponents.StatusHandler())
}

// ------------ Test StatusComponents --------------------
func TestStatusComponents_Close_ShouldWork(t *testing.T) {
	statusArgs, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	scf, _ := factory.NewStatusComponentsFactory(statusArgs)
	cc, err := scf.Create()
	require.Nil(t, err)

	err = cc.Close()
	require.NoError(t, err)
}

func getStatusComponentsFactoryArgsAndProcessComponents() (factory.StatusComponentsFactoryArgs, factory.ProcessComponentsHolder) {
	coreArgs := getCoreArgs()
	coreComponents := getCoreComponents()
	networkComponents := getNetworkComponents()
	dataComponents := getDataComponents(coreComponents)
	cryptoComponents := getCryptoComponents(coreComponents)
	stateComponents := getStateComponents(coreComponents)
	processComponents := getProcessComponents(
		coreComponents,
		networkComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
	)

	return factory.StatusComponentsFactoryArgs{
		Config:             coreArgs.Config,
		ExternalConfig:     config.ExternalConfig{},
		RoundDurationSec:   4,
		ElasticOptions:     &indexer.Options{},
		ShardCoordinator:   mock.NewMultiShardsCoordinatorMock(2),
		NodesCoordinator:   processComponents.NodesCoordinator(),
		EpochStartNotifier: processComponents.EpochStartNotifier(),
		CoreComponents:     coreComponents,
		DataComponents:     dataComponents,
		NetworkComponents:  networkComponents,
		StatusUtils:        &mock.StatusHandlersUtilsMock{},
	}, processComponents
}

// creating network components in parallel is not concurrent safe as it changes global variable from
// pubsub. This is is not an issue during normal operations as this is not called concurrently,
// but in unit tests, this might get called in parallel so a mutex should be used.
var mutNetworkComponentsCreate = sync.Mutex{}

func getNetworkComponents() factory.NetworkComponentsHolder {
	networkArgs := getNetworkArgs()
	networkComponents, _ := factory.NewManagedNetworkComponents(networkArgs)

	mutNetworkComponentsCreate.Lock()
	_ = networkComponents.Create()
	mutNetworkComponentsCreate.Unlock()

	return networkComponents
}

func getDataComponents(coreComponents factory.CoreComponentsHolder) factory.DataComponentsHolder {
	dataArgs := getDataArgs(coreComponents)
	dataComponents, _ := factory.NewManagedDataComponents(factory.DataComponentsHandlerArgs(dataArgs))
	_ = dataComponents.Create()
	return dataComponents
}

func getCryptoComponents(coreComponents factory.CoreComponentsHolder) factory.CryptoComponentsHolder {
	cryptoArgs := getCryptoArgs(coreComponents)
	cryptoComponents, err := factory.NewManagedCryptoComponents(factory.CryptoComponentsHandlerArgs(cryptoArgs))
	if err != nil {
		fmt.Println("getCryptoComponents NewManagedCryptoComponents", "error", err.Error())
		return nil
	}

	err = cryptoComponents.Create()
	if err != nil {
		fmt.Println("getCryptoComponents Create", "error", err.Error())
		return nil
	}
	return cryptoComponents
}

func getStateComponents(coreComponents factory.CoreComponentsHolder) factory.StateComponentsHolder {
	stateArgs := getStateArgs(coreComponents)
	stateComponents, err := factory.NewManagedStateComponents(stateArgs)
	if err != nil {
		fmt.Println("getStateComponents NewManagedStateComponents", "error", err.Error())
		return nil
	}
	err = stateComponents.Create()
	if err != nil {
		fmt.Println("getStateComponents Create", "error", err.Error())
	}
	return stateComponents
}

func getProcessComponents(
	coreComponents factory.CoreComponentsHolder,
	networkComponents factory.NetworkComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	cryptoComponents factory.CryptoComponentsHolder,
	stateComponents factory.StateComponentsHolder,
) factory.ProcessComponentsHolder {
	processArgs := getProcessArgs(
		getCoreArgs(),
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
	)
	processComponents, err := factory.NewManagedProcessComponents(processArgs)
	if err != nil {
		fmt.Println("getProcessComponents NewManagedProcessComponents", "error", err.Error())
		return nil
	}
	err = processComponents.Create()
	if err != nil {
		fmt.Println("getProcessComponents Create", "error", err.Error())
	}
	return processComponents
}
