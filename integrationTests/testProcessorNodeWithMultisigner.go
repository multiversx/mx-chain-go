package integrationTests

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	mclmultisig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/multisig"
	"github.com/multiversx/mx-chain-crypto-go/signing/multisig"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/factory/peerSignatureHandler"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/headerCheck"
	"github.com/multiversx/mx-chain-go/process/rating"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	vic "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
)

// CreateNodesWithNodesCoordinator returns a map with nodes per shard each using a real nodes coordinator
func CreateNodesWithNodesCoordinator(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
) map[uint32][]*TestProcessorNode {
	return CreateNodesWithNodesCoordinatorWithCacher(nodesPerShard, nbMetaNodes, nbShards, shardConsensusGroupSize, metaConsensusGroupSize)
}

// CreateNodesWithNodesCoordinatorWithCacher returns a map with nodes per shard each using a real nodes coordinator with cacher
func CreateNodesWithNodesCoordinatorWithCacher(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
) map[uint32][]*TestProcessorNode {
	coordinatorFactory := &IndexHashedNodesCoordinatorFactory{}
	return CreateNodesWithNodesCoordinatorFactory(nodesPerShard, nbMetaNodes, nbShards, shardConsensusGroupSize, metaConsensusGroupSize, coordinatorFactory)
}

// CreateNodesWithNodesCoordinatorAndTxKeys -
func CreateNodesWithNodesCoordinatorAndTxKeys(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
) map[uint32][]*TestProcessorNode {
	rater, _ := rating.NewBlockSigningRater(CreateRatingsData())
	coordinatorFactory := &IndexHashedNodesCoordinatorWithRaterFactory{
		PeerAccountListAndRatingHandler: rater,
	}
	cp := CreateCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards), 1)
	blsPubKeys := PubKeysMapFromNodesKeysMap(cp.NodesKeys)
	txPubKeys := PubKeysMapFromTxKeysMap(cp.TxKeys)
	validatorsMap := GenValidatorsFromPubKeysAndTxPubKeys(blsPubKeys, txPubKeys)
	validatorsMapForNodesCoordinator, _ := nodesCoordinator.NodesInfoToValidators(validatorsMap)

	waitingMap := make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
	for i := 0; i < nbShards; i++ {
		waitingMap[uint32(i)] = make([]nodesCoordinator.GenesisNodeInfoHandler, 0)
	}
	waitingMap[core.MetachainShardId] = make([]nodesCoordinator.GenesisNodeInfoHandler, 0)

	waitingMapForNodesCoordinator := make(map[uint32][]nodesCoordinator.Validator)
	for i := 0; i < nbShards; i++ {
		waitingMapForNodesCoordinator[uint32(i)] = make([]nodesCoordinator.Validator, 0)
	}
	waitingMapForNodesCoordinator[core.MetachainShardId] = make([]nodesCoordinator.Validator, 0)

	nodesSetup := &genesisMocks.NodesSetupStub{InitialNodesInfoCalled: func() (m map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, m2 map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
		return validatorsMap, waitingMap
	}}

	nodesMap := make(map[uint32][]*TestProcessorNode)
	completeNodesList := make([]Connectable, 0)
	for shardId, validatorList := range validatorsMap {
		nodesList := make([]*TestProcessorNode, len(validatorList))

		for i := range validatorList {
			dataCache, _ := cache.NewLRUCache(10000)
			tpn := CreateNodeWithBLSAndTxKeys(
				nodesPerShard,
				nbMetaNodes,
				shardConsensusGroupSize,
				metaConsensusGroupSize,
				shardId,
				nbShards,
				validatorsMapForNodesCoordinator,
				waitingMapForNodesCoordinator,
				i,
				cp,
				dataCache,
				coordinatorFactory,
				nodesSetup,
				nil,
			)

			nodesList[i] = tpn
			completeNodesList = append(completeNodesList, tpn)
		}

		nodesMap[shardId] = append(nodesMap[shardId], nodesList...)
	}

	ConnectNodes(completeNodesList)

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
	validatorsMap map[uint32][]nodesCoordinator.Validator,
	waitingMap map[uint32][]nodesCoordinator.Validator,
	keyIndex int,
	cp *CryptoParams,
	cache nodesCoordinator.Cacher,
	coordinatorFactory NodesCoordinatorFactory,
	nodesSetup sharding.GenesisNodesSetupHandler,
	ratingsData *rating.RatingsData,
) *TestProcessorNode {

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

	peerSigCache, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 1000})
	twa.PeerSigHandler, _ = peerSignatureHandler.NewPeerSignatureHandler(peerSigCache, twa.SingleSigner, keyGen)

	epochsConfig := config.EnableEpochs{
		StakingV2EnableEpoch:                 1,
		DelegationManagerEnableEpoch:         1,
		DelegationSmartContractEnableEpoch:   1,
		ScheduledMiniBlocksEnableEpoch:       UnreachableEpoch,
		MiniBlockPartialExecutionEnableEpoch: UnreachableEpoch,
		RefactorPeersMiniBlocksEnableEpoch:   UnreachableEpoch,
		AndromedaEnableEpoch:                 UnreachableEpoch,
	}

	return CreateNode(
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
		cache,
		coordinatorFactory,
		nodesSetup,
		ratingsData,
		twa,
		epochsConfig,
	)
}

