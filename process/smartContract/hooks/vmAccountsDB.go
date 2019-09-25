package hooks

import (
	"encoding/binary"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
)

// VMAccountsDB is a wrapper over AccountsAdapter that satisfy vmcommon.BlockchainHook interface
type VMAccountsDB struct {
	accounts state.AccountsAdapter
	addrConv state.AddressConverter

	mutTempAccounts sync.Mutex
	tempAccounts    map[string]state.AccountHandler
}

// NewVMAccountsDB creates a new VMAccountsDB instance
func NewVMAccountsDB(
	accounts state.AccountsAdapter,
	addrConv state.AddressConverter,
) (*VMAccountsDB, error) {

	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, state.ErrNilAccountsAdapter
	}
	if addrConv == nil || addrConv.IsInterfaceNil() {
		return nil, state.ErrNilAddressConverter
	}

	vmAccountsDB := &VMAccountsDB{
		accounts: accounts,
		addrConv: addrConv,
	}

	vmAccountsDB.tempAccounts = make(map[string]state.AccountHandler, 0)

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
	exists, err := vadb.AccountExists(address)
	if err != nil {
		return nil, err
	}
	if !exists {
		return big.NewInt(0), nil
	}

	shardAccount, err := vadb.getShardAccountFromAddressBytes(address)
	if err != nil {
		return nil, err
	}

	return shardAccount.Balance, nil
}

// GetNonce returns the nonce of a shard account
func (vadb *VMAccountsDB) GetNonce(address []byte) (uint64, error) {
	exists, err := vadb.AccountExists(address)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}

	shardAccount, err := vadb.getShardAccountFromAddressBytes(address)
	if err != nil {
		return 0, err
	}

	return shardAccount.Nonce, nil
}

// GetStorageData returns the storage value of a variable held in account's data trie
func (vadb *VMAccountsDB) GetStorageData(accountAddress []byte, index []byte) ([]byte, error) {
	exists, err := vadb.AccountExists(accountAddress)
	if err != nil {
		return nil, err
	}
	if !exists {
		return make([]byte, 0), nil
	}

	account, err := vadb.getAccountFromAddressBytes(accountAddress)
	if err != nil {
		return nil, err
	}

	return account.DataTrieTracker().RetrieveValue(index)
}

// IsCodeEmpty returns if the code is empty
func (vadb *VMAccountsDB) IsCodeEmpty(address []byte) (bool, error) {
	exists, err := vadb.AccountExists(address)
	if err != nil {
		return false, err
	}
	if !exists {
		return true, nil
	}

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

// NewAddress is a hook which creates a new smart contract address from the creators address and nonce
// The address is created by applied keccak256 on the appended value off creator address and nonce
// Prefix mask is applied for first 8 bytes 0, and for bytes 9-10 - VM type
// Suffix mask is applied - last 2 bytes are for the shard ID - mask is applied as suffix mask
func (vadb *VMAccountsDB) NewAddress(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
	addressLength := vadb.addrConv.AddressLen()
	if len(creatorAddress) != addressLength {
		return nil, ErrAddressLengthNotCorrect
	}

	if len(vmType) != VMTypeLen {
		return nil, ErrVMTypeLengthIsNotCorrect
	}

	_, err := vadb.getShardAccountFromAddressBytes(creatorAddress)
	if err != nil {
		return nil, err
	}

	base := hashFromAddressAndNonce(creatorAddress, creatorNonce)
	prefixMask := createPrefixMask(vmType)
	suffixMask := createSuffixMask(creatorAddress)

	copy(base[:NumInitCharactersForScAddress], prefixMask)
	copy(base[len(base)-ShardIdentiferLen:], suffixMask)

	return base, nil
}

func hashFromAddressAndNonce(creatorAddress []byte, creatorNonce uint64) []byte {
	buffNonce := make([]byte, 8)
	binary.LittleEndian.PutUint64(buffNonce, creatorNonce)
	adrAndNonce := append(creatorAddress, buffNonce...)
	scAddress := keccak.Keccak{}.Compute(string(adrAndNonce))

	return scAddress
}

func createPrefixMask(vmType []byte) []byte {
	prefixMask := make([]byte, NumInitCharactersForScAddress-VMTypeLen)
	prefixMask = append(prefixMask, vmType...)

	return prefixMask
}

func createSuffixMask(creatorAddress []byte) []byte {
	return creatorAddress[len(creatorAddress)-2:]
}

func (vadb *VMAccountsDB) getAccountFromAddressBytes(address []byte) (state.AccountHandler, error) {
	tempAcc, success := vadb.getAccountFromTemporaryAccounts(address)
	if success {
		return tempAcc, nil
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

func (vadb *VMAccountsDB) getAccountFromTemporaryAccounts(address []byte) (state.AccountHandler, bool) {
	vadb.mutTempAccounts.Lock()
	defer vadb.mutTempAccounts.Unlock()

	if tempAcc, ok := vadb.tempAccounts[string(address)]; ok {
		return tempAcc, true
	}

	return nil, false
}

// AddTempAccount will add a temporary account in temporary store
func (vadb *VMAccountsDB) AddTempAccount(address []byte, balance *big.Int, nonce uint64) {
	vadb.mutTempAccounts.Lock()
	vadb.tempAccounts[string(address)] = &state.Account{Balance: balance, Nonce: nonce}
	vadb.mutTempAccounts.Unlock()
}

// CleanTempAccounts cleans the map holding the temporary accounts
func (vadb *VMAccountsDB) CleanTempAccounts() {
	vadb.mutTempAccounts.Lock()
	vadb.tempAccounts = make(map[string]state.AccountHandler, 0)
	vadb.mutTempAccounts.Unlock()
}

// TempAccount can retrieve a temporary account from provided address
func (vadb *VMAccountsDB) TempAccount(address []byte) state.AccountHandler {
	tempAcc, success := vadb.getAccountFromTemporaryAccounts(address)
	if success {
		return tempAcc
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (vadb *VMAccountsDB) IsInterfaceNil() bool {
	if vadb == nil {
		return true
	}
	return false
}

// VMTypeFromAddressBytes gets VM type from scAddress
func VMTypeFromAddressBytes(scAddress []byte) []byte {
	return scAddress[NumInitCharactersForScAddress-VMTypeLen : NumInitCharactersForScAddress]
}
