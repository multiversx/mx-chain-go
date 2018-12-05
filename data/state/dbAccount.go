package state

// DbAccount is a struct that will be serialized/deserialized inside the trie
// CodeHash and RootHash are hashes to the location where the actual data resides
// The reason for making fields exported and also providing getters and setters
// is that this structure now complies to JSON marshaling requirements and also
// implements the AccountContainer interface that is Cap'n'Proto compliant
type DbAccount struct {
	ValNonce    uint64
	ValBalance  []byte
	ValCodeHash []byte
	ValRootHash []byte
}

// NewDbAccount creates a new account object
func NewDbAccount() *DbAccount {
	return &DbAccount{}
}

// Nonce returns the account's nonce
func (account *DbAccount) Nonce() uint64 {
	return account.ValNonce
}

// SetNonce sets the account's nonce
func (account *DbAccount) SetNonce(nonce uint64) {
	account.ValNonce = nonce
}

// Balance returns the account's balance as a slice of bytes
func (account *DbAccount) Balance() []byte {
	return account.ValBalance
}

// SetBalance sets the account's balance
func (account *DbAccount) SetBalance(balance []byte) {
	account.ValBalance = balance
}

// CodeHash returns the account's code hash
func (account *DbAccount) CodeHash() []byte {
	return account.ValCodeHash
}

// SetCodeHash sets the account's code hash
func (account *DbAccount) SetCodeHash(codeHash []byte) {
	account.ValCodeHash = codeHash
}

// RootHash returns the data root hash
func (account *DbAccount) RootHash() []byte {
	return account.ValRootHash
}

// SetRootHash sets the data root hash
func (account *DbAccount) SetRootHash(rootHash []byte) {
	account.ValRootHash = rootHash
}
