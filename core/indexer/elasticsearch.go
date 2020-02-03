package indexer

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("core/indexer")

// Options structure holds the indexer's configuration options
type Options struct {
	TxIndexingEnabled bool
}

//ElasticIndexerArgs is struct that is used to store all components that are needed to create a indexer
type ElasticIndexerArgs struct {
	Url              string
	UserName         string
	Password         string
	ShardCoordinator sharding.Coordinator
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	Options          *Options
}

type elasticIndexer struct {
	database         databaseHandler
	shardCoordinator sharding.Coordinator
	options          *Options
	isNilIndexer     bool
}

// NewElasticIndexer creates a new elasticIndexer where the server listens on the url, authentication for the server is
// using the username and password
func NewElasticIndexer(arguments ElasticIndexerArgs) (Indexer, error) {
	err := checkElasticSearchParams(arguments)
	if err != nil {
		return nil, err
	}

	databaseArguments := elasticSearchDatabaseArgs{
		url:         arguments.Url,
		userName:    arguments.UserName,
		password:    arguments.Password,
		marshalizer: arguments.Marshalizer,
		hasher:      arguments.Hasher,
	}
	client, err := newElasticSearchDatabase(databaseArguments)
	if err != nil {
		return nil, err
	}

	indexer := &elasticIndexer{
		database:         client,
		shardCoordinator: arguments.ShardCoordinator,
		options:          arguments.Options,
		isNilIndexer:     false,
	}

	return indexer, nil
}

// SaveBlock will build
func (ei *elasticIndexer) SaveBlock(
	bodyHandler data.BodyHandler,
	headerHandler data.HeaderHandler,
	txPool map[string]data.TransactionHandler,
	signersIndexes []uint64,
) {
	body, ok := bodyHandler.(block.Body)
	if !ok {
		log.Debug("indexer", "error", ErrBodyTypeAssertion.Error())
		return
	}

	if check.IfNil(headerHandler) {
		log.Debug("indexer: no header", "error", ErrNoHeader.Error())
		return
	}

	go ei.database.saveHeader(headerHandler, signersIndexes)

	if len(body) == 0 {
		log.Debug("indexer", "error", ErrNoMiniblocks.Error())
		return
	}

	if ei.options.TxIndexingEnabled {
		go ei.database.saveTransactions(body, headerHandler, txPool, ei.shardCoordinator.SelfId())
	}
}

// SaveMetaBlock will index a meta block in elastic search
func (ei *elasticIndexer) SaveMetaBlock(header data.HeaderHandler, signersIndexes []uint64) {
	if check.IfNil(header) {
		log.Debug("indexer: nil header", "error", ErrNoHeader.Error())
		return
	}

	go ei.database.saveHeader(header, signersIndexes)
}

// SaveRoundInfo will save data about a round on elastic search
func (ei *elasticIndexer) SaveRoundInfo(roundInfo RoundInfo) {
	ei.database.saveRoundInfo(roundInfo)
}

//SaveValidatorsPubKeys will send all validators public keys to elastic search
func (ei *elasticIndexer) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte) {
	valPubKeys := make(map[uint32][]string)
	for shardId, shardPubKeys := range validatorsPubKeys {
		for _, pubKey := range shardPubKeys {
			valPubKeys[shardId] = append(valPubKeys[shardId], hex.EncodeToString(pubKey))
		}
		go func(id uint32, publicKeys []string) {
			ei.database.saveShardValidatorsPubKeys(id, publicKeys)
		}(shardId, valPubKeys[shardId])
	}
}

// UpdateTPS updates the tps and statistics into elasticsearch index
func (ei *elasticIndexer) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	if tpsBenchmark == nil {
		log.Debug("indexer: update tps called, but the tpsBenchmark is nil")
		return
	}

	ei.database.saveShardStatistics(tpsBenchmark)
}

// IsNilIndexer will return a bool value that signals if the indexer's implementation is a NilIndexer
func (ei *elasticIndexer) IsNilIndexer() bool {
	return ei.isNilIndexer
}

// IsInterfaceNil returns true if there is no value under the interface
func (ei *elasticIndexer) IsInterfaceNil() bool {
	return ei == nil
}
