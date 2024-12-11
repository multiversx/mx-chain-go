package sync

import (
	"context"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/mock"
	"github.com/stretchr/testify/require"
)

func createMockArgsPendingEpochStartShardHeader() ArgsNewPendingEpochStartShardHeaderSyncer {

	return ArgsNewPendingEpochStartShardHeaderSyncer{
		HeadersPool:    &mock.HeadersCacherStub{},
		Marshalizer:    &mock.MarshalizerFake{},
		RequestHandler: &testscommon.RequestHandlerStub{},
	}
}

func TestNewPendingEpochStartShardHeaderSyncer(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingEpochStartShardHeader()

	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Nil(t, err)
	require.NotNil(t, syncer)
	require.False(t, syncer.IsInterfaceNil())
}

func TestNewPendingEpochStartShardHeaderSyncer_NilHeadersPool(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingEpochStartShardHeader()
	args.HeadersPool = nil

	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Equal(t, update.ErrNilHeadersPool, err)
	require.Nil(t, syncer)
}

func TestNewPendingEpochStartShardHeaderSyncer_NilMarshalizer(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingEpochStartShardHeader()
	args.Marshalizer = nil

	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Equal(t, dataRetriever.ErrNilMarshalizer, err)
	require.Nil(t, syncer)
}

func TestNewPendingEpochStartShardHeaderSyncer_NilRequestHandler(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingEpochStartShardHeader()
	args.RequestHandler = nil

	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Equal(t, process.ErrNilRequestHandler, err)
	require.Nil(t, syncer)
}

func TestSyncEpochStartShardHeader_Success(t *testing.T) {
	t.Parallel()

	shardID := uint32(1)
	epoch := uint32(10)
	startNonce := uint64(100)

	headerHash := []byte("headerHash")
	header := &block.Header{
		ShardID:            shardID,
		Nonce:              startNonce + 2,
		Epoch:              epoch,
		EpochStartMetaHash: []byte("metaHash"),
	}

	headersPool := &mock.HeadersCacherStub{}
	args := ArgsNewPendingEpochStartShardHeaderSyncer{
		HeadersPool: headersPool,
		Marshalizer: &mock.MarshalizerFake{},
		RequestHandler: &testscommon.RequestHandlerStub{
			RequestShardHeaderByNonceCalled: func(shardID uint32, nonce uint64) {},
		},
	}

	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Nil(t, err)

	// Simulate receiving headers
	go func() {
		// First received header not epoch start
		h1 := &block.Header{
			ShardID: shardID,
			Nonce:   startNonce + 1,
			Epoch:   epoch - 1,
		}
		syncer.receivedHeader(h1, []byte("hash1"))

		// Wait a bit, then receive epoch start header
		time.Sleep(100 * time.Millisecond)
		syncer.receivedHeader(header, headerHash)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)
	require.Nil(t, err)

	h, hHash, errGet := syncer.GetEpochStartHeader()
	require.Nil(t, errGet)
	require.Equal(t, header, h)
	require.Equal(t, headerHash, hHash)
}

func TestSyncEpochStartShardHeader_Timeout(t *testing.T) {
	t.Parallel()

	shardID := uint32(1)
	epoch := uint32(10)
	startNonce := uint64(100)

	headersPool := &mock.HeadersCacherStub{}
	args := ArgsNewPendingEpochStartShardHeaderSyncer{
		HeadersPool: headersPool,
		Marshalizer: &mock.MarshalizerFake{},
		RequestHandler: &testscommon.RequestHandlerStub{
			RequestShardHeaderByNonceCalled: func(shardID uint32, nonce uint64) {},
		},
	}

	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Nil(t, err)

	// Not sending any epoch start header; it should time out
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err = syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)
	require.Equal(t, update.ErrTimeIsOut, err)
}

func TestSyncEpochStartShardHeader_GetEpochStartHeaderNotSynced(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingEpochStartShardHeader()

	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Nil(t, err)

	_, _, errGet := syncer.GetEpochStartHeader()
	require.Equal(t, update.ErrNotSynced, errGet)
}

