package indexer

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"time"
)

// Indexer is an interface for saving node specific data to other storage.
// This could be an elastic search index, a MySql database or any other external services.
type Indexer interface {
	SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase)
	SaveBlock(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler, signersIndexes []uint64, notarizedHeadersHashes []string)
	RevertIndexedBlock(header data.HeaderHandler)
	SaveRoundsInfos(roundsInfos []RoundInfo)
	UpdateTPS(tpsBenchmark statistics.TPSBenchmark)
	SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32)
	SaveValidatorsRating(indexID string, infoRating []ValidatorRatingInfo)
	GetQueueLength() int
	IsInterfaceNil() bool
	IsNilIndexer() bool
}

// DispatcherHandler -
type DispatcherHandler interface {
	StartIndexData()
	Close() error
	Add(item *workItem)
	GetQueueLength() int
}

// ElasticIndexer defines the interface for the elastic search indexer
type ElasticIndexer interface {
	SaveShardStatistics(tpsBenchmark statistics.TPSBenchmark) error
	SaveHeader(header data.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, txsSize int) error
	SaveMiniblocks(header data.HeaderHandler, body *block.Body) (map[string]bool, error)
	SaveTransactions(body *block.Body, header data.HeaderHandler, txPool map[string]data.TransactionHandler, selfShardID uint32, mbsInDb map[string]bool) error
	SaveValidatorsRating(index string, validatorsRatingInfo []ValidatorRatingInfo) error
	SaveRoundsInfos(infos []RoundInfo) error
	SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error
}

// QueueHandler defines the interface for the queue
type QueueHandler interface {
	Next() *workItem
	Add(item *workItem)
	Done()
	GotBackOff()
	GetBackOffTime() int64
	GetCycleTime() time.Duration
	Length() int
}
