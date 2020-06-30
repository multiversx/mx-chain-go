package interceptors

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const maxElementSize = 32

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

	for _, identifier := range interceptedData.Identifiers() {
		if w.cache.Has(identifier) {
			return true
		}
	}

	return false
}

// Add adds all the list to the cache
func (w *whiteListDataVerifier) Add(keys [][]byte) {
	for _, key := range keys {
		if len(key) > maxElementSize {
			log.Warn("whitelist add", "error", "key too large",
				"len", len(key),
			)
			continue
		}

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
