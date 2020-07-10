package indexer

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

// Indexer is an interface for saving node specific data to other storage.
// This could be an elastic search index, a MySql database or any other external services.
type Indexer interface {
	SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase)
	SaveBlock(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler, signersIndexes []uint64, notarizedHeadersHashes []string)
	SaveRoundsInfos(roundsInfos []RoundInfo)
	UpdateTPS(tpsBenchmark statistics.TPSBenchmark)
	SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32)
	SaveValidatorsRating(indexID string, infoRating []ValidatorRatingInfo)
	IsInterfaceNil() bool
	IsNilIndexer() bool
}

type ElasticIndexer interface {
	SaveHeader(header data.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, txsSize int)
}

