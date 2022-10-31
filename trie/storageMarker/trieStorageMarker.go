package storageMarker

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
)

var log = logger.GetOrCreate("trie")

type trieStorageMarker struct {
}

// NewTrieStorageMarker creates a new instance of trieStorageMarker
func NewTrieStorageMarker() *trieStorageMarker {
	return &trieStorageMarker{}
}

// MarkStorerAsSyncedAndActive marks the storage as synced and active
func (sm *trieStorageMarker) MarkStorerAsSyncedAndActive(storer common.StorageManager) {
	epoch, err := storer.GetLatestStorageEpoch()
	if err != nil {
		log.Error("getLatestStorageEpoch error", "error", err)
	}

	err = storer.Put([]byte(common.TrieSyncedKey), []byte(common.TrieSyncedVal))
	if err != nil {
		log.Error("error while putting trieSynced value into main storer after sync", "error", err)
	}
	log.Debug("set trieSyncedKey in epoch", "epoch", epoch)

	lastEpoch := epoch - 1
	if epoch == 0 {
		lastEpoch = 0
	}

	err = storer.PutInEpochWithoutCache([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal), lastEpoch)
	if err != nil {
		log.Error("error while putting activeDB value into main storer after sync", "error", err)
	}
	log.Debug("set activeDB in epoch", "epoch", lastEpoch)
}
