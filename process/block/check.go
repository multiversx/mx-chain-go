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
func (sbWrapper StateBlockBodyWrapper) Check() bool {
	if !sbWrapper.Integrity() {
		return false
	}

	if !sbWrapper.validityCheck() {
		return false
	}

	return true
}

// Integrity checks the integrity of the state block
func (sbWrapper StateBlockBodyWrapper) Integrity() bool {
	if sbWrapper.Processor == nil {
		log.Error("state block Processor is nil")
		return false
	}

	if sbWrapper.StateBlockBody == nil {
		log.Debug("state block nil")
		return false
	}

	if sbWrapper.ShardID >= sbWrapper.Processor.GetNbShards() {
		log.Debug("state block invalid shard number")
		return false
	}

	if sbWrapper.RootHash == nil {
		log.Debug("state block root hash nil")
		return false
	}

	return true
}

func (sbWrapper StateBlockBodyWrapper) validityCheck() bool {
	if sbWrapper.Processor.GetRootHash() == nil {
		log.Error("state root hash nil")
		return false
	}

	return bytes.Equal(sbWrapper.Processor.GetRootHash(), sbWrapper.RootHash)
}

// Check checks the integrity of a transactions block
func (txbWrapper TxBlockBodyWrapper) Check() bool {
	if !txbWrapper.Integrity() {
		return false
	}

	if !txbWrapper.validityCheck() {
		return false
	}

	return true
}

// Integrity checks the integrity of the state block wrapper
func (txbWrapper TxBlockBodyWrapper) Integrity() bool {
	if txbWrapper.TxBlockBody == nil {
		log.Debug("transactions block body nil")
		return false
	}

	stateBlockWrapper := StateBlockBodyWrapper{
		StateBlockBody: &txbWrapper.StateBlockBody,
		Processor:      txbWrapper.Processor,
	}

	if !stateBlockWrapper.Integrity() {
		return false
	}

	if txbWrapper.MiniBlocks == nil {
		log.Debug("tx block miniblocks nil")
		return false
	}

	for _, miniBlock := range txbWrapper.MiniBlocks {
		if miniBlock.TxHashes == nil {
			log.Debug("tx block txHashes nil")
			return false
		}

		for _, txHash := range miniBlock.TxHashes {
			if txHash == nil {
				log.Debug("tx block tx hash is nil")
				return false
			}
		}
	}

	return true
}

func (txbWrapper TxBlockBodyWrapper) validityCheck() bool {
	// TODO: update with validity checks

	return true
}

// Check checks the integrity and validity of a peer block wrapper
func (pbWrapper PeerBlockBodyWrapper) Check() bool {
	if !pbWrapper.Integrity() {
		return false
	}

	if !pbWrapper.validityCheck() {
		return false
	}

	return true
}

// Integrity checks the integrity of the state block wrapper
func (pbWrapper PeerBlockBodyWrapper) Integrity() bool {
	if pbWrapper.PeerBlockBody == nil {
		log.Debug("peer block body nil")
		return false
	}

	stateBlockWrapper := StateBlockBodyWrapper{
		StateBlockBody: &pbWrapper.StateBlockBody,
		Processor:      pbWrapper.Processor,
	}

	if !stateBlockWrapper.Integrity() {
		return false
	}

	if pbWrapper.Changes == nil {
		log.Debug("peer block changes nil")
		return false
	}

	for _, change := range pbWrapper.Changes {
		if change.ShardIdDest >= pbWrapper.Processor.GetNbShards() {
			log.Debug("peer block change.shardId invalid")
			return false
		}

		if change.PubKey == nil {
			log.Debug("peer block change.pubkey nil")
			return false
		}
	}

	return true
}

func (pbWrapper PeerBlockBodyWrapper) validityCheck() bool {
	// TODO: check that the peer changes received are equal with what has been calculated

	return true
}

// Check checks the integrity and validity of a block header wrapper
func (hWrapper HeaderWrapper) Check() bool {
	if !hWrapper.Integrity() {
		return false
	}

	if !hWrapper.validityCheck() {
		return false
	}

	return true
}

// Integrity checks the integrity of the state block wrapper
func (hWrapper HeaderWrapper) Integrity() bool {
	if hWrapper.Processor == nil {
		log.Error("Processor is nil")
		return false
	}

	if hWrapper.Header == nil {
		log.Debug("header is nil")
		return false
	}

	if hWrapper.BlockBodyHash == nil {
		log.Debug("header block body hash is nil")
		return false
	}

	if hWrapper.PubKeysBitmap == nil {
		log.Debug("header publick keys bitmap is nil")
		return false
	}

	if hWrapper.ShardId >= hWrapper.Processor.GetNbShards() {
		log.Debug("header invalid shard number")
		return false
	}

	if hWrapper.PrevHash == nil {
		return false
	}

	if hWrapper.Signature == nil {
		return false
	}

	return true
}

func (hWrapper HeaderWrapper) validityCheck() bool {
	// TODO: need to check epoch is round - timestamp - epoch - nonce - requires chronology

	return true
}

// VerifySig verifies a signature
func (hWrapper HeaderWrapper) VerifySig() bool {
	if hWrapper.Header == nil || hWrapper.Header.Signature == nil {
		return false
	}

	// TODO: Check block signature after multisig will be implemented
	return true
}
