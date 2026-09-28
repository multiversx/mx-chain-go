package pruning

import (
	"errors"
	"fmt"

	"github.com/multiversx/mx-chain-go/storage"
)

// GetForRecovery reads retained checkpoint data without consulting the shared cache.
func (ps *PruningStorer) GetForRecovery(key []byte, epoch uint32) ([]byte, error) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	if ps.pruningEnabled {
		pd, ok := ps.persistersMapByEpoch[epoch]
		if !ok || pd.getIsClosed() {
			return nil, fmt.Errorf("recovery persister for epoch %d is not open", epoch)
		}
	}
	for _, pd := range ps.activePersisters {
		if ps.pruningEnabled && pd.epoch > epoch {
			continue
		}
		value, err := pd.getPersister().Get(key)
		if err == nil {
			return value, nil
		}
		if !errors.Is(err, storage.ErrKeyNotFound) {
			return nil, err
		}
	}
	return nil, storage.ErrKeyNotFound
}

// PutForRecovery persists a repaired node once.
func (ps *PruningStorer) PutForRecovery(key, value []byte, epoch uint32) error {
	if !ps.pruningEnabled {
		return ps.Put(key, value)
	}
	return ps.PutInEpoch(key, value, epoch)
}
