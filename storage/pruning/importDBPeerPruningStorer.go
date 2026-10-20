package pruning

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
)

// importDBPeerPruningStorer keeps at most two memory DBs, without an epoch cache.
// Epochs label operations; only successful snapshot completion retires a generation.
type importDBPeerPruningStorer struct {
	mut          sync.RWMutex
	current, old *database.MemDB
	bootstrap    storage.Storer
	epoch        uint32
	closed       bool
}

var _ storage.Storer = (*importDBPeerPruningStorer)(nil)

// NewImportDBPeerPruningStorer creates memory storage with an optional disk source
// used only until the restored starting root has been copied by a snapshot.
func NewImportDBPeerPruningStorer(epoch uint32, bootstrap storage.Storer) *importDBPeerPruningStorer {
	if check.IfNil(bootstrap) {
		bootstrap = nil
	}
	log.Debug("import-db peer memory storer created", "epoch", epoch, "has disk source", bootstrap != nil)
	return &importDBPeerPruningStorer{current: database.NewMemDB(), epoch: epoch, bootstrap: bootstrap}
}

// HasBootstrapSource reports whether the initial disk snapshot is still needed.
func (s *importDBPeerPruningStorer) HasBootstrapSource() bool {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.bootstrap != nil
}

// Rotate redirects every write immediately. Retargeting a pending rotation reuses
// its replacement, which may already contain newly committed state.
func (s *importDBPeerPruningStorer) Rotate() {
	s.mut.Lock()
	defer s.mut.Unlock()
	if !s.closed && s.old == nil && s.bootstrap == nil {
		s.old, s.current = s.current, database.NewMemDB()
		log.Debug("import-db peer memory DB rotated, old DB retained as read fallback", "epoch", s.epoch)
	} else {
		log.Debug("import-db peer memory DB rotation skipped", "epoch", s.epoch, "closed", s.closed,
			"has old memory DB", s.old != nil, "has disk source", s.bootstrap != nil)
	}
}

// SnapshotFinished closes the startup disk source or releases the previous
// memory generation after its snapshot completes.
func (s *importDBPeerPruningStorer) SnapshotFinished() {
	s.mut.Lock()
	defer s.mut.Unlock()
	if s.old != nil {
		err := s.old.Destroy()
		s.old = nil
		log.Debug("import-db peer old memory DB retired", "epoch", s.epoch, "error", err)
	}
	if s.bootstrap != nil {
		// The normal DB is a source, never a retired memory generation.
		log.Debug("import-db peer closing startup disk source", "epoch", s.epoch)
		err := s.bootstrap.Close()
		s.bootstrap = nil
		log.Debug("import-db peer startup disk source detached", "epoch", s.epoch, "error", err)
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
	if err == nil {
		return value, err
	}
	if s.old != nil {
		return s.old.Get(key)
	}
	if s.bootstrap != nil {
		return s.bootstrap.SearchFirst(key)
	}
	return nil, err
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
		if s.bootstrap != nil {
			return s.bootstrap.SearchFirst(key)
		}
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
	log.Debug("import-db peer memory storer closing", "epoch", s.epoch,
		"has old memory DB", s.old != nil, "has disk source", s.bootstrap != nil)
	s.closed = true
	_ = s.current.Destroy()
	if s.old != nil {
		_ = s.old.Destroy()
		s.old = nil
	}
	if s.bootstrap != nil {
		err := s.bootstrap.Close()
		s.bootstrap = nil
		log.Debug("import-db peer memory storer closed with disk source", "epoch", s.epoch, "error", err)
		return err
	}
	log.Debug("import-db peer memory storer closed", "epoch", s.epoch)
	return nil
}
func (s *importDBPeerPruningStorer) DestroyUnit() error   { return s.Close() }
func (s *importDBPeerPruningStorer) IsInterfaceNil() bool { return s == nil }
