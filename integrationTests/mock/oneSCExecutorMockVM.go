package mock

import (
	"errors"
	"math/big"
	"strings"

	vmcommon "github.com/ElrondNetwork/elrond-go/core/vm-common"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
)

var errNilValue = errors.New("nil value provided")
var errNoArgumentsProvided = errors.New("no arguments provided")
var errInsuficientFunds = errors.New("insufucient funds")

const addFunc = "add"
const withdrawFunc = "withdraw"
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
	blockchainHook  vmcommon.BlockchainHook
	hasher          hashing.Hasher
	GasForOperation uint64
}

// NewOneSCExecutorMockVM -
func NewOneSCExecutorMockVM(blockchainHook vmcommon.BlockchainHook, hasher hashing.Hasher) (*OneSCExecutorMockVM, error) {
	if blockchainHook == nil || hasher == nil {
		return nil, errNilValue
	}

	vm := &OneSCExecutorMockVM{
		blockchainHook: blockchainHook,
		hasher:         hasher,
	}

	return vm, nil
}

// RunSmartContractCreate -
func (vm *OneSCExecutorMockVM) RunSmartContractCreate(input *vmcommon.ContractCreateInput) (*vmcommon.VMOutput, error) {
	if input == nil {
		return nil, errNilValue
	}
	if len(input.Arguments) == 0 {
		return nil, errNoArgumentsProvided
	}
	if input.GasProvided < vm.GasForOperation {
		return vm.outOfGasFunc(&input.VMInput)
	}

	initialValue := make([]byte, 0)
	if input.Arguments[0] != nil {
		initialValue = append(initialValue, input.Arguments[0]...)
	}

	senderACcount, err := vm.blockchainHook.GetUserAccount(input.CallerAddr)
	if err != nil {
		return nil, err
	}

	newSCAddr, err := vm.blockchainHook.NewAddress(input.CallerAddr, senderACcount.GetNonce(), factory.InternalTestingVM)
	if err != nil {
		return nil, err
	}

	scOutputAccount := &vmcommon.OutputAccount{
		Nonce:        0,
		CodeMetadata: input.ContractCodeMetadata,
		Code:         input.ContractCode,
		BalanceDelta: input.CallValue,
		Address:      newSCAddr,
		StorageUpdates: makeStorageUpdatesMap(&vmcommon.StorageUpdate{
			//only one variable: a
			Offset: variableA,
			Data:   initialValue,
		}),
	}

	senderOutputAccount := &vmcommon.OutputAccount{
		Address: input.CallerAddr,
		//VM does not increment sender's nonce
		Nonce: senderACcount.GetNonce(),
		//tx succeed, return 0 back to the sender
		BalanceDelta: big.NewInt(0),
	}

	return &vmcommon.VMOutput{
		OutputAccounts:  makeOutputAccountsMap(scOutputAccount, senderOutputAccount),
		DeletedAccounts: make([][]byte, 0),
		GasRefund:       big.NewInt(0),
		GasRemaining:    input.GasProvided - vm.GasForOperation,
		Logs:            make([]*vmcommon.LogEntry, 0),
		ReturnCode:      vmcommon.Ok,
		ReturnData:      [][]byte{},
		TouchedAccounts: make([][]byte, 0),
	}, nil
}

// RunSmartContractCall -
func (vm *OneSCExecutorMockVM) RunSmartContractCall(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	if input == nil {
		return nil, errNilValue
	}
	if input.Arguments == nil {
		return nil, errNilValue
	}
	if input.GasProvided < vm.GasForOperation {
		return vm.outOfGasFunc(&input.VMInput)
	}

	//make a dummy call to get the SC code, not required here, but to emulate what happens in the real VM
	recipientAccount, err := vm.blockchainHook.GetUserAccount(input.RecipientAddr)
	if err != nil {
		return nil, err
	}

	if len(recipientAccount.GetCode()) == 0 {
		return nil, hooks.ErrEmptyCode
	}

	value := big.NewInt(0)
	if len(input.Arguments) > 0 {
		if input.Arguments[0] != nil {
			value.SetBytes(input.Arguments[0])
		}
	}

	method := strings.ToLower(input.Function)
	switch method {
	case addFunc:
		return vm.processAddFunc(input, value)
	case withdrawFunc:
		return vm.processWithdrawFunc(input, value)
	case getFunc:
		return vm.processGetFunc(input)
	default:
		return vm.unavailableFunc(input)
	}
}

