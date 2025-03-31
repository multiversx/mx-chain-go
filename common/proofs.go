package common

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/consensus"
)

// IsEpochStartProofForFlagActivation returns true if the provided proof is the proof of the epoch start block on the activation epoch of equivalent messages
func IsEpochStartProofForFlagActivation(proof consensus.ProofHandler, enableEpochsHandler EnableEpochsHandler) bool {
	isStartOfEpochProof := proof.GetIsStartOfEpoch()
	isProofInActivationEpoch := proof.GetHeaderEpoch() == enableEpochsHandler.GetActivationEpoch(EquivalentMessagesFlag)

	return isStartOfEpochProof && isProofInActivationEpoch
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
