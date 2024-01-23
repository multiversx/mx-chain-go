package transactionEvaluator

import (
	"context"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// simulationAccountsDB is a wrapper over an accounts db which works read-only. write operation are disabled
type simulationAccountsDB struct {
	mutex            sync.RWMutex
	cachedAccounts   map[string]vmcommon.AccountHandler
	originalAccounts state.AccountsAdapter
}

// NewSimulationAccountsDB returns a new instance of simulationAccountsDB
func NewSimulationAccountsDB(accountsDB state.AccountsAdapter) (*simulationAccountsDB, error) {
	if check.IfNil(accountsDB) {
		return nil, ErrNilAccountsAdapter
	}

	return &simulationAccountsDB{
		mutex:            sync.RWMutex{},
		cachedAccounts:   make(map[string]vmcommon.AccountHandler),
		originalAccounts: accountsDB,
	}, nil
}

// SetSyncer returns nil for this implementation
func (r *simulationAccountsDB) SetSyncer(_ state.AccountsDBSyncer) error {
	return nil
}

// StartSnapshotIfNeeded returns nil for this implementation
func (r *simulationAccountsDB) StartSnapshotIfNeeded() error {
	return nil
}

// GetCode returns the code for the given account
func (r *simulationAccountsDB) GetCode(codeHash []byte) []byte {
	return r.originalAccounts.GetCode(codeHash)
}

// GetExistingAccount will call the original accounts' function with the same name
func (r *simulationAccountsDB) GetExistingAccount(address []byte) (vmcommon.AccountHandler, error) {
	cachedAccount, ok := r.getFromCache(address)
	if ok {
		return cachedAccount, nil
	}

	account, err := r.originalAccounts.GetExistingAccount(address)
	if err != nil {
		return nil, err
	}

	r.addToCache(account)

	return account, nil
}

// GetAccountFromBytes will call the original accounts' function with the same name
func (r *simulationAccountsDB) GetAccountFromBytes(address []byte, accountBytes []byte) (vmcommon.AccountHandler, error) {
	return r.originalAccounts.GetAccountFromBytes(address, accountBytes)
}

// LoadAccount will call the original accounts' function with the same name
func (r *simulationAccountsDB) LoadAccount(address []byte) (vmcommon.AccountHandler, error) {
	cachedAccount, ok := r.getFromCache(address)
	if ok {
		return cachedAccount, nil
	}

	account, err := r.originalAccounts.LoadAccount(address)
	if err != nil {
		return nil, err
	}

	r.addToCache(account)

	return account, nil
}

// SaveAccount won't do anything as write operations are disabled on this component
func (r *simulationAccountsDB) SaveAccount(account vmcommon.AccountHandler) error {
	if check.IfNil(account) {
		return nil
	}

	r.addToCache(account)

	return nil
}

// RemoveAccount won't do anything as write operations are disabled on this component
func (r *simulationAccountsDB) RemoveAccount(_ []byte) error {
	return nil
}

// Commit won't do anything as write operations are disabled on this component
func (r *simulationAccountsDB) Commit() ([]byte, error) {
	return nil, nil
}

// JournalLen will call the original accounts' function with the same name
func (r *simulationAccountsDB) JournalLen() int {
	return r.originalAccounts.JournalLen()
}

// RevertToSnapshot won't do anything as write operations are disabled on this component
func (r *simulationAccountsDB) RevertToSnapshot(_ int) error {
	return nil
}

// RootHash will call the original accounts' function with the same name
func (r *simulationAccountsDB) RootHash() ([]byte, error) {
	return r.originalAccounts.RootHash()
}

// RecreateTrie won't do anything as write operations are disabled on this component
func (r *simulationAccountsDB) RecreateTrie(_ []byte) error {
	return nil
}

// RecreateTrieFromEpoch won't do anything as write operations are disabled on this component
func (r *simulationAccountsDB) RecreateTrieFromEpoch(_ common.RootHashHolder) error {
	return nil
}

// PruneTrie won't do anything as write operations are disabled on this component
func (r *simulationAccountsDB) PruneTrie(_ []byte, _ state.TriePruningIdentifier, _ state.PruningHandler) {
}

// CancelPrune won't do anything as write operations are disabled on this component
func (r *simulationAccountsDB) CancelPrune(_ []byte, _ state.TriePruningIdentifier) {
}

// SnapshotState won't do anything as write operations are disabled on this component
func (r *simulationAccountsDB) SnapshotState(_ []byte, _ uint32) {
}

// IsPruningEnabled will call the original accounts' function with the same name
func (r *simulationAccountsDB) IsPruningEnabled() bool {
	return r.originalAccounts.IsPruningEnabled()
}

// GetAllLeaves will call the original accounts' function with the same name
func (r *simulationAccountsDB) GetAllLeaves(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeafParser common.TrieLeafParser) error {
	return r.originalAccounts.GetAllLeaves(leavesChannels, ctx, rootHash, trieLeafParser)
}

// RecreateAllTries will return an error which indicates that this operation is not supported
func (r *simulationAccountsDB) RecreateAllTries(_ []byte) (map[string]common.Trie, error) {
	return nil, nil
}

// GetTrie will return an error which indicates that this operation is not supported
func (r *simulationAccountsDB) GetTrie(_ []byte) (common.Trie, error) {
	return nil, nil
}

// CommitInEpoch will do nothing for this implementation
func (r *simulationAccountsDB) CommitInEpoch(_ uint32, _ uint32) ([]byte, error) {
	return nil, nil
}

// GetStackDebugFirstEntry -
func (r *simulationAccountsDB) GetStackDebugFirstEntry() []byte {
	return nil
}

// Close will handle the closing of the underlying components
func (r *simulationAccountsDB) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *simulationAccountsDB) IsInterfaceNil() bool {
	return r == nil
}

// CleanCache will clean the internal map with the cached accounts
func (r *simulationAccountsDB) CleanCache() {
	r.mutex.Lock()
	r.cachedAccounts = make(map[string]vmcommon.AccountHandler)
	r.mutex.Unlock()
}

func (r *simulationAccountsDB) addToCache(account vmcommon.AccountHandler) {
	r.mutex.Lock()
	r.cachedAccounts[string(account.AddressBytes())] = account
	r.mutex.Unlock()
}

func (r *simulationAccountsDB) getFromCache(address []byte) (vmcommon.AccountHandler, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	account, found := r.cachedAccounts[string(address)]

	return account, found
}
