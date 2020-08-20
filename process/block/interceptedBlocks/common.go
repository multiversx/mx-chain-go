package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

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

	return nil
}

func checkMetaShardInfo(shardInfo []block.ShardData, coordinator sharding.Coordinator) error {
	for _, sd := range shardInfo {
		if sd.ShardID >= coordinator.NumberOfShards() && sd.ShardID != core.MetachainShardId {
			return process.ErrInvalidShardId
		}

		err := checkShardData(sd, coordinator)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkShardData(sd block.ShardData, coordinator sharding.Coordinator) error {
	for _, smbh := range sd.ShardMiniBlockHeaders {
		isWrongSenderShardId := smbh.SenderShardID >= coordinator.NumberOfShards() &&
			smbh.SenderShardID != core.MetachainShardId &&
			smbh.SenderShardID != core.AllShardId
		isWrongDestinationShardId := smbh.ReceiverShardID >= coordinator.NumberOfShards() &&
			smbh.ReceiverShardID != core.MetachainShardId &&
			smbh.ReceiverShardID != core.AllShardId
		isWrongShardId := isWrongSenderShardId || isWrongDestinationShardId
		if isWrongShardId {
			return process.ErrInvalidShardId
		}

		if len(smbh.Reserved) > 0 {
			return process.ErrReservedFieldNotSupportedYet
		}
	}

	return nil
}

func checkMiniblocks(miniblocks []block.MiniBlockHeader, coordinator sharding.Coordinator) error {
	for _, miniblock := range miniblocks {
		isWrongSenderShardId := miniblock.SenderShardID >= coordinator.NumberOfShards() &&
			miniblock.SenderShardID != core.MetachainShardId &&
			miniblock.SenderShardID != core.AllShardId
		isWrongDestinationShardId := miniblock.ReceiverShardID >= coordinator.NumberOfShards() &&
			miniblock.ReceiverShardID != core.MetachainShardId &&
			miniblock.ReceiverShardID != core.AllShardId
		isWrongShardId := isWrongSenderShardId || isWrongDestinationShardId
		if isWrongShardId {
			return process.ErrInvalidShardId
		}

		if len(miniblock.Reserved) > 0 {
			return process.ErrReservedFieldNotSupportedYet
		}
	}

	return nil
}
