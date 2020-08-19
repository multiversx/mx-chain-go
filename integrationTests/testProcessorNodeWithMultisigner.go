package integrationTests

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/forking"
	"github.com/ElrondNetwork/elrond-go/crypto/peerSignatureHandler"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/testscommon"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	mclmultisig "github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/multisig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/headerCheck"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

// NewTestProcessorNodeWithCustomNodesCoordinator returns a new TestProcessorNode instance with custom NodesCoordinator
func NewTestProcessorNodeWithCustomNodesCoordinator(
	maxShards uint32,
	nodeShardId uint32,
	initialNodeAddr string,
	epochStartNotifier notifier.EpochStartNotifier,
	nodesCoordinator sharding.NodesCoordinator,
	ratingsData *rating.RatingsData,
	cp *CryptoParams,
	keyIndex int,
	ownAccount *TestWalletAccount,
	headerSigVerifier process.InterceptedHeaderSigVerifier,
	headerVersioning process.HeaderVersioningHandler,
	nodeSetup sharding.GenesisNodesSetupHandler,
) *TestProcessorNode {

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	messenger := CreateMessengerWithKadDht(initialNodeAddr)
	tpn := &TestProcessorNode{
		ShardCoordinator:      shardCoordinator,
		Messenger:             messenger,
		NodesCoordinator:      nodesCoordinator,
		HeaderSigVerifier:     headerSigVerifier,
		HeaderVersioning:      headerVersioning,
		ChainID:               ChainID,
		NodesSetup:            nodeSetup,
		RatingsData:           ratingsData,
		MinTransactionVersion: MinTransactionVersion,
		HistoryRepository:     &mock.HistoryRepositoryStub{},
		EpochNotifier:         forking.NewGenericEpochNotifier(),
	}

	tpn.NodeKeys = cp.Keys[nodeShardId][keyIndex]
	blsHasher := &blake2b.Blake2b{HashSize: hashing.BlsHashSize}
	llsig := &mclmultisig.BlsMultiSigner{Hasher: blsHasher}

	pubKeysMap := PubKeysMapFromKeysMap(cp.Keys)

	tpn.MultiSigner, _ = multisig.NewBLSMultisig(
		llsig,
		pubKeysMap[nodeShardId],
		tpn.NodeKeys.Sk,
		cp.KeyGen,
		0,
	)
	if tpn.MultiSigner == nil {
		fmt.Println("Error generating multisigner")
	}
	accountShardId := nodeShardId
	if nodeShardId == core.MetachainShardId {
		accountShardId = 0
	}

	if ownAccount == nil {
		tpn.OwnAccount = CreateTestWalletAccount(shardCoordinator, accountShardId)
	} else {
		tpn.OwnAccount = ownAccount
	}

	tpn.EpochStartNotifier = epochStartNotifier
	tpn.initDataPools()
	tpn.initTestNode()

	return tpn
}

// CreateNodesWithNodesCoordinator returns a map with nodes per shard each using a real nodes coordinator
func CreateNodesWithNodesCoordinator(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	seedAddress string,
) map[uint32][]*TestProcessorNode {
	return CreateNodesWithNodesCoordinatorWithCacher(nodesPerShard, nbMetaNodes, nbShards, shardConsensusGroupSize, metaConsensusGroupSize, seedAddress)
}

// CreateNodesWithNodesCoordinatorWithCacher returns a map with nodes per shard each using a real nodes coordinator with cacher
func CreateNodesWithNodesCoordinatorWithCacher(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	seedAddress string,
) map[uint32][]*TestProcessorNode {
	coordinatorFactory := &IndexHashedNodesCoordinatorFactory{}
	return CreateNodesWithNodesCoordinatorFactory(nodesPerShard, nbMetaNodes, nbShards, shardConsensusGroupSize, metaConsensusGroupSize, seedAddress, coordinatorFactory)

}

