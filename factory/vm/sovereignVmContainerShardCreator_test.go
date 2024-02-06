package vm_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/factory/vm"
	factory2 "github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignVmContainerShardCreatorFactory(t *testing.T) {
	t.Parallel()

	bhhc := &factory.BlockChainHookHandlerFactoryMock{}
	sovereignVmContainerShardFactory, err := vm.NewSovereignVmContainerShardFactory(bhhc)
	require.Nil(t, err)
	require.False(t, sovereignVmContainerShardFactory.IsInterfaceNil())
}

func TestNewSovereignVmContainerShardFactory_CreateVmContainerFactoryShard(t *testing.T) {
	t.Parallel()

	bhhc := &factory.BlockChainHookHandlerFactoryMock{}
	sovereignVmContainerShardFactory, err := vm.NewSovereignVmContainerShardFactory(bhhc)
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
	}

	vmContainer, vmFactory, err := sovereignVmContainerShardFactory.CreateVmContainerFactoryShard(argsBlockchain, args)
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
