package headerCheck

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	cryptoCommon "github.com/multiversx/mx-chain-go/common/crypto"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

const headerWaitDelayAtTransition = 50 * time.Millisecond

var _ process.InterceptedHeaderSigVerifier = (*HeaderSigVerifier)(nil)

var log = logger.GetOrCreate("process/headerCheck")

// ArgsHeaderSigVerifier is used to store all components that are needed to create a new HeaderSigVerifier
type ArgsHeaderSigVerifier struct {
	Marshalizer             marshal.Marshalizer
	Hasher                  hashing.Hasher
	NodesCoordinator        nodesCoordinator.NodesCoordinator
	MultiSigContainer       cryptoCommon.MultiSignerContainer
	SingleSigVerifier       crypto.SingleSigner
	KeyGen                  crypto.KeyGenerator
	FallbackHeaderValidator process.FallbackHeaderValidator
	EnableEpochsHandler     common.EnableEpochsHandler
	HeadersPool             dataRetriever.HeadersPool
	ProofsPool              dataRetriever.ProofsPool
	StorageService          dataRetriever.StorageService
}

// HeaderSigVerifier is component used to check if a header is valid
type HeaderSigVerifier struct {
	marshalizer             marshal.Marshalizer
	hasher                  hashing.Hasher
	nodesCoordinator        nodesCoordinator.NodesCoordinator
	multiSigContainer       cryptoCommon.MultiSignerContainer
	singleSigVerifier       crypto.SingleSigner
	keyGen                  crypto.KeyGenerator
	fallbackHeaderValidator process.FallbackHeaderValidator
	enableEpochsHandler     common.EnableEpochsHandler
	headersPool             dataRetriever.HeadersPool
	proofsPool              dataRetriever.ProofsPool
	storageService          dataRetriever.StorageService
}

// NewHeaderSigVerifier will create a new instance of HeaderSigVerifier
func NewHeaderSigVerifier(arguments *ArgsHeaderSigVerifier) (*HeaderSigVerifier, error) {
	err := checkArgsHeaderSigVerifier(arguments)
	if err != nil {
		return nil, err
	}

	return &HeaderSigVerifier{
		marshalizer:             arguments.Marshalizer,
		hasher:                  arguments.Hasher,
		nodesCoordinator:        arguments.NodesCoordinator,
		multiSigContainer:       arguments.MultiSigContainer,
		singleSigVerifier:       arguments.SingleSigVerifier,
		keyGen:                  arguments.KeyGen,
		fallbackHeaderValidator: arguments.FallbackHeaderValidator,
		enableEpochsHandler:     arguments.EnableEpochsHandler,
		headersPool:             arguments.HeadersPool,
		proofsPool:              arguments.ProofsPool,
		storageService:          arguments.StorageService,
	}, nil
}

