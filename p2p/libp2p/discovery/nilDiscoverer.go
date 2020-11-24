package discovery

import "github.com/ElrondNetwork/elrond-go/p2p"

var _ p2p.PeerDiscoverer = (*NilDiscoverer)(nil)
var _ p2p.Reconnecter = (*NilDiscoverer)(nil)

const nilName = "no peer discovery"

// NilDiscoverer is the non-functional peer discoverer aimed to be used when peer discovery options are all disabled
type NilDiscoverer struct {
}

// NewNilDiscoverer creates a new NullDiscoverer implementation
func NewNilDiscoverer() *NilDiscoverer {
	return &NilDiscoverer{}
}

// Bootstrap will return nil. There is no implementation.
func (nd *NilDiscoverer) Bootstrap() error {
	return nil
}

// Name returns a message which says no peer discovery mechanism is used
func (nd *NilDiscoverer) Name() string {
	return nilName
}

// ReconnectToNetwork does nothing
func (nd *NilDiscoverer) ReconnectToNetwork() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (nd *NilDiscoverer) IsInterfaceNil() bool {
	return nd == nil
}
