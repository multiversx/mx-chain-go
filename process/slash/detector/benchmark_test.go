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
	mock2 "github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	"github.com/ElrondNetwork/elrond-go/testscommon/slashMocks"
	"github.com/stretchr/testify/require"
)

func BenchmarkMultipleHeaderSigningDetector_ValidateProof(b *testing.B) {
	consensusGroupSize := uint16(400)
	numOfSigners := uint16(268) // 2/3 * 400 + 14 // MALICIOUS SIGNERS
	noOfHeaders := uint32(2)

	//	bitmapSize := consensusGroupSize/8 + 1
	// set bitmap to select all members
	//	bitmap := make([]byte, bitmapSize)
	//	byteMask := 0xFF
	//	for i := uint16(0); i < bitmapSize; i++ {
	//		bitmap[i] = byte((((1 << consensusGroupSize) - 1) >> i) & byteMask)
	//	}

	hashSize := 16
	hasher, _ := blake2b.NewBlake2bWithSize(hashSize)
	suite := mcl.NewSuiteBLS12()
	kg := signing.NewKeyGenerator(suite)
	pubKeysStr, blsSigners, privateKeys := createMultiSignersBls(consensusGroupSize, consensusGroupSize, hasher, kg)

	args := createAll(b, hasher, blsSigners[0], kg, pubKeysStr)
	ssd, err := detector.NewMultipleHeaderSigningDetector(args)
	require.Nil(b, err)
	// Worst case scenario: 25% of a consensus group of 400 validators on metachain signed 3 different headers
	slashRes := GenerateSlashResults(b, hasher, uint32(numOfSigners), privateKeys, noOfHeaders, blsSigners, pubKeysStr, args.NodesCoordinator)
	//	b.StopTimer()
	proof, err := coreSlash.NewMultipleSigningProof(slashRes)
	require.NotNil(b, proof)
	require.Nil(b, err)
	//	b.StartTimer()
	_ = ssd
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		//x := 1 + 1
		//fmt.Println(x)
		err = ssd.ValidateProof(proof)

		require.Nil(b, err)
	}

}

const (
	shuffleBetweenShards    = false
	adaptivity              = false
	hysteresis              = float32(0.2)
	defaultSelectionChances = uint32(1)
	eligiblePerShard        = 100
	waitingPerShard         = 30
)

func createAll(b *testing.B, hasher hashing.Hasher, multiSigVerifier crypto.MultiSigVerifier, keyGen crypto.KeyGenerator, pubKeys []string) *detector.MultipleHeaderSigningDetectorArgs {
	argsNodesCoordinator := createArguments(pubKeys)
	nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argsNodesCoordinator)
	require.Nil(b, err)

	headerSigVerifierArgs := headerCheck.ArgsHeaderSigVerifier{
		Marshalizer:             &marshal.GogoProtoMarshalizer{},
		Hasher:                  hasher,
		NodesCoordinator:        nodesCoordinator,
		MultiSigVerifier:        multiSigVerifier,
		SingleSigVerifier:       singlesig.NewBlsSigner(),
		KeyGen:                  keyGen,
		FallbackHeaderValidator: &testscommon.FallBackHeaderValidatorStub{},
	}
	headerSig, err := headerCheck.NewHeaderSigVerifier(&headerSigVerifierArgs)
	require.Nil(b, err)

	return &detector.MultipleHeaderSigningDetectorArgs{
		NodesCoordinator:  nodesCoordinator,
		RoundHandler:      &mock.RoundHandlerMock{},
		Hasher:            hasher,
		Marshaller:        &marshal.GogoProtoMarshalizer{},
		SlashingCache:     &slashMocks.RoundDetectorCacheStub{},
		RoundHashCache:    &slashMocks.HeadersCacheStub{},
		HeaderSigVerifier: headerSig,
	}
}

func createMultiSignersBls(
	numOfSigners uint16,
	grSize uint16,
	hasher hashing.Hasher,
	kg crypto.KeyGenerator,
) ([]string, []crypto.MultiSigner, []crypto.PrivateKey) {
	var pubKeyBytes []byte

	privKeys := make([]crypto.PrivateKey, grSize)
	pubKeys := make([]crypto.PublicKey, grSize)
	pubKeysStr := make([]string, grSize)

	for i := uint16(0); i < grSize; i++ {
		sk, pk := kg.GeneratePair()
		privKeys[i] = sk
		pubKeys[i] = pk

		pubKeyBytes, _ = pk.ToByteArray()
		pubKeysStr[i] = string(pubKeyBytes)
	}

	multiSigners := make([]crypto.MultiSigner, numOfSigners)
	llSigner := &llsig.BlsMultiSigner{Hasher: hasher}

	for i := uint16(0); i < numOfSigners; i++ {
		multiSigners[i], _ = multisig.NewBLSMultisig(llSigner, pubKeysStr, privKeys[i], kg, i)
	}

	return pubKeysStr, multiSigners, privKeys
}

func createArguments(pubKeys []string) sharding.ArgNodesCoordinator {
	nbShards := uint32(1)
	eligibleMap := createDummyNodesMap(400, nbShards, pubKeys)
	waitingMap := createDummyNodesMap(3, nbShards, pubKeys)
	shufflerArgs := &sharding.NodesShufflerArgs{
		NodesShard:           400,
		NodesMeta:            400,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, _ := sharding.NewHashValidatorsShuffler(shufflerArgs)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize:    63,
		MetaConsensusGroupSize:     400,
		Marshalizer:                &marshal.GogoProtoMarshalizer{},
		Hasher:                     &hashingMocks.HasherMock{},
		Shuffler:                   nodeShuffler,
		EpochStartNotifier:         epochStartSubscriber,
		BootStorer:                 bootStorer,
		NbShards:                   nbShards,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("test"),
		ConsensusGroupCache:        &mock2.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:         &mock2.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		IsFullArchive:              false,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
	return arguments
}

func createDummyNodesMap(nodesPerShard uint32, nbShards uint32, pubKeys []string) map[uint32][]sharding.Validator {
	nodesMap := make(map[uint32][]sharding.Validator)

	var shard uint32

	for i := uint32(0); i <= nbShards; i++ {
		shard = i
		if i == nbShards {
			shard = core.MetachainShardId
		}
		list := createDummyNodesList(nodesPerShard, pubKeys)
		nodesMap[shard] = list
	}

	return nodesMap
}

func createDummyNodesList(nbNodes uint32, pubKeys []string) []sharding.Validator {
	list := make([]sharding.Validator, 0)

	for j := uint32(0); j < nbNodes; j++ {
		list = append(list, mock2.NewValidatorMock([]byte(pubKeys[j]), 1, defaultSelectionChances))
	}

	return list
}
