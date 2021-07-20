package sync

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

type concurrentTriesMap struct {
	tries map[string]temporary.Trie
	mutex sync.RWMutex
}

func newConcurrentTriesMap() *concurrentTriesMap {
	return &concurrentTriesMap{
		tries: make(map[string]temporary.Trie),
		mutex: sync.RWMutex{},
	}
}

func (ct *concurrentTriesMap) getTrie(id string) (temporary.Trie, bool) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	if trie, ok := ct.tries[id]; ok {
		return trie, true
	}

	return nil, false
}

func (ct *concurrentTriesMap) setTrie(id string, trie temporary.Trie) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	ct.tries[id] = trie
}

func (ct *concurrentTriesMap) getTries() map[string]temporary.Trie {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	return ct.tries
}
