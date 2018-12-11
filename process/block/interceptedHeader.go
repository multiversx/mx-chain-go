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
