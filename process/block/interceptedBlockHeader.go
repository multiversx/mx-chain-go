package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// InterceptedHeader represents the wrapper over HeaderWrapper struct.
// It implements Newer and Hashed interfaces
type InterceptedHeader struct {
	*block.Header
	multiSigVerifier crypto.MultiSigVerifier
	hash             []byte
}

// NewInterceptedHeader creates a new instance of InterceptedHeader struct
func NewInterceptedHeader(multiSigVerifier crypto.MultiSigVerifier) *InterceptedHeader {
	return &InterceptedHeader{
		Header:           &block.Header{},
		multiSigVerifier: multiSigVerifier,
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
	return NewInterceptedHeader(inHdr.multiSigVerifier)
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

// IntegrityAndValidity checks the integrity and validity of a block header wrapper
func (inHdr *InterceptedHeader) IntegrityAndValidity(coordinator sharding.ShardCoordinator) error {
	err := inHdr.Integrity(coordinator)
	if err != nil {
		return err
	}

	return inHdr.validityCheck()
}

// Integrity checks the integrity of the state block wrapper
func (inHdr *InterceptedHeader) Integrity(coordinator sharding.ShardCoordinator) error {
	if coordinator == nil {
		return process.ErrNilShardCoordinator
	}

	if inHdr.Header == nil {
		return process.ErrNilBlockHeader
	}

	if inHdr.BlockBodyHash == nil {
		return process.ErrNilBlockBodyHash
	}

	if inHdr.PubKeysBitmap == nil {
		return process.ErrNilPubKeysBitmap
	}

	if inHdr.ShardId >= coordinator.NoShards() {
		return process.ErrInvalidShardId
	}

	if inHdr.PrevHash == nil {
		return process.ErrNilPreviousBlockHash
	}

	if inHdr.Signature == nil {
		return process.ErrNilSignature
	}

	if inHdr.Commitment == nil {
		return process.ErrNilCommitment
	}

	switch inHdr.BlockBodyType {
	case block.PeerBlock:
	case block.StateBlock:
	case block.TxBlock:
	default:
		return process.ErrInvalidBlockBodyType
	}

	return nil
}

func (inHdr *InterceptedHeader) validityCheck() error {
	// TODO: need to check epoch is round - timestamp - epoch - nonce - requires chronology
	return nil
}

// VerifySig verifies a signature
func (inHdr *InterceptedHeader) VerifySig() error {
	// TODO: Check block signature after multisig will be implemented
	// TODO: the interceptors do not have access yet to consensus group selection to validate multisigs

	return nil
}