// CreateNodesWithNodesCoordinatorFactory returns a map with nodes per shard each using a real nodes coordinator
func CreateNodesWithNodesCoordinatorFactory(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	nodesCoordinatorFactory NodesCoordinatorFactory,
) map[uint32][]*TestProcessorNode {
	cp := CreateCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards), 1)
	pubKeys := PubKeysMapFromNodesKeysMap(cp.NodesKeys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, uint32(nbShards))
	validatorsMapForNodesCoordinator, _ := nodesCoordinator.NodesInfoToValidators(validatorsMap)

	cpWaiting := CreateCryptoParams(1, 1, uint32(nbShards), 1)
	pubKeysWaiting := PubKeysMapFromNodesKeysMap(cpWaiting.NodesKeys)
	waitingMap := GenValidatorsFromPubKeys(pubKeysWaiting, uint32(nbShards))
	waitingMapForNodesCoordinator, _ := nodesCoordinator.NodesInfoToValidators(waitingMap)

	numNodes := nbShards*nodesPerShard + nbMetaNodes

	nodesSetup := &genesisMocks.NodesSetupStub{
		InitialNodesInfoCalled: func() (m map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, m2 map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			return validatorsMap, waitingMap
		},
		MinNumberOfNodesCalled: func() uint32 {
			return uint32(numNodes)
		},
	}

	epochsConfig := config.EnableEpochs{
		StakingV2EnableEpoch:                            UnreachableEpoch,
		ScheduledMiniBlocksEnableEpoch:                  UnreachableEpoch,
		MiniBlockPartialExecutionEnableEpoch:            UnreachableEpoch,
		RefactorPeersMiniBlocksEnableEpoch:              UnreachableEpoch,
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: UnreachableEpoch,
		StakingV4Step1EnableEpoch:                       UnreachableEpoch,
		StakingV4Step2EnableEpoch:                       UnreachableEpoch,
		StakingV4Step3EnableEpoch:                       UnreachableEpoch,
		AndromedaEnableEpoch:                            UnreachableEpoch,
	}

	nodesMap := make(map[uint32][]*TestProcessorNode)
	completeNodesList := make([]Connectable, 0)
	for shardId, validatorList := range validatorsMap {
		nodesList := make([]*TestProcessorNode, len(validatorList))
		nodesListWaiting := make([]*TestProcessorNode, len(waitingMap[shardId]))

		for i := range validatorList {
			dataCache, _ := cache.NewLRUCache(10000)
			tpn := CreateNode(
				nodesPerShard,
				nbMetaNodes,
				shardConsensusGroupSize,
				metaConsensusGroupSize,
				shardId,
				nbShards,
				validatorsMapForNodesCoordinator,
				waitingMapForNodesCoordinator,
				i,
				cp,
				dataCache,
				nodesCoordinatorFactory,
				nodesSetup,
				nil,
				nil,
				epochsConfig,
			)
			nodesList[i] = tpn
			completeNodesList = append(completeNodesList, tpn)
		}

		for i := range waitingMap[shardId] {
			dataCache, _ := cache.NewLRUCache(10000)
			tpn := CreateNode(
				nodesPerShard,
				nbMetaNodes,
				shardConsensusGroupSize,
				metaConsensusGroupSize,
				shardId,
				nbShards,
				validatorsMapForNodesCoordinator,
				waitingMapForNodesCoordinator,
				i,
				cpWaiting,
				dataCache,
				nodesCoordinatorFactory,
				nodesSetup,
				nil,
				nil,
				epochsConfig,
			)
			nodesListWaiting[i] = tpn
			completeNodesList = append(completeNodesList, tpn)
		}

		nodesMap[shardId] = append(nodesList, nodesListWaiting...)
	}

	ConnectNodes(completeNodesList)

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
	validatorsMap map[uint32][]nodesCoordinator.Validator,
	waitingMap map[uint32][]nodesCoordinator.Validator,
	keyIndex int,
	cp *CryptoParams,
	cache nodesCoordinator.Cacher,
	coordinatorFactory NodesCoordinatorFactory,
	nodesSetup sharding.GenesisNodesSetupHandler,
	ratingsData *rating.RatingsData,
	ownAccount *TestWalletAccount,
	epochsConfig config.EnableEpochs,
) *TestProcessorNode {

	epochStartSubscriber := notifier.NewEpochStartSubscriptionHandler()
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
	nodesCoordinatorInstance := coordinatorFactory.CreateNodesCoordinator(argFactory)

	txSignPrivKeyShardId := shardId
	if shardId == core.MetachainShardId {
		txSignPrivKeyShardId = 0
	}

	multiSigner, err := createMultiSigner(*cp)
	if err != nil {
		log.Error("error generating multisigner: %s\n", err)
		return nil
	}

	return NewTestProcessorNode(ArgTestProcessorNode{
		MaxShards:            uint32(nbShards),
		NodeShardId:          shardId,
		TxSignPrivKeyShardId: txSignPrivKeyShardId,
		EpochsConfig:         &epochsConfig,
		NodeKeys:             cp.NodesKeys[shardId][keyIndex],
		NodesSetup:           nodesSetup,
		NodesCoordinator:     nodesCoordinatorInstance,
		RatingsData:          ratingsData,
		MultiSigner:          multiSigner,
		EpochStartSubscriber: epochStartSubscriber,
		OwnAccount:           ownAccount,
	})
}

