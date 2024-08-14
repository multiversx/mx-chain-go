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

// VerifyRandSeed returns nil as it is disabled
func (h *headerSigVerifier) VerifyRandSeed(_ data.HeaderHandler) error {
	return nil
}

// VerifyLeaderSignature returns nil as it is disabled
func (h *headerSigVerifier) VerifyLeaderSignature(_ data.HeaderHandler) error {
	return nil
}

// VerifyRandSeedAndLeaderSignature returns nil as it is disabled
func (h *headerSigVerifier) VerifyRandSeedAndLeaderSignature(_ data.HeaderHandler) error {
	return nil
}

// VerifySignature returns nil as it is disabled
func (h *headerSigVerifier) VerifySignature(_ data.HeaderHandler) error {
	return nil
}

// VerifySignatureForHash returns nil as it is disabled
func (h *headerSigVerifier) VerifySignatureForHash(_ data.HeaderHandler, _ []byte, _ []byte, _ []byte) error {
	return nil
}

// VerifyPreviousBlockProof returns nil as it is disabled
func (h *headerSigVerifier) VerifyPreviousBlockProof(_ data.HeaderHandler) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *headerSigVerifier) IsInterfaceNil() bool {
	return h == nil
}
