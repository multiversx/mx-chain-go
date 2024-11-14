package headerCheck

import (
	"fmt"
	"math/bits"

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

func (hsv *HeaderSigVerifier) getConsensusSigners(header data.HeaderHandler, pubKeysBitmap []byte) ([][]byte, error) {
	randSeed := header.GetPrevRandSeed()
	if len(pubKeysBitmap) == 0 {
		return nil, process.ErrNilPubKeysBitmap
	}

	if !hsv.enableEpochsHandler.IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, header.GetEpoch()) {
		if pubKeysBitmap[0]&1 == 0 {
			return nil, process.ErrBlockProposerSignatureMissing
		}
	}

	// TODO: remove if start of epochForConsensus block needs to be validated by the new epochForConsensus nodes
	epochForConsensus := header.GetEpoch()
	if header.IsStartOfEpochBlock() && epochForConsensus > 0 {
		epochForConsensus = epochForConsensus - 1
	}

	_, consensusPubKeys, err := hsv.nodesCoordinator.GetConsensusValidatorsPublicKeys(
		randSeed,
		header.GetRound(),
		header.GetShardID(),
		epochForConsensus,
	)
	if err != nil {
		return nil, err
	}

	err = hsv.verifyConsensusSize(consensusPubKeys, header, pubKeysBitmap)
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
// TODO: Adapt header signature verification for the changes related to equivalent proofs
func (hsv *HeaderSigVerifier) VerifySignature(header data.HeaderHandler) error {
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
	if hsv.enableEpochsHandler.IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, headerCopy.GetEpoch()) {
		headerCopy, hash, sig, bitmap, err = hsv.getPrevHeaderInfo(headerCopy)
		if err != nil {
			return err
		}
	}

	return hsv.VerifySignatureForHash(headerCopy, hash, bitmap, sig)
}

// VerifySignatureForHash will check if signature is correct for the provided hash
func (hsv *HeaderSigVerifier) VerifySignatureForHash(header data.HeaderHandler, hash []byte, pubkeysBitmap []byte, signature []byte) error {
	multiSigVerifier, err := hsv.multiSigContainer.GetMultiSigner(header.GetEpoch())
	if err != nil {
		return err
	}

	pubKeysSigners, err := hsv.getConsensusSigners(header, pubkeysBitmap)
	if err != nil {
		return err
	}

	return multiSigVerifier.VerifyAggregatedSig(pubKeysSigners, hash, signature)
}

// VerifyHeaderProof checks if the proof is correct for the header
func (hsv *HeaderSigVerifier) VerifyHeaderProof(proofHandler data.HeaderProofHandler) error {
	if check.IfNilReflect(proofHandler) {
		return process.ErrNilHeaderProof
	}

	if !hsv.enableEpochsHandler.IsFlagEnabledInEpoch(common.FixedOrderInConsensusFlag, proofHandler.GetHeaderEpoch()) {
		return fmt.Errorf("%w for %s", process.ErrFlagNotActive, common.FixedOrderInConsensusFlag)
	}

	consensusPubKeys, err := hsv.nodesCoordinator.GetAllEligibleValidatorsPublicKeysForShard(proofHandler.GetHeaderEpoch(), proofHandler.GetHeaderShardId())
	if err != nil {
		return err
	}

	multiSigVerifier, err := hsv.multiSigContainer.GetMultiSigner(proofHandler.GetHeaderEpoch())
	if err != nil {
		return err
	}

	pubKeysSigners := getPubKeySigners(consensusPubKeys, proofHandler.GetPubKeysBitmap())

	return multiSigVerifier.VerifyAggregatedSig(pubKeysSigners, proofHandler.GetHeaderHash(), proofHandler.GetAggregatedSignature())
}

func (hsv *HeaderSigVerifier) getPrevHeaderInfo(currentHeader data.HeaderHandler) (data.HeaderHandler, []byte, []byte, []byte, error) {
	previousProof := currentHeader.GetPreviousProof()

	var sig, bitmap []byte
	if previousProof != nil {
		sig, bitmap = previousProof.GetAggregatedSignature(), previousProof.GetPubKeysBitmap()
	}

	hash := currentHeader.GetPrevHash()
	prevHeader, err := hsv.headersPool.GetHeaderByHash(hash)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	headerCopy, err := hsv.copyHeaderWithoutSig(prevHeader)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	hash, err = core.CalculateHash(hsv.marshalizer, hsv.hasher, headerCopy)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return headerCopy, hash, sig, bitmap, nil
}

// VerifyPreviousBlockProof verifies if the structure of the header matches the expected structure in regards with the consensus flag
func (hsv *HeaderSigVerifier) VerifyPreviousBlockProof(header data.HeaderHandler) error {
	previousProof := header.GetPreviousProof()

	hasProof := false
	hasLeaderSignature := false

	if previousProof != nil {
		previousAggregatedSignature, previousBitmap := previousProof.GetAggregatedSignature(), previousProof.GetPubKeysBitmap()
		hasProof = len(previousAggregatedSignature) > 0 && len(previousBitmap) > 0

		if len(previousBitmap) > 0 {
			hasLeaderSignature = previousBitmap[0]&1 != 0
		}
	}

	isFlagEnabled := hsv.enableEpochsHandler.IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, header.GetEpoch())
	if isFlagEnabled && !hasProof {
		return fmt.Errorf("%w, received header without proof after flag activation", process.ErrInvalidHeader)
	}
	if !isFlagEnabled && hasProof {
		return fmt.Errorf("%w, received header with proof before flag activation", process.ErrInvalidHeader)
	}
	if isFlagEnabled && !hasLeaderSignature {
		return fmt.Errorf("%w, received header without leader signature after flag activation", process.ErrInvalidHeader)
	}

	return nil
}

func (hsv *HeaderSigVerifier) verifyConsensusSize(consensusPubKeys []string, header data.HeaderHandler, bitmap []byte) error {
	consensusSize := len(consensusPubKeys)

	expectedBitmapSize := consensusSize / 8
	if consensusSize%8 != 0 {
		expectedBitmapSize++
	}
	if len(bitmap) != expectedBitmapSize {
		log.Debug("wrong size bitmap",
			"expected number of bytes", expectedBitmapSize,
			"actual", len(bitmap))
		return ErrWrongSizeBitmap
	}

	numOfOnesInBitmap := 0
	for index := range bitmap {
		numOfOnesInBitmap += bits.OnesCount8(bitmap[index])
	}

	minNumRequiredSignatures := core.GetPBFTThreshold(consensusSize)
	if hsv.fallbackHeaderValidator.ShouldApplyFallbackValidation(header) {
		minNumRequiredSignatures = core.GetPBFTFallbackThreshold(consensusSize)
		log.Warn("HeaderSigVerifier.verifyConsensusSize: fallback validation has been applied",
			"minimum number of signatures required", minNumRequiredSignatures,
			"actual number of signatures in bitmap", numOfOnesInBitmap,
		)
	}

	if numOfOnesInBitmap >= minNumRequiredSignatures {
		return nil
	}

	log.Debug("not enough signatures",
		"minimum expected", minNumRequiredSignatures,
		"actual", numOfOnesInBitmap)

	return ErrNotEnoughSignatures
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
	randSeed := header.GetRandSeed()
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

	err = headerCopy.SetLeaderSignature(nil)
	if err != nil {
		return nil, err
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
