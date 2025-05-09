package integrationTests

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/versioning"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/hashing"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	mclMultiSig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/multisig"
	"github.com/multiversx/mx-chain-crypto-go/signing/multisig"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/round"
	"github.com/multiversx/mx-chain-go/consensus/spos/sposFactory"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	epochStartDisabled "github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	"github.com/multiversx/mx-chain-go/epochStart/metachain"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/epochStart/shardchain"
	cryptoFactory "github.com/multiversx/mx-chain-go/factory/crypto"
	"github.com/multiversx/mx-chain-go/factory/peerSignatureHandler"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/keysManagement"
	"github.com/multiversx/mx-chain-go/node"
	"github.com/multiversx/mx-chain-go/node/nodeDebugFactory"
	"github.com/multiversx/mx-chain-go/ntp"
	p2pFactory "github.com/multiversx/mx-chain-go/p2p/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/factory/interceptorscontainer"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	disabledInterceptors "github.com/multiversx/mx-chain-go/process/interceptors/disabled"
	interceptorsFactory "github.com/multiversx/mx-chain-go/process/interceptors/factory"
	processMock "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/scToProtocol"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	processSync "github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/sharding"
	chainShardingMocks "github.com/multiversx/mx-chain-go/sharding/mock"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/blockInfoProviders"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	testFactory "github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	vic "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/defaults"
)

// CreateNodesWithTestFullNode will create a set of nodes with full consensus and processing components
func CreateNodesWithTestFullNode(
	numMetaNodes int,
	nodesPerShard int,
	consensusSize int,
	roundTime uint64,
	consensusType string,
	numKeysOnEachNode int,
	enableEpochsConfig config.EnableEpochs,
	withSync bool,
) map[uint32][]*TestFullNode {

	nodes := make(map[uint32][]*TestFullNode, nodesPerShard)
	cp := CreateCryptoParams(nodesPerShard, numMetaNodes, maxShards, numKeysOnEachNode)
	keysMap := PubKeysMapFromNodesKeysMap(cp.NodesKeys)
	validatorsMap := GenValidatorsFromPubKeys(keysMap, maxShards)
	eligibleMap, _ := nodesCoordinator.NodesInfoToValidators(validatorsMap)
	waitingMap := make(map[uint32][]nodesCoordinator.Validator)
	connectableNodes := make(map[uint32][]Connectable, 0)

	startTime := time.Now().Unix()
	testHasher := createHasher(consensusType)

	for shardID := range cp.NodesKeys {
		for _, keysPair := range cp.NodesKeys[shardID] {
			multiSigner, _ := multisig.NewBLSMultisig(&mclMultiSig.BlsMultiSigner{Hasher: testHasher}, cp.KeyGen)
			multiSignerMock := createCustomMultiSignerMock(multiSigner)

			args := ArgsTestFullNode{
				ArgTestProcessorNode: &ArgTestProcessorNode{
					MaxShards:            2,
					NodeShardId:          0,
					TxSignPrivKeyShardId: 0,
					WithSync:             withSync,
					EpochsConfig:         &enableEpochsConfig,
					NodeKeys:             keysPair,
				},
				ShardID:       shardID,
				ConsensusSize: consensusSize,
				RoundTime:     roundTime,
				ConsensusType: consensusType,
				EligibleMap:   eligibleMap,
				WaitingMap:    waitingMap,
				KeyGen:        cp.KeyGen,
				P2PKeyGen:     cp.P2PKeyGen,
				MultiSigner:   multiSignerMock,
				StartTime:     startTime,
			}

			tfn := NewTestFullNode(args)
			nodes[shardID] = append(nodes[shardID], tfn)
			connectableNodes[shardID] = append(connectableNodes[shardID], tfn)
		}
	}

	for shardID := range nodes {
		ConnectNodes(connectableNodes[shardID])
	}

	return nodes
}

// ArgsTestFullNode defines arguments for test full node
type ArgsTestFullNode struct {
	*ArgTestProcessorNode

	ShardID       uint32
	ConsensusSize int
	RoundTime     uint64
	ConsensusType string
	EligibleMap   map[uint32][]nodesCoordinator.Validator
	WaitingMap    map[uint32][]nodesCoordinator.Validator
	KeyGen        crypto.KeyGenerator
	P2PKeyGen     crypto.KeyGenerator
	MultiSigner   *cryptoMocks.MultisignerMock
	StartTime     int64
}

// TestFullNode defines the structure for testing node with full processing and consensus components
type TestFullNode struct {
	*TestProcessorNode

	ShardCoordinator sharding.Coordinator
	MultiSigner      *cryptoMocks.MultisignerMock
	GenesisTimeField time.Time
}

// NewTestFullNode will create a new instance of full testing node
func NewTestFullNode(args ArgsTestFullNode) *TestFullNode {
	tpn := newBaseTestProcessorNode(*args.ArgTestProcessorNode)

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, args.ShardID)

	tfn := &TestFullNode{
		TestProcessorNode: tpn,
		ShardCoordinator:  shardCoordinator,
		MultiSigner:       args.MultiSigner,
	}

	tfn.initTestNodeWithArgs(*args.ArgTestProcessorNode, args)

	return tfn
}

func (tfn *TestFullNode) initNodesCoordinator(
	consensusSize int,
	hasher hashing.Hasher,
	epochStartRegistrationHandler notifier.EpochStartNotifier,
	eligibleMap map[uint32][]nodesCoordinator.Validator,
	waitingMap map[uint32][]nodesCoordinator.Validator,
	pkBytes []byte,
	cache storage.Cacher,
) {
	argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
		ChainParametersHandler: &chainParameters.ChainParametersHandlerStub{
			ChainParametersForEpochCalled: func(_ uint32) (config.ChainParametersByEpochConfig, error) {
				return config.ChainParametersByEpochConfig{
					ShardConsensusGroupSize:     uint32(consensusSize),
					MetachainConsensusGroupSize: uint32(consensusSize),
				}, nil
			},
		},
		Marshalizer:                     TestMarshalizer,
		Hasher:                          hasher,
		Shuffler:                        &shardingMocks.NodeShufflerMock{},
		EpochStartNotifier:              epochStartRegistrationHandler,
		BootStorer:                      CreateMemUnit(),
		NbShards:                        maxShards,
		EligibleNodes:                   eligibleMap,
		WaitingNodes:                    waitingMap,
		SelfPublicKey:                   pkBytes,
		ConsensusGroupCache:             cache,
		ShuffledOutHandler:              &chainShardingMocks.ShuffledOutHandlerStub{},
		ChanStopNode:                    endProcess.GetDummyEndProcessChannel(),
		NodeTypeProvider:                &nodeTypeProviderMock.NodeTypeProviderStub{},
		IsFullArchive:                   false,
		EnableEpochsHandler:             &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ValidatorInfoCacher:             &vic.ValidatorInfoCacherStub{},
		ShardIDAsObserver:               tfn.ShardCoordinator.SelfId(),
		GenesisNodesSetupHandler:        &genesisMocks.NodesSetupStub{},
		NodesCoordinatorRegistryFactory: &shardingMocks.NodesCoordinatorRegistryFactoryMock{},
	}

	tfn.NodesCoordinator, _ = nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
}