// CreateNodesWithNodesCoordinatorAndTxKeys -
func CreateNodesWithNodesCoordinatorAndTxKeys(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	seedAddress string,
) map[uint32][]*TestProcessorNode {
	rater, _ := rating.NewBlockSigningRater(CreateRatingsData())
	coordinatorFactory := &IndexHashedNodesCoordinatorWithRaterFactory{
		PeerAccountListAndRatingHandler: rater,
	}
	cp := CreateCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards))
	blsPubKeys := PubKeysMapFromKeysMap(cp.Keys)
	txPubKeys := PubKeysMapFromKeysMap(cp.TxKeys)
	validatorsMap := GenValidatorsFromPubKeysAndTxPubKeys(blsPubKeys, txPubKeys)
	validatorsMapForNodesCoordinator, _ := sharding.NodesInfoToValidators(validatorsMap)

	waitingMap := make(map[uint32][]sharding.GenesisNodeInfoHandler)
	for i := 0; i < nbShards; i++ {
		waitingMap[uint32(i)] = make([]sharding.GenesisNodeInfoHandler, 0)
	}
	waitingMap[core.MetachainShardId] = make([]sharding.GenesisNodeInfoHandler, 0)

	waitingMapForNodesCoordinator := make(map[uint32][]sharding.Validator)
	for i := 0; i < nbShards; i++ {
		waitingMapForNodesCoordinator[uint32(i)] = make([]sharding.Validator, 0)
	}
	waitingMapForNodesCoordinator[core.MetachainShardId] = make([]sharding.Validator, 0)

	nodesSetup := &mock.NodesSetupStub{InitialNodesInfoCalled: func() (m map[uint32][]sharding.GenesisNodeInfoHandler, m2 map[uint32][]sharding.GenesisNodeInfoHandler) {
		return validatorsMap, waitingMap
	}}

	nodesMap := make(map[uint32][]*TestProcessorNode)

	for shardId, validatorList := range validatorsMap {
		nodesList := make([]*TestProcessorNode, len(validatorList))

		for i := range validatorList {
			dataCache, _ := lrucache.NewCache(10000)
			nodesList[i] = CreateNodeWithBLSAndTxKeys(
				nodesPerShard,
				nbMetaNodes,
				shardConsensusGroupSize,
				metaConsensusGroupSize,
				shardId,
				nbShards,
				validatorsMapForNodesCoordinator,
				waitingMapForNodesCoordinator,
				i,
				seedAddress,
				cp,
				dataCache,
				coordinatorFactory,
				nodesSetup,
				nil,
			)
		}

		nodesMap[shardId] = append(nodesMap[shardId], nodesList...)
	}

	return nodesMap
}

