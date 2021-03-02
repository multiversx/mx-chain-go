package sync

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/mock"
	"github.com/stretchr/testify/require"
)

func createMockArgsPendingMiniBlock() ArgsNewPendingMiniBlocksSyncer {
	return ArgsNewPendingMiniBlocksSyncer{
		Storage: &mock.StorerStub{},
		Cache: &testscommon.CacherStub{
			RegisterHandlerCalled: func(f func(key []byte, val interface{})) {},
		},
		Marshalizer:    &mock.MarshalizerFake{},
		RequestHandler: &mock.RequestHandlerStub{},
	}
}

func TestNewPendingMiniBlocksSyncer(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingMiniBlock()

	pendingMiniBlocksSyncer, err := NewPendingMiniBlocksSyncer(args)
	require.NotNil(t, pendingMiniBlocksSyncer)
	require.Nil(t, err)
	require.False(t, pendingMiniBlocksSyncer.IsInterfaceNil())
}

func TestNewPendingMiniBlocksSyncer_NilStorage(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingMiniBlock()
	args.Storage = nil

	pendingMiniBlocksSyncer, err := NewPendingMiniBlocksSyncer(args)
	require.Equal(t, dataRetriever.ErrNilHeadersStorage, err)
	require.Nil(t, pendingMiniBlocksSyncer)
}

func TestNewPendingMiniBlocksSyncer_NilCache(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingMiniBlock()
	args.Cache = nil

	pendingMiniBlocksSyncer, err := NewPendingMiniBlocksSyncer(args)
	require.Equal(t, update.ErrNilCacher, err)
	require.Nil(t, pendingMiniBlocksSyncer)
}

func TestNewPendingMiniBlocksSyncer_NilMarshalizer(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingMiniBlock()
	args.Marshalizer = nil

	pendingMiniBlocksSyncer, err := NewPendingMiniBlocksSyncer(args)
	require.Equal(t, dataRetriever.ErrNilMarshalizer, err)
	require.Nil(t, pendingMiniBlocksSyncer)
}

func TestNewPendingMiniBlocksSyncer_NilRequestHandler(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingMiniBlock()
	args.RequestHandler = nil

	pendingMiniBlocksSyncer, err := NewPendingMiniBlocksSyncer(args)
	require.Equal(t, process.ErrNilRequestHandler, err)
	require.Nil(t, pendingMiniBlocksSyncer)
}

func TestSyncPendingMiniBlocksFromMeta_MiniBlocksInPool(t *testing.T) {
	t.Parallel()

	miniBlockInPool := false
	mbHash := []byte("mbHash")
	mb := &block.MiniBlock{}
	args := ArgsNewPendingMiniBlocksSyncer{
		Storage: &mock.StorerStub{},
		Cache: &testscommon.CacherStub{
			RegisterHandlerCalled: func(f func(key []byte, val interface{})) {},
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
				{
					ShardID:                 0,
					RootHash:                []byte("shardDataRootHash"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{{Hash: mbHash}},
					FirstPendingMetaBlock:   []byte("firstPending"),
				},
			},
		},
	}
	unFinished := make(map[string]*block.MetaBlock)
	unFinished["firstPending"] = metaBlock
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err = pendingMiniBlocksSyncer.SyncPendingMiniBlocksFromMeta(metaBlock, unFinished, ctx)
	cancel()
	require.Nil(t, err)
	require.True(t, miniBlockInPool)

	miniBlocks, err := pendingMiniBlocksSyncer.GetMiniBlocks()
	require.Equal(t, mb, miniBlocks[string(mbHash)])
	require.Nil(t, err)
}

