package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// InterceptedPeerBlockBody represents the wrapper over PeerBlockBodyWrapper struct.
type InterceptedPeerBlockBody struct {
	*block.PeerBlockBody
	hash []byte
}

// NewInterceptedPeerBlockBody creates a new instance of InterceptedPeerBlockBody struct
func NewInterceptedPeerBlockBody() *InterceptedPeerBlockBody {
	return &InterceptedPeerBlockBody{
		PeerBlockBody: &block.PeerBlockBody{},
	}
}

// SetHash sets the hash of this peer block body. The hash will also be the ID of this object
func (inPeerBlkBdy *InterceptedPeerBlockBody) SetHash(hash []byte) {
	inPeerBlkBdy.hash = hash
}

// Hash gets the hash of this peer block body
func (inPeerBlkBdy *InterceptedPeerBlockBody) Hash() []byte {
	return inPeerBlkBdy.hash
}

// Create returns a new instance of this struct (used in topics)
func (inPeerBlkBdy *InterceptedPeerBlockBody) Create() p2p.Creator {
	return NewInterceptedPeerBlockBody()
}

// ID returns the ID of this object. Set to return the hash of the peer block body
func (inPeerBlkBdy *InterceptedPeerBlockBody) ID() string {
	return string(inPeerBlkBdy.hash)
}

// Shard returns the shard ID for which this body is addressed
func (inPeerBlkBdy *InterceptedPeerBlockBody) Shard() uint32 {
	return inPeerBlkBdy.ShardID
}

// GetUnderlyingObject returns the underlying object
func (inPeerBlkBdy *InterceptedPeerBlockBody) GetUnderlyingObject() interface{} {
	return inPeerBlkBdy.PeerBlockBody
}

// IntegrityAndValidity checks the integrity and validity of a peer block wrapper
func (inPeerBlkBdy *InterceptedPeerBlockBody) IntegrityAndValidity(coordinator sharding.ShardCoordinator) error {
	err := inPeerBlkBdy.Integrity(coordinator)
	if err != nil {
		return err
	}

	return inPeerBlkBdy.validityCheck()
}

// Integrity checks the integrity of the state block wrapper
func (inPeerBlkBdy *InterceptedPeerBlockBody) Integrity(coordinator sharding.ShardCoordinator) error {
	if coordinator == nil {
		return process.ErrNilShardCoordinator
	}

	if inPeerBlkBdy.PeerBlockBody == nil {
		return process.ErrNilPeerBlockBody
	}

	interceptedStateBlock := InterceptedStateBlockBody{
		StateBlockBody: &inPeerBlkBdy.StateBlockBody,
	}

	err := interceptedStateBlock.Integrity(coordinator)
	if err != nil {
		return err
	}

	if inPeerBlkBdy.Changes == nil {
		return process.ErrNilPeerChanges
	}

	for _, change := range inPeerBlkBdy.Changes {
		if change.ShardIdDest >= coordinator.NoShards() {
			return process.ErrInvalidShardId
		}

		if change.PubKey == nil {
			return process.ErrNilPublicKey
		}
	}

	return nil
}

func (inPeerBlkBdy *InterceptedPeerBlockBody) validityCheck() error {
	// TODO: check that the peer changes received are equal with what has been calculated

	return nil
}
