package state

import (
	"context"
	"errors"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var errNotImplemented = errors.New("not implemented")

// AccountsStub -
type AccountsStub struct {
	GetExistingAccountCalled      func(addressContainer []byte) (vmcommon.AccountHandler, error)
	GetAccountFromBytesCalled     func(address []byte, accountBytes []byte) (vmcommon.AccountHandler, error)
	LoadAccountCalled             func(container []byte) (vmcommon.AccountHandler, error)
	SaveAccountCalled             func(account vmcommon.AccountHandler) error
	RemoveAccountCalled           func(addressContainer []byte) error
	CommitCalled                  func() ([]byte, error)
	CommitInEpochCalled           func(uint32, uint32) ([]byte, error)
	JournalLenCalled              func() int
	RevertToSnapshotCalled        func(snapshot int) error
	RootHashCalled                func() ([]byte, error)
	RecreateTrieCalled            func(rootHash []byte) error
	RecreateTrieFromEpochCalled   func(options common.RootHashHolder) error
	PruneTrieCalled               func(rootHash []byte, identifier state.TriePruningIdentifier, handler state.PruningHandler)
	CancelPruneCalled             func(rootHash []byte, identifier state.TriePruningIdentifier)
	SnapshotStateCalled           func(rootHash []byte, epoch uint32)
	SetStateCheckpointCalled      func(rootHash []byte)
	IsPruningEnabledCalled        func() bool
	GetAllLeavesCalled            func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeafParser common.TrieLeafParser) error
	RecreateAllTriesCalled        func(rootHash []byte) (map[string]common.Trie, error)
	GetCodeCalled                 func([]byte) []byte
	GetTrieCalled                 func([]byte) (common.Trie, error)
	GetStackDebugFirstEntryCalled func() []byte
	GetAccountWithBlockInfoCalled func(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error)
	GetCodeWithBlockInfoCalled    func(codeHash []byte, options common.RootHashHolder) ([]byte, common.BlockInfo, error)
	CloseCalled                   func() error
	SetSyncerCalled               func(syncer state.AccountsDBSyncer) error
	StartSnapshotIfNeededCalled   func() error
}

// CleanCache -
func (as *AccountsStub) CleanCache() {
}

// SetSyncer -
func (as *AccountsStub) SetSyncer(syncer state.AccountsDBSyncer) error {
	if as.SetSyncerCalled != nil {
		return as.SetSyncerCalled(syncer)
	}

	return nil
}

// StartSnapshotIfNeeded -
func (as *AccountsStub) StartSnapshotIfNeeded() error {
	if as.StartSnapshotIfNeededCalled != nil {
		return as.StartSnapshotIfNeededCalled()
	}

	return nil
}

// GetTrie -
func (as *AccountsStub) GetTrie(codeHash []byte) (common.Trie, error) {
	if as.GetTrieCalled != nil {
		return as.GetTrieCalled(codeHash)
	}

	return nil, nil
}

// GetCode -
func (as *AccountsStub) GetCode(codeHash []byte) []byte {
	if as.GetCodeCalled != nil {
		return as.GetCodeCalled(codeHash)
	}
	return nil
}

// RecreateAllTries -
func (as *AccountsStub) RecreateAllTries(rootHash []byte) (map[string]common.Trie, error) {
	if as.RecreateAllTriesCalled != nil {
		return as.RecreateAllTriesCalled(rootHash)
	}
	return nil, nil
}

// LoadAccount -
func (as *AccountsStub) LoadAccount(address []byte) (vmcommon.AccountHandler, error) {
	if as.LoadAccountCalled != nil {
		return as.LoadAccountCalled(address)
	}
	return NewAccountWrapMock(address), nil
}

// SaveAccount -
func (as *AccountsStub) SaveAccount(account vmcommon.AccountHandler) error {
	if as.SaveAccountCalled != nil {
		return as.SaveAccountCalled(account)
	}
	return nil
}

// GetAllLeaves -
func (as *AccountsStub) GetAllLeaves(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeafParser common.TrieLeafParser) error {
	if as.GetAllLeavesCalled != nil {
		return as.GetAllLeavesCalled(leavesChannels, ctx, rootHash, trieLeafParser)
	}
	return nil
}

// Commit -
func (as *AccountsStub) Commit() ([]byte, error) {
	if as.CommitCalled != nil {
		return as.CommitCalled()
	}

	return nil, errNotImplemented
}

// GetExistingAccount -
func (as *AccountsStub) GetExistingAccount(addressContainer []byte) (vmcommon.AccountHandler, error) {
	if as.GetExistingAccountCalled != nil {
		return as.GetExistingAccountCalled(addressContainer)
	}

	return nil, errNotImplemented
}

