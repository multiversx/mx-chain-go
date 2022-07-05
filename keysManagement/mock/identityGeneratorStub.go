package mock

import "github.com/ElrondNetwork/elrond-go-core/core"

// IdentityGeneratorStub -
type IdentityGeneratorStub struct {
	CreateRandomP2PIdentityCalled func() ([]byte, core.PeerID, error)
}

// CreateRandomP2PIdentity -
func (stub *IdentityGeneratorStub) CreateRandomP2PIdentity() ([]byte, core.PeerID, error) {
	if stub.CreateRandomP2PIdentityCalled != nil {
		return stub.CreateRandomP2PIdentityCalled()
	}

	return make([]byte, 0), "", nil
}

// IsInterfaceNil -
func (stub *IdentityGeneratorStub) IsInterfaceNil() bool {
	return stub == nil
}
