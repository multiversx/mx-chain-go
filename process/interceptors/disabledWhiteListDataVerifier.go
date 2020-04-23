package interceptors

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type disabledWhiteListVerifier struct {
}

// NewDisabledWhiteListDataVerifier returns a default data verifier
func NewDisabledWhiteListDataVerifier(_ storage.Cacher) (*disabledWhiteListVerifier, error) {
	return &disabledWhiteListVerifier{}, nil
}

// IsWhiteListed return true if intercepted data is accepted
func (w *disabledWhiteListVerifier) IsWhiteListed(interceptedData process.InterceptedData) bool {
	return false
}

// Add ads all the list to the cache
func (w *disabledWhiteListVerifier) Add(keys [][]byte) {
}

// Remove removes all the keys from the cache
func (w *disabledWhiteListVerifier) Remove(keys [][]byte) {
}

// IsInterfaceNil returns true if underlying object is nil
func (w *disabledWhiteListVerifier) IsInterfaceNil() bool {
	return w == nil
}
