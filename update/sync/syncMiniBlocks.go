package sync

import (
	"context"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
)

var _ update.EpochStartPendingMiniBlocksSyncHandler = (*miniBlocksSyncer)(nil)

type miniBlocksSyncer struct {
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

// ArgsNewMiniBlocksSyncer defines the arguments needed for the miniblocks syncer
type ArgsNewMiniBlocksSyncer struct {
	Storage        storage.Storer
	Cache          storage.Cacher
	Marshalizer    marshal.Marshalizer
	RequestHandler process.RequestHandler
}

// NewMiniBlocksSyncer creates a syncer for all required miniblocks
func NewMiniBlocksSyncer(args ArgsNewMiniBlocksSyncer) (*miniBlocksSyncer, error) {
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

	p := &miniBlocksSyncer{
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
func (syncer *miniBlocksSyncer) SyncPendingMiniBlocksFromMeta(epochStart data.MetaHeaderHandler, unFinished map[string]data.MetaHeaderHandler, ctx context.Context) error {
	if !epochStart.IsStartOfEpochBlock() && epochStart.GetNonce() > 0 {
		return update.ErrNotEpochStartBlock
	}
	if unFinished == nil {
		return update.ErrWrongUnFinishedMetaHdrsMap
	}

	for hash, meta := range unFinished {
		log.Debug("syncing miniblocks from unFinished meta", "hash", []byte(hash), "nonce", meta.GetNonce())
	}

	listPendingMiniBlocks, err := update.GetPendingMiniBlocks(epochStart, unFinished)
	if err != nil {
		return err
	}

	return syncer.syncMiniBlocks(listPendingMiniBlocks, ctx)
}

// SyncMiniBlocks will sync the miniblocks for the given miniblocks header handlers
func (syncer *miniBlocksSyncer) SyncMiniBlocks(miniBlockHeaders []data.MiniBlockHeaderHandler, ctx context.Context) error {
	return syncer.syncMiniBlocks(miniBlockHeaders, ctx)
}

func (syncer *miniBlocksSyncer) syncMiniBlocks(listMiniBlocks []data.MiniBlockHeaderHandler, ctx context.Context) error {
	_ = core.EmptyChannel(syncer.chReceivedAll)

	mapMiniBlocksToRequest := make(map[string]data.MiniBlockHeaderHandler)
	for _, mbHeader := range listMiniBlocks {
		mapMiniBlocksToRequest[string(mbHeader.GetHash())] = mbHeader
	}

	syncer.mutPendingMb.Lock()
	syncer.stopSyncing = false
	syncer.syncedAll = false
	syncer.mutPendingMb.Unlock()

	for {
		requestedMBs := 0
		syncer.mutPendingMb.Lock()
		syncer.stopSyncing = false
		for hash, miniBlockHeader := range mapMiniBlocksToRequest {
			if _, ok := syncer.mapMiniBlocks[hash]; ok {
				delete(mapMiniBlocksToRequest, hash)
				continue
			}

			syncer.mapHashes[hash] = struct{}{}
			miniBlock, ok := syncer.getMiniBlockFromPoolOrStorage([]byte(hash))
			if ok {
				syncer.mapMiniBlocks[hash] = miniBlock
				delete(mapMiniBlocksToRequest, hash)
				continue
			}

			syncer.requestMiniBlockHeader(miniBlockHeader, []byte(hash))
			requestedMBs++
		}
		syncer.mutPendingMb.Unlock()

		if requestedMBs == 0 {
			syncer.mutPendingMb.Lock()
			syncer.stopSyncing = true
			syncer.syncedAll = true
			syncer.mutPendingMb.Unlock()
			return nil
		}

		select {
		case <-syncer.chReceivedAll:
			syncer.mutPendingMb.Lock()
			syncer.stopSyncing = true
			syncer.syncedAll = true
			syncer.mutPendingMb.Unlock()
			return nil
		case <-time.After(syncer.waitTimeBetweenRequests):
			continue
		case <-ctx.Done():
			syncer.mutPendingMb.Lock()
			syncer.stopSyncing = true
			syncer.mutPendingMb.Unlock()
			return update.ErrTimeIsOut
		}
	}
}

func (syncer *miniBlocksSyncer) requestMiniBlockHeader(miniBlockHeader data.MiniBlockHeaderHandler, hash []byte) {
	log.Debug("requesting miniblock header from network",
		"hash", hash,
		"sender shard ID", miniBlockHeader.GetSenderShardID(),
		"receiver shard ID", miniBlockHeader.GetReceiverShardID())

	syncer.requestHandler.RequestMiniBlock(miniBlockHeader.GetSenderShardID(), hash)
	if miniBlockHeader.GetSenderShardID() != miniBlockHeader.GetReceiverShardID() {
		// request from receiver shard also, to increase the chance of getting the miniblock
		syncer.requestHandler.RequestMiniBlock(miniBlockHeader.GetReceiverShardID(), hash)
	}
}

// receivedMiniBlock is a callback function when a new miniblock was received
// it will further ask for missing transactions
func (syncer *miniBlocksSyncer) receivedMiniBlock(miniBlockHash []byte, val interface{}) {
	syncer.mutPendingMb.Lock()
	if syncer.stopSyncing {
		syncer.mutPendingMb.Unlock()
		return
	}

	if _, ok := syncer.mapHashes[string(miniBlockHash)]; !ok {
		syncer.mutPendingMb.Unlock()
		return
	}

	if _, ok := syncer.mapMiniBlocks[string(miniBlockHash)]; ok {
		syncer.mutPendingMb.Unlock()
		return
	}

	miniBlock, ok := val.(*block.MiniBlock)
	if !ok {
		syncer.mutPendingMb.Unlock()
		return
	}

	syncer.mapMiniBlocks[string(miniBlockHash)] = miniBlock
	receivedAll := len(syncer.mapHashes) == len(syncer.mapMiniBlocks)
	syncer.mutPendingMb.Unlock()
	if receivedAll {
		syncer.chReceivedAll <- true
	}
}

func (syncer *miniBlocksSyncer) getMiniBlockFromPoolOrStorage(hash []byte) (*block.MiniBlock, bool) {
	miniBlock, ok := syncer.getMiniBlockFromPool(hash)
	if ok {
		return miniBlock, true
	}

	mbData, err := GetDataFromStorage(hash, syncer.storage)
	if err != nil {
		return nil, false
	}

	mb := &block.MiniBlock{
		TxHashes: make([][]byte, 0),
	}
	err = syncer.marshalizer.Unmarshal(mb, mbData)
	if err != nil {
		return nil, false
	}

	return mb, true
}

func (syncer *miniBlocksSyncer) getMiniBlockFromPool(hash []byte) (*block.MiniBlock, bool) {
	val, ok := syncer.pool.Peek(hash)
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
func (syncer *miniBlocksSyncer) GetMiniBlocks() (map[string]*block.MiniBlock, error) {
	syncer.mutPendingMb.Lock()
	defer syncer.mutPendingMb.Unlock()
	if !syncer.syncedAll {
		return nil, update.ErrNotSynced
	}

	return syncer.mapMiniBlocks, nil
}

// ClearFields will clear all the maps
func (syncer *miniBlocksSyncer) ClearFields() {
	syncer.mutPendingMb.Lock()
	syncer.mapHashes = make(map[string]struct{})
	syncer.mapMiniBlocks = make(map[string]*block.MiniBlock)
	syncer.mutPendingMb.Unlock()
}

// IsInterfaceNil returns nil if underlying object is nil
func (syncer *miniBlocksSyncer) IsInterfaceNil() bool {
	return syncer == nil
}