func createHeaderIntegrityVerifier() process.HeaderIntegrityVerifier {
	hvh := &testscommon.HeaderVersionHandlerStub{}
	headerVersioning, _ := headerCheck.NewHeaderIntegrityVerifier(
		ChainID,
		hvh,
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
	signer crypto.SingleSigner,
	keyGen crypto.KeyGenerator,
) map[uint32][]*TestProcessorNode {
	cp := CreateCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards), 1)
	pubKeys := PubKeysMapFromNodesKeysMap(cp.NodesKeys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, uint32(nbShards))
	validatorsMapForNodesCoordinator, _ := nodesCoordinator.NodesInfoToValidators(validatorsMap)
	nodesMap := make(map[uint32][]*TestProcessorNode)

	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
		EnableEpochsHandler:  &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	nodeShuffler, _ := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	epochStartSubscriber := notifier.NewEpochStartSubscriptionHandler()
	bootStorer := CreateMemUnit()

	nodesSetup := &genesisMocks.NodesSetupStub{InitialNodesInfoCalled: func() (m map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, m2 map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
		return validatorsMap, nil
	}}

	nodesCoordinatorRegistryFactory, _ := nodesCoordinator.NewNodesCoordinatorRegistryFactory(
		&marshallerMock.MarshalizerMock{},
		StakingV4Step2EnableEpoch,
	)
	completeNodesList := make([]Connectable, 0)
	for shardId, validatorList := range validatorsMap {
		consensusCache, _ := cache.NewLRUCache(10000)
		argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
			ChainParametersHandler: &chainParameters.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(_ uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{
						ShardConsensusGroupSize:     uint32(shardConsensusGroupSize),
						ShardMinNumNodes:            uint32(nodesPerShard),
						MetachainConsensusGroupSize: uint32(metaConsensusGroupSize),
						MetachainMinNumNodes:        uint32(nbMetaNodes),
						Hysteresis:                  hysteresis,
						Adaptivity:                  adaptivity,
					}, nil
				},
			},
			Marshalizer:                     TestMarshalizer,
			Hasher:                          TestHasher,
			Shuffler:                        nodeShuffler,
			BootStorer:                      bootStorer,
			EpochStartNotifier:              epochStartSubscriber,
			ShardIDAsObserver:               shardId,
			NbShards:                        uint32(nbShards),
			EligibleNodes:                   validatorsMapForNodesCoordinator,
			WaitingNodes:                    make(map[uint32][]nodesCoordinator.Validator),
			SelfPublicKey:                   []byte(strconv.Itoa(int(shardId))),
			ConsensusGroupCache:             consensusCache,
			ShuffledOutHandler:              &mock.ShuffledOutHandlerStub{},
			ChanStopNode:                    endProcess.GetDummyEndProcessChannel(),
			NodeTypeProvider:                &nodeTypeProviderMock.NodeTypeProviderStub{},
			IsFullArchive:                   false,
			EnableEpochsHandler:             &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			ValidatorInfoCacher:             &vic.ValidatorInfoCacherStub{},
			GenesisNodesSetupHandler:        &genesisMocks.NodesSetupStub{},
			NodesCoordinatorRegistryFactory: nodesCoordinatorRegistryFactory,
		}
		nodesCoordinatorInstance, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)

		if err != nil {
			fmt.Println("Error creating node coordinator: " + err.Error())
		}

		nodesList := make([]*TestProcessorNode, len(validatorList))
		args := headerCheck.ArgsHeaderSigVerifier{
			Marshalizer:             TestMarshalizer,
			Hasher:                  TestHasher,
			NodesCoordinator:        nodesCoordinatorInstance,
			MultiSigContainer:       cryptoMocks.NewMultiSignerContainerMock(TestMultiSig),
			SingleSigVerifier:       signer,
			KeyGen:                  keyGen,
			FallbackHeaderValidator: &testscommon.FallBackHeaderValidatorStub{},
			EnableEpochsHandler:     enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
			HeadersPool:             &mock.HeadersCacherStub{},
			ProofsPool:              &dataRetriever.ProofsPoolMock{},
			StorageService:          &genericMocks.ChainStorerMock{},
		}
		headerSig, _ := headerCheck.NewHeaderSigVerifier(&args)

		txSignPrivKeyShardId := shardId
		if shardId == core.MetachainShardId {
			txSignPrivKeyShardId = 0
		}

		for i := range validatorList {
			multiSigner, errCreate := createMultiSigner(*cp)
			if errCreate != nil {
				log.Error("error generating multisigner: %s\n", errCreate)
				return nil
			}

			tpn := NewTestProcessorNode(ArgTestProcessorNode{
				MaxShards:            uint32(nbShards),
				NodeShardId:          shardId,
				TxSignPrivKeyShardId: txSignPrivKeyShardId,
				EpochsConfig: &config.EnableEpochs{
					StakingV2EnableEpoch:                 UnreachableEpoch,
					ScheduledMiniBlocksEnableEpoch:       UnreachableEpoch,
					MiniBlockPartialExecutionEnableEpoch: UnreachableEpoch,
					AndromedaEnableEpoch:                 UnreachableEpoch,
				},
				NodeKeys:                cp.NodesKeys[shardId][i],
				NodesSetup:              nodesSetup,
				NodesCoordinator:        nodesCoordinatorInstance,
				MultiSigner:             multiSigner,
				EpochStartSubscriber:    epochStartSubscriber,
				HeaderSigVerifier:       headerSig,
				HeaderIntegrityVerifier: createHeaderIntegrityVerifier(),
			})

			nodesList[i] = tpn
			completeNodesList = append(completeNodesList, tpn)
		}
		nodesMap[shardId] = nodesList
	}

	ConnectNodes(completeNodesList)

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
	singleSigner crypto.SingleSigner,
	keyGenForBlocks crypto.KeyGenerator,
) map[uint32][]*TestProcessorNode {
	cp := CreateCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards), 1)
	pubKeys := PubKeysMapFromNodesKeysMap(cp.NodesKeys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, uint32(nbShards))
	validatorsMapForNodesCoordinator, _ := nodesCoordinator.NodesInfoToValidators(validatorsMap)

	cpWaiting := CreateCryptoParams(2, 2, uint32(nbShards), 1)
	pubKeysWaiting := PubKeysMapFromNodesKeysMap(cpWaiting.NodesKeys)
	waitingMap := GenValidatorsFromPubKeys(pubKeysWaiting, uint32(nbShards))
	waitingMapForNodesCoordinator, _ := nodesCoordinator.NodesInfoToValidators(waitingMap)

	nodesMap := make(map[uint32][]*TestProcessorNode)
	epochStartSubscriber := notifier.NewEpochStartSubscriptionHandler()
	nodeShuffler := &shardingMocks.NodeShufflerMock{}

	nodesSetup := &genesisMocks.NodesSetupStub{
		InitialNodesInfoCalled: func() (m map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, m2 map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			return validatorsMap, waitingMap
		},
	}

	nodesCoordinatorRegistryFactory, _ := nodesCoordinator.NewNodesCoordinatorRegistryFactory(
		&marshallerMock.MarshalizerMock{},
		StakingV4Step2EnableEpoch,
	)
	completeNodesList := make([]Connectable, 0)
	for shardId, validatorList := range validatorsMap {
		bootStorer := CreateMemUnit()
		lruCache, _ := cache.NewLRUCache(10000)
		argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
			ChainParametersHandler: &chainParameters.ChainParametersHandlerStub{
				ChainParametersForEpochCalled: func(_ uint32) (config.ChainParametersByEpochConfig, error) {
					return config.ChainParametersByEpochConfig{
						ShardConsensusGroupSize:     uint32(shardConsensusGroupSize),
						MetachainConsensusGroupSize: uint32(metaConsensusGroupSize),
					}, nil
				},
			},
			Marshalizer:                     TestMarshalizer,
			Hasher:                          TestHasher,
			Shuffler:                        nodeShuffler,
			EpochStartNotifier:              epochStartSubscriber,
			BootStorer:                      bootStorer,
			ShardIDAsObserver:               shardId,
			NbShards:                        uint32(nbShards),
			EligibleNodes:                   validatorsMapForNodesCoordinator,
			WaitingNodes:                    waitingMapForNodesCoordinator,
			SelfPublicKey:                   []byte(strconv.Itoa(int(shardId))),
			ConsensusGroupCache:             lruCache,
			ShuffledOutHandler:              &mock.ShuffledOutHandlerStub{},
			ChanStopNode:                    endProcess.GetDummyEndProcessChannel(),
			NodeTypeProvider:                &nodeTypeProviderMock.NodeTypeProviderStub{},
			IsFullArchive:                   false,
			EnableEpochsHandler:             &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			ValidatorInfoCacher:             &vic.ValidatorInfoCacherStub{},
			GenesisNodesSetupHandler:        &genesisMocks.NodesSetupStub{},
			NodesCoordinatorRegistryFactory: nodesCoordinatorRegistryFactory,
		}
		nodesCoord, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)

		if err != nil {
			fmt.Println("Error creating node coordinator")
		}

		txSignPrivKeyShardId := shardId
		if shardId == core.MetachainShardId {
			txSignPrivKeyShardId = 0
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
				Marshalizer:             TestMarshalizer,
				Hasher:                  TestHasher,
				NodesCoordinator:        nodesCoord,
				MultiSigContainer:       cryptoMocks.NewMultiSignerContainerMock(TestMultiSig),
				SingleSigVerifier:       singleSigner,
				KeyGen:                  keyGenForBlocks,
				FallbackHeaderValidator: &testscommon.FallBackHeaderValidatorStub{},
				EnableEpochsHandler:     enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
				HeadersPool:             &mock.HeadersCacherStub{},
				ProofsPool:              &dataRetriever.ProofsPoolMock{},
				StorageService:          &genericMocks.ChainStorerMock{},
			}

			headerSig, _ := headerCheck.NewHeaderSigVerifier(&args)

			multiSigner, errCreate := createMultiSigner(*cp)
			if errCreate != nil {
				log.Error("error generating multisigner: %s\n", errCreate)
				return nil
			}

			tpn := NewTestProcessorNode(ArgTestProcessorNode{
				MaxShards:            uint32(nbShards),
				NodeShardId:          shardId,
				TxSignPrivKeyShardId: txSignPrivKeyShardId,
				EpochsConfig: &config.EnableEpochs{
					StakingV2EnableEpoch:                 UnreachableEpoch,
					ScheduledMiniBlocksEnableEpoch:       UnreachableEpoch,
					MiniBlockPartialExecutionEnableEpoch: UnreachableEpoch,
					AndromedaEnableEpoch:                 UnreachableEpoch,
				},
				NodeKeys:                cp.NodesKeys[shardId][i],
				NodesSetup:              nodesSetup,
				NodesCoordinator:        nodesCoord,
				MultiSigner:             multiSigner,
				EpochStartSubscriber:    epochStartSubscriber,
				HeaderSigVerifier:       headerSig,
				HeaderIntegrityVerifier: createHeaderIntegrityVerifier(),
				OwnAccount:              ownAccount,
			})

			nodesList[i] = tpn
			completeNodesList = append(completeNodesList, tpn)
		}
		nodesMap[shardId] = nodesList
	}

	ConnectNodes(completeNodesList)

	return nodesMap
}

