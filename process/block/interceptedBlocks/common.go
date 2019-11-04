package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// headerMultiSigVerifier is an "abstract" struct that is able to verify the signature of a header handler
type headerMultiSigVerifier struct {
	marshalizer          marshal.Marshalizer
	hasher               hashing.Hasher
	nodesCoordinator     sharding.NodesCoordinator
	multiSigVerifier     crypto.MultiSigVerifier
	copyHeaderWithoutSig func(src data.HeaderHandler) data.HeaderHandler
}

func (hmsv *headerMultiSigVerifier) verifySig(header data.HeaderHandler) error {

	randSeed := header.GetPrevRandSeed()
	bitmap := header.GetPubKeysBitmap()

	//TODO: check randSeed = Sig_proposer(prevRandSeed)

	if len(bitmap) == 0 {
		return process.ErrNilPubKeysBitmap
	}
	if bitmap[0]&1 == 0 {
		return process.ErrBlockProposerSignatureMissing
	}

	consensusPubKeys, err := hmsv.nodesCoordinator.GetValidatorsPublicKeys(
		randSeed,
		header.GetRound(),
		header.GetShardID(),
	)
	if err != nil {
		return err
	}

	verifier, err := hmsv.multiSigVerifier.Create(consensusPubKeys, 0)
	if err != nil {
		return err
	}

	err = verifier.SetAggregatedSig(header.GetSignature())
	if err != nil {
		return err
	}

	// get marshalled block header without signature and bitmap
	// as this is the message that was signed
	headerCopy := hmsv.copyHeaderWithoutSig(header)

	hash, err := core.CalculateHash(hmsv.marshalizer, hmsv.hasher, headerCopy)
	if err != nil {
		return err
	}

	return verifier.Verify(hash, bitmap)
}

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
	if check.IfNil(arg.MultiSigVerifier) {
		return process.ErrNilMultiSigVerifier
	}
	if check.IfNil(arg.NodesCoordinator) {
		return process.ErrNilNodesCoordinator
	}
	if check.IfNil(arg.ShardCoordinator) {
		return process.ErrNilShardCoordinator
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
		if sd.ShardID >= coordinator.NumberOfShards() {
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
		isWrongSenderShardId := smbh.SenderShardID >= coordinator.NumberOfShards()
		isWrongDestinationShardId := smbh.ReceiverShardID >= coordinator.NumberOfShards()
		isWrongShardId := isWrongSenderShardId || isWrongDestinationShardId
		if isWrongShardId {
			return process.ErrInvalidShardId
		}
	}

	return nil
}

func checkMiniblocks(miniblocks []block.MiniBlockHeader, coordinator sharding.Coordinator) error {
	for _, miniblock := range miniblocks {
		isWrongSenderShardId := miniblock.SenderShardID >= coordinator.NumberOfShards()
		isWrongDestinationShardId := miniblock.ReceiverShardID >= coordinator.NumberOfShards()
		isWrongShardId := isWrongSenderShardId || isWrongDestinationShardId
		if isWrongShardId {
			return process.ErrInvalidShardId
		}
	}

	return nil
}
