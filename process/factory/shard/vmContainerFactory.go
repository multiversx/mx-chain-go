package shard

import (
	"fmt"
	"io"
	"sort"

	arwen14 "github.com/ElrondNetwork/wasm-vm/arwen"
	arwenHost14 "github.com/ElrondNetwork/wasm-vm/arwen/host"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	arwen12 "github.com/ElrondNetwork/wasm-vm-v1_2/arwen"
	arwenHost12 "github.com/ElrondNetwork/wasm-vm-v1_2/arwen/host"
	arwen13 "github.com/ElrondNetwork/wasm-vm-v1_3/arwen"
	arwenHost13 "github.com/ElrondNetwork/wasm-vm-v1_3/arwen/host"
)

var _ process.VirtualMachinesContainerFactory = (*vmContainerFactory)(nil)

var logVMContainerFactory = logger.GetOrCreate("vmContainerFactory")

type vmContainerFactory struct {
	config              config.VirtualMachineConfig
	blockChainHook      process.BlockChainHookHandler
	cryptoHook          vmcommon.CryptoHook
	blockGasLimit       uint64
	gasSchedule         core.GasScheduleNotifier
	builtinFunctions    vmcommon.BuiltInFunctionContainer
	epochNotifier       process.EpochNotifier
	enableEpochsHandler vmcommon.EnableEpochsHandler
	container           process.VirtualMachinesContainer
	arwenVersions       []config.ArwenVersionByEpoch
	arwenChangeLocker   common.Locker
	esdtTransferParser  vmcommon.ESDTTransferParser
}

// ArgVMContainerFactory defines the arguments needed to the new VM factory
type ArgVMContainerFactory struct {
	Config              config.VirtualMachineConfig
	BlockGasLimit       uint64
	GasSchedule         core.GasScheduleNotifier
	EpochNotifier       process.EpochNotifier
	EnableEpochsHandler vmcommon.EnableEpochsHandler
	ArwenChangeLocker   common.Locker
	ESDTTransferParser  vmcommon.ESDTTransferParser
	BuiltInFunctions    vmcommon.BuiltInFunctionContainer
	BlockChainHook      process.BlockChainHookHandler
}

