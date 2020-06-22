package interceptors

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ process.WhiteListHandler = (*whiteListDataVerifier)(nil)

type whiteListDataVerifier struct {
	cache storage.Cacher
}

// NewWhiteListDataVerifier returns a default data verifier
func NewWhiteListDataVerifier(cacher storage.Cacher) (*whiteListDataVerifier, error) {
	if check.IfNil(cacher) {
		return nil, fmt.Errorf("%w in NewWhiteListDataVerifier", process.ErrNilCacher)
	}

	return &whiteListDataVerifier{
		cache: cacher,
	}, nil
}

// IsWhiteListed return true if intercepted data is accepted
func (w *whiteListDataVerifier) IsWhiteListed(interceptedData process.InterceptedData) bool {
	if check.IfNil(interceptedData) {
		return false
	}

	return w.cache.Has(interceptedData.Hash())
}

// Add adds all the list to the cache
func (w *whiteListDataVerifier) Add(keys [][]byte) {
	for _, key := range keys {
		_ = w.cache.Put(key, struct{}{}, 0)
	}
}

// Remove removes all the keys from the cache
func (w *whiteListDataVerifier) Remove(keys [][]byte) {
	for _, key := range keys {
		w.cache.Remove(key)
	}
}

// IsInterfaceNil returns true if underlying object is nil
func (w *whiteListDataVerifier) IsInterfaceNil() bool {
	return w == nil
}
