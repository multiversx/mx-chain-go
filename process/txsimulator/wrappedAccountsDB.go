package txsimulator

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// readOnlyAccountsDB is a wrapper over an accounts db which works read-only. write operation are disabled
type readOnlyAccountsDB struct {
	originalAccounts state.AccountsAdapter
}

// NewReadOnlyAccountsDB returns a new instance of readOnlyAccountsDB
func NewReadOnlyAccountsDB(accountsDB state.AccountsAdapter) (*readOnlyAccountsDB, error) {
	if check.IfNil(accountsDB) {
		return nil, ErrNilAccountsAdapter
	}

	return &readOnlyAccountsDB{originalAccounts: accountsDB}, nil
}

// SetSyncer returns nil for this implementation
func (r *readOnlyAccountsDB) SetSyncer(_ state.AccountsDBSyncer) error {
	return nil
}

// StartSnapshotIfNeeded returns nil for this implementation
func (r *readOnlyAccountsDB) StartSnapshotIfNeeded() error {
	return nil
}

// GetCode returns the code for the given account
func (r *readOnlyAccountsDB) GetCode(codeHash []byte) []byte {
	return r.originalAccounts.GetCode(codeHash)
}

// GetExistingAccount will call the original accounts' function with the same name
func (r *readOnlyAccountsDB) GetExistingAccount(address []byte) (vmcommon.AccountHandler, error) {
	return r.originalAccounts.GetExistingAccount(address)
}

// GetAccountFromBytes will call the original accounts' function with the same name
func (r *readOnlyAccountsDB) GetAccountFromBytes(address []byte, accountBytes []byte) (vmcommon.AccountHandler, error) {
	return r.originalAccounts.GetAccountFromBytes(address, accountBytes)
}

// LoadAccount will call the original accounts' function with the same name
func (r *readOnlyAccountsDB) LoadAccount(address []byte) (vmcommon.AccountHandler, error) {
	return r.originalAccounts.LoadAccount(address)
}

// SaveAccount won't do anything as write operations are disabled on this component
func (r *readOnlyAccountsDB) SaveAccount(_ vmcommon.AccountHandler) error {
	return nil
}

// RemoveAccount won't do anything as write operations are disabled on this component
func (r *readOnlyAccountsDB) RemoveAccount(_ []byte) error {
	return nil
}

// Commit won't do anything as write operations are disabled on this component
func (r *readOnlyAccountsDB) Commit() ([]byte, error) {
	return nil, nil
}

// JournalLen will call the original accounts' function with the same name
func (r *readOnlyAccountsDB) JournalLen() int {
	return r.originalAccounts.JournalLen()
}

// RevertToSnapshot won't do anything as write operations are disabled on this component
func (r *readOnlyAccountsDB) RevertToSnapshot(_ int) error {
	return nil
}

// RootHash will call the original accounts' function with the same name
func (r *readOnlyAccountsDB) RootHash() ([]byte, error) {
	return r.originalAccounts.RootHash()
}

// RecreateTrie won't do anything as write operations are disabled on this component
func (r *readOnlyAccountsDB) RecreateTrie(_ []byte) error {
	return nil
}

// RecreateTrieFromEpoch won't do anything as write operations are disabled on this component
func (r *readOnlyAccountsDB) RecreateTrieFromEpoch(_ common.RootHashHolder) error {
	return nil
}

// PruneTrie won't do anything as write operations are disabled on this component
func (r *readOnlyAccountsDB) PruneTrie(_ []byte, _ state.TriePruningIdentifier, _ state.PruningHandler) {
}

// CancelPrune won't do anything as write operations are disabled on this component
func (r *readOnlyAccountsDB) CancelPrune(_ []byte, _ state.TriePruningIdentifier) {
}

// SnapshotState won't do anything as write operations are disabled on this component
func (r *readOnlyAccountsDB) SnapshotState(_ []byte) {
}

// SetStateCheckpoint won't do anything as write operations are disabled on this component
func (r *readOnlyAccountsDB) SetStateCheckpoint(_ []byte) {
}

// IsPruningEnabled will call the original accounts' function with the same name
func (r *readOnlyAccountsDB) IsPruningEnabled() bool {
	return r.originalAccounts.IsPruningEnabled()
}

// GetAllLeaves will call the original accounts' function with the same name
func (r *readOnlyAccountsDB) GetAllLeaves(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeafParser common.TrieLeafParser) error {
	return r.originalAccounts.GetAllLeaves(leavesChannels, ctx, rootHash, trieLeafParser)
}

// RecreateAllTries will return an error which indicates that this operation is not supported
func (r *readOnlyAccountsDB) RecreateAllTries(_ []byte) (map[string]common.Trie, error) {
	return nil, nil
}

// GetTrie will return an error which indicates that this operation is not supported
func (r *readOnlyAccountsDB) GetTrie(_ []byte) (common.Trie, error) {
	return nil, nil
}

// CommitInEpoch will do nothing for this implementation
func (r *readOnlyAccountsDB) CommitInEpoch(_ uint32, _ uint32) ([]byte, error) {
	return nil, nil
}

// GetStackDebugFirstEntry -
func (r *readOnlyAccountsDB) GetStackDebugFirstEntry() []byte {
	return nil
}

// Close will handle the closing of the underlying components
func (r *readOnlyAccountsDB) Close() error {
	return r.originalAccounts.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *readOnlyAccountsDB) IsInterfaceNil() bool {
	return r == nil
}
