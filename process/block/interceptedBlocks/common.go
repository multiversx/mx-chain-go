package interceptedBlocks

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

const maxLenMiniBlockReservedField = 10
const maxLenMiniBlockHeaderReservedField = 32

func checkForDuplicateHashes(hashes [][]byte) error {
	mapHashes := make(map[string]struct{}, len(hashes))
	for _, hash := range hashes {
		hashStr := string(hash)
		if _, exists := mapHashes[hashStr]; exists {
			return process.ErrDuplicatedHashInBlock
		}
		mapHashes[hashStr] = struct{}{}
	}
	return nil
}

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
	if check.IfNil(arg.EpochChangeGracePeriodHandler) {
		return process.ErrNilEpochChangeGracePeriodHandler
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

func checkHeaderHandler(
	hdr data.HeaderHandler,
	enableEpochsHandler common.EnableEpochsHandler,
) error {
	equivalentMessagesEnabled := enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, hdr.GetEpoch())

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

	return hdr.CheckFieldsForNil()
}

func checkMetaShardInfo(
	shardInfo []data.ShardDataHandler,
	coordinator sharding.Coordinator,
) error {
	if coordinator.SelfId() != core.MetachainShardId {
		return nil
	}

	for _, sd := range shardInfo {
		if sd.GetShardID() >= coordinator.NumberOfShards() && sd.GetShardID() != core.MetachainShardId {
			return process.ErrInvalidShardId
		}

		err := checkShardData(sd, coordinator)
		if err != nil {
			return err
		}
	}

	shardDataHashes := make([][]byte, len(shardInfo))
	for i, sd := range shardInfo {
		shardDataHashes[i] = sd.GetHeaderHash()
	}

	return checkForDuplicateHashes(shardDataHashes)
}

func checkShardData(sd data.ShardDataHandler, coordinator sharding.Coordinator) error {
	shardMBHeaders := sd.GetShardMiniBlockHeaderHandlers()
	mbHashes := make([][]byte, len(shardMBHeaders))
	for i, smbh := range shardMBHeaders {
		mbHashes[i] = smbh.GetHash()
	}
	if err := checkForDuplicateHashes(mbHashes); err != nil {
		return err
	}

	for _, smbh := range shardMBHeaders {
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
	mbHashes := make([][]byte, len(mbHeaders))
	for i, mbHeader := range mbHeaders {
		mbHashes[i] = mbHeader.GetHash()
	}
	if err := checkForDuplicateHashes(mbHashes); err != nil {
		return err
	}

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
