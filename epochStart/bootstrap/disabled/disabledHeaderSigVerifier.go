package disabled

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type headerSigVerifier struct {
}

// NewHeaderSigVerifier returns a new instance of headerSigVerifier
func NewHeaderSigVerifier() *headerSigVerifier {
	return &headerSigVerifier{}
}

func (h *headerSigVerifier) VerifyRandSeedAndLeaderSignature(header data.HeaderHandler) error {
	return nil
}

func (h *headerSigVerifier) VerifySignature(header data.HeaderHandler) error {
	return nil
}

func (h *headerSigVerifier) IsInterfaceNil() bool {
	return h == nil
}
