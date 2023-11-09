package hashesHolder

import (
	"bytes"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type checkpointHashesHolder struct {
	hashes      []common.ModifiedHashes
	rootHashes  [][]byte
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
		hashes:      make([]common.ModifiedHashes, 0),
		rootHashes:  make([][]byte, 0),
		currentSize: 0,
		maxSize:     maxSize,
		hashSize:    hashSize,
		mutex:       sync.RWMutex{},
	}
}

// Put appends the given hashes to the underlying array of maps. Put returns true if the maxSize is reached,
// meaning that a commit operation needs to be done in order to clear the array of maps.
func (c *checkpointHashesHolder) Put(rootHash []byte, hashes common.ModifiedHashes) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.rootHashes) != 0 {
		lastRootHash := c.rootHashes[len(c.rootHashes)-1]
		if bytes.Equal(lastRootHash, rootHash) {
			log.Debug("checkpoint hashes holder rootHash did not change")
			return false
		}
	}

	c.rootHashes = append(c.rootHashes, rootHash)
	c.hashes = append(c.hashes, hashes)

	mapSize := getMapSize(hashes, c.hashSize)
	c.currentSize = c.currentSize + mapSize + uint64(len(rootHash))

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
		_, found := hashesMap[string(hash)]
		if found {
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
	for index, rootHash := range c.rootHashes {
		mapHashes := c.hashes[index]
		sizeOfRemovedHashes = sizeOfRemovedHashes + getMapSize(mapHashes, c.hashSize) + uint64(len(rootHash))

		lastCommittedRootHashNotFound := !bytes.Equal(rootHash, lastCommittedRootHash)
		if lastCommittedRootHashNotFound {
			continue
		}

		c.hashes = c.hashes[index+1:]
		c.rootHashes = c.rootHashes[index+1:]

		ok := checkCorrectSize(c.currentSize, sizeOfRemovedHashes)
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
	for index, hashesMap := range c.hashes {
		totalSize += getMapSize(hashesMap, c.hashSize) + uint64(len(c.rootHashes[index]))
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

func (c *checkpointHashesHolder) removeHashFromMap(hash []byte, hashesMap common.ModifiedHashes) {
	_, ok := hashesMap[string(hash)]
	if !ok {
		return
	}

	delete(hashesMap, string(hash))

	ok = checkCorrectSize(c.currentSize, c.hashSize)
	if !ok {
		c.computeCurrentSize()
		return
	}

	c.currentSize -= c.hashSize
}

func getMapSize(hashesMap common.ModifiedHashes, hashSize uint64) uint64 {
	return uint64(len(hashesMap)) * hashSize
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
