package elastic

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/outport/drivers/elastic/workItems"
	"github.com/ElrondNetwork/elrond-go/outport/types"
)

type dataIndexer struct {
	isNilIndexer     bool
	dispatcher       DispatcherHandler
	elasticProcessor ElasticProcessor
	options          *Options
	marshalizer      marshal.Marshalizer
}

// NewDataIndexer will create a new data indexer
func NewDataIndexer(arguments ArgDataIndexer) (*dataIndexer, error) {
	err := checkIndexerArgs(arguments)
	if err != nil {
		return nil, err
	}

	dataIndexerObj := &dataIndexer{
		isNilIndexer:     false,
		dispatcher:       arguments.DataDispatcher,
		elasticProcessor: arguments.ElasticProcessor,
		marshalizer:      arguments.Marshalizer,
		options:          arguments.Options,
	}

	return dataIndexerObj, nil
}

func checkIndexerArgs(arguments ArgDataIndexer) error {
	if check.IfNil(arguments.DataDispatcher) {
		return ErrNilDataDispatcher
	}
	if check.IfNil(arguments.ElasticProcessor) {
		return ErrNilElasticProcessor
	}
	if check.IfNil(arguments.Marshalizer) {
		return core.ErrNilMarshalizer
	}
	if check.IfNil(arguments.ShardCoordinator) {
		return ErrNilShardCoordinator
	}

	return nil
}

// SaveBlock saves the block info in the queue to be sent to elastic
func (di *dataIndexer) SaveBlock(args types.ArgsSaveBlocks) {
	wi := workItems.NewItemBlock(
		di.elasticProcessor,
		di.marshalizer,
		args.Body,
		args.Header,
		args.TxsFromPool,
		args.SignersIndexes,
		args.NotarizedHeadersHashes,
		args.HeaderHash,
	)
	di.dispatcher.Add(wi)
}

// Close will stop goroutine that index data in database
func (di *dataIndexer) Close() error {
	return di.dispatcher.Close()
}

// RevertBlock will remove from database block and miniblocks
func (di *dataIndexer) RevertBlock(header data.HeaderHandler, body data.BodyHandler) {
	wi := workItems.NewItemRemoveBlock(
		di.elasticProcessor,
		body,
		header,
	)
	di.dispatcher.Add(wi)
}

// SaveRoundsInfo will save data about a slice of rounds in elasticsearch
func (di *dataIndexer) SaveRoundsInfo(roundsInfo []types.RoundInfo) {
	wi := workItems.NewItemRounds(di.elasticProcessor, roundsInfo)
	di.dispatcher.Add(wi)
}

// SaveValidatorsRating will save all validators rating info to elasticsearch
func (di *dataIndexer) SaveValidatorsRating(indexID string, validatorsRatingInfo []types.ValidatorRatingInfo) {
	wi := workItems.NewItemRating(
		di.elasticProcessor,
		indexID,
		validatorsRatingInfo,
	)
	di.dispatcher.Add(wi)
}

// SaveValidatorsPubKeys will save all validators public keys to elasticsearch
func (di *dataIndexer) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) {
	wi := workItems.NewItemValidators(
		di.elasticProcessor,
		epoch,
		validatorsPubKeys,
	)
	di.dispatcher.Add(wi)
}

// UpdateTPS updates the tps and statistics into elasticsearch index
func (di *dataIndexer) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	if tpsBenchmark == nil {
		log.Debug("indexer: update tps called, but the tpsBenchmark is nil")
		return
	}

	wi := workItems.NewItemTpsBenchmark(di.elasticProcessor, tpsBenchmark)
	di.dispatcher.Add(wi)
}

// SaveAccounts will save the provided accounts
func (di *dataIndexer) SaveAccounts(accounts []state.UserAccountHandler) {
	wi := workItems.NewItemAccounts(di.elasticProcessor, accounts)
	di.dispatcher.Add(wi)
}

// IsInterfaceNil returns true if there is no value under the interface
func (di *dataIndexer) IsInterfaceNil() bool {
	return di == nil
}