func checkArgsHeaderSigVerifier(arguments *ArgsHeaderSigVerifier) error {
	if arguments == nil {
		return process.ErrNilArgumentStruct
	}
	if check.IfNil(arguments.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(arguments.KeyGen) {
		return process.ErrNilKeyGen
	}
	if check.IfNil(arguments.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(arguments.MultiSigContainer) {
		return process.ErrNilMultiSignerContainer
	}
	multiSigner, err := arguments.MultiSigContainer.GetMultiSigner(0)
	if err != nil {
		return err
	}
	if check.IfNil(multiSigner) {
		return process.ErrNilMultiSigVerifier
	}
	if check.IfNil(arguments.NodesCoordinator) {
		return process.ErrNilNodesCoordinator
	}
	if check.IfNil(arguments.SingleSigVerifier) {
		return process.ErrNilSingleSigner
	}
	if check.IfNil(arguments.FallbackHeaderValidator) {
		return process.ErrNilFallbackHeaderValidator
	}
	if check.IfNil(arguments.EnableEpochsHandler) {
		return process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(arguments.HeadersPool) {
		return process.ErrNilHeadersDataPool
	}
	if check.IfNil(arguments.ProofsPool) {
		return process.ErrNilProofsPool
	}
	if check.IfNil(arguments.StorageService) {
		return process.ErrNilStorageService
	}

	return nil
}

func isIndexInBitmap(index uint16, bitmap []byte) error {
	indexOutOfBounds := index >= uint16(len(bitmap)*8)
	if indexOutOfBounds {
		return ErrIndexOutOfBounds
	}

	indexNotInBitmap := bitmap[index/8]&(1<<uint8(index%8)) == 0
	if indexNotInBitmap {
		return ErrIndexNotSelected
	}

	return nil
}

func (hsv *HeaderSigVerifier) getConsensusSignersForEquivalentProofs(proof data.HeaderProofHandler) ([][]byte, error) {
	if check.IfNil(proof) {
		return nil, process.ErrNilHeaderProof
	}
	if !hsv.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, proof.GetHeaderEpoch()) {
		return nil, process.ErrUnexpectedHeaderProof
	}

	// TODO: remove if start of epochForConsensus block needs to be validated by the new epochForConsensus nodes
	epochForConsensus := common.GetEpochForConsensus(proof)

	consensusPubKeys, err := hsv.nodesCoordinator.GetAllEligibleValidatorsPublicKeysForShard(
		epochForConsensus,
		proof.GetHeaderShardId(),
	)
	if err != nil {
		return nil, err
	}

	shouldApplyFallbackValidation := hsv.fallbackHeaderValidator.ShouldApplyFallbackValidationForHeaderWith(
		proof.GetHeaderShardId(),
		proof.GetIsStartOfEpoch(),
		proof.GetHeaderRound(),
		proof.GetHeaderHash(),
	)

	err = common.IsConsensusBitmapValid(
		log,
		consensusPubKeys,
		proof.GetPubKeysBitmap(),
		shouldApplyFallbackValidation,
	)
	if err != nil {
		return nil, err
	}

	return getPubKeySigners(consensusPubKeys, proof.GetPubKeysBitmap()), nil
}

func (hsv *HeaderSigVerifier) getConsensusSigners(
	randSeed []byte,
	shardID uint32,
	epoch uint32,
	startOfEpochBlock bool,
	round uint64,
	prevHash []byte,
	pubKeysBitmap []byte,
) ([][]byte, error) {
	if len(pubKeysBitmap) == 0 {
		return nil, process.ErrNilPubKeysBitmap
	}

	if !hsv.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, epoch) {
		if pubKeysBitmap[0]&1 == 0 {
			return nil, process.ErrBlockProposerSignatureMissing
		}
	}

	// TODO: remove if start of epochForConsensus block needs to be validated by the new epochForConsensus nodes
	epochForConsensus := epoch
	if startOfEpochBlock && epochForConsensus > 0 {
		epochForConsensus = epochForConsensus - 1
	}

	_, consensusPubKeys, err := hsv.nodesCoordinator.GetConsensusValidatorsPublicKeys(
		randSeed,
		round,
		shardID,
		epochForConsensus,
	)
	if err != nil {
		return nil, err
	}

	shouldApplyFallbackValidation := hsv.fallbackHeaderValidator.ShouldApplyFallbackValidationForHeaderWith(
		shardID,
		startOfEpochBlock,
		round,
		prevHash,
	)

	err = common.IsConsensusBitmapValid(
		log,
		consensusPubKeys,
		pubKeysBitmap,
		shouldApplyFallbackValidation,
	)
	if err != nil {
		return nil, err
	}

	return getPubKeySigners(consensusPubKeys, pubKeysBitmap), nil
}

func getPubKeySigners(consensusPubKeys []string, pubKeysBitmap []byte) [][]byte {
	pubKeysSigners := make([][]byte, 0, len(consensusPubKeys))
	for i := range consensusPubKeys {
		err := isIndexInBitmap(uint16(i), pubKeysBitmap)
		if err != nil {
			continue
		}
		pubKeysSigners = append(pubKeysSigners, []byte(consensusPubKeys[i]))
	}

	return pubKeysSigners
}

// VerifySignature will check if signature is correct
func (hsv *HeaderSigVerifier) VerifySignature(header data.HeaderHandler) error {
	if hsv.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, header.GetEpoch()) {
		return nil
	}

	headerCopy, err := hsv.copyHeaderWithoutSig(header)
	if err != nil {
		return err
	}

	hash, err := core.CalculateHash(hsv.marshalizer, hsv.hasher, headerCopy)
	if err != nil {
		return err
	}

	bitmap := header.GetPubKeysBitmap()
	sig := header.GetSignature()
	return hsv.VerifySignatureForHash(headerCopy, hash, bitmap, sig)
}

