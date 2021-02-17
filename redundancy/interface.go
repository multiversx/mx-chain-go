package redundancy

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// NodeRedundancyHandler provides functionality to handle the redundancy mechanism for a node
type NodeRedundancyHandler interface {
	IsRedundancyNode() bool
	IsMasterMachineActive() bool
	AdjustInactivityIfNeeded(selfPubKey string, consensusPubKeys []string, roundIndex int64)
	ResetInactivityIfNeeded(selfPubKey string, consensusMsgPubKey string, consensusMsgPeerID core.PeerID)
	IsInterfaceNil() bool
}

// P2PMessenger defines a subset of the p2p.Messenger interface
type P2PMessenger interface {
	ID() core.PeerID
	IsInterfaceNil() bool
}
