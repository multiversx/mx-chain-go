package disabled

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/genesis"
)

// MessageSignVerifier represents the message verifier that accepts any message, sign, pk tuple
// and only verifies public key validity
type messageSignVerifier struct {
	kg crypto.KeyGenerator
}

// NewMessageSignVerifier creates a new message sign verifier - which verifies only the public key validity
func NewMessageSignVerifier(kg crypto.KeyGenerator) (*messageSignVerifier, error) {
	if check.IfNil(kg) {
		return nil, genesis.ErrNilKeyGenerator
	}
	return &messageSignVerifier{kg: kg}, nil
}

// Verify returns nil if public key is valid
func (msv *messageSignVerifier) Verify(_ []byte, _ []byte, pubKey []byte) error {
	if len(pubKey) == 0 {
		return genesis.ErrEmptyPubKey
	}
	err := msv.kg.CheckPublicKeyValid(pubKey)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (msv *messageSignVerifier) IsInterfaceNil() bool {
	return msv == nil
}
