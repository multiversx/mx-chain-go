package mock

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

// TrieFactoryStub -
type TrieFactoryStub struct {
	CreateCalled func(config config.StorageConfig, s string, b bool) (temporary.StorageManager, temporary.Trie, error)
}

// Create -
func (t *TrieFactoryStub) Create(config config.StorageConfig, s string, b bool) (temporary.StorageManager, temporary.Trie, error) {
	if t.CreateCalled != nil {
		return t.CreateCalled(config, s, b)
	}
	return nil, nil, nil
}

// IsInterfaceNil -
func (t *TrieFactoryStub) IsInterfaceNil() bool {
	return t == nil
}