// ProposeBlockData is a struct that holds some context data for the proposed block
type ProposeBlockData struct {
	Body           data.BodyHandler
	Header         data.HeaderHandler
	Txs            [][]byte
	Leader         *TestProcessorNode
	ConsensusGroup []*TestProcessorNode
}

// ProposeBlockWithConsensusSignature proposes
func ProposeBlockWithConsensusSignature(
	shardId uint32,
	nodesMap map[uint32][]*TestProcessorNode,
	round uint64,
	nonce uint64,
	randomness []byte,
	epoch uint32,
) *ProposeBlockData {
	nodesCoordinatorInstance := nodesMap[shardId][0].NodesCoordinator

	leaderPubKey, pubKeys, err := nodesCoordinatorInstance.GetConsensusValidatorsPublicKeys(randomness, round, shardId, epoch)
	if err != nil {
		log.Error("nodesCoordinator.GetConsensusValidatorsPublicKeys", "error", err)
	}

	// select nodes from map based on their pub keys
	leaderNode, consensusNodes := selectTestNodesForPubKeys(nodesMap[shardId], leaderPubKey, pubKeys)
	// first node is block proposer
	body, header, txHashes := leaderNode.ProposeBlock(round, nonce)
	err = header.SetPrevRandSeed(randomness)
	if err != nil {
		log.Error("header.SetPrevRandSeed", "error", err)
	}

	header = DoConsensusSigningOnBlock(header, leaderNode, consensusNodes, pubKeys)

	return &ProposeBlockData{
		Body:           body,
		Header:         header,
		Txs:            txHashes,
		Leader:         leaderNode,
		ConsensusGroup: consensusNodes,
	}
}

