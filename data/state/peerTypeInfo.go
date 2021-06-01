package state

import "github.com/ElrondNetwork/elrond-go/core"

// PeerTypeInfo contains information related to the peertypes needed by the peerTypeProvider
type PeerTypeInfo struct {
	PublicKey   string
	PeerType    string
	PeerSubType core.P2PPeerSubType
	ShardId     uint32
}
