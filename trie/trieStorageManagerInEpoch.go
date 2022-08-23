package trie

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/errors"
)

type trieStorageManagerInEpoch struct {
	*trieStorageManager
	mainStorer dbWithGetFromEpoch
	epoch      uint32
}

func newTrieStorageManagerInEpoch(storageManager common.StorageManager, epoch uint32) (*trieStorageManagerInEpoch, error) {
	if check.IfNil(storageManager) {
		return nil, ErrNilTrieStorage
	}

	tsm, ok := storageManager.GetBaseTrieStorageManager().(*trieStorageManager)
	if !ok {
		storerType := fmt.Sprintf("%T", storageManager.GetBaseTrieStorageManager())
		return nil, fmt.Errorf("invalid storage manager, type is %s", storerType)
	}

	storer, ok := tsm.mainStorer.(dbWithGetFromEpoch)
	if !ok {
		storerType := fmt.Sprintf("%T", tsm.mainStorer)
		return nil, fmt.Errorf("invalid storer, type is %s", storerType)
	}

	return &trieStorageManagerInEpoch{
		trieStorageManager: tsm,
		mainStorer:         storer,
		epoch:              epoch,
	}, nil
}

// Get checks all the storers for the given key, and returns it if it is found
func (tsmie *trieStorageManagerInEpoch) Get(key []byte) ([]byte, error) {
	tsmie.storageOperationMutex.Lock()
	defer tsmie.storageOperationMutex.Unlock()

	if tsmie.closed {
		log.Debug("trieStorageManagerInEpoch get context closing", "key", key)
		return nil, errors.ErrContextClosing
	}

	val, err := tsmie.mainStorer.GetFromEpoch(key, tsmie.epoch)
	if errors.IsClosingError(err) {
		return nil, err
	}
	if len(val) != 0 {
		return val, nil
	}

	return nil, ErrKeyNotFound
}
