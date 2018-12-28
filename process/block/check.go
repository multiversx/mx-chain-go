package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// StateBlockBodyWrapper is a wrapper for StateBlockBody, adding functionality for validity and integrity checks
type StateBlockBodyWrapper struct {
	*block.StateBlockBody
}

// PeerBlockBodyWrapper is a wrapper for StateBlockBody, adding functionality for validity and integrity checks
type PeerBlockBodyWrapper struct {
	*block.PeerBlockBody
}

// TxBlockBodyWrapper is a wrapper for StateBlockBody, adding functionality for validity and integrity checks
type TxBlockBodyWrapper struct {
	*block.TxBlockBody
}

// HeaderWrapper is a wrapper for StateBlockBody, adding functionality for validity and integrity checks
// and as well for signature verification
type HeaderWrapper struct {
	*block.Header
}

// IntegrityAndValidity checks the integrity and validity of a state block wrapper
func (sbWrapper StateBlockBodyWrapper) IntegrityAndValidity(coordinator sharding.ShardCoordinator) error {
	err := sbWrapper.Integrity(coordinator)

	if err != nil {
		return err
	}

	return sbWrapper.validityCheck(coordinator)
}

// Integrity checks the integrity of the state block
func (sbWrapper StateBlockBodyWrapper) Integrity(coordinator sharding.ShardCoordinator) error {
	if coordinator == nil {
		return process.ErrNilShardCoordinator
	}

	if sbWrapper.StateBlockBody == nil {
		return process.ErrNilStateBlockBody
	}

	if sbWrapper.ShardID >= coordinator.NoShards() {
		return process.ErrInvalidShardId
	}

	if sbWrapper.RootHash == nil {
		return process.ErrNilRootHash
	}

	return nil
}

func (sbWrapper StateBlockBodyWrapper) validityCheck(coordinator sharding.ShardCoordinator) error {
	return nil
}

// IntegrityAndValidity checks the integrity of a transactions block
func (txbWrapper TxBlockBodyWrapper) IntegrityAndValidity(coordinator sharding.ShardCoordinator) error {
	err := txbWrapper.Integrity(coordinator)

	if err != nil {
		return err
	}

	return txbWrapper.validityCheck()
}

// Integrity checks the integrity of the state block wrapper
func (txbWrapper TxBlockBodyWrapper) Integrity(coordinator sharding.ShardCoordinator) error {
	if coordinator == nil {
		return process.ErrNilShardCoordinator
	}

	if txbWrapper.TxBlockBody == nil {
		return process.ErrNilTxBlockBody
	}

	stateBlockWrapper := StateBlockBodyWrapper{
		StateBlockBody: &txbWrapper.StateBlockBody,
	}

	err := stateBlockWrapper.Integrity(coordinator)
	if err != nil {
		return err
	}

	if txbWrapper.MiniBlocks == nil {
		return process.ErrNilMiniBlocks
	}

	for _, miniBlock := range txbWrapper.MiniBlocks {
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

func (txbWrapper TxBlockBodyWrapper) validityCheck() error {
	// TODO: update with validity checks

	return nil
}

// IntegrityAndValidity checks the integrity and validity of a peer block wrapper
func (pbWrapper PeerBlockBodyWrapper) IntegrityAndValidity(coordinator sharding.ShardCoordinator) error {
	err := pbWrapper.Integrity(coordinator)
	if err != nil {
		return err
	}

	return pbWrapper.validityCheck()
}

// Integrity checks the integrity of the state block wrapper
func (pbWrapper PeerBlockBodyWrapper) Integrity(coordinator sharding.ShardCoordinator) error {
	if coordinator == nil {
		return process.ErrNilShardCoordinator
	}

	if pbWrapper.PeerBlockBody == nil {
		return process.ErrNilPeerBlockBody
	}

	stateBlockWrapper := StateBlockBodyWrapper{
		StateBlockBody: &pbWrapper.StateBlockBody,
	}

	err := stateBlockWrapper.Integrity(coordinator)
	if err != nil {
		return err
	}

	if pbWrapper.Changes == nil {
		return process.ErrNilPeerChanges
	}

	for _, change := range pbWrapper.Changes {
		if change.ShardIdDest >= coordinator.NoShards() {
			return process.ErrInvalidShardId
		}

		if change.PubKey == nil {
			return process.ErrNilPublicKey
		}
	}

	return nil
}

func (pbWrapper PeerBlockBodyWrapper) validityCheck() error {
	// TODO: check that the peer changes received are equal with what has been calculated

	return nil
}

// IntegrityAndValidity checks the integrity and validity of a block header wrapper
func (hWrapper *HeaderWrapper) IntegrityAndValidity(coordinator sharding.ShardCoordinator) error {
	err := hWrapper.Integrity(coordinator)
	if err != nil {
		return err
	}

	return hWrapper.validityCheck()
}

// Integrity checks the integrity of the state block wrapper
func (hWrapper *HeaderWrapper) Integrity(coordinator sharding.ShardCoordinator) error {
	if coordinator == nil {
		return process.ErrNilShardCoordinator
	}

	if hWrapper.Header == nil {
		return process.ErrNilBlockHeader
	}

	if hWrapper.BlockBodyHash == nil {
		return process.ErrNilBlockBodyHash
	}

	if hWrapper.PubKeysBitmap == nil {
		return process.ErrNilPubKeysBitmap
	}

	if hWrapper.ShardId >= coordinator.NoShards() {
		return process.ErrInvalidShardId
	}

	if hWrapper.PrevHash == nil {
		return process.ErrNilPreviousBlockHash
	}

	if hWrapper.Signature == nil {
		return process.ErrNilSignature
	}

	if hWrapper.Commitment == nil {
		return process.ErrNilCommitment
	}

	switch hWrapper.BlockBodyType {
	case block.PeerBlock:
	case block.StateBlock:
	case block.TxBlock:
	default:
		return process.ErrInvalidBlockBodyType
	}

	return nil
}

func (hWrapper *HeaderWrapper) validityCheck() error {
	// TODO: need to check epoch is round - timestamp - epoch - nonce - requires chronology
	return nil
}

// VerifySig verifies a signature
func (hWrapper *HeaderWrapper) VerifySig() error {
	// TODO: Check block signature after multisig will be implemented
	return nil
}
