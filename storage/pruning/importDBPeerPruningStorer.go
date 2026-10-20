package pruning

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
)

// importDBPeerPruningStorer keeps at most two memory DBs, without an epoch cache.
// Epochs label operations; only successful snapshot completion retires a generation.
type importDBPeerPruningStorer struct {
	mut          sync.RWMutex
	current, old *database.MemDB
	epoch        uint32
	closed       bool
}

var _ storage.Storer = (*importDBPeerPruningStorer)(nil)

// NewImportDBPeerPruningStorer creates cache-free memory storage for metachain replay.
func NewImportDBPeerPruningStorer(epoch uint32) *importDBPeerPruningStorer {
	return &importDBPeerPruningStorer{current: database.NewMemDB(), epoch: epoch}
}

// Rotate redirects every write immediately. Retargeting a pending rotation reuses
// its replacement, which may already contain newly committed state.
func (s *importDBPeerPruningStorer) Rotate() {
	s.mut.Lock()
	defer s.mut.Unlock()
	if !s.closed && s.old == nil {
		s.old, s.current = s.current, database.NewMemDB()
	}
}

// SnapshotFinished releases the source after a settled-root snapshot completes.
func (s *importDBPeerPruningStorer) SnapshotFinished() {
	s.mut.Lock()
	defer s.mut.Unlock()
	if s.old != nil {
		_ = s.old.Destroy()
		s.old = nil
	}
}

func (s *importDBPeerPruningStorer) Put(key, value []byte) error {
	s.mut.RLock()
	defer s.mut.RUnlock()
	if s.closed {
		return storage.ErrDBIsClosed
	}
	return s.current.Put(key, value)
}

func (s *importDBPeerPruningStorer) Get(key []byte) ([]byte, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	if s.closed {
		return nil, storage.ErrDBIsClosed
	}
	value, err := s.current.Get(key)
	if err == nil || s.old == nil {
		return value, err
	}
	return s.old.Get(key)
}

func (s *importDBPeerPruningStorer) Has(key []byte) error                   { _, err := s.Get(key); return err }
func (s *importDBPeerPruningStorer) SearchFirst(key []byte) ([]byte, error) { return s.Get(key) }
func (s *importDBPeerPruningStorer) PutInEpoch(key, value []byte, _ uint32) error {
	return s.Put(key, value)
}
func (s *importDBPeerPruningStorer) PutInEpochWithoutCache(key, value []byte, _ uint32) error {
	return s.Put(key, value)
}
func (s *importDBPeerPruningStorer) GetFromEpoch(key []byte, _ uint32) ([]byte, error) {
	return s.Get(key)
}

func (s *importDBPeerPruningStorer) GetForRecovery(key []byte, _ uint32) ([]byte, error) {
	return s.Get(key)
}
func (s *importDBPeerPruningStorer) PutForRecovery(key, value []byte, _ uint32) error {
	return s.Put(key, value)
}

// Content-addressed nodes are shared across epochs. No physical epoch is
// reported, avoiding the snapshot adapter's previous-epoch backfill.
func (s *importDBPeerPruningStorer) GetWithoutAddingToCache(key []byte, _ uint32) ([]byte, core.OptionalUint32, error) {
	value, err := s.Get(key)
	return value, core.OptionalUint32{}, err
}

// Control markers must describe the replacement itself, without fallback.
func (s *importDBPeerPruningStorer) GetFromCurrentEpoch(key []byte) ([]byte, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	if s.closed {
		return nil, storage.ErrDBIsClosed
	}
	return s.current.Get(key)
}

func (s *importDBPeerPruningStorer) GetFromLastEpoch(key []byte) ([]byte, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	if s.closed {
		return nil, storage.ErrDBIsClosed
	}
	if s.old == nil {
		return nil, fmt.Errorf("no previous peer memory DB")
	}
	return s.old.Get(key)
}

func (s *importDBPeerPruningStorer) GetLatestStorageEpoch() (uint32, error) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	if s.closed {
		return 0, storage.ErrDBIsClosed
	}
	return s.epoch, nil
}
func (s *importDBPeerPruningStorer) GetOldestEpoch() (uint32, error) {
	return s.GetLatestStorageEpoch()
}
func (s *importDBPeerPruningStorer) SetEpochForPutOperation(epoch uint32) {
	s.mut.Lock()
	s.epoch = epoch
	s.mut.Unlock()
}

func (s *importDBPeerPruningStorer) RemoveFromCurrentEpoch(key []byte) error {
	s.mut.RLock()
	defer s.mut.RUnlock()
	if s.closed {
		return storage.ErrDBIsClosed
	}
	return s.current.Remove(key)
}

func (s *importDBPeerPruningStorer) Remove(key []byte) error {
	s.mut.RLock()
	defer s.mut.RUnlock()
	if s.closed {
		return storage.ErrDBIsClosed
	}
	if err := s.current.Remove(key); err != nil {
		return err
	}
	if s.old != nil {
		return s.old.Remove(key)
	}
	return nil
}
func (s *importDBPeerPruningStorer) RemoveFromAllActiveEpochs(key []byte) error { return s.Remove(key) }
func (s *importDBPeerPruningStorer) ClearCache()                                {}

func (s *importDBPeerPruningStorer) GetBulkFromEpoch(keys [][]byte, epoch uint32) ([]data.KeyValuePair, error) {
	values := make([]data.KeyValuePair, 0, len(keys))
	for _, key := range keys {
		value, err := s.GetFromEpoch(key, epoch)
		if err != nil {
			return nil, err
		}
		values = append(values, data.KeyValuePair{Key: key, Value: value})
	}
	return values, nil
}

func (s *importDBPeerPruningStorer) RangeKeys(handler func([]byte, []byte) bool) {
	if handler == nil {
		return
	}
	s.mut.RLock()
	defer s.mut.RUnlock()
	if s.closed {
		return
	}
	seen := make(map[string]struct{})
	more := true
	s.current.RangeKeys(func(key, value []byte) bool {
		seen[string(key)] = struct{}{}
		more = handler(key, value)
		return more
	})
	if more && s.old != nil {
		s.old.RangeKeys(func(key, value []byte) bool {
			if _, exists := seen[string(key)]; exists {
				return true
			}
			return handler(key, value)
		})
	}
}

func (s *importDBPeerPruningStorer) Close() error {
	s.mut.Lock()
	defer s.mut.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	_ = s.current.Destroy()
	if s.old != nil {
		_ = s.old.Destroy()
		s.old = nil
	}
	return nil
}
func (s *importDBPeerPruningStorer) DestroyUnit() error   { return s.Close() }
func (s *importDBPeerPruningStorer) IsInterfaceNil() bool { return s == nil }
