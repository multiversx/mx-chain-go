package sync

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getMockChainHandler() data.ChainHandler {
	return &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Epoch: 0,
			}
		},
	}
}

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

	// make sure go routine started and waited a few cycles of boot.syncBlocks
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

	// make sure go routine started and waited a few cycles of boot.syncBlocks
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

func TestBaseBootstrap_GetNodeState(t *testing.T) {
	t.Parallel()

	boot := &baseBootstrap{
		isInImportMode:        true,
		isNodeStateCalculated: true,
		roundHandler:          &mock.RoundHandlerMock{},
		chainHandler:          getMockChainHandler(),
		currentEpochProvider:  &testscommon.CurrentEpochProviderStub{},
	}
	assert.Equal(t, common.NsNotSynchronized, boot.GetNodeState())

	boot = &baseBootstrap{
		isInImportMode:        false,
		isNodeStateCalculated: true,
		roundHandler:          &mock.RoundHandlerMock{},
		chainHandler:          getMockChainHandler(),
		currentEpochProvider:  &testscommon.CurrentEpochProviderStub{},
	}
	assert.Equal(t, common.NsNotSynchronized, boot.GetNodeState())

	boot = &baseBootstrap{
		roundIndex:            1,
		isInImportMode:        false,
		isNodeStateCalculated: true,
		roundHandler:          &mock.RoundHandlerMock{},
		chainHandler:          getMockChainHandler(),
		currentEpochProvider:  &testscommon.CurrentEpochProviderStub{},
	}
	assert.Equal(t, common.NsNotCalculated, boot.GetNodeState())

	boot = &baseBootstrap{
		roundIndex:            1,
		isInImportMode:        false,
		isNodeStateCalculated: true,
		roundHandler:          &mock.RoundHandlerMock{},
		chainHandler:          getMockChainHandler(),
		currentEpochProvider: &testscommon.CurrentEpochProviderStub{
			EpochIsActiveInNetworkCalled: func(epoch uint32) bool {
				return false
			},
		},
	}
	assert.Equal(t, common.NsNotSynchronized, boot.GetNodeState())
}

func TestBaseSync_getEpochOfCurrentBlockGenesis(t *testing.T) {
	t.Parallel()

	genesisEpoch := uint32(1123)
	boot := &baseBootstrap{
		chainHandler: &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{
					Epoch: genesisEpoch,
				}
			},
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return nil
			},
		},
	}

	epoch := boot.getEpochOfCurrentBlock()
	assert.Equal(t, genesisEpoch, epoch)
}

func TestBaseSync_getEpochOfCurrentBlockHeader(t *testing.T) {
	t.Parallel()

	genesisEpoch := uint32(1123)
	headerEpoch := uint32(97493)
	boot := &baseBootstrap{
		chainHandler: &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{
					Epoch: genesisEpoch,
				}
			},
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.Header{
					Epoch: headerEpoch,
				}
			},
		},
	}

	epoch := boot.getEpochOfCurrentBlock()
	assert.Equal(t, headerEpoch, epoch)
}

func TestBaseSync_shouldAllowRollback(t *testing.T) {
	t.Parallel()

	finalBlockHash := []byte("final block hash")
	notFinalBlockHash := []byte("not final block hash")
	firstBlockNonce := &core.OptionalUint64{
		HasValue: true,
		Value:    2,
	}
	boot := &baseBootstrap{
		forkDetector: &mock.ForkDetectorMock{
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return 10
			},
			GetHighestFinalBlockHashCalled: func() []byte {
				return finalBlockHash
			},
		},
		blockProcessor: &testscommon.BlockProcessorStub{
			NonceOfFirstCommittedBlockCalled: func() core.OptionalUint64 {
				return *firstBlockNonce
			},
		},
	}

	t.Run("should allow rollback nonces above final", func(t *testing.T) {
		header := &testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 11
			},
			HasScheduledMiniBlocksCalled: func() bool {
				return false
			},
		}
		require.True(t, boot.shouldAllowRollback(header, finalBlockHash))
		require.True(t, boot.shouldAllowRollback(header, notFinalBlockHash))

		header.HasScheduledMiniBlocksCalled = func() bool {
			return true
		}
		require.True(t, boot.shouldAllowRollback(header, finalBlockHash))
		require.True(t, boot.shouldAllowRollback(header, notFinalBlockHash))
	})

	t.Run("should not allow rollback of a final header with the same final hash if it doesn't have scheduled miniBlocks", func(t *testing.T) {
		header := &testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 10
			},
			HasScheduledMiniBlocksCalled: func() bool {
				return false
			},
		}
		require.False(t, boot.shouldAllowRollback(header, finalBlockHash))
	})

	t.Run("should allow rollback of a final header without the same final hash", func(t *testing.T) {
		header := &testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 10
			},
			HasScheduledMiniBlocksCalled: func() bool {
				return false
			},
		}
		require.True(t, boot.shouldAllowRollback(header, notFinalBlockHash))
	})

	t.Run("should allow rollback of a final header if it holds scheduled miniBlocks", func(t *testing.T) {
		header := &testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 10
			},
			HasScheduledMiniBlocksCalled: func() bool {
				return true
			},
		}
		require.True(t, boot.shouldAllowRollback(header, finalBlockHash))
	})

	t.Run("should not allow rollback of a final header if it holds scheduled miniBlocks but no commit was done", func(t *testing.T) {
		firstBlockNonce.HasValue = false
		header := &testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 10
			},
			HasScheduledMiniBlocksCalled: func() bool {
				return true
			},
		}
		require.False(t, boot.shouldAllowRollback(header, finalBlockHash))
		firstBlockNonce.HasValue = true
	})

	t.Run("should not allow rollback of a final header if it holds scheduled miniBlocks but first committed nonce is higher", func(t *testing.T) {
		firstBlockNonce.Value = 11
		header := &testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 10
			},
			HasScheduledMiniBlocksCalled: func() bool {
				return true
			},
		}
		require.False(t, boot.shouldAllowRollback(header, finalBlockHash))
		firstBlockNonce.Value = 2
	})

	t.Run("should not allow any rollBack of a header if nonce is behind final", func(t *testing.T) {
		header := &testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 9
			},
			HasScheduledMiniBlocksCalled: func() bool {
				return true
			},
		}
		require.False(t, boot.shouldAllowRollback(header, finalBlockHash))
		require.False(t, boot.shouldAllowRollback(header, notFinalBlockHash))

		header.HasScheduledMiniBlocksCalled = func() bool {
			return false
		}
		require.False(t, boot.shouldAllowRollback(header, finalBlockHash))
		require.False(t, boot.shouldAllowRollback(header, notFinalBlockHash))
	})
}
