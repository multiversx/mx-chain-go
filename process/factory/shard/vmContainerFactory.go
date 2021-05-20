package shard

import (
	arwen1_2 "github.com/ElrondNetwork/arwen-wasm-vm/arwen"
	arwenHost1_2 "github.com/ElrondNetwork/arwen-wasm-vm/arwen/host"
	ipcCommon1_2 "github.com/ElrondNetwork/arwen-wasm-vm/ipc/common"
	ipcMarshaling1_2 "github.com/ElrondNetwork/arwen-wasm-vm/ipc/marshaling"
	ipcNodePart1_2 "github.com/ElrondNetwork/arwen-wasm-vm/ipc/nodepart"
	arwen1_3 "github.com/ElrondNetwork/arwen-wasm-vm/v1_3/arwen"
	arwenHost1_3 "github.com/ElrondNetwork/arwen-wasm-vm/v1_3/arwen/host"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
)

var _ process.VirtualMachinesContainerFactory = (*vmContainerFactory)(nil)

var logVMContainerFactory = logger.GetOrCreate("vmContainerFactory")

type vmContainerFactory struct {
	config                               config.VirtualMachineConfig
	blockChainHookImpl                   *hooks.BlockChainHookImpl
	cryptoHook                           vmcommon.CryptoHook
	blockGasLimit                        uint64
	gasSchedule                          core.GasScheduleNotifier
	builtinFunctions                     vmcommon.FunctionNames
	epochNotifier                        process.EpochNotifier
	deployEnableEpoch                    uint32
	aheadOfTimeGasUsageEnableEpoch       uint32
	arwenV3EnableEpoch                   uint32
	ArwenESDTFunctionsEnableEpoch        uint32
	ArwenGasManagementRewriteEnableEpoch uint32
	container                            process.VirtualMachinesContainer
}

// ArgVMContainerFactory defines the arguments needed to the new VM factory
type ArgVMContainerFactory struct {
	Config                               config.VirtualMachineConfig
	BlockGasLimit                        uint64
	GasSchedule                          core.GasScheduleNotifier
	ArgBlockChainHook                    hooks.ArgBlockChainHook
	EpochNotifier                        process.EpochNotifier
	DeployEnableEpoch                    uint32
	AheadOfTimeGasUsageEnableEpoch       uint32
	ArwenV3EnableEpoch                   uint32
	ArwenESDTFunctionsEnableEpoch        uint32
	ArwenGasManagementRewriteEnableEpoch uint32
}

// NewVMContainerFactory is responsible for creating a new virtual machine factory object
func NewVMContainerFactory(args ArgVMContainerFactory) (*vmContainerFactory, error) {
	if check.IfNil(args.GasSchedule) {
		return nil, process.ErrNilGasSchedule
	}

	blockChainHookImpl, err := hooks.NewBlockChainHookImpl(args.ArgBlockChainHook)
	if err != nil {
		return nil, err
	}

	cryptoHook := hooks.NewVMCryptoHook()
	builtinFunctions := blockChainHookImpl.GetBuiltinFunctionNames()

	factory := &vmContainerFactory{
		config:                               args.Config,
		blockChainHookImpl:                   blockChainHookImpl,
		cryptoHook:                           cryptoHook,
		blockGasLimit:                        args.BlockGasLimit,
		gasSchedule:                          args.GasSchedule,
		builtinFunctions:                     builtinFunctions,
		epochNotifier:                        args.EpochNotifier,
		deployEnableEpoch:                    args.DeployEnableEpoch,
		aheadOfTimeGasUsageEnableEpoch:       args.AheadOfTimeGasUsageEnableEpoch,
		arwenV3EnableEpoch:                   args.ArwenV3EnableEpoch,
		ArwenESDTFunctionsEnableEpoch:        args.ArwenESDTFunctionsEnableEpoch,
		ArwenGasManagementRewriteEnableEpoch: args.ArwenGasManagementRewriteEnableEpoch,
		container:                            nil,
	}

	factory.epochNotifier.RegisterNotifyHandler(factory)
	return factory, nil
}

// Create sets up all the needed virtual machine returning a container of all the VMs
func (vmf *vmContainerFactory) Create() (process.VirtualMachinesContainer, error) {
	container := containers.NewVirtualMachinesContainer()

	currVm, err := vmf.createArwenVM()
	if err != nil {
		return nil, err
	}
	vmf.gasSchedule.RegisterNotifyHandler(currVm)

	err = container.Add(factory.ArwenVirtualMachine, currVm)
	if err != nil {
		return nil, err
	}

	// The vmContainerFactory keeps a reference to the container it has created,
	// in order to replace, from within the container, the VM instances that
	// become out-of-date after specific epochs.
	vmf.container = container

	return container, nil
}

// Close closes the vm container factory
func (vmf *vmContainerFactory) Close() error {
	return vmf.blockChainHookImpl.Close()
}

// EpochConfirmed updates VMs in all created containers depending on the epoch
func (vmf *vmContainerFactory) EpochConfirmed(epoch uint32) {
	if vmf.config.OutOfProcessEnabled {
		return
	}

	if epoch == vmf.ArwenGasManagementRewriteEnableEpoch {
		vmf.replaceInProcessArwen(epoch)
	}
}

func (vmf *vmContainerFactory) createArwenVM() (vmcommon.VMExecutionHandler, error) {
	if vmf.config.OutOfProcessEnabled {
		return vmf.createOutOfProcessArwenVM()
	}

	return vmf.createInProcessArwenVM()
}

