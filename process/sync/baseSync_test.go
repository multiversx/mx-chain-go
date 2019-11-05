package sync

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestBaseBootstrap_SyncBlocksShouldNotCallSyncIfNotConnectedToTheNetwork(t *testing.T) {
	t.Parallel()

	var numCalls uint32
	boot := &baseBootstrap{
		chStopSync: make(chan bool),
		syncStarter: &mock.SyncStarterStub{
			SyncBlockCalled: func() error {
				atomic.AddUint32(&numCalls, 1)
				return nil
			},
		},
		networkWatcher: &mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return false
			},
		},
	}

	go boot.syncBlocks()
	//make sure go routine started and waited a few cycles of boot.syncBlocks
	time.Sleep(time.Second + sleepTime*10)

	assert.Equal(t, uint32(0), atomic.LoadUint32(&numCalls))
}

func TestBaseBootstrap_SyncBlocksShouldCallSyncIfConnectedToTheNetwork(t *testing.T) {
	t.Parallel()

	var numCalls uint32
	boot := &baseBootstrap{
		chStopSync: make(chan bool),
		syncStarter: &mock.SyncStarterStub{
			SyncBlockCalled: func() error {
				atomic.AddUint32(&numCalls, 1)
				return nil
			},
		},
		networkWatcher: &mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
	}

	go boot.syncBlocks()

	//make sure go routine started and waited a few cycles of boot.syncBlocks
	time.Sleep(time.Second + sleepTime*10)

	assert.True(t, atomic.LoadUint32(&numCalls) > 0)
}
