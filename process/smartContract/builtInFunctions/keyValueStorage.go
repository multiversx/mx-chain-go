package builtInFunctions

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var _ process.BuiltinFunction = (*saveKeyValueStorage)(nil)

type saveKeyValueStorage struct {
	gasConfig   BaseOperationCost
	funcGasCost uint64
}

// NewSaveKeyValueStorageFunc returns the save key-value storage built in function
func NewSaveKeyValueStorageFunc(
	gasConfig BaseOperationCost,
	funcGasCost uint64,
) (*saveKeyValueStorage, error) {
	s := &saveKeyValueStorage{
		gasConfig:   gasConfig,
		funcGasCost: funcGasCost,
	}

	return s, nil
}

// ProcessBuiltinFunction will save the value for the selected key
func (k *saveKeyValueStorage) ProcessBuiltinFunction(
	_, acntDst state.UserAccountHandler,
	input *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
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
		if lengthOldValue < length {
			lengthChange = length - lengthOldValue
		} else {
			releaseLength := lengthOldValue - length
			refundValue := big.NewInt(0).Mul(big.NewInt(0).SetUint64(releaseLength), big.NewInt(0).SetUint64(k.gasConfig.ReleasePerByte))
			vmOutput.GasRefund.Add(vmOutput.GasRefund, refundValue)
		}

		useGas += k.gasConfig.StorePerByte * lengthChange
		if input.GasProvided < useGas {
			return nil, process.ErrNotEnoughGas
		}

		acntDst.DataTrieTracker().SaveKeyValue(key, value)
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
