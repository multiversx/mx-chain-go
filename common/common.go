package common

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-vm-v1_2-go/ipc/marshaling"
)

// IsEpochChangeBlockForFlagActivation returns true if the provided header is the first one after the specified flag's activation
func IsEpochChangeBlockForFlagActivation(header data.HeaderHandler, enableEpochsHandler EnableEpochsHandler, flag core.EnableEpochFlag) bool {
	isStartOfEpochBlock := header.IsStartOfEpochBlock()
	isBlockInActivationEpoch := header.GetEpoch() == enableEpochsHandler.GetActivationEpoch(flag)

	return isStartOfEpochBlock && isBlockInActivationEpoch
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
	if check.IfNilReflect(proof) {
		return ErrInvalidHeaderProof
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

// GetHeader tries to get the header from pool first and if not found, searches for it through storer
func GetHeader(
	headerHash []byte,
	headersPool HeadersPool,
	headersStorer storage.Storer,
	marshaller marshaling.Marshalizer,
) (data.HeaderHandler, error) {
	header, err := headersPool.GetHeaderByHash(headerHash)
	if err == nil {
		return header, nil
	}

	headerBytes, err := headersStorer.SearchFirst(headerHash)
	if err != nil {
		return nil, err
	}

	err = marshaller.Unmarshal(header, headerBytes)
	if err != nil {
		return nil, err
	}

	return header, nil
}
