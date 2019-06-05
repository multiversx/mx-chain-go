package indexer

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/core/statistics"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

// Indexer is an interface for saving node specific data to other storages.
// This could be an elasticsearch intex, a MySql database or any other external services.
type Indexer interface {
	SaveBlock(body data.BodyHandler, header data.HeaderHandler, txPool map[string]*transaction.Transaction)
	UpdateTPS(tpsBenchmark statistics.TPSBenchmark)
}
