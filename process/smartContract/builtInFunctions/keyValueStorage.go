package builtInFunctions

import (
	"bytes"
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.BuiltinFunction = (*saveKeyValueStorage)(nil)

type saveKeyValueStorage struct {
	gasConfig    process.BaseOperationCost
	funcGasCost  uint64
	mutExecution sync.RWMutex
}

// NewSaveKeyValueStorageFunc returns the save key-value storage built in function
func NewSaveKeyValueStorageFunc(
	gasConfig process.BaseOperationCost,
	funcGasCost uint64,
) (*saveKeyValueStorage, error) {
	s := &saveKeyValueStorage{
		gasConfig:   gasConfig,
		funcGasCost: funcGasCost,
	}

	return s, nil
}

// SetNewGasConfig is called whenever gas cost is changed
func (k *saveKeyValueStorage) SetNewGasConfig(gasCost *process.GasCost) {
	if gasCost == nil {
		return
	}

	k.mutExecution.Lock()
	k.funcGasCost = gasCost.BuiltInCost.SaveKeyValue
	k.gasConfig = gasCost.BaseOperationCost
	k.mutExecution.Unlock()
}

// ProcessBuiltinFunction will save the value for the selected key
func (k *saveKeyValueStorage) ProcessBuiltinFunction(
	_, acntDst state.UserAccountHandler,
	input *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	k.mutExecution.RLock()
	defer k.mutExecution.RUnlock()

	err := checkArgumentsForSaveKeyValue(acntDst, input)
	if err != nil {
		return nil, err
	}

	vmOutput := &vmcommon.VMOutput{
		GasRemaining: input.GasProvided,
		GasRefund:    big.NewInt(0),
	}

	useGas := k.funcGasCost
	for i := 0; i < len(input.Arguments); i += 2 {
		key := input.Arguments[i]
		value := input.Arguments[i+1]
		length := uint64(len(value) + len(key))
		useGas += length * k.gasConfig.PersistPerByte

		if !process.IsAllowedToSaveUnderKey(key) {
			return nil, fmt.Errorf("%w it is not allowed to save under key %s", process.ErrOperationNotPermitted, key)
		}

		oldValue, _ := acntDst.DataTrieTracker().RetrieveValue(key)
		if bytes.Equal(oldValue, value) {
			continue
		}

		lengthChange := uint64(0)
		lengthOldValue := uint64(len(oldValue))
		lengthNewValue := uint64(len(value))
		if lengthOldValue < lengthNewValue {
			lengthChange = lengthNewValue - lengthOldValue
		}

		useGas += k.gasConfig.StorePerByte * lengthChange
		if input.GasProvided < useGas {
			return nil, process.ErrNotEnoughGas
		}

		err = acntDst.DataTrieTracker().SaveKeyValue(key, value)
		if err != nil {
			return nil, err
		}
	}

	vmOutput.GasRemaining -= useGas

	return vmOutput, nil
}

func checkArgumentsForSaveKeyValue(acntDst state.UserAccountHandler, input *vmcommon.ContractCallInput) error {
	if input == nil {
		return process.ErrNilVmInput
	}
	if len(input.Arguments) < 2 {
		return process.ErrInvalidArguments
	}
	if len(input.Arguments)%2 != 0 {
		return process.ErrInvalidArguments
	}
	if input.CallValue.Cmp(zero) != 0 {
		return process.ErrBuiltInFunctionCalledWithValue
	}
	if check.IfNil(acntDst) {
		return process.ErrNilSCDestAccount
	}
	if !bytes.Equal(input.CallerAddr, input.RecipientAddr) {
		return fmt.Errorf("%w not the owner of the account", process.ErrOperationNotPermitted)
	}
	if core.IsSmartContractAddress(input.CallerAddr) {
		return fmt.Errorf("%w key-value builtin function not allowed for smart contracts", process.ErrOperationNotPermitted)
	}

	return nil
}

// IsInterfaceNil return true if underlying object in nil
func (k *saveKeyValueStorage) IsInterfaceNil() bool {
	return k == nil
}
