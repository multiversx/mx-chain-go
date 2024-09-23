package sync

import (
	"context"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/update"
)

// ArgsNewReceiptsDataSyncer defines the arguments needed for the sycner
type ArgsNewReceiptsDataSyncer struct {
	Cache          storage.Cacher
	RequestHandler process.RequestHandler
}

type receiptsDataSyncer struct {
	mapHashes               map[string]struct{}
	mapReceiptsData         map[string][]byte
	requestHandler          process.RequestHandler
	pool                    storage.Cacher
	mutex                   sync.Mutex
	waitTimeBetweenRequests time.Duration
	chReceivedAll           chan bool
	stopSync                bool
	syncedAll               bool
}

// NewReceiptsDataSyncer will create a new instance of receipt data syncer
func NewReceiptsDataSyncer(args ArgsNewReceiptsDataSyncer) (*receiptsDataSyncer, error) {
	if check.IfNil(args.Cache) {
		return nil, update.ErrNilCacher
	}
	if check.IfNil(args.RequestHandler) {
		return nil, update.ErrNilRequestHandler
	}

	syncer := &receiptsDataSyncer{
		requestHandler:          args.RequestHandler,
		pool:                    args.Cache,
		mutex:                   sync.Mutex{},
		waitTimeBetweenRequests: args.RequestHandler.RequestInterval(),
		chReceivedAll:           make(chan bool),
		mapHashes:               make(map[string]struct{}),
		stopSync:                true,
		syncedAll:               true,
	}

	syncer.pool.RegisterHandler(syncer.receivedReceiptsData, core.UniqueIdentifier())

	return syncer, nil
}

func (rds *receiptsDataSyncer) receivedReceiptsData(receiptDataHash []byte, val interface{}) {
	rds.mutex.Lock()
	if rds.stopSync {
		rds.mutex.Unlock()
		return
	}

	if _, ok := rds.mapHashes[string(receiptDataHash)]; !ok {
		rds.mutex.Unlock()
		return
	}

	if _, ok := rds.mapReceiptsData[string(receiptDataHash)]; ok {
		rds.mutex.Unlock()
		return
	}

	receiptDataBytes, ok := val.([]byte)
	if !ok {
		rds.mutex.Unlock()
		return
	}

	rds.mapReceiptsData[string(receiptDataHash)] = receiptDataBytes
	receivedAll := len(rds.mapHashes) == len(rds.mapReceiptsData)
	rds.mutex.Unlock()
	if receivedAll {
		rds.chReceivedAll <- true
	}
}

// SyncReceiptsDataFor syncs receipts data the provided hashes
func (rds *receiptsDataSyncer) SyncReceiptsDataFor(hashes [][]byte, ctx context.Context) error {
	_ = core.EmptyChannel(rds.chReceivedAll)

	for {
		rds.mutex.Lock()
		rds.requestReceiptsTrieNodes(hashes)
		rds.mutex.Unlock()

		select {
		case <-rds.chReceivedAll:
			rds.mutex.Lock()
			rds.stopSync = true
			rds.syncedAll = true
			rds.mutex.Unlock()
			return nil
		case <-time.After(rds.waitTimeBetweenRequests):
			rds.mutex.Lock()
			log.Debug("receiptsDataSyncer.SyncReceiptsDataFor", "num nodes needed", len(hashes), "num nodes got", len(rds.mapHashes))
			rds.mutex.Unlock()
			continue
		case <-ctx.Done():
			rds.mutex.Lock()
			rds.stopSync = true
			rds.mutex.Unlock()
			return update.ErrTimeIsOut
		}
	}
}

func (rds *receiptsDataSyncer) requestReceiptsTrieNodes(hashes [][]byte) {
	hashesToRequest := make([][]byte, 0)
	for _, hash := range hashes {
		_, ok := rds.mapReceiptsData[string(hash)]
		if ok {
			continue
		}

		rds.mapHashes[string(hash)] = struct{}{}
		receiptsDataBytes, ok := rds.getReceiptDataFromPool(hash)
		if ok {
			rds.mapReceiptsData[string(hash)] = receiptsDataBytes
			continue
		}

		hashesToRequest = append(hashesToRequest, hash)
	}

	rds.requestHandler.RequestReceiptsTrieNodes(hashesToRequest)
}

func (rds *receiptsDataSyncer) getReceiptDataFromPool(hash []byte) ([]byte, bool) {
	res, ok := rds.pool.Peek(hash)
	if !ok {
		return nil, false
	}

	receiptDataBytes, ok := res.([]byte)
	if !ok {
		return nil, false
	}

	return receiptDataBytes, true
}

// GetReceiptsData returns the synced receipts data
func (rds *receiptsDataSyncer) GetReceiptsData() (map[string][]byte, error) {
	rds.mutex.Lock()
	defer rds.mutex.Unlock()

	if !rds.syncedAll {
		return nil, update.ErrNotSynced
	}

	return rds.mapReceiptsData, nil
}

// ClearFields will clear all the maps
func (rds *receiptsDataSyncer) ClearFields() {
	rds.mutex.Lock()
	rds.mapHashes = make(map[string]struct{})
	rds.mapReceiptsData = make(map[string][]byte)
	rds.mutex.Unlock()
}

// IsInterfaceNil returns true if underlying object is nil
func (rds *receiptsDataSyncer) IsInterfaceNil() bool {
	return rds == nil
}
