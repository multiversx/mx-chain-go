package vm_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	runType "github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/vm"
	processFactory "github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
)

func createRunTypeComponents() runType.RunTypeComponentsHolder {
	coreComp := componentsMock.GetCoreComponents(testscommon.GetGeneralConfig())
	cryptoComp := componentsMock.GetCryptoComponents(coreComp)

	return componentsMock.GetRunTypeComponents(coreComp, cryptoComp)
}

func TestNewSovereignVmContainerShardCreatorFactory(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		runTypeComponents := createRunTypeComponents()
		sovereignVmContainerShardFactory, err := vm.NewSovereignVmContainerShardFactory(runTypeComponents.VmContainerMetaFactoryCreator(), runTypeComponents.VmContainerShardFactoryCreator())
		require.Nil(t, err)
		require.False(t, sovereignVmContainerShardFactory.IsInterfaceNil())
	})

	t.Run("nil vmContainerMetaFactory", func(t *testing.T) {
		t.Parallel()

		runTypeComponents := createRunTypeComponents()
		sovereignVmContainerShardFactory, err := vm.NewSovereignVmContainerShardFactory(nil, runTypeComponents.VmContainerShardFactoryCreator())
		require.ErrorIs(t, err, vm.ErrNilVmContainerMetaCreator)
		require.True(t, sovereignVmContainerShardFactory.IsInterfaceNil())
	})

	t.Run("nil vmContainerShardFactory", func(t *testing.T) {
		t.Parallel()

		runTypeComponents := createRunTypeComponents()
		sovereignVmContainerShardFactory, err := vm.NewSovereignVmContainerShardFactory(runTypeComponents.VmContainerMetaFactoryCreator(), nil)
		require.ErrorIs(t, err, vm.ErrNilVmContainerShardCreator)
		require.True(t, sovereignVmContainerShardFactory.IsInterfaceNil())
	})
}

func TestNewSovereignVmContainerShardFactory_CreateVmContainerFactoryShard(t *testing.T) {
	t.Parallel()
	if runtime.GOARCH == "arm64" {
		t.Skip("skipping test on arm64")
	}

	runTypeComponents := createRunTypeComponents()
	sovereignVmContainerShardFactory, err := vm.NewSovereignVmContainerShardFactory(runTypeComponents.VmContainerMetaFactoryCreator(), runTypeComponents.VmContainerShardFactoryCreator())
	require.Nil(t, err)
	require.False(t, sovereignVmContainerShardFactory.IsInterfaceNil())

	argsShard := createMockVMAccountsArguments()
	gasSchedule := makeGasSchedule()
	argsMeta := createVmContainerMockArgument(gasSchedule)
	args := vm.ArgsVmContainerFactory{
		BlockChainHook:      argsShard.BlockChainHook,
		Config:              argsShard.Config,
		BlockGasLimit:       argsShard.BlockGasLimit,
		GasSchedule:         argsMeta.GasSchedule,
		EpochNotifier:       argsShard.EpochNotifier,
		EnableEpochsHandler: argsShard.EnableEpochsHandler,
		WasmVMChangeLocker:  argsShard.WasmVMChangeLocker,
		ESDTTransferParser:  argsShard.ESDTTransferParser,
		BuiltInFunctions:    argsShard.BuiltInFunctions,
		Hasher:              argsShard.Hasher,
		Economics:           argsMeta.Economics,
		MessageSignVerifier: argsMeta.MessageSignVerifier,
		NodesConfigProvider: argsMeta.NodesConfigProvider,
		Marshalizer:         argsMeta.Marshalizer,
		SystemSCConfig:      argsMeta.SystemSCConfig,
		ValidatorAccountsDB: argsMeta.ValidatorAccountsDB,
		UserAccountsDB:      argsMeta.UserAccountsDB,
		ChanceComputer:      argsMeta.ChanceComputer,
		ShardCoordinator:    argsMeta.ShardCoordinator,
		PubkeyConv:          argsMeta.PubkeyConv,
		NodesCoordinator:    argsMeta.NodesCoordinator,
	}

	vmContainer, vmFactory, err := sovereignVmContainerShardFactory.CreateVmContainerFactory(args)
	require.Nil(t, err)
	require.Equal(t, "*containers.virtualMachinesContainer", fmt.Sprintf("%T", vmContainer))
	require.Equal(t, "*shard.vmContainerFactory", fmt.Sprintf("%T", vmFactory))

	require.Equal(t, 2, vmContainer.Len())
	svm, err := vmContainer.Get(processFactory.SystemVirtualMachine)
	require.Nil(t, err)
	require.NotNil(t, svm)
	require.Equal(t, "*process.systemVM", fmt.Sprintf("%T", svm))

	wasmvm, err := vmContainer.Get(processFactory.WasmVirtualMachine)
	require.Nil(t, err)
	require.NotNil(t, wasmvm)
	require.Equal(t, "*hostCore.vmHost", fmt.Sprintf("%T", wasmvm))
}