// NewVMContainerFactory is responsible for creating a new virtual machine factory object
func NewVMContainerFactory(args ArgVMContainerFactory) (*vmContainerFactory, error) {
	if check.IfNil(args.GasSchedule) {
		return nil, process.ErrNilGasSchedule
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	if check.IfNilReflect(args.ArwenChangeLocker) {
		return nil, process.ErrNilLocker
	}
	if check.IfNil(args.ESDTTransferParser) {
		return nil, process.ErrNilESDTTransferParser
	}
	if check.IfNil(args.BuiltInFunctions) {
		return nil, process.ErrNilBuiltInFunction
	}
	if check.IfNil(args.BlockChainHook) {
		return nil, process.ErrNilBlockChainHook
	}

	cryptoHook := hooks.NewVMCryptoHook()

	vmf := &vmContainerFactory{
		config:              args.Config,
		blockChainHook:      args.BlockChainHook,
		cryptoHook:          cryptoHook,
		blockGasLimit:       args.BlockGasLimit,
		gasSchedule:         args.GasSchedule,
		builtinFunctions:    args.BuiltInFunctions,
		epochNotifier:       args.EpochNotifier,
		enableEpochsHandler: args.EnableEpochsHandler,
		container:           nil,
		arwenChangeLocker:   args.ArwenChangeLocker,
		esdtTransferParser:  args.ESDTTransferParser,
	}

	vmf.arwenVersions = args.Config.ArwenVersions
	vmf.sortArwenVersions()
	err := vmf.validateArwenVersions()
	if err != nil {
		return nil, err
	}

	return vmf, nil
}

func (vmf *vmContainerFactory) sortArwenVersions() {
	sort.Slice(vmf.arwenVersions, func(i, j int) bool {
		return vmf.arwenVersions[i].StartEpoch < vmf.arwenVersions[j].StartEpoch
	})
}

func (vmf *vmContainerFactory) validateArwenVersions() error {
	if len(vmf.arwenVersions) == 0 {
		return ErrEmptyVersionsByEpochsList
	}

	currentEpoch := uint32(0)
	for idx, ver := range vmf.arwenVersions {
		if idx == 0 && ver.StartEpoch != 0 {
			return fmt.Errorf("%w first version should start on epoch 0",
				ErrInvalidVersionOnEpochValues)
		}

		if idx > 0 && currentEpoch >= ver.StartEpoch {
			return fmt.Errorf("%w, StartEpoch is greater or equal to next epoch StartEpoch value, version %s",
				ErrInvalidVersionOnEpochValues, ver.Version)
		}
		currentEpoch = ver.StartEpoch

		if len(ver.Version) > common.MaxSoftwareVersionLengthInBytes {
			return fmt.Errorf("%w for version %s",
				ErrInvalidVersionStringTooLong, ver.Version)
		}
	}

	return nil
}

// Create sets up all the needed virtual machine returning a container of all the VMs
func (vmf *vmContainerFactory) Create() (process.VirtualMachinesContainer, error) {
	container := containers.NewVirtualMachinesContainer()

	vmf.arwenChangeLocker.Lock()
	version := vmf.getMatchingVersion(vmf.epochNotifier.CurrentEpoch())
	currentVM, err := vmf.createArwenVM(version)
	if err != nil {
		vmf.arwenChangeLocker.Unlock()
		return nil, err
	}

	err = container.Add(factory.ArwenVirtualMachine, currentVM)
	if err != nil {
		vmf.arwenChangeLocker.Unlock()
		return nil, err
	}

	// The vmContainerFactory keeps a reference to the container it has created,
	// in order to replace, from within the container, the VM instances that
	// become out-of-date after specific epochs.
	vmf.container = container
	vmf.arwenChangeLocker.Unlock()

	vmf.epochNotifier.RegisterNotifyHandler(vmf)
	vmf.gasSchedule.RegisterNotifyHandler(vmf)

	return container, nil
}

// Close closes the vm container factory
func (vmf *vmContainerFactory) Close() error {
	errContainer := vmf.container.Close()
	if errContainer != nil {
		logVMContainerFactory.Error("cannot close container", "error", errContainer)
	}

	errBlockchain := vmf.blockChainHook.Close()
	if errBlockchain != nil {
		logVMContainerFactory.Error("cannot close blockchain hook implementation", "error", errBlockchain)
	}

	if errContainer != nil {
		return errContainer
	}

	return errBlockchain
}

// GasScheduleChange updates the gas schedule map in all the components
func (vmf *vmContainerFactory) GasScheduleChange(gasSchedule map[string]map[string]uint64) {
	// clear compiled codes always before changing gasSchedule
	vmf.blockChainHook.ClearCompiledCodes()
	for _, key := range vmf.container.Keys() {
		currentVM, err := vmf.container.Get(key)
		if err != nil {
			logVMContainerFactory.Error("cannot get VM on GasSchedule Change", "error", err)
			continue
		}

		currentVM.GasScheduleChange(gasSchedule)
	}
}

// EpochConfirmed updates the VM version in the container, depending on the epoch
func (vmf *vmContainerFactory) EpochConfirmed(epoch uint32, _ uint64) {
	vmf.ensureCorrectArwenVersion(epoch)
}

func (vmf *vmContainerFactory) ensureCorrectArwenVersion(epoch uint32) {
	newVersion := vmf.getMatchingVersion(epoch)
	currentArwenVM, err := vmf.container.Get(factory.ArwenVirtualMachine)
	if err != nil {
		logVMContainerFactory.Error("cannot retrieve Arwen VM from container", "epoch", epoch, "error", err)
		return
	}

	if !vmf.shouldReplaceArwenInstance(newVersion, currentArwenVM) {
		return
	}

	vmf.arwenChangeLocker.Lock()
	defer vmf.arwenChangeLocker.Unlock()

	vmf.closePreviousVM(currentArwenVM)
	newArwenVM, err := vmf.createArwenVM(newVersion)
	if err != nil {
		logVMContainerFactory.Error("cannot replace Arwen VM", "epoch", epoch, "error", err)
		return
	}

	err = vmf.container.Replace(factory.ArwenVirtualMachine, newArwenVM)
	if err != nil {
		logVMContainerFactory.Error("cannot replace Arwen VM", "epoch", epoch, "error", err)
		return
	}

	logVMContainerFactory.Debug("Arwen VM replaced", "epoch", epoch)
}

func (vmf *vmContainerFactory) shouldReplaceArwenInstance(
	newVersion config.ArwenVersionByEpoch,
	currentVM vmcommon.VMExecutionHandler,
) bool {
	specificVersionRequired := newVersion.Version != "*"
	differentVersion := newVersion.Version != currentVM.GetVersion()

	return specificVersionRequired && differentVersion
}

func (vmf *vmContainerFactory) createArwenVM(version config.ArwenVersionByEpoch) (vmcommon.VMExecutionHandler, error) {
	currentVM, err := vmf.createInProcessArwenVMByVersion(version)
	if err != nil {
		return nil, err
	}

	return currentVM, nil
}

func (vmf *vmContainerFactory) getMatchingVersion(epoch uint32) config.ArwenVersionByEpoch {
	matchingVersion := vmf.arwenVersions[len(vmf.arwenVersions)-1]
	for idx := 0; idx < len(vmf.arwenVersions)-1; idx++ {
		crtVer := vmf.arwenVersions[idx]
		nextVer := vmf.arwenVersions[idx+1]
		if crtVer.StartEpoch <= epoch && epoch < nextVer.StartEpoch {
			return crtVer
		}
	}

	return matchingVersion
}

func (vmf *vmContainerFactory) createInProcessArwenVMByVersion(version config.ArwenVersionByEpoch) (vmcommon.VMExecutionHandler, error) {
	logVMContainerFactory.Debug("createInProcessArwenVM", "version", version)
	switch version.Version {
	case "v1.2":
		return vmf.createInProcessArwenVMV12()
	case "v1.3":
		return vmf.createInProcessArwenVMV13()
	default:
		return vmf.createInProcessArwenVMV14()
	}
}

func (vmf *vmContainerFactory) createInProcessArwenVMV12() (vmcommon.VMExecutionHandler, error) {
	hostParameters := &arwen12.VMHostParameters{
		VMType:                   factory.ArwenVirtualMachine,
		BlockGasLimit:            vmf.blockGasLimit,
		GasSchedule:              vmf.gasSchedule.LatestGasSchedule(),
		ProtocolBuiltinFunctions: vmf.builtinFunctions.Keys(),
		ElrondProtectedKeyPrefix: []byte(core.ElrondProtectedKeyPrefix),
		EnableEpochsHandler:      vmf.enableEpochsHandler,
	}
	return arwenHost12.NewArwenVM(vmf.blockChainHook, hostParameters)
}

func (vmf *vmContainerFactory) createInProcessArwenVMV13() (vmcommon.VMExecutionHandler, error) {
	hostParameters := &arwen13.VMHostParameters{
		VMType:                   factory.ArwenVirtualMachine,
		BlockGasLimit:            vmf.blockGasLimit,
		GasSchedule:              vmf.gasSchedule.LatestGasSchedule(),
		BuiltInFuncContainer:     vmf.builtinFunctions,
		ElrondProtectedKeyPrefix: []byte(core.ElrondProtectedKeyPrefix),
		EnableEpochsHandler:      vmf.enableEpochsHandler,
	}
	return arwenHost13.NewArwenVM(vmf.blockChainHook, hostParameters)
}

func (vmf *vmContainerFactory) createInProcessArwenVMV14() (vmcommon.VMExecutionHandler, error) {
	hostParameters := &arwen14.VMHostParameters{
		VMType:                              factory.ArwenVirtualMachine,
		BlockGasLimit:                       vmf.blockGasLimit,
		GasSchedule:                         vmf.gasSchedule.LatestGasSchedule(),
		BuiltInFuncContainer:                vmf.builtinFunctions,
		ElrondProtectedKeyPrefix:            []byte(core.ElrondProtectedKeyPrefix),
		ESDTTransferParser:                  vmf.esdtTransferParser,
		WasmerSIGSEGVPassthrough:            vmf.config.WasmerSIGSEGVPassthrough,
		TimeOutForSCExecutionInMilliseconds: vmf.config.TimeOutForSCExecutionInMilliseconds,
		EpochNotifier:                       vmf.epochNotifier,
		EnableEpochsHandler:                 vmf.enableEpochsHandler,
	}
	return arwenHost14.NewArwenVM(vmf.blockChainHook, hostParameters)
}

func (vmf *vmContainerFactory) closePreviousVM(vm vmcommon.VMExecutionHandler) {
	vmf.blockChainHook.ClearCompiledCodes()
	logVMContainerFactory.Debug("AOT compilation cache cleared")

	vmAsCloser, ok := vm.(io.Closer)
	if ok {
		err := vmAsCloser.Close()
		logVMContainerFactory.LogIfError(err, "closePreviousVM", "error", err)
	}
}

// BlockChainHookImpl returns the created blockChainHookImpl
func (vmf *vmContainerFactory) BlockChainHookImpl() process.BlockChainHookHandler {
	return vmf.blockChainHook
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmf *vmContainerFactory) IsInterfaceNil() bool {
	return vmf == nil
}
