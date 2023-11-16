package headerCheck

import (
	"math/bits"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	cryptoCommon "github.com/multiversx/mx-chain-go/common/crypto"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ process.InterceptedHeaderSigVerifier = (*HeaderSigVerifier)(nil)

var log = logger.GetOrCreate("process/headerCheck")

// ArgsHeaderSigVerifier is used to store all components that are needed to create a new HeaderSigVerifier
type ArgsHeaderSigVerifier struct {
	Marshalizer                  marshal.Marshalizer
	Hasher                       hashing.Hasher
	NodesCoordinator             nodesCoordinator.NodesCoordinator
	MultiSigContainer            cryptoCommon.MultiSignerContainer
	SingleSigVerifier            crypto.SingleSigner
	KeyGen                       crypto.KeyGenerator
	FallbackHeaderValidator      process.FallbackHeaderValidator
	ExtraHeaderSigVerifierHolder ExtraHeaderSigVerifierHolder
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

	extraSigVerifierHolder ExtraHeaderSigVerifierHolder
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
		extraSigVerifierHolder:  arguments.ExtraHeaderSigVerifierHolder,
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
	if check.IfNil(arguments.ExtraHeaderSigVerifierHolder) {
		return errors.ErrNilExtraHeaderSigVerifierHolder
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

func (hsv *HeaderSigVerifier) getConsensusSigners(header data.HeaderHandler) ([][]byte, error) {
	randSeed := header.GetPrevRandSeed()
	bitmap := header.GetPubKeysBitmap()
	if len(bitmap) == 0 {
		return nil, process.ErrNilPubKeysBitmap
	}
	if bitmap[0]&1 == 0 {
		return nil, process.ErrBlockProposerSignatureMissing
	}

	// TODO: remove if start of epochForConsensus block needs to be validated by the new epochForConsensus nodes
	epochForConsensus := header.GetEpoch()
	if header.IsStartOfEpochBlock() && epochForConsensus > 0 {
		epochForConsensus = epochForConsensus - 1
	}

	consensusPubKeys, err := hsv.nodesCoordinator.GetConsensusValidatorsPublicKeys(
		randSeed,
		header.GetRound(),
		header.GetShardID(),
		epochForConsensus,
	)
	if err != nil {
		return nil, err
	}

	err = hsv.verifyConsensusSize(consensusPubKeys, header)
	if err != nil {
		return nil, err
	}

	pubKeysSigners := make([][]byte, 0, len(consensusPubKeys))
	for i := range consensusPubKeys {
		err = isIndexInBitmap(uint16(i), bitmap)
		if err != nil {
			continue
		}
		pubKeysSigners = append(pubKeysSigners, []byte(consensusPubKeys[i]))
	}

	return pubKeysSigners, nil
}

// VerifySignature will check if signature is correct
func (hsv *HeaderSigVerifier) VerifySignature(header data.HeaderHandler) error {
	multiSigVerifier, err := hsv.multiSigContainer.GetMultiSigner(header.GetEpoch())
	if err != nil {
		return err
	}

	headerCopy, err := hsv.copyHeaderWithoutSig(header)
	if err != nil {
		return err
	}

	hash, err := core.CalculateHash(hsv.marshalizer, hsv.hasher, headerCopy)
	if err != nil {
		return err
	}

	pubKeysSigners, err := hsv.getConsensusSigners(header)
	if err != nil {
		return err
	}

	err = multiSigVerifier.VerifyAggregatedSig(pubKeysSigners, hash, header.GetSignature())
	if err != nil {
		return err
	}

	return hsv.extraSigVerifierHolder.VerifyAggregatedSignature(header, multiSigVerifier, pubKeysSigners)
}

func (hsv *HeaderSigVerifier) verifyConsensusSize(consensusPubKeys []string, header data.HeaderHandler) error {
	consensusSize := len(consensusPubKeys)
	bitmap := header.GetPubKeysBitmap()

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

	err = hsv.singleSigVerifier.Verify(leaderPubKey, headerBytes, header.GetLeaderSignature())
	if err != nil {
		return err
	}

	return hsv.extraSigVerifierHolder.VerifyLeaderSignature(header, leaderPubKey)
}

func (hsv *HeaderSigVerifier) getLeader(header data.HeaderHandler) (crypto.PublicKey, error) {
	prevRandSeed := header.GetPrevRandSeed()

	// TODO: remove if start of epoch block needs to be validated by the new epoch nodes
	epoch := header.GetEpoch()
	if header.IsStartOfEpochBlock() && epoch > 0 {
		epoch = epoch - 1
	}

	headerConsensusGroup, err := hsv.nodesCoordinator.ComputeConsensusGroup(prevRandSeed, header.GetRound(), header.GetShardID(), epoch)
	if err != nil {
		return nil, err
	}

	leaderPubKeyValidator := headerConsensusGroup[0]
	return hsv.keyGen.PublicKeyFromByteArray(leaderPubKeyValidator.PubKey())
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

	err = hsv.extraSigVerifierHolder.RemoveAllSignatures(headerCopy)
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

	err = hsv.extraSigVerifierHolder.RemoveLeaderSignature(headerCopy)
	if err != nil {
		return nil, err
	}

	return headerCopy, nil
}
