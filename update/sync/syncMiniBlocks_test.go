package sync

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
)

func TestNewPendingMiniBlocksSyncer(t *testing.T) {
	t.Parallel()

	args := ArgsNewPendingMiniBlocksSyncer{
		Storage: &mock.StorerStub{},
		Cache: &mock.CacherStub{
			RegisterHandlerCalled: func(f func(key []byte)) {},
		},
		Marshalizer:    &mock.MarshalizerFake{},
		RequestHandler: &mock.RequestHandlerStub{},
	}

	pendingMiniBlocksSyncer, err := NewPendingMiniBlocksSyncer(args)
	require.NotNil(t, pendingMiniBlocksSyncer)
	require.Nil(t, err)
}

func TestSyncPendingMiniBlocksFromMeta_MiniBlocksInPool(t *testing.T) {
	t.Parallel()

	miniBlockInPool := false
	mbHash := []byte("mbHash")
	mb := &block.MiniBlock{}
	args := ArgsNewPendingMiniBlocksSyncer{
		Storage: &mock.StorerStub{},
		Cache: &mock.CacherStub{
			RegisterHandlerCalled: func(f func(key []byte)) {},
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				miniBlockInPool = true
				return mb, true
			},
		},
		Marshalizer:    &mock.MarshalizerFake{},
		RequestHandler: &mock.RequestHandlerStub{},
	}

	pendingMiniBlocksSyncer, err := NewPendingMiniBlocksSyncer(args)
	require.Nil(t, err)

	metaBlock := &block.MetaBlock{
		Nonce: 1, Epoch: 1, RootHash: []byte("metaRootHash"),
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{ShardId: 0, RootHash: []byte("shardDataRootHash"),
					PendingMiniBlockHeaders: []block.ShardMiniBlockHeader{
						{Hash: mbHash},
					},
				},
			},
		},
	}
	err = pendingMiniBlocksSyncer.SyncPendingMiniBlocksFromMeta(metaBlock, time.Second)
	require.Nil(t, err)
	require.True(t, miniBlockInPool)

	miniBlocks, err := pendingMiniBlocksSyncer.GetMiniBlocks()
	require.Equal(t, mb, miniBlocks[string(mbHash)])
	require.Nil(t, err)
}

func TestSyncPendingMiniBlocksFromMeta_MiniBlocksInPoolMissingTimeout(t *testing.T) {
	t.Parallel()

	mbHash := []byte("mbHash")
	localErr := errors.New("not found")
	args := ArgsNewPendingMiniBlocksSyncer{
		Storage: &mock.StorerStub{
			GetCalled: func(key []byte) (bytes []byte, err error) {
				return nil, localErr
			},
			GetFromEpochCalled: func(key []byte, epoch uint32) (bytes []byte, err error) {
				return nil, localErr
			},
		},
		Cache: &mock.CacherStub{
			RegisterHandlerCalled: func(f func(key []byte)) {},
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		},
		Marshalizer:    &mock.MarshalizerFake{},
		RequestHandler: &mock.RequestHandlerStub{},
	}

	pendingMiniBlocksSyncer, err := NewPendingMiniBlocksSyncer(args)
	require.Nil(t, err)

	metaBlock := &block.MetaBlock{
		Nonce: 1, Epoch: 1, RootHash: []byte("metaRootHash"),
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{ShardId: 0, RootHash: []byte("shardDataRootHash"),
					PendingMiniBlockHeaders: []block.ShardMiniBlockHeader{
						{Hash: mbHash},
					},
				},
			},
		},
	}
	err = pendingMiniBlocksSyncer.SyncPendingMiniBlocksFromMeta(metaBlock, time.Second)
	require.Equal(t, process.ErrTimeIsOut, err)
}

func TestSyncPendingMiniBlocksFromMeta_MiniBlocksInPoolReceive(t *testing.T) {
	t.Parallel()

	mbHash := []byte("mbHash")
	mb := &block.MiniBlock{}
	localErr := errors.New("not found")
	args := ArgsNewPendingMiniBlocksSyncer{
		Storage: &mock.StorerStub{
			GetCalled: func(key []byte) (bytes []byte, err error) {
				return nil, localErr
			},
			GetFromEpochCalled: func(key []byte, epoch uint32) (bytes []byte, err error) {
				return nil, localErr
			},
		},
		Cache:          mock.NewCacherMock(),
		Marshalizer:    &mock.MarshalizerFake{},
		RequestHandler: &mock.RequestHandlerStub{},
	}

	pendingMiniBlocksSyncer, err := NewPendingMiniBlocksSyncer(args)
	require.Nil(t, err)

	metaBlock := &block.MetaBlock{
		Nonce: 1, Epoch: 1, RootHash: []byte("metaRootHash"),
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{ShardId: 0, RootHash: []byte("shardDataRootHash"),
					PendingMiniBlockHeaders: []block.ShardMiniBlockHeader{
						{Hash: mbHash},
					},
				},
			},
		},
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = pendingMiniBlocksSyncer.pool.Put(mbHash, mb)
	}()

	err = pendingMiniBlocksSyncer.SyncPendingMiniBlocksFromMeta(metaBlock, time.Second)
	require.Nil(t, err)
}
