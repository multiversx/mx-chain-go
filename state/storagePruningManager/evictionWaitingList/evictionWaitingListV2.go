package evictionWaitingList

import (
	"bytes"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type hashInfo struct {
	roothashes [][]byte
}

// evictionWaitingList is a structure that caches keys that need to be removed from a certain database.
// If the cache is full, the keys will be stored in the underlying database. Writing at the same key in
// cacher and db will overwrite the previous values.
type evictionWaitingListV2 struct {
	cache         map[string]temporary.ModifiedHashes
	inversedCache map[string]hashInfo
	cacheSize     uint
	db            storage.Persister
	marshalizer   marshal.Marshalizer
	opMutex       sync.RWMutex
}

// NewEvictionWaitingListV2 creates a new instance of evictionWaitingList
func NewEvictionWaitingListV2(size uint, db storage.Persister, marshalizer marshal.Marshalizer) (*evictionWaitingListV2, error) {
	if size < 1 {
		return nil, data.ErrInvalidCacheSize
	}
	if check.IfNil(db) {
		return nil, data.ErrNilDatabase
	}
	if check.IfNil(marshalizer) {
		return nil, data.ErrNilMarshalizer
	}

	return &evictionWaitingListV2{
		cache:         make(map[string]temporary.ModifiedHashes),
		inversedCache: make(map[string]hashInfo),
		cacheSize:     size,
		db:            db,
		marshalizer:   marshalizer,
	}, nil
}

// Put stores the given hashes in the eviction waiting list, in the position given by the root hash
func (ewl *evictionWaitingListV2) Put(rootHash []byte, hashes temporary.ModifiedHashes) error {
	ewl.opMutex.Lock()
	defer ewl.opMutex.Unlock()

	log.Trace("trie eviction waiting list", "size", len(ewl.cache))

	ewl.putInInversedCache(rootHash, hashes)

	if uint(len(ewl.cache)) < ewl.cacheSize {
		ewl.cache[string(rootHash)] = hashes
		return nil
	}

	b := &batch.Batch{}

	for h := range hashes {
		b.Data = append(b.Data, []byte(h))
	}

	marshalizedHashes, err := ewl.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	err = ewl.db.Put(rootHash, marshalizedHashes)
	if err != nil {
		return err
	}
	ewl.cache[string(rootHash)] = nil

	return nil
}

func (ewl *evictionWaitingListV2) putInInversedCache(rootHash []byte, hashes temporary.ModifiedHashes) {
	for hash := range hashes {
		info, existing := ewl.inversedCache[hash]
		if ewl.index(info, rootHash) != -1 {
			continue
		}

		info.roothashes = append(info.roothashes, rootHash)
		if !existing {
			ewl.inversedCache[hash] = info
		}
	}
}

func (ewl *evictionWaitingListV2) index(info hashInfo, roothash []byte) int {
	for index, hash := range info.roothashes {
		if bytes.Equal(hash, roothash) {
			return index
		}
	}

	return -1
}

func (ewl *evictionWaitingListV2) removeFromInversedCache(rootHash []byte, hashes temporary.ModifiedHashes) {
	for hash := range hashes {
		info, ok := ewl.inversedCache[hash]
		if !ok {
			continue
		}
		idx := ewl.index(info, rootHash)
		if idx < 0 {
			continue
		}

		if len(info.roothashes) == 1 {
			delete(ewl.inversedCache, hash)
			continue
		}

		info.roothashes[idx] = info.roothashes[len(info.roothashes)-1]
		info.roothashes[len(info.roothashes)-1] = nil
		info.roothashes = info.roothashes[:len(info.roothashes)-1]
	}
}

// Evict returns and removes from the waiting list all the hashes from the position given by the root hash
func (ewl *evictionWaitingListV2) Evict(rootHash []byte) (temporary.ModifiedHashes, error) {
	ewl.opMutex.Lock()
	defer ewl.opMutex.Unlock()

	hashes, ok := ewl.cache[string(rootHash)]

	if !ok {
		return nil, nil
	}

	delete(ewl.cache, string(rootHash))

	defer ewl.removeFromInversedCache(rootHash, hashes)

	if len(hashes) != 0 {
		return hashes, nil
	}

	marshalizedHashes, err := ewl.db.Get(rootHash)
	if err != nil {
		return nil, err
	}

	b := &batch.Batch{}

	err = ewl.marshalizer.Unmarshal(b, marshalizedHashes)
	if err != nil {
		return nil, err
	}

	hashes = make(temporary.ModifiedHashes, len(b.Data))
	for _, h := range b.Data {
		hashes[string(h)] = struct{}{}
	}

	err = ewl.db.Remove(rootHash)
	if err != nil {
		return nil, err
	}

	return hashes, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ewl *evictionWaitingListV2) IsInterfaceNil() bool {
	return ewl == nil
}

// ShouldKeepHash searches for the given hash in all of the evictionWaitingList's newHashes.
// If the identifier is equal to oldRoot, then we should also search in oldHashes.
func (ewl *evictionWaitingListV2) ShouldKeepHash(hash string, identifier temporary.TriePruningIdentifier) (bool, error) {
	ewl.opMutex.RLock()
	defer ewl.opMutex.RUnlock()

	info, found := ewl.inversedCache[hash]
	if !found {
		return false, nil
	}

	for _, key := range info.roothashes {
		if len(key) == 0 {
			return false, state.ErrInvalidKey
		}

		lastByte := key[len(key)-1]
		if temporary.TriePruningIdentifier(lastByte) == temporary.OldRoot && identifier == temporary.OldRoot {
			continue
		}

		return true, nil
	}

	return false, nil
}

// Close - closes the underlying db
func (ewl *evictionWaitingListV2) Close() error {
	if !check.IfNil(ewl.db) {
		//TODO: @beni verify if a flush is needed before closing the DB
		return ewl.db.Close()
	}
	return nil
}
