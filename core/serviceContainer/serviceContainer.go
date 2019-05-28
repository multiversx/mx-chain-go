package serviceContainer

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/core/indexer"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/statistics"
)

type serviceContainer struct {
	indexer indexer.Indexer
	tpsBenchmark statistics.TPSBenchmark
}

// Option represents a functional configuration parameter that
//  can operate over the serviceContainer struct
type Option func(container *serviceContainer) error

// NewServiceContainer creates a new serviceContainer responsible in
//  providing access to all injected core features
func NewServiceContainer(opts ...Option) (Core, error) {
	sc :=  &serviceContainer{}
	for _, opt := range opts {
		err := opt(sc)
		if err != nil {
			return nil, err
		}
	}
	return sc, nil
}

// Indexer returns the core package's indexer
func (sc *serviceContainer) Indexer() indexer.Indexer {
	return sc.indexer
}

// TPSBenchmark returns the core package's tpsBenchmark
func (sc *serviceContainer) TPSBenchmark() statistics.TPSBenchmark {
	return sc.tpsBenchmark
}

// WithIndexer sets up the database indexer for the core serviceContainer
func WithIndexer(indexer indexer.Indexer) Option {
	return func(sc *serviceContainer) error {
		sc.indexer = indexer
		return nil
	}
}

// WithTPSBenchmark sets up the tpsBenchmark object for the core serviceContainer
func WithTPSBenchmark(tpsBenchmark statistics.TPSBenchmark) Option {
	return func(sc *serviceContainer) error {
		sc.tpsBenchmark = tpsBenchmark
		return nil
	}
}
