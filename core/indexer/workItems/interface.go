package workItems

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// WorkItemHandler defines the interface for item that needs to be saved in elasticsearch database
type WorkItemHandler interface {
	Save() error
	IsInterfaceNil() bool
}

type saveBlockIndexer interface {
	SaveHeader(header data.HeaderHandler, signersIndexes []uint64, body *block.Body, notarizedHeadersHashes []string, txsSize int) error
	SaveMiniblocks(header data.HeaderHandler, body *block.Body) (map[string]bool, error)
	SaveTransactions(body *block.Body, header data.HeaderHandler, txPool map[string]data.TransactionHandler, selfShardID uint32, mbsInDb map[string]bool) error
}

type saveRatingIndexer interface {
	SaveValidatorsRating(index string, validatorsRatingInfo []ValidatorRatingInfo) error
}

type removeIndexer interface {
	RemoveHeader(header data.HeaderHandler) error
	RemoveMiniblocks(header data.HeaderHandler, body *block.Body) error
}

type saveRounds interface {
	SaveRoundsInfo(infos []RoundInfo) error
}

type saveTpsBenchmark interface {
	SaveShardStatistics(tpsBenchmark statistics.TPSBenchmark) error
}

type saveValidatorsIndexer interface {
	SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error
}
