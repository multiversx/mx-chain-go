package mock

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-vm-common"
)

var errNilValue = errors.New("nil value provided")

const addFunc = "add"
const getFunc = "get"

var variableA = []byte("a")

// OneSCExecutorMockVM contains one hardcoded SC with the following behaviour (written in golang):
//-------------------------------------
// var a int
//
// func init(initial int){
//     a = initial
// }
//
// func Add(value int){
//     a += value
// }
//
// func Get() int{
//     return a
// }
//-------------------------------------
type OneSCExecutorMockVM struct {
	blockchainHook vmcommon.BlockchainHook
	cryptoHook     vmcommon.CryptoHook
}

func NewOneSCExecutorMockVM(blockchainHook vmcommon.BlockchainHook, cryptoHook vmcommon.CryptoHook) (*OneSCExecutorMockVM, error) {
	if blockchainHook == nil || cryptoHook == nil {
		return nil, errNilValue
	}

	vm := &OneSCExecutorMockVM{
		blockchainHook: blockchainHook,
		cryptoHook:     cryptoHook,
	}

	return vm, nil
}

func (vm *OneSCExecutorMockVM) G0Create(input *vmcommon.ContractCreateInput) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (vm *OneSCExecutorMockVM) G0Call(input *vmcommon.ContractCallInput) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (vm *OneSCExecutorMockVM) RunSmartContractCreate(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
	if input == nil {
		return nil, errNilValue
	}
	if input.Arguments == nil {
		return nil, errNilValue
	}

	initialValue := big.NewInt(0)
	if input.Arguments[0] != nil {
		initialValue = input.Arguments[0]
	}

	senderNonce, err := vm.blockchainHook.GetNonce(input.CallerAddr)
	if err != nil {
		return nil, err
	}

	newSCAddr, err := vm.cryptoHook.Sha256(string(input.CallerAddr) + fmt.Sprintf("%d", senderNonce))
	if err != nil {
		return nil, err
	}

	scOutputAccount := &vmcommon.OutputAccount{
		Nonce:   big.NewInt(0),
		Code:    input.ContractCode,
		Balance: input.CallValue,
		Address: []byte(newSCAddr),
		StorageUpdates: []*vmcommon.StorageUpdate{
			{
				//only one variable: a
				Data:   variableA,
				Offset: initialValue.Bytes(),
			},
		},
	}

	return &vmcommon.VMOutput{
		OutputAccounts:  []*vmcommon.OutputAccount{scOutputAccount},
		Error:           false,
		DeletedAccounts: make([][]byte, 0),
		GasRefund:       big.NewInt(0),
		GasRemaining:    big.NewInt(0),
		Logs:            make([]*vmcommon.LogEntry, 0),
		ReturnCode:      vmcommon.Ok,
		ReturnData:      make([]*big.Int, 0),
		TouchedAccounts: make([][]byte, 0),
	}, nil
}

func (vm *OneSCExecutorMockVM) RunSmartContractCall(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	if input == nil {
		return nil, errNilValue
	}
	if input.Arguments == nil {
		return nil, errNilValue
	}

	//make a dummy call to get the SC code, not required here, but to emulate what happens in the real VM
	_, err := vm.blockchainHook.GetCode(input.RecipientAddr)
	if err != nil {
		return nil, err
	}

	value := big.NewInt(0)
	if input.Arguments[0] != nil {
		value = input.Arguments[0]
	}

	method := strings.ToLower(input.Function)
	switch method {
	case addFunc:
		return vm.processAddFunc(input, value)
	case getFunc:
		return vm.processGetFunc(input)
	default:
		return vm.unavailableFunc(input)
	}
}

