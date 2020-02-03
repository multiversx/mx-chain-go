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

// databaseHandler --
type databaseHandler interface {
	saveHeader(header data.HeaderHandler, signersIndexes []uint64)
	saveTransactions(body block.Body, header data.HeaderHandler, txPool map[string]data.TransactionHandler, selfShardId uint32)
	saveRoundInfo(info RoundInfo)
	saveShardValidatorsPubKeys(shardId uint32, shardValidatorsPubKeys []string)
	saveShardStatistics(tpsBenchmark statistics.TPSBenchmark)
}

// databaseWriterHandler --
type databaseWriterHandler interface {
	doRequest(req esapi.IndexRequest) error
	doBulkRequest(buff *bytes.Buffer, index string) error
	checkAndCreateIndex(index string, body io.Reader) error
}
