package state

import (
	"math/big"
)

// Account is the struct used in serialization/deserialization
type Account struct {
	Nonce    uint64
	Balance  big.Int
	CodeHash []byte
	RootHash []byte
}

// NewAccount creates a new account object
func NewAccount() *Account {
	return &Account{}
}

//TODO add Cap'N'Proto converter funcs
