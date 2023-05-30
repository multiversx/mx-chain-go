package syncer

import (
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
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

	err := notifier.RegisterHandler(nil)
	assert.Equal(t, common.ErrNilStateSyncNotifierSubscriber, err)
	assert.Equal(t, 0, notifier.GetNumHandlers())

	err = notifier.RegisterHandler(&testscommon.StateSyncNotifierSubscriberStub{})
	assert.Nil(t, err)

	assert.Equal(t, 1, notifier.GetNumHandlers())
}

func TestMissingTrieNodesNotifier_AsyncNotifyMissingTrieNode(t *testing.T) {
	t.Parallel()

	numMissingDataTrieNodeFoundCalled := 0
	notifier := NewMissingTrieNodesNotifier()
	notifier.AsyncNotifyMissingTrieNode([]byte("hash1"))

	wg := sync.WaitGroup{}
	wg.Add(2)
	mutex := sync.Mutex{}

	err := notifier.RegisterHandler(&testscommon.StateSyncNotifierSubscriberStub{
		MissingDataTrieNodeFoundCalled: func(_ []byte) {
			mutex.Lock()
			numMissingDataTrieNodeFoundCalled++
			wg.Done()
			mutex.Unlock()
		},
	})
	assert.Nil(t, err)

	notifier.AsyncNotifyMissingTrieNode(nil)
	notifier.AsyncNotifyMissingTrieNode([]byte("hash2"))
	notifier.AsyncNotifyMissingTrieNode([]byte("hash3"))

	wg.Wait()

	assert.Equal(t, 1, notifier.GetNumHandlers())
	assert.Equal(t, 2, numMissingDataTrieNodeFoundCalled)
}
