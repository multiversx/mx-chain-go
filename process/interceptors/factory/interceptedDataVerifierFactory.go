package factory

import (
	"time"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
)

// InterceptedDataVerifierFactoryArgs holds the required arguments for InterceptedDataVerifierFactory
type InterceptedDataVerifierFactoryArgs struct {
	CacheSpan   time.Duration
	CacheExpiry time.Duration
}

// InterceptedDataVerifierFactory encapsulates the required arguments to create InterceptedDataVerifier
// Furthermore it will hold all such instances in an internal map.
type InterceptedDataVerifierFactory struct {
	cacheSpan                  time.Duration
	cacheExpiry                time.Duration
	interceptedDataVerifierMap map[string]storage.Cacher
}

// NewInterceptedDataVerifierFactory will create a factory instance that will create instance of InterceptedDataVerifiers
func NewInterceptedDataVerifierFactory(args InterceptedDataVerifierFactoryArgs) *InterceptedDataVerifierFactory {
	return &InterceptedDataVerifierFactory{
		cacheSpan:                  args.CacheSpan,
		cacheExpiry:                args.CacheExpiry,
		interceptedDataVerifierMap: make(map[string]storage.Cacher),
	}
}

// Create will return an instance of InterceptedDataVerifier
func (idvf *InterceptedDataVerifierFactory) Create(topic string) (process.InterceptedDataVerifier, error) {
	internalCache, err := cache.NewTimeCacher(cache.ArgTimeCacher{
		DefaultSpan: idvf.cacheSpan,
		CacheExpiry: idvf.cacheExpiry,
	})
	if err != nil {
		return nil, err
	}

	idvf.interceptedDataVerifierMap[topic] = internalCache
	verifier := interceptors.NewInterceptedDataVerifier(internalCache)
	return verifier, nil
}
