package sync

import (
	"context"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/update"
)

// TODO: there is some duplicated code between this syncer and the other syncers in this package that could be refactored

var _ update.PendingEpochStartShardHeaderSyncHandler = (*pendingEpochStartShardHeader)(nil)

type proofedHeaderInfo struct {
	header data.HeaderHandler
	hash   []byte
}

type pendingEpochStartShardHeader struct {
	mutPending              sync.RWMutex
	epochStartHeader        data.HeaderHandler
	epochStartHash          []byte
	targetEpoch             uint32
	targetShardId           uint32
	expectedNonce           uint64
	candidates              map[string]data.HeaderHandler
	headersPool             dataRetriever.HeadersPool
	chProofedHeader         chan proofedHeaderInfo
	marshaller              marshal.Marshalizer
	stopSyncing             bool
	synced                  bool
	requestHandler          process.RequestHandler
	waitTimeBetweenRequests time.Duration
	enableEpochsHandler     common.EnableEpochsHandler
	proofsPool              dataRetriever.ProofsPool
}

// ArgsPendingEpochStartShardHeaderSyncer defines the arguments needed for the sycner
type ArgsPendingEpochStartShardHeaderSyncer struct {
	HeadersPool         dataRetriever.HeadersPool
	ProofsPool          dataRetriever.ProofsPool
	Marshalizer         marshal.Marshalizer
	RequestHandler      process.RequestHandler
	EnableEpochsHandler common.EnableEpochsHandler
}

// NewPendingEpochStartShardHeaderSyncer creates a syncer for all pending miniblocks
func NewPendingEpochStartShardHeaderSyncer(args ArgsPendingEpochStartShardHeaderSyncer) (*pendingEpochStartShardHeader, error) {
	if check.IfNil(args.HeadersPool) {
		return nil, update.ErrNilHeadersPool
	}
	if check.IfNil(args.ProofsPool) {
		return nil, dataRetriever.ErrNilProofsPool
	}
	if check.IfNil(args.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(args.RequestHandler) {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, update.ErrNilEnableEpochsHandler
	}

	p := &pendingEpochStartShardHeader{
		mutPending:              sync.RWMutex{},
		epochStartHeader:        nil,
		epochStartHash:          nil,
		targetEpoch:             0,
		targetShardId:           0,
		candidates:              make(map[string]data.HeaderHandler),
		headersPool:             args.HeadersPool,
		proofsPool:              args.ProofsPool,
		chProofedHeader:         make(chan proofedHeaderInfo, 8),
		requestHandler:          args.RequestHandler,
		stopSyncing:             true,
		synced:                  false,
		marshaller:              args.Marshalizer,
		waitTimeBetweenRequests: args.RequestHandler.RequestInterval(),
		enableEpochsHandler:     args.EnableEpochsHandler,
	}

	p.headersPool.RegisterHandler(p.receivedHeader)
	p.proofsPool.RegisterHandler(p.receivedProof)

	return p, nil
}

// SyncEpochStartShardHeader will sync the epoch start header for a specific shard
func (p *pendingEpochStartShardHeader) SyncEpochStartShardHeader(shardId uint32, epoch uint32, startNonce uint64, ctx context.Context) error {
	return p.syncEpochStartShardHeader(shardId, epoch, startNonce, ctx)
}

func (p *pendingEpochStartShardHeader) hasProof(shardID uint32, hash []byte, epoch uint32) bool {
	if !p.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, epoch) {
		return true
	}

	return p.proofsPool.HasProof(shardID, hash)
}

// syncEpochStartShardHeader walks the shard chain nonce by nonce from startNonce+1 up to the target
// epoch's start block; only proofed headers at the exact requested nonce advance the walk
func (p *pendingEpochStartShardHeader) syncEpochStartShardHeader(shardId uint32, epoch uint32, startNonce uint64, ctx context.Context) error {
	p.drainProofedHeaderChannel()

	p.mutPending.Lock()
	p.stopSyncing = false
	p.targetEpoch = epoch
	p.targetShardId = shardId
	p.expectedNonce = startNonce + 1
	p.candidates = make(map[string]data.HeaderHandler)
	p.mutPending.Unlock()

	defer func() {
		p.mutPending.Lock()
		p.stopSyncing = true
		p.mutPending.Unlock()
	}()

	for {
		p.mutPending.RLock()
		nonceToRequest := p.expectedNonce
		p.mutPending.RUnlock()

		p.requestHandler.RequestShardHeaderByNonce(shardId, nonceToRequest)
		p.requestHandler.RequestEquivalentProofByNonce(shardId, nonceToRequest)

		select {
		case info := <-p.chProofedHeader:
			done, err := p.processProofedHeader(info)
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		case <-time.After(p.waitTimeBetweenRequests):
			continue
		case <-ctx.Done():
			return update.ErrTimeIsOut
		}
	}
}

// processProofedHeader decides for a proofed header at the expected nonce: done, walked past, or advance
func (p *pendingEpochStartShardHeader) processProofedHeader(info proofedHeaderInfo) (bool, error) {
	p.mutPending.Lock()
	defer p.mutPending.Unlock()

	if check.IfNil(info.header) || info.header.GetNonce() != p.expectedNonce {
		return false, nil
	}

	isTargetEpochStart := info.header.GetEpoch() == p.targetEpoch && info.header.IsStartOfEpochBlock()
	if isTargetEpochStart {
		p.epochStartHeader = info.header
		p.epochStartHash = info.hash
		p.synced = true
		return true, nil
	}

	if info.header.GetEpoch() >= p.targetEpoch {
		log.Warn("pendingEpochStartShardHeader: walked past the target epoch start block",
			"shard", p.targetShardId,
			"target epoch", p.targetEpoch,
			"nonce", info.header.GetNonce(),
			"header epoch", info.header.GetEpoch())
		return false, update.ErrEpochStartShardHeaderNotFound
	}

	p.expectedNonce++
	p.candidates = make(map[string]data.HeaderHandler)

	return false, nil
}

func (p *pendingEpochStartShardHeader) drainProofedHeaderChannel() {
	for {
		select {
		case <-p.chProofedHeader:
		default:
			return
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

	isExpected := header.GetShardID() == p.targetShardId && header.GetNonce() == p.expectedNonce
	if !isExpected {
		p.mutPending.Unlock()
		return
	}

	p.candidates[string(headerHash)] = header
	if !p.hasProof(header.GetShardID(), headerHash, header.GetEpoch()) {
		go p.requestHandler.RequestEquivalentProofByHash(header.GetShardID(), headerHash)
		p.mutPending.Unlock()
		return
	}
	p.mutPending.Unlock()

	p.signalProofedHeader(header, headerHash)
}

// receivedProof is a callback function when a new proof was received
func (p *pendingEpochStartShardHeader) receivedProof(proof data.HeaderProofHandler) {
	p.mutPending.Lock()
	if p.stopSyncing {
		p.mutPending.Unlock()
		return
	}

	header, ok := p.candidates[string(proof.GetHeaderHash())]
	p.mutPending.Unlock()
	if !ok {
		return
	}

	p.signalProofedHeader(header, proof.GetHeaderHash())
}

func (p *pendingEpochStartShardHeader) signalProofedHeader(header data.HeaderHandler, hash []byte) {
	select {
	case p.chProofedHeader <- proofedHeaderInfo{header: header, hash: hash}:
	default:
	}
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
