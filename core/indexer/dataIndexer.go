package indexer

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type dataIndexer struct {
	isNilIndexer bool
	dispatcher   DispatcherHandler
	coordinator  sharding.NodesCoordinator
}

// NewDataIndexer will create a new data indexer
func NewDataIndexer(arguments DataIndexerArgs) (Indexer, error) {
	err := checkDataIndexerParams(arguments)
	if err != nil {
		return nil, err
	}

	dispatcher, err := NewDataDispatcher(arguments)
	if err != nil {
		return nil, err
	}

	dispatcher.StartIndexData()

	dataIndexer := &dataIndexer{
		isNilIndexer: false,
		dispatcher:   dispatcher,
		coordinator:  arguments.NodesCoordinator,
	}

	if arguments.ShardID == core.MetachainShardId {
		arguments.EpochStartNotifier.RegisterHandler(dataIndexer.epochStartEventHandler())
	}

	return dataIndexer, nil
}

func (di *dataIndexer) epochStartEventHandler() epochStart.ActionHandler {
	subscribeHandler := notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {
		currentEpoch := hdr.GetEpoch()
		validatorsPubKeys, err := di.coordinator.GetAllEligibleValidatorsPublicKeys(currentEpoch)
		if err != nil {
			log.Warn("GetAllEligibleValidatorPublicKeys for current epoch failed",
				"epoch", currentEpoch,
				"error", err.Error())
		}

		go di.SaveValidatorsPubKeys(validatorsPubKeys, currentEpoch)

	}, func(_ data.HeaderHandler) {}, core.IndexerOrder)

	return subscribeHandler
}

// SaveBlock saves the block info in the queue to be sent to elastic
func (di *dataIndexer) SaveBlock(
	bodyHandler data.BodyHandler,
	headerHandler data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	signersIndexes []uint64,
	notarizedHeadersHashes []string,
) {
	wi := NewWorkItem(WorkTypeSaveBlock, &saveBlockData{
		bodyHandler:            bodyHandler,
		headerHandler:          headerHandler,
		txPool:                 txPool,
		signersIndexes:         signersIndexes,
		notarizedHeadersHashes: notarizedHeadersHashes,
	})
	di.dispatcher.Add(wi)
}

// GetQueueLength -
func (di *dataIndexer) GetQueueLength() int {
	return di.dispatcher.GetQueueLength()
}

// RevertIndexedBlock -
func (di *dataIndexer) RevertIndexedBlock(header data.HeaderHandler) {
	//TODO dont forget about this function
	// ask Cristi
}

// SaveRoundsInfos will save data about a slice of rounds on elasticsearch
func (di *dataIndexer) SaveRoundsInfos(roundsInfos []RoundInfo) {
	wi := NewWorkItem(WorkTypeSaveRoundsInfo, &roundsInfos)
	di.dispatcher.Add(wi)
}

// SaveValidatorsRating will send all validators rating info to elasticsearch
func (di *dataIndexer) SaveValidatorsRating(indexID string, validatorsRatingInfo []ValidatorRatingInfo) {
	wi := NewWorkItem(WorkTypeSaveValidatorsRating, &saveValidatorsRatingData{
		indexID:    indexID,
		infoRating: validatorsRatingInfo,
	})
	di.dispatcher.Add(wi)
}

// SaveValidatorsPubKeys will send all validators public keys to elasticsearch
func (di *dataIndexer) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) {
	wi := NewWorkItem(WorkTypeSaveValidatorsPubKeys, &saveValidatorsPubKeysData{
		epoch:             epoch,
		validatorsPubKeys: validatorsPubKeys,
	})
	di.dispatcher.Add(wi)
}

// UpdateTPS updates the tps and statistics into elasticsearch index
func (di *dataIndexer) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	wi := NewWorkItem(WorkTypeUpdateTPSBenchmark, &tpsBenchmark)
	di.dispatcher.Add(wi)
}

// SetTxLogsProcessor will set tx logs processor
func (di *dataIndexer) SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase) {

}

// IsNilIndexer will return a bool value that signals if the indexer's implementation is a NilIndexer
func (di *dataIndexer) IsNilIndexer() bool {
	return di.isNilIndexer
}

// IsInterfaceNil returns true if there is no value under the interface
func (di *dataIndexer) IsInterfaceNil() bool {
	return di == nil
}