func TestSyncPendingMiniBlocksFromMeta_MiniBlocksInPoolWithRewards(t *testing.T) {
	t.Parallel()

	miniBlockInPool := false
	mbHash := []byte("mbHash")
	rwdMBHash := []byte("rwdMBHash")
	mb := &block.MiniBlock{}
	rwdMB := &block.MiniBlock{
		SenderShardID:   core.MetachainShardId,
		ReceiverShardID: 0,
		Type:            block.RewardsBlock,
	}
	args := ArgsNewPendingMiniBlocksSyncer{
		Storage: &mock.StorerStub{},
		Cache: &testscommon.CacherStub{
			RegisterHandlerCalled: func(f func(key []byte, val interface{})) {},
			PeekCalled: func(key []byte) (value interface{}, ok bool) {
				miniBlockInPool = true
				if bytes.Equal(key, mbHash) {
					return mb, true
				}
				if bytes.Equal(key, rwdMBHash) {
					return rwdMB, true
				}
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
				{
					ShardID:  0,
					RootHash: []byte("shardDataRootHash"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{Hash: mbHash},
						{Hash: rwdMBHash},
					},
					FirstPendingMetaBlock: []byte("firstPending"),
				},
			},
		},
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				SenderShardID:   core.MetachainShardId,
				ReceiverShardID: 0,
				Hash:            rwdMBHash,
				Type:            block.RewardsBlock,
			},
			{
				Type: block.PeerBlock,
			},
		},
	}
	unFinished := make(map[string]*block.MetaBlock)
	unFinished["firstPending"] = metaBlock
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err = pendingMiniBlocksSyncer.SyncPendingMiniBlocksFromMeta(metaBlock, unFinished, ctx)
	cancel()
	require.Nil(t, err)
	require.True(t, miniBlockInPool)

	miniBlocks, err := pendingMiniBlocksSyncer.GetMiniBlocks()
	require.Equal(t, mb, miniBlocks[string(mbHash)])
	require.Equal(t, rwdMB, miniBlocks[string(rwdMBHash)])
	require.Equal(t, 2, len(miniBlocks))
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
		Cache: &testscommon.CacherStub{
			RegisterHandlerCalled: func(f func(key []byte, val interface{})) {},
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
				{
					ShardID:                 0,
					RootHash:                []byte("shardDataRootHash"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{{Hash: mbHash}},
					FirstPendingMetaBlock:   []byte("firstPending"),
				},
			},
		},
	}
	unFinished := make(map[string]*block.MetaBlock)
	unFinished["firstPending"] = metaBlock
	// we need a value larger than the request interval as to also test what happens after the normal request interval has expired
	timeout := time.Second + time.Millisecond*500
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	err = pendingMiniBlocksSyncer.SyncPendingMiniBlocksFromMeta(metaBlock, unFinished, ctx)
	cancel()
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
		Cache:          testscommon.NewCacherMock(),
		Marshalizer:    &mock.MarshalizerFake{},
		RequestHandler: &mock.RequestHandlerStub{},
	}

	pendingMiniBlocksSyncer, err := NewPendingMiniBlocksSyncer(args)
	require.Nil(t, err)

	metaBlock := &block.MetaBlock{
		Nonce: 1, Epoch: 1, RootHash: []byte("metaRootHash"),
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					ShardID:                 0,
					RootHash:                []byte("shardDataRootHash"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{{Hash: mbHash}},
					FirstPendingMetaBlock:   []byte("firstPending"),
				},
			},
		},
	}
	unFinished := make(map[string]*block.MetaBlock)
	unFinished["firstPending"] = metaBlock

	go func() {
		time.Sleep(100 * time.Millisecond)
		_ = pendingMiniBlocksSyncer.pool.Put(mbHash, mb, mb.Size())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err = pendingMiniBlocksSyncer.SyncPendingMiniBlocksFromMeta(metaBlock, unFinished, ctx)
	cancel()
	require.Nil(t, err)
}

func TestSyncPendingMiniBlocksFromMeta_MiniBlocksInStorageReceive(t *testing.T) {
	t.Parallel()

	mbHash := []byte("mbHash")
	mb := &block.MiniBlock{}
	marshalizer := &mock.MarshalizerMock{}
	args := ArgsNewPendingMiniBlocksSyncer{
		Storage: &mock.StorerStub{
			GetCalled: func(key []byte) (bytes []byte, err error) {
				mbBytes, _ := marshalizer.Marshal(mb)
				return mbBytes, nil
			},
		},
		Cache: &testscommon.CacherStub{
			RegisterHandlerCalled: func(_ func(_ []byte, _ interface{})) {},
			PeekCalled: func(key []byte) (interface{}, bool) {
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
				{
					ShardID:                 0,
					RootHash:                []byte("shardDataRootHash"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{{Hash: mbHash}},
					FirstPendingMetaBlock:   []byte("firstPending"),
				},
			},
		},
	}
	unFinished := make(map[string]*block.MetaBlock)
	unFinished["firstPending"] = metaBlock

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err = pendingMiniBlocksSyncer.SyncPendingMiniBlocksFromMeta(metaBlock, unFinished, ctx)
	cancel()
	require.Nil(t, err)
}

func TestSyncPendingMiniBlocksFromMeta_GetMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	mbHash := []byte("mbHash")
	mb := &block.MiniBlock{
		TxHashes: [][]byte{[]byte("txHash")},
	}
	localErr := errors.New("not found")
	marshalizer := &mock.MarshalizerMock{}
	args := ArgsNewPendingMiniBlocksSyncer{
		Storage: &mock.StorerStub{
			GetCalled: func(key []byte) (bytes []byte, err error) {
				mbBytes, _ := marshalizer.Marshal(mb)
				return mbBytes, nil
			},
			GetFromEpochCalled: func(key []byte, epoch uint32) (bytes []byte, err error) {
				return nil, localErr
			},
		},
		Cache: &testscommon.CacherStub{
			RegisterHandlerCalled: func(_ func(_ []byte, _ interface{})) {},
			PeekCalled: func(key []byte) (interface{}, bool) {
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
				{
					ShardID:                 0,
					RootHash:                []byte("shardDataRootHash"),
					PendingMiniBlockHeaders: []block.MiniBlockHeader{{Hash: mbHash}},
					FirstPendingMetaBlock:   []byte("firstPending"),
				},
			},
		},
	}
	unFinished := make(map[string]*block.MetaBlock)
	unFinished["firstPending"] = metaBlock

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err = pendingMiniBlocksSyncer.SyncPendingMiniBlocksFromMeta(metaBlock, unFinished, ctx)
	cancel()
	require.Nil(t, err)

	res, err := pendingMiniBlocksSyncer.GetMiniBlocks()
	require.NoError(t, err)
	require.Equal(t, mb, res[string(mbHash)])
}
