package common

import (
	"fmt"
	"math/bits"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/consensus"
)

// IsValidRelayedTxV3 returns true if the provided transaction is a valid transaction of type relayed v3
func IsValidRelayedTxV3(tx data.TransactionHandler) bool {
	relayedTx, isRelayedV3 := tx.(data.RelayedTransactionHandler)
	if !isRelayedV3 {
		return false
	}
	hasValidRelayer := len(relayedTx.GetRelayerAddr()) == len(tx.GetSndAddr()) && len(relayedTx.GetRelayerAddr()) > 0
	hasValidRelayerSignature := len(relayedTx.GetRelayerSignature()) == len(relayedTx.GetSignature()) && len(relayedTx.GetRelayerSignature()) > 0
	return hasValidRelayer && hasValidRelayerSignature
}

// IsRelayedTxV3 returns true if the provided transaction is a transaction of type relayed v3, without any further checks
func IsRelayedTxV3(tx data.TransactionHandler) bool {
	relayedTx, isRelayedV3 := tx.(data.RelayedTransactionHandler)
	if !isRelayedV3 {
		return false
	}

	hasRelayer := len(relayedTx.GetRelayerAddr()) > 0
	hasRelayerSignature := len(relayedTx.GetRelayerSignature()) > 0
	return hasRelayer || hasRelayerSignature
}

// IsEpochChangeBlockForFlagActivation returns true if the provided header is the first one after the specified flag's activation
func IsEpochChangeBlockForFlagActivation(header data.HeaderHandler, enableEpochsHandler EnableEpochsHandler, flag core.EnableEpochFlag) bool {
	isStartOfEpochBlock := header.IsStartOfEpochBlock()
	isBlockInActivationEpoch := header.GetEpoch() == enableEpochsHandler.GetActivationEpoch(flag)

	return isStartOfEpochBlock && isBlockInActivationEpoch
}

// IsEpochStartProofForFlagActivation returns true if the provided proof is the proof of the epoch start block on the activation epoch of equivalent messages
func IsEpochStartProofForFlagActivation(proof consensus.ProofHandler, enableEpochsHandler EnableEpochsHandler) bool {
	isStartOfEpochProof := proof.GetIsStartOfEpoch()
	isProofInActivationEpoch := proof.GetHeaderEpoch() == enableEpochsHandler.GetActivationEpoch(EquivalentMessagesFlag)

	return isStartOfEpochProof && isProofInActivationEpoch
}

// isFlagEnabledAfterEpochsStartBlock returns true if the flag is enabled for the header, but it is not the epoch start block
func isFlagEnabledAfterEpochsStartBlock(header data.HeaderHandler, enableEpochsHandler EnableEpochsHandler, flag core.EnableEpochFlag) bool {
	isFlagEnabled := enableEpochsHandler.IsFlagEnabledInEpoch(flag, header.GetEpoch())
	isEpochStartBlock := IsEpochChangeBlockForFlagActivation(header, enableEpochsHandler, flag)
	return isFlagEnabled && !isEpochStartBlock
}

// ShouldBlockHavePrevProof returns true if the block should have a proof
func ShouldBlockHavePrevProof(header data.HeaderHandler, enableEpochsHandler EnableEpochsHandler, flag core.EnableEpochFlag) bool {
	return isFlagEnabledAfterEpochsStartBlock(header, enableEpochsHandler, flag) && header.GetNonce() > 1
}

// VerifyProofAgainstHeader verifies the fields on the proof match the ones on the header
func VerifyProofAgainstHeader(proof data.HeaderProofHandler, header data.HeaderHandler) error {
	if check.IfNil(proof) {
		return ErrNilHeaderProof
	}
	if check.IfNil(header) {
		return ErrNilHeaderHandler
	}

	if proof.GetHeaderNonce() != header.GetNonce() {
		return fmt.Errorf("%w, nonce mismatch", ErrInvalidHeaderProof)
	}
	if proof.GetHeaderShardId() != header.GetShardID() {
		return fmt.Errorf("%w, shard id mismatch", ErrInvalidHeaderProof)
	}
	if proof.GetHeaderEpoch() != header.GetEpoch() {
		return fmt.Errorf("%w, epoch mismatch", ErrInvalidHeaderProof)
	}
	if proof.GetHeaderRound() != header.GetRound() {
		return fmt.Errorf("%w, round mismatch", ErrInvalidHeaderProof)
	}
	if proof.GetIsStartOfEpoch() != header.IsStartOfEpochBlock() {
		return fmt.Errorf("%w, is start of epoch mismatch", ErrInvalidHeaderProof)
	}

	return nil
}

// GetShardIDs returns a map of shard IDs based on the number of shards
func GetShardIDs(numShards uint32) map[uint32]struct{} {
	shardIdentifiers := make(map[uint32]struct{})
	for i := uint32(0); i < numShards; i++ {
		shardIdentifiers[i] = struct{}{}
	}
	shardIdentifiers[core.MetachainShardId] = struct{}{}

	return shardIdentifiers
}

// IsConsensusBitmapValid checks if the provided keys and bitmap match the consensus requirements
func IsConsensusBitmapValid(
	log logger.Logger,
	consensusPubKeys []string,
	bitmap []byte,
	shouldApplyFallbackValidation bool,
) error {
	consensusSize := len(consensusPubKeys)

	expectedBitmapSize := consensusSize / 8
	if consensusSize%8 != 0 {
		expectedBitmapSize++
	}
	if len(bitmap) != expectedBitmapSize {
		log.Debug("wrong size bitmap",
			"expected number of bytes", expectedBitmapSize,
			"actual", len(bitmap))
		return ErrWrongSizeBitmap
	}

	numOfOnesInBitmap := 0
	for index := range bitmap {
		numOfOnesInBitmap += bits.OnesCount8(bitmap[index])
	}

	minNumRequiredSignatures := core.GetPBFTThreshold(consensusSize)
	if shouldApplyFallbackValidation {
		minNumRequiredSignatures = core.GetPBFTFallbackThreshold(consensusSize)
		log.Warn("IsConsensusBitmapValid: fallback validation has been applied",
			"minimum number of signatures required", minNumRequiredSignatures,
			"actual number of signatures in bitmap", numOfOnesInBitmap,
		)
	}

	if numOfOnesInBitmap >= minNumRequiredSignatures {
		return nil
	}

	log.Debug("not enough signatures",
		"minimum expected", minNumRequiredSignatures,
		"actual", numOfOnesInBitmap)

	return ErrNotEnoughSignatures
}
