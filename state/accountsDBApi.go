package state

import (
	"context"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type accountsDBApi struct {
	innerAccountsAdapter        AccountsAdapter
	blockInfoProvider           BlockInfoProvider
	mutBlockInfoOfRecreatedTrie sync.RWMutex
	blockInfo                   common.BlockInfo
}

// NewAccountsDBApi will create a new instance of type accountsDBApi
func NewAccountsDBApi(innerAccountsAdapter AccountsAdapter, blockInfoProvider BlockInfoProvider) (*accountsDBApi, error) {
	if check.IfNil(innerAccountsAdapter) {
		return nil, ErrNilAccountsAdapter
	}
	if check.IfNil(blockInfoProvider) {
		return nil, ErrNilBlockInfoProvider
	}

	return &accountsDBApi{
		innerAccountsAdapter: innerAccountsAdapter,
		blockInfoProvider:    blockInfoProvider,
	}, nil
}

func (accountsDB *accountsDBApi) recreateTrieIfNecessary() (common.BlockInfo, error) {
	newBlockInfo := accountsDB.blockInfoProvider.GetBlockInfo()
	if check.IfNil(newBlockInfo) {
		return nil, fmt.Errorf("%w in accountsDBApi.recreateTrieIfNecessary", ErrNilBlockInfo)
	}
	if len(newBlockInfo.GetRootHash()) == 0 {
		return nil, fmt.Errorf("%w in accountsDBApi.recreateTrieIfNecessary", ErrNilRootHash)
	}

	accountsDB.mutBlockInfoOfRecreatedTrie.RLock()
	currentBlockInfo := accountsDB.blockInfo
	accountsDB.mutBlockInfoOfRecreatedTrie.RUnlock()

	if newBlockInfo.Equal(currentBlockInfo) {
		return currentBlockInfo, nil
	}

	return accountsDB.doRecreateTrieWithBlockInfo(newBlockInfo)
}

func (accountsDB *accountsDBApi) doRecreateTrieWithBlockInfo(newBlockInfo common.BlockInfo) (common.BlockInfo, error) {
	accountsDB.mutBlockInfoOfRecreatedTrie.Lock()
	defer accountsDB.mutBlockInfoOfRecreatedTrie.Unlock()

	// early exit for possible multiple re-entrances here
	currentBlockInfo := accountsDB.blockInfo
	if newBlockInfo.Equal(currentBlockInfo) {
		return currentBlockInfo, nil
	}

	err := accountsDB.innerAccountsAdapter.RecreateTrie(newBlockInfo.GetRootHash())
	if err != nil {
		accountsDB.blockInfo = nil
		return nil, err
	}

	accountsDB.blockInfo = newBlockInfo

	return newBlockInfo, nil
}

// GetExistingAccount will call the inner accountsAdapter method after trying to recreate the trie
func (accountsDB *accountsDBApi) GetExistingAccount(address []byte) (vmcommon.AccountHandler, error) {
	account, _, err := accountsDB.GetAccountWithBlockInfo(address)

	return account, err
}

// GetAccountFromBytes will call the inner accountsAdapter method after trying to recreate the trie
func (accountsDB *accountsDBApi) GetAccountFromBytes(address []byte, accountBytes []byte) (vmcommon.AccountHandler, error) {
	_, err := accountsDB.recreateTrieIfNecessary()
	if err != nil {
		return nil, err
	}

	return accountsDB.innerAccountsAdapter.GetAccountFromBytes(address, accountBytes)
}

// LoadAccount will call the inner accountsAdapter method after trying to recreate the trie
func (accountsDB *accountsDBApi) LoadAccount(address []byte) (vmcommon.AccountHandler, error) {
	_, err := accountsDB.recreateTrieIfNecessary()
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
	code, _, err := accountsDB.GetCodeWithBlockInfo(codeHash)
	if err != nil {
		log.Warn("accountsDBApi.GetCode", "error", err)
	}

	return code
}

// RootHash will return last loaded root hash
func (accountsDB *accountsDBApi) RootHash() ([]byte, error) {
	accountsDB.mutBlockInfoOfRecreatedTrie.RLock()
	defer accountsDB.mutBlockInfoOfRecreatedTrie.RUnlock()

	blockInfo := accountsDB.blockInfo
	if check.IfNil(blockInfo) || blockInfo.GetRootHash() == nil {
		return nil, ErrNilRootHash
	}

	return blockInfo.GetRootHash(), nil
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
	_, err := accountsDB.recreateTrieIfNecessary()
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

// GetAccountWithBlockInfo returns the account and the associated block info
func (accountsDB *accountsDBApi) GetAccountWithBlockInfo(address []byte) (vmcommon.AccountHandler, common.BlockInfo, error) {
	_, err := accountsDB.recreateTrieIfNecessary()
	if err != nil {
		return nil, nil, err
	}

	// We hold the read mutex over both <getting the current block info> AND <getting the account>.
	// -> desired side-effect: any concurrent "recreateTrieIfNecessary()" waits until the mutex is released.
	// -> under normal circumstances (node already synchronized), performance of GET account should not be impacted.
	// -> why don't we create a critical section that spans over the whole function? Because we'd like "recreateTrieIfNecessary()"
	// to happen as soon (and as often) as possible.
	accountsDB.mutBlockInfoOfRecreatedTrie.RLock()
	defer accountsDB.mutBlockInfoOfRecreatedTrie.RUnlock()

	blockInfo := accountsDB.blockInfo

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
func (accountsDB *accountsDBApi) GetCodeWithBlockInfo(codeHash []byte) ([]byte, common.BlockInfo, error) {
	blockInfo, err := accountsDB.recreateTrieIfNecessary()
	if err != nil {
		return nil, nil, err
	}

	return accountsDB.innerAccountsAdapter.GetCode(codeHash), blockInfo, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (accountsDB *accountsDBApi) IsInterfaceNil() bool {
	return accountsDB == nil
}
