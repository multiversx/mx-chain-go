package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

// InterceptedStateBlockBody contains data found in a StateBlockBodyType
// essentially is the state root
type InterceptedStateBlockBody struct {
	*block.StateBlockBody
	hash []byte
}

// NewInterceptedStateBlockBody creates a new instance of InterceptedStateBlockBody struct
func NewInterceptedStateBlockBody() *InterceptedStateBlockBody {
	return &InterceptedStateBlockBody{StateBlockBody: &block.StateBlockBody{}}
}

// Check returns true if the state block body data pass some sanity check and some validation checks
func (inStateBlkBdy *InterceptedStateBlockBody) Check() bool {
	notNilFields := inStateBlkBdy.StateBlockBody != nil &&
		inStateBlkBdy.RootHash != nil &&
		len(inStateBlkBdy.RootHash) != 0

	if !notNilFields {
		log.Debug("state block body with nil fields")
		return false
	}

	return true
}

// SetHash sets the hash of this state block body. The hash will also be the ID of this object
func (inStateBlkBdy *InterceptedStateBlockBody) SetHash(hash []byte) {
	inStateBlkBdy.hash = hash
}

// Hash gets the hash of this state block body
func (inStateBlkBdy *InterceptedStateBlockBody) Hash() []byte {
	return inStateBlkBdy.hash
}

// New returns a new instance of this struct (used in topics)
func (inStateBlkBdy *InterceptedStateBlockBody) New() p2p.Newer {
	return NewInterceptedStateBlockBody()
}

// ID returns the ID of this object. Set to return the hash of the state block body
func (inStateBlkBdy *InterceptedStateBlockBody) ID() string {
	return string(inStateBlkBdy.hash)
}

// Shard returns the shard ID for which this body is addressed
func (inStateBlkBdy *InterceptedStateBlockBody) Shard() uint32 {
	return inStateBlkBdy.ShardId
}
