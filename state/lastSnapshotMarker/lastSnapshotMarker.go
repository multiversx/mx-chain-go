package lastSnapshotMarker

import (
	"sync"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage/storageEpochChange"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("state/lastSnapshotMarker")

const (
	// LastSnapshot is the marker for the last snapshot started
	LastSnapshot = "lastSnapshot"
)

type lastSnapshotMarker struct {
	mutex                       sync.RWMutex
	latestFinishedSnapshotEpoch uint32
}

// NewLastSnapshotMarker creates a new instance of lastSnapshotMarker
func NewLastSnapshotMarker() *lastSnapshotMarker {
	return &lastSnapshotMarker{}
}

// AddMarker adds a marker for the last snapshot started in the given epoch
func (lsm *lastSnapshotMarker) AddMarker(trieStorageManager common.StorageManager, epoch uint32, rootHash []byte) {
	err := storageEpochChange.WaitForStorageEpochChange(storageEpochChange.StorageEpochChangeWaitArgs{
		TrieStorageManager:            trieStorageManager,
		Epoch:                         epoch,
		WaitTimeForSnapshotEpochCheck: storageEpochChange.WaitTimeForSnapshotEpochCheck,
		SnapshotWaitTimeout:           storageEpochChange.SnapshotWaitTimeout,
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

	err = trieStorageManager.PutInEpoch([]byte(LastSnapshot), rootHash, epoch)
	if err != nil {
		log.Warn("could not set lastSnapshot", err, "rootHash", rootHash, "epoch", epoch, "rootHash", rootHash)
	}
}

// RemoveMarker removes the marker for the last snapshot started
func (lsm *lastSnapshotMarker) RemoveMarker(trieStorageManager common.StorageManager, epoch uint32, rootHash []byte) {
	lsm.mutex.Lock()
	defer lsm.mutex.Unlock()

	err := trieStorageManager.RemoveFromAllActiveEpochs([]byte(LastSnapshot))
	if err != nil {
		log.Warn("could not remove lastSnapshot", err, "rootHash", rootHash, "epoch", epoch)
	}

	lsm.latestFinishedSnapshotEpoch = epoch
}

// GetMarkerInfo returns the root hash of the last snapshot started
func (lsm *lastSnapshotMarker) GetMarkerInfo(trieStorageManager common.StorageManager) ([]byte, error) {
	return trieStorageManager.GetFromCurrentEpoch([]byte(LastSnapshot))
}

// IsInterfaceNil returns true if there is no value under the interface
func (lsm *lastSnapshotMarker) IsInterfaceNil() bool {
	return lsm == nil
}
