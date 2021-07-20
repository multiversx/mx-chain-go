package mock

import (
	"github.com/ElrondNetwork/elrond-go/common/statistics"
	"github.com/ElrondNetwork/elrond-go/process"
)

// StatusComponentsMock -
type StatusComponentsMock struct {
	Indexer      process.Indexer
	TPSBenchmark statistics.TPSBenchmark
}

// ElasticIndexer -
func (scm *StatusComponentsMock) ElasticIndexer() process.Indexer {
	return scm.Indexer
}

// TpsBenchmark -
func (scm *StatusComponentsMock) TpsBenchmark() statistics.TPSBenchmark {
	return scm.TPSBenchmark
}

// IsInterfaceNil -
func (scm *StatusComponentsMock) IsInterfaceNil() bool {
	return scm == nil
}
