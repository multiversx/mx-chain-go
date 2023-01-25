package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
)

var _ process.InterceptedHeaderSigVerifier = (*headerSigVerifier)(nil)

type headerSigVerifier struct {
}

// NewHeaderSigVerifier returns a new instance of headerSigVerifier
func NewHeaderSigVerifier() *headerSigVerifier {
	return &headerSigVerifier{}
}

// VerifyRandSeed -
func (h *headerSigVerifier) VerifyRandSeed(_ data.HeaderHandler) error {
	return nil
}

// VerifyLeaderSignature -
func (h *headerSigVerifier) VerifyLeaderSignature(_ data.HeaderHandler) error {
	return nil
}

// VerifyRandSeedAndLeaderSignature -
func (h *headerSigVerifier) VerifyRandSeedAndLeaderSignature(_ data.HeaderHandler) error {
	return nil
}

// VerifySignature -
func (h *headerSigVerifier) VerifySignature(_ data.HeaderHandler) error {
	return nil
}

// IsInterfaceNil -
func (h *headerSigVerifier) IsInterfaceNil() bool {
	return h == nil
}
