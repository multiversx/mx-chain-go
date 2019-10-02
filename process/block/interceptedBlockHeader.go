package block

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// InterceptedHeader represents the wrapper over HeaderWrapper struct.
// It implements Newer and Hashed interfaces
type InterceptedHeader struct {
	*block.Header
	multiSigVerifier crypto.MultiSigVerifier
	hash             []byte
	nodesCoordinator sharding.NodesCoordinator
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
}

// NewInterceptedHeader creates a new instance of InterceptedHeader struct
func NewInterceptedHeader(
	multiSigVerifier crypto.MultiSigVerifier,
	nodesCoordinator sharding.NodesCoordinator,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
) *InterceptedHeader {

	return &InterceptedHeader{
		Header:           &block.Header{},
		multiSigVerifier: multiSigVerifier,
		nodesCoordinator: nodesCoordinator,
		marshalizer:      marshalizer,
		hasher:           hasher,
	}
}

// SetHash sets the hash of this hdr. The hash will also be the ID of this object
func (inHdr *InterceptedHeader) SetHash(hash []byte) {
	inHdr.hash = hash
}

// Hash gets the hash of this hdr
func (inHdr *InterceptedHeader) Hash() []byte {
	return inHdr.hash
}

// Shard returns the shard ID for which this hdr is addressed
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

// IntegrityAndValidity checks the integrity and validity of a block hdr wrapper
func (inHdr *InterceptedHeader) IntegrityAndValidity(coordinator sharding.Coordinator) error {
	err := inHdr.Integrity(coordinator)
	if err != nil {
		return err
	}

	return nil
}

// Integrity checks the integrity of the state block wrapper
func (inHdr *InterceptedHeader) Integrity(coordinator sharding.Coordinator) error {
	if coordinator == nil || coordinator.IsInterfaceNil() {
		return process.ErrNilShardCoordinator
	}
	if inHdr.Header == nil {
		return process.ErrNilBlockHeader
	}
	if inHdr.PubKeysBitmap == nil {
		return process.ErrNilPubKeysBitmap
	}
	if inHdr.ShardId >= coordinator.NumberOfShards() {
		return process.ErrInvalidShardId
	}
	if inHdr.PrevHash == nil {
		return process.ErrNilPreviousBlockHash
	}
	if inHdr.Signature == nil {
		return process.ErrNilSignature
	}
	if inHdr.RootHash == nil {
		return process.ErrNilRootHash
	}
	if inHdr.RandSeed == nil {
		return process.ErrNilRandSeed
	}
	if inHdr.PrevRandSeed == nil {
		return process.ErrNilPrevRandSeed
	}

	switch inHdr.BlockBodyType {
	case block.PeerBlock:
		return inHdr.validatePeerBlock()
	case block.StateBlock:
		return inHdr.validateStateBlock()
	case block.TxBlock:
		return inHdr.validateTxBlock()
	default:
		return process.ErrInvalidBlockBodyType
	}
}

// VerifySig verifies the intercepted Header block signature
func (inHdr *InterceptedHeader) VerifySig() error {
	randSeed := inHdr.GetPrevRandSeed()
	bitmap := inHdr.GetPubKeysBitmap()

	if len(bitmap) == 0 {
		return process.ErrNilPubKeysBitmap
	}

	if bitmap[0]&1 == 0 {
		return process.ErrBlockProposerSignatureMissing

	}

	consensusPubKeys, err := inHdr.nodesCoordinator.GetValidatorsPublicKeys(randSeed, inHdr.Round, inHdr.ShardId)
	if err != nil {
		return err
	}

	verifier, err := inHdr.multiSigVerifier.Create(consensusPubKeys, 0)
	if err != nil {
		return err
	}

	err = verifier.SetAggregatedSig(inHdr.Signature)
	if err != nil {
		return err
	}

	// get marshalled block hdr without signature and bitmap
	// as this is the message that was signed
	headerCopy := *inHdr.Header
	headerCopy.Signature = nil
	headerCopy.PubKeysBitmap = nil

	hash, err := core.CalculateHash(inHdr.marshalizer, inHdr.hasher, headerCopy)
	if err != nil {
		return err
	}

	err = verifier.Verify(hash, bitmap)

	return err
}

func (inHdr *InterceptedHeader) validatePeerBlock() error {
	return nil
}

func (inHdr *InterceptedHeader) validateStateBlock() error {
	return nil
}

func (inHdr *InterceptedHeader) validateTxBlock() error {
	if inHdr.MiniBlockHeaders == nil {
		return process.ErrNilMiniBlockHeaders
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (inHdr *InterceptedHeader) IsInterfaceNil() bool {
	if inHdr == nil {
		return true
	}
	return false
}
