package state

// PeerTypeInfo contains information related to the peertypes needed by the peerTypeProvider
type PeerTypeInfo struct {
	PublicKey string
	PeerType  string
	ShardId   uint32
}
