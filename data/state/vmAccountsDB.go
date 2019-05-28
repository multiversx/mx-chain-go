package state

import "math/big"

// VMAccountsDB is a wrapper over AccountsAdapter that satisfy vmcommon.BlockchainHook interface
type VMAccountsDB struct {
	accounts AccountsAdapter
	addrConv AddressConverter
}

// NewVMAccountsDB creates a new VMAccountsDB instance
func NewVMAccountsDB(
	accounts AccountsAdapter,
	addrConv AddressConverter,
) (*VMAccountsDB, error) {

	if accounts == nil {
		return nil, ErrNilAccountsAdapter
	}
	if addrConv == nil {
		return nil, ErrNilAddressContainer
	}

	return &VMAccountsDB{
		accounts: accounts,
		addrConv: addrConv,
	}, nil
}

// AccountExists checks if an account exists in provided AccountAdapter
func (vadb *VMAccountsDB) AccountExists(address []byte) (bool, error) {
	addr, err := vadb.addrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return false, err
	}

	_, err = vadb.accounts.GetExistingAccount(addr)
	if err != nil {
		if err == ErrAccNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetBalance returns the balance of an account
func (vadb *VMAccountsDB) GetBalance(address []byte) (*big.Int, error) {
	panic("implement me")
}

func (vadb *VMAccountsDB) GetNonce(address []byte) (*big.Int, error) {
	panic("implement me")
}

func (vadb *VMAccountsDB) GetStorageData(accountAddress []byte, index []byte) ([]byte, error) {
	panic("implement me")
}

func (vadb *VMAccountsDB) IsCodeEmpty(address []byte) (bool, error) {
	panic("implement me")
}

func (vadb *VMAccountsDB) GetCode(address []byte) ([]byte, error) {
	panic("implement me")
}

// GetBlockhash is deprecated
func (vadb *VMAccountsDB) GetBlockhash(offset *big.Int) ([]byte, error) {
	return nil, nil
}
