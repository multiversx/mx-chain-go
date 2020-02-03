package sync

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
)

type concurrentTriesMap struct {
	tries map[string]data.Trie
	mutex sync.RWMutex
}

func newConcurrentTriesMap() *concurrentTriesMap {
	return &concurrentTriesMap{
		tries: make(map[string]data.Trie),
		mutex: sync.RWMutex{},
	}
}

func (ct *concurrentTriesMap) getTrie(id string) (data.Trie, bool) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	if trie, ok := ct.tries[id]; ok {
		return trie, true
	}

	return nil, false
}

func (ct *concurrentTriesMap) setTrie(id string, trie data.Trie) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	ct.tries[id] = trie
}

func (ct *concurrentTriesMap) getTries() map[string]data.Trie {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	return ct.tries
}
