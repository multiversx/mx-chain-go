package peer

type PeerTypeInfo struct {
	PublicKey string
	PeerType  string
	ShardId   uint32
}

// GetPublicKey returns the PK
func (pti *PeerTypeInfo) GetPublicKey() string {
	return pti.PublicKey
}

// GetPeerType returns the peer type
func (pti *PeerTypeInfo) GetPeerType() string {
	return pti.PeerType
}

// GetShardId returns the shard id
func (pti *PeerTypeInfo) GetShardId() uint32 {
	return pti.ShardId
}
