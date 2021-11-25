package detector_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl"
	llsig "github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/multisig"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go/process/headerCheck"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/slash/detector"
	"github.com/ElrondNetwork/elrond-go/sharding"
	shardingMock "github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	"github.com/ElrondNetwork/elrond-go/testscommon/slashMocks"
	"github.com/stretchr/testify/require"
)

var marshaller = &marshal.GogoProtoMarshalizer{}

const (
	shuffleBetweenShards    = false
	adaptivity              = false
	hysteresis              = float32(0.2)
	defaultSelectionChances = uint32(1)
	hashSize                = 16
	metaConsensusGroupSize  = 100
	shardConsensusGroupSize = 63
	leaderGroupIndex        = 0
)

func BenchmarkMultipleHeaderSigningDetector_ValidateProof(b *testing.B) {
	hasher, err := blake2b.NewBlake2bWithSize(hashSize)
	require.Nil(b, err)

	blsSuite := mcl.NewSuiteBLS12()
	keyGenerator := signing.NewKeyGenerator(blsSuite)
	multiSigData := createMultiSignersBls(metaConsensusGroupSize, hasher, keyGenerator)

	args := createHeaderSigningDetectorArgs(b, hasher, keyGenerator, multiSigData)
	ssd, err := detector.NewMultipleHeaderSigningDetector(args)
	require.NotNil(b, ssd)
	require.Nil(b, err)

	// Worst case scenario: 25% * 400() + 1
	noOfMaliciousSigners := uint16(30)
	noOfSignedHeaders := uint32(2)
	slashRes := GenerateSlashResults(b, hasher, uint32(noOfMaliciousSigners), noOfSignedHeaders, args.NodesCoordinator, multiSigData)
	proof, err := coreSlash.NewMultipleSigningProof(slashRes)
	require.NotNil(b, proof)
	require.Nil(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err = ssd.ValidateProof(proof)
		require.Nil(b, err)
	}

}

type multiSignerData struct {
	multiSigner crypto.MultiSigner
	privateKey  crypto.PrivateKey
}

func createMultiSignersBls(
	noOfSigners uint16,
	hasher hashing.Hasher,
	keyGenerator crypto.KeyGenerator,
) map[string]multiSignerData {
	privateKeys := make([]crypto.PrivateKey, noOfSigners)
	pubKeysStr := make([]string, noOfSigners)
	for i := uint16(0); i < noOfSigners; i++ {
		sk, pk := keyGenerator.GeneratePair()
		privateKeys[i] = sk

		pubKeyBytes, _ := pk.ToByteArray()
		pubKeysStr[i] = string(pubKeyBytes)
	}

	multiSigners := make([]crypto.MultiSigner, noOfSigners)
	llSigner := &llsig.BlsMultiSigner{Hasher: hasher}
	for i := uint16(0); i < noOfSigners; i++ {
		multiSigners[i], _ = multisig.NewBLSMultisig(llSigner, pubKeysStr, privateKeys[i], keyGenerator, i)
	}

	allMultiSigData := make(map[string]multiSignerData)
	for i, pubKey := range pubKeysStr {
		allMultiSigData[pubKey] = multiSignerData{
			multiSigner: multiSigners[i],
			privateKey:  privateKeys[i],
		}
	}

	return allMultiSigData
}

