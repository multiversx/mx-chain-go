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
	if input == nil {
		return nil, process.ErrNilVmInput
	}
	if len(input.Arguments) != 2 {
		return nil, process.ErrInvalidArguments
	}
	if check.IfNil(acntDst) {
		return nil, process.ErrNilSCDestAccount
	}
	if !bytes.Equal(input.CallerAddr, acntDst.AddressContainer().Bytes()) {
		return nil, fmt.Errorf("%w not the owner of the account", process.ErrOperationNotPermitted)
	}
	if core.IsSmartContractAddress(input.CallerAddr) {
		return nil, fmt.Errorf("%w key-value builtin function not allowed for smart contracts", process.ErrOperationNotPermitted)
	}

	value := input.Arguments[1]
	length := uint64(len(value))
	useGas := k.funcGasCost + length*k.gasConfig.DataCopyPerByte
	if input.GasProvided < useGas {
		return nil, process.ErrNotEnoughGas
	}

	vmOutput := &vmcommon.VMOutput{GasRemaining: input.GasProvided - useGas}

	key := input.Arguments[0]
	if !process.IsAllowedToSaveUnderKey(key) {
		return nil, fmt.Errorf("%w it is not allowed to save under key %s", process.ErrOperationNotPermitted, key)
	}

	oldValue, _ := acntDst.DataTrieTracker().RetrieveValue(key)
	if bytes.Equal(oldValue, value) {
		return vmOutput, nil
	}

	lengthChange := uint64(0)
	lengthOldValue := uint64(len(oldValue))
	if lengthOldValue < length {
		lengthChange = length - lengthOldValue
	} else {
		releaseLength := lengthOldValue - length
		vmOutput.GasRefund = big.NewInt(0).Mul(big.NewInt(0).SetUint64(releaseLength), big.NewInt(0).SetUint64(k.gasConfig.ReleasePerByte))
	}

	useGas += k.gasConfig.StorePerByte * lengthChange
	if input.GasProvided < useGas {
		return nil, process.ErrNotEnoughGas
	}

	vmOutput.GasRemaining = input.GasProvided - useGas
	acntDst.DataTrieTracker().SaveKeyValue(key, value)

	return vmOutput, nil
}

// IsInterfaceNil return true if underlying object in nil
func (k *saveKeyValueStorage) IsInterfaceNil() bool {
	return k == nil
}