// CreateNodeWithBLSAndTxKeys -
func CreateNodeWithBLSAndTxKeys(
	nodesPerShard int,
	nbMetaNodes int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	shardId uint32,
	nbShards int,
	validatorsMap map[uint32][]sharding.Validator,
	waitingMap map[uint32][]sharding.Validator,
	keyIndex int,
	seedAddress string,
	cp *CryptoParams,
	cache sharding.Cacher,
	coordinatorFactory NodesCoordinatorFactory,
	nodesSetup sharding.GenesisNodesSetupHandler,
	ratingsData *rating.RatingsData,
) *TestProcessorNode {

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := CreateMemUnit()
	argFactory := ArgIndexHashedNodesCoordinatorFactory{
		nodesPerShard:           nodesPerShard,
		nbMetaNodes:             nbMetaNodes,
		shardConsensusGroupSize: shardConsensusGroupSize,
		metaConsensusGroupSize:  metaConsensusGroupSize,
		shardId:                 shardId,
		nbShards:                nbShards,
		validatorsMap:           validatorsMap,
		waitingMap:              waitingMap,
		keyIndex:                keyIndex,
		cp:                      cp,
		epochStartSubscriber:    epochStartSubscriber,
		hasher:                  TestHasher,
		consensusGroupCache:     cache,
		bootStorer:              bootStorer,
	}
	nodesCoordinator := coordinatorFactory.CreateNodesCoordinator(argFactory)

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(nbShards), shardId)

	messenger := CreateMessengerWithKadDht(seedAddress)
	tpn := &TestProcessorNode{
		ShardCoordinator:      shardCoordinator,
		Messenger:             messenger,
		NodesCoordinator:      nodesCoordinator,
		HeaderSigVerifier:     &mock.HeaderSigVerifierStub{},
		HeaderVersioning:      CreateHeaderVersioning(),
		ChainID:               ChainID,
		NodesSetup:            nodesSetup,
		RatingsData:           ratingsData,
		MinTransactionVersion: MinTransactionVersion,
		HistoryRepository:     &mock.HistoryRepositoryStub{},
		EpochNotifier:         forking.NewGenericEpochNotifier(),
	}

	tpn.NodeKeys = cp.Keys[shardId][keyIndex]
	blsHasher := &blake2b.Blake2b{HashSize: hashing.BlsHashSize}
	llsig := &mclmultisig.BlsMultiSigner{Hasher: blsHasher}

	pubKeysMap := PubKeysMapFromKeysMap(cp.Keys)

	tpn.MultiSigner, _ = multisig.NewBLSMultisig(
		llsig,
		pubKeysMap[shardId],
		tpn.NodeKeys.Sk,
		cp.KeyGen,
		0,
	)
	if tpn.MultiSigner == nil {
		fmt.Println("Error generating multisigner")
	}
	twa := &TestWalletAccount{}
	twa.SingleSigner = cp.SingleSigner
	twa.BlockSingleSigner = &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			return nil
		},
	}
	sk := cp.TxKeys[shardId][keyIndex].Sk
	pk := cp.TxKeys[shardId][keyIndex].Pk
	keyGen := cp.TxKeyGen

	pkBuff, _ := pk.ToByteArray()
	fmt.Printf("Found pk: %s in shard %d\n", hex.EncodeToString(pkBuff), shardId)

	twa.SkTxSign = sk
	twa.PkTxSign = pk
	twa.PkTxSignBytes, _ = pk.ToByteArray()
	twa.KeygenTxSign = keyGen
	twa.KeygenBlockSign = &mock.KeyGenMock{}
	twa.Address = twa.PkTxSignBytes

	peerSigCache, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 1000})
	twa.PeerSigHandler, _ = peerSignatureHandler.NewPeerSignatureHandler(peerSigCache, twa.SingleSigner, keyGen)
	tpn.OwnAccount = twa

	tpn.EpochStartNotifier = epochStartSubscriber
	tpn.initDataPools()
	tpn.initTestNode()

	return tpn
}

