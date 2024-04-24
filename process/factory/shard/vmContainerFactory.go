package shard

import (
	"fmt"
	"io"
	"sort"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/factory/containers"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
	wasmVMHost15 "github.com/multiversx/mx-chain-vm-go/vmhost/hostCore"
	wasmvm12 "github.com/multiversx/mx-chain-vm-v1_2-go/vmhost"
	wasmVMHost12 "github.com/multiversx/mx-chain-vm-v1_2-go/vmhost/hostCore"
	wasmvm13 "github.com/multiversx/mx-chain-vm-v1_3-go/vmhost"
	wasmVMHost13 "github.com/multiversx/mx-chain-vm-v1_3-go/vmhost/hostCore"
	wasmvm14 "github.com/multiversx/mx-chain-vm-v1_4-go/vmhost"
	wasmVMHost14 "github.com/multiversx/mx-chain-vm-v1_4-go/vmhost/hostCore"
)

var _ process.VirtualMachinesContainerFactory = (*vmContainerFactory)(nil)

var logVMContainerFactory = logger.GetOrCreate("vmContainerFactory")

type vmContainerFactory struct {
	config              config.VirtualMachineConfig
	blockChainHook      process.BlockChainHookWithAccountsAdapter
	cryptoHook          vmcommon.CryptoHook
	blockGasLimit       uint64
	gasSchedule         core.GasScheduleNotifier
	builtinFunctions    vmcommon.BuiltInFunctionContainer
	epochNotifier       process.EpochNotifier
	enableEpochsHandler common.EnableEpochsHandler
	container           process.VirtualMachinesContainer
	wasmVMVersions      []config.WasmVMVersionByEpoch
	wasmVMChangeLocker  common.Locker
	esdtTransferParser  vmcommon.ESDTTransferParser
	hasher              hashing.Hasher
	pubKeyConverter     core.PubkeyConverter

	mapOpcodeAddressIsAllowed map[string]map[string]struct{}
}

const managedMultiTransferESDTNFTExecuteByUser = "managedMultiTransferESDTNFTExecuteByUser"

// ArgVMContainerFactory defines the arguments needed to the new VM factory
type ArgVMContainerFactory struct {
	Config              config.VirtualMachineConfig
	BlockGasLimit       uint64
	GasSchedule         core.GasScheduleNotifier
	EpochNotifier       process.EpochNotifier
	EnableEpochsHandler common.EnableEpochsHandler
	WasmVMChangeLocker  common.Locker
	ESDTTransferParser  vmcommon.ESDTTransferParser
	BuiltInFunctions    vmcommon.BuiltInFunctionContainer
	BlockChainHook      process.BlockChainHookWithAccountsAdapter
	Hasher              hashing.Hasher
	PubKeyConverter     core.PubkeyConverter
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
	if check.IfNilReflect(args.WasmVMChangeLocker) {
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
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.PubKeyConverter) {
		return nil, process.ErrNilPubkeyConverter
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
		wasmVMChangeLocker:  args.WasmVMChangeLocker,
		esdtTransferParser:  args.ESDTTransferParser,
		hasher:              args.Hasher,
		pubKeyConverter:     args.PubKeyConverter,
	}

	vmf.wasmVMVersions = args.Config.WasmVMVersions
	vmf.sortWasmVMVersions()
	err := vmf.validateWasmVMVersions()
	if err != nil {
		return nil, err
	}

	err = vmf.createMapOpCodeAddressIsAllowed()
	if err != nil {
		return nil, err
	}

	return vmf, nil
}

func (vmf *vmContainerFactory) createMapOpCodeAddressIsAllowed() error {
	vmf.mapOpcodeAddressIsAllowed = make(map[string]map[string]struct{})

	transferAndExecuteByUserAddresses := vmf.config.TransferAndExecuteByUserAddresses
	if len(transferAndExecuteByUserAddresses) == 0 {
		return process.ErrTransferAndExecuteByUserAddressesIsNil
	}

	vmf.mapOpcodeAddressIsAllowed[managedMultiTransferESDTNFTExecuteByUser] = make(map[string]struct{})
	for _, address := range transferAndExecuteByUserAddresses {
		decodedAddress, errDecode := vmf.pubKeyConverter.Decode(address)
		if errDecode != nil {
			return errDecode
		}
		vmf.mapOpcodeAddressIsAllowed[managedMultiTransferESDTNFTExecuteByUser][string(decodedAddress)] = struct{}{}
	}

	return nil
}