// GasScheduleChange -
func (vm *OneSCExecutorMockVM) GasScheduleChange(_ map[string]map[string]uint64) {
}

func (vm *OneSCExecutorMockVM) processAddFunc(input *vmcommon.ContractCallInput, value *big.Int) (*vmcommon.VMOutput, error) {
	currentValueBuff, err := vm.blockchainHook.GetStorageData(input.RecipientAddr, variableA)
	if err != nil {
		return nil, err
	}

	currentValue := big.NewInt(0).SetBytes(currentValueBuff)
	newValue := big.NewInt(0).Add(currentValue, value)

	recipientAccount, err := vm.blockchainHook.GetUserAccount(input.RecipientAddr)
	if err != nil {
		return nil, err
	}

	scOutputAccount := &vmcommon.OutputAccount{
		Nonce:        recipientAccount.GetNonce(),
		BalanceDelta: input.CallValue,
		Address:      input.RecipientAddr,
		StorageUpdates: makeStorageUpdatesMap(&vmcommon.StorageUpdate{
			//only one variable: a
			Offset: variableA,
			Data:   newValue.Bytes(),
		}),
	}

	senderAccount, err := vm.blockchainHook.GetUserAccount(input.CallerAddr)
	if err != nil {
		return nil, err
	}

	senderOutputAccount := &vmcommon.OutputAccount{
		Address: input.CallerAddr,
		//VM does not increment sender's nonce
		Nonce: senderAccount.GetNonce(),
		//tx succeed, return 0 back to the sender
		BalanceDelta: big.NewInt(0),
	}

	return &vmcommon.VMOutput{
		OutputAccounts:  makeOutputAccountsMap(scOutputAccount, senderOutputAccount),
		DeletedAccounts: make([][]byte, 0),
		GasRefund:       big.NewInt(0),
		GasRemaining:    input.GasProvided - vm.GasForOperation,
		Logs:            make([]*vmcommon.LogEntry, 0),
		ReturnCode:      vmcommon.Ok,
		ReturnData:      [][]byte{},
		TouchedAccounts: make([][]byte, 0),
	}, nil
}

func (vm *OneSCExecutorMockVM) processWithdrawFunc(input *vmcommon.ContractCallInput, value *big.Int) (*vmcommon.VMOutput, error) {

	destAccount, err := vm.blockchainHook.GetUserAccount(input.RecipientAddr)
	if err != nil {
		return nil, err
	}

	destBalance := destAccount.GetBalance()
	if destBalance.Cmp(value) < 0 {
		return nil, errInsuficientFunds
	}

	newSCBalance := big.NewInt(0).Sub(big.NewInt(0), value)
	scOutputAccount := &vmcommon.OutputAccount{
		Nonce:        destAccount.GetNonce(),
		BalanceDelta: newSCBalance,
		Address:      input.RecipientAddr,
	}

	senderAccount, err := vm.blockchainHook.GetUserAccount(input.CallerAddr)
	if err != nil {
		return nil, err
	}

	senderOutputAccount := &vmcommon.OutputAccount{
		Address: input.CallerAddr,
		//VM does not increment sender's nonce
		Nonce: senderAccount.GetNonce(),
		//tx succeed, return 0 back to the sender
		BalanceDelta: value,
	}

	return &vmcommon.VMOutput{
		OutputAccounts:  makeOutputAccountsMap(scOutputAccount, senderOutputAccount),
		DeletedAccounts: make([][]byte, 0),
		GasRefund:       big.NewInt(0),
		GasRemaining:    input.GasProvided - vm.GasForOperation,
		Logs:            make([]*vmcommon.LogEntry, 0),
		ReturnCode:      vmcommon.Ok,
		ReturnData:      [][]byte{},
		TouchedAccounts: make([][]byte, 0),
	}, nil
}

