package interceptors

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/sync"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

type interceptedDataStatus int8

const (
	validInterceptedData           interceptedDataStatus = iota
	interceptedDataStatusBytesSize                       = 8
)

type interceptedDataVerifier struct {
	km    sync.KeyRWMutexHandler
	cache storage.Cacher
}

// NewInterceptedDataVerifier creates a new instance of intercepted data verifier
func NewInterceptedDataVerifier(cache storage.Cacher) (*interceptedDataVerifier, error) {
	if check.IfNil(cache) {
		return nil, process.ErrNilInterceptedDataCache
	}

	return &interceptedDataVerifier{
		km:    sync.NewKeyRWMutex(),
		cache: cache,
	}, nil
}

// Verify will check if the intercepted data has been validated before and put in the time cache.
// It will retrieve the status in the cache if it exists, otherwise it will validate it and store the status of the
// validation in the cache. Note that the entries are stored for a set period of time
func (idv *interceptedDataVerifier) Verify(interceptedData process.InterceptedData) error {
	if len(interceptedData.Hash()) == 0 {
		return interceptedData.CheckValidity()
	}

	hash := string(interceptedData.Hash())
	idv.km.Lock(hash)
	defer idv.km.Unlock(hash)

	if val, ok := idv.cache.Get(interceptedData.Hash()); ok {
		if val == validInterceptedData {
			return nil
		}

		return process.ErrInvalidInterceptedData
	}

	err := interceptedData.CheckValidity()
	if err != nil {
		log.Debug("Intercepted data is invalid", "hash", interceptedData.Hash(), "err", err)
		// TODO: investigate to selectively add as invalid intercepted data only when data is indeed invalid instead of missing
		// idv.cache.Put(interceptedData.Hash(), invalidInterceptedData, interceptedDataStatusBytesSize)
		return process.ErrInvalidInterceptedData
	}

	idv.cache.Put(interceptedData.Hash(), validInterceptedData, interceptedDataStatusBytesSize)
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (idv *interceptedDataVerifier) IsInterfaceNil() bool {
	return idv == nil
}
