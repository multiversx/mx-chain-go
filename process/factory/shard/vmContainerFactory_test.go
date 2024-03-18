package shard

import (
	"runtime"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	vmcommonBuiltInFunctions "github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"
	ipcNodePart1p2 "github.com/multiversx/mx-chain-vm-v1_2-go/ipc/nodepart"
	"github.com/stretchr/testify/assert"
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

func createMockVMAccountsArguments() ArgVMContainerFactory {
	esdtTransferParser, _ := parsers.NewESDTTransferParser(&mock.MarshalizerMock{})
	return ArgVMContainerFactory{
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
	args.WasmVMChangeLocker = nil
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

func TestNewVMContainerFactory_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.EnableEpochsHandler = nil
	vmf, err := NewVMContainerFactory(args)

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
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

func TestNewVMContainerFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockVMAccountsArguments()
	args.Hasher = nil
	vmf, err := NewVMContainerFactory(args)

	assert.Nil(t, vmf)
	assert.Equal(t, process.ErrNilHasher, err)
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
	if runtime.GOARCH == "arm64" {
		t.Skip("skipping test on arm64")
	}

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

	vm, err := container.Get(factory.WasmVirtualMachine)
	assert.Nil(t, err)
	assert.NotNil(t, vm)

	acc := vmf.BlockChainHookImpl()
	assert.NotNil(t, acc)
}

func TestVmContainerFactory_ResolveWasmVMVersion(t *testing.T) {
	if runtime.GOARCH == "arm64" {
		t.Skip("skipping test on arm64")
	}

	epochNotifierInstance := forking.NewGenericEpochNotifier()

	numCalled := 0
	gasScheduleNotifier := testscommon.NewGasScheduleNotifierMock(wasmConfig.MakeGasMapForTests())
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
	require.Equal(t, "v1.2", getWasmVMVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifierInstance.CheckEpoch(makeHeaderHandlerStub(1))
	require.Equal(t, "v1.2", getWasmVMVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifierInstance.CheckEpoch(makeHeaderHandlerStub(6))
	require.Equal(t, "v1.2", getWasmVMVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifierInstance.CheckEpoch(makeHeaderHandlerStub(10))
	require.Equal(t, "v1.2", getWasmVMVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifierInstance.CheckEpoch(makeHeaderHandlerStub(11))
	require.Equal(t, "v1.2", getWasmVMVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifierInstance.CheckEpoch(makeHeaderHandlerStub(12))
	require.Equal(t, "v1.3", getWasmVMVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifierInstance.CheckEpoch(makeHeaderHandlerStub(13))
	require.Equal(t, "v1.3", getWasmVMVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	epochNotifierInstance.CheckEpoch(makeHeaderHandlerStub(20))
	require.Equal(t, "v1.4", getWasmVMVersion(t, container))
	require.False(t, isOutOfProcess(t, container))

	require.Equal(t, numCalled, 1)
}

func makeHeaderHandlerStub(epoch uint32) data.HeaderHandler {
	return &testscommon.HeaderHandlerStub{
		EpochField: epoch,
	}
}

func isOutOfProcess(t testing.TB, container process.VirtualMachinesContainer) bool {
	vm, err := container.Get(factory.WasmVirtualMachine)
	require.Nil(t, err)
	require.NotNil(t, vm)

	_, ok := vm.(*ipcNodePart1p2.VMDriver)
	return ok
}

func getWasmVMVersion(t testing.TB, container process.VirtualMachinesContainer) string {
	vm, err := container.Get(factory.WasmVirtualMachine)
	require.Nil(t, err)
	require.NotNil(t, vm)

	return vm.GetVersion()
}
