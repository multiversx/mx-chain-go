package state

import (
	"context"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type accountsDBApi struct {
	innerAccountsAdapter      AccountsAdapter
	blockInfoProvider         BlockInfoProvider
	mutRecreatedTrieBlockInfo sync.RWMutex
	blockInfo                 common.BlockInfo
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

	accountsDB.mutRecreatedTrieBlockInfo.RLock()
	currentBlockInfo := accountsDB.blockInfo
	accountsDB.mutRecreatedTrieBlockInfo.RUnlock()

	if newBlockInfo.Equal(currentBlockInfo) {
		return currentBlockInfo, nil
	}

	return accountsDB.doRecreateTrieWithBlockInfo(newBlockInfo)
}

func (accountsDB *accountsDBApi) doRecreateTrieWithBlockInfo(newBlockInfo common.BlockInfo) (common.BlockInfo, error) {
	accountsDB.mutRecreatedTrieBlockInfo.Lock()
	defer accountsDB.mutRecreatedTrieBlockInfo.Unlock()

	// early exit for possible multiple re-entrances here
	currentBlockInfo := accountsDB.blockInfo
	if newBlockInfo.Equal(currentBlockInfo) {
		return currentBlockInfo, nil
	}

	rootHashHolder := holders.NewDefaultRootHashesHolder(newBlockInfo.GetRootHash())
	err := accountsDB.innerAccountsAdapter.RecreateTrie(rootHashHolder)
	if err != nil {
		accountsDB.blockInfo = nil
		return nil, err
	}

	accountsDB.blockInfo = newBlockInfo

	return newBlockInfo, nil
}

// SetSyncer returns nil for this implementation
func (accountsDB *accountsDBApi) SetSyncer(_ AccountsDBSyncer) error {
	return nil
}

// StartSnapshotIfNeeded returns nil for this implementation
func (accountsDB *accountsDBApi) StartSnapshotIfNeeded() error {
	return nil
}

// GetExistingAccount will call the inner accountsAdapter method after trying to recreate the trie
func (accountsDB *accountsDBApi) GetExistingAccount(address []byte) (vmcommon.AccountHandler, error) {
	account, _, err := accountsDB.GetAccountWithBlockInfo(address, holders.NewRootHashHolderAsEmpty())

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
	code, _, err := accountsDB.GetCodeWithBlockInfo(codeHash, holders.NewRootHashHolderAsEmpty())
	if err != nil {
		log.Warn("accountsDBApi.GetCode", "error", err)
	}

	return code
}

// RootHash will return last loaded root hash
func (accountsDB *accountsDBApi) RootHash() ([]byte, error) {
	accountsDB.mutRecreatedTrieBlockInfo.RLock()
	defer accountsDB.mutRecreatedTrieBlockInfo.RUnlock()

	blockInfo := accountsDB.blockInfo
	if check.IfNil(blockInfo) || blockInfo.GetRootHash() == nil {
		return nil, ErrNilRootHash
	}

	return blockInfo.GetRootHash(), nil
}

// RecreateTrie is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApi) RecreateTrie(options common.RootHashHolder) error {
	accountsDB.mutRecreatedTrieBlockInfo.Lock()
	defer accountsDB.mutRecreatedTrieBlockInfo.Unlock()

	if check.IfNil(options) {
		return ErrNilRootHashHolder
	}

	newBlockInfo := holders.NewBlockInfo([]byte{}, 0, options.GetRootHash())
	if newBlockInfo.Equal(accountsDB.blockInfo) {
		return nil
	}

	err := accountsDB.innerAccountsAdapter.RecreateTrie(options)
	if err != nil {
		accountsDB.blockInfo = nil
		return err
	}

	accountsDB.blockInfo = newBlockInfo

	return nil
}

// PruneTrie is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApi) PruneTrie(_ []byte, _ TriePruningIdentifier, _ PruningHandler) {
}

// CancelPrune is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApi) CancelPrune(_ []byte, _ TriePruningIdentifier) {
}

// SnapshotState is a not permitted operation in this implementation and thus, does nothing
func (accountsDB *accountsDBApi) SnapshotState(_ []byte, _ uint32) {
}

// IsPruningEnabled will call the inner accountsAdapter method
func (accountsDB *accountsDBApi) IsPruningEnabled() bool {
	return accountsDB.innerAccountsAdapter.IsPruningEnabled()
}

// GetAllLeaves will call the inner accountsAdapter method after trying to recreate the trie
func (accountsDB *accountsDBApi) GetAllLeaves(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeafParser common.TrieLeafParser) error {
	_, err := accountsDB.recreateTrieIfNecessary()
	if err != nil {
		return err
	}

	return accountsDB.innerAccountsAdapter.GetAllLeaves(leavesChannels, ctx, rootHash, trieLeafParser)
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
func (accountsDB *accountsDBApi) GetAccountWithBlockInfo(address []byte, _ common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
	_, err := accountsDB.recreateTrieIfNecessary()
	if err != nil {
		return nil, nil, err
	}

	// We hold the read mutex over both <getting the current block info> AND <getting the account>.
	// -> desired side-effect: any concurrent "recreateTrieIfNecessary()" waits until the mutex is released.
	// -> under normal circumstances (node already synchronized), performance of GET account should not be impacted.
	accountsDB.mutRecreatedTrieBlockInfo.RLock()
	defer accountsDB.mutRecreatedTrieBlockInfo.RUnlock()

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
func (accountsDB *accountsDBApi) GetCodeWithBlockInfo(codeHash []byte, _ common.RootHashHolder) ([]byte, common.BlockInfo, error) {
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
