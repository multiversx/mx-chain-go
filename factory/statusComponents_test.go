package factory_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatusComponentsFactory_NilCoreComponentsShouldErr(t *testing.T) {
	t.Parallel()

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	args.CoreComponents = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, errors.ErrNilCoreComponentsHolder, err)
}

func TestNewStatusComponentsFactory_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	args.NodesCoordinator = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, errors.ErrNilNodesCoordinator, err)
}

func TestNewStatusComponentsFactory_NilEpochStartNotifierShouldErr(t *testing.T) {
	t.Parallel()

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	args.EpochStartNotifier = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, errors.ErrNilEpochStartNotifier, err)
}

func TestNewStatusComponentsFactory_NilStatusHandlerErr(t *testing.T) {
	t.Parallel()

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	args.StatusUtils = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, errors.ErrNilStatusHandlersUtils, err)
}

func TestNewStatusComponentsFactory_NilNetworkComponentsShouldErr(t *testing.T) {
	t.Parallel()

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	args.NetworkComponents = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, errors.ErrNilNetworkComponentsHolder, err)
}

func TestNewStatusComponentsFactory_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	args.ShardCoordinator = nil
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, errors.ErrNilShardCoordinator, err)
}

func TestNewStatusComponentsFactory_InvalidRoundDurationShouldErr(t *testing.T) {
	t.Parallel()

	args, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	args.RoundDurationSec = 0
	scf, err := factory.NewStatusComponentsFactory(args)
	assert.True(t, check.IfNil(scf))
	assert.Equal(t, errors.ErrInvalidRoundDuration, err)
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
	statusArgs, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	coreComponents := getDefaultCoreComponents()
	statusArgs.CoreComponents = coreComponents

	statusComponentsFactory, _ := factory.NewStatusComponentsFactory(statusArgs)
	managedStatusComponents, err := factory.NewManagedStatusComponents(statusComponentsFactory)
	require.NoError(t, err)

	coreComponents.StatusHdl = nil
	err = managedStatusComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedStatusComponents.StatusHandler())
}

func TestManagedStatusComponents_Create_ShouldWork(t *testing.T) {
	statusArgs, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	statusComponentsFactory, _ := factory.NewStatusComponentsFactory(statusArgs)
	managedStatusComponents, err := factory.NewManagedStatusComponents(statusComponentsFactory)
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
	statusComponentsFactory, _ := factory.NewStatusComponentsFactory(statusArgs)
	managedStatusComponents, _ := factory.NewManagedStatusComponents(statusComponentsFactory)
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

func getStatusComponents(
	coreComponents factory.CoreComponentsHolder,
	networkComponents factory.NetworkComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	processComponents factory.ProcessComponentsHolder,
) factory.StatusComponentsHandler {
	statusArgs := factory.StatusComponentsFactoryArgs{
		Config:             testscommon.GetGeneralConfig(),
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
	}

	statusComponentsFactory, _ := factory.NewStatusComponentsFactory(statusArgs)
	managedStatusComponents, err := factory.NewManagedStatusComponents(statusComponentsFactory)
	if err != nil {
		fmt.Println("getStatusComponents NewManagedStatusComponents", "error", err.Error())
		return nil
	}
	err = managedStatusComponents.Create()
	if err != nil {
		fmt.Println("getStatusComponents Create", "error", err.Error())
	}
	return managedStatusComponents
}

func getStatusComponentsFactoryArgsAndProcessComponents() (factory.StatusComponentsFactoryArgs, factory.ProcessComponentsHolder) {
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
		Config:             testscommon.GetGeneralConfig(),
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

func getNetworkComponents() factory.NetworkComponentsHolder {
	networkArgs := getNetworkArgs()
	networkComponentsFactory, _ := factory.NewNetworkComponentsFactory(networkArgs)
	networkComponents, _ := factory.NewManagedNetworkComponents(networkComponentsFactory)

	_ = networkComponents.Create()

	return networkComponents
}

func getDataComponents(coreComponents factory.CoreComponentsHolder) factory.DataComponentsHolder {
	dataArgs := getDataArgs(coreComponents)
	dataComponentsFactory, _ := factory.NewDataComponentsFactory(dataArgs)
	dataComponents, _ := factory.NewManagedDataComponents(dataComponentsFactory)
	_ = dataComponents.Create()
	return dataComponents
}

func getCryptoComponents(coreComponents factory.CoreComponentsHolder) factory.CryptoComponentsHolder {
	cryptoArgs := getCryptoArgs(coreComponents)
	cryptoComponentsFactory, _ := factory.NewCryptoComponentsFactory(cryptoArgs)
	cryptoComponents, err := factory.NewManagedCryptoComponents(cryptoComponentsFactory)
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
	stateComponentsFactory, _ := factory.NewStateComponentsFactory(stateArgs)
	stateComponents, err := factory.NewManagedStateComponents(stateComponentsFactory)
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
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
	)
	processComponentsFactory, _ := factory.NewProcessComponentsFactory(processArgs)
	managedProcessComponents, err := factory.NewManagedProcessComponents(processComponentsFactory)
	if err != nil {
		fmt.Println("getProcessComponents NewManagedProcessComponents", "error", err.Error())
		return nil
	}
	err = managedProcessComponents.Create()
	if err != nil {
		fmt.Println("getProcessComponents Create", "error", err.Error())
	}
	return managedProcessComponents
}
