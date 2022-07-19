package trie

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/common"
)

func (ts *trieSyncer) trieNodeIntercepted(hash []byte, val interface{}) {
	ts.mutOperation.Lock()
	defer ts.mutOperation.Unlock()

	log.Trace("trie node intercepted", "hash", hash)

	n, ok := ts.nodesForTrie[string(hash)]
	if !ok || n.received {
		return
	}

	interceptedNode, err := trieNode(val, marshalizer, hasherMock)
	if err != nil {
		return
	}

	ts.nodesForTrie[string(hash)] = trieNodeInfo{
		trieNode: interceptedNode,
		received: true,
	}
}

// PruningBlockingOperations -
func (tsm *trieStorageManagerWithoutCheckpoints) PruningBlockingOperations() uint32 {
	return tsm.pruningBlockingOps
}

// WaitForOperationToComplete -
func WaitForOperationToComplete(tsm common.StorageManager) {
	for tsm.IsPruningBlocked() {
		time.Sleep(10 * time.Millisecond)
	}
}

// GetFromCheckpoint -
func (tsm *trieStorageManager) GetFromCheckpoint(key []byte) ([]byte, error) {
	tsm.storageOperationMutex.Lock()
	defer tsm.storageOperationMutex.Unlock()

	return tsm.checkpointsStorer.Get(key)
}

// CreateSmallTestTrieAndStorageManager -
func CreateSmallTestTrieAndStorageManager() (*patriciaMerkleTrie, *trieStorageManager) {
	tr, trieStorage := newEmptyTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("dogglesworth"), []byte("cat"))

	_ = tr.Commit()

	return tr, trieStorage
}

// GetDirtyHashes -
func GetDirtyHashes(tr common.Trie) common.ModifiedHashes {
	hashes, _ := tr.GetAllHashes()
	dirtyHashes := make(common.ModifiedHashes)
	for _, hash := range hashes {
		dirtyHashes[string(hash)] = struct{}{}
	}

	return dirtyHashes
}

// WriteInChanNonBlocking -
func WriteInChanNonBlocking(errChan chan error, err error) {
	writeInChanNonBlocking(errChan, err)
}