func (tpn *TestFullNode) initTestNodeWithArgs(args ArgTestProcessorNode, fullArgs ArgsTestFullNode) {
	tpn.AppStatusHandler = args.AppStatusHandler
	if check.IfNil(args.AppStatusHandler) {
		tpn.AppStatusHandler = TestAppStatusHandler
	}

	tpn.MainMessenger = CreateMessengerWithNoDiscovery()

	tpn.StatusMetrics = args.StatusMetrics
	if check.IfNil(args.StatusMetrics) {
		args.StatusMetrics = &testscommon.StatusMetricsStub{}
	}

	tpn.initChainHandler()
	tpn.initHeaderValidator()
	tpn.initRoundHandler()

	syncer := ntp.NewSyncTime(ntp.NewNTPGoogleConfig(), nil)
	syncer.StartSyncingTime()
	tpn.GenesisTimeField = time.Unix(fullArgs.StartTime, 0)

	roundHandler, _ := round.NewRound(
		tpn.GenesisTimeField,
		syncer.CurrentTime(),
		time.Millisecond*time.Duration(fullArgs.RoundTime),
		syncer,
		0)

	tpn.NetworkShardingCollector = mock.NewNetworkShardingCollectorMock()
	if check.IfNil(tpn.EpochNotifier) {
		tpn.EpochStartNotifier = notifier.NewEpochStartSubscriptionHandler()
	}
	tpn.initStorage()
	if check.IfNil(args.TrieStore) {
		tpn.initAccountDBsWithPruningStorer()
	} else {
		tpn.initAccountDBs(args.TrieStore)
	}

	economicsConfig := args.EconomicsConfig
	if economicsConfig == nil {
		economicsConfig = createDefaultEconomicsConfig()
	}

	tpn.initEconomicsData(economicsConfig)
	tpn.initRatingsData()
	tpn.initRequestedItemsHandler()
	tpn.initResolvers()
	tpn.initRequesters()
	tpn.initValidatorStatistics()
	tpn.initGenesisBlocks(args)
	tpn.initBlockTracker(roundHandler)

	gasMap := wasmConfig.MakeGasMapForTests()
	defaults.FillGasMapInternal(gasMap, 1)
	if args.GasScheduleMap != nil {
		gasMap = args.GasScheduleMap
	}
	vmConfig := getDefaultVMConfig()
	if args.VMConfig != nil {
		vmConfig = args.VMConfig
	}
	tpn.initInnerProcessors(gasMap, vmConfig)

	if check.IfNil(args.TrieStore) {
		var apiBlockchain data.ChainHandler
		if tpn.ShardCoordinator.SelfId() == core.MetachainShardId {
			apiBlockchain, _ = blockchain.NewMetaChain(statusHandlerMock.NewAppStatusHandlerMock())
		} else {
			apiBlockchain, _ = blockchain.NewBlockChain(statusHandlerMock.NewAppStatusHandlerMock())
		}
		argsNewScQueryService := smartContract.ArgsNewSCQueryService{
			VmContainer:              tpn.VMContainer,
			EconomicsFee:             tpn.EconomicsData,
			BlockChainHook:           tpn.BlockchainHook,
			MainBlockChain:           tpn.BlockChain,
			APIBlockChain:            apiBlockchain,
			WasmVMChangeLocker:       tpn.WasmVMChangeLocker,
			Bootstrapper:             tpn.Bootstrapper,
			AllowExternalQueriesChan: common.GetClosedUnbufferedChannel(),
			HistoryRepository:        tpn.HistoryRepository,
			ShardCoordinator:         tpn.ShardCoordinator,
			StorageService:           tpn.Storage,
			Marshaller:               TestMarshaller,
			Hasher:                   TestHasher,
			Uint64ByteSliceConverter: TestUint64Converter,
		}
		tpn.SCQueryService, _ = smartContract.NewSCQueryService(argsNewScQueryService)
	} else {
		tpn.createFullSCQueryService(gasMap, vmConfig)
	}

	testHasher := createHasher(fullArgs.ConsensusType)
	epochStartRegistrationHandler := notifier.NewEpochStartSubscriptionHandler()
	pkBytes, _ := tpn.NodeKeys.MainKey.Pk.ToByteArray()
	consensusCache, _ := cache.NewLRUCache(10000)

	tpn.initNodesCoordinator(
		fullArgs.ConsensusSize,
		testHasher,
		epochStartRegistrationHandler,
		fullArgs.EligibleMap,
		fullArgs.WaitingMap,
		pkBytes,
		consensusCache,
	)

	tpn.BroadcastMessenger, _ = sposFactory.GetBroadcastMessenger(
		TestMarshalizer,
		TestHasher,
		tpn.MainMessenger,
		tpn.ShardCoordinator,
		tpn.OwnAccount.PeerSigHandler,
		tpn.DataPool.Headers(),
		tpn.MainInterceptorsContainer,
		&testscommon.AlarmSchedulerStub{},
		testscommon.NewKeysHandlerSingleSignerMock(
			tpn.NodeKeys.MainKey.Sk,
			tpn.MainMessenger.ID(),
		),
	)

	if args.WithSync {
		tpn.initBootstrapper()
	}
	tpn.setGenesisBlock()
	tpn.initNode(fullArgs, syncer, roundHandler)
	tpn.addHandlersForCounters()
	tpn.addGenesisBlocksIntoStorage()

	if args.GenesisFile != "" {
		tpn.createHeartbeatWithHardforkTrigger()
	}
}

func (tpn *TestFullNode) setGenesisBlock() {
	genesisBlock := tpn.GenesisBlocks[tpn.ShardCoordinator.SelfId()]
	_ = tpn.BlockChain.SetGenesisHeader(genesisBlock)
	hash, _ := core.CalculateHash(TestMarshalizer, TestHasher, genesisBlock)
	tpn.BlockChain.SetGenesisHeaderHash(hash)
	log.Info("set genesis",
		"shard ID", tpn.ShardCoordinator.SelfId(),
		"hash", hex.EncodeToString(hash),
	)
}

