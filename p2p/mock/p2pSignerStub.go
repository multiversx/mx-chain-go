package mock

import "github.com/multiversx/mx-chain-core-go/core"

// P2PSignerStub -
type P2PSignerStub struct {
	SignCalled                func(payload []byte) ([]byte, error)
	VerifyCalled              func(payload []byte, pid core.PeerID, signature []byte) error
	SignUsingPrivateKeyCalled func(skBytes []byte, payload []byte) ([]byte, error)
}

// Sign -
func (stub *P2PSignerStub) Sign(payload []byte) ([]byte, error) {
	if stub.SignCalled != nil {
		return stub.SignCalled(payload)
	}

	return []byte{}, nil
}

// Verify -
func (stub *P2PSignerStub) Verify(payload []byte, pid core.PeerID, signature []byte) error {
	if stub.VerifyCalled != nil {
		return stub.VerifyCalled(payload, pid, signature)
	}

	return nil
}

// SignUsingPrivateKey -
func (stub *P2PSignerStub) SignUsingPrivateKey(skBytes []byte, payload []byte) ([]byte, error) {
	if stub.SignUsingPrivateKeyCalled != nil {
		return stub.SignUsingPrivateKeyCalled(skBytes, payload)
	}

	return []byte{}, nil
}

// IsInterfaceNil -
func (stub *P2PSignerStub) IsInterfaceNil() bool {
	return stub == nil
}
