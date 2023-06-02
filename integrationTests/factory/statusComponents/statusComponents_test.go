package statusComponents

import (
	"fmt"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	bootstrapComp "github.com/multiversx/mx-chain-go/factory/bootstrap"
	"github.com/multiversx/mx-chain-go/integrationTests/factory"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/multiversx/mx-chain-go/testscommon/goroutines"
	"github.com/stretchr/testify/require"
)

// ------------ Test StatusComponents --------------------
func TestStatusComponents_Create_Close_ShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	time.Sleep(time.Second * 4)

	gc := goroutines.NewGoCounter(goroutines.TestsRelevantGoRoutines)
	idxInitial, _ := gc.Snapshot()
	factory.PrintStack()

	configs := factory.CreateDefaultConfig(t)
	chanStopNodeProcess := make(chan endProcess.ArgEndProcess)
	nr, err := node.NewNodeRunner(configs)
	require.Nil(t, err)

	managedCoreComponents, err := nr.CreateManagedCoreComponents(chanStopNodeProcess)
	require.Nil(t, err)
	managedStatusCoreComponents, err := nr.CreateManagedStatusCoreComponents(managedCoreComponents)
	require.Nil(t, err)
	managedCryptoComponents, err := nr.CreateManagedCryptoComponents(managedCoreComponents)
	require.Nil(t, err)
	managedNetworkComponents, err := nr.CreateManagedNetworkComponents(managedCoreComponents, managedStatusCoreComponents, managedCryptoComponents)
	require.Nil(t, err)
	managedBootstrapComponents, err := nr.CreateManagedBootstrapComponents(managedStatusCoreComponents, managedCoreComponents, managedCryptoComponents, managedNetworkComponents)
	require.Nil(t, err)
	managedDataComponents, err := nr.CreateManagedDataComponents(managedStatusCoreComponents, managedCoreComponents, managedBootstrapComponents, managedCryptoComponents)
	require.Nil(t, err)
	managedStateComponents, err := nr.CreateManagedStateComponents(managedCoreComponents, managedDataComponents, managedStatusCoreComponents)
	require.Nil(t, err)
	nodesShufflerOut, err := bootstrapComp.CreateNodesShuffleOut(managedCoreComponents.GenesisNodesSetup(), configs.GeneralConfig.EpochStartConfig, managedCoreComponents.ChanStopNodeProcess())
	require.Nil(t, err)
	storer, err := managedDataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit)
	require.Nil(t, err)
	nodesCoordinator, err := bootstrapComp.CreateNodesCoordinator(
		nodesShufflerOut,
		managedCoreComponents.GenesisNodesSetup(),
		configs.PreferencesConfig.Preferences,
		managedCoreComponents.EpochStartNotifierWithConfirm(),
		managedCryptoComponents.PublicKey(),
		managedCoreComponents.InternalMarshalizer(),
		managedCoreComponents.Hasher(),
		managedCoreComponents.Rater(),
		storer,
		managedCoreComponents.NodesShuffler(),
		managedBootstrapComponents.ShardCoordinator().SelfId(),
		managedBootstrapComponents.EpochBootstrapParams(),
		managedBootstrapComponents.EpochBootstrapParams().Epoch(),
		managedCoreComponents.ChanStopNodeProcess(),
		managedCoreComponents.NodeTypeProvider(),
		managedCoreComponents.EnableEpochsHandler(),
		managedDataComponents.Datapool().CurrentEpochValidatorInfo(),
		configs.GeneralConfig.EpochStartConfig.NumNodesConfigEpochsToStore,
	)
	require.Nil(t, err)
	managedStatusComponents, err := nr.CreateManagedStatusComponents(
		managedStatusCoreComponents,
		managedCoreComponents,
		managedNetworkComponents,
		managedBootstrapComponents,
		managedStateComponents,
		nodesCoordinator,
		false,
	)
	require.Nil(t, err)
	require.NotNil(t, managedStatusComponents)

	argsGasScheduleNotifier := forking.ArgsNewGasScheduleNotifier{
		GasScheduleConfig:  configs.EpochConfig.GasSchedule,
		ConfigDir:          configs.ConfigurationPathsHolder.GasScheduleDirectoryName,
		EpochNotifier:      managedCoreComponents.EpochNotifier(),
		WasmVMChangeLocker: managedCoreComponents.WasmVMChangeLocker(),
	}
	gasScheduleNotifier, err := forking.NewGasScheduleNotifier(argsGasScheduleNotifier)
	require.Nil(t, err)
	managedProcessComponents, err := nr.CreateManagedProcessComponents(
		managedCoreComponents,
		managedCryptoComponents,
		managedNetworkComponents,
		managedBootstrapComponents,
		managedStateComponents,
		managedDataComponents,
		managedStatusComponents,
		managedStatusCoreComponents,
		gasScheduleNotifier,
		nodesCoordinator,
	)
	require.Nil(t, err)
	time.Sleep(2 * time.Second)

	err = managedStatusComponents.SetForkDetector(managedProcessComponents.ForkDetector())
	require.Nil(t, err)
	err = managedStatusComponents.StartPolling()
	require.Nil(t, err)

	time.Sleep(5 * time.Second)

	err = managedStatusComponents.Close()
	require.Nil(t, err)
	err = managedProcessComponents.Close()
	require.Nil(t, err)
	err = managedStateComponents.Close()
	require.Nil(t, err)
	err = managedDataComponents.Close()
	require.Nil(t, err)
	err = managedBootstrapComponents.Close()
	require.Nil(t, err)
	err = managedNetworkComponents.Close()
	require.Nil(t, err)
	err = managedCryptoComponents.Close()
	require.Nil(t, err)
	err = managedStatusCoreComponents.Close()
	require.Nil(t, err)
	err = managedCoreComponents.Close()
	require.Nil(t, err)

	time.Sleep(5 * time.Second)

	idx, _ := gc.Snapshot()
	diff := gc.DiffGoRoutines(idxInitial, idx)
	require.Equal(t, 0, len(diff), fmt.Sprintf("%v", diff))
}
