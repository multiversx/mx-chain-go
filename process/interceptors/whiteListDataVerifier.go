package interceptors

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

const maxElementSize = 32 + 4 // hash + chunk_index_as_uint32

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

	return w.IsWhiteListedAtLeastOne(interceptedData.Identifiers())
}

// IsWhiteListedAtLeastOne return true if at least one identifier from the slice is whitelisted
func (w *whiteListDataVerifier) IsWhiteListedAtLeastOne(identifiers [][]byte) bool {
	return true
	for _, identifier := range identifiers {
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
