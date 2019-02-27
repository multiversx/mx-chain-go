package discovery

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

const nullName = "no peer discovery"

type NullDiscoverer struct {
}

func NewNullDiscoverer() *NullDiscoverer {
	return &NullDiscoverer{}
}

func (nd *NullDiscoverer) Bootstrap() error {
	return nil
}

func (nd *NullDiscoverer) Close() error {
	return nil
}

// Name returns the name of the mdns peer discovery implementation
func (nd *NullDiscoverer) Name() string {
	return nullName
}

func (nd *NullDiscoverer) ApplyContext(ctxProvider p2p.ContextProvider) error {
	return nil
}
