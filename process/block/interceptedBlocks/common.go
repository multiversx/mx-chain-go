package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

func checkBlockHeaderArgument(arg *ArgInterceptedBlockHeader) error {
	if arg == nil {
		return process.ErrNilArguments
	}
	if arg.HdrBuff == nil {
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

	return nil
}

func checkTxBlockBodyArgument(arg *ArgInterceptedTxBlockBody) error {
	if arg == nil {
		return process.ErrNilArguments
	}
	if arg.TxBlockBodyBuff == nil {
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
	if hdr.GetPubKeysBitmap() == nil {
		return process.ErrNilPubKeysBitmap
	}
	if hdr.GetPrevHash() == nil {
		return process.ErrNilPreviousBlockHash
	}
	if hdr.GetSignature() == nil {
		return process.ErrNilSignature
	}
	if hdr.GetRootHash() == nil {
		return process.ErrNilRootHash
	}
	if hdr.GetRandSeed() == nil {
		return process.ErrNilRandSeed
	}
	if hdr.GetPrevRandSeed() == nil {
		return process.ErrNilPrevRandSeed
	}

	return nil
}

func checkMetaShardInfo(shardInfo []block.ShardData, coordinator sharding.Coordinator) error {
	for _, sd := range shardInfo {
		if sd.ShardID >= coordinator.NumberOfShards() && sd.ShardID != sharding.MetachainShardId {
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
			smbh.SenderShardID != sharding.MetachainShardId
		isWrongDestinationShardId := smbh.ReceiverShardID >= coordinator.NumberOfShards() &&
			smbh.ReceiverShardID != sharding.MetachainShardId
		isWrongShardId := isWrongSenderShardId || isWrongDestinationShardId
		if isWrongShardId {
			return process.ErrInvalidShardId
		}
	}

	return nil
}

func checkMiniblocks(miniblocks []block.MiniBlockHeader, coordinator sharding.Coordinator) error {
	for _, miniblock := range miniblocks {
		isWrongSenderShardId := miniblock.SenderShardID >= coordinator.NumberOfShards() &&
			miniblock.SenderShardID != sharding.MetachainShardId
		isWrongDestinationShardId := miniblock.ReceiverShardID >= coordinator.NumberOfShards() &&
			miniblock.ReceiverShardID != sharding.MetachainShardId
		isWrongShardId := isWrongSenderShardId || isWrongDestinationShardId
		if isWrongShardId {
			return process.ErrInvalidShardId
		}
	}

	return nil
}
