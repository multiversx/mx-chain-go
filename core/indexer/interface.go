package indexer

import (
	"github.com/ElrondNetwork/elastic-indexer-go/workItems"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
)

// Indexer is an interface for saving node specific data to other storage.
// This could be an elastic search index, a MySql database or any other external services.
type Indexer interface {
	SetTxLogsProcessor(txLogsProc process.TransactionLogProcessorDatabase)
	//SaveBlock TODO add a structure instead of these params
	SaveBlock(body data.BodyHandler, header data.HeaderHandler, txPool map[string]data.TransactionHandler,
		signersIndexes []uint64, notarizedHeadersHashes []string, headerHash []byte)
	RevertIndexedBlock(header data.HeaderHandler, body data.BodyHandler)
	SaveRoundsInfo(roundsInfos []workItems.RoundInfo)
	UpdateTPS(tpsBenchmark statistics.TPSBenchmark)
	SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32)
	SaveValidatorsRating(indexID string, infoRating []workItems.ValidatorRatingInfo)
	SaveAccounts(acc []state.UserAccountHandler)
	Close() error
	IsInterfaceNil() bool
	IsNilIndexer() bool
}
