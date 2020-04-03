package disabled

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
)

type multiSignatureVerifier struct {
}

// NewMultiSigVerifier returns a new instance of multiSignatureVerifier
func NewMultiSigVerifier() *multiSignatureVerifier {
	return &multiSignatureVerifier{}
}

// Create will return a disabled multi signer
func (m *multiSignatureVerifier) Create(_ []string, _ uint16) (crypto.MultiSigner, error) {
	return NewMultiSigner(), nil
}

// SetAggregatedSig will return nil
func (m *multiSignatureVerifier) SetAggregatedSig([]byte) error {
	return nil
}

// Verify will return nil
func (m *multiSignatureVerifier) Verify(_ []byte, _ []byte) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *multiSignatureVerifier) IsInterfaceNil() bool {
	return m == nil
}
