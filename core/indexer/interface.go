package indexer

import (
	"bytes"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/elastic/go-elasticsearch/v7/esapi"
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
	StopIndexing() error
	GetQueueLength() int // TODO instead of length return percent
	IsInterfaceNil() bool
	IsNilIndexer() bool
}

// DispatcherHandler is an interface for the data dispatcher data is used to index data in database
// TODO implement a new data dispatcher that knows to save every index in a different queue
type DispatcherHandler interface {
	StartIndexData()
	Close() error
	Add(item *workItem)
	GetQueueLength() int
	SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase)
}

// ElasticIndexer defines the interface for the elastic search indexer
type ElasticIndexer interface {
	SaveShardStatistics(tpsBenchmark statistics.TPSBenchmark) error
	SaveHeader(header data.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, txsSize int) error
	RemoveHeader(header data.HeaderHandler) error
	RemoveMiniblocks(header data.HeaderHandler) error
	SaveMiniblocks(header data.HeaderHandler, body *block.Body) (map[string]bool, error)
	SaveTransactions(body *block.Body, header data.HeaderHandler, txPool map[string]data.TransactionHandler, selfShardID uint32, mbsInDb map[string]bool) error
	SaveValidatorsRating(index string, validatorsRatingInfo []ValidatorRatingInfo) error
	SaveRoundsInfos(infos []RoundInfo) error
	SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error
	SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase)
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

// databaseClientHandler is an interface that do requests to elasticsearch server
type databaseClientHandler interface {
	DoRequest(req *esapi.IndexRequest) error
	DoBulkRequest(buff *bytes.Buffer, index string) error
	DoBulkRemove(index string, hashes []string) error
	DoMultiGet(query object, index string) (object, error)
	CheckAndCreateIndex(index string) error
	CheckAndCreateAlias(alias string, indexName string) error
	CheckAndCreateTemplate(templateName string, template *bytes.Buffer) error
	CheckAndCreatePolicy(policyName string, policy *bytes.Buffer) error
}
