package sync

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/mock"
)

func createMockArgsPendingEpochStartShardHeader() ArgsPendingEpochStartShardHeaderSyncer {

	return ArgsPendingEpochStartShardHeaderSyncer{
		HeadersPool:         &mock.HeadersCacherStub{},
		Marshalizer:         &mock.MarshalizerFake{},
		RequestHandler:      &testscommon.RequestHandlerStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ProofsPool:          &dataRetrieverMocks.ProofsPoolMock{},
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
	proof := &block.HeaderProof{
		HeaderShardId: shardID,
		HeaderNonce:   startNonce + 2,
		HeaderHash:    headerHash,
		HeaderEpoch:   epoch,
	}

	args := createPendingEpochStartShardHeaderSyncerArgs()
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
		h1Hash := []byte("hash1")
		p1 := &block.HeaderProof{
			HeaderShardId: shardID,
			HeaderNonce:   startNonce + 1,
			HeaderHash:    h1Hash,
			HeaderEpoch:   epoch - 1,
		}
		syncer.receivedHeader(h1, h1Hash)
		syncer.receivedProof(p1)

		// Wait a bit, then receive epoch start header
		time.Sleep(100 * time.Millisecond)
		syncer.receivedHeader(header, headerHash)
		syncer.receivedProof(proof)
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

	args := createPendingEpochStartShardHeaderSyncerArgs()
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
	proof := &block.HeaderProof{
		HeaderShardId: shardID,
		HeaderNonce:   startNonce + 1,
		HeaderHash:    headerHash,
		HeaderEpoch:   epoch,
	}

	args := createPendingEpochStartShardHeaderSyncerArgs()
	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Nil(t, err)

	// Simulate receiving the epoch start header
	go func() {
		time.Sleep(100 * time.Millisecond)
		syncer.receivedHeader(header, headerHash)
		syncer.receivedProof(proof)
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
	proof := &block.HeaderProof{
		HeaderShardId: shardID,
		HeaderNonce:   startNonce + 2,
		HeaderHash:    headerHash,
		HeaderEpoch:   epoch,
	}

	args := createPendingEpochStartShardHeaderSyncerArgs()
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
		syncer.receivedProof(proof)
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
	nonEpochStartHeaderHash := []byte("nonEpochStartHash")
	nonEpochStartHeader := &block.Header{
		ShardID:            shardID,
		Nonce:              startNonce + 1,
		Epoch:              epoch - 1, // not the target epoch
		EpochStartMetaHash: []byte("ignoreMetaHash"),
	}
	nonEpochStartProof := &block.HeaderProof{
		HeaderShardId: shardID,
		HeaderNonce:   startNonce + 1,
		HeaderHash:    []byte("ignoreHash"),
		HeaderEpoch:   epoch - 1,
	}

	epochStartHeader := &block.Header{
		ShardID:            shardID,
		Nonce:              startNonce + 2,
		Epoch:              epoch,
		EpochStartMetaHash: []byte("metaHash"),
	}
	epochStartProof := &block.HeaderProof{
		HeaderShardId: shardID,
		HeaderNonce:   startNonce + 2,
		HeaderHash:    headerHash,
		HeaderEpoch:   epoch,
	}

	args := createPendingEpochStartShardHeaderSyncerArgs()
	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Nil(t, err)

	go func() {
		// first receive non-epoch start header
		syncer.receivedHeader(nonEpochStartHeader, nonEpochStartHeaderHash)
		syncer.receivedProof(nonEpochStartProof)

		// after a small delay, receive epoch start header
		time.Sleep(100 * time.Millisecond)
		syncer.receivedHeader(epochStartHeader, headerHash)
		syncer.receivedProof(epochStartProof)
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

func TestSyncEpochStartShardHeader_MultipleGoroutines(t *testing.T) {
	t.Parallel()

	shardID := uint32(1)
	epoch := uint32(10)
	startNonce := uint64(100)

	headerHash := []byte("correctEpochStartHash")
	epochStartHeader := &block.Header{
		ShardID:            shardID,
		Nonce:              startNonce + 5, // simulate a few attempts
		Epoch:              epoch,
		EpochStartMetaHash: []byte("methaHash"),
	}
	epochStartProof := &block.HeaderProof{
		HeaderShardId: shardID,
		HeaderNonce:   startNonce + 5,
		HeaderHash:    headerHash,
		HeaderEpoch:   epoch,
	}

	args := createPendingEpochStartShardHeaderSyncerArgs()
	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Nil(t, err)

	numGoroutines := 5

	// Provide the correct epoch start header after a small delay.
	go func() {
		time.Sleep(200 * time.Millisecond)
		syncer.receivedHeader(epochStartHeader, headerHash)
		syncer.receivedProof(epochStartProof)
	}()

	// Use a wait group to wait for all noise goroutines to complete.
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Goroutines sending noise headers
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()

			localShardID := shardID
			localEpoch := epoch
			// Some headers might come from different shards
			if i%2 == 0 {
				localShardID = uint32(2) // different shard
			}
			// Some headers might come from a different epoch
			if i%3 == 0 {
				localEpoch = epoch - 1 // different epoch
			}

			for nonce := startNonce + 1; nonce < startNonce+5; nonce++ {
				noiseHash := []byte("noiseHash")
				hdr := &block.Header{
					ShardID: localShardID,
					Nonce:   nonce,
					Epoch:   localEpoch,
				}
				noiseProof := &block.HeaderProof{
					HeaderShardId: localShardID,
					HeaderNonce:   nonce,
					HeaderHash:    noiseHash,
					HeaderEpoch:   localEpoch,
				}
				syncer.receivedHeader(hdr, noiseHash)
				syncer.receivedProof(noiseProof)
				time.Sleep(10 * time.Millisecond) // small delay between headers
			}
		}(i)
	}

	// Wait for all noise goroutines to finish
	wg.Wait()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)
	require.Nil(t, err, "Should succeed after receiving correct epoch start header")

	h, hHash, errGet := syncer.GetEpochStartHeader()
	require.Nil(t, errGet, "Should be able to retrieve the epoch start header after sync")
	require.Equal(t, epochStartHeader, h, "Should be the correct epoch start header")
	require.Equal(t, headerHash, hHash, "Hash should match the epoch start header hash")
}

// Test no interface nil
func TestPendingEpochStartShardHeader_IsInterfaceNil(t *testing.T) {
	var p *pendingEpochStartShardHeader
	require.True(t, p.IsInterfaceNil())

	p = &pendingEpochStartShardHeader{}
	require.False(t, p.IsInterfaceNil())
}

func createPendingEpochStartShardHeaderSyncerArgs() ArgsPendingEpochStartShardHeaderSyncer {
	headersPool := &mock.HeadersCacherStub{}
	proofsPool := &dataRetrieverMocks.ProofsPoolMock{}
	args := ArgsPendingEpochStartShardHeaderSyncer{
		HeadersPool: headersPool,
		Marshalizer: &mock.MarshalizerFake{},
		RequestHandler: &testscommon.RequestHandlerStub{
			RequestShardHeaderByNonceCalled: func(shardID uint32, nonce uint64) {},
		},
		ProofsPool: proofsPool,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AndromedaFlag
			},
		},
	}
	return args
}
