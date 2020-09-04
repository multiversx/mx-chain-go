package indexer

import (
	"context"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/elastic/go-elasticsearch/v7"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/statistics"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

var log = logger.GetOrCreate("core/indexer")

// Options structure holds the indexer's configuration options
type Options struct {
	TxIndexingEnabled bool
}

type dataDispatcher struct {
	elasticIndexer ElasticIndexer
	workQueue      QueueHandler

	options     *Options
	marshalizer marshal.Marshalizer
	cancelFunc  func()
}

// NewDataDispatcher creates a new dataDispatcher instance, capable of selecting the correct es that will
//  handle saving different types
func NewDataDispatcher(arguments DataIndexerArgs) (*dataDispatcher, error) {
	// create elastic search client
	databaseClient, err := newElasticClient(elasticsearch.Config{
		Addresses: []string{arguments.Url},
		Username:  arguments.UserName,
		Password:  arguments.Password,
	})
	if err != nil {
		return nil, err
	}

	esIndexerArgs := ElasticIndexerArgs{
		IndexTemplates:           arguments.IndexTemplates,
		IndexPolicies:            arguments.IndexPolicies,
		Marshalizer:              arguments.Marshalizer,
		Hasher:                   arguments.Hasher,
		AddressPubkeyConverter:   arguments.AddressPubkeyConverter,
		ValidatorPubkeyConverter: arguments.ValidatorPubkeyConverter,
		Options:                  arguments.Options,
		DBClient:                 databaseClient,
	}

	ei, err := NewElasticIndexer(esIndexerArgs)
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

	return dd, nil
}

// StartIndexData will start index data in database
func (d *dataDispatcher) StartIndexData() {
	var ctx context.Context
	ctx, d.cancelFunc = context.WithCancel(context.Background())

	go d.startWorker(ctx)
}

func (d *dataDispatcher) startWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("dispatcher's go routine is stopping...")
			return
		default:
			d.doWork()
		}
	}
}

// Close will close the endless running go routine
func (d *dataDispatcher) Close() error {
	if d.cancelFunc != nil {
		d.cancelFunc()
	}

	return nil
}

// Add will add a new item in queue
func (d *dataDispatcher) Add(item *workItem) {
	d.workQueue.Add(item)
}

// GetQueueLength will  return the length of the queue
func (d *dataDispatcher) GetQueueLength() int {
	return d.workQueue.Length()
}

// SetTxLogsProcessor will set tx logs processor
func (d *dataDispatcher) SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase) {
	d.elasticIndexer.SetTxLogsProcessor(txLogsProc)
}

func (d *dataDispatcher) doWork() {
	if d.workQueue.Length() == 0 {
		time.Sleep(d.workQueue.GetCycleTime())
		return
	}

	if d.workQueue.GetBackOffTime() != 0 {
		time.Sleep(d.workQueue.GetCycleTime())
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
	case WorkTypeUpdateTPSBenchmark:
		d.updateTPSBenchmark(wi)
		break
	case WorkTypeSaveValidatorsRating:
		d.saveValidatorsRating(wi)
		break
	case WorkTypeSaveValidatorsPubKeys:
		d.saveValidatorsPubKeys(wi)
		break
	case WorkTypeSaveRoundsInfo:
		d.saveRoundsInfo(wi)
		break
	default:
		log.Error("dataDispatcher.doWork invalid work type received")
		d.workQueue.Done()
		break
	}
}

