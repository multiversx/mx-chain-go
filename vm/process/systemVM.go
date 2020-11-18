package process

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/vm"
)

type systemVM struct {
	systemEI             vm.ContextHandler
	vmType               []byte
	systemContracts      vm.SystemSCContainer
	asyncCallbackGasLock uint64
	asyncCallStepCost    uint64
	mutGasLock           sync.RWMutex
}

// ArgsNewSystemVM defines the needed arguments to create a new system vm
type ArgsNewSystemVM struct {
	SystemEI        vm.ContextHandler
	SystemContracts vm.SystemSCContainer
	VmType          []byte
	GasSchedule     core.GasScheduleNotifier
}

// NewSystemVM instantiates the system VM which is capable of running in protocol smart contracts
func NewSystemVM(args ArgsNewSystemVM) (*systemVM, error) {
	if check.IfNil(args.SystemEI) {
		return nil, vm.ErrNilSystemEnvironmentInterface
	}
	if check.IfNil(args.SystemContracts) {
		return nil, vm.ErrNilSystemContractsContainer
	}
	if len(args.VmType) == 0 { // no need for nil check, len() for nil returns 0
		return nil, vm.ErrNilVMType
	}
	if args.GasSchedule == nil {
		return nil, vm.ErrNilGasSchedule
	}

	apiCosts := args.GasSchedule.LatestGasSchedule()[core.ElrondAPICost]
	if apiCosts == nil {
		return nil, vm.ErrNilGasSchedule
	}

	sVm := &systemVM{
		systemEI:             args.SystemEI,
		systemContracts:      args.SystemContracts,
		vmType:               make([]byte, len(args.VmType)),
		asyncCallStepCost:    apiCosts[core.AsyncCallStepField],
		asyncCallbackGasLock: apiCosts[core.AsyncCallbackGasLockField],
	}
	copy(sVm.vmType, args.VmType)

	if sVm.asyncCallStepCost == 0 || sVm.asyncCallbackGasLock == 0 {
		return nil, vm.ErrNilGasSchedule
	}

	args.GasSchedule.RegisterNotifyHandler(sVm)

	return sVm, nil
}

// RunSmartContractCreate creates and saves a new smart contract to the trie
func (s *systemVM) RunSmartContractCreate(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
	if input == nil {
		return nil, vm.ErrInputArgsIsNil
	}
	if input.CallerAddr == nil {
		return nil, vm.ErrInputCallerAddrIsNil
	}

	s.systemEI.CleanCache()
	s.systemEI.SetSCAddress(input.CallerAddr)

	contract, err := s.systemContracts.Get(input.CallerAddr)
	if err != nil {
		return nil, vm.ErrUnknownSystemSmartContract
	}

	deployInput := &vmcommon.ContractCallInput{
		VMInput:       input.VMInput,
		RecipientAddr: input.CallerAddr,
		Function:      core.SCDeployInitFunctionName,
	}

	returnCode := contract.Execute(deployInput)

	s.systemEI.AddCode(input.CallerAddr, input.ContractCode)

	vmOutput := s.systemEI.CreateVMOutput()
	vmOutput.ReturnCode = returnCode

	return vmOutput, nil
}

// RunSmartContractCall executes a smart contract according to the input
func (s *systemVM) RunSmartContractCall(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	s.systemEI.CleanCache()
	s.systemEI.SetSCAddress(input.RecipientAddr)
	s.systemEI.AddTxValueToSmartContract(input.CallValue, input.RecipientAddr)
	s.systemEI.SetGasProvided(input.GasProvided)

	contract, err := s.systemContracts.Get(input.RecipientAddr)
	if err != nil {
		return nil, vm.ErrUnknownSystemSmartContract
	}

	if input.Function == core.SCDeployInitFunctionName {
		return &vmcommon.VMOutput{
			ReturnCode:    vmcommon.UserError,
			ReturnMessage: "cannot call smart contract init function",
		}, nil
	}

	lockedGas := input.GasLocked
	input.GasProvided -= lockedGas
	returnCode := contract.Execute(input)

	vmOutput := s.systemEI.CreateVMOutput()
	vmOutput.ReturnCode = returnCode
	vmOutput.GasRemaining += lockedGas

	return vmOutput, nil
}

// GasScheduleChange sets the new gas schedule where it is needed
func (s *systemVM) GasScheduleChange(gasSchedule map[string]map[string]uint64) {
	s.mutGasLock.Lock()
	defer s.mutGasLock.Unlock()

	apiCosts := gasSchedule[core.ElrondAPICost]
	if apiCosts == nil {
		return
	}

	s.asyncCallStepCost = apiCosts[core.AsyncCallStepField]
	s.asyncCallbackGasLock = apiCosts[core.AsyncCallbackGasLockField]
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *systemVM) IsInterfaceNil() bool {
	return s == nil
}
