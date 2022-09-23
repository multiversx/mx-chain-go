package trie

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/errors"
)

// numEpochsToVerify needs to be at least 2 due to a snapshotting edge-case.
// The trie nodes modified between when a start of epoch block is committed until it is notarized by meta
// are not copied during snapshot to the new storer. So in order to have access to all trie data related
// to a certain root hash, both current storer and previous storer need to be verified.
const numEpochsToVerify = uint32(2)

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
		return nil, fmt.Errorf("invalid storage manager, type is %T", storageManager.GetBaseTrieStorageManager())
	}

	storer, ok := tsm.mainStorer.(dbWithGetFromEpoch)
	if !ok {
		return nil, fmt.Errorf("invalid storer, type is %T", tsm.mainStorer)
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

	for i := uint32(0); i < numEpochsToVerify; i++ {
		if i > tsmie.epoch {
			break
		}
		epoch := tsmie.epoch - i

		val, err := tsmie.mainStorer.GetFromEpoch(key, epoch)
		treatGetFromEpochError(err, epoch)
		if len(val) != 0 {
			return val, nil
		}
	}

	return nil, ErrKeyNotFound
}

func treatGetFromEpochError(err error, epoch uint32) {
	if err == nil {
		return
	}

	if errors.IsClosingError(err) {
		log.Debug("trieStorageManagerInEpoch closing err", "error", err.Error(), "epoch", epoch)
		return
	}

	log.Warn("trieStorageManagerInEpoch", "error", err.Error(), "epoch", epoch)
}
