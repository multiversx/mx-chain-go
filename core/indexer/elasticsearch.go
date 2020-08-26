package indexer

import (
	"fmt"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("core/indexer")

// Options structure holds the indexer's configuration options
type Options struct {
	TxIndexingEnabled bool
}

//ElasticIndexerArgs is struct that is used to store all components that are needed to create a indexer
type ElasticIndexerArgs struct {
	ShardId                  uint32
	Url                      string
	UserName                 string
	Password                 string
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	EpochStartNotifier       sharding.EpochStartEventNotifier
	NodesCoordinator         sharding.NodesCoordinator
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	Options                  *Options
}

type elasticIndexer struct {
	database     DatabaseHandler
	options      *Options
	coordinator  sharding.NodesCoordinator
	marshalizer  marshal.Marshalizer
	isNilIndexer bool
}

// NewElasticIndexer creates a new elasticIndexer where the server listens on the url, authentication for the server is
// using the username and password
func NewElasticIndexer(arguments ElasticIndexerArgs) (Indexer, error) {
	err := checkElasticSearchParams(arguments)
	if err != nil {
		return nil, err
	}

	databaseArguments := ElasticSearchDatabaseArgs{
		AddressPubkeyConverter:   arguments.AddressPubkeyConverter,
		ValidatorPubkeyConverter: arguments.ValidatorPubkeyConverter,
		Url:                      arguments.Url,
		UserName:                 arguments.UserName,
		Password:                 arguments.Password,
		Marshalizer:              arguments.Marshalizer,
		Hasher:                   arguments.Hasher,
	}
	client, err := NewElasticSearchDatabase(databaseArguments)
	if err != nil {
		return nil, fmt.Errorf("cannot create indexer: %w", err)
	}

	indexer := &elasticIndexer{
		database:     client,
		options:      arguments.Options,
		coordinator:  arguments.NodesCoordinator,
		marshalizer:  arguments.Marshalizer,
		isNilIndexer: false,
	}

	if arguments.ShardId == core.MetachainShardId {
		arguments.EpochStartNotifier.RegisterHandler(indexer.epochStartEventHandler())
	}

	return indexer, nil
}

// SaveBlock will build
func (ei *elasticIndexer) SaveBlock(
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

	txsSizeInBytes := computeSizeOfTxs(ei.marshalizer, txPool)
	ei.database.SaveHeader(headerHandler, signersIndexes, body, notarizedHeadersHashes, txsSizeInBytes)

	if len(body.MiniBlocks) == 0 {
		return
	}

	mbsInDb := ei.database.SaveMiniblocks(headerHandler, body)

	if ei.options.TxIndexingEnabled {
		ei.database.SaveTransactions(body, headerHandler, txPool, headerHandler.GetShardID(), mbsInDb)
	}
}

// SaveRoundsInfos will save data about a slice of rounds on elasticsearch
func (ei *elasticIndexer) SaveRoundsInfos(roundsInfos []RoundInfo) {
	ei.database.SaveRoundsInfos(roundsInfos)
}

func (ei *elasticIndexer) epochStartEventHandler() epochStart.ActionHandler {
	subscribeHandler := notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {
		currentEpoch := hdr.GetEpoch()
		validatorsPubKeys, err := ei.coordinator.GetAllEligibleValidatorsPublicKeys(currentEpoch)
		if err != nil {
			log.Warn("GetAllEligibleValidatorPublicKeys for current epoch failed",
				"epoch", currentEpoch,
				"error", err.Error())
		}

		go ei.SaveValidatorsPubKeys(validatorsPubKeys, currentEpoch)

	}, func(_ data.HeaderHandler) {}, core.IndexerOrder)

	return subscribeHandler
}

// SaveValidatorsRating will send all validators rating info to elasticsearch
func (ei *elasticIndexer) SaveValidatorsRating(indexID string, validatorsRatingInfo []ValidatorRatingInfo) {
	if validatorsRatingInfo != nil && indexID != "" {
		ei.database.SaveValidatorsRating(indexID, validatorsRatingInfo)
	}
}

//SaveValidatorsPubKeys will send all validators public keys to elasticsearch
func (ei *elasticIndexer) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) {
	for shardID, shardPubKeys := range validatorsPubKeys {
		ei.database.SaveShardValidatorsPubKeys(shardID, epoch, shardPubKeys)
	}
}

// UpdateTPS updates the tps and statistics into elasticsearch index
func (ei *elasticIndexer) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	if tpsBenchmark == nil {
		log.Debug("indexer: update tps called, but the tpsBenchmark is nil")
		return
	}

	ei.database.SaveShardStatistics(tpsBenchmark)
}

// SetTxLogsProcessor will set tx logs processor
func (ei *elasticIndexer) SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase) {
	ei.database.SetTxLogsProcessor(txLogsProc)
}

// IsNilIndexer will return a bool value that signals if the indexer's implementation is a NilIndexer
func (ei *elasticIndexer) IsNilIndexer() bool {
	return ei.isNilIndexer
}

// IsInterfaceNil returns true if there is no value under the interface
func (ei *elasticIndexer) IsInterfaceNil() bool {
	return ei == nil
}
