package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// InterceptedTxBlockBody represents the wrapper over TxBlockBodyWrapper struct.
type InterceptedTxBlockBody struct {
	*block.TxBlockBody
	hash []byte
}

// NewInterceptedTxBlockBody creates a new instance of InterceptedTxBlockBody struct
func NewInterceptedTxBlockBody() *InterceptedTxBlockBody {
	return &InterceptedTxBlockBody{
		TxBlockBody: &block.TxBlockBody{},
	}
}

// SetHash sets the hash of this transaction block body. The hash will also be the ID of this object
func (inTxBlkBdy *InterceptedTxBlockBody) SetHash(hash []byte) {
	inTxBlkBdy.hash = hash
}

// Hash gets the hash of this transaction block body
func (inTxBlkBdy *InterceptedTxBlockBody) Hash() []byte {
	return inTxBlkBdy.hash
}

// Create returns a new instance of this struct (used in topics)
func (inTxBlkBdy *InterceptedTxBlockBody) Create() p2p.Creator {
	return NewInterceptedTxBlockBody()
}

// ID returns the ID of this object. Set to return the hash of the transaction block body
func (inTxBlkBdy *InterceptedTxBlockBody) ID() string {
	return string(inTxBlkBdy.hash)
}

// Shard returns the shard ID for which this body is addressed
func (inTxBlkBdy *InterceptedTxBlockBody) Shard() uint32 {
	return inTxBlkBdy.ShardID
}

// GetUnderlyingObject returns the underlying object
func (inTxBlkBdy *InterceptedTxBlockBody) GetUnderlyingObject() interface{} {
	return inTxBlkBdy.TxBlockBody
}

// IntegrityAndValidity checks the integrity of a transactions block
func (inTxBlkBdy *InterceptedTxBlockBody) IntegrityAndValidity(coordinator sharding.ShardCoordinator) error {
	err := inTxBlkBdy.Integrity(coordinator)

	if err != nil {
		return err
	}

	return inTxBlkBdy.validityCheck()
}

// Integrity checks the integrity of the state block wrapper
func (inTxBlkBdy *InterceptedTxBlockBody) Integrity(coordinator sharding.ShardCoordinator) error {
	if coordinator == nil {
		return process.ErrNilShardCoordinator
	}

	if inTxBlkBdy.TxBlockBody == nil {
		return process.ErrNilTxBlockBody
	}

	interceptedStateBlockBody := InterceptedStateBlockBody{
		StateBlockBody: &inTxBlkBdy.StateBlockBody,
	}

	err := interceptedStateBlockBody.Integrity(coordinator)
	if err != nil {
		return err
	}

	if inTxBlkBdy.MiniBlocks == nil {
		return process.ErrNilMiniBlocks
	}

	for _, miniBlock := range inTxBlkBdy.MiniBlocks {
		if miniBlock.TxHashes == nil {
			return process.ErrNilTxHashes
		}

		if miniBlock.ShardID >= coordinator.NoShards() {
			return process.ErrInvalidShardId
		}

		for _, txHash := range miniBlock.TxHashes {
			if txHash == nil {
				return process.ErrNilTxHash
			}
		}
	}

	return nil
}

func (inTxBlkBdy *InterceptedTxBlockBody) validityCheck() error {
	// TODO: update with validity checks

	return nil
}
