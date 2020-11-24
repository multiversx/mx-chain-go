package integrationTests

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/peerSignatureHandler"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	mclsig "github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/networksharding"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/testscommon"
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
	NodesCoordinator       sharding.NodesCoordinator
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
	coordinator sharding.NodesCoordinator,
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

	pidPk, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 1000})
	pkShardId, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 1000})
	pidShardId, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 1000})
	startInEpoch := uint32(0)
	tP2pNode.NetworkShardingUpdater, err = networksharding.NewPeerShardMapper(
		pidPk,
		pkShardId,
		pidShardId,
		coordinator,
		startInEpoch,
	)
	if err != nil {
		fmt.Printf("Error creating NewPeerShardMapper: %s\n", err.Error())
	}

	tP2pNode.Messenger = CreateMessengerFromConfig(p2pConfig)
	localId := tP2pNode.Messenger.ID()
	tP2pNode.NetworkShardingUpdater.UpdatePeerIdShardId(localId, shardCoordinator.SelfId())

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
	tP2pNode.SingleSigner = &mclsig.BlsSingleSigner{}
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

	tP2pNode.Node, err = node.NewNode(
		node.WithMessenger(tP2pNode.Messenger),
		node.WithInternalMarshalizer(TestMarshalizer, 100),
		node.WithHasher(TestHasher),
		node.WithKeyGen(tP2pNode.KeyGen),
		node.WithShardCoordinator(tP2pNode.ShardCoordinator),
		node.WithNodesCoordinator(tP2pNode.NodesCoordinator),
		node.WithSingleSigner(tP2pNode.SingleSigner),
		node.WithPrivKey(tP2pNode.NodeKeys.Sk),
		node.WithPubKey(tP2pNode.NodeKeys.Pk),
		node.WithNetworkShardingCollector(tP2pNode.NetworkShardingUpdater),
		node.WithDataStore(tP2pNode.Storage),
		node.WithInitialNodesPubKeys(pubkeys),
		node.WithInputAntifloodHandler(&mock.NilAntifloodHandler{}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithEpochStartEventNotifier(epochStartNotifier),
		node.WithValidatorStatistics(&mock.ValidatorStatisticsProcessorStub{
			GetValidatorInfoForRootHashCalled: func(_ []byte) (map[uint32][]*state.ValidatorInfo, error) {
				return map[uint32][]*state.ValidatorInfo{
					0: {{PublicKey: []byte("pk0")}},
				}, nil
			},
		}),
		node.WithHardforkTrigger(hardforkTrigger),
		node.WithPeerDenialEvaluator(&mock.PeerDenialEvaluatorStub{}),
		node.WithValidatorPubkeyConverter(TestValidatorPubkeyConverter),
		node.WithValidatorsProvider(&mock.ValidatorsProviderStub{}),
		node.WithPeerHonestyHandler(&mock.PeerHonestyHandlerStub{}),
		node.WithFallbackHeaderValidator(&testscommon.FallBackHeaderValidatorStub{}),
		node.WithPeerSignatureHandler(psh),
		node.WithBlockChain(&mock.BlockChainMock{}),
	)
	log.LogIfError(err)

	hbConfig := config.HeartbeatConfig{
		MinTimeToWaitBetweenBroadcastsInSec: 4,
		MaxTimeToWaitBetweenBroadcastsInSec: 6,
		DurationToConsiderUnresponsiveInSec: 60,
		HeartbeatRefreshIntervalInSec:       5,
		HideInactiveValidatorIntervalInSec:  600,
	}
	err = tP2pNode.Node.StartHeartbeat(hbConfig, "test", config.PreferencesConfig{})
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

	err = tP2pNode.Messenger.RegisterMessageProcessor(topic, processor)
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
	validatorsForNodesCoordinator, _ := sharding.NodesInfoToValidators(validatorsMap)
	nodesMap := make(map[uint32][]*TestP2PNode)
	cacherCfg := storageUnit.CacheConfig{Capacity: 10000, Type: storageUnit.LRUCache, Shards: 1}
	cache, _ := storageUnit.NewCache(cacherCfg)
	for shardId, validatorList := range validatorsMap {
		argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
			ShardConsensusGroupSize: shardConsensusGroupSize,
			MetaConsensusGroupSize:  metaConsensusGroupSize,
			Marshalizer:             TestMarshalizer,
			Hasher:                  TestHasher,
			ShardIDAsObserver:       shardId,
			NbShards:                uint32(numShards),
			EligibleNodes:           validatorsForNodesCoordinator,
			SelfPublicKey:           []byte(strconv.Itoa(int(shardId))),
			ConsensusGroupCache:     cache,
			Shuffler:                &mock.NodeShufflerMock{},
			BootStorer:              CreateMemUnit(),
			WaitingNodes:            make(map[uint32][]sharding.Validator),
			Epoch:                   0,
			EpochStartNotifier:      notifier.NewEpochStartSubscriptionHandler(),
			ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
		}
		nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
		log.LogIfError(err)

		nodesList := make([]*TestP2PNode, len(validatorList))
		for i := range validatorList {
			kp := cp.Keys[shardId][i]
			nodesList[i] = NewTestP2PNode(
				uint32(numShards),
				shardId,
				p2pConfig,
				nodesCoordinator,
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

			argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
				ShardConsensusGroupSize: shardConsensusGroupSize,
				MetaConsensusGroupSize:  metaConsensusGroupSize,
				Marshalizer:             TestMarshalizer,
				Hasher:                  TestHasher,
				ShardIDAsObserver:       shardId,
				NbShards:                uint32(numShards),
				EligibleNodes:           validatorsForNodesCoordinator,
				SelfPublicKey:           []byte(strconv.Itoa(int(shardId))),
				ConsensusGroupCache:     cache,
				Shuffler:                &mock.NodeShufflerMock{},
				BootStorer:              CreateMemUnit(),
				WaitingNodes:            make(map[uint32][]sharding.Validator),
				Epoch:                   0,
				EpochStartNotifier:      notifier.NewEpochStartSubscriptionHandler(),
				ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
			}
			nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
			log.LogIfError(err)

			n := NewTestP2PNode(
				uint32(numShards),
				shardId,
				p2pConfig,
				nodesCoordinator,
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
	header := []string{"pk", "pid", "shard ID", "messages global", "messages intra", "messages cross", "conns Total/IntraVal/CrossVal/IntraObs/CrossObs/Unk"}
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
					fmt.Sprintf("%d/%d/%d/%d/%d/%d",
						len(n.Messenger.ConnectedPeers()),
						peerInfo.NumIntraShardValidators,
						peerInfo.NumCrossShardValidators,
						peerInfo.NumIntraShardObservers,
						peerInfo.NumCrossShardObservers,
						len(peerInfo.UnknownPeers),
					),
				},
			)

			dataLines = append(dataLines, lineData)
		}
	}
	table, _ := display.CreateTableString(header, dataLines)

	return table
}
