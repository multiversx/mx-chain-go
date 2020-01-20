package libp2p

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type unknownPeerShardResolver struct {
}

// ByID returns the sharding.UnknownShardId value
func (upsr *unknownPeerShardResolver) ByID(_ p2p.PeerID) uint32 {
	return sharding.UnknownShardId
}

// IsInterfaceNil returns true if there is no value under the interface
func (upsr *unknownPeerShardResolver) IsInterfaceNil() bool {
	return upsr == nil
}
