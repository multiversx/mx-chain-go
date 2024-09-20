package interceptors

import (
	"errors"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

type interceptedDataStatus int

const (
	ValidInterceptedData interceptedDataStatus = iota
	InvalidInterceptedData
)

var (
	ErrInvalidInterceptedData = errors.New("invalid intercepted data")
)

type interceptedDataVerifier struct {
	cache storage.Cacher
}

func NewInterceptedDataVerifier(cache storage.Cacher) *interceptedDataVerifier {
	return &interceptedDataVerifier{cache: cache}
}

// Verify will check if the intercepted data has been validated before and put in the time cache.
// It will retrieve the status in the cache if it exists, otherwise it will validate it and store the status of the
// validation in the cache. Note that the entries are stored for a set period of time
func (idv *interceptedDataVerifier) Verify(interceptedData process.InterceptedData) error {
	if len(interceptedData.Hash()) == 0 {
		return nil
	}

	if val, ok := idv.cache.Get(interceptedData.Hash()); ok {
		if val == ValidInterceptedData {
			return nil
		}

		return ErrInvalidInterceptedData
	}

	err := interceptedData.CheckValidity()
	if err != nil {
		idv.cache.Put(interceptedData.Hash(), InvalidInterceptedData, 8)
		return ErrInvalidInterceptedData
	}

	idv.cache.Put(interceptedData.Hash(), ValidInterceptedData, 100)
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (idv *interceptedDataVerifier) IsInterfaceNil() bool {
	return idv == nil
}
