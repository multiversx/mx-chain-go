package state

import (
	"bytes"
	//"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"math/big"
)

type Account struct {
	Nonce    uint64
	Balance  *big.Int
	CodeHash []byte
	Root     []byte
}

type AccountState struct {
	Account
	Addr     Address
	AddrHash []byte
	Code     []byte
	Data     trie.Trier
	hasher   hashing.Hasher
	prevRoot []byte
}

func NewAccountState(address Address, account Account, hasher hashing.Hasher) *AccountState {
	acState := AccountState{Account: account, Addr: address, prevRoot: account.Root}
	if acState.Balance == nil {
		//an account is inconsistent if Balance is nil.
		acState.Balance = big.NewInt(0)
	}

	acState.hasher = hasher

	acState.AddrHash = address.ComputeHash(hasher)

	return &acState
}

func (as *AccountState) Dirty() bool {
	if as.Data == nil {
		return false
	}

	if (as.prevRoot == nil) || (as.Root == nil) {
		return true
	}

	return !bytes.Equal(as.Data.Root(), as.prevRoot)
}

func (as *AccountState) resetDataCode() {
	//reset code and data
	as.CodeHash = nil
	as.Code = nil
	as.Root = nil
	as.Data = nil
}