func (tpn *TestFullNode) initChainHandler() {
	if tpn.ShardCoordinator.SelfId() == core.MetachainShardId {
		tpn.BlockChain = CreateMetaChain()
	} else {
		tpn.BlockChain = CreateShardChain()
	}
}

func (tpn *TestFullNode) initNode(
	args ArgsTestFullNode,
	syncer ntp.SyncTimer,
	roundHandler consensus.RoundHandler,
) {
	var err error

	statusCoreComponents := &testFactory.StatusCoreComponentsStub{
		StatusMetricsField:    tpn.StatusMetrics,
		AppStatusHandlerField: tpn.AppStatusHandler,
	}
	if tpn.EpochNotifier == nil {
		tpn.EpochNotifier = forking.NewGenericEpochNotifier()
	}
	if tpn.EnableEpochsHandler == nil {
		tpn.EnableEpochsHandler, _ = enablers.NewEnableEpochsHandler(CreateEnableEpochsConfig(), tpn.EpochNotifier)
	}

	epochTrigger := tpn.createEpochStartTrigger(args.StartTime)
	tpn.EpochStartTrigger = epochTrigger

	strPk := ""
	if !check.IfNil(args.HardforkPk) {
		buff, err := args.HardforkPk.ToByteArray()
		log.LogIfError(err)

		strPk = hex.EncodeToString(buff)
	}
	_ = tpn.createHardforkTrigger(strPk)

	coreComponents := GetDefaultCoreComponents(tpn.EnableEpochsHandler, tpn.EpochNotifier)
	coreComponents.SyncTimerField = syncer
	coreComponents.RoundHandlerField = roundHandler

	coreComponents.InternalMarshalizerField = TestMarshalizer
	coreComponents.VmMarshalizerField = TestVmMarshalizer
	coreComponents.TxMarshalizerField = TestTxSignMarshalizer
	coreComponents.HasherField = TestHasher
	coreComponents.AddressPubKeyConverterField = TestAddressPubkeyConverter
	coreComponents.ValidatorPubKeyConverterField = TestValidatorPubkeyConverter
	coreComponents.ChainIdCalled = func() string {
		return string(tpn.ChainID)
	}

	coreComponents.GenesisTimeField = tpn.GenesisTimeField
	coreComponents.GenesisNodesSetupField = &genesisMocks.NodesSetupStub{
		GetShardConsensusGroupSizeCalled: func() uint32 {
			return uint32(args.ConsensusSize)
		},
		GetMetaConsensusGroupSizeCalled: func() uint32 {
			return uint32(args.ConsensusSize)
		},
	}
	coreComponents.MinTransactionVersionCalled = func() uint32 {
		return tpn.MinTransactionVersion
	}
	coreComponents.TxVersionCheckField = versioning.NewTxVersionChecker(tpn.MinTransactionVersion)
	hardforkPubKeyBytes, _ := coreComponents.ValidatorPubKeyConverterField.Decode(hardforkPubKey)
	coreComponents.HardforkTriggerPubKeyField = hardforkPubKeyBytes
	coreComponents.Uint64ByteSliceConverterField = TestUint64Converter
	coreComponents.EconomicsDataField = tpn.EconomicsData
	coreComponents.APIEconomicsHandler = tpn.EconomicsData
	coreComponents.EnableEpochsHandlerField = tpn.EnableEpochsHandler
	coreComponents.EpochNotifierField = tpn.EpochNotifier
	coreComponents.RoundNotifierField = tpn.RoundNotifier
	coreComponents.WasmVMChangeLockerInternal = tpn.WasmVMChangeLocker
	coreComponents.EconomicsDataField = tpn.EconomicsData

	dataComponents := GetDefaultDataComponents()
	dataComponents.BlockChain = tpn.BlockChain
	dataComponents.DataPool = tpn.DataPool
	dataComponents.Store = tpn.Storage

	bootstrapComponents := getDefaultBootstrapComponents(tpn.ShardCoordinator, tpn.EnableEpochsHandler)

	tpn.BlockBlackListHandler = cache.NewTimeCache(TimeSpanForBadHeaders)
	tpn.ForkDetector = tpn.createForkDetector(args.StartTime, roundHandler)

	argsKeysHolder := keysManagement.ArgsManagedPeersHolder{
		KeyGenerator:          args.KeyGen,
		P2PKeyGenerator:       args.P2PKeyGen,
		MaxRoundsOfInactivity: 0, // 0 for main node, non-0 for backup node
		PrefsConfig:           config.Preferences{},
		P2PKeyConverter:       p2pFactory.NewP2PKeyConverter(),
	}
	keysHolder, _ := keysManagement.NewManagedPeersHolder(argsKeysHolder)

	// adding provided handled keys
	for _, key := range args.NodeKeys.HandledKeys {
		skBytes, _ := key.Sk.ToByteArray()
		_ = keysHolder.AddManagedPeer(skBytes)
	}

	multiSigContainer := cryptoMocks.NewMultiSignerContainerMock(args.MultiSigner)
	pubKey := tpn.NodeKeys.MainKey.Sk.GeneratePublic()
	pubKeyBytes, _ := pubKey.ToByteArray()
	pubKeyString := coreComponents.ValidatorPubKeyConverterField.SilentEncode(pubKeyBytes, log)
	argsKeysHandler := keysManagement.ArgsKeysHandler{
		ManagedPeersHolder: keysHolder,
		PrivateKey:         tpn.NodeKeys.MainKey.Sk,
		Pid:                tpn.MainMessenger.ID(),
	}
	keysHandler, _ := keysManagement.NewKeysHandler(argsKeysHandler)

	signingHandlerArgs := cryptoFactory.ArgsSigningHandler{
		PubKeys:              []string{pubKeyString},
		MultiSignerContainer: multiSigContainer,
		KeyGenerator:         args.KeyGen,
		KeysHandler:          keysHandler,
		SingleSigner:         TestSingleBlsSigner,
	}
	sigHandler, _ := cryptoFactory.NewSigningHandler(signingHandlerArgs)

	cryptoComponents := GetDefaultCryptoComponents()
	cryptoComponents.PrivKey = tpn.NodeKeys.MainKey.Sk
	cryptoComponents.PubKey = tpn.NodeKeys.MainKey.Pk
	cryptoComponents.TxSig = tpn.OwnAccount.SingleSigner
	cryptoComponents.BlockSig = tpn.OwnAccount.SingleSigner
	cryptoComponents.MultiSigContainer = cryptoMocks.NewMultiSignerContainerMock(tpn.MultiSigner)
	cryptoComponents.BlKeyGen = tpn.OwnAccount.KeygenTxSign
	cryptoComponents.TxKeyGen = TestKeyGenForAccounts

	peerSigCache, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 1000})
	peerSigHandler, _ := peerSignatureHandler.NewPeerSignatureHandler(peerSigCache, TestSingleBlsSigner, args.KeyGen)
	cryptoComponents.PeerSignHandler = peerSigHandler
	cryptoComponents.SigHandler = sigHandler
	cryptoComponents.KeysHandlerField = keysHandler

	tpn.initInterceptors(coreComponents, cryptoComponents, roundHandler, tpn.EnableEpochsHandler, tpn.Storage, epochTrigger)

	if args.WithSync {
		tpn.initBlockProcessorWithSync(coreComponents, dataComponents, roundHandler)
	} else {
		tpn.initBlockProcessor(coreComponents, dataComponents, args, roundHandler)
	}

	processComponents := GetDefaultProcessComponents()
	processComponents.ForkDetect = tpn.ForkDetector
	processComponents.BlockProcess = tpn.BlockProcessor
	processComponents.ReqFinder = tpn.RequestersFinder
	processComponents.HeaderIntegrVerif = tpn.HeaderIntegrityVerifier
	processComponents.HeaderSigVerif = tpn.HeaderSigVerifier
	processComponents.BlackListHdl = tpn.BlockBlackListHandler
	processComponents.NodesCoord = tpn.NodesCoordinator
	processComponents.ShardCoord = tpn.ShardCoordinator
	processComponents.IntContainer = tpn.MainInterceptorsContainer
	processComponents.FullArchiveIntContainer = tpn.FullArchiveInterceptorsContainer
	processComponents.HistoryRepositoryInternal = tpn.HistoryRepository
	processComponents.WhiteListHandlerInternal = tpn.WhiteListHandler
	processComponents.WhiteListerVerifiedTxsInternal = tpn.WhiteListerVerifiedTxs
	processComponents.TxsSenderHandlerField = createTxsSender(tpn.ShardCoordinator, tpn.MainMessenger)
	processComponents.HardforkTriggerField = tpn.HardforkTrigger
	processComponents.ScheduledTxsExecutionHandlerInternal = &testscommon.ScheduledTxsExecutionStub{}
	processComponents.ProcessedMiniBlocksTrackerInternal = &testscommon.ProcessedMiniBlocksTrackerStub{}
	processComponents.SentSignaturesTrackerInternal = &testscommon.SentSignatureTrackerStub{}

	processComponents.RoundHandlerField = roundHandler
	processComponents.EpochNotifier = tpn.EpochStartNotifier

	stateComponents := GetDefaultStateComponents()
	stateComponents.Accounts = tpn.AccntState
	stateComponents.AccountsAPI = tpn.AccntState

	finalProvider, _ := blockInfoProviders.NewFinalBlockInfo(dataComponents.BlockChain)
	finalAccountsApi, _ := state.NewAccountsDBApi(tpn.AccntState, finalProvider)

	currentProvider, _ := blockInfoProviders.NewCurrentBlockInfo(dataComponents.BlockChain)
	currentAccountsApi, _ := state.NewAccountsDBApi(tpn.AccntState, currentProvider)

	historicalAccountsApi, _ := state.NewAccountsDBApiWithHistory(tpn.AccntState)

	argsAccountsRepo := state.ArgsAccountsRepository{
		FinalStateAccountsWrapper:      finalAccountsApi,
		CurrentStateAccountsWrapper:    currentAccountsApi,
		HistoricalStateAccountsWrapper: historicalAccountsApi,
	}
	stateComponents.AccountsRepo, _ = state.NewAccountsRepository(argsAccountsRepo)

	networkComponents := GetDefaultNetworkComponents()
	networkComponents.Messenger = tpn.MainMessenger
	networkComponents.FullArchiveNetworkMessengerField = tpn.FullArchiveMessenger
	networkComponents.PeersRatingHandlerField = tpn.PeersRatingHandler
	networkComponents.PeersRatingMonitorField = tpn.PeersRatingMonitor
	networkComponents.InputAntiFlood = &mock.NilAntifloodHandler{}
	networkComponents.PeerHonesty = &mock.PeerHonestyHandlerStub{}

	tpn.Node, err = node.NewNode(
		node.WithAddressSignatureSize(64),
		node.WithValidatorSignatureSize(48),
		node.WithBootstrapComponents(bootstrapComponents),
		node.WithCoreComponents(coreComponents),
		node.WithStatusCoreComponents(statusCoreComponents),
		node.WithDataComponents(dataComponents),
		node.WithProcessComponents(processComponents),
		node.WithCryptoComponents(cryptoComponents),
		node.WithNetworkComponents(networkComponents),
		node.WithStateComponents(stateComponents),
		node.WithPeerDenialEvaluator(&mock.PeerDenialEvaluatorStub{}),
		node.WithStatusCoreComponents(statusCoreComponents),
		node.WithRoundDuration(args.RoundTime),
		node.WithPublicKeySize(publicKeySize),
	)
	log.LogIfError(err)

	err = nodeDebugFactory.CreateInterceptedDebugHandler(
		tpn.Node,
		tpn.MainInterceptorsContainer,
		tpn.ResolversContainer,
		tpn.RequestersFinder,
		config.InterceptorResolverDebugConfig{
			Enabled:                    true,
			CacheSize:                  1000,
			EnablePrint:                true,
			IntervalAutoPrintInSeconds: 1,
			NumRequestsThreshold:       1,
			NumResolveFailureThreshold: 1,
			DebugLineExpiration:        1000,
		},
	)
	log.LogIfError(err)
}

