package indexer

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// DataIndexerFactory can create new instances of Indexer
type DataIndexerFactory interface {
	Create() (Indexer, error)
	IsInterfaceNil() bool
}

// Indexer is an interface for saving node specific data to other storage.
// This could be an elastic search index, a MySql database or any other external services.
type Indexer interface {
	SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase)
	SaveBlock(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler, signersIndexes []uint64, notarizedHeadersHashes []string)
	RevertIndexedBlock(header data.HeaderHandler, body data.BodyHandler)
	SaveRoundsInfo(roundsInfos []workItems.RoundInfo)
	UpdateTPS(tpsBenchmark statistics.TPSBenchmark)
	SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32)
	SaveValidatorsRating(indexID string, infoRating []workItems.ValidatorRatingInfo)
	StopIndexing() error
	IsInterfaceNil() bool
	IsNilIndexer() bool
}

// DispatcherHandler is an interface for the data dispatcher data is used to index data in database
type DispatcherHandler interface {
	StartIndexData()
	Close() error
	Add(item workItems.WorkItemHandler)
}

// ElasticProcessor defines the interface for the elastic search indexer
type ElasticProcessor interface {
	SaveShardStatistics(tpsBenchmark statistics.TPSBenchmark) error
	SaveHeader(header data.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, txsSize int) error
	RemoveHeader(header data.HeaderHandler) error
	RemoveMiniblocks(header data.HeaderHandler, body *block.Body) error
	SaveMiniblocks(header data.HeaderHandler, body *block.Body) (map[string]bool, error)
	SaveTransactions(body *block.Body, header data.HeaderHandler, txPool map[string]data.TransactionHandler, selfShardID uint32, mbsInDb map[string]bool) error
	SaveValidatorsRating(index string, validatorsRatingInfo []workItems.ValidatorRatingInfo) error
	SaveRoundsInfo(infos []workItems.RoundInfo) error
	SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error
	SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase)
}

// DatabaseClientHandler is an interface that do requests to elasticsearch server
type DatabaseClientHandler interface {
	DoRequest(req *esapi.IndexRequest) error
	DoBulkRequest(buff *bytes.Buffer, index string) error
	DoBulkRemove(index string, hashes []string) error
	DoMultiGet(query object, index string) (object, error)

	CheckAndCreateIndex(index string) error
	CheckAndCreateAlias(alias string, indexName string) error
	CheckAndCreateTemplate(templateName string, template *bytes.Buffer) error
	CheckAndCreatePolicy(policyName string, policy *bytes.Buffer) error
}
