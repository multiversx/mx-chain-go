package serviceContainer

import (
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/scwatcher"
)

// Core interface will abstract all the subpackage functionalities and will
//  provide access to it's members where needed
type Core interface {
	Indexer() indexer.Indexer
	TPSBenchmark() statistics.TPSBenchmark
	ScWatcherDriver() scwatcher.Driver
	IsInterfaceNil() bool
}
