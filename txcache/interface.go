package txcache

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-storage-go/types"
)

// TxGasHandler handles a transaction gas and gas cost
type TxGasHandler interface {
	ComputeTxFee(tx data.TransactionWithFeeHandler) *big.Int
	IsInterfaceNil() bool
}

// AccountStateProvider defines the behavior of a component able to provide the state of an account
type AccountStateProvider interface {
	GetAccountState(accountKey []byte) (*types.AccountState, error)
	IsInterfaceNil() bool
}

// ForEachTransaction is an iterator callback
type ForEachTransaction func(txHash []byte, value *WrappedTransaction)
