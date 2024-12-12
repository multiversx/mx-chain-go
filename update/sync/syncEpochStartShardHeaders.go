package sync

import (
	"context"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/update"
)

var _ update.PendingEpochStartShardHeaderSyncHandler = (*pendingEpochStartShardHeader)(nil)

type pendingEpochStartShardHeader struct {
	mutPending              sync.RWMutex
	epochStartHeader        data.HeaderHandler
	epochStartHash          []byte
	latestReceivedHeader    data.HeaderHandler
	latestReceivedHash      []byte
	targetEpoch             uint32
	targetShardId           uint32
	headersPool             dataRetriever.HeadersPool
	chReceived              chan bool
	chNew                   chan bool
	marshaller              marshal.Marshalizer
	stopSyncing             bool
	synced                  bool
	requestHandler          process.RequestHandler
	waitTimeBetweenRequests time.Duration
}

// ArgsPendingEpochStartShardHeaderSyncer defines the arguments needed for the sycner
type ArgsPendingEpochStartShardHeaderSyncer struct {
	HeadersPool    dataRetriever.HeadersPool
	Marshalizer    marshal.Marshalizer
	RequestHandler process.RequestHandler
}

// NewPendingEpochStartShardHeaderSyncer creates a syncer for all pending miniblocks
func NewPendingEpochStartShardHeaderSyncer(args ArgsPendingEpochStartShardHeaderSyncer) (*pendingEpochStartShardHeader, error) {
	if check.IfNil(args.HeadersPool) {
		return nil, update.ErrNilHeadersPool
	}
	if check.IfNil(args.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(args.RequestHandler) {
		return nil, process.ErrNilRequestHandler
	}

	p := &pendingEpochStartShardHeader{
		mutPending:              sync.RWMutex{},
		epochStartHeader:        nil,
		epochStartHash:          nil,
		targetEpoch:             0,
		targetShardId:           0,
		headersPool:             args.HeadersPool,
		chReceived:              make(chan bool),
		chNew:                   make(chan bool),
		requestHandler:          args.RequestHandler,
		stopSyncing:             true,
		synced:                  false,
		marshaller:              args.Marshalizer,
		waitTimeBetweenRequests: args.RequestHandler.RequestInterval(),
	}

	p.headersPool.RegisterHandler(p.receivedHeader)

	return p, nil
}

// SyncEpochStartShardHeader will sync the epoch start header for a specific shard
func (p *pendingEpochStartShardHeader) SyncEpochStartShardHeader(shardId uint32, epoch uint32, startNonce uint64, ctx context.Context) error {
	return p.syncEpochStartShardHeader(shardId, epoch, startNonce, ctx)
}

func (p *pendingEpochStartShardHeader) syncEpochStartShardHeader(shardId uint32, epoch uint32, startNonce uint64, ctx context.Context) error {
	_ = core.EmptyChannel(p.chReceived)
	_ = core.EmptyChannel(p.chNew)

	p.mutPending.Lock()
	p.stopSyncing = false
	p.targetEpoch = epoch
	p.targetShardId = shardId
	p.mutPending.Unlock()

	nonce := startNonce
	for {
		p.mutPending.Lock()
		p.stopSyncing = false
		p.requestHandler.RequestShardHeaderByNonce(shardId, nonce+1)
		p.mutPending.Unlock()

		select {
		case <-p.chReceived:
			p.mutPending.Lock()
			p.stopSyncing = true
			p.synced = true
			p.mutPending.Unlock()
			return nil
		case <-p.chNew:
			nonce = p.latestReceivedHeader.GetNonce()
			continue
		case <-ctx.Done():
			p.mutPending.Lock()
			p.stopSyncing = true
			p.mutPending.Unlock()
			return update.ErrTimeIsOut
		}
	}
}

// receivedHeader is a callback function when a new header was received
func (p *pendingEpochStartShardHeader) receivedHeader(header data.HeaderHandler, headerHash []byte) {
	p.mutPending.Lock()
	if p.stopSyncing {
		p.mutPending.Unlock()
		return
	}

	if header.GetShardID() != p.targetShardId {
		p.mutPending.Unlock()
		return
	}

	p.latestReceivedHash = headerHash
	p.latestReceivedHeader = header

	if header.GetEpoch() != p.targetEpoch || !header.IsStartOfEpochBlock() {
		p.mutPending.Unlock()
		p.chNew <- true
		return
	}

	p.epochStartHash = headerHash
	p.epochStartHeader = header
	p.mutPending.Unlock()

	p.chReceived <- true
}

// GetEpochStartHeader returns the synced epoch start header
func (p *pendingEpochStartShardHeader) GetEpochStartHeader() (data.HeaderHandler, []byte, error) {
	p.mutPending.RLock()
	defer p.mutPending.RUnlock()

	if !p.synced || p.epochStartHeader == nil || p.epochStartHash == nil {
		return nil, nil, update.ErrNotSynced
	}

	return p.epochStartHeader, p.epochStartHash, nil
}

// ClearFields will reset the state
func (p *pendingEpochStartShardHeader) ClearFields() {
	p.mutPending.Lock()
	p.epochStartHash = nil
	p.epochStartHeader = nil
	p.synced = false
	p.mutPending.Unlock()
}

// IsInterfaceNil returns nil if underlying object is nil
func (p *pendingEpochStartShardHeader) IsInterfaceNil() bool {
	return p == nil
}
