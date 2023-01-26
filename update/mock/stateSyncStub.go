package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

// StateSyncStub -
type StateSyncStub struct {
	GetEpochStartMetaBlockCalled  func() (data.MetaHeaderHandler, error)
	GetUnFinishedMetaBlocksCalled func() (map[string]data.MetaHeaderHandler, error)
	SyncAllStateCalled            func(epoch uint32) error
	GetAllTriesCalled             func() (map[string]common.Trie, error)
	GetAllTransactionsCalled      func() (map[string]data.TransactionHandler, error)
	GetAllValidatorsInfoCalled    func() (map[string]*state.ShardValidatorInfo, error)
	GetAllMiniBlocksCalled        func() (map[string]*block.MiniBlock, error)
}

// GetEpochStartMetaBlock -
func (sss *StateSyncStub) GetEpochStartMetaBlock() (data.MetaHeaderHandler, error) {
	if sss.GetEpochStartMetaBlockCalled != nil {
		return sss.GetEpochStartMetaBlockCalled()
	}
	return nil, nil
}

// GetUnFinishedMetaBlocks -
func (sss *StateSyncStub) GetUnFinishedMetaBlocks() (map[string]data.MetaHeaderHandler, error) {
	if sss.GetUnFinishedMetaBlocksCalled != nil {
		return sss.GetUnFinishedMetaBlocksCalled()
	}
	return nil, nil
}

// SyncAllState -
func (sss *StateSyncStub) SyncAllState(epoch uint32) error {
	if sss.SyncAllStateCalled != nil {
		return sss.SyncAllStateCalled(epoch)
	}
	return nil
}

// GetAllTries -
func (sss *StateSyncStub) GetAllTries() (map[string]common.Trie, error) {
	if sss.GetAllTriesCalled != nil {
		return sss.GetAllTriesCalled()
	}
	return nil, nil
}

// GetAllTransactions -
func (sss *StateSyncStub) GetAllTransactions() (map[string]data.TransactionHandler, error) {
	if sss.GetAllTransactionsCalled != nil {
		return sss.GetAllTransactionsCalled()
	}
	return nil, nil
}

// GetAllValidatorsInfo -
func (sss *StateSyncStub) GetAllValidatorsInfo() (map[string]*state.ShardValidatorInfo, error) {
	if sss.GetAllValidatorsInfoCalled != nil {
		return sss.GetAllValidatorsInfoCalled()
	}
	return nil, nil
}

// GetAllMiniBlocks -
func (sss *StateSyncStub) GetAllMiniBlocks() (map[string]*block.MiniBlock, error) {
	if sss.GetAllMiniBlocksCalled != nil {
		return sss.GetAllMiniBlocksCalled()
	}
	return nil, nil
}

// IsInterfaceNil -
func (sss *StateSyncStub) IsInterfaceNil() bool {
	return sss == nil
}
