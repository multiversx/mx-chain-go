package trie

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/errors"
)

type trieStorageManagerInEpoch struct {
	common.StorageManager
	storageManagerExtension storageManager
	mainStorer              dbWithGetFromEpoch
	epoch                   uint32
}

func newTrieStorageManagerInEpoch(tsm common.StorageManager, epoch uint32) (*trieStorageManagerInEpoch, error) {
	sme, ok := tsm.(storageManager)
	if !ok {
		storerType := fmt.Sprintf("%T", tsm)
		return nil, fmt.Errorf("invalid storage manager, type is %s", storerType)
	}

	storer, ok := sme.getStorer().(dbWithGetFromEpoch)
	if !ok {
		storerType := fmt.Sprintf("%T", sme.getStorer())
		return nil, fmt.Errorf("invalid storer, type is %s", storerType)
	}

	return &trieStorageManagerInEpoch{
		StorageManager:          tsm,
		storageManagerExtension: sme,
		mainStorer:              storer,
		epoch:                   epoch,
	}, nil
}

// Get checks all the storers for the given key, and returns it if it is found
func (tsmie *trieStorageManagerInEpoch) Get(key []byte) ([]byte, error) {
	tsmie.storageManagerExtension.lockMutex()
	defer tsmie.storageManagerExtension.unlockMutex()

	if tsmie.storageManagerExtension.isClosed() {
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
