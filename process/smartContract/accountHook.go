package smartContract

import "math/big"

type accountHook struct {
}

// An account with Balance = 0 and Nonce = 0 is considered to not exist
func (ah *accountHook) AccountExists(address []byte) (bool, error) {
	return true, nil
}

// Should retrieve the balance of an account
func (ah *accountHook) GetBalance(address []byte) (*big.Int, error) {
	return big.NewInt(0), nil
}

// Should retrieve the nonce of an account
func (ah *accountHook) GetNonce(address []byte) (*big.Int, error) {
	return big.NewInt(0), nil
}

// Get the storage value for a certain account and index.
// Should return an empty byte array if the key is missing from the account storage
func (ah *accountHook) GetStorageData(accountAddress []byte, index []byte) ([]byte, error) {
	return []byte(""), nil
}

// Should return whether of not an account is SC.
func (ah *accountHook) IsCodeEmpty(address []byte) (bool, error) {
	return true, nil
}

// Should return the compiled and assembled SC code.
// Empty byte array if the account is a wallet.
func (ah *accountHook) GetCode(address []byte) ([]byte, error) {
	return []byte(""), nil
}

// Should return the hash of the nth previous blockchain.
// Offset specifies how many blocks we need to look back.
func (ah *accountHook) GetBlockhash(offset *big.Int) ([]byte, error) {
	return []byte(""), nil
}
