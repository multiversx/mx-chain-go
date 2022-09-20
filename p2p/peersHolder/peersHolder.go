package peersHolder

import (
	"github.com/ElrondNetwork/elrond-go-p2p/peersHolder"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// NewPeersHolder returns a new instance of peersHolder
func NewPeersHolder(preferredConnectionAddresses []string) (p2p.PreferredPeersHolderHandler, error) {
	return peersHolder.NewPeersHolder(preferredConnectionAddresses)
}
