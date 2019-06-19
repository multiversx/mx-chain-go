package serviceContainer

import (
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
)

// Core interface will abstract all the subpackage functionalities and will
//  provide access to it's members where needed
type Core interface {
	Indexer() indexer.Indexer
	TPSBenchmark() statistics.TPSBenchmark
}
