package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// EpochStartTriesSyncHandlerMock -
type EpochStartTriesSyncHandlerMock struct {
	SyncTriesFromCalled func(meta data.HeaderHandler) error
	GetTriesCalled      func() (map[string]data.Trie, error)
}

// SyncTriesFrom -
func (es *EpochStartTriesSyncHandlerMock) SyncTriesFrom(meta data.HeaderHandler) error {
	if es.SyncTriesFromCalled != nil {
		return es.SyncTriesFromCalled(meta)
	}
	return nil
}

// GetTries -
func (es *EpochStartTriesSyncHandlerMock) GetTries() (map[string]data.Trie, error) {
	if es.GetTriesCalled != nil {
		return es.GetTriesCalled()
	}
	return nil, nil
}

// IsInterfaceNil -
func (es *EpochStartTriesSyncHandlerMock) IsInterfaceNil() bool {
	return es == nil
}
