package integrationTests

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/display"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl"
	"github.com/ElrondNetwork/elrond-go/common/enablers"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	heartbeatComp "github.com/ElrondNetwork/elrond-go/factory/heartbeat"
	"github.com/ElrondNetwork/elrond-go/factory/peerSignatureHandler"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/networksharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage/storageunit"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	vic "github.com/ElrondNetwork/elrond-go/testscommon/validatorInfoCacher"
	"github.com/ElrondNetwork/elrond-go/update/trigger"
)

// ShardTopic is the topic string generator for sharded topics
// Will generate topics in the following pattern: shard_0, shard_0_1, shard_0_META, shard_1 and so on
const ShardTopic = "shard"

// GlobalTopic is a global testing that all nodes will bind an interceptor
const GlobalTopic = "global"

// TestP2PNode represents a container type of class used in integration tests
// with all its fields exported, used mainly on P2P tests
type TestP2PNode struct {
	Messenger              p2p.Messenger
	NodesCoordinator       nodesCoordinator.NodesCoordinator
	ShardCoordinator       sharding.Coordinator
	NetworkShardingUpdater NetworkShardingUpdater
	Node                   *node.Node
	NodeKeys               TestKeyPair
	SingleSigner           crypto.SingleSigner
	KeyGen                 crypto.KeyGenerator
	Storage                dataRetriever.StorageService
	Interceptor            *CountInterceptor
}

// NewTestP2PNode returns a new NewTestP2PNode instance
func NewTestP2PNode(
	maxShards uint32,
	nodeShardId uint32,
	p2pConfig config.P2PConfig,
	coordinator nodesCoordinator.NodesCoordinator,
	keys TestKeyPair,
) *TestP2PNode {

	tP2pNode := &TestP2PNode{
		NodesCoordinator: coordinator,
		NodeKeys:         keys,
		Interceptor:      NewCountInterceptor(),
	}

	shardCoordinator, err := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)
	if err != nil {
		fmt.Printf("Error creating shard coordinator: %s\n", err.Error())
	}

	tP2pNode.ShardCoordinator = shardCoordinator

	pidPk, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 1000})
	pkShardId, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 1000})
	pidShardId, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 1000})
	arg := networksharding.ArgPeerShardMapper{
		PeerIdPkCache:         pidPk,
		FallbackPkShardCache:  pkShardId,
		FallbackPidShardCache: pidShardId,
		NodesCoordinator:      coordinator,
		PreferredPeersHolder:  &p2pmocks.PeersHolderStub{},
	}
	tP2pNode.NetworkShardingUpdater, err = networksharding.NewPeerShardMapper(arg)
	if err != nil {
		fmt.Printf("Error creating NewPeerShardMapper: %s\n", err.Error())
	}

	tP2pNode.Messenger = CreateMessengerFromConfig(p2pConfig)
	localId := tP2pNode.Messenger.ID()
	tP2pNode.NetworkShardingUpdater.UpdatePeerIDInfo(localId, []byte(""), shardCoordinator.SelfId())

	err = tP2pNode.Messenger.SetPeerShardResolver(tP2pNode.NetworkShardingUpdater)
	if err != nil {
		fmt.Printf("Error setting messenger.SetPeerShardResolver: %s\n", err.Error())
	}

	tP2pNode.initStorage()
	tP2pNode.initCrypto()
	tP2pNode.initNode()

	return tP2pNode
}

func (tP2pNode *TestP2PNode) initStorage() {
	tP2pNode.Storage = CreateStore(tP2pNode.ShardCoordinator.NumberOfShards())
}

func (tP2pNode *TestP2PNode) initCrypto() {
	tP2pNode.SingleSigner = TestSingleBlsSigner
	suite := mcl.NewSuiteBLS12()
	tP2pNode.KeyGen = signing.NewKeyGenerator(suite)
}