func (tfn *TestFullNode) createForkDetector(
	startTime int64,
	roundHandler consensus.RoundHandler,
) process.ForkDetector {
	var err error
	var forkDetector process.ForkDetector

	if tfn.ShardCoordinator.SelfId() != core.MetachainShardId {
		forkDetector, err = processSync.NewShardForkDetector(
			roundHandler,
			tfn.BlockBlackListHandler,
			tfn.BlockTracker,
			tfn.GenesisTimeField.Unix(),
			tfn.EnableEpochsHandler,
			tfn.DataPool.Proofs())
	} else {
		forkDetector, err = processSync.NewMetaForkDetector(
			roundHandler,
			tfn.BlockBlackListHandler,
			tfn.BlockTracker,
			tfn.GenesisTimeField.Unix(),
			tfn.EnableEpochsHandler,
			tfn.DataPool.Proofs())
	}
	if err != nil {
		log.Error("error creating fork detector", "error", err)
		return nil
	}

	return forkDetector
}

func (tfn *TestFullNode) createEpochStartTrigger(startTime int64) TestEpochStartTrigger {
	var epochTrigger TestEpochStartTrigger
	if tfn.ShardCoordinator.SelfId() == core.MetachainShardId {
		argsNewMetaEpochStart := &metachain.ArgsNewMetaEpochStartTrigger{
			GenesisTime:        tfn.GenesisTimeField,
			EpochStartNotifier: notifier.NewEpochStartSubscriptionHandler(),
			Settings: &config.EpochStartConfig{
				MinRoundsBetweenEpochs: 1,
				RoundsPerEpoch:         1000,
			},
			Epoch:            0,
			Storage:          createTestStore(),
			Marshalizer:      TestMarshalizer,
			Hasher:           TestHasher,
			AppStatusHandler: &statusHandlerMock.AppStatusHandlerStub{},
			DataPool:         tfn.DataPool,
		}
		epochStartTrigger, err := metachain.NewEpochStartTrigger(argsNewMetaEpochStart)
		if err != nil {
			fmt.Println(err.Error())
		}
		epochTrigger = &metachain.TestTrigger{}
		epochTrigger.SetTrigger(epochStartTrigger)
	} else {
		argsPeerMiniBlocksSyncer := shardchain.ArgPeerMiniBlockSyncer{
			MiniBlocksPool:     tfn.DataPool.MiniBlocks(),
			ValidatorsInfoPool: tfn.DataPool.ValidatorsInfo(),
			RequestHandler:     &testscommon.RequestHandlerStub{},
		}
		peerMiniBlockSyncer, _ := shardchain.NewPeerMiniBlockSyncer(argsPeerMiniBlocksSyncer)

		argsShardEpochStart := &shardchain.ArgsShardEpochStartTrigger{
			Marshalizer:          TestMarshalizer,
			Hasher:               TestHasher,
			HeaderValidator:      &mock.HeaderValidatorStub{},
			Uint64Converter:      TestUint64Converter,
			DataPool:             tfn.DataPool,
			Storage:              tfn.Storage,
			RequestHandler:       &testscommon.RequestHandlerStub{},
			Epoch:                0,
			Validity:             1,
			Finality:             1,
			EpochStartNotifier:   notifier.NewEpochStartSubscriptionHandler(),
			PeerMiniBlocksSyncer: peerMiniBlockSyncer,
			RoundHandler:         tfn.RoundHandler,
			AppStatusHandler:     &statusHandlerMock.AppStatusHandlerStub{},
			EnableEpochsHandler:  tfn.EnableEpochsHandler,
		}
		epochStartTrigger, err := shardchain.NewEpochStartTrigger(argsShardEpochStart)
		if err != nil {
			fmt.Println(err.Error())
		}
		epochTrigger = &shardchain.TestTrigger{}
		epochTrigger.SetTrigger(epochStartTrigger)
	}

	return epochTrigger
}