func (vmf *vmContainerFactory) createOutOfProcessArwenVM() (vmcommon.VMExecutionHandler, error) {
	logVMContainerFactory.Debug("createOutOfProcessArwenVM", "config", vmf.config)

	outOfProcessConfig := vmf.config.OutOfProcessConfig
	logsMarshalizer := ipcMarshaling1_2.ParseKind(outOfProcessConfig.LogsMarshalizer)
	messagesMarshalizer := ipcMarshaling1_2.ParseKind(outOfProcessConfig.MessagesMarshalizer)
	maxLoopTime := outOfProcessConfig.MaxLoopTime

	logger.GetLogLevelPattern()

	arwenVM, err := ipcNodePart1_2.NewArwenDriver(
		vmf.blockChainHookImpl,
		ipcCommon1_2.ArwenArguments{
			VMHostParameters: arwen1_2.VMHostParameters{
				VMType:                   factory.ArwenVirtualMachine,
				BlockGasLimit:            vmf.blockGasLimit,
				GasSchedule:              vmf.gasSchedule.LatestGasSchedule(),
				ProtocolBuiltinFunctions: vmf.builtinFunctions,
				ElrondProtectedKeyPrefix: []byte(core.ElrondProtectedKeyPrefix),
				ArwenV2EnableEpoch:       vmf.deployEnableEpoch,
				AheadOfTimeEnableEpoch:   vmf.aheadOfTimeGasUsageEnableEpoch,
				DynGasLockEnableEpoch:    vmf.deployEnableEpoch,
				ArwenV3EnableEpoch:       vmf.arwenV3EnableEpoch,
			},
			LogsMarshalizer:     logsMarshalizer,
			MessagesMarshalizer: messagesMarshalizer,
		},
		ipcNodePart1_2.Config{MaxLoopTime: maxLoopTime},
	)
	return arwenVM, err
}

func (vmf *vmContainerFactory) createInProcessArwenVM() (vmcommon.VMExecutionHandler, error) {
	logVMContainerFactory.Debug("createInProcessArwenVM", "config", vmf.config)

	return vmf.createInProcessArwenVMByEpoch(vmf.epochNotifier.CurrentEpoch())
}

func (vmf *vmContainerFactory) createInProcessArwenVMByEpoch(epoch uint32) (vmcommon.VMExecutionHandler, error) {
	if epoch < vmf.ArwenGasManagementRewriteEnableEpoch {
		return vmf.createInProcessArwenVM_v1_2()
	}

	return vmf.createInProcessArwenVM_v1_3()
}

func (vmf *vmContainerFactory) replaceInProcessArwen(epoch uint32) {
	newArwenVM, err := vmf.createInProcessArwenVMByEpoch(epoch)
	if err != nil {
		logVMContainerFactory.Error("cannot replace Arwen VM in epoch", "epoch", epoch)
		return
	}

	vmf.container.Replace(factory.ArwenVirtualMachine, newArwenVM)
	logVMContainerFactory.Debug("Arwen VM replaced", "epoch", epoch)
}

func (vmf *vmContainerFactory) createInProcessArwenVM_v1_2() (vmcommon.VMExecutionHandler, error) {
	hostParameters := &arwen1_2.VMHostParameters{
		VMType:                   factory.ArwenVirtualMachine,
		BlockGasLimit:            vmf.blockGasLimit,
		GasSchedule:              vmf.gasSchedule.LatestGasSchedule(),
		ProtocolBuiltinFunctions: vmf.builtinFunctions,
		ElrondProtectedKeyPrefix: []byte(core.ElrondProtectedKeyPrefix),
		ArwenV2EnableEpoch:       vmf.deployEnableEpoch,
		AheadOfTimeEnableEpoch:   vmf.aheadOfTimeGasUsageEnableEpoch,
		DynGasLockEnableEpoch:    vmf.deployEnableEpoch,
		ArwenV3EnableEpoch:       vmf.arwenV3EnableEpoch,
	}
	return arwenHost1_2.NewArwenVM(vmf.blockChainHookImpl, hostParameters)
}

func (vmf *vmContainerFactory) createInProcessArwenVM_v1_3() (vmcommon.VMExecutionHandler, error) {
	hostParameters := &arwen1_3.VMHostParameters{
		VMType:                   factory.ArwenVirtualMachine,
		BlockGasLimit:            vmf.blockGasLimit,
		GasSchedule:              vmf.gasSchedule.LatestGasSchedule(),
		ProtocolBuiltinFunctions: vmf.builtinFunctions,
		ElrondProtectedKeyPrefix: []byte(core.ElrondProtectedKeyPrefix),
		ArwenV2EnableEpoch:       vmf.deployEnableEpoch,
		AheadOfTimeEnableEpoch:   vmf.aheadOfTimeGasUsageEnableEpoch,
		DynGasLockEnableEpoch:    vmf.deployEnableEpoch,
		ArwenV3EnableEpoch:       vmf.arwenV3EnableEpoch,
	}
	return arwenHost1_3.NewArwenVM(vmf.blockChainHookImpl, hostParameters)
}

// BlockChainHookImpl returns the created blockChainHookImpl
func (vmf *vmContainerFactory) BlockChainHookImpl() process.BlockChainHookHandler {
	return vmf.blockChainHookImpl
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmf *vmContainerFactory) IsInterfaceNil() bool {
	return vmf == nil
}
