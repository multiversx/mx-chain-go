package common

import "sync"

// ModifiedHashes is used to memorize all old hashes and new hashes from when a trie is committed
type ModifiedHashes map[string]struct{}

// Clone is used to create a clone of the map
func (mh ModifiedHashes) Clone() ModifiedHashes {
	newMap := make(ModifiedHashes)
	for key := range mh {
		newMap[key] = struct{}{}
	}

	return newMap
}

// ModifiedHashesSlice is used to memorize all old hashes and new hashes from when a trie is committed
type modifiedHashesSlice struct {
	hashes [][]byte
	sync.RWMutex
}

// NewModifiedHashesSlice is used to create a new instance of modifiedHashesSlice
func NewModifiedHashesSlice(initialCapacity int) *modifiedHashesSlice {
	return &modifiedHashesSlice{
		hashes: make([][]byte, 0, initialCapacity),
	}
}

// Append is used to add new hashes to the slice
func (mhs *modifiedHashesSlice) Append(hashes [][]byte) {
	mhs.Lock()
	defer mhs.Unlock()

	mhs.hashes = append(mhs.hashes, hashes...)
}

// Get is used to get the hashes from the slice
func (mhs *modifiedHashesSlice) Get() [][]byte {
	mhs.RLock()
	defer mhs.RUnlock()

	hashes := make([][]byte, len(mhs.hashes))
	copy(hashes, mhs.hashes)
	return hashes
}
