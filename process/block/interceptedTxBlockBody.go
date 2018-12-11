package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

// InterceptedTxBlockBody contains data found in a TxBlockBodyType
// essentially is a slice of miniblocks
type InterceptedTxBlockBody struct {
	*block.TxBlockBody
	hash []byte
}

// NewInterceptedTxBlockBody creates a new instance of InterceptedTxBlockBody struct
func NewInterceptedTxBlockBody() *InterceptedTxBlockBody {
	return &InterceptedTxBlockBody{TxBlockBody: &block.TxBlockBody{}}
}

// Check returns true if the tx block body data pass some sanity check and some validation checks
func (inTxBlkBdy *InterceptedTxBlockBody) Check() bool {
	notNilFields := inTxBlkBdy.TxBlockBody != nil &&
		inTxBlkBdy.MiniBlocks != nil &&
		len(inTxBlkBdy.MiniBlocks) != 0

	if !notNilFields {
		log.Debug("tx block body with nil fields")
		return false
	}

	return inTxBlkBdy.checkEachMiniBlock()
}

func (inTxBlkBdy *InterceptedTxBlockBody) checkEachMiniBlock() bool {
	for i := 0; i < len(inTxBlkBdy.MiniBlocks); i++ {
		miniBlock := inTxBlkBdy.MiniBlocks[i]

		if len(miniBlock.TxHashes) == 0 {
			return false
		}
	}

	return true
}

// SetHash sets the hash of this transaction block body. The hash will also be the ID of this object
func (inTxBlkBdy *InterceptedTxBlockBody) SetHash(hash []byte) {
	inTxBlkBdy.hash = hash
}

// Hash gets the hash of this transaction block body
func (inTxBlkBdy *InterceptedTxBlockBody) Hash() []byte {
	return inTxBlkBdy.hash
}

// New returns a new instance of this struct (used in topics)
func (inTxBlkBdy *InterceptedTxBlockBody) New() p2p.Newer {
	return NewInterceptedTxBlockBody()
}

// ID returns the ID of this object. Set to return the hash of the transaction block body
func (inTxBlkBdy *InterceptedTxBlockBody) ID() string {
	return string(inTxBlkBdy.hash)
}

// Shard returns the shard ID for which this body is addressed
func (inTxBlkBdy *InterceptedTxBlockBody) Shard() uint32 {
	return inTxBlkBdy.ShardId
}