func (tP2pNode *TestP2PNode) initNode() {
	var err error

	pubkeys := tP2pNode.getPubkeys()

	epochStartNotifier := notifier.NewEpochStartSubscriptionHandler()
	pkBytes, _ := tP2pNode.NodeKeys.Pk.ToByteArray()
	argHardforkTrigger := trigger.ArgHardforkTrigger{
		TriggerPubKeyBytes:        []byte("invalid trigger public key"),
		Enabled:                   false,
		EnabledAuthenticated:      false,
		ArgumentParser:            smartContract.NewArgumentParser(),
		EpochProvider:             &mock.EpochStartTriggerStub{},
		ExportFactoryHandler:      &mock.ExportFactoryHandlerStub{},
		CloseAfterExportInMinutes: 5,
		ChanStopNodeProcess:       make(chan endProcess.ArgEndProcess),
		EpochConfirmedNotifier:    epochStartNotifier,
		SelfPubKeyBytes:           pkBytes,
		ImportStartHandler:        &mock.ImportStartHandlerStub{},
		RoundHandler:              &mock.RoundHandlerMock{},
	}
	argHardforkTrigger.SelfPubKeyBytes, _ = tP2pNode.NodeKeys.Pk.ToByteArray()
	hardforkTrigger, err := trigger.NewTrigger(argHardforkTrigger)
	log.LogIfError(err)

	cacher := testscommon.NewCacherMock()
	psh, err := peerSignatureHandler.NewPeerSignatureHandler(
		cacher,
		tP2pNode.SingleSigner,
		tP2pNode.KeyGen,
	)
	log.LogIfError(err)

	coreComponents := GetDefaultCoreComponents()
	coreComponents.InternalMarshalizerField = TestMarshalizer
	coreComponents.HasherField = TestHasher
	coreComponents.ValidatorPubKeyConverterField = TestValidatorPubkeyConverter
	cfg := config.EnableEpochs{
		HeartbeatDisableEpoch: UnreachableEpoch,
	}
	coreComponents.EnableEpochsHandlerField, _ = enablers.NewEnableEpochsHandler(cfg, coreComponents.EpochNotifierField)

	cryptoComponents := GetDefaultCryptoComponents()
	cryptoComponents.BlKeyGen = tP2pNode.KeyGen
	cryptoComponents.PrivKey = tP2pNode.NodeKeys.Sk
	cryptoComponents.PubKey = tP2pNode.NodeKeys.Pk
	cryptoComponents.BlockSig = tP2pNode.SingleSigner
	cryptoComponents.PeerSignHandler = psh

	processComponents := GetDefaultProcessComponents()
	processComponents.ShardCoord = tP2pNode.ShardCoordinator
	processComponents.NodesCoord = tP2pNode.NodesCoordinator
	processComponents.ValidatorProvider = &mock.ValidatorsProviderStub{}
	processComponents.ValidatorStatistics = &mock.ValidatorStatisticsProcessorStub{
		GetValidatorInfoForRootHashCalled: func(_ []byte) (map[uint32][]*state.ValidatorInfo, error) {
			return map[uint32][]*state.ValidatorInfo{
				0: {{PublicKey: []byte("pk0")}},
			}, nil
		},
	}
	processComponents.EpochNotifier = epochStartNotifier
	processComponents.EpochTrigger = &mock.EpochStartTriggerStub{}
	processComponents.PeerMapper = tP2pNode.NetworkShardingUpdater
	processComponents.HardforkTriggerField = hardforkTrigger

	networkComponents := GetDefaultNetworkComponents()
	networkComponents.Messenger = tP2pNode.Messenger
	networkComponents.PeerHonesty = &mock.PeerHonestyHandlerStub{}
	networkComponents.InputAntiFlood = &mock.NilAntifloodHandler{}

	dataComponents := GetDefaultDataComponents()
	dataComponents.Store = tP2pNode.Storage
	dataComponents.BlockChain = &testscommon.ChainHandlerStub{}

	redundancyHandler := &mock.RedundancyHandlerStub{}

	tP2pNode.Node, err = node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithCryptoComponents(cryptoComponents),
		node.WithProcessComponents(processComponents),
		node.WithNetworkComponents(networkComponents),
		node.WithDataComponents(dataComponents),
		node.WithInitialNodesPubKeys(pubkeys),
		node.WithPeerDenialEvaluator(&mock.PeerDenialEvaluatorStub{}),
	)
	log.LogIfError(err)

	hbConfig := config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 4,
		MaxTimeToWaitBetweenBroadcastsInSec: 6,
		DurationToConsiderUnresponsiveInSec: 60,
		HeartbeatRefreshIntervalInSec:       5,
		HideInactiveValidatorIntervalInSec:  600,
	}

	hbCompArgs := heartbeatComp.HeartbeatComponentsFactoryArgs{
		Config: config.Config{
			Heartbeat: hbConfig,
		},
		Prefs:             config.Preferences{},
		AppVersion:        "test",
		GenesisTime:       time.Time{},
		RedundancyHandler: redundancyHandler,
		CoreComponents:    coreComponents,
		DataComponents:    dataComponents,
		NetworkComponents: networkComponents,
		CryptoComponents:  cryptoComponents,
		ProcessComponents: processComponents,
	}
	heartbeatComponentsFactory, _ := heartbeatComp.NewHeartbeatComponentsFactory(hbCompArgs)
	managedHBComponents, err := heartbeatComp.NewManagedHeartbeatComponents(heartbeatComponentsFactory)
	log.LogIfError(err)

	err = managedHBComponents.Create()
	log.LogIfError(err)
}

