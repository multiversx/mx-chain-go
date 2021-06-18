package hashesHolder

import (
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

type checkpointHashesHolder struct {
	hashes      []map[string]data.ModifiedHashes
	currentSize uint64
	maxSize     uint64
	hashSize    uint64
	mutex       sync.RWMutex
}

var log = logger.GetOrCreate("trie/hashesHolder")

// NewCheckpointHashesHolder creates a new instance of hashesHolder
func NewCheckpointHashesHolder(maxSize uint64, hashSize uint64) *checkpointHashesHolder {
	log.Debug("created a new instance of checkpoint hashes holder",
		"max size", core.ConvertBytes(maxSize),
		"hash size", hashSize,
	)

	return &checkpointHashesHolder{
		hashes:      make([]map[string]data.ModifiedHashes, 0),
		currentSize: 0,
		maxSize:     maxSize,
		hashSize:    hashSize,
		mutex:       sync.RWMutex{},
	}
}

// Put appends the given hashes to the underlying array of maps. Put returns true if the maxSize is reached,
// meaning that a commit operation needs to be done in order to clear the array of maps.
func (c *checkpointHashesHolder) Put(rootHash []byte, hashes data.ModifiedHashes) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	newHashes := make(map[string]data.ModifiedHashes)
	newHashes[string(rootHash)] = hashes
	c.hashes = append(c.hashes, newHashes)

	mapSize := getMapSize(newHashes, c.hashSize)
	c.currentSize = c.currentSize + mapSize

	log.Debug("checkpoint hashes holder size after put",
		"current size", core.ConvertBytes(c.currentSize),
		"len", len(c.hashes),
	)

	return c.currentSize >= c.maxSize
}

// ShouldCommit returns true if the given hash is found.
// That means that the hash was modified since the last checkpoint,
// and needs to be committed into the snapshot DB.
func (c *checkpointHashesHolder) ShouldCommit(hash []byte) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, hashesMap := range c.hashes {
		if isInMap(hash, hashesMap) {
			return true
		}
	}

	return false
}

// RemoveCommitted removes entries from the array until it reaches the lastCommittedRootHash.
func (c *checkpointHashesHolder) RemoveCommitted(lastCommittedRootHash []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	sizeOfRemovedHashes := uint64(0)
	for index, hashesMap := range c.hashes {
		sizeOfRemovedHashes = sizeOfRemovedHashes + getMapSize(hashesMap, c.hashSize)

		_, ok := hashesMap[string(lastCommittedRootHash)]
		if !ok {
			continue
		}

		c.hashes = c.hashes[index+1:]
		ok = checkCorrectSize(c.currentSize, sizeOfRemovedHashes)
		if !ok {
			c.computeCurrentSize()
			return
		}

		c.currentSize = c.currentSize - sizeOfRemovedHashes
		log.Debug("checkpoint hashes holder size after remove",
			"current size", core.ConvertBytes(c.currentSize),
			"len", len(c.hashes),
		)
		return
	}
}

func (c *checkpointHashesHolder) computeCurrentSize() {
	totalSize := uint64(0)
	for _, hashesMap := range c.hashes {
		totalSize += getMapSize(hashesMap, c.hashSize)
	}

	c.currentSize = totalSize
}

// Remove removes the given hash from all the entries
func (c *checkpointHashesHolder) Remove(hash []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, hashesMap := range c.hashes {
		c.removeHashFromMap(hash, hashesMap)
	}
}

func (c *checkpointHashesHolder) removeHashFromMap(hash []byte, hashesMap map[string]data.ModifiedHashes) {
	for _, hashes := range hashesMap {
		_, ok := hashes[string(hash)]
		if !ok {
			continue
		}

		delete(hashes, string(hash))

		ok = checkCorrectSize(c.currentSize, c.hashSize)
		if !ok {
			c.computeCurrentSize()
			continue
		}

		c.currentSize = c.currentSize - c.hashSize
	}
}

func isInMap(hash []byte, hashesMap map[string]data.ModifiedHashes) bool {
	for _, hashes := range hashesMap {
		_, ok := hashes[string(hash)]
		if ok {
			return true
		}
	}

	return false
}

func getMapSize(hashesMap map[string]data.ModifiedHashes, hashSize uint64) uint64 {
	mapSize := uint64(0)
	for key, values := range hashesMap {
		keySize := uint64(len(key))
		hashesSize := uint64(len(values)) * hashSize
		mapSize = mapSize + keySize + hashesSize
	}

	return mapSize
}

func checkCorrectSize(currentSize uint64, sizeToRemove uint64) bool {
	if sizeToRemove > currentSize {
		log.Error("hashesHolder sizeOfRemovedHashes is greater than hashesSize",
			"size of removed hashes", sizeToRemove,
			"hashes size", currentSize,
		)
		return false
	}

	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *checkpointHashesHolder) IsInterfaceNil() bool {
	return c == nil
}
