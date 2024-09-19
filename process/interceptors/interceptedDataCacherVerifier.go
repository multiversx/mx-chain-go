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

func (idv *interceptedDataVerifier) IsInterfaceNil() bool {
	return idv == nil
}
