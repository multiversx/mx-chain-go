package sync

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	ctx, cancelFunc := context.WithCancel(context.Background())
	go boot.syncBlocks(ctx)

	//make sure go routine started and waited a few cycles of boot.syncBlocks
	time.Sleep(time.Second + sleepTime*10)
	cancelFunc()

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
		roundHandler: &mock.RoundHandlerMock{
			BeforeGenesisCalled: func() bool {
				return false
			},
		},
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	go boot.syncBlocks(ctx)

	//make sure go routine started and waited a few cycles of boot.syncBlocks
	time.Sleep(time.Second + sleepTime*10)
	cancelFunc()

	assert.True(t, atomic.LoadUint32(&numCalls) > 0)
}

func TestBaseBootstrap_GetOrderedMiniBlocksShouldErrMissingBody(t *testing.T) {
	t.Parallel()

	hashes := [][]byte{[]byte("hash1")}
	orderedMiniBlocks, err := getOrderedMiniBlocks(hashes, nil)

	assert.Nil(t, orderedMiniBlocks)
	assert.Equal(t, process.ErrMissingBody, err)
}

func TestBaseBootstrap_GetOrderedMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	hash1 := []byte("hash1")
	miniBlock1 := &block.MiniBlock{SenderShardID: 0}

	hash2 := []byte("hash2")
	miniBlock2 := &block.MiniBlock{SenderShardID: 1}

	hash3 := []byte("hash3")
	miniBlock3 := &block.MiniBlock{SenderShardID: 2}

	hashes := [][]byte{hash1, hash2, hash3}
	miniBlocksAndHashes := []*block.MiniblockAndHash{
		{
			Hash:      hash1,
			Miniblock: miniBlock1,
		},
		{
			Hash:      hash2,
			Miniblock: miniBlock2,
		},
		{
			Hash:      hash3,
			Miniblock: miniBlock3,
		},
	}

	orderedMiniBlocks, err := getOrderedMiniBlocks(hashes, miniBlocksAndHashes)

	assert.Nil(t, err)
	require.Equal(t, 3, len(orderedMiniBlocks))
	assert.Equal(t, uint32(0), orderedMiniBlocks[0].SenderShardID)
	assert.Equal(t, uint32(1), orderedMiniBlocks[1].SenderShardID)
	assert.Equal(t, uint32(2), orderedMiniBlocks[2].SenderShardID)
}
