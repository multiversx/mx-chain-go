package metachain

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/parsers"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/vm"
	systemVMFactory "github.com/ElrondNetwork/elrond-go/vm/factory"
	systemVMProcess "github.com/ElrondNetwork/elrond-go/vm/process"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
)

var _ process.VirtualMachinesContainerFactory = (*vmContainerFactory)(nil)

type vmContainerFactory struct {
	chanceComputer      sharding.ChanceComputer
	validatorAccountsDB state.AccountsAdapter
	blockChainHookImpl  *hooks.BlockChainHookImpl
	cryptoHook          vmcommon.CryptoHook
	systemContracts     vm.SystemSCContainer
	economics           *economics.EconomicsData
	messageSigVerifier  vm.MessageSignVerifier
	nodesConfigProvider vm.NodesConfigProvider
	gasSchedule         core.GasScheduleNotifier
	hasher              hashing.Hasher
	marshalizer         marshal.Marshalizer
	systemSCConfig      *config.SystemSmartContractsConfig
	epochNotifier       process.EpochNotifier
}

// ArgsNewVMContainerFactory defines the arguments needed to create a new VM container factory
type ArgsNewVMContainerFactory struct {
	ArgBlockChainHook   hooks.ArgBlockChainHook
	Economics           *economics.EconomicsData
	MessageSignVerifier vm.MessageSignVerifier
	GasSchedule         core.GasScheduleNotifier
	NodesConfigProvider vm.NodesConfigProvider
	Hasher              hashing.Hasher
	Marshalizer         marshal.Marshalizer
	SystemSCConfig      *config.SystemSmartContractsConfig
	ValidatorAccountsDB state.AccountsAdapter
	ChanceComputer      sharding.ChanceComputer
	EpochNotifier       process.EpochNotifier
}

// NewVMContainerFactory is responsible for creating a new virtual machine factory object
func NewVMContainerFactory(args ArgsNewVMContainerFactory) (*vmContainerFactory, error) {
	if args.Economics == nil {
		return nil, process.ErrNilEconomicsData
	}
	if check.IfNil(args.MessageSignVerifier) {
		return nil, process.ErrNilKeyGen
	}
	if check.IfNil(args.NodesConfigProvider) {
		return nil, process.ErrNilNodesConfigProvider
	}
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if args.SystemSCConfig == nil {
		return nil, process.ErrNilSystemSCConfig
	}
	if check.IfNil(args.ValidatorAccountsDB) {
		return nil, vm.ErrNilValidatorAccountsDB
	}
	if check.IfNil(args.ChanceComputer) {
		return nil, vm.ErrNilChanceComputer
	}
	if check.IfNil(gasSchedule) {
		return nil, vm.ErrNilGasSchedule
	}

	blockChainHookImpl, err := hooks.NewBlockChainHookImpl(args.ArgBlockChainHook)
	if err != nil {
		return nil, err
	}
	cryptoHook := hooks.NewVMCryptoHook()

	return &vmContainerFactory{
		blockChainHookImpl:  blockChainHookImpl,
		cryptoHook:          cryptoHook,
		economics:           args.Economics,
		messageSigVerifier:  args.MessageSignVerifier,
		gasSchedule:         args.GasSchedule,
		nodesConfigProvider: args.NodesConfigProvider,
		hasher:              args.Hasher,
		marshalizer:         args.Marshalizer,
		systemSCConfig:      args.SystemSCConfig,
		validatorAccountsDB: args.ValidatorAccountsDB,
		chanceComputer:      args.ChanceComputer,
		epochNotifier:       args.EpochNotifier,
	}, nil
}

// Create sets up all the needed virtual machine returning a container of all the VMs
func (vmf *vmContainerFactory) Create() (process.VirtualMachinesContainer, error) {
	container := containers.NewVirtualMachinesContainer()

	currVm, err := vmf.createSystemVM()
	if err != nil {
		return nil, err
	}

	err = container.Add(factory.SystemVirtualMachine, currVm)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (vmf *vmContainerFactory) createSystemVM() (vmcommon.VMExecutionHandler, error) {
	atArgumentParser := parsers.NewCallArgsParser()
	systemEI, err := systemSmartContracts.NewVMContext(
		vmf.blockChainHookImpl,
		vmf.cryptoHook,
		atArgumentParser,
		vmf.validatorAccountsDB,
		vmf.chanceComputer,
	)
	if err != nil {
		return nil, err
	}

	argsNewSystemScFactory := systemVMFactory.ArgsNewSystemSCFactory{
		SystemEI:            systemEI,
		SigVerifier:         vmf.messageSigVerifier,
		GasSchedule:         vmf.gasSchedule,
		NodesConfigProvider: vmf.nodesConfigProvider,
		Hasher:              vmf.hasher,
		Marshalizer:         vmf.marshalizer,
		SystemSCConfig:      vmf.systemSCConfig,
		Economics:           vmf.economics,
		EpochNotifier:       vmf.epochNotifier,
	}
	scFactory, err := systemVMFactory.NewSystemSCFactory(argsNewSystemScFactory)
	if err != nil {
		return nil, err
	}

	vmf.systemContracts, err = scFactory.Create()
	if err != nil {
		return nil, err
	}

	err = systemEI.SetSystemSCContainer(vmf.systemContracts)
	if err != nil {
		return nil, err
	}

	argsNewSystemVM := systemVMProcess.ArgsNewSystemVM{
		SystemEI:        systemEI,
		SystemContracts: vmf.systemContracts,
		VmType:          factory.SystemVirtualMachine,
		GasSchedule:     vmf.gasSchedule,
	}
	systemVM, err := systemVMProcess.NewSystemVM(argsNewSystemVM)
	if err != nil {
		return nil, err
	}

	vmf.gasSchedule.RegisterNotifyHandler(systemVM)

	return systemVM, nil
}

// BlockChainHookImpl returns the created blockChainHookImpl
func (vmf *vmContainerFactory) BlockChainHookImpl() process.BlockChainHookHandler {
	return vmf.blockChainHookImpl
}

// SystemSmartContractContainer return the created system smart contracts
func (vmf *vmContainerFactory) SystemSmartContractContainer() vm.SystemSCContainer {
	return vmf.systemContracts
}

// IsInterfaceNil returns true if there is no value under the interface
func (vmf *vmContainerFactory) IsInterfaceNil() bool {
	return vmf == nil
}