// CreateNodesWithNodesCoordinatorFactory returns a map with nodes per shard each using a real nodes coordinator
func CreateNodesWithNodesCoordinatorFactory(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	seedAddress string,
	nodesCoordinatorFactory NodesCoordinatorFactory,
) map[uint32][]*TestProcessorNode {
	cp := CreateCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards))
	pubKeys := PubKeysMapFromKeysMap(cp.Keys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, uint32(nbShards))
	validatorsMapForNodesCoordinator, _ := sharding.NodesInfoToValidators(validatorsMap)

	cpWaiting := CreateCryptoParams(1, 1, uint32(nbShards))
	pubKeysWaiting := PubKeysMapFromKeysMap(cpWaiting.Keys)
	waitingMap := GenValidatorsFromPubKeys(pubKeysWaiting, uint32(nbShards))
	waitingMapForNodesCoordinator, _ := sharding.NodesInfoToValidators(waitingMap)

	numNodes := nbShards*nodesPerShard + nbMetaNodes

	nodesSetup := &mock.NodesSetupStub{
		InitialNodesInfoCalled: func() (m map[uint32][]sharding.GenesisNodeInfoHandler, m2 map[uint32][]sharding.GenesisNodeInfoHandler) {
			return validatorsMap, waitingMap
		},
		MinNumberOfNodesCalled: func() uint32 {
			return uint32(numNodes)
		},
	}

	nodesMap := make(map[uint32][]*TestProcessorNode)

	for shardId, validatorList := range validatorsMap {
		nodesList := make([]*TestProcessorNode, len(validatorList))
		nodesListWaiting := make([]*TestProcessorNode, len(waitingMap[shardId]))

		for i := range validatorList {
			dataCache, _ := lrucache.NewCache(10000)
			nodesList[i] = CreateNode(
				nodesPerShard,
				nbMetaNodes,
				shardConsensusGroupSize,
				metaConsensusGroupSize,
				shardId,
				nbShards,
				validatorsMapForNodesCoordinator,
				waitingMapForNodesCoordinator,
				i,
				seedAddress,
				cp,
				dataCache,
				nodesCoordinatorFactory,
				nodesSetup,
				nil,
			)
		}

		for i := range waitingMap[shardId] {
			dataCache, _ := lrucache.NewCache(10000)
			nodesListWaiting[i] = CreateNode(
				nodesPerShard,
				nbMetaNodes,
				shardConsensusGroupSize,
				metaConsensusGroupSize,
				shardId,
				nbShards,
				validatorsMapForNodesCoordinator,
				waitingMapForNodesCoordinator,
				i,
				seedAddress,
				cpWaiting,
				dataCache,
				nodesCoordinatorFactory,
				nodesSetup,
				nil,
			)
		}

		nodesMap[shardId] = append(nodesList, nodesListWaiting...)
	}

	return nodesMap
}

// CreateNode -
func CreateNode(
	nodesPerShard int,
	nbMetaNodes int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	shardId uint32,
	nbShards int,
	validatorsMap map[uint32][]sharding.Validator,
	waitingMap map[uint32][]sharding.Validator,
	keyIndex int,
	seedAddress string,
	cp *CryptoParams,
	cache sharding.Cacher,
	coordinatorFactory NodesCoordinatorFactory,
	nodesSetup sharding.GenesisNodesSetupHandler,
	ratingsData *rating.RatingsData,
) *TestProcessorNode {

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := CreateMemUnit()

	argFactory := ArgIndexHashedNodesCoordinatorFactory{
		nodesPerShard,
		nbMetaNodes,
		shardConsensusGroupSize,
		metaConsensusGroupSize,
		shardId,
		nbShards,
		validatorsMap,
		waitingMap,
		keyIndex,
		cp,
		epochStartSubscriber,
		TestHasher,
		cache,
		bootStorer,
	}
	nodesCoordinator := coordinatorFactory.CreateNodesCoordinator(argFactory)

	return NewTestProcessorNodeWithCustomNodesCoordinator(
		uint32(nbShards),
		shardId,
		seedAddress,
		epochStartSubscriber,
		nodesCoordinator,
		ratingsData,
		cp,
		keyIndex,
		nil,
		&mock.HeaderSigVerifierStub{},
		CreateHeaderVersioning(),
		nodesSetup,
	)
}

func createVersioningHandler() process.HeaderVersioningHandler {
	headerVersioning, _ := headerCheck.NewHeaderVersioningHandler(
		ChainID,
		[]config.VersionByEpochs{
			{
				StartEpoch: 0,
				Version:    "*",
			},
		},
		"default",
		testscommon.NewCacherMock(),
	)

	return headerVersioning
}

