package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

// InterceptedHeader represents the wrapper over HeaderWrapper struct.
// It implements Newer and Hashed interfaces
type InterceptedHeader struct {
	*HeaderWrapper
	hash []byte
}

// InterceptedPeerBlockBody represents the wrapper over PeerBlockBodyWrapper struct.
type InterceptedPeerBlockBody struct {
	*PeerBlockBodyWrapper
	hash []byte
}

// InterceptedStateBlockBody represents the wrapper over InterceptedStateBlockBody struct.
type InterceptedStateBlockBody struct {
	*StateBlockBodyWrapper
	hash []byte
}

// InterceptedTxBlockBody represents the wrapper over TxBlockBodyWrapper struct.
type InterceptedTxBlockBody struct {
	*TxBlockBodyWrapper
	hash []byte
}

//------- InterceptedHeader

// NewInterceptedHeader creates a new instance of InterceptedHeader struct
func NewInterceptedHeader() *InterceptedHeader {
	return &InterceptedHeader{
		HeaderWrapper: &HeaderWrapper{Header: &block.Header{}},
	}
}

// SetHash sets the hash of this header. The hash will also be the ID of this object
func (inHdr *InterceptedHeader) SetHash(hash []byte) {
	inHdr.hash = hash
}

// Hash gets the hash of this header
func (inHdr *InterceptedHeader) Hash() []byte {
	return inHdr.hash
}

// Create returns a new instance of this struct (used in topics)
func (inHdr *InterceptedHeader) Create() p2p.Creator {
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

// GetUnderlyingObject returns the underlying object
func (inHdr *InterceptedHeader) GetUnderlyingObject() interface{} {
	return inHdr.Header
}

//------- InterceptedPeerBlockBody

// NewInterceptedPeerBlockBody creates a new instance of InterceptedPeerBlockBody struct
func NewInterceptedPeerBlockBody() *InterceptedPeerBlockBody {
	return &InterceptedPeerBlockBody{
		PeerBlockBodyWrapper: &PeerBlockBodyWrapper{PeerBlockBody: &block.PeerBlockBody{}},
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

//------- InterceptedStateBlockBody

// NewInterceptedStateBlockBody creates a new instance of InterceptedStateBlockBody struct
func NewInterceptedStateBlockBody() *InterceptedStateBlockBody {
	return &InterceptedStateBlockBody{
		StateBlockBodyWrapper: &StateBlockBodyWrapper{StateBlockBody: &block.StateBlockBody{}},
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

//------- InterceptedTxBlockBody

// NewInterceptedTxBlockBody creates a new instance of InterceptedTxBlockBody struct
func NewInterceptedTxBlockBody() *InterceptedTxBlockBody {
	return &InterceptedTxBlockBody{
		TxBlockBodyWrapper: &TxBlockBodyWrapper{TxBlockBody: &block.TxBlockBody{}},
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