// GetAccountFromBytes -
func (as *AccountsStub) GetAccountFromBytes(address []byte, accountBytes []byte) (vmcommon.AccountHandler, error) {
	if as.GetAccountFromBytesCalled != nil {
		return as.GetAccountFromBytesCalled(address, accountBytes)
	}

	return nil, errNotImplemented
}

// JournalLen -
func (as *AccountsStub) JournalLen() int {
	if as.JournalLenCalled != nil {
		return as.JournalLenCalled()
	}

	return 0
}

// RemoveAccount -
func (as *AccountsStub) RemoveAccount(addressContainer []byte) error {
	if as.RemoveAccountCalled != nil {
		return as.RemoveAccountCalled(addressContainer)
	}

	return errNotImplemented
}

// RevertToSnapshot -
func (as *AccountsStub) RevertToSnapshot(snapshot int) error {
	if as.RevertToSnapshotCalled != nil {
		return as.RevertToSnapshotCalled(snapshot)
	}

	return errNotImplemented
}

// RootHash -
func (as *AccountsStub) RootHash() ([]byte, error) {
	if as.RootHashCalled != nil {
		return as.RootHashCalled()
	}

	return nil, errNotImplemented
}

// RecreateTrie -
func (as *AccountsStub) RecreateTrie(rootHash []byte) error {
	if as.RecreateTrieCalled != nil {
		return as.RecreateTrieCalled(rootHash)
	}

	return errNotImplemented
}

// RecreateTrieFromEpoch -
func (as *AccountsStub) RecreateTrieFromEpoch(options common.RootHashHolder) error {
	if as.RecreateTrieFromEpochCalled != nil {
		return as.RecreateTrieFromEpochCalled(options)
	}

	return errNotImplemented
}

// PruneTrie -
func (as *AccountsStub) PruneTrie(rootHash []byte, identifier state.TriePruningIdentifier, handler state.PruningHandler) {
	as.PruneTrieCalled(rootHash, identifier, handler)
}

// CancelPrune -
func (as *AccountsStub) CancelPrune(rootHash []byte, identifier state.TriePruningIdentifier) {
	if as.CancelPruneCalled != nil {
		as.CancelPruneCalled(rootHash, identifier)
	}
}

// SnapshotState -
func (as *AccountsStub) SnapshotState(rootHash []byte, epoch uint32) {
	if as.SnapshotStateCalled != nil {
		as.SnapshotStateCalled(rootHash, epoch)
	}
}

// SetStateCheckpoint -
func (as *AccountsStub) SetStateCheckpoint(rootHash []byte) {
	if as.SetStateCheckpointCalled != nil {
		as.SetStateCheckpointCalled(rootHash)
	}
}

// IsPruningEnabled -
func (as *AccountsStub) IsPruningEnabled() bool {
	if as.IsPruningEnabledCalled != nil {
		return as.IsPruningEnabledCalled()
	}

	return false
}

// CommitInEpoch -
func (as *AccountsStub) CommitInEpoch(currentEpoch uint32, epochToCommit uint32) ([]byte, error) {
	if as.CommitInEpochCalled != nil {
		return as.CommitInEpochCalled(currentEpoch, epochToCommit)
	}

	return nil, nil
}

// GetStackDebugFirstEntry -
func (as *AccountsStub) GetStackDebugFirstEntry() []byte {
	if as.GetStackDebugFirstEntryCalled != nil {
		return as.GetStackDebugFirstEntryCalled()
	}

	return nil
}

// GetAccountWithBlockInfo -
func (as *AccountsStub) GetAccountWithBlockInfo(address []byte, options common.RootHashHolder) (vmcommon.AccountHandler, common.BlockInfo, error) {
	if as.GetAccountWithBlockInfoCalled != nil {
		return as.GetAccountWithBlockInfoCalled(address, options)
	}

	return nil, nil, nil
}

// GetCodeWithBlockInfo -
func (as *AccountsStub) GetCodeWithBlockInfo(codeHash []byte, options common.RootHashHolder) ([]byte, common.BlockInfo, error) {
	if as.GetCodeWithBlockInfoCalled != nil {
		return as.GetCodeWithBlockInfoCalled(codeHash, options)
	}

	return nil, nil, nil
}

// Close -
func (as *AccountsStub) Close() error {
	if as.CloseCalled != nil {
		return as.CloseCalled()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (as *AccountsStub) IsInterfaceNil() bool {
	return as == nil
}