func (tP2pNode *TestP2PNode) getPubkeys() map[uint32][]string {
	eligible, err := tP2pNode.NodesCoordinator.GetAllEligibleValidatorsPublicKeys(0)
	log.LogIfError(err)

	validators := make(map[uint32][]string)

	for shardId, shardValidators := range eligible {
		for _, v := range shardValidators {
			validators[shardId] = append(validators[shardId], string(v))
		}
	}

	return validators
}

// CreateTestInterceptors creates test interceptors that count the number of received messages
func (tP2pNode *TestP2PNode) CreateTestInterceptors() {
	tP2pNode.RegisterTopicValidator(GlobalTopic, tP2pNode.Interceptor)

	metaIdentifier := ShardTopic + tP2pNode.ShardCoordinator.CommunicationIdentifier(core.MetachainShardId)
	tP2pNode.RegisterTopicValidator(metaIdentifier, tP2pNode.Interceptor)

	for i := uint32(0); i < tP2pNode.ShardCoordinator.NumberOfShards(); i++ {
		identifier := ShardTopic + tP2pNode.ShardCoordinator.CommunicationIdentifier(i)
		tP2pNode.RegisterTopicValidator(identifier, tP2pNode.Interceptor)
	}
}

// CountGlobalMessages returns the messages count on the global topic
func (tP2pNode *TestP2PNode) CountGlobalMessages() int {
	return tP2pNode.Interceptor.MessageCount(GlobalTopic)
}

// CountIntraShardMessages returns the messages count on the intra-shard topic
func (tP2pNode *TestP2PNode) CountIntraShardMessages() int {
	identifier := ShardTopic + tP2pNode.ShardCoordinator.CommunicationIdentifier(tP2pNode.ShardCoordinator.SelfId())
	return tP2pNode.Interceptor.MessageCount(identifier)
}

// CountCrossShardMessages returns the messages count on the cross-shard topics
func (tP2pNode *TestP2PNode) CountCrossShardMessages() int {
	messages := 0

	if tP2pNode.ShardCoordinator.SelfId() != core.MetachainShardId {
		metaIdentifier := ShardTopic + tP2pNode.ShardCoordinator.CommunicationIdentifier(core.MetachainShardId)
		messages += tP2pNode.Interceptor.MessageCount(metaIdentifier)
	}

	for i := uint32(0); i < tP2pNode.ShardCoordinator.NumberOfShards(); i++ {
		if i == tP2pNode.ShardCoordinator.SelfId() {
			continue
		}

		metaIdentifier := ShardTopic + tP2pNode.ShardCoordinator.CommunicationIdentifier(i)
		messages += tP2pNode.Interceptor.MessageCount(metaIdentifier)
	}

	return messages
}

// RegisterTopicValidator registers a message processor instance on the provided topic
func (tP2pNode *TestP2PNode) RegisterTopicValidator(topic string, processor p2p.MessageProcessor) {
	err := tP2pNode.Messenger.CreateTopic(topic, true)
	if err != nil {
		fmt.Printf("error while creating topic %s: %s\n", topic, err.Error())
		return
	}

	err = tP2pNode.Messenger.RegisterMessageProcessor(topic, "test", processor)
	if err != nil {
		fmt.Printf("error while registering topic validator %s: %s\n", topic, err.Error())
		return
	}
}

