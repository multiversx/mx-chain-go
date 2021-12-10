package trie

import "fmt"

type snapshotTrieStorageManager struct {
	*trieStorageManager
	mainSnapshotStorer snapshotPruningStorer
}

func newSnapshotTrieStorageManager(tsm *trieStorageManager) (*snapshotTrieStorageManager, error) {
	storer, ok := tsm.mainStorer.(snapshotPruningStorer)
	if !ok {
		return nil, fmt.Errorf("invalid storer type")
	}

	return &snapshotTrieStorageManager{
		trieStorageManager: tsm,
		mainSnapshotStorer: storer,
	}, nil
}

//Get checks all the storers for the given key, and returns it if it is found
func (stsm *snapshotTrieStorageManager) Get(key []byte) ([]byte, error) {
	stsm.storageOperationMutex.Lock()
	defer stsm.storageOperationMutex.Unlock()

	val, _ := stsm.mainSnapshotStorer.GetFromOldEpochsWithoutCache(key)
	if len(val) != 0 {
		return val, nil
	}

	return stsm.getFromOtherStorers(key)
}

// Put adds the given value to the main storer
func (stsm *snapshotTrieStorageManager) Put(key, data []byte) error {
	stsm.storageOperationMutex.Lock()
	defer stsm.storageOperationMutex.Unlock()

	return stsm.mainSnapshotStorer.PutWithoutCache(key, data)
}
