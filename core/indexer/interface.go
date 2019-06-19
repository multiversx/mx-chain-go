package indexer

import (
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// Indexer is an interface for saving node specific data to other storage.
// This could be an elasticsearch index, a MySql database or any other external services.
type Indexer interface {
	SaveBlock(body data.BodyHandler, header data.HeaderHandler, txPool map[string]*transaction.Transaction)
	UpdateTPS(tpsBenchmark statistics.TPSBenchmark)
}
