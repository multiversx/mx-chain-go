package block

import "github.com/ElrondNetwork/elrond-go/data"

// SetPubKey - setter for public key
func (m *PeerChange) SetPubKey(pubKey []byte) error {
	if m == nil {
		return data.ErrNilPointerReceiver
	}
	m.PubKey = pubKey
	return nil
}

// SetShardIdDest - setter for destination shardID
func (m *PeerChange) SetShardIdDest(shardID uint32) error {
	if m == nil {
		return data.ErrNilPointerReceiver
	}
	m.ShardIdDest = shardID
	return nil
}
