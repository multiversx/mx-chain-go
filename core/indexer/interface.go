package indexer

import (
	"bytes"
	"io"

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
	SaveRoundInfo(roundInfo RoundInfo)
	UpdateTPS(tpsBenchmark statistics.TPSBenchmark)
	SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32)
	SaveValidatorsRating(indexID string, infoRating []ValidatorRatingInfo)
	IsInterfaceNil() bool
	IsNilIndexer() bool
}

// databaseHandler is an interface used by elasticsearch component to prepare data to be saved on elasticseach server
type databaseHandler interface {
	SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase)
	SaveHeader(header data.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, txsSize int)
	SaveMiniblocks(header data.HeaderHandler, body *block.Body)
	SaveTransactions(body *block.Body, header data.HeaderHandler, txPool map[string]data.TransactionHandler, selfShardId uint32)
	SaveRoundInfo(info RoundInfo)
	SaveShardValidatorsPubKeys(shardId, epoch uint32, shardValidatorsPubKeys [][]byte)
	SaveValidatorsRating(Index string, validatorsRatingInfo []ValidatorRatingInfo)
	SaveShardStatistics(tpsBenchmark statistics.TPSBenchmark)
}

// databaseWriterHandler is an interface that do requests to elasticsearch server do save data
type databaseWriterHandler interface {
	DoRequest(req *esapi.IndexRequest) error
	DoBulkRequest(buff *bytes.Buffer, index string) error
	CheckAndCreateIndex(index string, body io.Reader) error
}
