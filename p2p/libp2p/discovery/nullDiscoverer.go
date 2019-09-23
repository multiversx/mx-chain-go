package discovery

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

const nullName = "no peer discovery"

// NullDiscoverer is the non-functional peer discoverer aimed to be used when peer discovery options are all disabled
type NullDiscoverer struct {
}

// NewNullDiscoverer creates a new NullDiscoverer implementation
func NewNullDiscoverer() *NullDiscoverer {
	return &NullDiscoverer{}
}

// Bootstrap will return nil. There is no implementation.
func (nd *NullDiscoverer) Bootstrap() error {
	return nil
}

// Name returns a message which says no peer discovery mechanism is used
func (nd *NullDiscoverer) Name() string {
	return nullName
}

// ApplyContext is an empty func as the context is not required
func (nd *NullDiscoverer) ApplyContext(ctxProvider p2p.ContextProvider) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (nd *NullDiscoverer) IsInterfaceNil() bool {
	if nd == nil {
		return true
	}
	return false
}
