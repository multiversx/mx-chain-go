package builtInFunctions

import (
	"bytes"
	"fmt"
	"math/big"

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

// NewSaveKeyValueStorageFunc returns the key-value
func NewSaveKeyValueStorageFunc(
	gasConfig BaseOperationCost,
	funcGasCost uint64,
) *saveKeyValueStorage {
	return &saveKeyValueStorage{gasConfig: gasConfig, funcGasCost: funcGasCost}
}

// ProcessBuiltinFunction will save the value for the selected key
func (k *saveKeyValueStorage) ProcessBuiltinFunction(
	_, acntDst state.UserAccountHandler,
	input *vmcommon.ContractCallInput,
) (*big.Int, uint64, error) {
	if input == nil {
		return nil, 0, process.ErrNilVmInput
	}
	if len(input.Arguments) != 2 {
		return nil, input.GasProvided, process.ErrInvalidArguments
	}
	if check.IfNil(acntDst) {
		return nil, input.GasProvided, process.ErrNilSCDestAccount
	}
	if !bytes.Equal(input.CallerAddr, acntDst.AddressContainer().Bytes()) {
		return nil, 0, fmt.Errorf("%w not the owner of the account", process.ErrOperationNotPermitted)
	}

	value := input.Arguments[1]
	length := uint64(len(value))
	useGas := k.funcGasCost + length*k.gasConfig.DataCopyPerByte
	if input.GasProvided < useGas {
		return nil, input.GasProvided, process.ErrNotEnoughGas
	}

	key := input.Arguments[0]
	oldValue, _ := acntDst.DataTrieTracker().RetrieveValue(key)
	if bytes.Equal(oldValue, value) {
		return big.NewInt(0), useGas, nil
	}

	lengthChange := uint64(0)
	lengthOldValue := uint64(len(oldValue))
	if lengthOldValue < length {
		lengthChange = length - lengthOldValue
	}

	useGas += k.gasConfig.StorePerByte * lengthChange
	if input.GasProvided < useGas {
		return nil, input.GasProvided, process.ErrNotEnoughGas
	}

	acntDst.DataTrieTracker().SaveKeyValue(key, value)

	return big.NewInt(0), useGas, nil
}

// IsInterfaceNil return true if underlying object in nil
func (k *saveKeyValueStorage) IsInterfaceNil() bool {
	return k == nil
}
