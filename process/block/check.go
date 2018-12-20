package block

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// StateBlockBodyWrapper is a wrapper for StateBlockBody, adding functionality for validity and integrity checks
type StateBlockBodyWrapper struct {
	*block.StateBlockBody
	Processor process.BlockProcessor
}

// PeerBlockBodyWrapper is a wrapper for StateBlockBody, adding functionality for validity and integrity checks
type PeerBlockBodyWrapper struct {
	*block.PeerBlockBody
	Processor process.BlockProcessor
}

// TxBlockBodyWrapper is a wrapper for StateBlockBody, adding functionality for validity and integrity checks
type TxBlockBodyWrapper struct {
	*block.TxBlockBody
	Processor process.BlockProcessor
}

// HeaderWrapper is a wrapper for StateBlockBody, adding functionality for validity and integrity checks
// and as well for signature verification
type HeaderWrapper struct {
	*block.Header
	Processor process.BlockProcessor
}

// Check checks the integrity and validity of a state block
func (sbWrapper StateBlockBodyWrapper) Check() error {
	err := sbWrapper.Integrity()

	if err != nil {
		return err
	}

	return sbWrapper.validityCheck()
}

// Integrity checks the integrity of the state block
func (sbWrapper StateBlockBodyWrapper) Integrity() error {
	if sbWrapper.Processor == nil {
		return process.ErrNilProcessor
	}

	if sbWrapper.StateBlockBody == nil {
		return process.ErrNilStateBlockBody
	}

	if sbWrapper.ShardID >= sbWrapper.Processor.NoShards() {
		return process.ErrInvalidShardId
	}

	if sbWrapper.RootHash == nil {
		return process.ErrNilRootHash
	}

	return nil
}

func (sbWrapper StateBlockBodyWrapper) validityCheck() error {
	if sbWrapper.Processor.GetRootHash() == nil {
		return process.ErrNilRootHash
	}

	if !bytes.Equal(sbWrapper.Processor.GetRootHash(), sbWrapper.RootHash) {
		return process.ErrInvalidRootHash
	}

	return nil
}

// Check checks the integrity of a transactions block
func (txbWrapper TxBlockBodyWrapper) Check() error {
	err := txbWrapper.Integrity()

	if err != nil {
		return err
	}

	return txbWrapper.validityCheck()
}

// Integrity checks the integrity of the state block wrapper
func (txbWrapper TxBlockBodyWrapper) Integrity() error {

	if txbWrapper.TxBlockBody == nil {
		return process.ErrNilTxBlockBody
	}

	stateBlockWrapper := StateBlockBodyWrapper{
		StateBlockBody: &txbWrapper.StateBlockBody,
		Processor:      txbWrapper.Processor,
	}

	err := stateBlockWrapper.Integrity()

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

// Check checks the integrity and validity of a peer block wrapper
func (pbWrapper PeerBlockBodyWrapper) Check() error {
	err := pbWrapper.Integrity()
	if err != nil {
		return err
	}

	return pbWrapper.validityCheck()
}

// Integrity checks the integrity of the state block wrapper
func (pbWrapper PeerBlockBodyWrapper) Integrity() error {
	if pbWrapper.PeerBlockBody == nil {
		return process.ErrNilPeerBlockBody
	}

	stateBlockWrapper := StateBlockBodyWrapper{
		StateBlockBody: &pbWrapper.StateBlockBody,
		Processor:      pbWrapper.Processor,
	}

	err := stateBlockWrapper.Integrity()

	if err != nil {
		return err
	}

	if pbWrapper.Changes == nil {
		return process.ErrNilPeerChanges
	}

	for _, change := range pbWrapper.Changes {
		if change.ShardIdDest >= pbWrapper.Processor.NoShards() {
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

// Check checks the integrity and validity of a block header wrapper
func (hWrapper HeaderWrapper) Check() error {
	err := hWrapper.Integrity()
	if err != nil {
		return err
	}

	return hWrapper.validityCheck()
}

// Integrity checks the integrity of the state block wrapper
func (hWrapper HeaderWrapper) Integrity() error {
	if hWrapper.Processor == nil {
		return process.ErrNilProcessor
	}

	if hWrapper.Header == nil {
		return process.ErrNilProcessor
	}

	if hWrapper.BlockBodyHash == nil {
		return process.ErrNilBlockBodyHash
	}

	if hWrapper.PubKeysBitmap == nil {
		return process.ErrNilPubKeysBitmap
	}

	if hWrapper.ShardId >= hWrapper.Processor.NoShards() {
		return process.ErrInvalidShardId
	}

	if hWrapper.PrevHash == nil {
		return process.ErrNilPreviousBlockHash
	}

	if hWrapper.Signature == nil {
		return process.ErrNilSignature
	}

	return nil
}

func (hWrapper HeaderWrapper) validityCheck() error {
	// TODO: need to check epoch is round - timestamp - epoch - nonce - requires chronology

	return nil
}

// VerifySig verifies a signature
func (hWrapper HeaderWrapper) VerifySig() error {
	if hWrapper.Header == nil {
		return process.ErrNilBlockHeader
	}

	if hWrapper.Header.Signature == nil {
		return process.ErrNilSignature
	}
	// TODO: Check block signature after multisig will be implemented
	return nil
}
