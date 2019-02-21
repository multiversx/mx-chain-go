package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// InterceptedStateBlockBody represents the wrapper over InterceptedStateBlockBody struct.
type InterceptedStateBlockBody struct {
	*block.StateBlockBody
	hash []byte
}

// NewInterceptedStateBlockBody creates a new instance of InterceptedStateBlockBody struct
func NewInterceptedStateBlockBody() *InterceptedStateBlockBody {
	return &InterceptedStateBlockBody{
		StateBlockBody: &block.StateBlockBody{},
	}
}

// SetHash sets the hash of this state block body. The hash will also be the ID of this object
func (inStateBlkBdy *InterceptedStateBlockBody) SetHash(hash []byte) {
	inStateBlkBdy.hash = hash
}

// Hash gets the hash of this state block body
func (inStateBlkBdy *InterceptedStateBlockBody) Hash() []byte {
	return inStateBlkBdy.hash
}

// Create returns a new instance of this struct (used in topics)
func (inStateBlkBdy *InterceptedStateBlockBody) Create() p2p.Creator {
	return NewInterceptedStateBlockBody()
}

// ID returns the ID of this object. Set to return the hash of the state block body
func (inStateBlkBdy *InterceptedStateBlockBody) ID() string {
	return string(inStateBlkBdy.hash)
}

// Shard returns the shard ID for which this body is addressed
func (inStateBlkBdy *InterceptedStateBlockBody) Shard() uint32 {
	return inStateBlkBdy.ShardID
}

// GetUnderlyingObject returns the underlying object
func (inStateBlkBdy *InterceptedStateBlockBody) GetUnderlyingObject() interface{} {
	return inStateBlkBdy.StateBlockBody
}

// IntegrityAndValidity checks the integrity and validity of a state block wrapper
func (inStateBlkBdy *InterceptedStateBlockBody) IntegrityAndValidity(coordinator sharding.ShardCoordinator) error {
	err := inStateBlkBdy.Integrity(coordinator)

	if err != nil {
		return err
	}

	return inStateBlkBdy.validityCheck(coordinator)
}

// Integrity checks the integrity of the state block
func (inStateBlkBdy *InterceptedStateBlockBody) Integrity(coordinator sharding.ShardCoordinator) error {
	if coordinator == nil {
		return process.ErrNilShardCoordinator
	}

	if inStateBlkBdy.StateBlockBody == nil {
		return process.ErrNilStateBlockBody
	}

	if inStateBlkBdy.ShardID >= coordinator.NoShards() {
		return process.ErrInvalidShardId
	}

	if inStateBlkBdy.RootHash == nil {
		return process.ErrNilRootHash
	}

	return nil
}

func (inStateBlkBdy *InterceptedStateBlockBody) validityCheck(coordinator sharding.ShardCoordinator) error {
	return nil
}
