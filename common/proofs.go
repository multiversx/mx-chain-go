package common

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

// IsEpochStartProofForFlagActivation returns true if the provided proof is the proof of the epoch start block on the activation epoch of equivalent messages
func IsEpochStartProofForFlagActivation(proof consensus.ProofHandler, enableEpochsHandler EnableEpochsHandler) bool {
	isStartOfEpochProof := proof.GetIsStartOfEpoch()
	isProofInActivationEpoch := proof.GetHeaderEpoch() == enableEpochsHandler.GetActivationEpoch(AndromedaFlag)

	return isStartOfEpochProof && isProofInActivationEpoch
}

// IsProofsFlagEnabledForHeader returns true if proofs flag has to be enabled for the provided header
func IsProofsFlagEnabledForHeader(
	enableEpochsHandler EnableEpochsHandler,
	header data.HeaderHandler,
) bool {
	ifFlagActive := enableEpochsHandler.IsFlagEnabledInEpoch(AndromedaFlag, header.GetEpoch())
	isGenesisBlock := header.GetNonce() == 0

	return ifFlagActive && !isGenesisBlock
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

// GetEpochForConsensus will get epoch to be used by consensus based on equivalent proof data
func GetEpochForConsensus(proof data.HeaderProofHandler) uint32 {
	epochForConsensus := proof.GetHeaderEpoch()
	if proof.GetIsStartOfEpoch() && epochForConsensus > 0 {
		epochForConsensus = epochForConsensus - 1
	}

	return epochForConsensus
}