// CreateNodesWithNodesCoordinatorAndHeaderSigVerifier returns a map with nodes per shard each using a real nodes coordinator and header sig verifier
func CreateNodesWithNodesCoordinatorAndHeaderSigVerifier(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	seedAddress string,
	signer crypto.SingleSigner,
	keyGen crypto.KeyGenerator,
) map[uint32][]*TestProcessorNode {
	cp := CreateCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards))
	pubKeys := PubKeysMapFromKeysMap(cp.Keys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, uint32(nbShards))
	validatorsMapForNodesCoordinator, _ := sharding.NodesInfoToValidators(validatorsMap)
	nodesMap := make(map[uint32][]*TestProcessorNode)
	nodeShuffler := sharding.NewHashValidatorsShuffler(
		uint32(nodesPerShard),
		uint32(nbMetaNodes),
		hysteresis,
		adaptivity,
		shuffleBetweenShards,
	)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := CreateMemUnit()

	nodesSetup := &mock.NodesSetupStub{InitialNodesInfoCalled: func() (m map[uint32][]sharding.GenesisNodeInfoHandler, m2 map[uint32][]sharding.GenesisNodeInfoHandler) {
		return validatorsMap, nil
	}}

	for shardId, validatorList := range validatorsMap {
		consensusCache, _ := lrucache.NewCache(10000)
		argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
			ShardConsensusGroupSize: shardConsensusGroupSize,
			MetaConsensusGroupSize:  metaConsensusGroupSize,
			Marshalizer:             TestMarshalizer,
			Hasher:                  TestHasher,
			Shuffler:                nodeShuffler,
			BootStorer:              bootStorer,
			EpochStartNotifier:      epochStartSubscriber,
			ShardIDAsObserver:       shardId,
			NbShards:                uint32(nbShards),
			EligibleNodes:           validatorsMapForNodesCoordinator,
			WaitingNodes:            make(map[uint32][]sharding.Validator),
			SelfPublicKey:           []byte(strconv.Itoa(int(shardId))),
			ConsensusGroupCache:     consensusCache,
			ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
		}
		nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)

		if err != nil {
			fmt.Println("Error creating node coordinator: " + err.Error())
		}

		nodesList := make([]*TestProcessorNode, len(validatorList))
		args := headerCheck.ArgsHeaderSigVerifier{
			Marshalizer:       TestMarshalizer,
			Hasher:            TestHasher,
			NodesCoordinator:  nodesCoordinator,
			MultiSigVerifier:  TestMultiSig,
			SingleSigVerifier: signer,
			KeyGen:            keyGen,
		}
		headerSig, _ := headerCheck.NewHeaderSigVerifier(&args)

		for i := range validatorList {
			nodesList[i] = NewTestProcessorNodeWithCustomNodesCoordinator(
				uint32(nbShards),
				shardId,
				seedAddress,
				epochStartSubscriber,
				nodesCoordinator,
				nil,
				cp,
				i,
				nil,
				headerSig,
				createVersioningHandler(),
				nodesSetup,
			)
		}
		nodesMap[shardId] = nodesList
	}

	return nodesMap
}