func selectTestNodesForPubKeys(nodes []*TestProcessorNode, leaderPubKey string, pubKeys []string) (*TestProcessorNode, []*TestProcessorNode) {
	selectedNodes := make([]*TestProcessorNode, len(pubKeys))
	cntNodes := 0
	var leaderNode *TestProcessorNode
	for i, pk := range pubKeys {
		for j, node := range nodes {
			pubKeyBytes, _ := node.NodeKeys.MainKey.Pk.ToByteArray()
			if bytes.Equal(pubKeyBytes, []byte(pk)) {
				selectedNodes[i] = nodes[j]
				cntNodes++
			}
			if string(pubKeyBytes) == leaderPubKey {
				leaderNode = nodes[j]
			}
		}
	}

	if cntNodes != len(pubKeys) {
		fmt.Println("Error selecting nodes from public keys")
	}

	return leaderNode, selectedNodes
}

// DoConsensusSigningOnBlock simulates a ConsensusGroup aggregated signature on the provided block
func DoConsensusSigningOnBlock(
	blockHeader data.HeaderHandler,
	leaderNode *TestProcessorNode,
	consensusNodes []*TestProcessorNode,
	pubKeys []string,
) data.HeaderHandler {
	// set bitmap for all consensus nodes signing
	bitmap := make([]byte, len(consensusNodes)/8+1)
	for i := range bitmap {
		bitmap[i] = 0xFF
	}

	bitmap[len(consensusNodes)/8] >>= uint8(8 - (len(consensusNodes) % 8))
	err := blockHeader.SetPubKeysBitmap(bitmap)
	if err != nil {
		log.Error("blockHeader.SetPubKeysBitmap", "error", err)
	}

	// clear signature, as we need to compute it below
	err = blockHeader.SetSignature(nil)
	if err != nil {
		log.Error("blockHeader.SetSignature", "error", err)
	}

	err = blockHeader.SetPubKeysBitmap(nil)
	if err != nil {
		log.Error("blockHeader.SetPubKeysBitmap", "error", err)
	}

	blockHeaderHash, _ := core.CalculateHash(TestMarshalizer, TestHasher, blockHeader)

	pubKeysBytes := make([][]byte, len(consensusNodes))
	sigShares := make([][]byte, len(consensusNodes))
	msig := leaderNode.MultiSigner

	for i := 0; i < len(consensusNodes); i++ {
		pubKeysBytes[i] = []byte(pubKeys[i])
		sk, _ := consensusNodes[i].NodeKeys.MainKey.Sk.ToByteArray()
		sigShares[i], _ = msig.CreateSignatureShare(sk, blockHeaderHash)
	}

	sig, _ := msig.AggregateSigs(pubKeysBytes, sigShares)
	err = blockHeader.SetSignature(sig)
	if err != nil {
		log.Error("blockHeader.SetSignature", "error", err)
	}

	err = blockHeader.SetPubKeysBitmap(bitmap)
	if err != nil {
		log.Error("blockHeader.SetPubKeysBitmap", "error", err)
	}

	err = blockHeader.SetLeaderSignature([]byte("leader sign"))
	if err != nil {
		log.Error("blockHeader.SetLeaderSignature", "error", err)
	}

	return blockHeader
}

