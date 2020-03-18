package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type EpochStartTriesSyncHandlerMock struct {
	SyncTriesFromCalled func(meta *block.MetaBlock, waitTime time.Duration) error
	GetTriesCalled      func() (map[string]data.Trie, error)
}

func (es *EpochStartTriesSyncHandlerMock) SyncTriesFrom(meta *block.MetaBlock, waitTime time.Duration) error {
	if es.SyncTriesFromCalled != nil {
		return es.SyncTriesFromCalled(meta, waitTime)
	}
	return nil
}

func (es *EpochStartTriesSyncHandlerMock) GetTries() (map[string]data.Trie, error) {
	if es.GetTriesCalled != nil {
		return es.GetTriesCalled()
	}
	return nil, nil
}

func (es *EpochStartTriesSyncHandlerMock) IsInterfaceNil() bool {
	return es == nil
}
