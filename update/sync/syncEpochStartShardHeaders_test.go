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

// createSyncerWithRequestCapture wires the request handler to publish every requested shard nonce
// exactly once on the returned channel, so tests can deliver responses event-driven
func createSyncerWithRequestCapture(t *testing.T) (*pendingEpochStartShardHeader, chan uint64) {
	requestedNonces := make(chan uint64, 100)
	seen := make(map[uint64]struct{})
	var mutSeen sync.Mutex

	args := createPendingEpochStartShardHeaderSyncerArgs()
	args.RequestHandler = &testscommon.RequestHandlerStub{
		RequestShardHeaderByNonceCalled: func(shardID uint32, nonce uint64) {
			mutSeen.Lock()
			defer mutSeen.Unlock()
			if _, ok := seen[nonce]; ok {
				return
			}
			seen[nonce] = struct{}{}
			requestedNonces <- nonce
		},
	}

	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Nil(t, err)

	return syncer, requestedNonces
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

func TestNewPendingEpochStartShardHeaderSyncer_NilProofsPool(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingEpochStartShardHeader()
	args.ProofsPool = nil

	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Equal(t, process.ErrNilProofsPool, err)
	require.Nil(t, syncer)
}

func TestNewPendingEpochStartShardHeaderSyncer_NilEnableEpochsHandler(t *testing.T) {
	t.Parallel()

	args := createMockArgsPendingEpochStartShardHeader()
	args.EnableEpochsHandler = nil

	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Equal(t, process.ErrNilEnableEpochsHandler, err)
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

	syncer, requestedNonces := createSyncerWithRequestCapture(t)

	go func() {
		for nonce := range requestedNonces {
			switch nonce {
			case startNonce + 1:
				h1Hash := []byte("hash1")
				h1 := &block.Header{ShardID: shardID, Nonce: startNonce + 1, Epoch: epoch - 1}
				p1 := &block.HeaderProof{HeaderShardId: shardID, HeaderNonce: startNonce + 1, HeaderHash: h1Hash, HeaderEpoch: epoch - 1}
				syncer.receivedHeader(h1, h1Hash)
				syncer.receivedProof(p1)
			case startNonce + 2:
				syncer.receivedHeader(header, headerHash)
				syncer.receivedProof(proof)
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)
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

	syncer, requestedNonces := createSyncerWithRequestCapture(t)

	go func() {
		for nonce := range requestedNonces {
			if nonce == startNonce+1 {
				syncer.receivedHeader(header, headerHash)
				syncer.receivedProof(proof)
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)
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

	syncer, requestedNonces := createSyncerWithRequestCapture(t)

	go func() {
		for nonce := range requestedNonces {
			if nonce != startNonce+1 {
				continue
			}
			// same nonce, different shard - must be ignored
			differentShardHeader := &block.Header{
				ShardID:            otherShardID,
				Nonce:              startNonce + 1,
				Epoch:              epoch,
				EpochStartMetaHash: []byte("ignoreMetaHash"),
			}
			syncer.receivedHeader(differentShardHeader, []byte("ignoreHash"))

			syncer.receivedHeader(header, headerHash)
			syncer.receivedProof(proof)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)
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
		ShardID: shardID,
		Nonce:   startNonce + 1,
		Epoch:   epoch - 1, // not the target epoch
	}
	nonEpochStartProof := &block.HeaderProof{
		HeaderShardId: shardID,
		HeaderNonce:   startNonce + 1,
		HeaderHash:    nonEpochStartHeaderHash,
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

	syncer, requestedNonces := createSyncerWithRequestCapture(t)

	go func() {
		for nonce := range requestedNonces {
			switch nonce {
			case startNonce + 1:
				syncer.receivedHeader(nonEpochStartHeader, nonEpochStartHeaderHash)
				syncer.receivedProof(nonEpochStartProof)
			case startNonce + 2:
				syncer.receivedHeader(epochStartHeader, headerHash)
				syncer.receivedProof(epochStartProof)
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)
	require.Nil(t, err)

	h, hHash, errGet := syncer.GetEpochStartHeader()
	require.Nil(t, errGet)
	require.Equal(t, epochStartHeader, h)
	require.Equal(t, headerHash, hHash)
}

// regression for the shuffle-bootstrap hijack: an unsolicited live tip header (proofed, target
// epoch, not an epoch start) must not move the walk cursor past the target epoch start block
func TestSyncEpochStartShardHeader_UnsolicitedTipHeaderDoesNotHijackWalk(t *testing.T) {
	t.Parallel()

	shardID := uint32(1)
	epoch := uint32(10)
	startNonce := uint64(100)
	tipNonce := uint64(500)

	tipHash := []byte("tipHash")
	tipHeader := &block.Header{ShardID: shardID, Nonce: tipNonce, Epoch: epoch}
	tipProof := &block.HeaderProof{HeaderShardId: shardID, HeaderNonce: tipNonce, HeaderHash: tipHash, HeaderEpoch: epoch}

	headerHash := []byte("epochStartHash")
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

	syncer, requestedNonces := createSyncerWithRequestCapture(t)

	var mutRequested sync.Mutex
	requested := make([]uint64, 0)

	go func() {
		for nonce := range requestedNonces {
			mutRequested.Lock()
			requested = append(requested, nonce)
			mutRequested.Unlock()

			// live gossip interleaved with every response
			syncer.receivedHeader(tipHeader, tipHash)
			syncer.receivedProof(tipProof)

			switch nonce {
			case startNonce + 1:
				h1Hash := []byte("hash1")
				h1 := &block.Header{ShardID: shardID, Nonce: startNonce + 1, Epoch: epoch - 1}
				p1 := &block.HeaderProof{HeaderShardId: shardID, HeaderNonce: startNonce + 1, HeaderHash: h1Hash, HeaderEpoch: epoch - 1}
				syncer.receivedHeader(h1, h1Hash)
				syncer.receivedProof(p1)

				syncer.receivedHeader(tipHeader, tipHash)
				syncer.receivedProof(tipProof)
			case startNonce + 2:
				syncer.receivedHeader(epochStartHeader, headerHash)
				syncer.receivedProof(epochStartProof)
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)
	require.Nil(t, err)

	h, hHash, errGet := syncer.GetEpochStartHeader()
	require.Nil(t, errGet)
	require.Equal(t, epochStartHeader, h)
	require.Equal(t, headerHash, hHash)

	mutRequested.Lock()
	defer mutRequested.Unlock()
	for _, nonce := range requested {
		require.LessOrEqual(t, nonce, startNonce+2, "walk cursor was hijacked past the target")
	}
}

// a proofed header of the target epoch that is not an epoch start at the expected nonce proves the
// walk started past the target; the syncer must fail fast instead of chasing the tip forever
func TestSyncEpochStartShardHeader_WalkedPastTargetReturnsError(t *testing.T) {
	t.Parallel()

	shardID := uint32(1)
	epoch := uint32(10)
	startNonce := uint64(100)

	pastHash := []byte("pastHash")
	pastHeader := &block.Header{ShardID: shardID, Nonce: startNonce + 1, Epoch: epoch}
	pastProof := &block.HeaderProof{HeaderShardId: shardID, HeaderNonce: startNonce + 1, HeaderHash: pastHash, HeaderEpoch: epoch}

	syncer, requestedNonces := createSyncerWithRequestCapture(t)

	go func() {
		for nonce := range requestedNonces {
			if nonce == startNonce+1 {
				syncer.receivedHeader(pastHeader, pastHash)
				syncer.receivedProof(pastProof)
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)
	require.Equal(t, update.ErrEpochStartShardHeaderNotFound, err)

	_, _, errGet := syncer.GetEpochStartHeader()
	require.Equal(t, update.ErrNotSynced, errGet)
}

func TestSyncEpochStartShardHeader_ConcurrentNoiseDoesNotDerailWalk(t *testing.T) {
	t.Parallel()

	shardID := uint32(1)
	epoch := uint32(10)
	startNonce := uint64(100)
	targetNonce := startNonce + 5

	headerHash := []byte("correctEpochStartHash")
	epochStartHeader := &block.Header{
		ShardID:            shardID,
		Nonce:              targetNonce,
		Epoch:              epoch,
		EpochStartMetaHash: []byte("metaHash"),
	}
	epochStartProof := &block.HeaderProof{
		HeaderShardId: shardID,
		HeaderNonce:   targetNonce,
		HeaderHash:    headerHash,
		HeaderEpoch:   epoch,
	}

	syncer, requestedNonces := createSyncerWithRequestCapture(t)

	stopNoise := make(chan struct{})
	var wgNoise sync.WaitGroup

	// concurrent noise: other-shard headers at any nonce and same-shard live tip headers, all proofed
	numNoiseGoroutines := 4
	wgNoise.Add(numNoiseGoroutines)
	for i := 0; i < numNoiseGoroutines; i++ {
		go func(i int) {
			defer wgNoise.Done()

			noiseShard := shardID
			noiseNonce := uint64(700 + i) // same-shard tip noise
			if i%2 == 0 {
				noiseShard = uint32(2) // other-shard noise
				noiseNonce = startNonce + uint64(i)
			}
			noiseHash := []byte{byte(i), 0x01, 0x02}
			hdr := &block.Header{ShardID: noiseShard, Nonce: noiseNonce, Epoch: epoch}
			noiseProof := &block.HeaderProof{HeaderShardId: noiseShard, HeaderNonce: noiseNonce, HeaderHash: noiseHash, HeaderEpoch: epoch}
			for {
				select {
				case <-stopNoise:
					return
				default:
					syncer.receivedHeader(hdr, noiseHash)
					syncer.receivedProof(noiseProof)
					time.Sleep(2 * time.Millisecond)
				}
			}
		}(i)
	}

	go func() {
		for nonce := range requestedNonces {
			if nonce >= startNonce+1 && nonce < targetNonce {
				chainHash := []byte{0x10, byte(nonce)}
				hdr := &block.Header{ShardID: shardID, Nonce: nonce, Epoch: epoch - 1}
				chainProof := &block.HeaderProof{HeaderShardId: shardID, HeaderNonce: nonce, HeaderHash: chainHash, HeaderEpoch: epoch - 1}
				syncer.receivedHeader(hdr, chainHash)
				syncer.receivedProof(chainProof)
			}
			if nonce == targetNonce {
				syncer.receivedHeader(epochStartHeader, headerHash)
				syncer.receivedProof(epochStartProof)
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)

	close(stopNoise)
	wgNoise.Wait()

	require.Nil(t, err, "should succeed despite concurrent noise")

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

func TestSyncEpochStartShardHeader_HeadersAtWrongNonceAreIgnored(t *testing.T) {
	t.Parallel()

	shardID := uint32(1)
	epoch := uint32(10)
	startNonce := uint64(100)

	// target-looking header, but two nonces ahead of the walk: must be ignored
	headerHash := []byte("headerHash")
	epochStartHeader := &block.Header{
		ShardID:            shardID,
		Nonce:              startNonce + 3,
		Epoch:              epoch,
		EpochStartMetaHash: []byte("metaHash"),
	}
	epochStartProof := &block.HeaderProof{
		HeaderShardId: shardID,
		HeaderNonce:   startNonce + 3,
		HeaderHash:    headerHash,
		HeaderEpoch:   epoch,
	}

	syncer, requestedNonces := createSyncerWithRequestCapture(t)

	go func() {
		for range requestedNonces {
			syncer.receivedHeader(epochStartHeader, headerHash)
			syncer.receivedProof(epochStartProof)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)
	require.Equal(t, update.ErrTimeIsOut, err)

	_, _, errGet := syncer.GetEpochStartHeader()
	require.Equal(t, update.ErrNotSynced, errGet)
}

func TestSyncEpochStartShardHeader_ProofsBeforeHeaderShouldWork(t *testing.T) {
	t.Parallel()

	shardID := uint32(1)
	epoch := uint32(10)
	startNonce := uint64(100)

	headerHash := []byte("epochStartHash")

	epochStartHeader := &block.Header{
		ShardID:            shardID,
		Nonce:              startNonce + 1,
		Epoch:              epoch,
		EpochStartMetaHash: []byte("metaHash"),
	}
	epochStartProof := &block.HeaderProof{
		HeaderShardId: shardID,
		HeaderNonce:   startNonce + 1,
		HeaderHash:    headerHash,
		HeaderEpoch:   epoch,
	}

	requestedNonces := make(chan uint64, 100)
	seen := make(map[uint64]struct{})
	var mutSeen sync.Mutex

	headersPool := &mock.HeadersCacherStub{}
	proofsPool := &dataRetrieverMocks.ProofsPoolMock{
		HasProofCalled: func(shardID uint32, headerHash []byte) bool {
			return true
		},
	}
	args := ArgsPendingEpochStartShardHeaderSyncer{
		HeadersPool: headersPool,
		Marshalizer: &mock.MarshalizerFake{},
		RequestHandler: &testscommon.RequestHandlerStub{
			RequestShardHeaderByNonceCalled: func(shardID uint32, nonce uint64) {
				mutSeen.Lock()
				defer mutSeen.Unlock()
				if _, ok := seen[nonce]; ok {
					return
				}
				seen[nonce] = struct{}{}
				requestedNonces <- nonce
			},
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
	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Nil(t, err)

	go func() {
		for nonce := range requestedNonces {
			if nonce == startNonce+1 {
				// proof already in pool (HasProof = true); header alone must complete the nonce
				syncer.receivedProof(epochStartProof)
				syncer.receivedHeader(epochStartHeader, headerHash)
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)
	require.Nil(t, err)

	h, hHash, errGet := syncer.GetEpochStartHeader()
	require.Nil(t, errGet)
	require.Equal(t, epochStartHeader, h)
	require.Equal(t, headerHash, hHash)
}

func TestSyncEpochStartShardHeader_ShouldWorkWithoutAndromedaActivated(t *testing.T) {
	t.Parallel()

	shardID := uint32(1)
	epoch := uint32(10)
	startNonce := uint64(100)

	headerHash := []byte("epochStartHash")

	epochStartHeader := &block.Header{
		ShardID:            shardID,
		Nonce:              startNonce + 1,
		Epoch:              epoch,
		EpochStartMetaHash: []byte("metaHash"),
	}

	requestedNonces := make(chan uint64, 100)
	seen := make(map[uint64]struct{})
	var mutSeen sync.Mutex

	headersPool := &mock.HeadersCacherStub{}
	proofsPool := &dataRetrieverMocks.ProofsPoolMock{
		HasProofCalled: func(shardID uint32, headerHash []byte) bool {
			return false
		},
	}
	args := ArgsPendingEpochStartShardHeaderSyncer{
		HeadersPool: headersPool,
		Marshalizer: &mock.MarshalizerFake{},
		RequestHandler: &testscommon.RequestHandlerStub{
			RequestShardHeaderByNonceCalled: func(shardID uint32, nonce uint64) {
				mutSeen.Lock()
				defer mutSeen.Unlock()
				if _, ok := seen[nonce]; ok {
					return
				}
				seen[nonce] = struct{}{}
				requestedNonces <- nonce
			},
		},
		ProofsPool: proofsPool,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return false
			},
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.AndromedaFlag
			},
		},
	}
	syncer, err := NewPendingEpochStartShardHeaderSyncer(args)
	require.Nil(t, err)

	go func() {
		for nonce := range requestedNonces {
			if nonce == startNonce+1 {
				// pre-Andromeda: no proof needed, the header alone completes the nonce
				syncer.receivedHeader(epochStartHeader, headerHash)
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = syncer.SyncEpochStartShardHeader(shardID, epoch, startNonce, ctx)
	require.Nil(t, err)

	h, hHash, errGet := syncer.GetEpochStartHeader()
	require.Nil(t, errGet)
	require.Equal(t, epochStartHeader, h)
	require.Equal(t, headerHash, hHash)
}