// AllShardsProposeBlock simulates each shard selecting a ConsensusGroup group and proposing/broadcasting/committing a block
func AllShardsProposeBlock(
	round uint64,
	nonce uint64,
	nodesMap map[uint32][]*TestProcessorNode,
) map[uint32]*ProposeBlockData {

	proposalData := make(map[uint32]*ProposeBlockData)
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
		proposalData[i] = ProposeBlockWithConsensusSignature(
			i, nodesMap, round, nonce, prevRandomness, epoch,
		)
		proposalData[i].Leader.WhiteListBody(nodesList, proposalData[i].Body)
		newRandomness[i] = proposalData[i].Header.GetRandSeed()
	}

	// propagate blocks
	for i := range nodesMap {
		leader := proposalData[i].Leader
		pk := proposalData[i].Leader.NodeKeys.MainKey.Pk
		leader.BroadcastBlock(proposalData[i].Body, proposalData[i].Header, pk)
		leader.CommitBlock(proposalData[i].Body, proposalData[i].Header)
	}

	time.Sleep(2 * StepDelay)

	return proposalData
}

// SyncAllShardsWithRoundBlock enforces all nodes in each shard synchronizing the block for the given round
func SyncAllShardsWithRoundBlock(
	t *testing.T,
	proposalData map[uint32]*ProposeBlockData,
	nodesMap map[uint32][]*TestProcessorNode,
	round uint64,
) {
	for shard, nodesList := range nodesMap {
		proposal := proposalData[shard]
		SyncBlock(t, nodesList, []*TestProcessorNode{proposal.Leader}, round)
	}
	time.Sleep(4 * StepDelay)
}

func createMultiSigner(cp CryptoParams) (crypto.MultiSigner, error) {
	blsHasher, _ := blake2b.NewBlake2bWithSize(hashing.BlsHashSize)
	llsig := &mclmultisig.BlsMultiSigner{Hasher: blsHasher}
	return multisig.NewBLSMultisig(
		llsig,
		cp.KeyGen,
	)
}
