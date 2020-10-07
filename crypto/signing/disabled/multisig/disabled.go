package multisig

import "github.com/ElrondNetwork/elrond-go/crypto"

const signature = "signature"

// Disabled represents a disabled multisigner implementation
type Disabled struct {
}

// Create returns a new disabled instance
func (d *Disabled) Create(_ []string, _ uint16) (crypto.MultiSigner, error) {
	return &Disabled{}, nil
}

// SetAggregatedSig returns nil
func (d *Disabled) SetAggregatedSig(_ []byte) error {
	return nil
}

// Verify returns nil
func (d *Disabled) Verify(_ []byte, _ []byte) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *Disabled) IsInterfaceNil() bool {
	return d == nil
}

// Reset returns nil
func (d *Disabled) Reset(_ []string, _ uint16) error {
	return nil
}

// CreateSignatureShare returns a mock signature value
func (d *Disabled) CreateSignatureShare(_ []byte, _ []byte) ([]byte, error) {
	return []byte(signature), nil
}

// StoreSignatureShare returns nil
func (d *Disabled) StoreSignatureShare(_ uint16, _ []byte) error {
	return nil
}

// SignatureShare returns a mock signature value
func (d *Disabled) SignatureShare(_ uint16) ([]byte, error) {
	return []byte(signature), nil
}

// VerifySignatureShare returns nil
func (d *Disabled) VerifySignatureShare(_ uint16, _ []byte, _ []byte, _ []byte) error {
	return nil
}

// AggregateSigs returns a mock signature value
func (d *Disabled) AggregateSigs(_ []byte) ([]byte, error) {
	return []byte(signature), nil
}
