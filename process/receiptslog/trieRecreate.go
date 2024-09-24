package receiptslog

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/trie"
)

// RecreateTrieFromDB will recreate the trie from the provided storer
func (ti *trieInteractor) RecreateTrieFromDB(rootHash []byte, db storage.Persister) (common.Trie, error) {
	storageManager, err := NewStorageManagerOnlyGet(db)
	if err != nil {
		return nil, err
	}

	localTrie, err := trie.NewTrie(storageManager, ti.marshaller, ti.hasher, ti.enableEpochsHandler, maxTrieLevelInMemory)
	if err != nil {
		return nil, err
	}

	rootHashHolder := holders.NewDefaultRootHashesHolder(rootHash)
	return localTrie.Recreate(rootHashHolder)
}
