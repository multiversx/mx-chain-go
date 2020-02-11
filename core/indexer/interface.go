package indexer

import (
	"bytes"
	"io"

	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// Indexer is an interface for saving node specific data to other storage.
// This could be an elastic search index, a MySql database or any other external services.
type Indexer interface {
	SaveBlock(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler, signersIndexes []uint64)
	SaveMetaBlock(header data.HeaderHandler, signersIndexes []uint64)
	SaveRoundInfo(roundInfo RoundInfo)
	UpdateTPS(tpsBenchmark statistics.TPSBenchmark)
	SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte)
	IsInterfaceNil() bool
	IsNilIndexer() bool
}

// databaseHandler is an interface used by elasticsearch component to prepare data to be saved on elasticseach server
type databaseHandler interface {
	SaveHeader(header data.HeaderHandler, signersIndexes []uint64)
	SaveTransactions(body block.Body, header data.HeaderHandler, txPool map[string]data.TransactionHandler, selfShardId uint32)
	SaveRoundInfo(info RoundInfo)
	SaveShardValidatorsPubKeys(shardId uint32, shardValidatorsPubKeys []string)
	SaveShardStatistics(tpsBenchmark statistics.TPSBenchmark)
}

// databaseWriterHandler is an interface that do requests to elasticsearch server do save data
type databaseWriterHandler interface {
	DoRequest(req esapi.IndexRequest) error
	DoBulkRequest(buff *bytes.Buffer, index string) error
	CheckAndCreateIndex(index string, body io.Reader) error
}
