package redundancy

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// P2PMessenger defines a subset of the p2p.Messenger interface
type P2PMessenger interface {
	ID() core.PeerID
	IsInterfaceNil() bool
}