func TestSyncEpochStartShardHeader_ClearFields(t *testing.T) {
	t.Parallel()

	shardID := uint32(1)
	epoch := uint32(10)
	startNonce := uint64(100)

	headerHash := []byte("headerHash")
	header := &block.Header{
		ShardID:            shardID,
		Nonce:              startNonce + 1,
		Epoch:              epoch,
		EpochStartMetaHash: []byte("metaHash"),
	}

	headersPool := &mock.HeadersCacherStub{}
	args := ArgsNewPendingEpochStartShardHeaderSyncer{
		HeadersPool: headersPool,
		Marshalizer: &mock.MarshalizerFake{},
		RequestHandler: &testscommon.RequestHandlerStub{
			RequestShardHeaderByNonceCalled: func(shardID uint32, nonce uint64) {},
		},
	}

	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Nil(t, err)

	// Simulate receiving the epoch start header
	go func() {
		time.Sleep(100 * time.Millisecond)
		syncer.receivedHeader(header, headerHash)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)
	require.Nil(t, err)

	// Check fields before clear
	h, hHash, errGet := syncer.GetEpochStartHeader()
	require.Nil(t, errGet)
	require.Equal(t, header, h)
	require.Equal(t, headerHash, hHash)

	// Clear fields
	syncer.ClearFields()

	_, _, errGet = syncer.GetEpochStartHeader()
	require.Equal(t, update.ErrNotSynced, errGet)
}

func TestSyncEpochStartShardHeader_DifferentShardIDsShouldNotInterfere(t *testing.T) {
	t.Parallel()

	shardID := uint32(1)
	otherShardID := uint32(2)
	epoch := uint32(10)
	startNonce := uint64(100)

	headerHash := []byte("epochStartHash")
	header := &block.Header{
		ShardID:            shardID,
		Nonce:              startNonce + 2,
		Epoch:              epoch,
		EpochStartMetaHash: []byte("metaHash"),
	}

	headersPool := &mock.HeadersCacherStub{}
	args := ArgsNewPendingEpochStartShardHeaderSyncer{
		HeadersPool: headersPool,
		Marshalizer: &mock.MarshalizerFake{},
		RequestHandler: &testscommon.RequestHandlerStub{
			RequestShardHeaderByNonceCalled: func(shardID uint32, nonce uint64) {},
		},
	}

	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Nil(t, err)

	go func() {
		// Receive a header from a different shard - should be ignored
		differentShardHeader := &block.Header{
			ShardID:            otherShardID,
			Nonce:              startNonce + 1,
			Epoch:              epoch,
			EpochStartMetaHash: []byte("ignoreMetaHash"),
		}
		syncer.receivedHeader(differentShardHeader, []byte("ignoreHash"))

		// Wait and then send correct shard header
		time.Sleep(100 * time.Millisecond)
		syncer.receivedHeader(header, headerHash)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)
	require.Nil(t, err)

	h, hHash, errGet := syncer.GetEpochStartHeader()
	require.Nil(t, errGet)
	require.Equal(t, header, h)
	require.Equal(t, headerHash, hHash)
}

func TestSyncEpochStartShardHeader_NonEpochStartHeadersShouldTriggerNextAttempt(t *testing.T) {
	t.Parallel()

	shardID := uint32(1)
	epoch := uint32(10)
	startNonce := uint64(100)

	headerHash := []byte("epochStartHash")
	nonEpochStartHeader := &block.Header{
		ShardID:            shardID,
		Nonce:              startNonce + 1,
		Epoch:              epoch - 1, // not the target epoch
		EpochStartMetaHash: []byte("ignoreMetaHash"),
	}

	epochStartHeader := &block.Header{
		ShardID:            shardID,
		Nonce:              startNonce + 2,
		Epoch:              epoch,
		EpochStartMetaHash: []byte("metaHash"),
	}

	headersPool := &mock.HeadersCacherStub{}
	args := ArgsNewPendingEpochStartShardHeaderSyncer{
		HeadersPool: headersPool,
		Marshalizer: &mock.MarshalizerFake{},
		RequestHandler: &testscommon.RequestHandlerStub{
			RequestShardHeaderByNonceCalled: func(shardID uint32, nonce uint64) {},
		},
	}

	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Nil(t, err)

	go func() {
		// first receive non-epoch start header
		syncer.receivedHeader(nonEpochStartHeader, []byte("nonEpochStartHash"))

		// after a small delay, receive epoch start header
		time.Sleep(100 * time.Millisecond)
		syncer.receivedHeader(epochStartHeader, headerHash)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)
	require.Nil(t, err)

	h, hHash, errGet := syncer.GetEpochStartHeader()
	require.Nil(t, errGet)
	require.Equal(t, epochStartHeader, h)
	require.Equal(t, headerHash, hHash)
}

// Test no interface nil
func TestPendingEpochStartShardHeader_IsInterfaceNil(t *testing.T) {
	var p *pendingEpochStartShardHeader
	require.True(t, p.IsInterfaceNil())

	p = &pendingEpochStartShardHeader{}
	require.False(t, p.IsInterfaceNil())
}
