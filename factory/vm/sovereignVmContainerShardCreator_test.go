package vm_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/process"
	factory2 "github.com/multiversx/mx-chain-go/process/factory"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignVmContainerShardCreatorFactory(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		bhhc := &factory.BlockChainHookHandlerFactoryMock{}
		runTypeComponents := componentsMock.GetRunTypeComponents()
		sovereignVmContainerShardFactory, err := vm.NewSovereignVmContainerShardFactory(bhhc, runTypeComponents.VmContainerMetaFactoryCreator(), runTypeComponents.VmContainerShardFactoryCreator())
		require.Nil(t, err)
		require.False(t, sovereignVmContainerShardFactory.IsInterfaceNil())
	})

	t.Run("nil blockChainHookHandlerCreator", func(t *testing.T) {
		t.Parallel()

		runTypeComponents := componentsMock.GetRunTypeComponents()
		sovereignVmContainerShardFactory, err := vm.NewSovereignVmContainerShardFactory(nil, runTypeComponents.VmContainerMetaFactoryCreator(), runTypeComponents.VmContainerShardFactoryCreator())
		require.ErrorIs(t, err, process.ErrNilBlockChainHook)
		require.True(t, sovereignVmContainerShardFactory.IsInterfaceNil())
	})

	t.Run("nil vmContainerMetaFactory", func(t *testing.T) {
		t.Parallel()

		bhhc := &factory.BlockChainHookHandlerFactoryMock{}
		runTypeComponents := componentsMock.GetRunTypeComponents()
		sovereignVmContainerShardFactory, err := vm.NewSovereignVmContainerShardFactory(bhhc, nil, runTypeComponents.VmContainerShardFactoryCreator())
		require.ErrorIs(t, err, vm.ErrNilVmContainerMetaCreator)
		require.True(t, sovereignVmContainerShardFactory.IsInterfaceNil())
	})

	t.Run("nil vmContainerShardFactory", func(t *testing.T) {
		t.Parallel()

		bhhc := &factory.BlockChainHookHandlerFactoryMock{}
		runTypeComponents := componentsMock.GetRunTypeComponents()
		sovereignVmContainerShardFactory, err := vm.NewSovereignVmContainerShardFactory(bhhc, runTypeComponents.VmContainerMetaFactoryCreator(), nil)
		require.ErrorIs(t, err, vm.ErrNilVmContainerShardCreator)
		require.True(t, sovereignVmContainerShardFactory.IsInterfaceNil())
	})
}

func TestNewSovereignVmContainerShardFactory_CreateVmContainerFactoryShard(t *testing.T) {
	t.Parallel()

	bhhc := &factory.BlockChainHookHandlerFactoryMock{}
	runTypeComponents := componentsMock.GetRunTypeComponents()
	sovereignVmContainerShardFactory, err := vm.NewSovereignVmContainerShardFactory(bhhc, runTypeComponents.VmContainerMetaFactoryCreator(), runTypeComponents.VmContainerShardFactoryCreator())
	require.Nil(t, err)
	require.False(t, sovereignVmContainerShardFactory.IsInterfaceNil())

	argsBlockchain := createMockBlockChainHookArgs()
	argsShard := createMockVMAccountsArguments()
	gasSchedule := makeGasSchedule()
	argsMeta := createVmContainerMockArgument(gasSchedule)
	args := vm.ArgsVmContainerFactory{
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

	vmContainer, vmFactory, err := sovereignVmContainerShardFactory.CreateVmContainerFactory(argsBlockchain, args)
	require.Nil(t, err)
	require.Equal(t, "*containers.virtualMachinesContainer", fmt.Sprintf("%T", vmContainer))
	require.Equal(t, "*shard.vmContainerFactory", fmt.Sprintf("%T", vmFactory))

	require.Equal(t, 2, vmContainer.Len())
	svm, err := vmContainer.Get(factory2.SystemVirtualMachine)
	require.Nil(t, err)
	require.NotNil(t, svm)
	require.Equal(t, "*process.systemVM", fmt.Sprintf("%T", svm))

	wasmvm, err := vmContainer.Get(factory2.WasmVirtualMachine)
	require.Nil(t, err)
	require.NotNil(t, wasmvm)
	require.Equal(t, "*hostCore.vmHost", fmt.Sprintf("%T", wasmvm))
}
