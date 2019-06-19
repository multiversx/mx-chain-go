package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
)

// ServiceContainerMock is a mock implementation of the Core interface
type ServiceContainerMock struct {
	IndexerCalled      func() indexer.Indexer
	TPSBenchmarkCalled func() statistics.TPSBenchmark
}

// Indexer returns a mock implementation for core.Indexer
func (scm *ServiceContainerMock) Indexer() indexer.Indexer {
	if scm.IndexerCalled != nil {
		return scm.IndexerCalled()
	}
	return nil
}

// TPSBenchmark returns a mock implementation for core.TPSBenchmark
func (scm *ServiceContainerMock) TPSBenchmark() statistics.TPSBenchmark {
	if scm.TPSBenchmarkCalled != nil {
		return scm.TPSBenchmarkCalled()
	}
	return nil
}
