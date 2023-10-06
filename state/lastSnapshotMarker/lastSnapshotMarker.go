package lastSnapshotMarker

import (
	"sync"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("state/lastSnapshotMarker")

const (
	lastSnapshot = "lastSnapshot"
)

type lastSnapshotMarker struct {
	mutex                       sync.RWMutex
	latestFinishedSnapshotEpoch uint32
}

func NewLastSnapshotMarker() *lastSnapshotMarker {
	return &lastSnapshotMarker{}
}

func (lsm *lastSnapshotMarker) AddMarker(trieStorageManager common.StorageManager, epoch uint32, rootHash []byte) {
	err := storage.WaitForStorageEpochChange(storage.StorageEpochChangeWaitArgs{
		TrieStorageManager:            trieStorageManager,
		Epoch:                         epoch,
		WaitTimeForSnapshotEpochCheck: storage.WaitTimeForSnapshotEpochCheck,
		SnapshotWaitTimeout:           storage.SnapshotWaitTimeout,
	})
	if err != nil {
		log.Warn("err while waiting for storage epoch change", "err", err, "epoch", epoch, "rootHash", rootHash)
		return
	}

	lsm.mutex.Lock()
	defer lsm.mutex.Unlock()

	if epoch <= lsm.latestFinishedSnapshotEpoch {
		log.Debug("will not put lastSnapshot marker in epoch storage",
			"epoch", epoch,
			"latestFinishedSnapshotEpoch", lsm.latestFinishedSnapshotEpoch,
		)
		return
	}

	err = trieStorageManager.PutInEpoch([]byte(lastSnapshot), rootHash, epoch)
	if err != nil {
		log.Warn("could not set lastSnapshot", err, "rootHash", rootHash, "epoch", epoch, "rootHash", rootHash)
	}
}

func (lsm *lastSnapshotMarker) RemoveMarker(trieStorageManager common.StorageManager, epoch uint32, rootHash []byte) {
	lsm.mutex.RLock()
	defer lsm.mutex.RUnlock()

	err := trieStorageManager.RemoveFromAllActiveEpochs([]byte(lastSnapshot))
	if err != nil {
		log.Warn("could not remove lastSnapshot", err, "rootHash", rootHash, "epoch", epoch)
	}

	lsm.latestFinishedSnapshotEpoch = epoch
}

func (lsm *lastSnapshotMarker) GetMarkerInfo(trieStorageManager common.StorageManager) ([]byte, error) {
	lsm.mutex.RLock()
	defer lsm.mutex.RUnlock()

	rootHash, err := trieStorageManager.GetFromCurrentEpoch([]byte(lastSnapshot))
	if err != nil {
		return nil, err
	}

	return rootHash, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (lsm *lastSnapshotMarker) IsInterfaceNil() bool {
	return lsm == nil
}
