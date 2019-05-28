package indexer

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/core/statistics"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

// Indexer is an interface for saving node specific data to other storages.
// This could be an elasticsearch intex, a MySql database or any other external services.
type Indexer interface {
	SaveBlock(body block.Body, header *block.Header, txPool map[string]*transaction.Transaction)
	SaveMetaBlock(metaBlock *block.MetaBlock, headerPool map[string]*block.Header)
	UpdateTPS(tpsBenchmark statistics.TPSBenchmark)
}
