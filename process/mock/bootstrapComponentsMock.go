package mock

import (
	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// BootstrapComponentsMock -
type BootstrapComponentsMock struct {
	Coordinator          sharding.Coordinator
	HdrIntegrityVerifier factory.HeaderIntegrityVerifierHandler
}

// ShardCoordinator -
func (bcm *BootstrapComponentsMock) ShardCoordinator() sharding.Coordinator {
	return bcm.Coordinator
}

// HeaderIntegrityVerifier -
func (bcm *BootstrapComponentsMock) HeaderIntegrityVerifier() factory.HeaderIntegrityVerifierHandler {
	return bcm.HdrIntegrityVerifier
}

// IsInterfaceNil -
func (bcm *BootstrapComponentsMock) IsInterfaceNil() bool {
	return bcm == nil
}