func (vm *OneSCExecutorMockVM) processAddFunc(input *vmcommon.ContractCallInput, value *big.Int) (*vmcommon.VMOutput, error) {
	currentValueBuff, err := vm.blockchainHook.GetStorageData(input.RecipientAddr, variableA)
	if err != nil {
		return nil, err
	}

	currentValue := big.NewInt(0).SetBytes(currentValueBuff)
	newValue := big.NewInt(0).Add(currentValue, value)

	destNonce, err := vm.blockchainHook.GetNonce(input.RecipientAddr)
	if err != nil {
		return nil, err
	}

	destBalance, err := vm.blockchainHook.GetBalance(input.RecipientAddr)
	if err != nil {
		return nil, err
	}

	newBalance := big.NewInt(0).Add(destBalance, input.CallValue)
	scOutputAccount := &vmcommon.OutputAccount{
		Nonce:   destNonce,
		Balance: newBalance,
		Address: input.RecipientAddr,
		StorageUpdates: []*vmcommon.StorageUpdate{
			{
				//only one variable: a
				Data:   variableA,
				Offset: newValue.Bytes(),
			},
		},
	}

	return &vmcommon.VMOutput{
		OutputAccounts:  []*vmcommon.OutputAccount{scOutputAccount},
		Error:           false,
		DeletedAccounts: make([][]byte, 0),
		GasRefund:       big.NewInt(0),
		GasRemaining:    big.NewInt(0),
		Logs:            make([]*vmcommon.LogEntry, 0),
		ReturnCode:      vmcommon.Ok,
		ReturnData:      make([]*big.Int, 0),
		TouchedAccounts: make([][]byte, 0),
	}, nil
}

func (vm *OneSCExecutorMockVM) processGetFunc(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	currentValueBuff, err := vm.blockchainHook.GetStorageData(input.RecipientAddr, variableA)
	if err != nil {
		return nil, err
	}

	currentValue := big.NewInt(0).SetBytes(currentValueBuff)
	destNonce, err := vm.blockchainHook.GetNonce(input.RecipientAddr)
	if err != nil {
		return nil, err
	}

	destBalance, err := vm.blockchainHook.GetBalance(input.RecipientAddr)
	if err != nil {
		return nil, err
	}

	newBalance := big.NewInt(0).Add(destBalance, input.CallValue)
	scOutputAccount := &vmcommon.OutputAccount{
		Nonce:          destNonce,
		Balance:        newBalance,
		Address:        input.RecipientAddr,
		StorageUpdates: make([]*vmcommon.StorageUpdate, 0),
	}

	return &vmcommon.VMOutput{
		OutputAccounts:  []*vmcommon.OutputAccount{scOutputAccount},
		Error:           false,
		DeletedAccounts: make([][]byte, 0),
		GasRefund:       big.NewInt(0),
		GasRemaining:    big.NewInt(0),
		Logs:            make([]*vmcommon.LogEntry, 0),
		ReturnCode:      vmcommon.Ok,
		ReturnData:      []*big.Int{currentValue},
		TouchedAccounts: make([][]byte, 0),
	}, nil
}

func (vm *OneSCExecutorMockVM) unavailableFunc(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	destNonce, err := vm.blockchainHook.GetNonce(input.RecipientAddr)
	if err != nil {
		return nil, err
	}

	destBalance, err := vm.blockchainHook.GetBalance(input.RecipientAddr)
	if err != nil {
		return nil, err
	}

	scOutputAccount := &vmcommon.OutputAccount{
		Nonce:          destNonce,
		Balance:        destBalance,
		Address:        input.RecipientAddr,
		StorageUpdates: make([]*vmcommon.StorageUpdate, 0),
	}

	return &vmcommon.VMOutput{
		OutputAccounts:  []*vmcommon.OutputAccount{scOutputAccount},
		Error:           false,
		DeletedAccounts: make([][]byte, 0),
		GasRefund:       big.NewInt(0),
		GasRemaining:    big.NewInt(0),
		Logs:            make([]*vmcommon.LogEntry, 0),
		ReturnCode:      vmcommon.FunctionNotFound,
		ReturnData:      make([]*big.Int, 0),
		TouchedAccounts: make([][]byte, 0),
	}, nil
}
