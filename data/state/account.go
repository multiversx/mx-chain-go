package state

import (
	"math/big"
)

// Account is a struct that will be used as a base for all account wrappers
type Account struct {
	nonce    uint64
	balance  big.Int
	codeHash []byte
	rootHash []byte
}

// NewAccount creates a new account object
func NewAccount() *Account {
	return &Account{}
}

// Nonce returns the account's nonce
func (account *Account) Nonce() uint64 {
	return account.nonce
}

// SetNonce sets the account's nonce
func (account *Account) SetNonce(nonce uint64) {
	account.nonce = nonce
}

// Balance returns the account's balance as a big integer
func (account *Account) Balance() big.Int {
	return account.balance
}

// SetBalance sets the account's balance
func (account *Account) SetBalance(balance big.Int) {
	account.balance = balance
}

// CodeHash returns the account's code hash
func (account *Account) CodeHash() []byte {
	return account.codeHash
}

// SetCodeHash sets the account's code hash
func (account *Account) SetCodeHash(codeHash []byte) {
	account.codeHash = codeHash
}

// RootHash returns the data root hash
func (account *Account) RootHash() []byte {
	return account.rootHash
}

// SetRootHash sets the data root hash
func (account *Account) SetRootHash(rootHash []byte) {
	account.rootHash = rootHash
}

// LoadFromDbAccount loads and overwrites current fields data with the source representation
func (account *Account) LoadFromDbAccount(source DbAccountContainer) error {
	if source == nil {
		return ErrNilValue
	}

	balance := big.NewInt(0)
	err := balance.GobDecode(source.Balance())

	if err != nil {
		return err
	}

	account.nonce = source.Nonce()
	account.balance = *balance
	account.rootHash = source.RootHash()
	account.codeHash = source.CodeHash()

	return nil
}

// SaveToDbAccount saves in target object the representation of current values
func (account *Account) SaveToDbAccount(target DbAccountContainer) error {
	if target == nil {
		return ErrNilValue
	}

	buff, err := account.balance.GobEncode()

	if err != nil {
		//according to current implementation, this can not be hit, maintaining to be 100% safe
		return err
	}

	target.SetNonce(account.nonce)
	target.SetBalance(buff)
	target.SetCodeHash(account.codeHash)
	target.SetRootHash(account.rootHash)
	return nil
}
