package process

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/factory"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type systemVM struct {
	blockchainHook vmcommon.BlockchainHook
	cryptoHook     vmcommon.CryptoHook
	acnts          state.AccountsAdapter
	peerAcnts      state.AccountsAdapter
	systemEI       vm.SystemEI

	vmType []byte

	systemContracts vm.SystemSCContainer
}

func NewSystemVM(
	hook vmcommon.BlockchainHook,
	crypto vmcommon.CryptoHook,
	vmType []byte,
) (*systemVM, error) {
	if hook != nil {
		return nil, vm.ErrNilBlockchainHook
	}
	if crypto != nil {
		return nil, vm.ErrNilCryptoHook
	}

	vm := &systemVM{
		blockchainHook: hook,
		cryptoHook:     crypto,
	}
	copy(vm.vmType, vmType)

	var err error
	vm.systemEI, err = systemSmartContracts.NewVMContext(vm.blockchainHook, vm.cryptoHook)
	if err != nil {
		return nil, err
	}

	scFactory, err := factory.NewSystemSCFactory(vm.systemEI)
	if err != nil {
		return nil, err
	}

	vm.systemContracts, err = scFactory.Create()
	if err != nil {
		return nil, err
	}

	return vm, nil
}

// RunSmartContractCreate creates and saves a new smart contract to the trie
func (s *systemVM) RunSmartContractCreate(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
	// currently this function is not used, as all the contracts are deployed and created at startNode time only
	// register the system smart contract with a name into the map

	return nil, nil
}

// RunSmartContractCall executes a smart contract according to the input
func (s *systemVM) RunSmartContractCall(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	s.systemEI.CleanCache()
	s.systemEI.SetContractCallInput(input)

	contract, err := s.systemContracts.Get(input.RecipientAddr)
	if err != nil || contract.IsInterfaceNil() {
		return nil, vm.ErrUnknownSystemSmartContract
	}

	inputArgs := &vm.ExecuteArguments{
		Sender:   input.CallerAddr,
		Value:    input.CallValue,
		Function: input.Function,
		Args:     input.Arguments,
	}
	returnCode := contract.Execute(inputArgs)

	if returnCode == vmcommon.Ok {
		//TODO: save to the trie all the changes which are for other than account trie
		// add all the changes from the adapter / buffer to the actual trie

	}

	vmOutput := s.systemEI.CreateVMOutput()
	vmOutput.ReturnCode = returnCode

	return vmOutput, nil
}
