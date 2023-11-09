package state

import "github.com/multiversx/mx-chain-core-go/core"

// PeerTypeInfo contains information related to the peertypes needed by the peerTypeProvider
type PeerTypeInfo struct {
	PublicKey   string
	PeerType    string
	PeerSubType core.P2PPeerSubType
	ShardId     uint32
}
