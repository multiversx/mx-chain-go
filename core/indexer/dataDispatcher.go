package indexer

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
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

type dataDispatcher struct {
	elasticIndexer ElasticIndexer
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
