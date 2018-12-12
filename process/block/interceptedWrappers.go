package block

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

// InterceptedHeader represents the wrapper over block.Header struct.
// It implements Newer, Checker and SigVerifier interfaces
type InterceptedHeader struct {
	*block.Header
	hash []byte
}

// InterceptedPeerBlockBody contains data found in a PeerBlockBodyType
// essentially is the changing list of peers on a shard
type InterceptedPeerBlockBody struct {
	*block.PeerBlockBody
	hash []byte
}

// InterceptedStateBlockBody contains data found in a StateBlockBodyType
// essentially is the state root
type InterceptedStateBlockBody struct {
	*block.StateBlockBody
	hash []byte
}

// InterceptedTxBlockBody contains data found in a TxBlockBodyType
// essentially is a slice of miniblocks
type InterceptedTxBlockBody struct {
	*block.TxBlockBody
	hash []byte
}

//------- InterceptedHeader

// NewInterceptedHeader creates a new instance of InterceptedHeader struct
func NewInterceptedHeader() *InterceptedHeader {
	return &InterceptedHeader{
		Header: &block.Header{},
	}
}

// Check returns true if the transaction data pass some sanity check and some validation checks
func (inHdr *InterceptedHeader) Check() bool {
	if !inHdr.sanityCheck() {
		return false
	}

	return true
}

func (inHdr *InterceptedHeader) sanityCheck() bool {
	notNilFields := inHdr.Header != nil &&
		inHdr.PrevHash != nil &&
		inHdr.PubKeysBitmap != nil &&
		inHdr.BlockBodyHash != nil &&
		inHdr.Signature != nil &&
		inHdr.Commitment != nil &&
		inHdr.RootHash != nil

	if !notNilFields {
		log.Debug("hdr with nil fields")
		return false
	}

	switch inHdr.BlockBodyType {
	case block.BlockBodyPeer:
	case block.BlockBodyState:
	case block.BlockBodyTx:
	default:
		log.Debug(fmt.Sprintf("invalid BlockBodyType: %v", inHdr.BlockBodyType))
		return false
	}

	return true
}

// VerifySig checks if the header is correctly signed
func (inHdr *InterceptedHeader) VerifySig() bool {
	//TODO add real sig verify here
	return true
}

// SetHash sets the hash of this header. The hash will also be the ID of this object
func (inHdr *InterceptedHeader) SetHash(hash []byte) {
	inHdr.hash = hash
}

// Hash gets the hash of this header
func (inHdr *InterceptedHeader) Hash() []byte {
	return inHdr.hash
}

// New returns a new instance of this struct (used in topics)
func (inHdr *InterceptedHeader) New() p2p.Newer {
	return NewInterceptedHeader()
}

// ID returns the ID of this object. Set to return the hash of the header
func (inHdr *InterceptedHeader) ID() string {
	return string(inHdr.hash)
}

// Shard returns the shard ID for which this header is addressed
func (inHdr *InterceptedHeader) Shard() uint32 {
	return inHdr.ShardId
}

// GetHeader returns the Header pointer that holds the data
func (inHdr *InterceptedHeader) GetHeader() *block.Header {
	return inHdr.Header
}

//------- InterceptedPeerBlockBody

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

//------- InterceptedStateBlockBody

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

//------- InterceptedTxBlockBody

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
