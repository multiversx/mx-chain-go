package builtInFunctions

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// KeyValueGasCost -
type KeyValueGasCost struct {
	BaseCost       uint64
	SavePerByte    uint64
	ChangePerByte  uint64
	ReleasePerByte uint64
}

type keyValueStorage struct {
	gasConfig KeyValueGasCost
}

// NewKeyValueStorageFunc -
func NewKeyValueStorageFunc(
	gasConfig KeyValueGasCost,
) *keyValueStorage {
	return &keyValueStorage{gasConfig: gasConfig}
}

// ProcessBuiltinFunction -
func (k *keyValueStorage) ProcessBuiltinFunction(acntSnd, acntDst state.UserAccountHandler, vmInput *vmcommon.ContractCallInput) (*big.Int, error) {
	return big.NewInt(0), nil
}

// GasUsed -
func (k *keyValueStorage) GasUsed() uint64 {
	return 0
}

// IsInterfaceNil -
func (k *keyValueStorage) IsInterfaceNil() bool {
	return k == nil
}
