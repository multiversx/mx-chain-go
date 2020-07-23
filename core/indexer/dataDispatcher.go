package indexer

import (
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


	return &dataDispatcher{
		elasticIndexer: ei,

		options:     arguments.Options,
		marshalizer: arguments.Marshalizer,
	}, nil
}

// SaveBlock saves the block info in the queue to be sent to elastic
func (d *dataDispatcher) SaveBlock(
	bodyHandler data.BodyHandler,
	headerHandler data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	signersIndexes []uint64,
	notarizedHeadersHashes []string,
) {
	body, ok := bodyHandler.(*block.Body)
	if !ok {
		log.Debug("indexer", "error", ErrBodyTypeAssertion.Error())
		return
	}

	if check.IfNil(headerHandler) {
		log.Debug("indexer: no header", "error", ErrNoHeader.Error())
		return
	}

	txsSizeInBytes := computeSizeOfTxs(d.marshalizer, txPool)
	d.elasticIndexer.SaveHeader(headerHandler, signersIndexes, body, notarizedHeadersHashes, txsSizeInBytes)

	if len(body.MiniBlocks) == 0 {
		return
	}

	mbsInDb := d.elasticIndexer.SaveMiniblocks(headerHandler, body)
	if d.options.TxIndexingEnabled {
		d.elasticIndexer.SaveTransactions(body, headerHandler, txPool, headerHandler.GetShardID(), mbsInDb)
	}
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

// IsInterfaceNil returns true if there is no value under the interface
func (d *dataDispatcher) IsInterfaceNil() bool {
	return d == nil
}
