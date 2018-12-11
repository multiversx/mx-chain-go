package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

// InterceptedPeerBlockBody contains data found in a PeerBlockBodyType
// essentially is the changing list of peers on a shard
type InterceptedPeerBlockBody struct {
	*block.PeerBlockBody
	hash []byte
}

// NewInterceptedPeerBlockBody creates a new instance of InterceptedPeerBlockBody struct
func NewInterceptedPeerBlockBody() *InterceptedPeerBlockBody {
	return &InterceptedPeerBlockBody{PeerBlockBody: &block.PeerBlockBody{}}
}

// Check returns true if the peer block body data pass some sanity check and some validation checks
func (inPeerBlkBdy *InterceptedPeerBlockBody) Check() bool {
	notNilFields := inPeerBlkBdy.PeerBlockBody != nil &&
		inPeerBlkBdy.Changes != nil &&
		len(inPeerBlkBdy.Changes) != 0

	if !notNilFields {
		log.Debug("peer block body with nil fields")
		return false
	}

	return inPeerBlkBdy.checkEachChange()
}

func (inPeerBlkBdy *InterceptedPeerBlockBody) checkEachChange() bool {
	for i := 0; i < len(inPeerBlkBdy.Changes); i++ {
		change := inPeerBlkBdy.Changes[i]

		if len(change.PubKey) == 0 {
			return false
		}
	}

	return true
}

// SetHash sets the hash of this peer block body. The hash will also be the ID of this object
func (inPeerBlkBdy *InterceptedPeerBlockBody) SetHash(hash []byte) {
	inPeerBlkBdy.hash = hash
}

// Hash gets the hash of this peer block body
func (inPeerBlkBdy *InterceptedPeerBlockBody) Hash() []byte {
	return inPeerBlkBdy.hash
}

// New returns a new instance of this struct (used in topics)
func (inPeerBlkBdy *InterceptedPeerBlockBody) New() p2p.Newer {
	return NewInterceptedPeerBlockBody()
}

// ID returns the ID of this object. Set to return the hash of the peer block body
func (inPeerBlkBdy *InterceptedPeerBlockBody) ID() string {
	return string(inPeerBlkBdy.hash)
}

// Shard returns the shard ID for which this body is addressed
func (inPeerBlkBdy *InterceptedPeerBlockBody) Shard() uint32 {
	return inPeerBlkBdy.ShardId
}