// VerifySignatureForHash will check if signature is correct for the provided hash
func (hsv *HeaderSigVerifier) VerifySignatureForHash(header data.HeaderHandler, hash []byte, pubkeysBitmap []byte, signature []byte) error {
	multiSigVerifier, err := hsv.multiSigContainer.GetMultiSigner(header.GetEpoch())
	if err != nil {
		return err
	}

	randSeed := header.GetPrevRandSeed()
	if randSeed == nil {
		return process.ErrNilPrevRandSeed
	}
	pubKeysSigners, err := hsv.getConsensusSigners(
		randSeed,
		header.GetShardID(),
		header.GetEpoch(),
		header.IsStartOfEpochBlock(),
		header.GetRound(),
		header.GetPrevHash(),
		pubkeysBitmap,
	)
	if err != nil {
		return err
	}

	return multiSigVerifier.VerifyAggregatedSig(pubKeysSigners, hash, signature)
}

func (hsv *HeaderSigVerifier) getHeaderForProofAtTransition(proof data.HeaderProofHandler) (data.HeaderHandler, error) {
	var header data.HeaderHandler
	var err error

	for {
		header, err = process.GetHeader(proof.GetHeaderHash(), hsv.headersPool, hsv.storageService, hsv.marshalizer, proof.GetHeaderShardId())
		if err == nil {
			break
		}

		log.Debug("getHeaderForProofAtTransition: failed to get header, will wait and try again",
			"headerHash", proof.GetHeaderHash(),
			"error", err.Error(),
		)

		time.Sleep(headerWaitDelayAtTransition)
	}

	return header, nil
}

func (hsv *HeaderSigVerifier) verifyHeaderProofAtTransition(proof data.HeaderProofHandler) error {
	if check.IfNil(proof) {
		return process.ErrNilHeaderProof
	}
	header, err := hsv.getHeaderForProofAtTransition(proof)
	if err != nil {
		return err
	}

	consensusPubKeys, err := hsv.getConsensusSigners(
		header.GetPrevRandSeed(),
		proof.GetHeaderShardId(),
		proof.GetHeaderEpoch(),
		proof.GetIsStartOfEpoch(),
		proof.GetHeaderRound(),
		proof.GetHeaderHash(),
		proof.GetPubKeysBitmap())
	if err != nil {
		return err
	}

	multiSigVerifier, err := hsv.multiSigContainer.GetMultiSigner(proof.GetHeaderEpoch())
	if err != nil {
		return err
	}

	return multiSigVerifier.VerifyAggregatedSig(consensusPubKeys, proof.GetHeaderHash(), proof.GetAggregatedSignature())
}

// VerifyHeaderProof checks if the proof is correct for the header
func (hsv *HeaderSigVerifier) VerifyHeaderProof(proofHandler data.HeaderProofHandler) error {
	if check.IfNil(proofHandler) {
		return process.ErrNilHeaderProof
	}
	if !hsv.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, proofHandler.GetHeaderEpoch()) {
		return fmt.Errorf("%w for flag %s", process.ErrFlagNotActive, common.AndromedaFlag)
	}

	if common.IsEpochStartProofForFlagActivation(proofHandler, hsv.enableEpochsHandler) {
		return hsv.verifyHeaderProofAtTransition(proofHandler)
	}

	multiSigVerifier, err := hsv.multiSigContainer.GetMultiSigner(proofHandler.GetHeaderEpoch())
	if err != nil {
		return err
	}

	consensusPubKeys, err := hsv.getConsensusSignersForEquivalentProofs(proofHandler)
	if err != nil {
		return err
	}

	return multiSigVerifier.VerifyAggregatedSig(consensusPubKeys, proofHandler.GetHeaderHash(), proofHandler.GetAggregatedSignature())
}

