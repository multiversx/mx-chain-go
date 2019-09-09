package hooks

import (
	"encoding/hex"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// VMAccountsDB is a wrapper over AccountsAdapter that satisfy vmcommon.BlockchainHook interface
type VMAccountsDB struct {
	accounts         state.AccountsAdapter
	addrConv         state.AddressConverter
	shardCoordinator sharding.Coordinator

	mutTempAccounts sync.Mutex
	tempAccounts    map[string]state.AccountHandler
}

// NewVMAccountsDB creates a new VMAccountsDB instance
func NewVMAccountsDB(
	accounts state.AccountsAdapter,
	addrConv state.AddressConverter,
	coordinator sharding.Coordinator,
) (*VMAccountsDB, error) {

	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, state.ErrNilAccountsAdapter
	}
	if addrConv == nil || addrConv.IsInterfaceNil() {
		return nil, state.ErrNilAddressConverter
	}
	if coordinator == nil || coordinator.IsInterfaceNil() {
		return nil, state.ErrNilShardCoordinator
	}

	vmAccountsDB := &VMAccountsDB{
		accounts:         accounts,
		addrConv:         addrConv,
		shardCoordinator: coordinator,
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

const numInitCharactersForScAddress = 10
const vmTypeLen = 2

// NewAddress is a hook which creates a new smart contract address from the creators address and nonce
// The first 8 bytes are 0, bytes 9-10 are for the VM type, bytes 11-30 are keccak256 from address and nonce of creator
// The last 2 bytes are for the shard ID
func (vadb *VMAccountsDB) NewAddress(creatorAddress []byte, creatorNonce uint64, vmType []byte) ([]byte, error) {
	addressLength := vadb.addrConv.AddressLen()
	if len(creatorAddress) != addressLength {
		return nil, ErrAddressLengthNotCorrect
	}

	if len(vmType) != vmTypeLen {
		return nil, ErrVMTypeLengthIsNotCorrect
	}

	creatorAddr, err := vadb.addrConv.CreateAddressFromPublicKeyBytes(creatorAddress)
	if err != nil {
		return nil, err
	}

	shId := vadb.shardCoordinator.ComputeId(creatorAddr)
	if shId >= vadb.shardCoordinator.NumberOfShards() {
		return nil, ErrAddressIsInUnknownShard
	}

	nonce := big.NewInt(0).SetUint64(creatorNonce)

	adrAndNonce := append(creatorAddress, nonce.Bytes()...)
	kecAdrAndNonce := hex.EncodeToString(keccak.Keccak{}.Compute(string(adrAndNonce)))

	scAddress := []byte(kecAdrAndNonce)
	identifier := make([]byte, numInitCharactersForScAddress-vmTypeLen)
	identifier = append(identifier, vmType...)
	copy(scAddress[:numInitCharactersForScAddress], identifier)

	bytesNeed := int(vadb.shardCoordinator.NumberOfShards()/256) + 1
	for i := addressLength - bytesNeed; i < addressLength; i++ {
		scAddress[i] = creatorAddress[i]
		scAddress[i] = creatorAddress[i]
	}

	return scAddress, nil
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
