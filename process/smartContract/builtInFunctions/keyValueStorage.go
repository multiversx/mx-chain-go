package builtInFunctions

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type keyValueStorage struct {
	gasConfig smartContract.BaseOperationCost
}

// NewKeyValueStorageFunc -
func NewKeyValueStorageFunc(
	gasConfig smartContract.BaseOperationCost,
) *keyValueStorage {
	return &keyValueStorage{gasConfig: gasConfig}
}

// ProcessBuiltinFunction -
func (k *keyValueStorage) ProcessBuiltinFunction(acntSnd, acntDst state.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*big.Int, uint64, error) {
	return big.NewInt(0), 0, nil
}

// GasUsed -
func (k *keyValueStorage) GasUsed() uint64 {
	return 0
}

// IsInterfaceNil -
func (k *keyValueStorage) IsInterfaceNil() bool {
	return k == nil
}
