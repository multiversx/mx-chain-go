package singlesig

import "github.com/ElrondNetwork/elrond-go/crypto"

const signature = "signature"

// Disabled represents a disabled singlesigner implementation
type Disabled struct {
}

// Sign returns a mock signature value
func (d *Disabled) Sign(_ crypto.PrivateKey, _ []byte) ([]byte, error) {
	return []byte(signature), nil
}

// Verify returns nil
func (d *Disabled) Verify(_ crypto.PublicKey, _ []byte, _ []byte) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *Disabled) IsInterfaceNil() bool {
	return d == nil
}
