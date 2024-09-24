package receiptslog

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie"
)

// RecreateTrieFromDB will recreate the trie from the provided storer
func (ti *trieInteractor) RecreateTrieFromDB(rootHash []byte, db storage.Storer) (common.Trie, error) {
	storageManagerStub := &storageManager.StorageManagerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return db.Get(key)
		},
	}

	localTrie, err := trie.NewTrie(storageManagerStub, ti.marshaller, ti.hasher, ti.enableEpochsHandler, maxTrieLevelInMemory)
	if err != nil {
		return nil, err
	}

	rootHashHolder := holders.NewDefaultRootHashesHolder(rootHash)
	return localTrie.Recreate(rootHashHolder)
}
