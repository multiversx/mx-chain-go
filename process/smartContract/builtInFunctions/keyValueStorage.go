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

const numInitCharactersInProtectedKeys = 4

type saveKeyValueStorage struct {
	gasConfig      BaseOperationCost
	funcGasCost    uint64
	mapInvalidKeys map[string]struct{}
}

// NewSaveKeyValueStorageFunc returns the key-value
func NewSaveKeyValueStorageFunc(
	gasConfig BaseOperationCost,
	funcGasCost uint64,
	mapInvalidKeys map[string]struct{},
) (*saveKeyValueStorage, error) {
	if mapInvalidKeys == nil {
		return nil, process.ErrNilInvalidKeysMap
	}

	s := &saveKeyValueStorage{
		gasConfig:   gasConfig,
		funcGasCost: funcGasCost,
	}
	s.mapInvalidKeys = make(map[string]struct{}, len(mapInvalidKeys))
	for key := range mapInvalidKeys {
		trimmedKey := key[:numInitCharactersInProtectedKeys]
		s.mapInvalidKeys[trimmedKey] = struct{}{}
	}

	return s, nil
}

func (k *saveKeyValueStorage) isAllowedToSaveUnderKey(key []byte) bool {
	if len(key) < numInitCharactersInProtectedKeys {
		return true
	}

	trimmedKey := key[:numInitCharactersInProtectedKeys]
	_, ok := k.mapInvalidKeys[string(trimmedKey)]
	return !ok
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
		return nil, input.GasProvided, fmt.Errorf("%w not the owner of the account", process.ErrOperationNotPermitted)
	}

	value := input.Arguments[1]
	length := uint64(len(value))
	useGas := k.funcGasCost + length*k.gasConfig.DataCopyPerByte
	if input.GasProvided < useGas {
		return nil, input.GasProvided, process.ErrNotEnoughGas
	}

	key := input.Arguments[0]
	if !k.isAllowedToSaveUnderKey(key) {
		return nil, input.GasProvided, fmt.Errorf("%w it is not allowed to save under key %s", process.ErrOperationNotPermitted, key)
	}

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
