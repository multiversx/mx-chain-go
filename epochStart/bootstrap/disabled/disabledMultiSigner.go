package disabled

import crypto "github.com/multiversx/mx-chain-crypto-go"

type multiSigner struct {
}

// NewMultiSigner returns a new instance of disabled multiSigner
func NewMultiSigner() *multiSigner {
	return &multiSigner{}
}

// CreateSignatureShare returns nil byte slice and nil error
func (m *multiSigner) CreateSignatureShare(_ []byte, _ []byte) ([]byte, error) {
	return nil, nil
}

// CreateSignatureShare returns nil byte slice and nil error
func (m *multiSigner) CreateSignatureShareV2(_ crypto.PrivateKey, _ []byte) ([]byte, error) {
	return nil, nil
}

// VerifySignatureShare returns nil
func (m *multiSigner) VerifySignatureShare(_ []byte, _ []byte, _ []byte) error {
	return nil
}

// VerifySignatureShare returns nil
func (m *multiSigner) VerifySignatureShareV2(_ crypto.PublicKey, _ []byte, _ []byte) error {
	return nil
}

// AggregateSigs returns nil byte slice and nil error
func (m *multiSigner) AggregateSigs(_ [][]byte, _ [][]byte) ([]byte, error) {
	return nil, nil
}

// AggregateSigsV2 returns nil byte slice and nil error
func (m *multiSigner) AggregateSigsV2(_ []crypto.PublicKey, _ [][]byte) ([]byte, error) {
	return nil, nil
}

// VerifyAggregatedSig returns nil
func (m *multiSigner) VerifyAggregatedSig(_ [][]byte, _ []byte, _ []byte) error {
	return nil
}

// VerifyAggregatedSigV2 returns nil
func (m *multiSigner) VerifyAggregatedSigV2(_ []crypto.PublicKey, _ []byte, _ []byte) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *multiSigner) IsInterfaceNil() bool {
	return m == nil
}
