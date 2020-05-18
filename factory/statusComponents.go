package factory

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
)

// TODO: finish this implementation

type statusComponents struct {
	statusHandler   core.AppStatusHandler
	tpsBenchmark    statistics.TPSBenchmark
	elasticSearch   indexer.Indexer
	softwareVersion statistics.SoftwareVersionChecker
}
