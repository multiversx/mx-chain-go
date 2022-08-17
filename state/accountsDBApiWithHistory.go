package state

import (
	"context"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/holders"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type accountsDBApiWithHistory struct {
	innerAccountsAdapter AccountsAdapter
	mutRecreateAndGet    sync.Mutex
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

// IsPruningEnabled will return false
func (accountsDB *accountsDBApiWithHistory) IsPruningEnabled() bool {
	return false
}

// GetAllLeaves will return an error
func (accountsDB *accountsDBApiWithHistory) GetAllLeaves(leavesChannel chan core.KeyValueHolder, ctx context.Context, rootHash []byte) error {
	return ErrOperationNotPermitted
}

// RecreateAllTries is a not permitted operation in this implementation and thus, will return an error
func (accountsDB *accountsDBApiWithHistory) RecreateAllTries(_ []byte) (map[string]common.Trie, error) {
	return nil, ErrOperationNotPermitted
}

// GetTrie is not implemented
func (accountsDB *accountsDBApiWithHistory) GetTrie(rootHash []byte) (common.Trie, error) {
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
func (accountsDB *accountsDBApiWithHistory) GetAccountWithBlockInfo(address []byte, options common.GetAccountsStateOptions) (vmcommon.AccountHandler, common.BlockInfo, error) {
	blockInfo := holders.NewBlockInfo(nil, 0, options.GetBlockRootHash())

	// We cannot allow concurrent requests to <recreate a trie & load state>.
	accountsDB.mutRecreateAndGet.Lock()
	defer accountsDB.mutRecreateAndGet.Unlock()

	err := accountsDB.recreateTrie(options.GetBlockRootHash())
	if err != nil {
		return nil, nil, err
	}

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
func (accountsDB *accountsDBApiWithHistory) GetCodeWithBlockInfo(codeHash []byte, options common.GetAccountsStateOptions) ([]byte, common.BlockInfo, error) {
	blockInfo := holders.NewBlockInfo(nil, 0, options.GetBlockRootHash())

	// We cannot allow concurrent requests to <recreate a trie & load state>.
	accountsDB.mutRecreateAndGet.Lock()
	defer accountsDB.mutRecreateAndGet.Unlock()

	err := accountsDB.recreateTrie(options.GetBlockRootHash())
	if err != nil {
		return nil, nil, err
	}

	code := accountsDB.innerAccountsAdapter.GetCode(codeHash)
	return code, blockInfo, nil
}

func (accountsDB *accountsDBApiWithHistory) recreateTrie(rootHash []byte) error {
	// TODO: perhaps we can optimize this in the future: check the latest rootHash used for re-creation, and skip re-creation, if possible.
	return accountsDB.innerAccountsAdapter.RecreateTrie(rootHash)
}

// IsInterfaceNil returns true if there is no value under the interface
func (accountsDB *accountsDBApiWithHistory) IsInterfaceNil() bool {
	return accountsDB == nil
}
