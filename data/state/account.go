package state

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

// Account is a struct that will be serialized/deserialized
type Account struct {
	Nonce    uint64
	Balance  *big.Int
	CodeHash []byte
	Root     []byte
}

// AccountState is a struct that wraps Account and add functionalities to it
type AccountState struct {
	Account
	Addr     Address
	Code     []byte
	Data     trie.PatriciaMerkelTree
	hasher   hashing.Hasher
	prevRoot []byte
}

// NewAccountState creates new wrapper for an Account (that has just been retrieved)
func NewAccountState(address Address, account Account, hasher hashing.Hasher) *AccountState {
	acState := AccountState{Account: account, Addr: address, prevRoot: account.Root}
	if acState.Balance == nil {
		//an account is inconsistent if Balance is nil.
		acState.Balance = big.NewInt(0)
	}

	acState.hasher = hasher

	return &acState
}

// Dirty returns true if data inside data trie has changed
// Useful when we track all data tries that need to be committed/undo-ed in persistence unit
// The status is computed as a difference between previous root hash of the data trie and the current
// root hash of the data trie. When committing data to persistence, prevRoot becomes Root as it will
// cause Dirty() to return false. When undo-ing data trie, Root will become prevRoot and so Dirty() will
// also return false
func (as *AccountState) Dirty() bool {
	if as.Data == nil {
		return false
	}

	if (as.prevRoot == nil) || (as.Root == nil) {
		return true
	}

	return !bytes.Equal(as.Data.Root(), as.prevRoot)
}

// Resets (nils) the fields inside Account
func (as *AccountState) reset() {
	as.CodeHash = nil
	as.Code = nil
	as.Root = nil
	as.Data = nil
}