func (vm *OneSCExecutorMockVM) processGetFunc(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	currentValueBuff, err := vm.blockchainHook.GetStorageData(input.RecipientAddr, variableA)
	if err != nil {
		return nil, err
	}

	destAccount, err := vm.blockchainHook.GetUserAccount(input.RecipientAddr)
	if err != nil {
		return nil, err
	}

	scOutputAccount := &vmcommon.OutputAccount{
		Nonce:          destAccount.GetNonce(),
		BalanceDelta:   input.CallValue,
		Address:        input.RecipientAddr,
		StorageUpdates: make(map[string]*vmcommon.StorageUpdate),
	}

	return &vmcommon.VMOutput{
		OutputAccounts:  makeOutputAccountsMap(scOutputAccount),
		DeletedAccounts: make([][]byte, 0),
		GasRefund:       big.NewInt(0),
		GasRemaining:    input.GasProvided - vm.GasForOperation,
		Logs:            make([]*vmcommon.LogEntry, 0),
		ReturnCode:      vmcommon.Ok,
		ReturnData:      [][]byte{currentValueBuff},
		TouchedAccounts: make([][]byte, 0),
	}, nil
}

func (vm *OneSCExecutorMockVM) unavailableFunc(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
	destAccount, err := vm.blockchainHook.GetUserAccount(input.RecipientAddr)
	if err != nil {
		return nil, err
	}

	scOutputAccount := &vmcommon.OutputAccount{
		Nonce:          destAccount.GetNonce(),
		Address:        input.RecipientAddr,
		StorageUpdates: make(map[string]*vmcommon.StorageUpdate),
	}

	return &vmcommon.VMOutput{
		OutputAccounts:  makeOutputAccountsMap(scOutputAccount),
		DeletedAccounts: make([][]byte, 0),
		GasRefund:       big.NewInt(0),
		GasRemaining:    0,
		Logs:            make([]*vmcommon.LogEntry, 0),
		ReturnCode:      vmcommon.FunctionNotFound,
		ReturnData:      [][]byte{},
		TouchedAccounts: make([][]byte, 0),
	}, nil
}

func (vm *OneSCExecutorMockVM) outOfGasFunc(input *vmcommon.VMInput) (*vmcommon.VMOutput, error) {
	callerAccount, err := vm.blockchainHook.GetUserAccount(input.CallerAddr)
	if err != nil {
		return nil, err
	}

	vmo := &vmcommon.OutputAccount{
		BalanceDelta: big.NewInt(0),
		Address:      input.CallerAddr,
		Nonce:        callerAccount.GetNonce(),
	}

	return &vmcommon.VMOutput{
		OutputAccounts:  makeOutputAccountsMap(vmo),
		DeletedAccounts: make([][]byte, 0),
		GasRefund:       big.NewInt(0),
		GasRemaining:    0,
		Logs:            make([]*vmcommon.LogEntry, 0),
		ReturnCode:      vmcommon.OutOfGas,
		ReturnData:      [][]byte{},
		TouchedAccounts: make([][]byte, 0),
	}, nil
}

func makeStorageUpdatesMap(updates ...*vmcommon.StorageUpdate) map[string]*vmcommon.StorageUpdate {
	if len(updates) == 0 {
		return make(map[string]*vmcommon.StorageUpdate)
	}

	updatesMap := make(map[string]*vmcommon.StorageUpdate, len(updates))
	for _, update := range updates {
		updatesMap[string(update.Offset)] = update
	}
	return updatesMap
}

func makeOutputAccountsMap(accounts ...*vmcommon.OutputAccount) map[string]*vmcommon.OutputAccount {
	if accounts == nil {
		return make(map[string]*vmcommon.OutputAccount)
	}

	accountsMap := make(map[string]*vmcommon.OutputAccount, len(accounts))
	for _, account := range accounts {
		accountsMap[string(account.Address)] = account
	}
	return accountsMap
}
