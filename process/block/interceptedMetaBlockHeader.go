package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// InterceptedHeader represents the wrapper over HeaderWrapper struct.
// It implements Newer and Hashed interfaces
type InterceptedMetaHeader struct {
	*block.MetaBlock
	multiSigVerifier    crypto.MultiSigVerifier
	chronologyValidator process.ChronologyValidator
	hash                []byte
}

// NewInterceptedHeader creates a new instance of InterceptedHeader struct
func NewInterceptedMetaHeader(
	multiSigVerifier crypto.MultiSigVerifier,
	chronologyValidator process.ChronologyValidator,
) *InterceptedMetaHeader {

	return &InterceptedMetaHeader{
		MetaBlock:           &block.MetaBlock{},
		multiSigVerifier:    multiSigVerifier,
		chronologyValidator: chronologyValidator,
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

	return imh.validityCheck()
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

func (imh *InterceptedMetaHeader) validityCheck() error {
	if imh.chronologyValidator == nil {
		return process.ErrNilChronologyValidator
	}

	return imh.chronologyValidator.ValidateReceivedBlock(
		sharding.MetachainShardId,
		imh.Epoch,
		imh.Nonce,
		imh.Round,
	)
}

// VerifySig verifies a signature
func (imh *InterceptedMetaHeader) VerifySig() error {
	// TODO: Check block signature after multisig will be implemented
	// TODO: the interceptors do not have access yet to consensus group selection to validate multisigs
	// TODO: verify that the block proposer is among the signers
	return nil
}
