package mock

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data"
)

// TrieFactoryStub -
type TrieFactoryStub struct {
	CreateCalled func(config config.StorageConfig, s string, b bool) (data.StorageManager, data.Trie, error)
}

// Create -
func (t *TrieFactoryStub) Create(config config.StorageConfig, s string, b bool) (data.StorageManager, data.Trie, error) {
	if t.CreateCalled != nil {
		return t.CreateCalled(config, s, b)
	}
	return nil, nil, nil
}

// IsInterfaceNil -
func (t *TrieFactoryStub) IsInterfaceNil() bool {
	return t == nil
}
