package state

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go/common"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type accountsDBApiWithHistory struct {
	innerAccountsAdapter AccountsAdapter
}

// NewAccountsDBApiWithHistory will create a new instance of type accountsDBApiWithHistory
func NewAccountsDBApiWithHistory(innerAccountsAdapter AccountsAdapter) (*accountsDBApiWithHistory, error) {
	if check.IfNil(innerAccountsAdapter) {
		return nil, ErrNilAccountsAdapter
	}

	return &accountsDBApiWithHistory{
		innerAccountsAdapter: innerAccountsAdapter,
	}, nil
}

// GetExistingAccount will return an error
func (accountsDB *accountsDBApiWithHistory) GetExistingAccount(address []byte) (vmcommon.AccountHandler, error) {
	return nil, ErrFunctionalityNotImplemented
}

// GetAccountFromBytes will return an error
func (accountsDB *accountsDBApiWithHistory) GetAccountFromBytes(address []byte, accountBytes []byte) (vmcommon.AccountHandler, error) {
	return nil, ErrFunctionalityNotImplemented
}

// LoadAccount will return an error
func (accountsDB *accountsDBApiWithHistory) LoadAccount(address []byte) (vmcommon.AccountHandler, error) {
	return nil, ErrFunctionalityNotImplemented
}

// SaveAccount is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApiWithHistory) SaveAccount(_ vmcommon.AccountHandler) error {
	return ErrOperationNotPermitted
}

// RemoveAccount is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApiWithHistory) RemoveAccount(_ []byte) error {
	return ErrOperationNotPermitted
}

// CommitInEpoch is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApiWithHistory) CommitInEpoch(_ uint32, _ uint32) ([]byte, error) {
	return nil, ErrOperationNotPermitted
}

// Commit is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApiWithHistory) Commit() ([]byte, error) {
	return nil, ErrOperationNotPermitted
}

// JournalLen will always return 0
func (accountsDB *accountsDBApiWithHistory) JournalLen() int {
	return 0
}

// RevertToSnapshot is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApiWithHistory) RevertToSnapshot(_ int) error {
	return ErrOperationNotPermitted
}

// GetCode will call the inner accountsAdapter method after trying to recreate the trie
func (accountsDB *accountsDBApiWithHistory) GetCode(codeHash []byte) []byte {
	return nil
}

// RootHash will return an error
func (accountsDB *accountsDBApiWithHistory) RootHash() ([]byte, error) {
	return nil, ErrOperationNotPermitted
}

// RecreateTrie is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApiWithHistory) RecreateTrie(_ []byte) error {
	return ErrOperationNotPermitted
}

// PruneTrie is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApiWithHistory) PruneTrie(_ []byte, _ TriePruningIdentifier, _ PruningHandler) {
}

// CancelPrune is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApiWithHistory) CancelPrune(_ []byte, _ TriePruningIdentifier) {
}

// SnapshotState is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApiWithHistory) SnapshotState(_ []byte) {
}

// SetStateCheckpoint is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApiWithHistory) SetStateCheckpoint(_ []byte) {
}

// IsPruningEnabled will call the inner accountsAdapter method
func (accountsDB *accountsDBApiWithHistory) IsPruningEnabled() bool {
	return accountsDB.innerAccountsAdapter.IsPruningEnabled()
}

// GetAllLeaves will return an error
func (accountsDB *accountsDBApiWithHistory) GetAllLeaves(leavesChannel chan core.KeyValueHolder, ctx context.Context, rootHash []byte) error {
	return ErrOperationNotPermitted
}

// RecreateAllTries is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApiWithHistory) RecreateAllTries(_ []byte) (map[string]common.Trie, error) {
	return nil, ErrOperationNotPermitted
}

// GetTrie will call the inner accountsAdapter method
func (accountsDB *accountsDBApiWithHistory) GetTrie(rootHash []byte) (common.Trie, error) {
	return accountsDB.innerAccountsAdapter.GetTrie(rootHash)
}

// GetStackDebugFirstEntry will call the inner accountsAdapter method
func (accountsDB *accountsDBApiWithHistory) GetStackDebugFirstEntry() []byte {
	return accountsDB.innerAccountsAdapter.GetStackDebugFirstEntry()
}

// Close will handle the closing of the underlying components
func (accountsDB *accountsDBApiWithHistory) Close() error {
	return accountsDB.innerAccountsAdapter.Close()
}

// GetAccountWithBlockInfo returns the account and the associated block info
func (accountsDB *accountsDBApiWithHistory) GetAccountWithBlockInfo(address []byte, options api.AccountQueryOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
	return nil, nil, ErrFunctionalityNotImplemented
}

// GetCodeWithBlockInfo returns the code and the associated block info
func (accountsDB *accountsDBApiWithHistory) GetCodeWithBlockInfo(codeHash []byte, options api.AccountQueryOptions) ([]byte, common.BlockInfo, error) {
	return nil, nil, ErrFunctionalityNotImplemented
}

func (accountsDB *accountsDBApiWithHistory) recreateTrie(rootHash []byte) error {
	// TODO: perhaps we can optimize this in the future: check the latest rootHash used for re-creation, and skip re-creation, if possible.
	return accountsDB.innerAccountsAdapter.RecreateTrie(rootHash)
}

// IsInterfaceNil returns true if there is no value under the interface
func (accountsDB *accountsDBApiWithHistory) IsInterfaceNil() bool {
	return accountsDB == nil
}
