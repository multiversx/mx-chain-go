package builtInFunctions

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

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

	var zero []byte
	if bytes.Equal(oldValue, zero) {
		useGas += k.gasConfig.StorePerByte * length
		if input.GasProvided < useGas {
			return nil, input.GasProvided, process.ErrNotEnoughGas
		}

		return big.NewInt(0), useGas, nil
	}
	if bytes.Equal(value, zero) {
		freeGas := metering.GasSchedule().BaseOperationCost.StorePerByte * uint64(len(oldValue))
		metering.FreeGas(freeGas)
	}

	useGas := metering.GasSchedule().BaseOperationCost.PersistPerByte * uint64(length)
	metering.UseGas(useGas)

	return big.NewInt(0), 0, nil
}

// IsInterfaceNil return true if underlying object in nil
func (k *saveKeyValueStorage) IsInterfaceNil() bool {
	return k == nil
}
