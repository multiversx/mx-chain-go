package sync

import (
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

type headersToSync struct {
	mutMeta          sync.Mutex
	metaBlockToSync  *block.MetaBlock
	chReceivedAll    chan bool
	metaBlockStorage update.HistoryStorer
	metaBlockPool    storage.Cacher
	epochHandler     update.EpochStartVerifier
	marshalizer      marshal.Marshalizer
	stopSyncing      bool
	epochToSync      uint32
	requestHandler   process.RequestHandler
}

// ArgsNewHeadersSyncHandler defines the arguments needed for the new header syncer
type ArgsNewHeadersSyncHandler struct {
	Storage        storage.Storer
	Cache          storage.Cacher
	Marshalizer    marshal.Marshalizer
	EpochHandler   update.EpochStartVerifier
	RequestHandler process.RequestHandler
}

// NewHeadersSyncHandler creates a new header syncer
func NewHeadersSyncHandler(args ArgsNewHeadersSyncHandler) (*headersToSync, error) {
	if check.IfNil(args.Storage) {
		return nil, dataRetriever.ErrNilHeadersStorage
	}
	if check.IfNil(args.Cache) {
		return nil, dataRetriever.ErrNilCacher
	}
	if check.IfNil(args.EpochHandler) {
		return nil, dataRetriever.ErrNilEpochHandler
	}
	if check.IfNil(args.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(args.RequestHandler) {
		return nil, process.ErrNilRequestHandler
	}

	headers := &headersToSync{
		mutMeta:          sync.Mutex{},
		metaBlockToSync:  &block.MetaBlock{},
		chReceivedAll:    make(chan bool),
		metaBlockStorage: args.Storage,
		metaBlockPool:    args.Cache,
		epochHandler:     args.EpochHandler,
		stopSyncing:      true,
		requestHandler:   args.RequestHandler,
		marshalizer:      args.Marshalizer,
	}

	headers.metaBlockPool.RegisterHandler(headers.receivedMetaBlock)

	return headers, nil
}

func (h *headersToSync) receivedMetaBlock(hash []byte) {
	h.mutMeta.Lock()
	if h.stopSyncing {
		h.mutMeta.Unlock()
		return
	}

	val, ok := h.metaBlockPool.Peek(hash)
	if !ok {
		h.mutMeta.Unlock()
		return
	}

	meta, ok := val.(*block.MetaBlock)
	if !ok {
		h.mutMeta.Unlock()
		return
	}

	isWrongEpoch := meta.Epoch > h.epochToSync || meta.Epoch < h.epochToSync-1
	if isWrongEpoch {
		h.mutMeta.Unlock()
		return
	}

	h.epochHandler.ReceivedHeader(meta)
	if !h.epochHandler.IsEpochStart() {
		h.mutMeta.Unlock()
		return
	}

	epochStartMetaHash := h.epochHandler.EpochStartMetaHdrHash()
	metaData, ok := h.metaBlockPool.Peek(epochStartMetaHash)
	if !ok {
		h.mutMeta.Unlock()
		return
	}

	metaBlock, ok := metaData.(*block.MetaBlock)
	if !ok {
		h.mutMeta.Unlock()
		return
	}

	h.metaBlockToSync = metaBlock
	h.stopSyncing = true
	h.mutMeta.Unlock()

	h.chReceivedAll <- true
}

// SyncEpochStartMetaHeader syncs and validates an epoch start metaheader
func (h *headersToSync) SyncEpochStartMetaHeader(epoch uint32, waitTime time.Duration) (*block.MetaBlock, error) {
	meta := &block.MetaBlock{}
	h.epochToSync = epoch
	epochStartId := core.EpochStartIdentifier(epoch)
	epochStartData, err := GetDataFromStorage([]byte(epochStartId), h.metaBlockStorage, epoch)
	if err != nil {
		_ = process.EmptyChannel(h.chReceivedAll)

		h.mutMeta.Lock()
		h.stopSyncing = false
		h.mutMeta.Unlock()

		h.requestHandler.RequestStartOfEpochMetaBlock(epoch)

		err = WaitFor(h.chReceivedAll, waitTime)
		log.Warn("timeOut for requesting epoch metaHdr")
		if err != nil {
			return nil, err
		}

		h.mutMeta.Lock()
		meta = h.metaBlockToSync
		h.mutMeta.Unlock()

		return meta, nil
	}

	err = h.marshalizer.Unmarshal(meta, epochStartData)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

// GetMetaBlock returns the synced metablock
func (h *headersToSync) GetMetaBlock() (*block.MetaBlock, error) {
	h.mutMeta.Lock()
	meta := h.metaBlockToSync
	h.mutMeta.Unlock()

	if meta.IsStartOfEpochBlock() {
		return meta, nil
	}

	return nil, update.ErrNotSynced
}

// IsInterfaceNil returns true if underlying object is nil
func (h *headersToSync) IsInterfaceNil() bool {
	return h == nil
}
