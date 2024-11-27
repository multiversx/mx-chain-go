package disabled

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core/check"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

type accountsAdapter struct {
	accountFactory state.AccountFactory
}

// NewAccountsAdapter returns a nil implementation of accountsAdapter
func NewAccountsAdapter(accountFactory state.AccountFactory) (*accountsAdapter, error) {
	if check.IfNil(accountFactory) {
		return nil, state.ErrNilAccountFactory
	}

	return &accountsAdapter{
		accountFactory: accountFactory,
	}, nil
}

// SetSyncer -
func (a *accountsAdapter) SetSyncer(_ state.AccountsDBSyncer) error {
	return nil
}

// StartSnapshotIfNeeded -
func (a *accountsAdapter) StartSnapshotIfNeeded() error {
	return nil
}

// GetTrie -
func (a *accountsAdapter) GetTrie(_ []byte) (common.Trie, error) {
	return nil, nil
}

// GetCode -
func (a *accountsAdapter) GetCode(_ []byte) []byte {
	return nil
}

// LoadAccount -
func (a *accountsAdapter) LoadAccount(address []byte) (vmcommon.AccountHandler, error) {
	return a.accountFactory.CreateAccount(address)
}

// SaveAccount -
func (a *accountsAdapter) SaveAccount(_ vmcommon.AccountHandler) error {
	return nil
}

// PruneTrie -
func (a *accountsAdapter) PruneTrie(_ []byte, _ state.TriePruningIdentifier, _ state.PruningHandler) {
}

// GetExistingAccount -
func (a *accountsAdapter) GetExistingAccount(address []byte) (vmcommon.AccountHandler, error) {
	return a.accountFactory.CreateAccount(address)
}

// GetAccountFromBytes -
func (a *accountsAdapter) GetAccountFromBytes(address []byte, _ []byte) (vmcommon.AccountHandler, error) {
	return a.accountFactory.CreateAccount(address)
}

// RemoveAccount -
func (a *accountsAdapter) RemoveAccount(_ []byte) error {
	return nil
}

// Commit -
func (a *accountsAdapter) Commit() ([]byte, error) {
	return nil, nil
}

// JournalLen -
func (a *accountsAdapter) JournalLen() int {
	return 0
}

// RevertToSnapshot -
func (a *accountsAdapter) RevertToSnapshot(_ int) error {
	return nil
}

// RootHash -
func (a *accountsAdapter) RootHash() ([]byte, error) {
	return nil, nil
}

// RecreateTrie -
func (a *accountsAdapter) RecreateTrie(_ common.RootHashHolder) error {
	return nil
}

// CancelPrune -
func (a *accountsAdapter) CancelPrune(_ []byte, _ state.TriePruningIdentifier) {
}

// SnapshotState -
func (a *accountsAdapter) SnapshotState(_ []byte, _ uint32) {
}

// IsPruningEnabled -
func (a *accountsAdapter) IsPruningEnabled() bool {
	return false
}

// ClosePersister -
func (a *accountsAdapter) ClosePersister() error {
	return nil
}

// GetAllLeaves -
func (a *accountsAdapter) GetAllLeaves(_ *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.TrieLeafParser) error {
	return nil
}

// RecreateAllTries -
func (a *accountsAdapter) RecreateAllTries(_ []byte) (map[string]common.Trie, error) {
	return nil, nil
}

// CommitInEpoch -
func (a *accountsAdapter) CommitInEpoch(_ uint32, _ uint32) ([]byte, error) {
	return nil, nil
}

// GetStackDebugFirstEntry -
func (a *accountsAdapter) GetStackDebugFirstEntry() []byte {
	return nil
}

// Close -
func (a *accountsAdapter) Close() error {
	return nil
}

// IsInterfaceNil -
func (a *accountsAdapter) IsInterfaceNil() bool {
	return a == nil
}
