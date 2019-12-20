package headerCheck

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("process/headerCheck")

// ArgsHeaderSigVerifier is used to store all components that are needed to create a new HeaderSigVerifier
type ArgsHeaderSigVerifier struct {
	Marshalizer       marshal.Marshalizer
	Hasher            hashing.Hasher
	NodesCoordinator  sharding.NodesCoordinator
	MultiSigVerifier  crypto.MultiSigVerifier
	SingleSigVerifier crypto.SingleSigner
	KeyGen            crypto.KeyGenerator
}

//HeaderSigVerifier is component used to check if a header is valid
type HeaderSigVerifier struct {
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	nodesCoordinator  sharding.NodesCoordinator
	multiSigVerifier  crypto.MultiSigVerifier
	singleSigVerifier crypto.SingleSigner
	keyGen            crypto.KeyGenerator
}

// NewHeaderSigVerifier will create a new instance of HeaderSigVerifier
func NewHeaderSigVerifier(arguments *ArgsHeaderSigVerifier) (*HeaderSigVerifier, error) {
	err := checkArgsHeaderSigVerifier(arguments)
	if err != nil {
		return nil, err
	}

	return &HeaderSigVerifier{
		marshalizer:       arguments.Marshalizer,
		hasher:            arguments.Hasher,
		nodesCoordinator:  arguments.NodesCoordinator,
		multiSigVerifier:  arguments.MultiSigVerifier,
		singleSigVerifier: arguments.SingleSigVerifier,
		keyGen:            arguments.KeyGen,
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
	if check.IfNil(arguments.MultiSigVerifier) {
		return process.ErrNilMultiSigVerifier
	}
	if check.IfNil(arguments.NodesCoordinator) {
		return process.ErrNilNodesCoordinator
	}
	if check.IfNil(arguments.SingleSigVerifier) {
		return process.ErrNilSingleSigner
	}

	return nil
}

// VerifySignature will check if signature is correct
func (hsv *HeaderSigVerifier) VerifySignature(header data.HeaderHandler) error {

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
	headerCopy := hsv.copyHeaderWithoutLeaderSig(header)
	headerBytes, err := hsv.marshalizer.Marshal(&headerCopy)
	if err != nil {
		return err
	}

	return hsv.singleSigVerifier.Verify(leaderPubKey, headerBytes, header.GetLeaderSignature())
}

func (hsv *HeaderSigVerifier) getLeader(header data.HeaderHandler) (crypto.PublicKey, error) {
	prevRandSeed := header.GetPrevRandSeed()
	headerConsensusGroup, err := hsv.nodesCoordinator.ComputeValidatorsGroup(prevRandSeed, header.GetRound(), header.GetShardID())
	if err != nil {
		return nil, err
	}

	leaderPubKeyValidator := headerConsensusGroup[0]
	return hsv.keyGen.PublicKeyFromByteArray(leaderPubKeyValidator.PubKey())
}

func (hsv *HeaderSigVerifier) copyHeaderWithoutSig(header data.HeaderHandler) data.HeaderHandler {
	headerCopy := header.Clone()
	headerCopy.SetSignature(nil)
	headerCopy.SetPubKeysBitmap(nil)
	headerCopy.SetLeaderSignature(nil)
	if header.GetShardID() == sharding.MetachainShardId {
		headerCopy.SetValidatorStatsRootHash(nil)
	}

	return headerCopy
}

func (hsv *HeaderSigVerifier) copyHeaderWithoutLeaderSig(header data.HeaderHandler) data.HeaderHandler {
	headerCopy := header.Clone()
	headerCopy.SetLeaderSignature(nil)

	return headerCopy
}
