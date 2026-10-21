package trie

import (
	"errors"
	"fmt"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
)

type recoveryNodeStore interface {
	GetForRecovery([]byte, uint32) ([]byte, error)
	PutForRecovery([]byte, []byte, uint32) error
}

type recoveryTrieStorageManager struct {
	common.StorageManager
	nodes recoveryNodeStore
	epoch uint32
}

func newStorageForTrieSync(arg ArgTrieSyncer) (common.TrieStorageInteractor, error) {
	if !arg.RecoveryEpoch.HasValue {
		return NewSyncTrieStorageManager(arg.DB)
	}
	nodes, ok := arg.DB.(recoveryNodeStore)
	if !ok {
		nodes, ok = arg.DB.GetBaseTrieStorageManager().(recoveryNodeStore)
	}
	if !ok {
		return nil, fmt.Errorf("trie storage does not support recovery reads")
	}
	return &recoveryTrieStorageManager{StorageManager: arg.DB, nodes: nodes, epoch: arg.RecoveryEpoch.Value}, nil
}

func (rs *recoveryTrieStorageManager) Get(key []byte) ([]byte, error) {
	return rs.nodes.GetForRecovery(key, rs.epoch)
}

func (rs *recoveryTrieStorageManager) Put(key, value []byte) error {
	return rs.nodes.PutForRecovery(key, value, rs.epoch)
}

func commitSyncedNode(n node, db common.TrieStorageInteractor) (int, error) {
	_, recovery := db.(*recoveryTrieStorageManager)
	// Decoded local nodes are clean; intercepted nodes are marked dirty.
	if recovery && !n.isDirty() {
		encoded, err := collapseAndEncodeNode(n)
		return len(encoded), err
	}
	return encodeNodeAndCommitToDB(n, db)
}

// GetForRecovery reads only storage retained for the checkpoint.
func (tsm *trieStorageManager) GetForRecovery(key []byte, epoch uint32) ([]byte, error) {
	storer, ok := tsm.mainStorer.(recoveryNodeStore)
	if !ok {
		return nil, fmt.Errorf("main storer does not support recovery reads")
	}
	return storer.GetForRecovery(key, epoch)
}

// PutForRecovery persists a repaired node in the checkpoint epoch.
func (tsm *trieStorageManager) PutForRecovery(key, value []byte, epoch uint32) error {
	storer, ok := tsm.mainStorer.(recoveryNodeStore)
	if !ok {
		return fmt.Errorf("main storer does not support recovery writes")
	}
	return storer.PutForRecovery(key, value, epoch)
}

// GetForRecovery uses the static storer when snapshots are disabled.
func (tsm *trieStorageManagerWithoutSnapshot) GetForRecovery(key []byte, _ uint32) ([]byte, error) {
	value, err := tsm.Get(key)
	if errors.Is(err, ErrKeyNotFound) {
		return nil, storage.ErrKeyNotFound
	}
	return value, err
}

// PutForRecovery uses the static storer when snapshots are disabled.
func (tsm *trieStorageManagerWithoutSnapshot) PutForRecovery(key, value []byte, _ uint32) error {
	return tsm.Put(key, value)
}
