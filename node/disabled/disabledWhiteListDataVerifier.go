package disabled

import (
	"github.com/multiversx/mx-chain-go/process"
)

type disabledWhiteListVerifier struct {
}

// NewDisabledWhiteListDataVerifier returns a default data verifier
func NewDisabledWhiteListDataVerifier() *disabledWhiteListVerifier {
	return &disabledWhiteListVerifier{}
}

// IsWhiteListed returns false
func (w *disabledWhiteListVerifier) IsWhiteListed(_ process.InterceptedData) bool {
	return false
}

// IsWhiteListedAtLeastOne returns true
func (w *disabledWhiteListVerifier) IsWhiteListedAtLeastOne(_ [][]byte) bool {
	return true
}

// Add does nothing
func (w *disabledWhiteListVerifier) Add(_ [][]byte) {
}

// Remove does nothing
func (w *disabledWhiteListVerifier) Remove(_ [][]byte) {
}

// IsInterfaceNil returns true if underlying object is nil
func (w *disabledWhiteListVerifier) IsInterfaceNil() bool {
	return w == nil
}
