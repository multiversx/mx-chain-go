package builtInFunctions

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type keyValueStorage struct {
	gasConfig BaseOperationCost
}

// NewKeyValueStorageFunc -
func NewKeyValueStorageFunc(
	gasConfig BaseOperationCost,
) *keyValueStorage {
	return &keyValueStorage{gasConfig: gasConfig}
}

// ProcessBuiltinFunction -
func (k *keyValueStorage) ProcessBuiltinFunction(_, _ state.UserAccountHandler, _ *vmcommon.ContractCallInput) (*big.Int, uint64, error) {
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
