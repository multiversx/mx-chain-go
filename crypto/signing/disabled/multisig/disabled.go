package multisig

import "github.com/ElrondNetwork/elrond-go/crypto"

const signature = "signature"

// DisabledMultiSig represents a disabled multisigner implementation
type DisabledMultiSig struct {
}

// Create returns a new disabled instance
func (dms *DisabledMultiSig) Create(_ []string, _ uint16) (crypto.MultiSigner, error) {
	return &DisabledMultiSig{}, nil
}

// SetAggregatedSig returns nil
func (dms *DisabledMultiSig) SetAggregatedSig(_ []byte) error {
	return nil
}

// Verify returns nil
func (dms *DisabledMultiSig) Verify(_ []byte, _ []byte) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dms *DisabledMultiSig) IsInterfaceNil() bool {
	return dms == nil
}

// Reset returns nil
func (dms *DisabledMultiSig) Reset(_ []string, _ uint16) error {
	return nil
}

// CreateSignatureShare returns a mock signature value
func (dms *DisabledMultiSig) CreateSignatureShare(_ []byte, _ []byte) ([]byte, error) {
	return []byte(signature), nil
}

// StoreSignatureShare returns nil
func (dms *DisabledMultiSig) StoreSignatureShare(_ uint16, _ []byte) error {
	return nil
}

// SignatureShare returns a mock signature value
func (dms *DisabledMultiSig) SignatureShare(_ uint16) ([]byte, error) {
	return []byte(signature), nil
}

// VerifySignatureShare returns nil
func (dms *DisabledMultiSig) VerifySignatureShare(_ uint16, _ []byte, _ []byte, _ []byte) error {
	return nil
}

// AggregateSigs returns a mock signature value
func (dms *DisabledMultiSig) AggregateSigs(_ []byte) ([]byte, error) {
	return []byte(signature), nil
}