func (d *dataDispatcher) saveBlock(item *workItem) {
	saveBlockParams, body, ok := getBlockParamsAndBody(item)
	if !ok {
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

	mbsInDb, err := d.elasticIndexer.SaveMiniblocks(saveBlockParams.headerHandler, body)
	if err != nil && err == ErrBackOff {
		log.Warn("dataDispatcher.saveBlock", "could not index miniblocks, received back off:", err.Error())
		d.workQueue.GotBackOff()
		return
	}

	shardID := saveBlockParams.headerHandler.GetShardID()
	if d.options.TxIndexingEnabled {
		err = d.elasticIndexer.SaveTransactions(body, saveBlockParams.headerHandler, saveBlockParams.txPool, shardID, mbsInDb)
		if err != nil && err == ErrBackOff {
			log.Warn("dataDispatcher.saveBlock", "could not index header, received back off:", err.Error())
			d.workQueue.GotBackOff()
			return
		}
	}

	d.workQueue.Done()
}

func (d *dataDispatcher) updateTPSBenchmark(item *workItem) {
	tpsBenchmark := item.Data.(*statistics.TPSBenchmark)

	err := d.elasticIndexer.SaveShardStatistics(*tpsBenchmark)
	if err != nil && err == ErrBackOff {
		log.Warn("dataDispatcher.updateTPSBenchmark", "could not index tps benchmark, received back off:", err.Error())
		d.workQueue.GotBackOff()
		return
	}
	if err != nil {
		log.Warn("dataDispatcher.updateTPSBenchmark", "removing item from queue", err.Error())
		d.workQueue.Done()
		return
	}

	d.workQueue.Done()
}

func (d *dataDispatcher) saveValidatorsRating(item *workItem) {
	validatorsRatingData, ok := item.Data.(*saveValidatorsRatingData)
	if !ok {
		log.Warn("dataDispatcher.saveValidatorsRating",
			"removing item from queue", ErrInvalidWorkItemData.Error())
		d.workQueue.Done()
		return
	}

	err := d.elasticIndexer.SaveValidatorsRating(validatorsRatingData.indexID, validatorsRatingData.infoRating)
	if err != nil && err == ErrBackOff {
		log.Warn("dataDispatcher.saveValidatorsRating",
			"could not index validators rating, received back off:", err.Error())
		d.workQueue.GotBackOff()
		return
	}
	if err != nil {
		log.Warn("dataDispatcher.saveValidatorsRating",
			"removing item from queue", err.Error())
		d.workQueue.Done()
		return
	}

	d.workQueue.Done()
}

func (d *dataDispatcher) saveValidatorsPubKeys(item *workItem) {
	validatorsPubKeysData, ok := item.Data.(*saveValidatorsPubKeysData)
	if !ok {
		log.Warn("dataDispatcher.saveValidatorsPubKeys",
			"removing item from queue", ErrInvalidWorkItemData.Error())
		d.workQueue.Done()
		return
	}

	for shardID, shardPubKeys := range validatorsPubKeysData.validatorsPubKeys {
		err := d.elasticIndexer.SaveShardValidatorsPubKeys(shardID, validatorsPubKeysData.epoch, shardPubKeys)
		if err != nil && err == ErrBackOff {
			log.Warn("dataDispatcher.saveValidatorsPubKeys",
				"could not index validators public keys, ",
				"for shard", shardID,
				"received back off:", err.Error())
			d.workQueue.GotBackOff()
			return
		}
		if err != nil {
			log.Warn("dataDispatcher.saveValidatorsPubKeys",
				"removing item from queue", err.Error())
			continue
		}
	}

	d.workQueue.Done()
}

func (d *dataDispatcher) saveRoundsInfo(item *workItem) {
	roundsInfo, ok := item.Data.(*[]RoundInfo)
	if !ok {
		log.Warn("dataDispatcher.saveRoundsInfo",
			"removing item from queue", ErrInvalidWorkItemData.Error())
		d.workQueue.Done()
		return
	}

	err := d.elasticIndexer.SaveRoundsInfos(*roundsInfo)
	if err != nil && err == ErrBackOff {
		log.Warn("dataDispatcher.saveRoundsInfo",
			"could not index rounds info, received back off:", err.Error())
		d.workQueue.GotBackOff()
		return
	}
	if err != nil {
		log.Warn("dataDispatcher.saveRoundsInfo",
			"removing item from queue", err.Error())
		d.workQueue.Done()
		return
	}

	d.workQueue.Done()
}
