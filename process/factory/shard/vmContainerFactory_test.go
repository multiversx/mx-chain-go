package shard

import (
	"sync"
	"testing"

	ipcNodePart1p2 "github.com/ElrondNetwork/arwen-wasm-vm/v1_2/ipc/nodepart"
	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/config"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/common/forking"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	vmcommonBuiltInFunctions "github.com/ElrondNetwork/elrond-vm-common/builtInFunctions"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeVMConfig() config.VirtualMachineConfig {
	return config.VirtualMachineConfig{
		ArwenVersions: []config.ArwenVersionByEpoch{
			{StartEpoch: 0, Version: "v1.2"},
			{StartEpoch: 10, Version: "v1.2"},
			{StartEpoch: 12, Version: "v1.3"},
			{StartEpoch: 14, Version: "v1.4"},
		},
	}
}

func createMockVMAccountsArguments() ArgVMContainerFactory {
	esdtTransferParser, _ := parsers.NewESDTTransferParser(&mock.MarshalizerMock{})
	return ArgVMContainerFactory{
		Config:              makeVMConfig(),
		BlockGasLimit:       10000,
		GasSchedule:         mock.NewGasScheduleNotifierMock(arwenConfig.MakeGasMapForTests()),
		EpochNotifier:       &epochNotifier.EpochNotifierStub{},
		EnableEpochsHandler: &testscommon.EnableEpochsHandlerStub{},
		ArwenChangeLocker:   &sync.RWMutex{},
		ESDTTransferParser:  esdtTransferParser,
		BuiltInFunctions:    vmcommonBuiltInFunctions.NewBuiltInFunctionContainer(),
		BlockChainHook:      &testscommon.BlockChainHookStub{},
	}
}

func TestNewVMContainerFactory_NilGasScheduleShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.GasSchedule = nil
	vmf, err := NewVMContainerFactory(args)

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilGasSchedule, err)
}

func TestNewVMContainerFactory_NilESDTTransferParserShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.ESDTTransferParser = nil
	vmf, err := NewVMContainerFactory(args)

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilESDTTransferParser, err)
}

func TestNewVMContainerFactory_NilLockerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.ArwenChangeLocker = nil
	vmf, err := NewVMContainerFactory(args)

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilLocker, err)
}

func TestNewVMContainerFactory_NilEpochNotifierShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.EpochNotifier = nil
	vmf, err := NewVMContainerFactory(args)

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilEpochNotifier, err)
}

func TestNewVMContainerFactory_NilBuiltinFunctionsShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.BuiltInFunctions = nil
	vmf, err := NewVMContainerFactory(args)

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilBuiltInFunction, err)
}

func TestNewVMContainerFactory_NilBlockChainHookShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.BlockChainHook = nil
	vmf, err := NewVMContainerFactory(args)

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilBlockChainHook, err)
}

func TestNewVMContainerFactory_OkValues(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	vmf, err := NewVMContainerFactory(args)

	assert.NotNil(t, vmf)
	assert.Nil(t, err)
	assert.False(t, vmf.IsInterfaceNil())
}

func TestVmContainerFactory_Create(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	vmf, _ := NewVMContainerFactory(args)
	require.NotNil(t, vmf)

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
	epochNotifierInstance := forking.NewGenericEpochNotifier()

	numCalled := 0
	gasScheduleNotifier := mock.NewGasScheduleNotifierMock(arwenConfig.MakeGasMapForTests())
	gasScheduleNotifier.RegisterNotifyHandlerCalled = func(handler core.GasScheduleSubscribeHandler) {
		numCalled++
		handler.GasScheduleChange(gasScheduleNotifier.GasSchedule)
	}
	args := createMockVMAccountsArguments()
	args.GasSchedule = gasScheduleNotifier
	args.EpochNotifier = epochNotifierInstance
	vmf, _ := NewVMContainerFactory(args)
	require.NotNil(t, vmf)

	container, err := vmf.Create()
	require.Nil(t, err)
	require.NotNil(t, container)
	defer func() {
		_ = container.Close()
	}()
	require.Equal(t, "v1.2", getArwenVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifierInstance.CheckEpoch(makeHeaderHandlerStub(1))
	require.Equal(t, "v1.2", getArwenVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifierInstance.CheckEpoch(makeHeaderHandlerStub(6))
	require.Equal(t, "v1.2", getArwenVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifierInstance.CheckEpoch(makeHeaderHandlerStub(10))
	require.Equal(t, "v1.2", getArwenVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifierInstance.CheckEpoch(makeHeaderHandlerStub(11))
	require.Equal(t, "v1.2", getArwenVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifierInstance.CheckEpoch(makeHeaderHandlerStub(12))
	require.Equal(t, "v1.3", getArwenVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifierInstance.CheckEpoch(makeHeaderHandlerStub(13))
	require.Equal(t, "v1.3", getArwenVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifierInstance.CheckEpoch(makeHeaderHandlerStub(20))
	require.Equal(t, "v1.4", getArwenVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	require.Equal(t, numCalled, 1)
}

func makeHeaderHandlerStub(epoch uint32) data.HeaderHandler {
	return &testscommon.HeaderHandlerStub{
		EpochField: epoch,
	}
}

func isOutOfProcess(t testing.TB, container process.VirtualMachinesContainer) bool {
	vm, err := container.Get(factory.ArwenVirtualMachine)
	require.Nil(t, err)
	require.NotNil(t, vm)

	_, ok := vm.(*ipcNodePart1p2.ArwenDriver)
	return ok
}

func getArwenVersion(t testing.TB, container process.VirtualMachinesContainer) string {
	vm, err := container.Get(factory.ArwenVirtualMachine)
	require.Nil(t, err)
	require.NotNil(t, vm)

	return vm.GetVersion()
}
