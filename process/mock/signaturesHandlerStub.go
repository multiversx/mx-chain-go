package mock

import "github.com/multiversx/mx-chain-core-go/core"

// SignaturesHandlerStub -
type SignaturesHandlerStub struct {
	VerifyCalled func(payload []byte, pid core.PeerID, signature []byte) error
}

// Verify -
func (s *SignaturesHandlerStub) Verify(payload []byte, pid core.PeerID, signature []byte) error {
	if s.VerifyCalled != nil {
		return s.VerifyCalled(payload, pid, signature)
	}
	return nil
}

// IsInterfaceNil -
func (s *SignaturesHandlerStub) IsInterfaceNil() bool {
	return false
}