func (vmf *vmContainerFactory) sortWasmVMVersions() {
	sort.Slice(vmf.wasmVMVersions, func(i, j int) bool {
		return vmf.wasmVMVersions[i].StartEpoch < vmf.wasmVMVersions[j].StartEpoch
	})
}

func (vmf *vmContainerFactory) validateWasmVMVersions() error {
	if len(vmf.wasmVMVersions) == 0 {
		return ErrEmptyVersionsByEpochsList
	}

	currentEpoch := uint32(0)
	for idx, ver := range vmf.wasmVMVersions {
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

	vmf.wasmVMChangeLocker.Lock()
	version := vmf.getMatchingVersion(vmf.epochNotifier.CurrentEpoch())
	currentVM, err := vmf.createWasmVM(version)
	if err != nil {
		vmf.wasmVMChangeLocker.Unlock()
		return nil, err
	}

	err = container.Add(factory.WasmVirtualMachine, currentVM)
	if err != nil {
		vmf.wasmVMChangeLocker.Unlock()
		return nil, err
	}

	// The vmContainerFactory keeps a reference to the container it has created,
	// in order to replace, from within the container, the VM instances that
	// become out-of-date after specific epochs.
	vmf.container = container
	vmf.wasmVMChangeLocker.Unlock()

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
	vmf.ensureCorrectWasmVMVersion(epoch)
}

func (vmf *vmContainerFactory) ensureCorrectWasmVMVersion(epoch uint32) {
	newVersion := vmf.getMatchingVersion(epoch)
	currentWasmVM, err := vmf.container.Get(factory.WasmVirtualMachine)
	if err != nil {
		logVMContainerFactory.Error("cannot retrieve Wasm VM from container", "epoch", epoch, "error", err)
		return
	}

	if !vmf.shouldReplaceWasmVMInstance(newVersion, currentWasmVM) {
		return
	}

	vmf.wasmVMChangeLocker.Lock()
	defer vmf.wasmVMChangeLocker.Unlock()

	vmf.closePreviousVM(currentWasmVM)
	newWasmVM, err := vmf.createWasmVM(newVersion)
	if err != nil {
		logVMContainerFactory.Error("cannot replace Wasm VM", "epoch", epoch, "error", err)
		return
	}

	err = vmf.container.Replace(factory.WasmVirtualMachine, newWasmVM)
	if err != nil {
		logVMContainerFactory.Error("cannot replace Wasm VM", "epoch", epoch, "error", err)
		return
	}

	logVMContainerFactory.Debug("Wasm VM replaced", "epoch", epoch)
}

func (vmf *vmContainerFactory) shouldReplaceWasmVMInstance(
	newVersion config.WasmVMVersionByEpoch,
	currentVM vmcommon.VMExecutionHandler,
) bool {
	specificVersionRequired := newVersion.Version != "*"
	differentVersion := newVersion.Version != currentVM.GetVersion()

	return specificVersionRequired && differentVersion
}

func (vmf *vmContainerFactory) createWasmVM(version config.WasmVMVersionByEpoch) (vmcommon.VMExecutionHandler, error) {
	currentVM, err := vmf.createInProcessWasmVMByVersion(version)
	if err != nil {
		return nil, err
	}

	return currentVM, nil
}

func (vmf *vmContainerFactory) getMatchingVersion(epoch uint32) config.WasmVMVersionByEpoch {
	matchingVersion := vmf.wasmVMVersions[len(vmf.wasmVMVersions)-1]
	for idx := 0; idx < len(vmf.wasmVMVersions)-1; idx++ {
		crtVer := vmf.wasmVMVersions[idx]
		nextVer := vmf.wasmVMVersions[idx+1]
		if crtVer.StartEpoch <= epoch && epoch < nextVer.StartEpoch {
			return crtVer
		}
	}

	return matchingVersion
}

func (vmf *vmContainerFactory) createInProcessWasmVMByVersion(version config.WasmVMVersionByEpoch) (vmcommon.VMExecutionHandler, error) {
	logVMContainerFactory.Debug("createInProcessWasmVMByVersion", "version", version)
	switch version.Version {
	case "v1.2":
		return vmf.createInProcessWasmVMV12()
	case "v1.3":
		return vmf.createInProcessWasmVMV13()
	case "v1.4":
		return vmf.createInProcessWasmVMV14()
	default:
		return vmf.createInProcessWasmVMV15()
	}
}

func (vmf *vmContainerFactory) createInProcessWasmVMV12() (vmcommon.VMExecutionHandler, error) {
	logVMContainerFactory.Info("VM 1.2 created")
	hostParameters := &wasmvm12.VMHostParameters{
		VMType:                   factory.WasmVirtualMachine,
		BlockGasLimit:            vmf.blockGasLimit,
		GasSchedule:              vmf.gasSchedule.LatestGasSchedule(),
		ProtocolBuiltinFunctions: vmf.builtinFunctions.Keys(),
		ProtectedKeyPrefix:       []byte(core.ProtectedKeyPrefix),
		EnableEpochsHandler:      vmf.enableEpochsHandler,
	}
	return wasmVMHost12.NewVMHost(vmf.blockChainHook, hostParameters)
}

func (vmf *vmContainerFactory) createInProcessWasmVMV13() (vmcommon.VMExecutionHandler, error) {
	logVMContainerFactory.Info("VM 1.3 created")
	hostParameters := &wasmvm13.VMHostParameters{
		VMType:               factory.WasmVirtualMachine,
		BlockGasLimit:        vmf.blockGasLimit,
		GasSchedule:          vmf.gasSchedule.LatestGasSchedule(),
		BuiltInFuncContainer: vmf.builtinFunctions,
		ProtectedKeyPrefix:   []byte(core.ProtectedKeyPrefix),
		EnableEpochsHandler:  vmf.enableEpochsHandler,
	}
	return wasmVMHost13.NewVMHost(vmf.blockChainHook, hostParameters)
}

func (vmf *vmContainerFactory) createInProcessWasmVMV14() (vmcommon.VMExecutionHandler, error) {
	logVMContainerFactory.Info("VM 1.4 created")
	hostParameters := &wasmvm14.VMHostParameters{
		VMType:                              factory.WasmVirtualMachine,
		BlockGasLimit:                       vmf.blockGasLimit,
		GasSchedule:                         vmf.gasSchedule.LatestGasSchedule(),
		BuiltInFuncContainer:                vmf.builtinFunctions,
		ProtectedKeyPrefix:                  []byte(core.ProtectedKeyPrefix),
		ESDTTransferParser:                  vmf.esdtTransferParser,
		WasmerSIGSEGVPassthrough:            vmf.config.WasmerSIGSEGVPassthrough,
		TimeOutForSCExecutionInMilliseconds: vmf.config.TimeOutForSCExecutionInMilliseconds,
		EpochNotifier:                       vmf.epochNotifier,
		EnableEpochsHandler:                 vmf.enableEpochsHandler,
		Hasher:                              vmf.hasher,
	}
	return wasmVMHost14.NewVMHost(vmf.blockChainHook, hostParameters)
}

func (vmf *vmContainerFactory) createInProcessWasmVMV15() (vmcommon.VMExecutionHandler, error) {
	logVMContainerFactory.Info("VM 1.5 created")
	hostParameters := &vmhost.VMHostParameters{
		VMType:                              factory.WasmVirtualMachine,
		BlockGasLimit:                       vmf.blockGasLimit,
		GasSchedule:                         vmf.gasSchedule.LatestGasSchedule(),
		BuiltInFuncContainer:                vmf.builtinFunctions,
		ProtectedKeyPrefix:                  []byte(core.ProtectedKeyPrefix),
		ESDTTransferParser:                  vmf.esdtTransferParser,
		WasmerSIGSEGVPassthrough:            vmf.config.WasmerSIGSEGVPassthrough,
		TimeOutForSCExecutionInMilliseconds: vmf.config.TimeOutForSCExecutionInMilliseconds,
		EpochNotifier:                       vmf.epochNotifier,
		EnableEpochsHandler:                 vmf.enableEpochsHandler,
		Hasher:                              vmf.hasher,
		MapOpcodeAddressIsAllowed:           vmf.mapOpcodeAddressIsAllowed,
	}

	return wasmVMHost15.NewVMHost(vmf.blockChainHook, hostParameters)
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
func (vmf *vmContainerFactory) BlockChainHookImpl() process.BlockChainHookWithAccountsAdapter {
	return vmf.blockChainHook
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmf *vmContainerFactory) IsInterfaceNil() bool {
	return vmf == nil
}
