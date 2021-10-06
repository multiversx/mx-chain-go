package evictionWaitingList

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
)

type hashInfo struct {
	roothashes [][]byte
}

// MemoryEvictionWaitingListArgs is the DTO used in the NewMemoryEvictionWaitingList function
type MemoryEvictionWaitingListArgs struct {
	RootHashesSize uint
	HashesSize     uint
}

// memoryEvictionWaitingList is a structure that caches keys that need to be removed from a certain database.
// If the cache is full, the caches will be emptied automatically. Writing at the same key in
// cacher and db will overwrite the previous values.
type memoryEvictionWaitingList struct {
	cache          map[string]common.ModifiedHashes
	reversedCache  map[string]*hashInfo
	rootHashesSize uint
	hashesSize     uint
	opMutex        sync.RWMutex
}

// NewMemoryEvictionWaitingList creates a new instance of memoryEvictionWaitingList
func NewMemoryEvictionWaitingList(args MemoryEvictionWaitingListArgs) (*memoryEvictionWaitingList, error) {
	if args.RootHashesSize < 1 {
		return nil, fmt.Errorf("%w for RootHashesSize in NewMemoryEvictionWaitingList", data.ErrInvalidCacheSize)
	}
	if args.HashesSize < 1 {
		return nil, fmt.Errorf("%w for HashesSize in NewMemoryEvictionWaitingList", data.ErrInvalidCacheSize)
	}

	return &memoryEvictionWaitingList{
		cache:          make(map[string]common.ModifiedHashes),
		reversedCache:  make(map[string]*hashInfo),
		rootHashesSize: args.RootHashesSize,
		hashesSize:     args.HashesSize,
	}, nil
}

// Put stores the given hashes in the eviction waiting list, in the position given by the root hash
func (mewl *memoryEvictionWaitingList) Put(rootHash []byte, hashes common.ModifiedHashes) error {
	mewl.opMutex.Lock()
	defer mewl.opMutex.Unlock()

	log.Trace("trie eviction waiting list", "size", len(mewl.cache))

	mewl.putInReversedCache(rootHash, hashes)
	mewl.cache[string(rootHash)] = hashes

	if !mewl.cachesFull() {
		return nil
	}

	log.Warn("trie nodes eviction waiting list full, emptying...")
	mewl.cache = make(map[string]common.ModifiedHashes)
	mewl.reversedCache = make(map[string]*hashInfo)

	return nil
}

func (mewl *memoryEvictionWaitingList) cachesFull() bool {
	if uint(len(mewl.cache)) > mewl.rootHashesSize {
		return true
	}
	if uint(len(mewl.reversedCache)) > mewl.hashesSize {
		return true
	}

	return false
}

func (mewl *memoryEvictionWaitingList) putInReversedCache(rootHash []byte, hashes common.ModifiedHashes) {
	for hash := range hashes {
		info, existing := mewl.reversedCache[hash]
		if !existing {
			info = &hashInfo{
				roothashes: [][]byte{rootHash},
			}
			mewl.reversedCache[hash] = info
			continue
		}

		if mewl.index(info, rootHash) != -1 {
			continue
		}

		info.roothashes = append(info.roothashes, rootHash)
	}
}

func (mewl *memoryEvictionWaitingList) index(info *hashInfo, roothash []byte) int {
	for index, hash := range info.roothashes {
		if bytes.Equal(hash, roothash) {
			return index
		}
	}

	return -1
}

func (mewl *memoryEvictionWaitingList) removeFromReversedCache(rootHash []byte, hashes common.ModifiedHashes) {
	for hash := range hashes {
		info, ok := mewl.reversedCache[hash]
		if !ok {
			continue
		}
		idx := mewl.index(info, rootHash)
		if idx < 0 {
			continue
		}

		if len(info.roothashes) == 1 {
			delete(mewl.reversedCache, hash)
			continue
		}

		info.roothashes = append(info.roothashes[:idx], info.roothashes[idx+1:]...)
	}
}

// Evict returns and removes from the waiting list all the hashes from the position given by the root hash
func (mewl *memoryEvictionWaitingList) Evict(rootHash []byte) (common.ModifiedHashes, error) {
	mewl.opMutex.Lock()
	defer mewl.opMutex.Unlock()

	hashes, ok := mewl.cache[string(rootHash)]
	if !ok {
		return make(common.ModifiedHashes), nil
	}

	delete(mewl.cache, string(rootHash))
	defer mewl.removeFromReversedCache(rootHash, hashes)

	return hashes, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mewl *memoryEvictionWaitingList) IsInterfaceNil() bool {
	return mewl == nil
}

// ShouldKeepHash searches for the given hash in all of the evictionWaitingList's newHashes.
// If the identifier is equal to oldRoot, then we should also search in oldHashes.
func (mewl *memoryEvictionWaitingList) ShouldKeepHash(hash string, identifier state.TriePruningIdentifier) (bool, error) {
	mewl.opMutex.RLock()
	defer mewl.opMutex.RUnlock()

	info, found := mewl.reversedCache[hash]
	if !found {
		return false, nil
	}

	for _, key := range info.roothashes {
		if len(key) == 0 {
			return false, state.ErrInvalidKey
		}

		lastByte := key[len(key)-1]
		if state.TriePruningIdentifier(lastByte) == state.OldRoot && identifier == state.OldRoot {
			continue
		}

		return true, nil
	}

	return false, nil
}

// Close returns nil
func (mewl *memoryEvictionWaitingList) Close() error {
	return nil
}