// VerifyRandSeed will check if rand seed is correct
func (hsv *HeaderSigVerifier) VerifyRandSeed(header data.HeaderHandler) error {
	leaderPubKey, err := hsv.getLeader(header)
	if err != nil {
		return err
	}

	err = hsv.verifyRandSeed(leaderPubKey, header)
	if err != nil {
		log.Trace("block rand seed",
			"error", err.Error())
		return err
	}

	return nil
}

// VerifyLeaderSignature will check if leader signature is correct
func (hsv *HeaderSigVerifier) VerifyLeaderSignature(header data.HeaderHandler) error {
	leaderPubKey, err := hsv.getLeader(header)
	if err != nil {
		return err
	}

	err = hsv.verifyLeaderSignature(leaderPubKey, header)
	if err != nil {
		log.Trace("block leader's signature",
			"error", err.Error())
		return err
	}

	return nil
}

// VerifyRandSeedAndLeaderSignature will check if rand seed and leader signature is correct
func (hsv *HeaderSigVerifier) VerifyRandSeedAndLeaderSignature(header data.HeaderHandler) error {
	leaderPubKey, err := hsv.getLeader(header)
	if err != nil {
		return err
	}

	err = hsv.verifyRandSeed(leaderPubKey, header)
	if err != nil {
		log.Trace("block rand seed",
			"error", err.Error())
		return err
	}

	err = hsv.verifyLeaderSignature(leaderPubKey, header)
	if err != nil {
		log.Trace("block leader's signature",
			"error", err.Error())
		return err
	}

	return nil
}

// IsInterfaceNil will check if interface is nil
func (hsv *HeaderSigVerifier) IsInterfaceNil() bool {
	return hsv == nil
}

func (hsv *HeaderSigVerifier) verifyRandSeed(leaderPubKey crypto.PublicKey, header data.HeaderHandler) error {
	prevRandSeed := header.GetPrevRandSeed()
	if prevRandSeed == nil {
		return process.ErrNilPrevRandSeed
	}

	randSeed := header.GetRandSeed()
	if randSeed == nil {
		return process.ErrNilRandSeed
	}

	return hsv.singleSigVerifier.Verify(leaderPubKey, prevRandSeed, randSeed)
}

func (hsv *HeaderSigVerifier) verifyLeaderSignature(leaderPubKey crypto.PublicKey, header data.HeaderHandler) error {
	headerCopy, err := hsv.copyHeaderWithoutLeaderSig(header)
	if err != nil {
		return err
	}

	headerBytes, err := hsv.marshalizer.Marshal(headerCopy)
	if err != nil {
		return err
	}

	return hsv.singleSigVerifier.Verify(leaderPubKey, headerBytes, header.GetLeaderSignature())
}

func (hsv *HeaderSigVerifier) getLeader(header data.HeaderHandler) (crypto.PublicKey, error) {
	leader, _, err := ComputeConsensusGroup(header, hsv.nodesCoordinator)
	if err != nil {
		return nil, err
	}
	return hsv.keyGen.PublicKeyFromByteArray(leader.PubKey())
}

func (hsv *HeaderSigVerifier) copyHeaderWithoutSig(header data.HeaderHandler) (data.HeaderHandler, error) {
	headerCopy := header.ShallowClone()
	err := headerCopy.SetSignature(nil)
	if err != nil {
		return nil, err
	}

	err = headerCopy.SetPubKeysBitmap(nil)
	if err != nil {
		return nil, err
	}

	if !hsv.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, header.GetEpoch()) {
		err = headerCopy.SetLeaderSignature(nil)
		if err != nil {
			return nil, err
		}
	}

	return headerCopy, nil
}

func (hsv *HeaderSigVerifier) copyHeaderWithoutLeaderSig(header data.HeaderHandler) (data.HeaderHandler, error) {
	headerCopy := header.ShallowClone()
	err := headerCopy.SetLeaderSignature(nil)
	if err != nil {
		return nil, err
	}

	return headerCopy, nil
}