// CreateNodesWithNodesCoordinatorKeygenAndSingleSigner returns a map with nodes per shard each using a real nodes coordinator
// and a given single signer for blocks and a given key gen for blocks
func CreateNodesWithNodesCoordinatorKeygenAndSingleSigner(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	seedAddress string,
	singleSigner crypto.SingleSigner,
	keyGenForBlocks crypto.KeyGenerator,
) map[uint32][]*TestProcessorNode {
	cp := CreateCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards))
	pubKeys := PubKeysMapFromKeysMap(cp.Keys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, uint32(nbShards))
	validatorsMapForNodesCoordinator, _ := sharding.NodesInfoToValidators(validatorsMap)

	cpWaiting := CreateCryptoParams(2, 2, uint32(nbShards))
	pubKeysWaiting := PubKeysMapFromKeysMap(cpWaiting.Keys)
	waitingMap := GenValidatorsFromPubKeys(pubKeysWaiting, uint32(nbShards))
	waitingMapForNodesCoordinator, _ := sharding.NodesInfoToValidators(waitingMap)

	nodesMap := make(map[uint32][]*TestProcessorNode)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	nodeShuffler := &mock.NodeShufflerMock{}

	nodesSetup := &mock.NodesSetupStub{
		InitialNodesInfoCalled: func() (m map[uint32][]sharding.GenesisNodeInfoHandler, m2 map[uint32][]sharding.GenesisNodeInfoHandler) {
			return validatorsMap, waitingMap
		},
	}

	for shardId, validatorList := range validatorsMap {
		bootStorer := CreateMemUnit()
		cache, _ := lrucache.NewCache(10000)
		argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
			ShardConsensusGroupSize: shardConsensusGroupSize,
			MetaConsensusGroupSize:  metaConsensusGroupSize,
			Marshalizer:             TestMarshalizer,
			Hasher:                  TestHasher,
			Shuffler:                nodeShuffler,
			EpochStartNotifier:      epochStartSubscriber,
			BootStorer:              bootStorer,
			ShardIDAsObserver:       shardId,
			NbShards:                uint32(nbShards),
			EligibleNodes:           validatorsMapForNodesCoordinator,
			WaitingNodes:            waitingMapForNodesCoordinator,
			SelfPublicKey:           []byte(strconv.Itoa(int(shardId))),
			ConsensusGroupCache:     cache,
			ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
		}
		nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)

		if err != nil {
			fmt.Println("Error creating node coordinator")
		}

		nodesList := make([]*TestProcessorNode, len(validatorList))
		shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(nbShards), shardId)
		for i := range validatorList {
			ownAccount := CreateTestWalletAccountWithKeygenAndSingleSigner(
				shardCoordinator,
				shardId,
				singleSigner,
				keyGenForBlocks,
			)

			args := headerCheck.ArgsHeaderSigVerifier{
				Marshalizer:       TestMarshalizer,
				Hasher:            TestHasher,
				NodesCoordinator:  nodesCoordinator,
				MultiSigVerifier:  TestMultiSig,
				SingleSigVerifier: singleSigner,
				KeyGen:            keyGenForBlocks,
			}

			headerSig, _ := headerCheck.NewHeaderSigVerifier(&args)
			nodesList[i] = NewTestProcessorNodeWithCustomNodesCoordinator(
				uint32(nbShards),
				shardId,
				seedAddress,
				epochStartSubscriber,
				nodesCoordinator,
				nil,
				cp,
				i,
				ownAccount,
				headerSig,
				createVersioningHandler(),
				nodesSetup,
			)
		}
		nodesMap[shardId] = nodesList
	}

	return nodesMap
}

// ProposeBlockWithConsensusSignature proposes
func ProposeBlockWithConsensusSignature(
	shardId uint32,
	nodesMap map[uint32][]*TestProcessorNode,
	round uint64,
	nonce uint64,
	randomness []byte,
	epoch uint32,
) (data.BodyHandler, data.HeaderHandler, [][]byte, []*TestProcessorNode) {
	nodesCoordinator := nodesMap[shardId][0].NodesCoordinator

	pubKeys, err := nodesCoordinator.GetConsensusValidatorsPublicKeys(randomness, round, shardId, epoch)
	if err != nil {
		fmt.Println("Error getting the validators public keys: ", err)
	}

	// select nodes from map based on their pub keys
	consensusNodes := selectTestNodesForPubKeys(nodesMap[shardId], pubKeys)
	// first node is block proposer
	body, header, txHashes := consensusNodes[0].ProposeBlock(round, nonce)
	header.SetPrevRandSeed(randomness)
	header = DoConsensusSigningOnBlock(header, consensusNodes, pubKeys)

	return body, header, txHashes, consensusNodes
}

func selectTestNodesForPubKeys(nodes []*TestProcessorNode, pubKeys []string) []*TestProcessorNode {
	selectedNodes := make([]*TestProcessorNode, len(pubKeys))
	cntNodes := 0

	for i, pk := range pubKeys {
		for _, node := range nodes {
			pubKeyBytes, _ := node.NodeKeys.Pk.ToByteArray()
			if bytes.Equal(pubKeyBytes, []byte(pk)) {
				selectedNodes[i] = node
				cntNodes++
			}
		}
	}

	if cntNodes != len(pubKeys) {
		fmt.Println("Error selecting nodes from public keys")
	}

	return selectedNodes
}

