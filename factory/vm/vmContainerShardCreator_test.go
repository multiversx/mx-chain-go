package vm_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/factory/vm"
	"github.com/multiversx/mx-chain-go/process/factory/shard"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	vmcommonBuiltInFunctions "github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"
	"github.com/stretchr/testify/require"
)

func makeVMConfig() config.VirtualMachineConfig {
	return config.VirtualMachineConfig{
		WasmVMVersions: []config.WasmVMVersionByEpoch{
			{StartEpoch: 0, Version: "v1.2"},
			{StartEpoch: 10, Version: "v1.2"},
			{StartEpoch: 12, Version: "v1.3"},
			{StartEpoch: 14, Version: "v1.4"},
		},
	}
}

func createMockVMAccountsArguments() shard.ArgVMContainerFactory {
	esdtTransferParser, _ := parsers.NewESDTTransferParser(&mock.MarshalizerMock{})
	return shard.ArgVMContainerFactory{
		Config:              makeVMConfig(),
		BlockGasLimit:       10000,
		GasSchedule:         testscommon.NewGasScheduleNotifierMock(wasmConfig.MakeGasMapForTests()),
		EpochNotifier:       &epochNotifier.EpochNotifierStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		WasmVMChangeLocker:  &sync.RWMutex{},
		ESDTTransferParser:  esdtTransferParser,
		BuiltInFunctions:    vmcommonBuiltInFunctions.NewBuiltInFunctionContainer(),
		BlockChainHook:      &testscommon.BlockChainHookStub{},
		Hasher:              &hashingMocks.HasherMock{},
	}
}

func TestNewVmContainerShardCreatorFactory(t *testing.T) {
	t.Parallel()

	bhhc := &factory.BlockChainHookHandlerFactoryMock{}
	vmContainerShardFactory, err := vm.NewVmContainerShardFactory(bhhc)
	require.Nil(t, err)
	require.False(t, vmContainerShardFactory.IsInterfaceNil())
}

func TestNewVmContainerShardFactory_CreateVmContainerFactoryShard(t *testing.T) {
	t.Parallel()

	bhhc := &factory.BlockChainHookHandlerFactoryMock{}
	vmContainerShardFactory, err := vm.NewVmContainerShardFactory(bhhc)
	require.Nil(t, err)
	require.False(t, vmContainerShardFactory.IsInterfaceNil())

	argsBlockchain := createMockBlockChainHookArgs()
	argsShard := createMockVMAccountsArguments()
	args := vm.ArgsVmContainerFactory{
		Config:              argsShard.Config,
		BlockGasLimit:       argsShard.BlockGasLimit,
		GasSchedule:         argsShard.GasSchedule,
		EpochNotifier:       argsShard.EpochNotifier,
		EnableEpochsHandler: argsShard.EnableEpochsHandler,
		WasmVMChangeLocker:  argsShard.WasmVMChangeLocker,
		ESDTTransferParser:  argsShard.ESDTTransferParser,
		BuiltInFunctions:    argsShard.BuiltInFunctions,
		Hasher:              argsShard.Hasher,
	}

	vmContainer, vmFactory, err := vmContainerShardFactory.CreateVmContainerFactoryShard(argsBlockchain, args)
	require.Nil(t, err)
	require.Equal(t, "*containers.virtualMachinesContainer", fmt.Sprintf("%T", vmContainer))
	require.Equal(t, "*shard.vmContainerFactory", fmt.Sprintf("%T", vmFactory))
}
