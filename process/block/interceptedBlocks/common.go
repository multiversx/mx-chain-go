package interceptedBlocks

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
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

func checkHeaderHandler(hdr data.HeaderHandler) error {
	if len(hdr.GetPubKeysBitmap()) == 0 {
		return process.ErrNilPubKeysBitmap
	}
	if len(hdr.GetPrevHash()) == 0 {
		return process.ErrNilPreviousBlockHash
	}
	if len(hdr.GetSignature()) == 0 {
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

func checkMetaShardInfo(shardInfo []data.ShardDataHandler, coordinator sharding.Coordinator) error {
	for _, sd := range shardInfo {
		//if sd.GetShardID() >= coordinator.NumberOfShards() && sd.GetShardID() != core.MetachainShardId {
		//	return process.ErrInvalidShardId
		//}

		err := checkShardData(sd, coordinator)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkShardData(sd data.ShardDataHandler, coordinator sharding.Coordinator) error {
	for _, smbh := range sd.GetShardMiniBlockHeaderHandlers() {
		//isWrongSenderShardId := smbh.GetSenderShardID() >= coordinator.NumberOfShards() &&
		//	smbh.GetSenderShardID() != core.MetachainShardId &&
		//	smbh.GetSenderShardID() != core.AllShardId
		//isWrongDestinationShardId := smbh.GetReceiverShardID() >= coordinator.NumberOfShards() &&
		//	smbh.GetReceiverShardID() != core.MetachainShardId &&
		//	smbh.GetReceiverShardID() != core.AllShardId
		//isWrongShardId := isWrongSenderShardId || isWrongDestinationShardId
		//if isWrongShardId {
		//	return process.ErrInvalidShardId
		//}

		if len(smbh.GetReserved()) > maxLenMiniBlockHeaderReservedField {
			return process.ErrReservedFieldInvalid
		}
	}

	return nil
}

func checkMiniBlocksHeaders(mbHeaders []data.MiniBlockHeaderHandler, coordinator sharding.Coordinator) error {
	for _, mbHeader := range mbHeaders {
		//isWrongSenderShardId := mbHeader.GetSenderShardID() >= coordinator.NumberOfShards() &&
		//	mbHeader.GetSenderShardID() != core.MetachainShardId &&
		//	mbHeader.GetSenderShardID() != core.AllShardId
		//isWrongDestinationShardId := mbHeader.GetReceiverShardID() >= coordinator.NumberOfShards() &&
		//	mbHeader.GetReceiverShardID() != core.MetachainShardId &&
		//	mbHeader.GetReceiverShardID() != core.AllShardId
		//isWrongShardId := isWrongSenderShardId || isWrongDestinationShardId
		//if isWrongShardId {
		//	return process.ErrInvalidShardId
		//}

		if len(mbHeader.GetReserved()) > maxLenMiniBlockHeaderReservedField {
			return process.ErrReservedFieldInvalid
		}
	}

	return nil
}