// DoConsensusSigningOnBlock simulates a consensus aggregated signature on the provided block
func DoConsensusSigningOnBlock(
	blockHeader data.HeaderHandler,
	consensusNodes []*TestProcessorNode,
	pubKeys []string,
) data.HeaderHandler {
	// set bitmap for all consensus nodes signing
	bitmap := make([]byte, len(consensusNodes)/8+1)
	for i := range bitmap {
		bitmap[i] = 0xFF
	}

	bitmap[len(consensusNodes)/8] >>= uint8(8 - (len(consensusNodes) % 8))
	blockHeader.SetPubKeysBitmap(bitmap)
	// clear signature, as we need to compute it below
	blockHeader.SetSignature(nil)
	blockHeader.SetPubKeysBitmap(nil)
	blockHeaderHash, _ := core.CalculateHash(TestMarshalizer, TestHasher, blockHeader)

	var msig crypto.MultiSigner
	msigProposer, _ := consensusNodes[0].MultiSigner.Create(pubKeys, 0)
	_, _ = msigProposer.CreateSignatureShare(blockHeaderHash, bitmap)

	for i := 1; i < len(consensusNodes); i++ {
		msig, _ = consensusNodes[i].MultiSigner.Create(pubKeys, uint16(i))
		sigShare, _ := msig.CreateSignatureShare(blockHeaderHash, bitmap)
		_ = msigProposer.StoreSignatureShare(uint16(i), sigShare)
	}

	sig, _ := msigProposer.AggregateSigs(bitmap)
	blockHeader.SetSignature(sig)
	blockHeader.SetPubKeysBitmap(bitmap)
	blockHeader.SetLeaderSignature([]byte("leader sign"))

	return blockHeader
}

// AllShardsProposeBlock simulates each shard selecting a consensus group and proposing/broadcasting/committing a block
func AllShardsProposeBlock(
	round uint64,
	nonce uint64,
	nodesMap map[uint32][]*TestProcessorNode,
) (
	map[uint32]data.BodyHandler,
	map[uint32]data.HeaderHandler,
	map[uint32][]*TestProcessorNode,
) {

	body := make(map[uint32]data.BodyHandler)
	header := make(map[uint32]data.HeaderHandler)
	consensusNodes := make(map[uint32][]*TestProcessorNode)
	newRandomness := make(map[uint32][]byte)

	nodesList := make([]*TestProcessorNode, 0)
	for shardID := range nodesMap {
		nodesList = append(nodesList, nodesMap[shardID]...)
	}

	// propose blocks
	for i := range nodesMap {
		currentBlockHeader := nodesMap[i][0].BlockChain.GetCurrentBlockHeader()
		if check.IfNil(currentBlockHeader) {
			currentBlockHeader = nodesMap[i][0].BlockChain.GetGenesisHeader()
		}

		// TODO: remove if start of epoch block needs to be validated by the new epoch nodes
		epoch := currentBlockHeader.GetEpoch()
		prevRandomness := currentBlockHeader.GetRandSeed()
		body[i], header[i], _, consensusNodes[i] = ProposeBlockWithConsensusSignature(
			i, nodesMap, round, nonce, prevRandomness, epoch,
		)
		nodesMap[i][0].WhiteListBody(nodesList, body[i])
		newRandomness[i] = header[i].GetRandSeed()
	}

	// propagate blocks
	for i := range nodesMap {
		consensusNodes[i][0].BroadcastBlock(body[i], header[i])
		consensusNodes[i][0].CommitBlock(body[i], header[i])
	}

	time.Sleep(2 * time.Second)

	return body, header, consensusNodes
}

// SyncAllShardsWithRoundBlock enforces all nodes in each shard synchronizing the block for the given round
func SyncAllShardsWithRoundBlock(
	t *testing.T,
	nodesMap map[uint32][]*TestProcessorNode,
	indexProposers map[uint32]int,
	round uint64,
) {
	for shard, nodeList := range nodesMap {
		SyncBlock(t, nodeList, []int{indexProposers[shard]}, round)
	}
	time.Sleep(2 * time.Second)
}
