package libp2p

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type unknownPeerShardResolver struct {
}

// GetShardID returns the sharding.UnknownShardId value
func (upsr *unknownPeerShardResolver) GetShardID(_ p2p.PeerID) uint32 {
	return core.UnknownShardId
}

// IsInterfaceNil returns true if there is no value under the interface
func (upsr *unknownPeerShardResolver) IsInterfaceNil() bool {
	return upsr == nil
}
