package components

import "github.com/multiversx/mx-chain-go/process"

type whiteListVerifier struct {
	shardID uint32
}

// NewWhiteListDataVerifier returns a default data verifier
func NewWhiteListDataVerifier(shardID uint32) (*whiteListVerifier, error) {
	return &whiteListVerifier{
		shardID: shardID,
	}, nil
}

// IsWhiteListed returns true
func (w *whiteListVerifier) IsWhiteListed(interceptedData process.InterceptedData) bool {
	interceptedTx, ok := interceptedData.(process.InterceptedTransactionHandler)
	if !ok {
		return true
	}

	if interceptedTx.SenderShardId() == w.shardID {
		return false
	}

	return true
}

// IsWhiteListedAtLeastOne returns true
func (w *whiteListVerifier) IsWhiteListedAtLeastOne(_ [][]byte) bool {
	return true
}

// Add does nothing
func (w *whiteListVerifier) Add(_ [][]byte) {
}

// Remove does nothing
func (w *whiteListVerifier) Remove(_ [][]byte) {
}

// IsInterfaceNil returns true if underlying object is nil
func (w *whiteListVerifier) IsInterfaceNil() bool {
	return w == nil
}
