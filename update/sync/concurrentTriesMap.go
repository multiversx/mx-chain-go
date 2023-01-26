package sync

import (
	"sync"

	"github.com/multiversx/mx-chain-go/common"
)

type concurrentTriesMap struct {
	tries map[string]common.Trie
	mutex sync.RWMutex
}

func newConcurrentTriesMap() *concurrentTriesMap {
	return &concurrentTriesMap{
		tries: make(map[string]common.Trie),
		mutex: sync.RWMutex{},
	}
}

func (ct *concurrentTriesMap) getTrie(id string) (common.Trie, bool) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	if trie, ok := ct.tries[id]; ok {
		return trie, true
	}

	return nil, false
}

func (ct *concurrentTriesMap) setTrie(id string, trie common.Trie) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	ct.tries[id] = trie
}

func (ct *concurrentTriesMap) getTries() map[string]common.Trie {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	return ct.tries
}
