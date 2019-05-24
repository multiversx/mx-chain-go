package core

type serviceContainer struct {
	indexer Indexer
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
func (sc *serviceContainer) Indexer() Indexer {
	return sc.indexer
}

// WithIndexer sets up the database indexer for the core serviceContainer
func WithIndexer(indexer Indexer) Option {
	return func(sc *serviceContainer) error {
		sc.indexer = indexer
		return nil
	}
}
