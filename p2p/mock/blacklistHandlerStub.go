package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// BlacklistHandlerStub -
type BlacklistHandlerStub struct {
	HasCalled func(pid core.PeerID) bool
}

// Has -
func (bhs *BlacklistHandlerStub) Has(pid core.PeerID) bool {
	return bhs.HasCalled(pid)
}

// IsInterfaceNil -
func (bhs *BlacklistHandlerStub) IsInterfaceNil() bool {
	return bhs == nil
}
