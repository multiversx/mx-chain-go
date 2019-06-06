package hooks

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

// VMAccountsDB is a wrapper over AccountsAdapter that satisfy vmcommon.BlockchainHook interface
type VMAccountsDB struct {
	accounts state.AccountsAdapter
	addrConv state.AddressConverter

	mutFakeAccounts sync.Mutex
	fakeAccounts    map[string]state.AccountHandler
}

// NewVMAccountsDB creates a new VMAccountsDB instance
func NewVMAccountsDB(
	accounts state.AccountsAdapter,
	addrConv state.AddressConverter,
) (*VMAccountsDB, error) {

	if accounts == nil {
		return nil, state.ErrNilAccountsAdapter
	}
	if addrConv == nil {
		return nil, state.ErrNilAddressConverter
	}

	vmAccountsDB := &VMAccountsDB{
		accounts: accounts,
		addrConv: addrConv,
	}

	vmAccountsDB.fakeAccounts = make(map[string]state.AccountHandler, 0)

	return vmAccountsDB, nil
}

// AccountExists checks if an account exists in provided AccountAdapter
func (vadb *VMAccountsDB) AccountExists(address []byte) (bool, error) {
	_, err := vadb.getAccountFromAddressBytes(address)
	if err != nil {
		if err == state.ErrAccNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetBalance returns the balance of a shard account
func (vadb *VMAccountsDB) GetBalance(address []byte) (*big.Int, error) {
	shardAccount, err := vadb.getShardAccountFromAddressBytes(address)
	if err != nil {
		return nil, err
	}

	return shardAccount.Balance, nil
}

// GetNonce returns the nonce of a shard account
func (vadb *VMAccountsDB) GetNonce(address []byte) (*big.Int, error) {
	shardAccount, err := vadb.getShardAccountFromAddressBytes(address)
	if err != nil {
		return nil, err
	}

	nonce := big.NewInt(0)
	nonce.SetUint64(shardAccount.Nonce)

	return nonce, nil
}

// GetStorageData returns the storage value of a variable held in account's data trie
func (vadb *VMAccountsDB) GetStorageData(accountAddress []byte, index []byte) ([]byte, error) {
	account, err := vadb.getAccountFromAddressBytes(accountAddress)
	if err != nil {
		return nil, err
	}

	return account.DataTrieTracker().RetrieveValue(index)
}

// IsCodeEmpty returns if the code is empty
func (vadb *VMAccountsDB) IsCodeEmpty(address []byte) (bool, error) {
	account, err := vadb.getAccountFromAddressBytes(address)
	if err != nil {
		return false, err
	}

	isCodeEmpty := len(account.GetCode()) == 0
	return isCodeEmpty, nil
}

// GetCode retrieves the account's code
func (vadb *VMAccountsDB) GetCode(address []byte) ([]byte, error) {
	account, err := vadb.getAccountFromAddressBytes(address)
	if err != nil {
		return nil, err
	}

	code := account.GetCode()
	if len(code) == 0 {
		return nil, ErrEmptyCode
	}

	return code, nil
}

// GetBlockhash is deprecated
func (vadb *VMAccountsDB) GetBlockhash(offset *big.Int) ([]byte, error) {
	return nil, nil
}

func (vadb *VMAccountsDB) getAccountFromAddressBytes(address []byte) (state.AccountHandler, error) {
	fakeAcc, success := vadb.getAccountFromFakeAccounts(address)
	if success {
		return fakeAcc, nil
	}

	addr, err := vadb.addrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return nil, err
	}

	return vadb.accounts.GetExistingAccount(addr)
}

func (vadb *VMAccountsDB) getShardAccountFromAddressBytes(address []byte) (*state.Account, error) {
	account, err := vadb.getAccountFromAddressBytes(address)
	if err != nil {
		return nil, err
	}

	shardAccount, ok := account.(*state.Account)
	if !ok {
		return nil, state.ErrWrongTypeAssertion
	}

	return shardAccount, nil
}

func (vadb *VMAccountsDB) getAccountFromFakeAccounts(address []byte) (state.AccountHandler, bool) {
	vadb.mutFakeAccounts.Lock()
	defer vadb.mutFakeAccounts.Unlock()

	if fakeAcc, ok := vadb.fakeAccounts[string(address)]; ok {
		return fakeAcc, true
	}

	return nil, false
}

func (vadb *VMAccountsDB) CreateFakeAccounts(address []byte, balance *big.Int, nonce uint64) {
	vadb.mutFakeAccounts.Lock()
	vadb.fakeAccounts[string(address)] = &state.Account{Balance: balance, Nonce: nonce}
	vadb.mutFakeAccounts.Unlock()
}

func (vadb *VMAccountsDB) CleanFakeAccounts() {
	vadb.mutFakeAccounts.Lock()
	vadb.fakeAccounts = make(map[string]state.AccountHandler, 0)
	vadb.mutFakeAccounts.Unlock()
}

func (vadb *VMAccountsDB) GetFakeAccount(address []byte) state.AccountHandler {
	fakeAcc, success := vadb.getAccountFromFakeAccounts(address)
	if success {
		return fakeAcc
	}

	return nil
}
