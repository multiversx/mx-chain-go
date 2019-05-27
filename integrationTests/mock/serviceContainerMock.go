package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/core"
)

// ServiceContainerMock is a mock implementation of the Core interface
type ServiceContainerMock struct {
	IndexerCalled func() core.Indexer
	TPSBenchmarkCalled func() core.TPSBenchmark
}

// Indexer returns a mock implementation for core.Indexer
func (scm *ServiceContainerMock) Indexer() core.Indexer {
	if scm.IndexerCalled != nil {
		return scm.IndexerCalled()
	}
	return nil
}

// TPSBenchmark returns a mock implementation for core.TPSBenchmark
func (scm *ServiceContainerMock) TPSBenchmark() core.TPSBenchmark {
	if scm.TPSBenchmarkCalled != nil {
		return scm.TPSBenchmarkCalled()
	}
	return nil
}