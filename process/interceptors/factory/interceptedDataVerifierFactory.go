package factory

import (
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/storage/disabled"
)

// InterceptedDataVerifierFactoryArgs holds the required arguments for interceptedDataVerifierFactory
type InterceptedDataVerifierFactoryArgs struct {
	InterceptedDataVerifierConfig config.InterceptedDataVerifierConfig
}

// interceptedDataVerifierFactory encapsulates the required arguments to create InterceptedDataVerifier
// Furthermore it will hold all such instances in an internal map.
type interceptedDataVerifierFactory struct {
	cfg config.InterceptedDataVerifierConfig

	interceptedDataVerifierMap map[string]storage.Cacher
	mutex                      sync.Mutex
}

// NewInterceptedDataVerifierFactory will create a factory instance that will create instance of InterceptedDataVerifiers
func NewInterceptedDataVerifierFactory(args InterceptedDataVerifierFactoryArgs) *interceptedDataVerifierFactory {
	return &interceptedDataVerifierFactory{
		cfg:                        args.InterceptedDataVerifierConfig,
		interceptedDataVerifierMap: make(map[string]storage.Cacher),
		mutex:                      sync.Mutex{},
	}
}

// Create will return an instance of InterceptedDataVerifier
func (idvf *interceptedDataVerifierFactory) Create(topic string) (process.InterceptedDataVerifier, error) {
	internalCache, err := idvf.createCache()
	if err != nil {
		return nil, err
	}

	idvf.mutex.Lock()
	idvf.interceptedDataVerifierMap[topic] = internalCache
	idvf.mutex.Unlock()

	return interceptors.NewInterceptedDataVerifier(internalCache)
}

func (idvf *interceptedDataVerifierFactory) createCache() (storage.Cacher, error) {
	if !idvf.cfg.EnableCaching {
		return disabled.NewCache(), nil
	}

	return cache.NewTimeCacher(cache.ArgTimeCacher{
		DefaultSpan: time.Duration(idvf.cfg.CacheSpanInSec) * time.Second,
		CacheExpiry: time.Duration(idvf.cfg.CacheExpiryInSec) * time.Second,
	})
}

// Close will close all the sweeping routines created by the cache.
func (idvf *interceptedDataVerifierFactory) Close() error {
	for topic, cacher := range idvf.interceptedDataVerifierMap {
		err := cacher.Close()
		if err != nil {
			return fmt.Errorf("failed to close cacher on topic %q: %w", topic, err)
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (idvf *interceptedDataVerifierFactory) IsInterfaceNil() bool {
	return idvf == nil
}
