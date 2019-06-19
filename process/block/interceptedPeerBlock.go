package block

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// InterceptedPeerBlockBody represents the wrapper over PeerBlockBodyWrapper struct.
type InterceptedPeerBlockBody struct {
	PeerBlockBody []*block.PeerChange
	hash          []byte
}

// NewInterceptedPeerBlockBody creates a new instance of InterceptedPeerBlockBody struct
func NewInterceptedPeerBlockBody() *InterceptedPeerBlockBody {
	return &InterceptedPeerBlockBody{
		PeerBlockBody: make([]*block.PeerChange, 0),
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

// GetUnderlyingObject returns the underlying object
func (inPeerBlkBdy *InterceptedPeerBlockBody) GetUnderlyingObject() interface{} {
	return inPeerBlkBdy.PeerBlockBody
}

// IntegrityAndValidity checks the integrity and validity of a peer block wrapper
func (inPeerBlkBdy *InterceptedPeerBlockBody) IntegrityAndValidity(coordinator sharding.Coordinator) error {
	err := inPeerBlkBdy.Integrity(coordinator)
	if err != nil {
		return err
	}

	return inPeerBlkBdy.validityCheck()
}

// Integrity checks the integrity of the state block wrapper
func (inPeerBlkBdy *InterceptedPeerBlockBody) Integrity(coordinator sharding.Coordinator) error {
	if coordinator == nil {
		return process.ErrNilShardCoordinator
	}

	if inPeerBlkBdy.PeerBlockBody == nil {
		return process.ErrNilPeerBlockBody
	}

	for _, change := range inPeerBlkBdy.PeerBlockBody {
		if change.ShardIdDest >= coordinator.NumberOfShards() {
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
