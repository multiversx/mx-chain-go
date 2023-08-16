package processOnlyNode

import (
	"bytes"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/core/versioning"
	"github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	"github.com/multiversx/mx-chain-core-go/marshal"
	nodeFactory "github.com/multiversx/mx-chain-go/cmd/node/factory"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	hdrFactory "github.com/multiversx/mx-chain-go/factory/block"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/integrationTests/processOnlyNode/customComponents"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	processSync "github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	trieFactory "github.com/multiversx/mx-chain-go/trie/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const addressPubKeyLength = 32
const validatorPubKeyLength = 96
const minTransactionVersion = 1

var log = logger.GetOrCreate("integrationTests/processOnlyNode")

// ArgsProcessOnlyNode specifies the arguments for constructing a new ProcessOnlyNode instance
type ArgsProcessOnlyNode struct {
	CurrentShardID   uint32
	NumShards        uint32
	ChainID          []byte
	GenesisTimeStamp time.Time
	RoundDuration    time.Duration

	EconomicsConfig    config.EconomicsConfig
	EnableEpochsConfig config.EnableEpochs
	EnableRoundsConfig config.RoundConfig
	CurrentGasMap      map[string]map[string]uint64
}

// ProcessOnlyNode is able to produce & execute blocks of transactions
type ProcessOnlyNode struct {
	StatusCoreComponents *factory.StatusCoreComponentsStub
	CoreComponents       *mock.CoreComponentsStub
	DataComponents       *mock.DataComponentsStub
	StatusComponents     *mock.StatusComponentsStub
	BootstrapComponents  *mainFactoryMocks.BootstrapComponentsStub

	GenesisTimeStamp       time.Time
	ShardCoordinator       sharding.Coordinator
	RequestHandler         process.RequestHandler        // TODO: have a custom, dedicated request handler here
	GenesisBlocks          map[uint32]data.HeaderHandler // TODO: have genesis blocks populated here
	TrieContainer          common.TriesHolder
	TrieStorageManagers    map[string]common.StorageManager
	Accounts               map[state.AccountsDbIdentifier]state.AccountsAdapter
	HeaderValidator        process.HeaderConstructionValidator
	WhiteListHandler       process.WhiteListHandler
	WhiteListerVerifiedTxs process.WhiteListHandler
	BlockTracker           process.BlockTracker
	BlockProcessor         process.BlockProcessor
}

// NewProcessOnlyNode creates a new instance of type ProcessOnlyNode
func NewProcessOnlyNode(args ArgsProcessOnlyNode) (*ProcessOnlyNode, error) {
	node := &ProcessOnlyNode{
		GenesisTimeStamp: args.GenesisTimeStamp,
	}

	err := node.initCommonComponents(args)
	if err != nil {
		return nil, err
	}

	err = node.initBlockProcessor()
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (node *ProcessOnlyNode) initCommonComponents(args ArgsProcessOnlyNode) error {
	var err error
	node.ShardCoordinator, err = sharding.NewMultiShardCoordinator(args.NumShards, args.CurrentShardID)
	if err != nil {
		return err
	}

	node.initStatusCoreComponents()

	err = node.initCoreComponents(args)
	if err != nil {
		return err
	}

	err = node.initDataComponents()
	if err != nil {
		return err
	}

	node.initStatusComponents()

	err = node.initBootstrapComponents()
	if err != nil {
		return err
	}

	node.initAccountDBs()

	err = node.initWhiteListHandlers()
	if err != nil {
		return err
	}

	return nil
}

func (node *ProcessOnlyNode) initStatusCoreComponents() {
	node.StatusCoreComponents = &factory.StatusCoreComponentsStub{
		AppStatusHandlerField: &statusHandlerMock.AppStatusHandlerStub{},
	}
}

func (node *ProcessOnlyNode) initCoreComponents(args ArgsProcessOnlyNode) error {
	epochNotifier := forking.NewGenericEpochNotifier()
	enableRoundsHandler, err := enablers.NewEnableRoundsHandler(args.EnableRoundsConfig)
	if err != nil {
		return err
	}

	enableEpochsHandler, err := enablers.NewEnableEpochsHandler(args.EnableEpochsConfig, epochNotifier)
	if err != nil {
		return err
	}

	addressPubKey, err := pubkeyConverter.NewBech32PubkeyConverter(addressPubKeyLength, log)
	if err != nil {
		return err
	}

	validatorPubKey, err := pubkeyConverter.NewHexPubkeyConverter(validatorPubKeyLength)
	if err != nil {
		return err
	}

	builtInCostHandler, err := economics.NewBuiltInFunctionsCost(&economics.ArgsBuiltInFunctionCost{
		ArgsParser:  smartContract.NewArgumentParser(),
		GasSchedule: mock.NewGasScheduleNotifierMock(args.CurrentGasMap),
	})
	if err != nil {
		return err
	}

	argsNewEconomicsData := economics.ArgsNewEconomicsData{
		Economics:                   &args.EconomicsConfig,
		EpochNotifier:               epochNotifier,
		EnableEpochsHandler:         enableEpochsHandler,
		BuiltInFunctionsCostHandler: builtInCostHandler,
	}
	economicsData, err := economics.NewEconomicsData(argsNewEconomicsData)
	if err != nil {
		return err
	}

	apiEconomicsData, err := economics.NewAPIEconomicsData(economicsData)
	if err != nil {
		return err
	}

	node.CoreComponents = &mock.CoreComponentsStub{
		InternalMarshalizerField:      &marshal.GogoProtoMarshalizer{},
		TxMarshalizerField:            &marshal.JsonMarshalizer{},
		VmMarshalizerField:            &marshal.JsonMarshalizer{},
		HasherField:                   blake2b.NewBlake2b(),
		TxSignHasherField:             keccak.NewKeccak(),
		Uint64ByteSliceConverterField: uint64ByteSlice.NewBigEndianConverter(),
		AddressPubKeyConverterField:   addressPubKey,
		ValidatorPubKeyConverterField: validatorPubKey,
		PathHandlerField:              &testscommon.PathManagerStub{},
		ChainIdCalled: func() string {
			return string(args.ChainID)
		},
		MinTransactionVersionCalled: func() uint32 {
			return minTransactionVersion
		},
		StatusHandlerField:           &statusHandlerMock.AppStatusHandlerStub{},
		WatchdogField:                &testscommon.WatchdogMock{},
		AlarmSchedulerField:          &testscommon.AlarmSchedulerStub{},
		SyncTimerField:               &testscommon.SyncTimerStub{},
		RoundHandlerField:            customComponents.NewRound(args.GenesisTimeStamp, args.RoundDuration),
		EconomicsDataField:           economicsData,
		APIEconomicsHandler:          apiEconomicsData,
		RatingsDataField:             &testscommon.RatingsInfoMock{},
		RaterField:                   &testscommon.RaterMock{},
		GenesisNodesSetupField:       &testscommon.NodesSetupStub{}, // TODO: replace with the real component
		NodesShufflerField:           nil,                           // TODO: replace with the real component
		EpochNotifierField:           epochNotifier,
		EnableRoundsHandlerField:     enableRoundsHandler,
		GenesisTimeField:             args.GenesisTimeStamp,
		TxVersionCheckField:          versioning.NewTxVersionChecker(minTransactionVersion),
		NodeTypeProviderField:        &nodeTypeProviderMock.NodeTypeProviderStub{},
		WasmVMChangeLockerInternal:   &sync.RWMutex{},
		ProcessStatusHandlerInternal: &testscommon.ProcessStatusHandlerStub{},
		HardforkTriggerPubKeyField:   bytes.Repeat([]byte{1}, validatorPubKeyLength),
		EnableEpochsHandlerField:     enableEpochsHandler,
	}

	return nil
}

func (node *ProcessOnlyNode) initDataComponents() error {
	var blockchainInstance data.ChainHandler
	if node.ShardCoordinator.SelfId() == core.MetachainShardId {
		blockchainInstance = node.initMetaChain()
	} else {
		blockchainInstance = node.initShardChain()
	}

	numShards := node.ShardCoordinator.NumberOfShards()
	shardID := node.ShardCoordinator.SelfId()

	node.DataComponents = &mock.DataComponentsStub{
		BlockChain: blockchainInstance,
		Store:      integrationTests.CreateStore(numShards),
		DataPool:   dataRetrieverMock.CreatePoolsHolder(numShards, shardID),
		MbProvider: &mock.MiniBlocksProviderStub{},
	}

	return nil
}

func (node *ProcessOnlyNode) initShardChain() data.ChainHandler {
	blockChain, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = blockChain.SetGenesisHeader(&dataBlock.Header{})
	genesisHeader, _ := node.CoreComponents.InternalMarshalizer().Marshal(blockChain.GetGenesisHeader())
	genesisHeaderHash := node.CoreComponents.Hasher().Compute(string(genesisHeader))
	blockChain.SetGenesisHeaderHash(genesisHeaderHash)

	return blockChain
}

func (node *ProcessOnlyNode) initMetaChain() data.ChainHandler {
	metaChain, _ := blockchain.NewMetaChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = metaChain.SetGenesisHeader(&dataBlock.MetaBlock{})
	genesisHeader, _ := node.CoreComponents.InternalMarshalizer().Marshal(metaChain.GetGenesisHeader())
	genesisHeaderHash := node.CoreComponents.Hasher().Compute(string(genesisHeader))
	metaChain.SetGenesisHeaderHash(genesisHeaderHash)

	return metaChain
}

func (node *ProcessOnlyNode) initStatusComponents() {
	node.StatusComponents = &mock.StatusComponentsStub{
		Outport:              mock.NewNilOutport(),
		SoftwareVersionCheck: &mock.SoftwareVersionCheckerMock{},
		AppStatusHandler:     &statusHandlerMock.AppStatusHandlerStub{},
	}
}

func (node *ProcessOnlyNode) initBootstrapComponents() error {
	var versionedHeaderFactory nodeFactory.VersionedHeaderFactory
	headerVersionHandler := &testscommon.HeaderVersionHandlerStub{}
	versionedHeaderFactory, _ = hdrFactory.NewShardHeaderFactory(headerVersionHandler)
	if node.ShardCoordinator.SelfId() == core.MetachainShardId {
		versionedHeaderFactory, _ = hdrFactory.NewMetaHeaderFactory(headerVersionHandler)
	}

	node.BootstrapComponents = &mainFactoryMocks.BootstrapComponentsStub{
		Bootstrapper: &bootstrapMocks.EpochStartBootstrapperStub{
			TrieHolder:      &trieMock.TriesHolderStub{},
			StorageManagers: map[string]common.StorageManager{"0": &testscommon.StorageManagerStub{}},
			BootstrapCalled: nil,
		},
		BootstrapParams:      &bootstrapMocks.BootstrapParamsHandlerMock{},
		NodeRole:             "",
		ShCoordinator:        node.ShardCoordinator,
		HdrVersionHandler:    headerVersionHandler,
		VersionedHdrFactory:  versionedHeaderFactory,
		HdrIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
	}

	return nil
}

func (node *ProcessOnlyNode) initAccountDBs() {
	node.TrieContainer = state.NewDataTriesHolder()

	userAccountsStore := integrationTests.CreateMemUnit()
	userTrieStorageManager, _ := integrationTests.CreateTrieStorageManager(userAccountsStore)
	userAccounts, stateTrie := integrationTests.CreateAccountsDB(integrationTests.UserAccount, userTrieStorageManager)
	node.TrieContainer.Put([]byte(trieFactory.UserAccountTrie), stateTrie)
	node.Accounts[state.UserAccountsState] = userAccounts

	validatorAccountsStore := integrationTests.CreateMemUnit()
	validatorTrieStorageManager, _ := integrationTests.CreateTrieStorageManager(validatorAccountsStore)
	peerAccounts, peerTrie := integrationTests.CreateAccountsDB(integrationTests.ValidatorAccount, validatorTrieStorageManager)
	node.TrieContainer.Put([]byte(trieFactory.PeerAccountTrie), peerTrie)
	node.Accounts[state.PeerAccountsState] = peerAccounts

	node.TrieStorageManagers = make(map[string]common.StorageManager)
	node.TrieStorageManagers[trieFactory.UserAccountTrie] = userTrieStorageManager
	node.TrieStorageManagers[trieFactory.PeerAccountTrie] = validatorTrieStorageManager
}

func (node *ProcessOnlyNode) initWhiteListHandlers() error {
	cacherCfg := storageunit.CacheConfig{Capacity: 10000, Type: storageunit.LRUCache, Shards: 1}
	suCache, err := storageunit.NewCache(cacherCfg)
	if err != nil {
		return err
	}

	node.WhiteListHandler, err = interceptors.NewWhiteListDataVerifier(suCache)
	if err != nil {
		return err
	}

	cacherVerifiedCfg := storageunit.CacheConfig{Capacity: 5000, Type: storageunit.LRUCache, Shards: 1}
	cacheVerified, err := storageunit.NewCache(cacherVerifiedCfg)
	if err != nil {
		return err
	}

	node.WhiteListerVerifiedTxs, err = interceptors.NewWhiteListDataVerifier(cacheVerified)

	return err
}

func (node *ProcessOnlyNode) initBlockProcessor() error {
	blockProcessorConfig := config.Config{
		StateTriesConfig: config.StateTriesConfig{
			CheckpointRoundsModulus: 100,
		},
		Debug: config.DebugConfig{
			EpochStart: config.EpochStartDebugConfig{
				ProcessDataTrieOnCommitEpoch: false,
			},
		},
	}

	err := node.initHeaderValidator()
	if err != nil {
		return err
	}

	if node.ShardCoordinator.SelfId() == core.MetachainShardId {
		return node.initMetaBlockProcessor(blockProcessorConfig)
	}

	return node.initShardBlockProcessor(blockProcessorConfig)
}

func (node *ProcessOnlyNode) initHeaderValidator() error {
	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:      node.CoreComponents.Hasher(),
		Marshalizer: node.CoreComponents.InternalMarshalizer(),
	}

	var err error
	node.HeaderValidator, err = block.NewHeaderValidator(argsHeaderValidator)

	return err
}

func (node *ProcessOnlyNode) initMetaBlockProcessor(blockProcessorConfig config.Config) error {
	// TODO: implement this
	return nil
}

func (node *ProcessOnlyNode) initShardBlockProcessor(blockProcessorConfig config.Config) error {
	err := node.initBlockTracker()
	if err != nil {
		return err
	}

	blockBlackListHandler := cache.NewTimeCache(integrationTests.TimeSpanForBadHeaders)
	forkDetector, err := processSync.NewShardForkDetector(
		node.CoreComponents.RoundHandler(),
		blockBlackListHandler,
		node.BlockTracker,
		node.GenesisTimeStamp.Unix(),
	)
	if err != nil {
		return err
	}

	args := block.ArgShardProcessor{
		ArgBaseProcessor: block.ArgBaseProcessor{
			CoreComponents:                 node.CoreComponents,
			DataComponents:                 node.DataComponents,
			BootstrapComponents:            node.BootstrapComponents,
			StatusComponents:               node.StatusComponents,
			StatusCoreComponents:           node.StatusCoreComponents,
			Config:                         blockProcessorConfig,
			AccountsDB:                     node.Accounts,
			ForkDetector:                   forkDetector,
			NodesCoordinator:               nil,
			FeeHandler:                     nil,
			RequestHandler:                 nil,
			BlockChainHook:                 nil,
			TxCoordinator:                  nil,
			EpochStartTrigger:              nil,
			HeaderValidator:                nil,
			BootStorer:                     nil,
			BlockTracker:                   nil,
			BlockSizeThrottler:             nil,
			Version:                        "",
			HistoryRepository:              nil,
			EnableRoundsHandler:            nil,
			VMContainersFactory:            nil,
			VmContainer:                    nil,
			GasHandler:                     nil,
			OutportDataProvider:            nil,
			ScheduledTxsExecutionHandler:   nil,
			ScheduledMiniBlocksEnableEpoch: 0,
			ProcessedMiniBlocksTracker:     nil,
			ReceiptsRepository:             nil,
		},
	}

	node.BlockProcessor, err = block.NewShardProcessor(args)

	return err
}

func (node *ProcessOnlyNode) initBlockTracker() error {
	argBaseTracker := track.ArgBaseTracker{
		Hasher:           node.CoreComponents.Hasher(),
		HeaderValidator:  node.HeaderValidator,
		Marshalizer:      node.CoreComponents.InternalMarshalizer(),
		RequestHandler:   node.RequestHandler,
		RoundHandler:     node.CoreComponents.RoundHandler(),
		ShardCoordinator: node.ShardCoordinator,
		Store:            node.DataComponents.Store,
		StartHeaders:     node.GenesisBlocks,
		PoolsHolder:      node.DataComponents.DataPool,
		WhitelistHandler: node.WhiteListHandler,
		FeeHandler:       node.CoreComponents.EconomicsData(),
	}

	var err error
	if node.ShardCoordinator.SelfId() != core.MetachainShardId {
		arguments := track.ArgShardTracker{
			ArgBaseTracker: argBaseTracker,
		}

		node.BlockTracker, err = track.NewShardBlockTrack(arguments)
	} else {
		arguments := track.ArgMetaTracker{
			ArgBaseTracker: argBaseTracker,
		}

		node.BlockTracker, err = track.NewMetaBlockTrack(arguments)
	}

	return err
}
