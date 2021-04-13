package block

import "github.com/ElrondNetwork/elrond-go/data"

// SetPubKey - setter for public key
func (pc *PeerChange) SetPubKey(pubKey []byte) error {
	if pc == nil {
		return data.ErrNilPointerReceiver
	}

	pc.PubKey = pubKey

	return nil
}

// SetShardIdDest - setter for destination shardID
func (pc *PeerChange) SetShardIdDest(shardID uint32) error {
	if pc == nil {
		return data.ErrNilPointerReceiver
	}

	pc.ShardIdDest = shardID

	return nil
}
