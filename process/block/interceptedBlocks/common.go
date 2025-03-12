package interceptedBlocks

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

const maxLenMiniBlockReservedField = 10
const maxLenMiniBlockHeaderReservedField = 32

func checkBlockHeaderArgument(arg *ArgInterceptedBlockHeader) error {
	if arg == nil {
		return process.ErrNilArgumentStruct
	}
	if len(arg.HdrBuff) == 0 {
		return process.ErrNilBuffer
	}
	if check.IfNil(arg.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(arg.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(arg.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(arg.HeaderSigVerifier) {
		return process.ErrNilHeaderSigVerifier
	}
	if check.IfNil(arg.HeaderIntegrityVerifier) {
		return process.ErrNilHeaderIntegrityVerifier
	}
	if check.IfNil(arg.EpochStartTrigger) {
		return process.ErrNilEpochStartTrigger
	}
	if check.IfNil(arg.ValidityAttester) {
		return process.ErrNilValidityAttester
	}
	if check.IfNil(arg.EnableEpochsHandler) {
		return process.ErrNilEnableEpochsHandler
	}

	return nil
}

func checkMiniblockArgument(arg *ArgInterceptedMiniblock) error {
	if arg == nil {
		return process.ErrNilArgumentStruct
	}
	if len(arg.MiniblockBuff) == 0 {
		return process.ErrNilBuffer
	}
	if check.IfNil(arg.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(arg.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(arg.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}

	return nil
}

func checkHeaderHandler(hdr data.HeaderHandler, enableEpochsHandler common.EnableEpochsHandler) error {
	equivalentMessagesEnabled := enableEpochsHandler.IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, hdr.GetEpoch())

	if len(hdr.GetPubKeysBitmap()) == 0 && !equivalentMessagesEnabled {
		return process.ErrNilPubKeysBitmap
	}
	if len(hdr.GetPrevHash()) == 0 {
		return process.ErrNilPreviousBlockHash
	}
	if len(hdr.GetSignature()) == 0 && !equivalentMessagesEnabled {
		return process.ErrNilSignature
	}
	if len(hdr.GetRootHash()) == 0 {
		return process.ErrNilRootHash
	}
	if len(hdr.GetRandSeed()) == 0 {
		return process.ErrNilRandSeed
	}
	if len(hdr.GetPrevRandSeed()) == 0 {
		return process.ErrNilPrevRandSeed
	}

	err := checkProofIntegrity(hdr, enableEpochsHandler)
	if err != nil {
		return err
	}

	return hdr.CheckFieldsForNil()
}

func checkProofIntegrity(hdr data.HeaderHandler, enableEpochsHandler common.EnableEpochsHandler) error {
	prevHeaderProof := hdr.GetPreviousProof()
	nilPreviousProof := check.IfNil(prevHeaderProof)
	shouldHavePrevProof := common.ShouldBlockHavePrevProof(hdr, enableEpochsHandler, common.EquivalentMessagesFlag)
	missingPrevProof := nilPreviousProof && shouldHavePrevProof
	unexpectedPrevProof := !nilPreviousProof && !shouldHavePrevProof
	hasPrevProof := !nilPreviousProof && !missingPrevProof

	if missingPrevProof {
		return process.ErrMissingPrevHeaderProof
	}
	if unexpectedPrevProof {
		return process.ErrUnexpectedHeaderProof
	}
	if hasPrevProof && isIncompleteProof(prevHeaderProof) {
		return process.ErrInvalidHeaderProof
	}

	return nil
}

func checkMetaShardInfo(
	shardInfo []data.ShardDataHandler,
	coordinator sharding.Coordinator,
	headerSigVerifier process.InterceptedHeaderSigVerifier,
) error {
	if coordinator.SelfId() != core.MetachainShardId {
		return nil
	}

	wgProofsVerification := sync.WaitGroup{}

	errChan := make(chan error, len(shardInfo))
	defer close(errChan)

	for _, sd := range shardInfo {
		if sd.GetShardID() >= coordinator.NumberOfShards() && sd.GetShardID() != core.MetachainShardId {
			return process.ErrInvalidShardId
		}

		err := checkShardData(sd, coordinator)
		if err != nil {
			return err
		}

		isSelfMeta := coordinator.SelfId() == core.MetachainShardId
		isHeaderFromSelf := sd.GetShardID() == coordinator.SelfId()
		if !(isSelfMeta || isHeaderFromSelf) {
			continue
		}

		wgProofsVerification.Add(1)
		checkProofAsync(sd.GetPreviousProof(), headerSigVerifier, &wgProofsVerification, errChan)
	}

	wgProofsVerification.Wait()

	return readFromChanNonBlocking(errChan)
}

func readFromChanNonBlocking(errChan chan error) error {
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

func checkProofAsync(
	proof data.HeaderProofHandler,
	headerSigVerifier process.InterceptedHeaderSigVerifier,
	wg *sync.WaitGroup,
	errChan chan error,
) {
	go func(proof data.HeaderProofHandler) {
		errCheckProof := checkProof(proof, headerSigVerifier)
		if errCheckProof != nil {
			errChan <- errCheckProof
		}

		wg.Done()
	}(proof)
}

func checkProof(proof data.HeaderProofHandler, headerSigVerifier process.InterceptedHeaderSigVerifier) error {
	if check.IfNil(proof) {
		return nil
	}

	if isIncompleteProof(proof) {
		return process.ErrInvalidHeaderProof
	}

	return headerSigVerifier.VerifyHeaderProof(proof)
}

func isIncompleteProof(proof data.HeaderProofHandler) bool {
	return len(proof.GetAggregatedSignature()) == 0 ||
		len(proof.GetPubKeysBitmap()) == 0 ||
		len(proof.GetHeaderHash()) == 0
}

func checkShardData(sd data.ShardDataHandler, coordinator sharding.Coordinator) error {
	for _, smbh := range sd.GetShardMiniBlockHeaderHandlers() {
		isWrongSenderShardId := smbh.GetSenderShardID() >= coordinator.NumberOfShards() &&
			smbh.GetSenderShardID() != core.MetachainShardId &&
			smbh.GetSenderShardID() != core.AllShardId
		isWrongDestinationShardId := smbh.GetReceiverShardID() >= coordinator.NumberOfShards() &&
			smbh.GetReceiverShardID() != core.MetachainShardId &&
			smbh.GetReceiverShardID() != core.AllShardId
		isWrongShardId := isWrongSenderShardId || isWrongDestinationShardId
		if isWrongShardId {
			return process.ErrInvalidShardId
		}

		if len(smbh.GetReserved()) > maxLenMiniBlockHeaderReservedField {
			return process.ErrReservedFieldInvalid
		}
	}

	return nil
}

func checkMiniBlocksHeaders(mbHeaders []data.MiniBlockHeaderHandler, coordinator sharding.Coordinator) error {
	for _, mbHeader := range mbHeaders {
		isWrongSenderShardId := mbHeader.GetSenderShardID() >= coordinator.NumberOfShards() &&
			mbHeader.GetSenderShardID() != core.MetachainShardId &&
			mbHeader.GetSenderShardID() != core.AllShardId
		isWrongDestinationShardId := mbHeader.GetReceiverShardID() >= coordinator.NumberOfShards() &&
			mbHeader.GetReceiverShardID() != core.MetachainShardId &&
			mbHeader.GetReceiverShardID() != core.AllShardId
		isWrongShardId := isWrongSenderShardId || isWrongDestinationShardId
		if isWrongShardId {
			return process.ErrInvalidShardId
		}

		if len(mbHeader.GetReserved()) > maxLenMiniBlockHeaderReservedField {
			return process.ErrReservedFieldInvalid
		}
	}

	return nil
}
