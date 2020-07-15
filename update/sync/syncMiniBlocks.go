package sync

import (
	"context"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
)

var _ update.EpochStartPendingMiniBlocksSyncHandler = (*pendingMiniBlocks)(nil)

type pendingMiniBlocks struct {
	mutPendingMb            sync.Mutex
	mapMiniBlocks           map[string]*block.MiniBlock
	mapHashes               map[string]struct{}
	pool                    storage.Cacher
	storage                 update.HistoryStorer
	chReceivedAll           chan bool
	marshalizer             marshal.Marshalizer
	stopSyncing             bool
	syncedAll               bool
	requestHandler          process.RequestHandler
	waitTimeBetweenRequests time.Duration
}

// ArgsNewPendingMiniBlocksSyncer defines the arguments needed for the sycner
type ArgsNewPendingMiniBlocksSyncer struct {
	Storage        storage.Storer
	Cache          storage.Cacher
	Marshalizer    marshal.Marshalizer
	RequestHandler process.RequestHandler
}

// NewPendingMiniBlocksSyncer creates a syncer for all pending miniblocks
func NewPendingMiniBlocksSyncer(args ArgsNewPendingMiniBlocksSyncer) (*pendingMiniBlocks, error) {
	if check.IfNil(args.Storage) {
		return nil, dataRetriever.ErrNilHeadersStorage
	}
	if check.IfNil(args.Cache) {
		return nil, update.ErrNilCacher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(args.RequestHandler) {
		return nil, process.ErrNilRequestHandler
	}

	p := &pendingMiniBlocks{
		mutPendingMb:            sync.Mutex{},
		mapMiniBlocks:           make(map[string]*block.MiniBlock),
		mapHashes:               make(map[string]struct{}),
		pool:                    args.Cache,
		storage:                 args.Storage,
		chReceivedAll:           make(chan bool),
		requestHandler:          args.RequestHandler,
		stopSyncing:             true,
		syncedAll:               false,
		marshalizer:             args.Marshalizer,
		waitTimeBetweenRequests: args.RequestHandler.RequestInterval(),
	}

	p.pool.RegisterHandler(p.receivedMiniBlock, core.UniqueIdentifier())

	return p, nil
}

// SyncPendingMiniBlocksFromMeta syncs the pending miniblocks from an epoch start metaBlock
func (p *pendingMiniBlocks) SyncPendingMiniBlocksFromMeta(epochStart *block.MetaBlock, unFinished map[string]*block.MetaBlock, ctx context.Context) error {
	if !epochStart.IsStartOfEpochBlock() && epochStart.Nonce > 0 {
		return update.ErrNotEpochStartBlock
	}
	if unFinished == nil {
		return update.ErrWrongUnfinishedMetaHdrsMap
	}

	for hash, meta := range unFinished {
		log.Debug("syncing miniblocks from unFinished meta", "hash", []byte(hash), "nonce", meta.Nonce)
	}

	listPendingMiniBlocks := make([]block.MiniBlockHeader, 0)
	nonceToHashMap := p.createNonceToHashMap(unFinished)

	for _, shardData := range epochStart.EpochStart.LastFinalizedHeaders {
		computedPending, err := p.computePendingMiniBlocksFromUnFinished(shardData, unFinished, nonceToHashMap, epochStart.GetNonce())
		if err != nil {
			return err
		}

		listPendingMiniBlocks = append(listPendingMiniBlocks, computedPending...)
	}

	for _, val := range epochStart.MiniBlockHeaders {
		if val.SenderShardID != core.MetachainShardId || val.Type == block.PeerBlock {
			continue
		}
		listPendingMiniBlocks = append(listPendingMiniBlocks, val)
	}

	return p.syncMiniBlocks(listPendingMiniBlocks, ctx)
}

// SyncPendingMiniBlocks will sync the miniblocks for the given epoch start meta block
func (p *pendingMiniBlocks) SyncPendingMiniBlocks(miniBlockHeaders []block.MiniBlockHeader, ctx context.Context) error {
	return p.syncMiniBlocks(miniBlockHeaders, ctx)
}

func (p *pendingMiniBlocks) syncMiniBlocks(listPendingMiniBlocks []block.MiniBlockHeader, ctx context.Context) error {
	_ = core.EmptyChannel(p.chReceivedAll)

	mapHashesToRequest := make(map[string]uint32)
	for _, mbHeader := range listPendingMiniBlocks {
		mapHashesToRequest[string(mbHeader.Hash)] = mbHeader.SenderShardID
	}

	p.mutPendingMb.Lock()
	p.stopSyncing = false
	p.mutPendingMb.Unlock()

	for {
		requestedMBs := 0
		p.mutPendingMb.Lock()
		p.stopSyncing = false
		for hash, shardId := range mapHashesToRequest {
			if _, ok := p.mapMiniBlocks[hash]; ok {
				delete(mapHashesToRequest, hash)
			}

			p.mapHashes[hash] = struct{}{}
			miniBlock, ok := p.getMiniBlockFromPoolOrStorage([]byte(hash))
			if ok {
				p.mapMiniBlocks[hash] = miniBlock
				delete(mapHashesToRequest, hash)
				continue
			}

			p.requestHandler.RequestMiniBlock(shardId, []byte(hash))
			requestedMBs++
		}
		p.mutPendingMb.Unlock()

		if requestedMBs == 0 {
			p.mutPendingMb.Lock()
			p.stopSyncing = true
			p.syncedAll = true
			p.mutPendingMb.Unlock()
			return nil
		}

		select {
		case <-p.chReceivedAll:
			p.mutPendingMb.Lock()
			p.stopSyncing = true
			p.syncedAll = true
			p.mutPendingMb.Unlock()
			return nil
		case <-time.After(p.waitTimeBetweenRequests):
			continue
		case <-ctx.Done():
			p.mutPendingMb.Lock()
			p.stopSyncing = true
			p.mutPendingMb.Unlock()
			return update.ErrTimeIsOut
		}
	}
}

func (p *pendingMiniBlocks) createNonceToHashMap(unFinished map[string]*block.MetaBlock) map[uint64]string {
	nonceToHash := make(map[uint64]string, len(unFinished))
	for hash, meta := range unFinished {
		nonceToHash[meta.GetNonce()] = hash
	}

	return nonceToHash
}

func (p *pendingMiniBlocks) computePendingMiniBlocksFromUnFinished(
	shardData block.EpochStartShardData,
	unFinished map[string]*block.MetaBlock,
	nonceToHash map[uint64]string,
	epochStartNonce uint64,
) ([]block.MiniBlockHeader, error) {
	pending := make([]block.MiniBlockHeader, 0)
	pending = append(pending, shardData.PendingMiniBlockHeaders...)

	firstPendingMeta, ok := unFinished[string(shardData.FirstPendingMetaBlock)]
	if !ok {
		return nil, update.ErrWrongUnfinishedMetaHdrsMap
	}

	firstUnFinishedNonce := firstPendingMeta.GetNonce()
	for nonce := firstUnFinishedNonce + 1; nonce <= epochStartNonce; nonce++ {
		metaHash, exists := nonceToHash[nonce]
		if !exists {
			return nil, update.ErrWrongUnfinishedMetaHdrsMap
		}

		meta, exists := unFinished[metaHash]
		if !exists {
			return nil, update.ErrWrongUnfinishedMetaHdrsMap
		}

		pendingFromCurrentMeta := getAllMiniBlocksWithDst(meta, shardData.ShardID)
		pending = append(pending, pendingFromCurrentMeta...)
	}

	return pending, nil
}

func getAllMiniBlocksWithDst(m *block.MetaBlock, destId uint32) []block.MiniBlockHeader {
	mbHdrs := make([]block.MiniBlockHeader, 0)
	for i := 0; i < len(m.ShardInfo); i++ {
		if m.ShardInfo[i].ShardID == destId {
			continue
		}

		for _, val := range m.ShardInfo[i].ShardMiniBlockHeaders {
			if val.ReceiverShardID == destId && val.SenderShardID != destId {
				mbHdrs = append(mbHdrs, val)
			}
		}
	}

	for _, val := range m.MiniBlockHeaders {
		if val.ReceiverShardID == destId && val.SenderShardID != destId {
			mbHdrs = append(mbHdrs, val)
		}
	}

	return mbHdrs
}

// receivedMiniBlock is a callback function when a new miniblock was received
// it will further ask for missing transactions
func (p *pendingMiniBlocks) receivedMiniBlock(miniBlockHash []byte, val interface{}) {
	p.mutPendingMb.Lock()
	if p.stopSyncing {
		p.mutPendingMb.Unlock()
		return
	}

	if _, ok := p.mapHashes[string(miniBlockHash)]; !ok {
		p.mutPendingMb.Unlock()
		return
	}

	if _, ok := p.mapMiniBlocks[string(miniBlockHash)]; ok {
		p.mutPendingMb.Unlock()
		return
	}

	miniBlock, ok := val.(*block.MiniBlock)
	if !ok {
		p.mutPendingMb.Unlock()
		return
	}

	p.mapMiniBlocks[string(miniBlockHash)] = miniBlock
	receivedAll := len(p.mapHashes) == len(p.mapMiniBlocks)
	p.mutPendingMb.Unlock()
	if receivedAll {
		p.chReceivedAll <- true
	}
}

func (p *pendingMiniBlocks) getMiniBlockFromPoolOrStorage(hash []byte) (*block.MiniBlock, bool) {
	miniBlock, ok := p.getMiniBlockFromPool(hash)
	if ok {
		return miniBlock, true
	}

	mbData, err := GetDataFromStorage(hash, p.storage)
	if err != nil {
		return nil, false
	}

	mb := &block.MiniBlock{
		TxHashes: make([][]byte, 0),
	}
	err = p.marshalizer.Unmarshal(mb, mbData)
	if err != nil {
		return nil, false
	}

	return mb, true
}

func (p *pendingMiniBlocks) getMiniBlockFromPool(hash []byte) (*block.MiniBlock, bool) {
	val, ok := p.pool.Peek(hash)
	if !ok {
		return nil, false
	}

	miniBlock, ok := val.(*block.MiniBlock)
	if !ok {
		return nil, false
	}

	return miniBlock, true
}

// GetMiniBlocks returns the synced miniblocks
func (p *pendingMiniBlocks) GetMiniBlocks() (map[string]*block.MiniBlock, error) {
	p.mutPendingMb.Lock()
	defer p.mutPendingMb.Unlock()
	if !p.syncedAll {
		return nil, update.ErrNotSynced
	}

	return p.mapMiniBlocks, nil
}

// ClearFields will clear all the maps
func (p *pendingMiniBlocks) ClearFields() {
	p.mutPendingMb.Lock()
	p.mapHashes = make(map[string]struct{})
	p.mapMiniBlocks = make(map[string]*block.MiniBlock)
	p.mutPendingMb.Unlock()
}

// IsInterfaceNil returns nil if underlying object is nil
func (p *pendingMiniBlocks) IsInterfaceNil() bool {
	return p == nil
}
