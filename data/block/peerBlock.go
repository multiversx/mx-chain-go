package block

// SetPubKey - setter for public key
func (m *PeerChange) SetPubKey(pubKey []byte) {
	m.PubKey = pubKey
}

// SetShardIdDest - setter for destination shardID
func (m *PeerChange) SetShardIdDest(shardID uint32) {
	m.ShardIdDest = shardID
}
