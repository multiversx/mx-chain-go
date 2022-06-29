package state

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	chainData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/common"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type accountsDBApi struct {
	innerAccountsAdapter AccountsAdapter
	chainHandler         chainData.ChainHandler
	mutLastRootHash      sync.RWMutex
	lastRootHash         []byte
}

// NewAccountsDBApi will create a new instance of type accountsDBApi
func NewAccountsDBApi(innerAccountsAdapter AccountsAdapter, chainHandler chainData.ChainHandler) (*accountsDBApi, error) {
	if check.IfNil(innerAccountsAdapter) {
		return nil, ErrNilAccountsAdapter
	}
	if check.IfNil(chainHandler) {
		return nil, ErrNilChainHandler
	}

	return &accountsDBApi{
		innerAccountsAdapter: innerAccountsAdapter,
		chainHandler:         chainHandler,
	}, nil
}

func (accountsDB *accountsDBApi) recreateTrieIfNecessary() error {
	targetRootHash := accountsDB.chainHandler.GetCurrentBlockRootHash()
	if len(targetRootHash) == 0 {
		return fmt.Errorf("%w in accountsDBApi when fetching GetCurrentBlockRootHash", ErrNilRootHash)
	}

	accountsDB.mutLastRootHash.RLock()
	lastRootHash := accountsDB.lastRootHash
	accountsDB.mutLastRootHash.RUnlock()

	if bytes.Equal(lastRootHash, targetRootHash) {
		return nil
	}

	return accountsDB.doRecreateTrie(targetRootHash)
}

func (accountsDB *accountsDBApi) doRecreateTrie(targetRootHash []byte) error {
	accountsDB.mutLastRootHash.Lock()
	defer accountsDB.mutLastRootHash.Unlock()

	// early exit for possible multiple re-entrances here
	lastRootHash := accountsDB.lastRootHash
	if bytes.Equal(lastRootHash, targetRootHash) {
		return nil
	}

	err := accountsDB.innerAccountsAdapter.RecreateTrie(targetRootHash)
	if err != nil {
		accountsDB.lastRootHash = nil
		return err
	}

	accountsDB.lastRootHash = targetRootHash

	return nil
}

// GetExistingAccount will call the inner accountsAdapter method after trying to recreate the trie
func (accountsDB *accountsDBApi) GetExistingAccount(address []byte) (vmcommon.AccountHandler, error) {
	err := accountsDB.recreateTrieIfNecessary()
	if err != nil {
		return nil, err
	}

	return accountsDB.innerAccountsAdapter.GetExistingAccount(address)
}

// GetAccountFromBytes will call the inner accountsAdapter method after trying to recreate the trie
func (accountsDB *accountsDBApi) GetAccountFromBytes(address []byte, accountBytes []byte) (vmcommon.AccountHandler, error) {
	err := accountsDB.recreateTrieIfNecessary()
	if err != nil {
		return nil, err
	}

	return accountsDB.innerAccountsAdapter.GetAccountFromBytes(address, accountBytes)
}

// LoadAccount will call the inner accountsAdapter method after trying to recreate the trie
func (accountsDB *accountsDBApi) LoadAccount(address []byte) (vmcommon.AccountHandler, error) {
	err := accountsDB.recreateTrieIfNecessary()
	if err != nil {
		return nil, err
	}

	return accountsDB.innerAccountsAdapter.LoadAccount(address)
}

// SaveAccount is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApi) SaveAccount(_ vmcommon.AccountHandler) error {
	return ErrOperationNotPermitted
}

// RemoveAccount is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApi) RemoveAccount(_ []byte) error {
	return ErrOperationNotPermitted
}

// CommitInEpoch is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApi) CommitInEpoch(_ uint32, _ uint32) ([]byte, error) {
	return nil, ErrOperationNotPermitted
}

// Commit is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApi) Commit() ([]byte, error) {
	return nil, ErrOperationNotPermitted
}

// JournalLen will always return 0
func (accountsDB *accountsDBApi) JournalLen() int {
	return 0
}

// RevertToSnapshot is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApi) RevertToSnapshot(_ int) error {
	return ErrOperationNotPermitted
}

// GetCode will call the inner accountsAdapter method after trying to recreate the trie
func (accountsDB *accountsDBApi) GetCode(codeHash []byte) []byte {
	err := accountsDB.recreateTrieIfNecessary()
	if err != nil {
		return nil
	}

	return accountsDB.innerAccountsAdapter.GetCode(codeHash)
}

// RootHash will return last loaded root hash
func (accountsDB *accountsDBApi) RootHash() ([]byte, error) {
	accountsDB.mutLastRootHash.RLock()
	defer accountsDB.mutLastRootHash.RUnlock()

	if accountsDB.lastRootHash == nil {
		return nil, ErrNilRootHash
	}

	return accountsDB.lastRootHash, nil
}

// RecreateTrie is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApi) RecreateTrie(_ []byte) error {
	return ErrOperationNotPermitted
}

// PruneTrie is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApi) PruneTrie(_ []byte, _ TriePruningIdentifier, _ PruningHandler) {
}

// CancelPrune is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApi) CancelPrune(_ []byte, _ TriePruningIdentifier) {
}

// SnapshotState is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApi) SnapshotState(_ []byte) {
}

// SetStateCheckpoint is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApi) SetStateCheckpoint(_ []byte) {
}

// IsPruningEnabled will call the inner accountsAdapter method
func (accountsDB *accountsDBApi) IsPruningEnabled() bool {
	return accountsDB.innerAccountsAdapter.IsPruningEnabled()
}

// GetAllLeaves will call the inner accountsAdapter method after trying to recreate the trie
func (accountsDB *accountsDBApi) GetAllLeaves(leavesChannel chan core.KeyValueHolder, ctx context.Context, rootHash []byte) error {
	err := accountsDB.recreateTrieIfNecessary()
	if err != nil {
		return err
	}

	return accountsDB.innerAccountsAdapter.GetAllLeaves(leavesChannel, ctx, rootHash)
}

// RecreateAllTries is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApi) RecreateAllTries(_ []byte) (map[string]common.Trie, error) {
	return nil, ErrOperationNotPermitted
}

// GetTrie will call the inner accountsAdapter method
func (accountsDB *accountsDBApi) GetTrie(rootHash []byte) (common.Trie, error) {
	return accountsDB.innerAccountsAdapter.GetTrie(rootHash)
}

// GetStackDebugFirstEntry will call the inner accountsAdapter method
func (accountsDB *accountsDBApi) GetStackDebugFirstEntry() []byte {
	return accountsDB.innerAccountsAdapter.GetStackDebugFirstEntry()
}

// Close will handle the closing of the underlying components
func (accountsDB *accountsDBApi) Close() error {
	return accountsDB.innerAccountsAdapter.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (accountsDB *accountsDBApi) IsInterfaceNil() bool {
	return accountsDB == nil
}
