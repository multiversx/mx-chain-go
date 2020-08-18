package indexer

import (
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

var log = logger.GetOrCreate("core/indexer")

// Options structure holds the indexer's configuration options
type Options struct {
	TxIndexingEnabled bool
}

type dataDispatcher struct {
	elasticIndexer ElasticIndexer
	workQueue      *workQueue

	options     *Options
	marshalizer marshal.Marshalizer
}

// NewDataDispatcher creates a new dataDispatcher instance, capable of selecting the correct es that will
//  handle saving different types
func NewDataDispatcher(arguments ElasticIndexerArgs) (Indexer, error) {
	ei, err := NewElasticIndexer(arguments)
	if err != nil {
		return nil, err
	}

	wq, err := NewWorkQueue()
	if err != nil {
		return nil, err
	}

	dd := &dataDispatcher{
		elasticIndexer: ei,
		workQueue:      wq,

		options:     arguments.Options,
		marshalizer: arguments.Marshalizer,
	}

	go dd.startWorker()

	return dd, nil
}

// SaveBlock saves the block info in the queue to be sent to elastic
func (d *dataDispatcher) SaveBlock(
	bodyHandler data.BodyHandler,
	headerHandler data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	signersIndexes []uint64,
	notarizedHeadersHashes []string,
) {
	wi, err := NewWorkItem(WorkTypeSaveBlock, &saveBlockData{
		bodyHandler:            bodyHandler,
		headerHandler:          headerHandler,
		txPool:                 txPool,
		signersIndexes:         signersIndexes,
		notarizedHeadersHashes: notarizedHeadersHashes,
	})
	if err != nil {
		log.Error("dataDispatcher.SaveBlock", "error creating work item", err.Error())
		return
	}

	d.workQueue.Add(wi)
}

// RevertIndexedBlock -
func (d *dataDispatcher) RevertIndexedBlock(header data.HeaderHandler) {
	
}

// SaveRoundsInfos will save data about a slice of rounds on elasticsearch
func (d *dataDispatcher) SaveRoundsInfos(roundsInfos []RoundInfo) {

}

// SaveValidatorsRating will send all validators rating info to elasticsearch
func (d *dataDispatcher) SaveValidatorsRating(indexID string, validatorsRatingInfo []ValidatorRatingInfo) {

}

// SaveValidatorsPubKeys will send all validators public keys to elasticsearch
func (d *dataDispatcher) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) {

}

// UpdateTPS updates the tps and statistics into elasticsearch index
func (d *dataDispatcher) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {

}

// SetTxLogsProcessor will set tx logs processor
func (d *dataDispatcher) SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase) {

}

// IsNilIndexer will return a bool value that signals if the indexer's implementation is a NilIndexer
func (d *dataDispatcher) IsNilIndexer() bool {
	return false
}

func (d *dataDispatcher) saveBlock(item *workItem) {
	saveBlockParams, ok := item.Data.(*saveBlockData)
	if !ok {
		log.Warn("dataDispatcher.saveBlock", "removing item from queue", ErrInvalidWorkItemData.Error())
		d.workQueue.Done()
		return
	}

	if len(saveBlockParams.signersIndexes) == 0 {
		log.Warn("block has no signers, returning")
		d.workQueue.Done()
		return
	}

	body, ok := saveBlockParams.bodyHandler.(*block.Body)
	if !ok {
		log.Warn("dataDispatcher.saveBlock", "removing item from queue", ErrBodyTypeAssertion.Error())
		d.workQueue.Done()
		return
	}

	if check.IfNil(saveBlockParams.headerHandler) {
		log.Warn("dataDispatcher.saveBlock", "removing item from queue", ErrNoHeader.Error())
		d.workQueue.Done()
		return
	}

	txsSizeInBytes := computeSizeOfTxs(d.marshalizer, saveBlockParams.txPool)
	err := d.elasticIndexer.SaveHeader(saveBlockParams.headerHandler, saveBlockParams.signersIndexes, body, saveBlockParams.notarizedHeadersHashes, txsSizeInBytes)
	if err != nil && err == ErrBackOff {
		log.Warn("dataDispatcher.saveBlock", "could not index header, received back off:", err.Error())
		d.workQueue.GotBackOff()
		return
	}
	if err != nil {
		log.Warn("dataDispatcher.saveBlock", "removing item from queue", err.Error())
		d.workQueue.Done()
		return
	}

	if len(body.MiniBlocks) == 0 {
		d.workQueue.Done()
		return
	}

	mbsInDb := d.elasticIndexer.SaveMiniblocks(saveBlockParams.headerHandler, body)

	shardId := saveBlockParams.headerHandler.GetShardID()
	if d.options.TxIndexingEnabled {
		err = d.elasticIndexer.SaveTransactions(body, saveBlockParams.headerHandler, saveBlockParams.txPool, shardId, mbsInDb)
		if err != nil && err == ErrBackOff {
			log.Warn("dataDispatcher.saveBlock", "could not index header, received back off:", err.Error())
			d.workQueue.GotBackOff()
			return
		}
	}

	d.workQueue.Done()
}

func (d *dataDispatcher) startWorker() {
	for {
		time.Sleep(d.workQueue.GetCycleTime())
		d.doWork()
	}
}

func (d *dataDispatcher) doWork() {
	if d.workQueue.Length() == 0 {
		return
	}

	wi := d.workQueue.Next()
	if wi == nil {
		return
	}

	switch wi.Type {
	case WorkTypeSaveBlock:
		d.saveBlock(wi)
		break
	default:
		log.Error("dataDispatcher.doWork invalid work type received")
		break
	}
}


// IsInterfaceNil returns true if there is no value under the interface
func (d *dataDispatcher) IsInterfaceNil() bool {
	return d == nil
}
