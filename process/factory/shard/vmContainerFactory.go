package shard

import (
	"fmt"
	"io"
	"sort"

	arwen12 "github.com/ElrondNetwork/arwen-wasm-vm/v1_2/arwen"
	arwenHost12 "github.com/ElrondNetwork/arwen-wasm-vm/v1_2/arwen/host"
	arwen13 "github.com/ElrondNetwork/arwen-wasm-vm/v1_3/arwen"
	arwenHost13 "github.com/ElrondNetwork/arwen-wasm-vm/v1_3/arwen/host"
	arwen14 "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/arwen"
	arwenHost14 "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/arwen/host"
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
)

var _ process.VirtualMachinesContainerFactory = (*vmContainerFactory)(nil)

var logVMContainerFactory = logger.GetOrCreate("vmContainerFactory")

type vmContainerFactory struct {
	config             config.VirtualMachineConfig
	blockChainHookImpl *hooks.BlockChainHookImpl
	cryptoHook         vmcommon.CryptoHook
	blockGasLimit      uint64
	gasSchedule        core.GasScheduleNotifier
	builtinFunctions   vmcommon.BuiltInFunctionContainer
	epochNotifier      process.EpochNotifier
	epochConfig        config.EnableEpochs
	container          process.VirtualMachinesContainer
	arwenVersions      []config.ArwenVersionByEpoch
	arwenChangeLocker  process.Locker
	esdtTransferParser vmcommon.ESDTTransferParser
}

// ArgVMContainerFactory defines the arguments needed to the new VM factory
type ArgVMContainerFactory struct {
	Config             config.VirtualMachineConfig
	BlockGasLimit      uint64
	GasSchedule        core.GasScheduleNotifier
	ArgBlockChainHook  hooks.ArgBlockChainHook
	EpochNotifier      process.EpochNotifier
	EpochConfig        config.EnableEpochs
	ArwenChangeLocker  process.Locker
	ESDTTransferParser vmcommon.ESDTTransferParser
}

// NewVMContainerFactory is responsible for creating a new virtual machine factory object
func NewVMContainerFactory(args ArgVMContainerFactory) (*vmContainerFactory, error) {
	if check.IfNil(args.GasSchedule) {
		return nil, process.ErrNilGasSchedule
	}
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}
	if check.IfNilReflect(args.ArwenChangeLocker) {
		return nil, process.ErrNilLocker
	}
	if check.IfNil(args.ESDTTransferParser) {
		return nil, process.ErrNilESDTTransferParser
	}

	blockChainHookImpl, err := hooks.NewBlockChainHookImpl(args.ArgBlockChainHook)
	if err != nil {
		return nil, err
	}

	cryptoHook := hooks.NewVMCryptoHook()

	vmf := &vmContainerFactory{
		config:             args.Config,
		blockChainHookImpl: blockChainHookImpl,
		cryptoHook:         cryptoHook,
		blockGasLimit:      args.BlockGasLimit,
		gasSchedule:        args.GasSchedule,
		builtinFunctions:   args.ArgBlockChainHook.BuiltInFunctions,
		epochNotifier:      args.EpochNotifier,
		epochConfig:        args.EpochConfig,
		container:          nil,
		arwenChangeLocker:  args.ArwenChangeLocker,
		esdtTransferParser: args.ESDTTransferParser,
	}

	vmf.arwenVersions = args.Config.ArwenVersions
	vmf.sortArwenVersions()
	err = vmf.validateArwenVersions()
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

	return container, nil
}

// Close closes the vm container factory
func (vmf *vmContainerFactory) Close() error {
	return vmf.blockChainHookImpl.Close()
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

	vmf.gasSchedule.RegisterNotifyHandler(currentVM)
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
		ArwenV2EnableEpoch:       vmf.epochConfig.SCDeployEnableEpoch,
		AheadOfTimeEnableEpoch:   vmf.epochConfig.AheadOfTimeGasUsageEnableEpoch,
		DynGasLockEnableEpoch:    vmf.epochConfig.SCDeployEnableEpoch,
		ArwenV3EnableEpoch:       vmf.epochConfig.RepairCallbackEnableEpoch,
	}
	return arwenHost12.NewArwenVM(vmf.blockChainHookImpl, hostParameters)
}

func (vmf *vmContainerFactory) createInProcessArwenVMV13() (vmcommon.VMExecutionHandler, error) {
	hostParameters := &arwen13.VMHostParameters{
		VMType:                   factory.ArwenVirtualMachine,
		BlockGasLimit:            vmf.blockGasLimit,
		GasSchedule:              vmf.gasSchedule.LatestGasSchedule(),
		BuiltInFuncContainer:     vmf.builtinFunctions,
		ElrondProtectedKeyPrefix: []byte(core.ElrondProtectedKeyPrefix),
		ArwenV2EnableEpoch:       vmf.epochConfig.SCDeployEnableEpoch,
		AheadOfTimeEnableEpoch:   vmf.epochConfig.AheadOfTimeGasUsageEnableEpoch,
		DynGasLockEnableEpoch:    vmf.epochConfig.SCDeployEnableEpoch,
		ArwenV3EnableEpoch:       vmf.epochConfig.RepairCallbackEnableEpoch,
	}
	return arwenHost13.NewArwenVM(vmf.blockChainHookImpl, hostParameters)
}

func (vmf *vmContainerFactory) createInProcessArwenVMV14() (vmcommon.VMExecutionHandler, error) {
	hostParameters := &arwen14.VMHostParameters{
		VMType:                   factory.ArwenVirtualMachine,
		BlockGasLimit:            vmf.blockGasLimit,
		GasSchedule:              vmf.gasSchedule.LatestGasSchedule(),
		BuiltInFuncContainer:     vmf.builtinFunctions,
		ElrondProtectedKeyPrefix: []byte(core.ElrondProtectedKeyPrefix),
		ESDTTransferParser:       vmf.esdtTransferParser,
		EpochNotifier:            vmf.epochNotifier,
		MultiESDTTransferAsyncCallBackEnableEpoch: vmf.epochConfig.MultiESDTTransferFixOnCallBackOnEnableEpoch,
	}
	return arwenHost14.NewArwenVM(vmf.blockChainHookImpl, hostParameters)
}

func (vmf *vmContainerFactory) closePreviousVM(vm vmcommon.VMExecutionHandler) {
	vmf.blockChainHookImpl.ClearCompiledCodes()
	logVMContainerFactory.Debug("AOT compilation cache cleared")

	vmAsCloser, ok := vm.(io.Closer)
	if ok {
		err := vmAsCloser.Close()
		logVMContainerFactory.LogIfError(err, "closePreviousVM", "error", err)
	}
}

// BlockChainHookImpl returns the created blockChainHookImpl
func (vmf *vmContainerFactory) BlockChainHookImpl() process.BlockChainHookHandler {
	return vmf.blockChainHookImpl
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmf *vmContainerFactory) IsInterfaceNil() bool {
	return vmf == nil
}
