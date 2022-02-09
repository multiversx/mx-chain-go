package process

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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
	if check.IfNil(args.GasSchedule) {
		return nil, vm.ErrNilGasSchedule
	}

	apiCosts := args.GasSchedule.LatestGasSchedule()[common.ElrondAPICost]
	if apiCosts == nil {
		return nil, vm.ErrNilGasSchedule
	}

	s := &systemVM{
		systemEI:             args.SystemEI,
		systemContracts:      args.SystemContracts,
		vmType:               make([]byte, len(args.VmType)),
		asyncCallStepCost:    apiCosts[common.AsyncCallStepField],
		asyncCallbackGasLock: apiCosts[common.AsyncCallbackGasLockField],
	}
	copy(s.vmType, args.VmType)

	if s.asyncCallStepCost == 0 || s.asyncCallbackGasLock == 0 {
		return nil, vm.ErrNilGasSchedule
	}

	args.GasSchedule.RegisterNotifyHandler(s)

	return s, nil
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

	contract, err := s.systemEI.GetContract(input.RecipientAddr)
	if err != nil {
		return nil, vm.ErrUnknownSystemSmartContract
	}

	if input.Function == core.SCDeployInitFunctionName {
		return &vmcommon.VMOutput{
			ReturnCode:    vmcommon.UserError,
			ReturnMessage: "cannot call smart contract init function",
		}, nil
	}

	returnCode := contract.Execute(input)
	vmOutput := s.systemEI.CreateVMOutput()
	vmOutput.ReturnCode = returnCode

	return vmOutput, nil
}

// GetVersion returns an empty string, because the system VM is not versioned
func (s *systemVM) GetVersion() string {
	return ""
}

// GasScheduleChange sets the new gas schedule where it is needed
func (s *systemVM) GasScheduleChange(gasSchedule map[string]map[string]uint64) {
	s.mutGasLock.Lock()
	defer s.mutGasLock.Unlock()

	apiCosts := gasSchedule[common.ElrondAPICost]
	if apiCosts == nil {
		return
	}

	s.asyncCallStepCost = apiCosts[common.AsyncCallStepField]
	s.asyncCallbackGasLock = apiCosts[common.AsyncCallbackGasLockField]
}

// Close does nothing, only to implement interface
func (s *systemVM) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *systemVM) IsInterfaceNil() bool {
	return s == nil
}