func (tcn *TestFullNode) initInterceptors(
	coreComponents process.CoreComponentsHolder,
	cryptoComponents process.CryptoComponentsHolder,
	roundHandler consensus.RoundHandler,
	enableEpochsHandler common.EnableEpochsHandler,
	storage dataRetriever.StorageService,
	epochStartTrigger TestEpochStartTrigger,
) {
	interceptorDataVerifierArgs := interceptorsFactory.InterceptedDataVerifierFactoryArgs{
		CacheSpan:   time.Second * 10,
		CacheExpiry: time.Second * 10,
	}

	accountsAdapter := epochStartDisabled.NewAccountsAdapter()

	blockBlackListHandler := cache.NewTimeCache(TimeSpanForBadHeaders)

	genesisBlocks := make(map[uint32]data.HeaderHandler)
	blockTracker := processMock.NewBlockTrackerMock(tcn.ShardCoordinator, genesisBlocks)

	whiteLstHandler, _ := disabledInterceptors.NewDisabledWhiteListDataVerifier()

	cacherVerifiedCfg := storageunit.CacheConfig{Capacity: 5000, Type: storageunit.LRUCache, Shards: 1}
	cacheVerified, _ := storageunit.NewCache(cacherVerifiedCfg)
	whiteListerVerifiedTxs, _ := interceptors.NewWhiteListDataVerifier(cacheVerified)

	interceptorContainerFactoryArgs := interceptorscontainer.CommonInterceptorsContainerFactoryArgs{
		CoreComponents:                 coreComponents,
		CryptoComponents:               cryptoComponents,
		Accounts:                       accountsAdapter,
		ShardCoordinator:               tcn.ShardCoordinator,
		NodesCoordinator:               tcn.NodesCoordinator,
		MainMessenger:                  tcn.MainMessenger,
		FullArchiveMessenger:           tcn.FullArchiveMessenger,
		Store:                          storage,
		DataPool:                       tcn.DataPool,
		MaxTxNonceDeltaAllowed:         common.MaxTxNonceDeltaAllowed,
		TxFeeHandler:                   &economicsmocks.EconomicsHandlerMock{},
		BlockBlackList:                 blockBlackListHandler,
		HeaderSigVerifier:              &consensusMocks.HeaderSigVerifierMock{},
		HeaderIntegrityVerifier:        CreateHeaderIntegrityVerifier(),
		ValidityAttester:               blockTracker,
		EpochStartTrigger:              epochStartTrigger,
		WhiteListHandler:               whiteLstHandler,
		WhiteListerVerifiedTxs:         whiteListerVerifiedTxs,
		AntifloodHandler:               &mock.NilAntifloodHandler{},
		ArgumentsParser:                smartContract.NewArgumentParser(),
		PreferredPeersHolder:           &p2pmocks.PeersHolderStub{},
		SizeCheckDelta:                 sizeCheckDelta,
		RequestHandler:                 &testscommon.RequestHandlerStub{},
		PeerSignatureHandler:           &processMock.PeerSignatureHandlerStub{},
		SignaturesHandler:              &processMock.SignaturesHandlerStub{},
		HeartbeatExpiryTimespanInSec:   30,
		MainPeerShardMapper:            mock.NewNetworkShardingCollectorMock(),
		FullArchivePeerShardMapper:     mock.NewNetworkShardingCollectorMock(),
		HardforkTrigger:                &testscommon.HardforkTriggerStub{},
		NodeOperationMode:              common.NormalOperation,
		InterceptedDataVerifierFactory: interceptorsFactory.NewInterceptedDataVerifierFactory(interceptorDataVerifierArgs),
	}
	if tcn.ShardCoordinator.SelfId() == core.MetachainShardId {
		interceptorContainerFactory, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(interceptorContainerFactoryArgs)
		if err != nil {
			fmt.Println(err.Error())
		}

		tcn.MainInterceptorsContainer, _, err = interceptorContainerFactory.Create()
		if err != nil {
			log.Debug("interceptor container factory Create", "error", err.Error())
		}
	} else {
		argsPeerMiniBlocksSyncer := shardchain.ArgPeerMiniBlockSyncer{
			MiniBlocksPool:     tcn.DataPool.MiniBlocks(),
			ValidatorsInfoPool: tcn.DataPool.ValidatorsInfo(),
			RequestHandler:     &testscommon.RequestHandlerStub{},
		}
		peerMiniBlockSyncer, _ := shardchain.NewPeerMiniBlockSyncer(argsPeerMiniBlocksSyncer)
		argsShardEpochStart := &shardchain.ArgsShardEpochStartTrigger{
			Marshalizer:          TestMarshalizer,
			Hasher:               TestHasher,
			HeaderValidator:      &mock.HeaderValidatorStub{},
			Uint64Converter:      TestUint64Converter,
			DataPool:             tcn.DataPool,
			Storage:              storage,
			RequestHandler:       &testscommon.RequestHandlerStub{},
			Epoch:                0,
			Validity:             1,
			Finality:             1,
			EpochStartNotifier:   notifier.NewEpochStartSubscriptionHandler(),
			PeerMiniBlocksSyncer: peerMiniBlockSyncer,
			RoundHandler:         roundHandler,
			AppStatusHandler:     &statusHandlerMock.AppStatusHandlerStub{},
			EnableEpochsHandler:  enableEpochsHandler,
		}
		_, _ = shardchain.NewEpochStartTrigger(argsShardEpochStart)

		interceptorContainerFactory, err := interceptorscontainer.NewShardInterceptorsContainerFactory(interceptorContainerFactoryArgs)
		if err != nil {
			fmt.Println(err.Error())
		}

		tcn.MainInterceptorsContainer, _, err = interceptorContainerFactory.Create()
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func (tpn *TestFullNode) initBlockProcessor(
	coreComponents *mock.CoreComponentsStub,
	dataComponents *mock.DataComponentsStub,
	args ArgsTestFullNode,
	roundHandler consensus.RoundHandler,
) {
	var err error

	accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDb[state.UserAccountsState] = tpn.AccntState
	accountsDb[state.PeerAccountsState] = tpn.PeerState

	if tpn.EpochNotifier == nil {
		tpn.EpochNotifier = forking.NewGenericEpochNotifier()
	}
	if tpn.EnableEpochsHandler == nil {
		tpn.EnableEpochsHandler, _ = enablers.NewEnableEpochsHandler(CreateEnableEpochsConfig(), tpn.EpochNotifier)
	}

	bootstrapComponents := getDefaultBootstrapComponents(tpn.ShardCoordinator, tpn.EnableEpochsHandler)
	bootstrapComponents.HdrIntegrityVerifier = tpn.HeaderIntegrityVerifier

	statusComponents := GetDefaultStatusComponents()

	statusCoreComponents := &testFactory.StatusCoreComponentsStub{
		AppStatusHandlerField: &statusHandlerMock.AppStatusHandlerStub{},
	}

	argumentsBase := block.ArgBaseProcessor{
		CoreComponents:       coreComponents,
		DataComponents:       dataComponents,
		BootstrapComponents:  bootstrapComponents,
		StatusComponents:     statusComponents,
		StatusCoreComponents: statusCoreComponents,
		Config:               config.Config{},
		AccountsDB:           accountsDb,
		ForkDetector:         tpn.ForkDetector,
		NodesCoordinator:     tpn.NodesCoordinator,
		FeeHandler:           tpn.FeeAccumulator,
		RequestHandler:       tpn.RequestHandler,
		BlockChainHook:       tpn.BlockchainHook,
		HeaderValidator:      tpn.HeaderValidator,
		BootStorer: &mock.BoostrapStorerMock{
			PutCalled: func(round int64, bootData bootstrapStorage.BootstrapData) error {
				return nil
			},
		},
		BlockTracker:                 tpn.BlockTracker,
		BlockSizeThrottler:           TestBlockSizeThrottler,
		HistoryRepository:            tpn.HistoryRepository,
		GasHandler:                   tpn.GasHandler,
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
		ReceiptsRepository:           &testscommon.ReceiptsRepositoryStub{},
		OutportDataProvider:          &outport.OutportDataProviderStub{},
		BlockProcessingCutoffHandler: &testscommon.BlockProcessingCutoffStub{},
		ManagedPeersHolder:           &testscommon.ManagedPeersHolderStub{},
		SentSignaturesTracker:        &testscommon.SentSignatureTrackerStub{},
	}

	if check.IfNil(tpn.EpochStartNotifier) {
		tpn.EpochStartNotifier = notifier.NewEpochStartSubscriptionHandler()
	}

	if tpn.ShardCoordinator.SelfId() == core.MetachainShardId {
		argumentsBase.EpochStartTrigger = tpn.EpochStartTrigger
		argumentsBase.TxCoordinator = tpn.TxCoordinator

		argsStakingToPeer := scToProtocol.ArgStakingToPeer{
			PubkeyConv:          TestValidatorPubkeyConverter,
			Hasher:              TestHasher,
			Marshalizer:         TestMarshalizer,
			PeerState:           tpn.PeerState,
			BaseState:           tpn.AccntState,
			ArgParser:           tpn.ArgsParser,
			CurrTxs:             tpn.DataPool.CurrentBlockTxs(),
			RatingsData:         tpn.RatingsData,
			EnableEpochsHandler: tpn.EnableEpochsHandler,
		}
		scToProtocolInstance, _ := scToProtocol.NewStakingToPeer(argsStakingToPeer)

		argsEpochStartData := metachain.ArgsNewEpochStartData{
			Marshalizer:         TestMarshalizer,
			Hasher:              TestHasher,
			Store:               tpn.Storage,
			DataPool:            tpn.DataPool,
			BlockTracker:        tpn.BlockTracker,
			ShardCoordinator:    tpn.ShardCoordinator,
			EpochStartTrigger:   tpn.EpochStartTrigger,
			RequestHandler:      tpn.RequestHandler,
			EnableEpochsHandler: tpn.EnableEpochsHandler,
		}
		epochStartDataCreator, _ := metachain.NewEpochStartData(argsEpochStartData)

		economicsDataProvider := metachain.NewEpochEconomicsStatistics()
		argsEpochEconomics := metachain.ArgsNewEpochEconomics{
			Marshalizer:           TestMarshalizer,
			Hasher:                TestHasher,
			Store:                 tpn.Storage,
			ShardCoordinator:      tpn.ShardCoordinator,
			RewardsHandler:        tpn.EconomicsData,
			RoundTime:             roundHandler,
			GenesisTotalSupply:    tpn.EconomicsData.GenesisTotalSupply(),
			EconomicsDataNotified: economicsDataProvider,
			StakingV2EnableEpoch:  tpn.EnableEpochs.StakingV2EnableEpoch,
		}
		epochEconomics, _ := metachain.NewEndOfEpochEconomicsDataCreator(argsEpochEconomics)

		systemVM, _ := mock.NewOneSCExecutorMockVM(tpn.BlockchainHook, TestHasher)

		argsStakingDataProvider := metachain.StakingDataProviderArgs{
			EnableEpochsHandler: coreComponents.EnableEpochsHandler(),
			SystemVM:            systemVM,
			MinNodePrice:        "1000",
		}
		stakingDataProvider, errRsp := metachain.NewStakingDataProvider(argsStakingDataProvider)
		if errRsp != nil {
			log.Error("initBlockProcessor NewRewardsStakingProvider", "error", errRsp)
		}

		rewardsStorage, _ := tpn.Storage.GetStorer(dataRetriever.RewardTransactionUnit)
		miniBlockStorage, _ := tpn.Storage.GetStorer(dataRetriever.MiniBlockUnit)
		argsEpochRewards := metachain.RewardsCreatorProxyArgs{
			BaseRewardsCreatorArgs: metachain.BaseRewardsCreatorArgs{
				ShardCoordinator:      tpn.ShardCoordinator,
				PubkeyConverter:       TestAddressPubkeyConverter,
				RewardsStorage:        rewardsStorage,
				MiniBlockStorage:      miniBlockStorage,
				Hasher:                TestHasher,
				Marshalizer:           TestMarshalizer,
				DataPool:              tpn.DataPool,
				NodesConfigProvider:   tpn.NodesCoordinator,
				UserAccountsDB:        tpn.AccntState,
				EnableEpochsHandler:   tpn.EnableEpochsHandler,
				ExecutionOrderHandler: tpn.TxExecutionOrderHandler,
				RewardsHandler:        tpn.EconomicsData,
			},
			StakingDataProvider:   stakingDataProvider,
			EconomicsDataProvider: economicsDataProvider,
		}
		epochStartRewards, err := metachain.NewRewardsCreatorProxy(argsEpochRewards)
		if err != nil {
			log.Error("error creating rewards proxy", "error", err)
		}

		validatorInfoStorage, _ := tpn.Storage.GetStorer(dataRetriever.UnsignedTransactionUnit)
		argsEpochValidatorInfo := metachain.ArgsNewValidatorInfoCreator{
			ShardCoordinator:     tpn.ShardCoordinator,
			ValidatorInfoStorage: validatorInfoStorage,
			MiniBlockStorage:     miniBlockStorage,
			Hasher:               TestHasher,
			Marshalizer:          TestMarshalizer,
			DataPool:             tpn.DataPool,
			EnableEpochsHandler:  tpn.EnableEpochsHandler,
		}
		epochStartValidatorInfo, _ := metachain.NewValidatorInfoCreator(argsEpochValidatorInfo)

		maxNodesChangeConfigProvider, _ := notifier.NewNodesConfigProvider(
			tpn.EpochNotifier,
			nil,
		)
		auctionCfg := config.SoftAuctionConfig{
			TopUpStep:             "10",
			MinTopUp:              "1",
			MaxTopUp:              "32000000",
			MaxNumberOfIterations: 100000,
		}
		ald, _ := metachain.NewAuctionListDisplayer(metachain.ArgsAuctionListDisplayer{
			TableDisplayHandler:      metachain.NewTableDisplayer(),
			ValidatorPubKeyConverter: &testscommon.PubkeyConverterMock{},
			AddressPubKeyConverter:   &testscommon.PubkeyConverterMock{},
			AuctionConfig:            auctionCfg,
		})

		argsAuctionListSelector := metachain.AuctionListSelectorArgs{
			ShardCoordinator:             tpn.ShardCoordinator,
			StakingDataProvider:          stakingDataProvider,
			MaxNodesChangeConfigProvider: maxNodesChangeConfigProvider,
			AuctionListDisplayHandler:    ald,
			SoftAuctionConfig:            auctionCfg,
		}
		auctionListSelector, _ := metachain.NewAuctionListSelector(argsAuctionListSelector)

		argsEpochSystemSC := metachain.ArgsNewEpochStartSystemSCProcessing{
			SystemVM:                     systemVM,
			UserAccountsDB:               tpn.AccntState,
			PeerAccountsDB:               tpn.PeerState,
			Marshalizer:                  TestMarshalizer,
			StartRating:                  tpn.RatingsData.StartRating(),
			ValidatorInfoCreator:         tpn.ValidatorStatisticsProcessor,
			EndOfEpochCallerAddress:      vm.EndOfEpochAddress,
			StakingSCAddress:             vm.StakingSCAddress,
			ChanceComputer:               tpn.NodesCoordinator,
			EpochNotifier:                tpn.EpochNotifier,
			GenesisNodesConfig:           tpn.NodesSetup,
			StakingDataProvider:          stakingDataProvider,
			NodesConfigProvider:          tpn.NodesCoordinator,
			ShardCoordinator:             tpn.ShardCoordinator,
			ESDTOwnerAddressBytes:        vm.EndOfEpochAddress,
			EnableEpochsHandler:          tpn.EnableEpochsHandler,
			AuctionListSelector:          auctionListSelector,
			MaxNodesChangeConfigProvider: maxNodesChangeConfigProvider,
		}
		epochStartSystemSCProcessor, _ := metachain.NewSystemSCProcessor(argsEpochSystemSC)
		tpn.EpochStartSystemSCProcessor = epochStartSystemSCProcessor

		arguments := block.ArgMetaProcessor{
			ArgBaseProcessor:             argumentsBase,
			SCToProtocol:                 scToProtocolInstance,
			PendingMiniBlocksHandler:     &mock.PendingMiniBlocksHandlerStub{},
			EpochEconomics:               epochEconomics,
			EpochStartDataCreator:        epochStartDataCreator,
			EpochRewardsCreator:          epochStartRewards,
			EpochValidatorInfoCreator:    epochStartValidatorInfo,
			ValidatorStatisticsProcessor: tpn.ValidatorStatisticsProcessor,
			EpochSystemSCProcessor:       epochStartSystemSCProcessor,
		}

		tpn.BlockProcessor, err = block.NewMetaProcessor(arguments)
		if err != nil {
			log.Error("error creating meta blockprocessor", "error", err)
		}
	} else {
		argumentsBase.EpochStartTrigger = tpn.EpochStartTrigger
		argumentsBase.BlockChainHook = tpn.BlockchainHook
		argumentsBase.TxCoordinator = tpn.TxCoordinator
		argumentsBase.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{}

		arguments := block.ArgShardProcessor{
			ArgBaseProcessor: argumentsBase,
		}

		tpn.BlockProcessor, err = block.NewShardProcessor(arguments)
		if err != nil {
			log.Error("error creating shard blockprocessor", "error", err)
		}
	}

}

func (tpn *TestFullNode) initBlockProcessorWithSync(
	coreComponents *mock.CoreComponentsStub,
	dataComponents *mock.DataComponentsStub,
	roundHandler consensus.RoundHandler,
) {
	var err error

	accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDb[state.UserAccountsState] = tpn.AccntState
	accountsDb[state.PeerAccountsState] = tpn.PeerState

	if tpn.EpochNotifier == nil {
		tpn.EpochNotifier = forking.NewGenericEpochNotifier()
	}
	if tpn.EnableEpochsHandler == nil {
		tpn.EnableEpochsHandler, _ = enablers.NewEnableEpochsHandler(CreateEnableEpochsConfig(), tpn.EpochNotifier)
	}

	bootstrapComponents := getDefaultBootstrapComponents(tpn.ShardCoordinator, tpn.EnableEpochsHandler)
	bootstrapComponents.HdrIntegrityVerifier = tpn.HeaderIntegrityVerifier

	statusComponents := GetDefaultStatusComponents()

	statusCoreComponents := &factory.StatusCoreComponentsStub{
		AppStatusHandlerField: &statusHandlerMock.AppStatusHandlerStub{},
	}

	argumentsBase := block.ArgBaseProcessor{
		CoreComponents:       coreComponents,
		DataComponents:       dataComponents,
		BootstrapComponents:  bootstrapComponents,
		StatusComponents:     statusComponents,
		StatusCoreComponents: statusCoreComponents,
		Config:               config.Config{},
		AccountsDB:           accountsDb,
		ForkDetector:         nil,
		NodesCoordinator:     tpn.NodesCoordinator,
		FeeHandler:           tpn.FeeAccumulator,
		RequestHandler:       tpn.RequestHandler,
		BlockChainHook:       &testscommon.BlockChainHookStub{},
		EpochStartTrigger:    &mock.EpochStartTriggerStub{},
		HeaderValidator:      tpn.HeaderValidator,
		BootStorer: &mock.BoostrapStorerMock{
			PutCalled: func(round int64, bootData bootstrapStorage.BootstrapData) error {
				return nil
			},
		},
		BlockTracker:                 tpn.BlockTracker,
		BlockSizeThrottler:           TestBlockSizeThrottler,
		HistoryRepository:            tpn.HistoryRepository,
		GasHandler:                   tpn.GasHandler,
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
		ReceiptsRepository:           &testscommon.ReceiptsRepositoryStub{},
		OutportDataProvider:          &outport.OutportDataProviderStub{},
		BlockProcessingCutoffHandler: &testscommon.BlockProcessingCutoffStub{},
		ManagedPeersHolder:           &testscommon.ManagedPeersHolderStub{},
		SentSignaturesTracker:        &testscommon.SentSignatureTrackerStub{},
	}

	if tpn.ShardCoordinator.SelfId() == core.MetachainShardId {
		argumentsBase.ForkDetector = tpn.ForkDetector
		argumentsBase.TxCoordinator = &mock.TransactionCoordinatorMock{}
		arguments := block.ArgMetaProcessor{
			ArgBaseProcessor:          argumentsBase,
			SCToProtocol:              &mock.SCToProtocolStub{},
			PendingMiniBlocksHandler:  &mock.PendingMiniBlocksHandlerStub{},
			EpochStartDataCreator:     &mock.EpochStartDataCreatorStub{},
			EpochEconomics:            &mock.EpochEconomicsStub{},
			EpochRewardsCreator:       &testscommon.RewardsCreatorStub{},
			EpochValidatorInfoCreator: &testscommon.EpochValidatorInfoCreatorStub{},
			ValidatorStatisticsProcessor: &testscommon.ValidatorStatisticsProcessorStub{
				UpdatePeerStateCalled: func(header data.MetaHeaderHandler) ([]byte, error) {
					return []byte("validator stats root hash"), nil
				},
			},
			EpochSystemSCProcessor: &testscommon.EpochStartSystemSCStub{},
		}

		tpn.BlockProcessor, err = block.NewMetaProcessor(arguments)
	} else {
		argumentsBase.ForkDetector = tpn.ForkDetector
		argumentsBase.BlockChainHook = tpn.BlockchainHook
		argumentsBase.TxCoordinator = tpn.TxCoordinator
		argumentsBase.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{}
		arguments := block.ArgShardProcessor{
			ArgBaseProcessor: argumentsBase,
		}

		tpn.BlockProcessor, err = block.NewShardProcessor(arguments)
	}

	if err != nil {
		panic(fmt.Sprintf("Error creating blockprocessor: %s", err.Error()))
	}
}

func (tpn *TestFullNode) initBlockTracker(
	roundHandler consensus.RoundHandler,
) {
	argBaseTracker := track.ArgBaseTracker{
		Hasher:                        TestHasher,
		HeaderValidator:               tpn.HeaderValidator,
		Marshalizer:                   TestMarshalizer,
		RequestHandler:                tpn.RequestHandler,
		RoundHandler:                  roundHandler,
		ShardCoordinator:              tpn.ShardCoordinator,
		Store:                         tpn.Storage,
		StartHeaders:                  tpn.GenesisBlocks,
		PoolsHolder:                   tpn.DataPool,
		WhitelistHandler:              tpn.WhiteListHandler,
		FeeHandler:                    tpn.EconomicsData,
		EnableEpochsHandler:           tpn.EnableEpochsHandler,
		ProofsPool:                    tpn.DataPool.Proofs(),
		EpochChangeGracePeriodHandler: tpn.EpochChangeGracePeriodHandler,
	}

	var err error
	if tpn.ShardCoordinator.SelfId() != core.MetachainShardId {
		arguments := track.ArgShardTracker{
			ArgBaseTracker: argBaseTracker,
		}

		tpn.BlockTracker, err = track.NewShardBlockTrack(arguments)
		if err != nil {
			log.Error("NewShardBlockTrack", "error", err)
		}
	} else {
		arguments := track.ArgMetaTracker{
			ArgBaseTracker: argBaseTracker,
		}

		tpn.BlockTracker, err = track.NewMetaBlockTrack(arguments)
		if err != nil {
			log.Error("NewMetaBlockTrack", "error", err)
		}
	}
}
