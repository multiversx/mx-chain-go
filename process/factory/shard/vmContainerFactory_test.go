package shard

import (
	"testing"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/config"
	ipcNodePart1_2 "github.com/ElrondNetwork/arwen-wasm-vm/ipc/nodepart"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/forking"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockVMAccountsArguments() hooks.ArgBlockChainHook {
	datapool := testscommon.NewPoolsHolderMock()
	arguments := hooks.ArgBlockChainHook{
		Accounts: &mock.AccountsStub{
			GetExistingAccountCalled: func(address []byte) (handler state.AccountHandler, e error) {
				return &mock.AccountWrapMock{}, nil
			},
		},
		PubkeyConv:         mock.NewPubkeyConverterMock(32),
		StorageService:     &mock.ChainStorerMock{},
		BlockChain:         &mock.BlockChainMock{},
		ShardCoordinator:   mock.NewOneShardCoordinatorMock(),
		Marshalizer:        &mock.MarshalizerMock{},
		Uint64Converter:    &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions:   builtInFunctions.NewBuiltInFunctionContainer(),
		DataPool:           datapool,
		CompiledSCPool:     datapool.SmartContracts(),
		NilCompiledSCStore: true,
	}
	return arguments
}

func TestNewVMContainerFactory_NilGasScheduleShouldErr(t *testing.T) {
	t.Parallel()

	argsNewVMFactory := ArgVMContainerFactory{
		Config:                         config.VirtualMachineConfig{},
		BlockGasLimit:                  10000,
		GasSchedule:                    nil,
		ArgBlockChainHook:              createMockVMAccountsArguments(),
		DeployEnableEpoch:              0,
		AheadOfTimeGasUsageEnableEpoch: 0,
		ArwenV3EnableEpoch:             0,
		ArwenESDTFunctionsEnableEpoch:  0,
	}
	vmf, err := NewVMContainerFactory(argsNewVMFactory)

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilGasSchedule, err)
}

func TestNewVMContainerFactory_OkValues(t *testing.T) {
	t.Parallel()

	argsNewVMFactory := ArgVMContainerFactory{
		Config:                         config.VirtualMachineConfig{},
		BlockGasLimit:                  10000,
		GasSchedule:                    mock.NewGasScheduleNotifierMock(arwenConfig.MakeGasMapForTests()),
		ArgBlockChainHook:              createMockVMAccountsArguments(),
		DeployEnableEpoch:              0,
		AheadOfTimeGasUsageEnableEpoch: 0,
		ArwenV3EnableEpoch:             0,
		ArwenESDTFunctionsEnableEpoch:  0,
	}
	vmf, err := NewVMContainerFactory(argsNewVMFactory)

	assert.NotNil(t, vmf)
	assert.Nil(t, err)
	assert.False(t, vmf.IsInterfaceNil())
}

func TestVmContainerFactory_Create(t *testing.T) {
	t.Parallel()

	argsNewVMFactory := ArgVMContainerFactory{
		Config:                         config.VirtualMachineConfig{},
		BlockGasLimit:                  10000,
		GasSchedule:                    mock.NewGasScheduleNotifierMock(arwenConfig.MakeGasMapForTests()),
		ArgBlockChainHook:              createMockVMAccountsArguments(),
		DeployEnableEpoch:              0,
		AheadOfTimeGasUsageEnableEpoch: 0,
		ArwenV3EnableEpoch:             0,
		ArwenESDTFunctionsEnableEpoch:  0,
	}
	vmf, err := NewVMContainerFactory(argsNewVMFactory)
	assert.NotNil(t, vmf)
	assert.Nil(t, err)

	container, err := vmf.Create()
	require.Nil(t, err)
	require.NotNil(t, container)
	defer func() {
		_ = container.Close()
	}()

	assert.Nil(t, err)
	assert.NotNil(t, container)

	vm, err := container.Get(factory.ArwenVirtualMachine)
	assert.Nil(t, err)
	assert.NotNil(t, vm)

	acc := vmf.BlockChainHookImpl()
	assert.NotNil(t, acc)
}

func TestVmContainerFactory_ResolveArwenVersion(t *testing.T) {
	vmConfig := config.VirtualMachineConfig{
		OutOfProcessEnabled: true,
		OutOfProcessConfig: config.VirtualMachineOutOfProcessConfig{
			LogsMarshalizer:     "json",
			MessagesMarshalizer: "json",
			MaxLoopTime:         1000,
		},
		ArwenVersions: []config.ArwenVersionByEpoch{
			{StartEpoch: 0, Version: "v1.2", OutOfProcessSupported: false},
			{StartEpoch: 10, Version: "v1.2", OutOfProcessSupported: true},
			{StartEpoch: 12, Version: "v1.3", OutOfProcessSupported: false},
		},
	}

	epochNotifier := forking.NewGenericEpochNotifier()

	argsNewVMFactory := ArgVMContainerFactory{
		Config:                         vmConfig,
		BlockGasLimit:                  10000,
		GasSchedule:                    mock.NewGasScheduleNotifierMock(arwenConfig.MakeGasMapForTests()),
		ArgBlockChainHook:              createMockVMAccountsArguments(),
		EpochNotifier:                  epochNotifier,
		DeployEnableEpoch:              0,
		AheadOfTimeGasUsageEnableEpoch: 0,
		ArwenV3EnableEpoch:             0,
	}

	vmf, err := NewVMContainerFactory(argsNewVMFactory)
	require.NotNil(t, vmf)
	require.Nil(t, err)

	container, err := vmf.Create()
	require.Nil(t, err)
	require.NotNil(t, container)
	defer func() {
		_ = container.Close()
	}()
	require.Equal(t, "v1.2", getArwenVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifier.CheckEpoch(1)
	require.Equal(t, "v1.2", getArwenVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifier.CheckEpoch(6)
	require.Equal(t, "v1.2", getArwenVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifier.CheckEpoch(10)
	require.Equal(t, "v1.2", getArwenVersion(t, container))
	require.True(t, isOutOfProcess(t, container))

	epochNotifier.CheckEpoch(11)
	require.Equal(t, "v1.2", getArwenVersion(t, container))
	require.True(t, isOutOfProcess(t, container))

	epochNotifier.CheckEpoch(12)
	require.Equal(t, "v1.3", getArwenVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifier.CheckEpoch(13)
	require.Equal(t, "v1.3", getArwenVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifier.CheckEpoch(20)
	require.Equal(t, "v1.3", getArwenVersion(t, container))
	require.False(t, isOutOfProcess(t, container))
}

func isOutOfProcess(t testing.TB, container process.VirtualMachinesContainer) bool {
	vm, err := container.Get(factory.ArwenVirtualMachine)
	require.Nil(t, err)
	require.NotNil(t, vm)

	_, ok := vm.(*ipcNodePart1_2.ArwenDriver)
	return ok
}

func getArwenVersion(t testing.TB, container process.VirtualMachinesContainer) string {
	vm, err := container.Get(factory.ArwenVirtualMachine)
	require.Nil(t, err)
	require.NotNil(t, vm)

	return vm.GetVersion()
}
