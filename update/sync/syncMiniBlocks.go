package sync

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
)

type pendingMiniBlocks struct {
	mutPendingMb   sync.Mutex
	mapMiniBlocks  map[string]*block.MiniBlock
	mapHashes      map[string]struct{}
	pool           storage.Cacher
	storage        update.HistoryStorer
	chReceivedAll  chan bool
	marshalizer    marshal.Marshalizer
	stopSyncing    bool
	epochToSync    uint32
	syncedAll      bool
	requestHandler process.RequestHandler
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
		return nil, dataRetriever.ErrNilCacher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(args.RequestHandler) {
		return nil, process.ErrNilRequestHandler
	}

	p := &pendingMiniBlocks{
		mutPendingMb:   sync.Mutex{},
		mapMiniBlocks:  make(map[string]*block.MiniBlock),
		mapHashes:      make(map[string]struct{}),
		pool:           args.Cache,
		storage:        args.Storage,
		chReceivedAll:  make(chan bool),
		requestHandler: args.RequestHandler,
		stopSyncing:    true,
		syncedAll:      false,
	}

	p.pool.RegisterHandler(p.receivedMiniBlock)

	return p, nil
}

// SyncPendingMiniBlocksFromMeta syncs the pending miniblocks from an epoch start metaBlock
func (p *pendingMiniBlocks) SyncPendingMiniBlocksFromMeta(
	epochStart *block.MetaBlock,
	unFinished map[string]*block.MetaBlock,
	waitTime time.Duration,
) error {
	if !epochStart.IsStartOfEpochBlock() {
		return update.ErrNotEpochStartBlock
	}
	if unFinished == nil {
		return update.ErrWrongUnfinishedMetaHdrsMap
	}

	listPendingMiniBlocks := make([]block.ShardMiniBlockHeader, 0)
	nonceToHashMap := p.createNonceToHashMap(unFinished)

	for _, shardData := range epochStart.EpochStart.LastFinalizedHeaders {
		computedPending, err := p.computePendingMiniBlocksFromUnFinished(shardData, unFinished, nonceToHashMap, epochStart.GetNonce())
		if err != nil {
			return err
		}

		listPendingMiniBlocks = append(listPendingMiniBlocks, computedPending...)
	}

	_ = process.EmptyChannel(p.chReceivedAll)

	requestedMBs := 0
	p.mutPendingMb.Lock()
	p.stopSyncing = false
	for _, mbHeader := range listPendingMiniBlocks {
		p.mapHashes[string(mbHeader.Hash)] = struct{}{}
		miniBlock, ok := p.getMiniBlockFromPoolOrStorage(mbHeader.Hash)
		if ok {
			p.mapMiniBlocks[string(mbHeader.Hash)] = miniBlock
			continue
		}

		requestedMBs++
		p.requestHandler.RequestMiniBlock(mbHeader.SenderShardID, mbHeader.Hash)
	}
	p.mutPendingMb.Unlock()

	var err error
	defer func() {
		p.mutPendingMb.Lock()
		p.stopSyncing = true
		if err == nil {
			p.syncedAll = true
		}
		p.mutPendingMb.Unlock()
	}()

	if requestedMBs > 0 {
		err = WaitFor(p.chReceivedAll, waitTime)
		if err != nil {
			return err
		}
	}

	return nil
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
) ([]block.ShardMiniBlockHeader, error) {
	pending := make([]block.ShardMiniBlockHeader, 0)
	pending = append(pending, shardData.PendingMiniBlockHeaders...)

	firstPendingMeta, ok := unFinished[string(shardData.FirstPendingMetaBlock)]
	if !ok {
		return nil, update.ErrWrongUnfinishedMetaHdrsMap
	}

	firstUnFinishedNonce := firstPendingMeta.GetNonce()
	for nonce := firstUnFinishedNonce + 1; nonce <= epochStartNonce; nonce++ {
		metaHash, ok := nonceToHash[nonce]
		if !ok {
			return nil, update.ErrWrongUnfinishedMetaHdrsMap
		}

		log.Debug("unFinished access")
		meta, ok := unFinished[metaHash]
		if !ok {
			return nil, update.ErrWrongUnfinishedMetaHdrsMap
		}

		pendingFromCurrentMeta := getAllMiniBlocksWithDst(meta, shardData.ShardId)
		pending = append(pending, pendingFromCurrentMeta...)
	}

	return pending, nil
}

func getAllMiniBlocksWithDst(m *block.MetaBlock, destId uint32) []block.ShardMiniBlockHeader {
	mbHdrs := make([]block.ShardMiniBlockHeader, 0)
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
			shardMiniBlockHdr := block.ShardMiniBlockHeader{
				Hash:            val.Hash,
				ReceiverShardID: val.ReceiverShardID,
				SenderShardID:   val.SenderShardID,
				TxCount:         val.TxCount,
			}
			mbHdrs = append(mbHdrs, shardMiniBlockHdr)
		}
	}

	return mbHdrs
}

// receivedMiniBlock is a callback function when a new miniblock was received
// it will further ask for missing transactions
func (p *pendingMiniBlocks) receivedMiniBlock(miniBlockHash []byte) {
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

	miniBlock, ok := p.getMiniBlockFromPool(miniBlockHash)
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

	mbData, err := GetDataFromStorage(hash, p.storage, p.epochToSync)
	if err != nil {
		return nil, false
	}

	mb := &block.MiniBlock{}
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

// IsInterfaceNil returns nil if underlying object is nil
func (p *pendingMiniBlocks) IsInterfaceNil() bool {
	return p == nil
}
