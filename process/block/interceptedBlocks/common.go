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
type headerSigVerifier struct {
	marshalizer          marshal.Marshalizer
	hasher               hashing.Hasher
	nodesCoordinator     sharding.NodesCoordinator
	multiSigVerifier     crypto.MultiSigVerifier
	singleSigVerifier    crypto.SingleSigner
	keyGen               crypto.KeyGenerator
	copyHeaderWithoutSig func(src data.HeaderHandler) data.HeaderHandler
}

func (hsv *headerSigVerifier) verifySig(header data.HeaderHandler) error {

	randSeed := header.GetPrevRandSeed()
	bitmap := header.GetPubKeysBitmap()
	if len(bitmap) == 0 {
		return process.ErrNilPubKeysBitmap
	}
	if bitmap[0]&1 == 0 {
		return process.ErrBlockProposerSignatureMissing
	}

	consensusPubKeys, err := hsv.nodesCoordinator.GetValidatorsPublicKeys(
		randSeed,
		header.GetRound(),
		header.GetShardID(),
	)
	if err != nil {
		return err
	}

	verifier, err := hsv.multiSigVerifier.Create(consensusPubKeys, 0)
	if err != nil {
		return err
	}

	err = verifier.SetAggregatedSig(header.GetSignature())
	if err != nil {
		return err
	}

	// get marshalled block header without signature and bitmap
	// as this is the message that was signed
	headerCopy := hsv.copyHeaderWithoutSig(header)

	hash, err := core.CalculateHash(hsv.marshalizer, hsv.hasher, headerCopy)
	if err != nil {
		return err
	}

	return verifier.Verify(hash, bitmap)
}

func (hsv *headerSigVerifier) verifyRandSeed(header data.HeaderHandler) error {
	prevRandSeed := header.GetPrevRandSeed()
	headerConsensusGroup, err := hsv.nodesCoordinator.ComputeValidatorsGroup(prevRandSeed, header.GetRound(), header.GetShardID())
	if err != nil {
		return err
	}

	randSeed := header.GetRandSeed()
	leaderPubKeyValidator := headerConsensusGroup[0]
	leaderPubKey, err := hsv.keyGen.PublicKeyFromByteArray(leaderPubKeyValidator.PubKey())
	if err != nil {
		return err
	}

	err = hsv.singleSigVerifier.Verify(leaderPubKey, prevRandSeed, randSeed)
	if err != nil {
		return err
	}

	return nil
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
	if check.IfNil(arg.KeyGen) {
		return process.ErrNilKeyGen
	}
	if check.IfNil(arg.SingleSigVerifier) {
		return process.ErrNilSingleSigner
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