// CreateNodesWithTestP2PNodes returns a map with nodes per shard each using a real nodes coordinator
// and TestP2PNodes
func CreateNodesWithTestP2PNodes(
	nodesPerShard int,
	numMetaNodes int,
	numShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	numObserversOnShard int,
	p2pConfig config.P2PConfig,
) map[uint32][]*TestP2PNode {

	cp := CreateCryptoParams(nodesPerShard, numMetaNodes, uint32(numShards))
	pubKeys := PubKeysMapFromKeysMap(cp.Keys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, uint32(numShards))
	validatorsForNodesCoordinator, _ := nodesCoordinator.NodesInfoToValidators(validatorsMap)
	nodesMap := make(map[uint32][]*TestP2PNode)
	cacherCfg := storageunit.CacheConfig{Capacity: 10000, Type: storageunit.LRUCache, Shards: 1}
	cache, _ := storageunit.NewCache(cacherCfg)
	for shardId, validatorList := range validatorsMap {
		argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
			ShardConsensusGroupSize: shardConsensusGroupSize,
			MetaConsensusGroupSize:  metaConsensusGroupSize,
			Marshalizer:             TestMarshalizer,
			Hasher:                  TestHasher,
			ShardIDAsObserver:       shardId,
			NbShards:                uint32(numShards),
			EligibleNodes:           validatorsForNodesCoordinator,
			SelfPublicKey:           []byte(strconv.Itoa(int(shardId))),
			ConsensusGroupCache:     cache,
			Shuffler:                &shardingMocks.NodeShufflerMock{},
			BootStorer:              CreateMemUnit(),
			WaitingNodes:            make(map[uint32][]nodesCoordinator.Validator),
			Epoch:                   0,
			EpochStartNotifier:      notifier.NewEpochStartSubscriptionHandler(),
			ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
			ChanStopNode:            endProcess.GetDummyEndProcessChannel(),
			NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
			IsFullArchive:           false,
			EnableEpochsHandler:     &testscommon.EnableEpochsHandlerStub{},
			ValidatorInfoCacher:     &vic.ValidatorInfoCacherStub{},
		}
		nodesCoord, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
		log.LogIfError(err)

		nodesList := make([]*TestP2PNode, len(validatorList))
		for i := range validatorList {
			kp := cp.Keys[shardId][i]
			nodesList[i] = NewTestP2PNode(
				uint32(numShards),
				shardId,
				p2pConfig,
				nodesCoord,
				*kp,
			)
		}
		nodesMap[shardId] = nodesList
	}

	for counter := uint32(0); counter < uint32(numShards+1); counter++ {
		for j := 0; j < numObserversOnShard; j++ {
			shardId := counter
			if shardId == uint32(numShards) {
				shardId = core.MetachainShardId
			}

			argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
				ShardConsensusGroupSize: shardConsensusGroupSize,
				MetaConsensusGroupSize:  metaConsensusGroupSize,
				Marshalizer:             TestMarshalizer,
				Hasher:                  TestHasher,
				ShardIDAsObserver:       shardId,
				NbShards:                uint32(numShards),
				EligibleNodes:           validatorsForNodesCoordinator,
				SelfPublicKey:           []byte(strconv.Itoa(int(shardId))),
				ConsensusGroupCache:     cache,
				Shuffler:                &shardingMocks.NodeShufflerMock{},
				BootStorer:              CreateMemUnit(),
				WaitingNodes:            make(map[uint32][]nodesCoordinator.Validator),
				Epoch:                   0,
				EpochStartNotifier:      notifier.NewEpochStartSubscriptionHandler(),
				ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
				ChanStopNode:            endProcess.GetDummyEndProcessChannel(),
				NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
				IsFullArchive:           false,
				EnableEpochsHandler:     &testscommon.EnableEpochsHandlerStub{},
				ValidatorInfoCacher:     &vic.ValidatorInfoCacherStub{},
			}
			nodesCoord, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
			log.LogIfError(err)

			n := NewTestP2PNode(
				uint32(numShards),
				shardId,
				p2pConfig,
				nodesCoord,
				createCryptoPair(),
			)

			nodesMap[shardId] = append(nodesMap[shardId], n)
		}
	}

	return nodesMap
}

func createCryptoPair() TestKeyPair {
	suite := mcl.NewSuiteBLS12()
	keyGen := signing.NewKeyGenerator(suite)

	kp := TestKeyPair{}
	kp.Sk, kp.Pk = keyGen.GeneratePair()

	return kp
}

// MakeDisplayTableForP2PNodes will output a string containing counters for received messages for all provided test nodes
func MakeDisplayTableForP2PNodes(nodes map[uint32][]*TestP2PNode) string {
	header := []string{"pk", "pid", "shard ID", "messages global", "messages intra", "messages cross", "conns Total/IntraVal/CrossVal/IntraObs/CrossObs/FullObs/Unk/Sed"}
	dataLines := make([]*display.LineData, 0)

	for shardId, nodesList := range nodes {
		for _, n := range nodesList {
			buffPk, _ := n.NodeKeys.Pk.ToByteArray()

			peerInfo := n.Messenger.GetConnectedPeersInfo()

			pid := n.Messenger.ID().Pretty()
			lineData := display.NewLineData(
				false,
				[]string{
					core.GetTrimmedPk(hex.EncodeToString(buffPk)),
					pid[len(pid)-6:],
					fmt.Sprintf("%d", shardId),
					fmt.Sprintf("%d", n.CountGlobalMessages()),
					fmt.Sprintf("%d", n.CountIntraShardMessages()),
					fmt.Sprintf("%d", n.CountCrossShardMessages()),
					fmt.Sprintf("%d/%d/%d/%d/%d/%d/%d/%d",
						len(n.Messenger.ConnectedPeers()),
						peerInfo.NumIntraShardValidators,
						peerInfo.NumCrossShardValidators,
						peerInfo.NumIntraShardObservers,
						peerInfo.NumCrossShardObservers,
						peerInfo.NumFullHistoryObservers,
						len(peerInfo.UnknownPeers),
						len(peerInfo.Seeders),
					),
				},
			)

			dataLines = append(dataLines, lineData)
		}
	}
	table, _ := display.CreateTableString(header, dataLines)

	return table
}
