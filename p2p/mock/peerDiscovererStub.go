package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// PeerDiscovererStub -
type PeerDiscovererStub struct {
	BootstrapCalled    func() error
	CloseCalled        func() error
	ApplyContextCalled func(ctxProvider p2p.ContextProvider) error
}

// Bootstrap -
func (pds *PeerDiscovererStub) Bootstrap() error {
	return pds.BootstrapCalled()
}

// Name -
func (pds *PeerDiscovererStub) Name() string {
	return "PeerDiscovererStub"
}

// ApplyContext -
func (pds *PeerDiscovererStub) ApplyContext(ctxProvider p2p.ContextProvider) error {
	return pds.ApplyContextCalled(ctxProvider)
}

// IsInterfaceNil returns true if there is no value under the interface
func (pds *PeerDiscovererStub) IsInterfaceNil() bool {
	if pds == nil {
		return true
	}
	return false
}
