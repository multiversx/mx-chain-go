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
type InterceptedMetaHeader struct {
	*block.MetaBlock
	multiSigVerifier crypto.MultiSigVerifier
	hash             []byte
	nodesCoordinator sharding.NodesCoordinator
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
}

// NewInterceptedHeader creates a new instance of InterceptedHeader struct
func NewInterceptedMetaHeader(
	multiSigVerifier crypto.MultiSigVerifier,
	nodesCoordinator sharding.NodesCoordinator,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
) *InterceptedMetaHeader {

	return &InterceptedMetaHeader{
		MetaBlock:        &block.MetaBlock{},
		multiSigVerifier: multiSigVerifier,
		nodesCoordinator: nodesCoordinator,
		marshalizer:      marshalizer,
		hasher:           hasher,
	}
}

// SetHash sets the hash of this header. The hash will also be the ID of this object
func (imh *InterceptedMetaHeader) SetHash(hash []byte) {
	imh.hash = hash
}

// Hash gets the hash of this header
func (imh *InterceptedMetaHeader) Hash() []byte {
	return imh.hash
}

// GetMetaHeader returns the MetaBlock pointer that holds the data
func (imh *InterceptedMetaHeader) GetMetaHeader() *block.MetaBlock {
	return imh.MetaBlock
}

// IntegrityAndValidity checks the integrity and validity of a block header wrapper
func (imh *InterceptedMetaHeader) IntegrityAndValidity(coordinator sharding.Coordinator) error {
	err := imh.Integrity(coordinator)
	if err != nil {
		return err
	}

	return nil
}

// Integrity checks the integrity of the state block wrapper
func (imh *InterceptedMetaHeader) Integrity(coordinator sharding.Coordinator) error {
	if coordinator == nil {
		return process.ErrNilShardCoordinator
	}
	if imh.MetaBlock == nil {
		return process.ErrNilMetaBlockHeader
	}
	if imh.PubKeysBitmap == nil {
		return process.ErrNilPubKeysBitmap
	}
	if imh.PrevHash == nil {
		return process.ErrNilPreviousBlockHash
	}
	if imh.Signature == nil {
		return process.ErrNilSignature
	}
	if imh.RootHash == nil {
		return process.ErrNilRootHash
	}
	if imh.RandSeed == nil {
		return process.ErrNilRandSeed
	}
	if imh.PrevRandSeed == nil {
		return process.ErrNilPrevRandSeed
	}

	for _, sd := range imh.ShardInfo {
		if sd.ShardId >= coordinator.NumberOfShards() {
			return process.ErrInvalidShardId
		}
		for _, smbh := range sd.ShardMiniBlockHeaders {
			if smbh.SenderShardId >= coordinator.NumberOfShards() {
				return process.ErrInvalidShardId
			}
			if smbh.ReceiverShardId >= coordinator.NumberOfShards() {
				return process.ErrInvalidShardId
			}
		}
	}

	return nil
}

// VerifySig verifies a signature
func (imh *InterceptedMetaHeader) VerifySig() error {
	randSeed := imh.GetPrevRandSeed()
	bitmap := imh.GetPubKeysBitmap()

	if len(bitmap) == 0 {
		return process.ErrNilPubKeysBitmap
	}

	if bitmap[0]&1 == 0 {
		return process.ErrBlockProposerSignatureMissing

	}

	consensusPubKeys, err := imh.nodesCoordinator.GetValidatorsPublicKeys(randSeed)
	if err != nil {
		return err
	}

	verifier, err := imh.multiSigVerifier.Create(consensusPubKeys, 0)
	if err != nil {
		return err
	}

	err = verifier.SetAggregatedSig(imh.Signature)
	if err != nil {
		return err
	}

	// get marshalled block header without signature and bitmap
	// as this is the message that was signed
	headerCopy := *imh.MetaBlock
	headerCopy.Signature = nil
	headerCopy.PubKeysBitmap = nil

	hash, err := core.CalculateHash(imh.marshalizer, imh.hasher, headerCopy)
	if err != nil {
		return err
	}

	err = verifier.Verify(hash, bitmap)

	return err
}

// IsInterfaceNil return if there is no value under the interface
func (mb *InterceptedMetaHeader) IsInterfaceNil() bool {
	if mb == nil {
		return true
	}
	return false
}
