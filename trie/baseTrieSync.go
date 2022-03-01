package trie

import (
	"sync"
	"time"
)

type baseSyncTrie struct {
	mutStatistics sync.RWMutex
	numTrieNodes  uint64
	numLeaves     uint64
	numBytes      uint64
	duration      time.Duration
}

func (bst *baseSyncTrie) updateStats(bytesToAdd uint64, element node) {
	_, isLeaf := element.(*leafNode)

	bst.mutStatistics.Lock()
	bst.numBytes += bytesToAdd
	bst.numTrieNodes++
	if isLeaf {
		bst.numLeaves++
	}
	bst.mutStatistics.Unlock()
}

func (bst *baseSyncTrie) setSyncDuration(duration time.Duration) {
	bst.mutStatistics.Lock()
	bst.duration = duration
	bst.mutStatistics.Unlock()
}

// NumLeaves return the total number of leaves for the provided trie
func (bst *baseSyncTrie) NumLeaves() uint64 {
	bst.mutStatistics.RLock()
	defer bst.mutStatistics.RUnlock()

	return bst.numLeaves
}

// NumBytes returns the total number of bytes for the provided trie
func (bst *baseSyncTrie) NumBytes() uint64 {
	bst.mutStatistics.RLock()
	defer bst.mutStatistics.RUnlock()

	return bst.numBytes
}

// NumTrieNodes returns the total number of trie for the provided trie
func (bst *baseSyncTrie) NumTrieNodes() uint64 {
	bst.mutStatistics.RLock()
	defer bst.mutStatistics.RUnlock()

	return bst.numTrieNodes
}

// Duration returns the total sync duration for the provided trie
func (bst *baseSyncTrie) Duration() time.Duration {
	bst.mutStatistics.RLock()
	defer bst.mutStatistics.RUnlock()

	return bst.duration
}