func createHeaderSigningDetectorArgs(
	b *testing.B,
	hasher hashing.Hasher,
	keyGenerator crypto.KeyGenerator,
	multiSignersData map[string]multiSignerData,
) *detector.MultipleHeaderSigningDetectorArgs {
	var multiSigVerifier crypto.MultiSigVerifier
	for pubKey := range multiSignersData {
		multiSigVerifier = multiSignersData[pubKey].multiSigner
		break
	}

	pubKeys := make([]string, 0, len(multiSignersData))
	for pubKey := range multiSignersData {
		pubKeys = append(pubKeys, pubKey)
	}

	nodesCoordinatorArgs := createNodesCoordinatorArgs(hasher, pubKeys)
	nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(nodesCoordinatorArgs)
	require.Nil(b, err)

	headerSigVerifierArgs := createHeaderSigVerifierArgs(hasher, multiSigVerifier, keyGenerator, nodesCoordinator)
	headerSigVerifier, err := headerCheck.NewHeaderSigVerifier(&headerSigVerifierArgs)
	require.Nil(b, err)

	return &detector.MultipleHeaderSigningDetectorArgs{
		NodesCoordinator:  nodesCoordinator,
		RoundHandler:      &mock.RoundHandlerMock{},
		Hasher:            hasher,
		Marshaller:        &marshal.GogoProtoMarshalizer{},
		SlashingCache:     &slashMocks.RoundDetectorCacheStub{},
		RoundHashCache:    &slashMocks.HeadersCacheStub{},
		HeaderSigVerifier: headerSigVerifier,
	}
}
func createHeaderSigVerifierArgs(
	hasher hashing.Hasher,
	multiSigVerifier crypto.MultiSigVerifier,
	keyGenerator crypto.KeyGenerator,
	nodesCoordinator sharding.NodesCoordinator,
) headerCheck.ArgsHeaderSigVerifier {
	return headerCheck.ArgsHeaderSigVerifier{
		Marshalizer:             marshaller,
		Hasher:                  hasher,
		NodesCoordinator:        nodesCoordinator,
		MultiSigVerifier:        multiSigVerifier,
		SingleSigVerifier:       singlesig.NewBlsSigner(),
		KeyGen:                  keyGenerator,
		FallbackHeaderValidator: &testscommon.FallBackHeaderValidatorStub{},
	}
}

func createNodesCoordinatorArgs(hasher hashing.Hasher, pubKeys []string) sharding.ArgNodesCoordinator {
	noOfShards := uint32(1)
	eligibleMap := createShardValidatorMap(metaConsensusGroupSize, noOfShards, pubKeys)
	waitingMap := createShardValidatorMap(1, noOfShards, pubKeys)
	nodeShuffler := createNodesShuffler()

	return sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize:    shardConsensusGroupSize,
		MetaConsensusGroupSize:     metaConsensusGroupSize,
		Marshalizer:                marshaller,
		Hasher:                     hasher,
		Shuffler:                   nodeShuffler,
		EpochStartNotifier:         &mock.EpochStartNotifierStub{},
		BootStorer:                 mock.NewStorerMock(),
		NbShards:                   noOfShards,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("test"),
		ConsensusGroupCache:        &shardingMock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:         &shardingMock.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		IsFullArchive:              false,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
}

func createNodesShuffler() sharding.NodesShuffler {
	shufflerArgs := &sharding.NodesShufflerArgs{
		NodesShard:           metaConsensusGroupSize,
		NodesMeta:            metaConsensusGroupSize,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}

	nodeShuffler, _ := sharding.NewHashValidatorsShuffler(shufflerArgs)
	return nodeShuffler
}

func createShardValidatorMap(validatorsPerShard uint32, noOfShards uint32, pubKeys []string) map[uint32][]sharding.Validator {
	shardValidatorsMap := make(map[uint32][]sharding.Validator)
	for i := uint32(0); i <= noOfShards; i++ {
		shard := i
		if i == noOfShards {
			shard = core.MetachainShardId
		}
		validators := createValidatorList(validatorsPerShard, pubKeys)
		shardValidatorsMap[shard] = validators
	}

	return shardValidatorsMap
}

func createValidatorList(nbNodes uint32, pubKeys []string) []sharding.Validator {
	validators := make([]sharding.Validator, 0)

	for i := uint32(0); i < nbNodes; i++ {
		validator := shardingMock.NewValidatorMock([]byte(pubKeys[i]), 1, defaultSelectionChances)
		validators = append(validators, validator)
	}

	return validators
}
