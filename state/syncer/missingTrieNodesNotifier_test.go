package syncer

import (
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewMissingTrieNodesNotifier(t *testing.T) {
	t.Parallel()

	assert.False(t, check.IfNil(NewMissingTrieNodesNotifier()))
}

func TestMissingTrieNodesNotifier_RegisterHandler(t *testing.T) {
	t.Parallel()

	notifier := NewMissingTrieNodesNotifier()

	notifier.RegisterHandler(nil)
	notifier.RegisterHandler(&testscommon.StateSyncNotifierSubscriberStub{})
	notifier.RegisterHandler(nil)

	assert.Equal(t, 1, notifier.GetNumHandlers())
}

func TestMissingTrieNodesNotifier_NotifyMissingTrieNode(t *testing.T) {
	t.Parallel()

	numMissingDataTrieNodeFoundCalled := 0
	notifier := NewMissingTrieNodesNotifier()
	notifier.NotifyMissingTrieNode([]byte("hash1"))

	wg := sync.WaitGroup{}
	wg.Add(2)
	mutex := sync.Mutex{}

	notifier.RegisterHandler(&testscommon.StateSyncNotifierSubscriberStub{
		MissingDataTrieNodeFoundCalled: func(_ []byte) {
			mutex.Lock()
			numMissingDataTrieNodeFoundCalled++
			wg.Done()
			mutex.Unlock()
		},
	})

	notifier.NotifyMissingTrieNode(nil)
	notifier.NotifyMissingTrieNode([]byte("hash2"))
	notifier.NotifyMissingTrieNode([]byte("hash3"))

	wg.Wait()

	assert.Equal(t, 1, notifier.GetNumHandlers())
	assert.Equal(t, 2, numMissingDataTrieNodeFoundCalled)
}
