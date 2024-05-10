package state

import (
	"bytes"
	"context"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type accountsDBApiWithHistory struct {
	innerAccountsAdapter    AccountsAdapter
	latestRecreatedRootHash []byte
	mutRecreateAndGet       sync.RWMutex
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

// SetSyncer  is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApiWithHistory) SetSyncer(_ AccountsDBSyncer) error {
	return nil
}

// StartSnapshotIfNeeded  is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApiWithHistory) StartSnapshotIfNeeded() error {
	return nil
}

// GetExistingAccount will return an error
func (accountsDB *accountsDBApiWithHistory) GetExistingAccount(_ []byte) (vmcommon.AccountHandler, error) {
	return nil, ErrFunctionalityNotImplemented
}

// GetAccountFromBytes will return an error
func (accountsDB *accountsDBApiWithHistory) GetAccountFromBytes(_ []byte, _ []byte) (vmcommon.AccountHandler, error) {
	return nil, ErrFunctionalityNotImplemented
}

// LoadAccount will return an error
func (accountsDB *accountsDBApiWithHistory) LoadAccount(_ []byte) (vmcommon.AccountHandler, error) {
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
func (accountsDB *accountsDBApiWithHistory) GetCode(_ []byte) []byte {
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

// RecreateTrieFromEpoch is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApiWithHistory) RecreateTrieFromEpoch(_ common.RootHashHolder) error {
	return ErrOperationNotPermitted
}

// PruneTrie is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApiWithHistory) PruneTrie(_ []byte, _ TriePruningIdentifier, _ PruningHandler) {
}

// CancelPrune is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApiWithHistory) CancelPrune(_ []byte, _ TriePruningIdentifier) {
}

// SnapshotState is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApiWithHistory) SnapshotState(_ []byte, _ uint32) {
}

// IsPruningEnabled will return false
func (accountsDB *accountsDBApiWithHistory) IsPruningEnabled() bool {
	return false
}

// GetAllLeaves will return an error
func (accountsDB *accountsDBApiWithHistory) GetAllLeaves(_ *common.TrieIteratorChannels, _ context.Context, _ []byte, _ common.TrieLeafParser) error {
	return ErrOperationNotPermitted
}

// RecreateAllTries is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApiWithHistory) RecreateAllTries(_ []byte) (map[string]common.Trie, error) {
	return nil, ErrOperationNotPermitted
}

// GetTrie is not implemented
func (accountsDB *accountsDBApiWithHistory) GetTrie(_ []byte) (common.Trie, error) {
	return nil, ErrFunctionalityNotImplemented
}

// GetStackDebugFirstEntry returns nil
func (accountsDB *accountsDBApiWithHistory) GetStackDebugFirstEntry() []byte {
	return nil
}

// Close will handle the closing of the underlying components
func (accountsDB *accountsDBApiWithHistory) Close() error {
	return accountsDB.innerAccountsAdapter.Close()
}

// GetAccountWithBlockInfo returns the account and the associated block info
func (accountsDB *accountsDBApiWithHistory) GetAccountWithBlockInfo(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
	rootHash := options.GetRootHash()

	// First check to avoid re-creation:
	accountsDB.mutRecreateAndGet.RLock()
	if !accountsDB.shouldRecreateTrieUnprotected(rootHash) {
		// Re-creation is not necessary, but make sure no other routine re-creates it.
		defer accountsDB.mutRecreateAndGet.RUnlock()
		return accountsDB.doGetAccountWithBlockInfoUnprotected(address, rootHash)
	}

	accountsDB.mutRecreateAndGet.RUnlock()

	// Re-creation seems to be necessary, promote to a "write" lock.
	accountsDB.mutRecreateAndGet.Lock()
	defer accountsDB.mutRecreateAndGet.Unlock()

	// Second check to avoid re-creation:
	if accountsDB.shouldRecreateTrieUnprotected(rootHash) {
		err := accountsDB.recreateTrieUnprotected(options)
		if err != nil {
			return nil, nil, err
		}
	}

	return accountsDB.doGetAccountWithBlockInfoUnprotected(address, rootHash)
}

func (accountsDB *accountsDBApiWithHistory) doGetAccountWithBlockInfoUnprotected(address []byte, rootHash []byte) (vmcommon.AccountHandler, common.BlockInfo, error) {
	blockInfo := holders.NewBlockInfo(nil, 0, rootHash)

	account, err := accountsDB.innerAccountsAdapter.GetExistingAccount(address)
	if err == ErrAccNotFound {
		return nil, nil, NewErrAccountNotFoundAtBlock(blockInfo)
	}
	if err != nil {
		return nil, nil, err
	}

	return account, blockInfo, nil
}

// GetCodeWithBlockInfo returns the code and the associated block info
func (accountsDB *accountsDBApiWithHistory) GetCodeWithBlockInfo(codeHash []byte, options common.RootHashHolder) ([]byte, common.BlockInfo, error) {
	rootHash := options.GetRootHash()

	// First check to avoid re-creation:
	accountsDB.mutRecreateAndGet.RLock()
	if !accountsDB.shouldRecreateTrieUnprotected(rootHash) {
		// Re-creation is not necessary, but make sure no other routine re-creates it.
		defer accountsDB.mutRecreateAndGet.RUnlock()
		return accountsDB.doGetCodeWithBlockInfoUnprotected(codeHash, rootHash)
	}

	accountsDB.mutRecreateAndGet.RUnlock()

	// Re-creation seems to be necessary, promote to a "write" lock.
	accountsDB.mutRecreateAndGet.Lock()
	defer accountsDB.mutRecreateAndGet.Unlock()

	// Second check to avoid re-creation:
	if accountsDB.shouldRecreateTrieUnprotected(rootHash) {
		err := accountsDB.recreateTrieUnprotected(options)
		if err != nil {
			return nil, nil, err
		}
	}

	return accountsDB.doGetCodeWithBlockInfoUnprotected(codeHash, rootHash)
}

func (accountsDB *accountsDBApiWithHistory) doGetCodeWithBlockInfoUnprotected(codeHash []byte, rootHash []byte) ([]byte, common.BlockInfo, error) {
	blockInfo := holders.NewBlockInfo(nil, 0, rootHash)
	code := accountsDB.innerAccountsAdapter.GetCode(codeHash)
	return code, blockInfo, nil
}

func (accountsDB *accountsDBApiWithHistory) shouldRecreateTrieUnprotected(rootHash []byte) bool {
	return !bytes.Equal(accountsDB.latestRecreatedRootHash, rootHash)
}

func (accountsDB *accountsDBApiWithHistory) recreateTrieUnprotected(options common.RootHashHolder) error {
	err := accountsDB.innerAccountsAdapter.RecreateTrieFromEpoch(options)
	if err != nil {
		return err
	}

	accountsDB.latestRecreatedRootHash = options.GetRootHash()
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (accountsDB *accountsDBApiWithHistory) IsInterfaceNil() bool {
	return accountsDB == nil
}
