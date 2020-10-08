package singlesig

import "github.com/ElrondNetwork/elrond-go/crypto"

const signature = "signature"

// DisabledSingleSig represents a disabled singleSigner implementation
type DisabledSingleSig struct {
}

// Sign returns a mock signature value
func (dss *DisabledSingleSig) Sign(_ crypto.PrivateKey, _ []byte) ([]byte, error) {
	return []byte(signature), nil
}

// Verify returns nil
func (dss *DisabledSingleSig) Verify(_ crypto.PublicKey, _ []byte, _ []byte) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dss *DisabledSingleSig) IsInterfaceNil() bool {
	return dss == nil
}
