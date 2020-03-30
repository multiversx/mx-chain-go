package sync

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
)

type syncHeadersByHash struct {
	mutMissingHdrs sync.Mutex
	mapHeaders     map[string]data.HeaderHandler
	mapHashes      map[string]struct{}
	pool           dataRetriever.HeadersPool
	storage        update.HistoryStorer
	chReceivedAll  chan bool
	marshalizer    marshal.Marshalizer
	stopSyncing    bool
	epochToSync    uint32
	syncedAll      bool
	requestHandler process.RequestHandler
}

// ArgsNewMissingHeadersByHashSyncer defines the arguments needed for the sycner
type ArgsNewMissingHeadersByHashSyncer struct {
	Storage        storage.Storer
	Cache          dataRetriever.HeadersPool
	Marshalizer    marshal.Marshalizer
	RequestHandler process.RequestHandler
}

// NewMissingheadersByHashSyncer creates a syncer for all missing headers
func NewMissingheadersByHashSyncer(args ArgsNewMissingHeadersByHashSyncer) (*syncHeadersByHash, error) {
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

	p := &syncHeadersByHash{
		mutMissingHdrs: sync.Mutex{},
		mapHeaders:     make(map[string]data.HeaderHandler),
		mapHashes:      make(map[string]struct{}),
		pool:           args.Cache,
		storage:        args.Storage,
		chReceivedAll:  make(chan bool),
		requestHandler: args.RequestHandler,
		stopSyncing:    true,
		syncedAll:      false,
		marshalizer:    args.Marshalizer,
	}

	p.pool.RegisterHandler(p.receivedHeader)

	return p, nil
}

// SyncMissingHeadersByHash syncs the missing headers
func (m *syncHeadersByHash) SyncMissingHeadersByHash(
	shardIDs []uint32,
	headersHashes [][]byte,
	waitTime time.Duration,
) error {
	_ = core.EmptyChannel(m.chReceivedAll)

	requestedMBs := 0
	m.mutMissingHdrs.Lock()
	m.stopSyncing = false
	for index, hash := range headersHashes {
		m.mapHashes[string(hash)] = struct{}{}
		header, ok := m.getHeaderFromPoolOrStorage(hash)
		if ok {
			m.mapHeaders[string(hash)] = header
			continue
		}

		requestedMBs++
		if shardIDs[index] == core.MetachainShardId {
			m.requestHandler.RequestMetaHeader(hash)
			continue
		}

		m.requestHandler.RequestShardHeader(shardIDs[index], hash)
	}
	m.mutMissingHdrs.Unlock()

	var err error
	if requestedMBs > 0 {
		err = WaitFor(m.chReceivedAll, waitTime)
	}

	m.mutMissingHdrs.Lock()
	m.stopSyncing = true
	if err == nil {
		m.syncedAll = true
	}
	m.mutMissingHdrs.Unlock()

	return err
}

// receivedHeader is a callback function when a new header was received
// it will further ask for missing transactions
func (m *syncHeadersByHash) receivedHeader(hdrHandler data.HeaderHandler, hdrHash []byte) {
	m.mutMissingHdrs.Lock()
	if m.stopSyncing {
		m.mutMissingHdrs.Unlock()
		return
	}

	if _, ok := m.mapHashes[string(hdrHash)]; !ok {
		m.mutMissingHdrs.Unlock()
		return
	}

	if _, ok := m.mapHeaders[string(hdrHash)]; ok {
		m.mutMissingHdrs.Unlock()
		return
	}

	m.mapHeaders[string(hdrHash)] = hdrHandler
	receivedAll := len(m.mapHashes) == len(m.mapHeaders)
	m.mutMissingHdrs.Unlock()
	if receivedAll {
		m.chReceivedAll <- true
	}
}

func (m *syncHeadersByHash) getHeaderFromPoolOrStorage(hash []byte) (data.HeaderHandler, bool) {
	header, ok := m.getHeaderFromPool(hash)
	if ok {
		return header, true
	}

	hdrData, err := GetDataFromStorage(hash, m.storage, m.epochToSync)
	if err != nil {
		return nil, false
	}

	var hdr block.Header
	err = m.marshalizer.Unmarshal(hdr, hdrData)
	if err != nil {
		return nil, false
	}

	return &hdr, true
}

func (m *syncHeadersByHash) getHeaderFromPool(hash []byte) (data.HeaderHandler, bool) {
	val, err := m.pool.GetHeaderByHash(hash)
	if err != nil {
		return nil, false
	}

	return val, true
}

// GetHeaders returns the synced headers
func (m *syncHeadersByHash) GetHeaders() (map[string]data.HeaderHandler, error) {
	m.mutMissingHdrs.Lock()
	defer m.mutMissingHdrs.Unlock()
	if !m.syncedAll {
		return nil, update.ErrNotSynced
	}

	return m.mapHeaders, nil
}

// ClearFields will clear all the maps
func (m *syncHeadersByHash) ClearFields() {
	m.mutMissingHdrs.Lock()
	m.mapHashes = make(map[string]struct{})
	m.mapHeaders = make(map[string]data.HeaderHandler)
	m.mutMissingHdrs.Unlock()
}

// IsInterfaceNil returns nil if underlying object is nil
func (m *syncHeadersByHash) IsInterfaceNil() bool {
	return m == nil
}
