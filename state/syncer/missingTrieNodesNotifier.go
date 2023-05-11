package syncer

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
)

type missingTrieNodesNotifier struct {
	handlers []common.StateSyncNotifierSubscriber
	mutex    sync.RWMutex
}

// NewMissingTrieNodesNotifier creates a new missing trie nodes notifier
func NewMissingTrieNodesNotifier() *missingTrieNodesNotifier {
	return &missingTrieNodesNotifier{
		handlers: make([]common.StateSyncNotifierSubscriber, 0),
		mutex:    sync.RWMutex{},
	}
}

// RegisterHandler registers a new handler for the missing trie nodes notifier
func (mtnn *missingTrieNodesNotifier) RegisterHandler(handler common.StateSyncNotifierSubscriber) error {
	if check.IfNil(handler) {
		return common.ErrNilStateSyncNotifierSubscriber
	}

	mtnn.mutex.Lock()
	mtnn.handlers = append(mtnn.handlers, handler)
	mtnn.mutex.Unlock()

	return nil
}

// AsyncNotifyMissingTrieNode asynchronously notifies all the registered handlers that a trie node is missing
func (mtnn *missingTrieNodesNotifier) AsyncNotifyMissingTrieNode(hash []byte) {
	if common.IsEmptyTrie(hash) {
		log.Warn("missingTrieNodesNotifier: empty trie hash")
		return
	}

	mtnn.mutex.RLock()
	defer mtnn.mutex.RUnlock()

	for _, handler := range mtnn.handlers {
		go handler.MissingDataTrieNodeFound(hash)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (mtnn *missingTrieNodesNotifier) IsInterfaceNil() bool {
	return mtnn == nil
}
