package shard

import (
	"github.com/ElrondNetwork/arwen-wasm-vm/arwen"
	arwenHost "github.com/ElrondNetwork/arwen-wasm-vm/arwen/host"
	ipcCommon "github.com/ElrondNetwork/arwen-wasm-vm/ipc/common"
	ipcMarshaling "github.com/ElrondNetwork/arwen-wasm-vm/ipc/marshaling"
	ipcNodePart "github.com/ElrondNetwork/arwen-wasm-vm/ipc/nodepart"
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
	config                         config.VirtualMachineConfig
	blockChainHookImpl             *hooks.BlockChainHookImpl
	cryptoHook                     vmcommon.CryptoHook
	blockGasLimit                  uint64
	gasSchedule                    core.GasScheduleNotifier
	builtinFunctions               vmcommon.FunctionNames
	deployEnableEpoch              uint32
	aheadOfTimeGasUsageEnableEpoch uint32
	arwenV3EnableEpoch             uint32
}

// ArgVMContainerFactory defines the arguments needed to the new VM factory
type ArgVMContainerFactory struct {
	Config                         config.VirtualMachineConfig
	BlockGasLimit                  uint64
	GasSchedule                    core.GasScheduleNotifier
	ArgBlockChainHook              hooks.ArgBlockChainHook
	DeployEnableEpoch              uint32
	AheadOfTimeGasUsageEnableEpoch uint32
	ArwenV3EnableEpoch             uint32
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

	return &vmContainerFactory{
		config:                         args.Config,
		blockChainHookImpl:             blockChainHookImpl,
		cryptoHook:                     cryptoHook,
		blockGasLimit:                  args.BlockGasLimit,
		gasSchedule:                    args.GasSchedule,
		builtinFunctions:               builtinFunctions,
		deployEnableEpoch:              args.DeployEnableEpoch,
		aheadOfTimeGasUsageEnableEpoch: args.AheadOfTimeGasUsageEnableEpoch,
		arwenV3EnableEpoch:             args.ArwenV3EnableEpoch,
	}, nil
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

	return container, nil
}

// Close closes the vm container factory
func (vmf *vmContainerFactory) Close() error{
	return vmf.blockChainHookImpl.Close()
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
	logsMarshalizer := ipcMarshaling.ParseKind(outOfProcessConfig.LogsMarshalizer)
	messagesMarshalizer := ipcMarshaling.ParseKind(outOfProcessConfig.MessagesMarshalizer)
	maxLoopTime := outOfProcessConfig.MaxLoopTime

	logger.GetLogLevelPattern()

	arwenVM, err := ipcNodePart.NewArwenDriver(
		vmf.blockChainHookImpl,
		ipcCommon.ArwenArguments{
			VMHostParameters: arwen.VMHostParameters{
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
		ipcNodePart.Config{MaxLoopTime: maxLoopTime},
	)
	return arwenVM, err
}

func (vmf *vmContainerFactory) createInProcessArwenVM() (vmcommon.VMExecutionHandler, error) {
	logVMContainerFactory.Debug("createInProcessArwenVM", "config", vmf.config)

	return arwenHost.NewArwenVM(
		vmf.blockChainHookImpl,
		&arwen.VMHostParameters{
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
	)
}

// BlockChainHookImpl returns the created blockChainHookImpl
func (vmf *vmContainerFactory) BlockChainHookImpl() process.BlockChainHookHandler {
	return vmf.blockChainHookImpl
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmf *vmContainerFactory) IsInterfaceNil() bool {
	return vmf == nil
}
