package detector_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	llsig "github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/multisig"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/multisig"
	mockEpochStart "github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/headerCheck"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/process/slash/detector"
	"github.com/ElrondNetwork/elrond-go/sharding"
	shardingMock "github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	"github.com/ElrondNetwork/elrond-go/testscommon/slashMocks"
	"github.com/stretchr/testify/require"
)

var validatorPubKey = []byte("validator pub key")
var marshaller = &marshal.GogoProtoMarshalizer{}

const (
	shuffleBetweenShards    = false
	adaptivity              = false
	hysteresis              = float32(0.2)
	defaultSelectionChances = uint32(1)
	hashSize                = 16
	metaConsensusGroupSize  = 400
	shardConsensusGroupSize = 63
	leaderGroupIndex        = 0
	cacheSize               = 3
	// Worst case scenario: 1/4 * metaConsensusGroupSize + 1
	maxNoOfMaliciousValidatorsOnMetaChain = uint32(metaConsensusGroupSize/4 + 1)
)

func generateMockMultipleHeaderDetectorArgs() detector.MultipleHeaderDetectorArgs {
	nodesCoordinator := &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			validator := mock.NewValidatorMock(validatorPubKey)
			return []sharding.Validator{validator}, nil
		},
	}

	return detector.MultipleHeaderDetectorArgs{
		NodesCoordinator:  nodesCoordinator,
		RoundHandler:      &mock.RoundHandlerMock{},
		SlashingCache:     &slashMocks.RoundDetectorCacheStub{},
		Hasher:            &hashingMocks.HasherMock{},
		Marshaller:        &testscommon.MarshalizerMock{},
		HeaderSigVerifier: &mock.HeaderSigVerifierStub{},
	}
}

func createInterceptedHeaders(headersInfo []data.HeaderInfoHandler) []process.InterceptedHeader {
	interceptedHeaders := make([]process.InterceptedHeader, 0)

	for _, headerInfo := range headersInfo {
		hash := headerInfo.GetHash()
		header := headerInfo.GetHeaderHandler()

		interceptedData := testscommon.InterceptedDataStub{
			HashCalled: func() []byte {
				return hash
			},
		}
		interceptedHeader := &testscommon.InterceptedHeaderStub{
			InterceptedDataStub: interceptedData,
			HeaderHandlerCalled: func() data.HeaderHandler {
				return header
			},
		}
		interceptedHeaders = append(interceptedHeaders, interceptedHeader)
	}

	return interceptedHeaders
}

func verifyInterceptedHeaders(b *testing.B, interceptedHeaders []process.InterceptedHeader, ssd slash.SlashingDetector) {
	var proof coreSlash.SlashingProofHandler
	var err error

	for idx, interceptedHeader := range interceptedHeaders {
		if idx == 0 {
			proof, err = ssd.VerifyData(interceptedHeader)
			require.Equal(b, process.ErrNoSlashingEventDetected, err)
			require.Nil(b, proof)
			continue
		}
		proof, err = ssd.VerifyData(interceptedHeader)
		require.Nil(b, err)
		require.NotNil(b, proof)
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

func createMultipleHeaderDetectorArgs(
	b *testing.B,
	hasher hashing.Hasher,
	keyGenerator crypto.KeyGenerator,
	multiSignersData map[string]multiSignerData,
) detector.MultipleHeaderDetectorArgs {
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

	return detector.MultipleHeaderDetectorArgs{
		NodesCoordinator:  nodesCoordinator,
		RoundHandler:      &mock.RoundHandlerMock{},
		Hasher:            hasher,
		Marshaller:        &marshal.GogoProtoMarshalizer{},
		SlashingCache:     detector.NewRoundValidatorHeaderCache(cacheSize),
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
	consensusGroupCache, _ := lrucache.NewCache(1000)

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
		ConsensusGroupCache:        consensusGroupCache,
		ShuffledOutHandler:         &shardingMock.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		IsFullArchive:              false,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
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
		pubKey := []byte(pubKeys[i])
		validator := shardingMock.NewValidatorMock(pubKey, defaultSelectionChances, defaultSelectionChances)
		validators = append(validators, validator)
	}

	return validators
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
