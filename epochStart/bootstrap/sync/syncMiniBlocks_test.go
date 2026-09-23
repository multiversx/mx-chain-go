package sync

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

func createMockArgsPendingMiniBlock() ArgsNewPendingMiniBlocksSyncer {
	return ArgsNewPendingMiniBlocksSyncer{
		Storage: &storageStubs.StorerStub{},
		Cache: &cache.CacherStub{
			RegisterHandlerCalled: func(f func(key []byte, val interface{})) {},
		},
		Marshalizer:    &mock.MarshalizerFake{},
		RequestHandler: &testscommon.RequestHandlerStub{},
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
	require.Equal(t, epochStart.ErrNilCacher, err)
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

func TestPendingMiniBlocks_SyncPendingMiniBlocksShouldRequestWithoutHoldingMutex(t *testing.T) {
	t.Parallel()

	localErr := errors.New("not found")
	ctx, cancel := context.WithCancel(context.Background())
	var pendingMiniBlocksSyncer *pendingMiniBlocks
	args := ArgsNewPendingMiniBlocksSyncer{
		Storage: &storageStubs.StorerStub{
			GetCalled: func(_ []byte) ([]byte, error) {
				return nil, localErr
			},
		},
		Cache: &cache.CacherStub{
			RegisterHandlerCalled: func(_ func(_ []byte, _ interface{})) {},
			PeekCalled: func(_ []byte) (interface{}, bool) {
				return nil, false
			},
		},
		Marshalizer: &mock.MarshalizerFake{},
		RequestHandler: &testscommon.RequestHandlerStub{
			RequestMiniBlockHandlerCalled: func(_ uint32, _ []byte) {
				require.True(t, pendingMiniBlocksSyncer.mutPendingMb.TryLock())
				pendingMiniBlocksSyncer.mutPendingMb.Unlock()
				cancel()
			},
		},
	}

	var err error
	pendingMiniBlocksSyncer, err = NewPendingMiniBlocksSyncer(args)
	require.NoError(t, err)

	err = pendingMiniBlocksSyncer.SyncPendingMiniBlocks(
		[]data.MiniBlockHeaderHandler{&block.MiniBlockHeader{Hash: []byte("mbHash")}},
		ctx,
	)
	require.ErrorIs(t, err, process.ErrTimeIsOut)
}

func TestPendingMiniBlocks_SyncPendingMiniBlocksShouldStopRequestPassWhenContextIsDone(t *testing.T) {
	t.Parallel()

	localErr := errors.New("not found")
	ctx, cancel := context.WithCancel(context.Background())
	numRequests := 0
	args := ArgsNewPendingMiniBlocksSyncer{
		Storage: &storageStubs.StorerStub{
			GetCalled: func(_ []byte) ([]byte, error) {
				return nil, localErr
			},
		},
		Cache: &cache.CacherStub{
			RegisterHandlerCalled: func(_ func(_ []byte, _ interface{})) {},
			PeekCalled: func(_ []byte) (interface{}, bool) {
				return nil, false
			},
		},
		Marshalizer: &mock.MarshalizerFake{},
		RequestHandler: &testscommon.RequestHandlerStub{
			RequestMiniBlockHandlerCalled: func(_ uint32, _ []byte) {
				numRequests++
				cancel()
			},
		},
	}

	pendingMiniBlocksSyncer, err := NewPendingMiniBlocksSyncer(args)
	require.NoError(t, err)

	err = pendingMiniBlocksSyncer.SyncPendingMiniBlocks(
		[]data.MiniBlockHeaderHandler{
			&block.MiniBlockHeader{Hash: []byte("mbHash0")},
			&block.MiniBlockHeader{Hash: []byte("mbHash1")},
			&block.MiniBlockHeader{Hash: []byte("mbHash2")},
		},
		ctx,
	)
	require.ErrorIs(t, err, process.ErrTimeIsOut)
	require.Equal(t, 1, numRequests)
}

func TestPendingMiniBlocks_ReceivedMiniBlockShouldNotBlockWhenCompletionIsAlreadySignaled(t *testing.T) {
	t.Parallel()

	pendingMiniBlocksSyncer := &pendingMiniBlocks{
		mapHashes: map[string]struct{}{
			"mbHash0": {},
			"mbHash1": {},
		},
		mapMiniBlocks: map[string]*block.MiniBlock{
			"mbHash0": {},
		},
		chReceivedAll: make(chan bool, 1),
	}
	pendingMiniBlocksSyncer.chReceivedAll <- true

	callbackDone := make(chan struct{})
	go func() {
		pendingMiniBlocksSyncer.receivedMiniBlock([]byte("mbHash1"), &block.MiniBlock{})
		close(callbackDone)
	}()

	select {
	case <-callbackDone:
	case <-time.After(time.Second):
		require.Fail(t, "received miniblock callback blocked on completion notification")
	}
}

func TestPendingMiniBlocks_SyncPendingMiniBlocksShouldIgnoreStaleCompletion(t *testing.T) {
	t.Parallel()

	localErr := errors.New("not found")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	completionSent := false
	var pendingMiniBlocksSyncer *pendingMiniBlocks

	args := ArgsNewPendingMiniBlocksSyncer{
		Storage: &storageStubs.StorerStub{
			GetCalled: func(_ []byte) ([]byte, error) {
				return nil, localErr
			},
		},
		Cache: &cache.CacherStub{
			RegisterHandlerCalled: func(_ func(_ []byte, _ interface{})) {},
			PeekCalled: func(_ []byte) (interface{}, bool) {
				return nil, false
			},
		},
		Marshalizer: &mock.MarshalizerFake{},
		RequestHandler: &testscommon.RequestHandlerStub{
			RequestMiniBlockHandlerCalled: func(_ uint32, _ []byte) {
				if completionSent {
					return
				}

				completionSent = true
				pendingMiniBlocksSyncer.chReceivedAll <- true
			},
		},
	}

	var err error
	pendingMiniBlocksSyncer, err = NewPendingMiniBlocksSyncer(args)
	require.NoError(t, err)

	err = pendingMiniBlocksSyncer.SyncPendingMiniBlocks(
		[]data.MiniBlockHeaderHandler{&block.MiniBlockHeader{Hash: []byte("mbHash")}},
		ctx,
	)
	require.ErrorIs(t, err, process.ErrTimeIsOut)
	require.False(t, pendingMiniBlocksSyncer.syncedAll)
}

func TestPendingMiniBlocks_SyncPendingMiniBlocksInPool(t *testing.T) {
	t.Parallel()

	mb := &block.MiniBlock{TxHashes: [][]byte{[]byte("tx1")}}
	args := createMockArgsPendingMiniBlock()
	args.Cache = &cache.CacherStub{
		RegisterHandlerCalled: func(_ func(_ []byte, _ interface{})) {},
		PeekCalled: func(key []byte) (interface{}, bool) {
			if string(key) == "hash1" {
				return mb, true
			}
			return nil, false
		},
	}

	syncer, err := NewPendingMiniBlocksSyncer(args)
	require.NoError(t, err)

	ctx := context.Background()
	err = syncer.SyncPendingMiniBlocks(
		[]data.MiniBlockHeaderHandler{&block.MiniBlockHeader{Hash: []byte("hash1")}},
		ctx,
	)
	require.NoError(t, err)

	mbs, err := syncer.GetMiniBlocks()
	require.NoError(t, err)
	require.Equal(t, 1, len(mbs))
	require.Equal(t, mb, mbs["hash1"])
}

func TestPendingMiniBlocks_GetMiniBlocksNotSynced(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingMiniBlock()
	syncer, err := NewPendingMiniBlocksSyncer(args)
	require.NoError(t, err)

	mbs, err := syncer.GetMiniBlocks()
	require.Equal(t, epochStart.ErrNotSynced, err)
	require.Nil(t, mbs)
}

func TestPendingMiniBlocks_ClearFields(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingMiniBlock()
	syncer, err := NewPendingMiniBlocksSyncer(args)
	require.NoError(t, err)

	syncer.mapHashes["hash1"] = struct{}{}
	syncer.mapMiniBlocks["hash1"] = &block.MiniBlock{}

	syncer.ClearFields()
	require.Equal(t, 0, len(syncer.mapHashes))
	require.Equal(t, 0, len(syncer.mapMiniBlocks))
}
