package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

func checkArgument(arg *ArgInterceptedBlockHeader) error {
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
	if check.IfNil(arg.MultiSigVerifier) {
		return process.ErrNilMultiSigVerifier
	}
	if check.IfNil(arg.ChronologyValidator) {
		return process.ErrNilChronologyValidator
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
		if sd.ShardId >= coordinator.NumberOfShards() {
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
		if smbh.SenderShardId >= coordinator.NumberOfShards() {
			return process.ErrInvalidShardId
		}
		if smbh.ReceiverShardId >= coordinator.NumberOfShards() {
			return process.ErrInvalidShardId
		}
	}

	return nil
}

func checkMiniblocks(miniblocks []block.MiniBlockHeader, coordinator sharding.Coordinator) error {
	for _, miniblock := range miniblocks {
		if miniblock.SenderShardID >= coordinator.NumberOfShards() {
			return process.ErrInvalidShardId
		}
		if miniblock.ReceiverShardID >= coordinator.NumberOfShards() {
			return process.ErrInvalidShardId
		}
	}

	return nil
}
