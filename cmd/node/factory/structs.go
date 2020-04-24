package factory

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/core/serviceContainer"
	"github.com/ElrondNetwork/elrond-go/core/statistics/softwareVersion"
	factorySoftwareVersion "github.com/ElrondNetwork/elrond-go/core/statistics/softwareVersion/factory"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519/singlesig"
	mclmultisig "github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/multisig"
	mclsig "github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/state"
	factoryState "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	dataRetrieverFactory "github.com/ElrondNetwork/elrond-go/dataRetriever/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	metachainEpochStart "github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/epochStart/shardchain"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/parser"
	genesisProcess "github.com/ElrondNetwork/elrond-go/genesis/process"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	factoryHasher "github.com/ElrondNetwork/elrond-go/hashing/factory"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/marshal"
	factoryMarshalizer "github.com/ElrondNetwork/elrond-go/marshal/factory"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/block/pendingMb"
	"github.com/ElrondNetwork/elrond-go/process/block/poolsCleaner"
	"github.com/ElrondNetwork/elrond-go/process/block/postprocess"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory/interceptorscontainer"
	"github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/headerCheck"
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	"github.com/ElrondNetwork/elrond-go/process/scToProtocol"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	processSync "github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/process/throttle"
	antifloodFactory "github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/factory"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/process/transactionLog"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/networksharding"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/vm"
	systemVM "github.com/ElrondNetwork/elrond-go/vm/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/urfave/cli"
)

const (
	// MaxTxsToRequest specifies the maximum number of txs to request
	MaxTxsToRequest = 1000
)

//TODO remove this
var log = logger.GetOrCreate("main")

// timeSpanForBadHeaders is the expiry time for an added block header hash
var timeSpanForBadHeaders = time.Minute * 2

// EpochStartNotifier defines which actions should be done for handling new epoch's events
type EpochStartNotifier interface {
	RegisterHandler(handler epochStart.ActionHandler)
	UnregisterHandler(handler epochStart.ActionHandler)
	NotifyAll(hdr data.HeaderHandler)
	NotifyAllPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
	IsInterfaceNil() bool
}

// CryptoParams is a DTO for holding block signing parameters
type CryptoParams struct {
	KeyGenerator    crypto.KeyGenerator
	PrivateKey      crypto.PrivateKey
	PublicKey       crypto.PublicKey
	PublicKeyBytes  []byte
	PublicKeyString string
}

// Network struct holds the network components of the Elrond protocol
type Network struct {
	NetMessenger           p2p.Messenger
	InputAntifloodHandler  P2PAntifloodHandler
	OutputAntifloodHandler P2PAntifloodHandler
	PeerBlackListHandler   process.BlackListHandler
}

// Core struct holds the core components of the Elrond protocol
type Core struct {
	Hasher                   hashing.Hasher
	InternalMarshalizer      marshal.Marshalizer
	VmMarshalizer            marshal.Marshalizer
	TxSignMarshalizer        marshal.Marshalizer
	TriesContainer           state.TriesHolder
	TrieStorageManagers      map[string]data.StorageManager
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	StatusHandler            core.AppStatusHandler
	ChainID                  []byte
}

// State struct holds the state components of the Elrond protocol
type State struct {
	AddressPubkeyConverter   state.PubkeyConverter
	ValidatorPubkeyConverter state.PubkeyConverter
	PeerAccounts             state.AccountsAdapter
	AccountsAdapter          state.AccountsAdapter
	AccountsForShard         []*genesis.InitialAccount
}

// Data struct holds the data components of the Elrond protocol
type Data struct {
	Blkc     data.ChainHandler
	Store    dataRetriever.StorageService
	Datapool dataRetriever.PoolsHolder
}

// Crypto struct holds the crypto components of the Elrond protocol
type Crypto struct {
	TxSingleSigner      crypto.SingleSigner
	SingleSigner        crypto.SingleSigner
	MultiSigner         crypto.MultiSigner
	BlockSignKeyGen     crypto.KeyGenerator
	TxSignKeyGen        crypto.KeyGenerator
	InitialPubKeys      map[uint32][]string
	MessageSignVerifier vm.MessageSignVerifier
}

// Process struct holds the process components of the Elrond protocol
type Process struct {
	InterceptorsContainer    process.InterceptorsContainer
	ResolversFinder          dataRetriever.ResolversFinder
	Rounder                  consensus.Rounder
	EpochStartTrigger        epochStart.TriggerHandler
	ForkDetector             process.ForkDetector
	BlockProcessor           process.BlockProcessor
	BlackListHandler         process.BlackListHandler
	BootStorer               process.BootStorer
	HeaderSigVerifier        HeaderSigVerifierHandler
	ValidatorsStatistics     process.ValidatorStatisticsProcessor
	ValidatorsProvider       process.ValidatorsProvider
	BlockTracker             process.BlockTracker
	PendingMiniBlocksHandler process.PendingMiniBlocksHandler
	RequestHandler           process.RequestHandler
}

type coreComponentsFactoryArgs struct {
	config      *config.Config
	pathManager storage.PathManagerHandler
	shardId     string
	chainID     []byte
}

// NewCoreComponentsFactoryArgs initializes the arguments necessary for creating the core components
func NewCoreComponentsFactoryArgs(config *config.Config, pathManager storage.PathManagerHandler, shardId string, chainID []byte) *coreComponentsFactoryArgs {
	return &coreComponentsFactoryArgs{
		config:      config,
		pathManager: pathManager,
		shardId:     shardId,
		chainID:     chainID,
	}
}

// CoreComponentsFactory creates the core components
func CoreComponentsFactory(args *coreComponentsFactoryArgs) (*Core, error) {
	hasher, err := factoryHasher.NewHasher(args.config.Hasher.Type)
	if err != nil {
		return nil, errors.New("could not create hasher: " + err.Error())
	}

	internalMarshalizer, err := factoryMarshalizer.NewMarshalizer(args.config.Marshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w for internalMarshalizer", err)
	}

	vmMarshalizer, err := factoryMarshalizer.NewMarshalizer(args.config.VmMarshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w for vmMarshalizer", err)
	}

	txSignMarshalizer, err := factoryMarshalizer.NewMarshalizer(args.config.TxSignMarshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w for txSignMarshalizer", err)
	}

	uint64ByteSliceConverter := uint64ByteSlice.NewBigEndianConverter()

	trieStorageManagers, trieContainer, err := createTries(args, internalMarshalizer, hasher)

	if err != nil {
		return nil, err
	}

	return &Core{
		Hasher:                   hasher,
		InternalMarshalizer:      internalMarshalizer,
		VmMarshalizer:            vmMarshalizer,
		TxSignMarshalizer:        txSignMarshalizer,
		TriesContainer:           trieContainer,
		TrieStorageManagers:      trieStorageManagers,
		Uint64ByteSliceConverter: uint64ByteSliceConverter,
		StatusHandler:            statusHandler.NewNilStatusHandler(),
		ChainID:                  args.chainID,
	}, nil
}

func createTries(
	args *coreComponentsFactoryArgs,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
) (map[string]data.StorageManager, state.TriesHolder, error) {

	trieContainer := state.NewDataTriesHolder()
	trieFactoryArgs := factory.TrieFactoryArgs{
		EvictionWaitingListCfg:   args.config.EvictionWaitingList,
		SnapshotDbCfg:            args.config.TrieSnapshotDB,
		Marshalizer:              marshalizer,
		Hasher:                   hasher,
		PathManager:              args.pathManager,
		TrieStorageManagerConfig: args.config.TrieStorageManagerConfig,
	}
	trieFactory, err := factory.NewTrieFactory(trieFactoryArgs)
	if err != nil {
		return nil, nil, err
	}

	trieStorageManagers := make(map[string]data.StorageManager)
	userStorageManager, userAccountTrie, err := trieFactory.Create(
		args.config.AccountsTrieStorage,
		args.shardId,
		args.config.StateTriesConfig.AccountsStatePruningEnabled,
	)
	if err != nil {
		return nil, nil, err
	}
	trieContainer.Put([]byte(factory.UserAccountTrie), userAccountTrie)
	trieStorageManagers[factory.UserAccountTrie] = userStorageManager

	peerStorageManager, peerAccountsTrie, err := trieFactory.Create(
		args.config.PeerAccountsTrieStorage,
		args.shardId,
		args.config.StateTriesConfig.PeerStatePruningEnabled,
	)
	if err != nil {
		return nil, nil, err
	}
	trieContainer.Put([]byte(factory.PeerAccountTrie), peerAccountsTrie)
	trieStorageManagers[factory.PeerAccountTrie] = peerStorageManager

	return trieStorageManagers, trieContainer, nil
}

type stateComponentsFactoryArgs struct {
	config           *config.Config
	genesisParser    *parser.Genesis
	shardCoordinator sharding.Coordinator
	core             *Core
	pathManager      storage.PathManagerHandler
}

// NewStateComponentsFactoryArgs initializes the arguments necessary for creating the state components
func NewStateComponentsFactoryArgs(
	config *config.Config,
	genesisParser *parser.Genesis,
	shardCoordinator sharding.Coordinator,
	core *Core,
	pathManager storage.PathManagerHandler,
) *stateComponentsFactoryArgs {
	return &stateComponentsFactoryArgs{
		config:           config,
		genesisParser:    genesisParser,
		shardCoordinator: shardCoordinator,
		core:             core,
		pathManager:      pathManager,
	}
}

// StateComponentsFactory creates the state components
func StateComponentsFactory(args *stateComponentsFactoryArgs) (*State, error) {
	processPubkeyConverter, err := factoryState.NewPubkeyConverter(args.config.AddressPubkeyConverter)
	if err != nil {
		return nil, fmt.Errorf("%w for ProcessPubkeyConverter", err)
	}

	validatorPubkeyConverter, err := factoryState.NewPubkeyConverter(args.config.ValidatorPubkeyConverter)
	if err != nil {
		return nil, fmt.Errorf("%w for ValidatorPubkeyConverter", err)
	}

	accountFactory := factoryState.NewAccountCreator()
	merkleTrie := args.core.TriesContainer.Get([]byte(factory.UserAccountTrie))
	accountsAdapter, err := state.NewAccountsDB(merkleTrie, args.core.Hasher, args.core.InternalMarshalizer, accountFactory)
	if err != nil {
		return nil, errors.New("could not create accounts adapter: " + err.Error())
	}

	accountsForShard, err := args.genesisParser.InitialAccountsSplitOnAddressesShards(args.shardCoordinator)
	if err != nil {
		return nil, errors.New("initial balances could not be processed " + err.Error())
	}

	accountFactory = factoryState.NewPeerAccountCreator()
	merkleTrie = args.core.TriesContainer.Get([]byte(factory.PeerAccountTrie))
	peerAdapter, err := state.NewPeerAccountsDB(merkleTrie, args.core.Hasher, args.core.InternalMarshalizer, accountFactory)
	if err != nil {
		return nil, err
	}

	return &State{
		PeerAccounts:             peerAdapter,
		AddressPubkeyConverter:   processPubkeyConverter,
		ValidatorPubkeyConverter: validatorPubkeyConverter,
		AccountsAdapter:          accountsAdapter,
		AccountsForShard:         accountsForShard[args.shardCoordinator.SelfId()],
	}, nil
}

type dataComponentsFactoryArgs struct {
	config             *config.Config
	economicsData      *economics.EconomicsData
	shardCoordinator   sharding.Coordinator
	core               *Core
	pathManager        storage.PathManagerHandler
	epochStartNotifier EpochStartNotifier
	currentEpoch       uint32
}

// NewDataComponentsFactoryArgs initializes the arguments necessary for creating the data components
func NewDataComponentsFactoryArgs(
	config *config.Config,
	economicsData *economics.EconomicsData,
	shardCoordinator sharding.Coordinator,
	core *Core,
	pathManager storage.PathManagerHandler,
	epochStartNotifier EpochStartNotifier,
	currentEpoch uint32,
) *dataComponentsFactoryArgs {
	return &dataComponentsFactoryArgs{
		config:             config,
		economicsData:      economicsData,
		shardCoordinator:   shardCoordinator,
		core:               core,
		pathManager:        pathManager,
		epochStartNotifier: epochStartNotifier,
		currentEpoch:       currentEpoch,
	}
}

// DataComponentsFactory creates the data components
func DataComponentsFactory(args *dataComponentsFactoryArgs) (*Data, error) {
	var datapool dataRetriever.PoolsHolder
	blkc, err := createBlockChainFromConfig(args.shardCoordinator, args.core.StatusHandler)
	if err != nil {
		return nil, errors.New("could not create block chain: " + err.Error())
	}

	store, err := createDataStoreFromConfig(
		args.config,
		args.shardCoordinator,
		args.pathManager,
		args.epochStartNotifier,
		args.currentEpoch,
	)
	if err != nil {
		return nil, errors.New("could not create local data store: " + err.Error())
	}

	dataPoolArgs := dataRetrieverFactory.ArgsDataPool{
		Config:           args.config,
		EconomicsData:    args.economicsData,
		ShardCoordinator: args.shardCoordinator,
	}
	datapool, err = dataRetrieverFactory.NewDataPoolFromConfig(dataPoolArgs)
	if err != nil {
		return nil, errors.New("could not create data pools: ")
	}

	return &Data{
		Blkc:     blkc,
		Store:    store,
		Datapool: datapool,
	}, nil
}

type cryptoComponentsFactoryArgs struct {
	ctx              *cli.Context
	config           *config.Config
	nodesConfig      *sharding.NodesSetup
	shardCoordinator sharding.Coordinator
	keyGen           crypto.KeyGenerator
	privKey          crypto.PrivateKey
	log              logger.Logger
}

// NewCryptoComponentsFactoryArgs initializes the arguments necessary for creating the crypto components
func NewCryptoComponentsFactoryArgs(
	ctx *cli.Context,
	config *config.Config,
	nodesConfig *sharding.NodesSetup,
	shardCoordinator sharding.Coordinator,
	keyGen crypto.KeyGenerator,
	privKey crypto.PrivateKey,
	log logger.Logger,
) *cryptoComponentsFactoryArgs {
	return &cryptoComponentsFactoryArgs{
		ctx:              ctx,
		config:           config,
		nodesConfig:      nodesConfig,
		shardCoordinator: shardCoordinator,
		keyGen:           keyGen,
		privKey:          privKey,
		log:              log,
	}
}

// CryptoComponentsFactory creates the crypto components
func CryptoComponentsFactory(args *cryptoComponentsFactoryArgs) (*Crypto, error) {
	initialPubKeys := args.nodesConfig.InitialNodesPubKeys()
	txSingleSigner := &singlesig.Ed25519Signer{}
	singleSigner, err := createSingleSigner(args.config)
	if err != nil {
		return nil, errors.New("could not create singleSigner: " + err.Error())
	}

	multisigHasher, err := getMultisigHasherFromConfig(args.config)
	if err != nil {
		return nil, errors.New("could not create multisig hasher: " + err.Error())
	}

	currentShardNodesPubKeys, err := args.nodesConfig.InitialEligibleNodesPubKeysForShard(args.shardCoordinator.SelfId())
	if err != nil {
		return nil, errors.New("could not start creation of multiSigner: " + err.Error())
	}

	multiSigner, err := createMultiSigner(args.config, multisigHasher, currentShardNodesPubKeys, args.privKey, args.keyGen)
	if err != nil {
		return nil, err
	}

	txSignKeyGen := signing.NewKeyGenerator(ed25519.NewEd25519())

	messageSignVerifier, err := systemVM.NewMessageSigVerifier(args.keyGen, singleSigner)
	if err != nil {
		return nil, err
	}

	return &Crypto{
		TxSingleSigner:      txSingleSigner,
		SingleSigner:        singleSigner,
		MultiSigner:         multiSigner,
		BlockSignKeyGen:     args.keyGen,
		TxSignKeyGen:        txSignKeyGen,
		InitialPubKeys:      initialPubKeys,
		MessageSignVerifier: messageSignVerifier,
	}, nil
}

// NetworkComponentsFactory creates the network components
func NetworkComponentsFactory(
	p2pConfig config.P2PConfig,
	mainConfig config.Config,
	statusHandler core.AppStatusHandler,
) (*Network, error) {

	arg := libp2p.ArgsNetworkMessenger{
		Context:       context.Background(),
		ListenAddress: libp2p.ListenAddrWithIp4AndTcp,
		P2pConfig:     p2pConfig,
	}

	netMessenger, err := libp2p.NewNetworkMessenger(arg)
	if err != nil {
		return nil, err
	}

	inAntifloodHandler, p2pPeerBlackList, errNewAntiflood := antifloodFactory.NewP2PAntiFloodAndBlackList(mainConfig, statusHandler)
	if errNewAntiflood != nil {
		return nil, errNewAntiflood
	}

	inputAntifloodHandler, ok := inAntifloodHandler.(P2PAntifloodHandler)
	if !ok {
		return nil, fmt.Errorf("%w when casting input antiflood handler to structs/P2PAntifloodHandler", errWrongTypeAssertion)
	}

	outAntifloodHandler, errOutputAntiflood := antifloodFactory.NewP2POutputAntiFlood(mainConfig)
	if errOutputAntiflood != nil {
		return nil, errOutputAntiflood
	}

	outputAntifloodHandler, ok := outAntifloodHandler.(P2PAntifloodHandler)
	if !ok {
		return nil, fmt.Errorf("%w when casting output antiflood handler to structs/P2PAntifloodHandler", errWrongTypeAssertion)
	}

	err = netMessenger.SetPeerBlackListHandler(p2pPeerBlackList)
	if err != nil {
		return nil, err
	}

	return &Network{
		NetMessenger:           netMessenger,
		InputAntifloodHandler:  inputAntifloodHandler,
		OutputAntifloodHandler: outputAntifloodHandler,
		PeerBlackListHandler:   p2pPeerBlackList,
	}, nil
}

type processComponentsFactoryArgs struct {
	coreComponents            *coreComponentsFactoryArgs
	genesisParser             *parser.Genesis
	economicsData             *economics.EconomicsData
	nodesConfig               *sharding.NodesSetup
	gasSchedule               map[string]map[string]uint64
	rounder                   consensus.Rounder
	shardCoordinator          sharding.Coordinator
	nodesCoordinator          sharding.NodesCoordinator
	data                      *Data
	coreData                  *Core
	crypto                    *Crypto
	state                     *State
	network                   *Network
	coreServiceContainer      serviceContainer.Core
	requestedItemsHandler     dataRetriever.RequestedItemsHandler
	whiteListHandler          process.WhiteListHandler
	epochStartNotifier        EpochStartNotifier
	epochStart                *config.EpochStartConfig
	rater                     sharding.PeerAccountListAndRatingHandler
	ratingsData               process.RatingsInfoHandler
	startEpochNum             uint32
	sizeCheckDelta            uint32
	stateCheckpointModulus    uint
	maxComputableRounds       uint64
	numConcurrentResolverJobs int32
	minSizeInBytes            uint32
	maxSizeInBytes            uint32
	maxRating                 uint32
	validatorPubkeyConverter  state.PubkeyConverter
}

// NewProcessComponentsFactoryArgs initializes the arguments necessary for creating the process components
func NewProcessComponentsFactoryArgs(
	coreComponents *coreComponentsFactoryArgs,
	genesisParser *parser.Genesis,
	economicsData *economics.EconomicsData,
	nodesConfig *sharding.NodesSetup,
	gasSchedule map[string]map[string]uint64,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	data *Data,
	coreData *Core,
	crypto *Crypto,
	state *State,
	network *Network,
	coreServiceContainer serviceContainer.Core,
	requestedItemsHandler dataRetriever.RequestedItemsHandler,
	whiteListHandler process.WhiteListHandler,
	epochStartNotifier EpochStartNotifier,
	epochStart *config.EpochStartConfig,
	startEpochNum uint32,
	rater sharding.PeerAccountListAndRatingHandler,
	sizeCheckDelta uint32,
	stateCheckpointModulus uint,
	maxComputableRounds uint64,
	numConcurrentResolverJobs int32,
	minSizeInBytes uint32,
	maxSizeInBytes uint32,
	maxRating uint32,
	validatorPubkeyConverter state.PubkeyConverter,
	ratingsData process.RatingsInfoHandler,
) *processComponentsFactoryArgs {
	return &processComponentsFactoryArgs{
		coreComponents:            coreComponents,
		genesisParser:             genesisParser,
		economicsData:             economicsData,
		nodesConfig:               nodesConfig,
		gasSchedule:               gasSchedule,
		rounder:                   rounder,
		shardCoordinator:          shardCoordinator,
		nodesCoordinator:          nodesCoordinator,
		data:                      data,
		coreData:                  coreData,
		crypto:                    crypto,
		state:                     state,
		network:                   network,
		coreServiceContainer:      coreServiceContainer,
		requestedItemsHandler:     requestedItemsHandler,
		whiteListHandler:          whiteListHandler,
		epochStartNotifier:        epochStartNotifier,
		epochStart:                epochStart,
		startEpochNum:             startEpochNum,
		rater:                     rater,
		ratingsData:               ratingsData,
		sizeCheckDelta:            sizeCheckDelta,
		stateCheckpointModulus:    stateCheckpointModulus,
		maxComputableRounds:       maxComputableRounds,
		numConcurrentResolverJobs: numConcurrentResolverJobs,
		minSizeInBytes:            minSizeInBytes,
		maxSizeInBytes:            maxSizeInBytes,
		maxRating:                 maxRating,
		validatorPubkeyConverter:  validatorPubkeyConverter,
	}
}

// ProcessComponentsFactory creates the process components
func ProcessComponentsFactory(args *processComponentsFactoryArgs) (*Process, error) {
	argsHeaderSig := &headerCheck.ArgsHeaderSigVerifier{
		Marshalizer:       args.coreData.InternalMarshalizer,
		Hasher:            args.coreData.Hasher,
		NodesCoordinator:  args.nodesCoordinator,
		MultiSigVerifier:  args.crypto.MultiSigner,
		SingleSigVerifier: args.crypto.SingleSigner,
		KeyGen:            args.crypto.BlockSignKeyGen,
	}
	headerSigVerifier, err := headerCheck.NewHeaderSigVerifier(argsHeaderSig)
	if err != nil {
		return nil, err
	}

	resolversContainerFactory, err := newResolverContainerFactory(
		args.shardCoordinator,
		args.data,
		args.coreData,
		args.network,
		args.sizeCheckDelta,
		args.numConcurrentResolverJobs,
	)
	if err != nil {
		return nil, err
	}

	resolversContainer, err := resolversContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	resolversFinder, err := containers.NewResolversFinder(resolversContainer, args.shardCoordinator)
	if err != nil {
		return nil, err
	}

	requestHandler, err := requestHandlers.NewResolverRequestHandler(
		resolversFinder,
		args.requestedItemsHandler,
		args.whiteListHandler,
		MaxTxsToRequest,
		args.shardCoordinator.SelfId(),
		time.Second,
	)
	if err != nil {
		return nil, err
	}

	validatorStatisticsProcessor, err := newValidatorStatisticsProcessor(args)
	if err != nil {
		return nil, err
	}

	validatorsProvider, err := peer.NewValidatorsProvider(
		validatorStatisticsProcessor,
		args.maxRating,
		args.validatorPubkeyConverter,
	)
	if err != nil {
		return nil, err
	}

	epochStartTrigger, err := newEpochStartTrigger(args, requestHandler)
	if err != nil {
		return nil, err
	}

	requestHandler.SetEpoch(epochStartTrigger.Epoch())

	err = dataRetriever.SetEpochHandlerToHdrResolver(resolversContainer, epochStartTrigger)
	if err != nil {
		return nil, err
	}

	validatorStatsRootHash, err := validatorStatisticsProcessor.RootHash()
	if err != nil {
		return nil, err
	}

	log.Trace("Validator stats created", "validatorStatsRootHash", validatorStatsRootHash)

	txLogsStorage := args.data.Store.GetStorer(dataRetriever.TxLogsUnit)
	txLogsProcessor, err := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer:      txLogsStorage,
		Marshalizer: args.coreData.InternalMarshalizer,
	})
	if err != nil {
		return nil, err
	}

	genesisBlocks, err := generateGenesisHeadersAndApplyInitialBalances(args)
	if err != nil {
		return nil, err
	}

	err = prepareGenesisBlock(args, genesisBlocks)
	if err != nil {
		return nil, err
	}

	bootStr := args.data.Store.GetStorer(dataRetriever.BootstrapUnit)
	bootStorer, err := bootstrapStorage.NewBootstrapStorer(args.coreData.InternalMarshalizer, bootStr)
	if err != nil {
		return nil, err
	}

	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:      args.coreData.Hasher,
		Marshalizer: args.coreData.InternalMarshalizer,
	}
	headerValidator, err := block.NewHeaderValidator(argsHeaderValidator)
	if err != nil {
		return nil, err
	}

	blockTracker, err := newBlockTracker(
		args,
		headerValidator,
		requestHandler,
		args.rounder,
		genesisBlocks,
	)
	if err != nil {
		return nil, err
	}

	_, err = poolsCleaner.NewMiniBlocksPoolsCleaner(
		args.data.Datapool.MiniBlocks(),
		args.rounder,
		args.shardCoordinator,
	)
	if err != nil {
		return nil, err
	}

	_, err = poolsCleaner.NewTxsPoolsCleaner(
		args.state.AddressPubkeyConverter,
		args.data.Datapool,
		args.rounder,
		args.shardCoordinator,
	)
	if err != nil {
		return nil, err
	}

	interceptorContainerFactory, blackListHandler, err := newInterceptorContainerFactory(
		args.shardCoordinator,
		args.nodesCoordinator,
		args.data,
		args.coreData,
		args.crypto,
		args.state,
		args.network,
		args.economicsData,
		headerSigVerifier,
		args.sizeCheckDelta,
		blockTracker,
		epochStartTrigger,
		args.whiteListHandler,
	)
	if err != nil {
		return nil, err
	}

	//TODO refactor all these factory calls
	interceptorsContainer, err := interceptorContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	var pendingMiniBlocksHandler process.PendingMiniBlocksHandler
	if args.shardCoordinator.SelfId() == core.MetachainShardId {
		pendingMiniBlocksHandler, err = pendingMb.NewPendingMiniBlocks()
		if err != nil {
			return nil, err
		}
	}

	forkDetector, err := newForkDetector(
		args.rounder,
		args.shardCoordinator,
		blackListHandler,
		blockTracker,
		args.nodesConfig.StartTime,
	)
	if err != nil {
		return nil, err
	}

	blockProcessor, err := newBlockProcessor(
		args,
		requestHandler,
		forkDetector,
		epochStartTrigger,
		bootStorer,
		validatorStatisticsProcessor,
		headerValidator,
		blockTracker,
		pendingMiniBlocksHandler,
		txLogsProcessor,
	)
	if err != nil {
		return nil, err
	}

	return &Process{
		InterceptorsContainer:    interceptorsContainer,
		ResolversFinder:          resolversFinder,
		Rounder:                  args.rounder,
		ForkDetector:             forkDetector,
		BlockProcessor:           blockProcessor,
		EpochStartTrigger:        epochStartTrigger,
		BlackListHandler:         blackListHandler,
		BootStorer:               bootStorer,
		HeaderSigVerifier:        headerSigVerifier,
		ValidatorsStatistics:     validatorStatisticsProcessor,
		ValidatorsProvider:       validatorsProvider,
		BlockTracker:             blockTracker,
		PendingMiniBlocksHandler: pendingMiniBlocksHandler,
		RequestHandler:           requestHandler,
	}, nil
}

func prepareGenesisBlock(args *processComponentsFactoryArgs, genesisBlocks map[uint32]data.HeaderHandler) error {
	genesisBlock, ok := genesisBlocks[args.shardCoordinator.SelfId()]
	if !ok {
		return errors.New("genesis block does not exists")
	}

	genesisBlockHash, err := core.CalculateHash(args.coreData.InternalMarshalizer, args.coreData.Hasher, genesisBlock)
	if err != nil {
		return err
	}

	err = args.data.Blkc.SetGenesisHeader(genesisBlock)
	if err != nil {
		return err
	}

	args.data.Blkc.SetGenesisHeaderHash(genesisBlockHash)

	marshalizedBlock, err := args.coreData.InternalMarshalizer.Marshal(genesisBlock)
	if err != nil {
		return err
	}

	if args.shardCoordinator.SelfId() == core.MetachainShardId {
		errNotCritical := args.data.Store.Put(dataRetriever.MetaBlockUnit, genesisBlockHash, marshalizedBlock)
		if errNotCritical != nil {
			log.Error("error storing genesis metablock", "error", errNotCritical.Error())
		}
	} else {
		errNotCritical := args.data.Store.Put(dataRetriever.BlockHeaderUnit, genesisBlockHash, marshalizedBlock)
		if errNotCritical != nil {
			log.Error("error storing genesis shardblock", "error", errNotCritical.Error())
		}
	}

	return nil
}

func newEpochStartTrigger(
	args *processComponentsFactoryArgs,
	requestHandler process.RequestHandler,
) (epochStart.TriggerHandler, error) {
	if args.shardCoordinator.SelfId() < args.shardCoordinator.NumberOfShards() {
		argsHeaderValidator := block.ArgsHeaderValidator{
			Hasher:      args.coreData.Hasher,
			Marshalizer: args.coreData.InternalMarshalizer,
		}
		headerValidator, err := block.NewHeaderValidator(argsHeaderValidator)
		if err != nil {
			return nil, err
		}

		argsPeerMiniBlockSyncer := shardchain.ArgPeerMiniBlockSyncer{
			MiniBlocksPool: args.data.Datapool.MiniBlocks(),
			Requesthandler: requestHandler,
		}

		peerMiniBlockSyncer, err := shardchain.NewPeerMiniBlockSyncer(argsPeerMiniBlockSyncer)
		if err != nil {
			return nil, err
		}

		argEpochStart := &shardchain.ArgsShardEpochStartTrigger{
			Marshalizer:          args.coreData.InternalMarshalizer,
			Hasher:               args.coreData.Hasher,
			HeaderValidator:      headerValidator,
			Uint64Converter:      args.coreData.Uint64ByteSliceConverter,
			DataPool:             args.data.Datapool,
			Storage:              args.data.Store,
			RequestHandler:       requestHandler,
			Epoch:                0,
			EpochStartNotifier:   args.epochStartNotifier,
			Validity:             process.MetaBlockValidity,
			Finality:             process.BlockFinality,
			PeerMiniBlocksSyncer: peerMiniBlockSyncer,
		}
		epochStartTrigger, err := shardchain.NewEpochStartTrigger(argEpochStart)
		if err != nil {
			return nil, errors.New("error creating new start of epoch trigger" + err.Error())
		}
		err = epochStartTrigger.SetAppStatusHandler(args.coreData.StatusHandler)
		if err != nil {
			return nil, err
		}

		return epochStartTrigger, nil
	}

	if args.shardCoordinator.SelfId() == core.MetachainShardId {
		argEpochStart := &metachainEpochStart.ArgsNewMetaEpochStartTrigger{
			GenesisTime:        time.Unix(args.nodesConfig.StartTime, 0),
			Settings:           args.epochStart,
			Epoch:              0,
			EpochStartNotifier: args.epochStartNotifier,
			Storage:            args.data.Store,
			Marshalizer:        args.coreData.InternalMarshalizer,
			Hasher:             args.coreData.Hasher,
		}
		epochStartTrigger, err := metachainEpochStart.NewEpochStartTrigger(argEpochStart)
		if err != nil {
			return nil, errors.New("error creating new start of epoch trigger" + err.Error())
		}
		err = epochStartTrigger.SetAppStatusHandler(args.coreData.StatusHandler)
		if err != nil {
			return nil, err
		}

		return epochStartTrigger, nil
	}

	return nil, errors.New("error creating new start of epoch trigger because of invalid shard id")
}

// CreateSoftwareVersionChecker will create a new software version checker and will start check if a new software version
// is available
func CreateSoftwareVersionChecker(statusHandler core.AppStatusHandler) (*softwareVersion.SoftwareVersionChecker, error) {
	softwareVersionCheckerFactory, err := factorySoftwareVersion.NewSoftwareVersionFactory(statusHandler)
	if err != nil {
		return nil, err
	}

	softwareVersionChecker, err := softwareVersionCheckerFactory.Create()
	if err != nil {
		return nil, err
	}

	return softwareVersionChecker, nil
}

func createBlockChainFromConfig(coordinator sharding.Coordinator, ash core.AppStatusHandler) (data.ChainHandler, error) {

	if coordinator == nil {
		return nil, state.ErrNilShardCoordinator
	}

	if coordinator.SelfId() < coordinator.NumberOfShards() {
		blockChain := blockchain.NewBlockChain()

		err := blockChain.SetAppStatusHandler(ash)
		if err != nil {
			return nil, err
		}

		return blockChain, nil
	}
	if coordinator.SelfId() == core.MetachainShardId {
		blockChain := blockchain.NewMetaChain()

		err := blockChain.SetAppStatusHandler(ash)
		if err != nil {
			return nil, err
		}

		return blockChain, nil
	}
	return nil, errors.New("can not create blockchain")
}

func createDataStoreFromConfig(
	config *config.Config,
	shardCoordinator sharding.Coordinator,
	pathManager storage.PathManagerHandler,
	epochStartNotifier EpochStartNotifier,
	currentEpoch uint32,
) (dataRetriever.StorageService, error) {
	storageServiceFactory, err := storageFactory.NewStorageServiceFactory(
		config,
		shardCoordinator,
		pathManager,
		epochStartNotifier,
		currentEpoch,
	)
	if err != nil {
		return nil, err
	}
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return storageServiceFactory.CreateForShard()
	}
	if shardCoordinator.SelfId() == core.MetachainShardId {
		return storageServiceFactory.CreateForMeta()
	}
	return nil, errors.New("can not create data store")
}

func createSingleSigner(config *config.Config) (crypto.SingleSigner, error) {
	switch config.Consensus.Type {
	case consensus.BlsConsensusType:
		return &mclsig.BlsSingleSigner{}, nil
	default:
		return nil, errors.New("no consensus type provided in config file")
	}
}

func getMultisigHasherFromConfig(cfg *config.Config) (hashing.Hasher, error) {
	if cfg.Consensus.Type == consensus.BlsConsensusType && cfg.MultisigHasher.Type != "blake2b" {
		return nil, errors.New("wrong multisig hasher provided for bls consensus type")
	}

	switch cfg.MultisigHasher.Type {
	case "sha256":
		return sha256.Sha256{}, nil
	case "blake2b":
		if cfg.Consensus.Type == consensus.BlsConsensusType {
			return &blake2b.Blake2b{HashSize: multisig.BlsHashSize}, nil
		}
		return &blake2b.Blake2b{}, nil
	}

	return nil, errors.New("no multisig hasher provided in config file")
}

func createMultiSigner(
	config *config.Config,
	hasher hashing.Hasher,
	pubKeys []string,
	privateKey crypto.PrivateKey,
	keyGen crypto.KeyGenerator,
) (crypto.MultiSigner, error) {

	switch config.Consensus.Type {
	case consensus.BlsConsensusType:
		blsSigner := &mclmultisig.BlsMultiSigner{Hasher: hasher}
		return multisig.NewBLSMultisig(blsSigner, pubKeys, privateKey, keyGen, uint16(0))
	default:
		return nil, errors.New("no consensus type provided in config file")
	}
}

func newInterceptorContainerFactory(
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	data *Data,
	coreData *Core,
	crypto *Crypto,
	state *State,
	network *Network,
	economics *economics.EconomicsData,
	headerSigVerifier HeaderSigVerifierHandler,
	sizeCheckDelta uint32,
	validityAttester process.ValidityAttester,
	epochStartTrigger process.EpochStartTriggerHandler,
	whiteListHandler process.WhiteListHandler,
) (process.InterceptorsContainerFactory, process.BlackListHandler, error) {

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return newShardInterceptorContainerFactory(
			shardCoordinator,
			nodesCoordinator,
			data,
			coreData,
			crypto,
			state,
			network,
			economics,
			headerSigVerifier,
			sizeCheckDelta,
			validityAttester,
			epochStartTrigger,
			whiteListHandler,
		)
	}
	if shardCoordinator.SelfId() == core.MetachainShardId {
		return newMetaInterceptorContainerFactory(
			shardCoordinator,
			nodesCoordinator,
			data,
			coreData,
			crypto,
			network,
			state,
			economics,
			headerSigVerifier,
			sizeCheckDelta,
			validityAttester,
			epochStartTrigger,
			whiteListHandler,
		)
	}

	return nil, nil, errors.New("could not create interceptor container factory")
}

func newResolverContainerFactory(
	shardCoordinator sharding.Coordinator,
	data *Data,
	coreData *Core,
	network *Network,
	sizeCheckDelta uint32,
	numConcurrentResolverJobs int32,
) (dataRetriever.ResolversContainerFactory, error) {

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return newShardResolverContainerFactory(
			shardCoordinator,
			data,
			coreData,
			network,
			sizeCheckDelta,
			numConcurrentResolverJobs,
		)
	}
	if shardCoordinator.SelfId() == core.MetachainShardId {
		return newMetaResolverContainerFactory(
			shardCoordinator,
			data,
			coreData,
			network,
			sizeCheckDelta,
			numConcurrentResolverJobs,
		)
	}

	return nil, errors.New("could not create interceptor and resolver container factory")
}

func newShardInterceptorContainerFactory(
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	data *Data,
	dataCore *Core,
	crypto *Crypto,
	state *State,
	network *Network,
	economics *economics.EconomicsData,
	headerSigVerifier HeaderSigVerifierHandler,
	sizeCheckDelta uint32,
	validityAttester process.ValidityAttester,
	epochStartTrigger process.EpochStartTriggerHandler,
	whiteListHandler process.WhiteListHandler,
) (process.InterceptorsContainerFactory, process.BlackListHandler, error) {
	headerBlackList := timecache.NewTimeCache(timeSpanForBadHeaders)
	shardInterceptorsContainerFactoryArgs := interceptorscontainer.ShardInterceptorsContainerFactoryArgs{
		Accounts:               state.AccountsAdapter,
		ShardCoordinator:       shardCoordinator,
		NodesCoordinator:       nodesCoordinator,
		Messenger:              network.NetMessenger,
		Store:                  data.Store,
		ProtoMarshalizer:       dataCore.InternalMarshalizer,
		TxSignMarshalizer:      dataCore.TxSignMarshalizer,
		Hasher:                 dataCore.Hasher,
		KeyGen:                 crypto.TxSignKeyGen,
		BlockSignKeyGen:        crypto.BlockSignKeyGen,
		SingleSigner:           crypto.TxSingleSigner,
		BlockSingleSigner:      crypto.SingleSigner,
		MultiSigner:            crypto.MultiSigner,
		DataPool:               data.Datapool,
		AddressPubkeyConverter: state.AddressPubkeyConverter,
		MaxTxNonceDeltaAllowed: core.MaxTxNonceDeltaAllowed,
		TxFeeHandler:           economics,
		BlackList:              headerBlackList,
		HeaderSigVerifier:      headerSigVerifier,
		ChainID:                dataCore.ChainID,
		SizeCheckDelta:         sizeCheckDelta,
		ValidityAttester:       validityAttester,
		EpochStartTrigger:      epochStartTrigger,
		WhiteListHandler:       whiteListHandler,
		AntifloodHandler:       network.InputAntifloodHandler,
		NonceConverter:         dataCore.Uint64ByteSliceConverter,
	}
	interceptorContainerFactory, err := interceptorscontainer.NewShardInterceptorsContainerFactory(shardInterceptorsContainerFactoryArgs)
	if err != nil {
		return nil, nil, err
	}

	return interceptorContainerFactory, headerBlackList, nil
}

func newMetaInterceptorContainerFactory(
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	data *Data,
	dataCore *Core,
	crypto *Crypto,
	network *Network,
	state *State,
	economics *economics.EconomicsData,
	headerSigVerifier HeaderSigVerifierHandler,
	sizeCheckDelta uint32,
	validityAttester process.ValidityAttester,
	epochStartTrigger process.EpochStartTriggerHandler,
	whiteListHandler process.WhiteListHandler,
) (process.InterceptorsContainerFactory, process.BlackListHandler, error) {
	headerBlackList := timecache.NewTimeCache(timeSpanForBadHeaders)
	metaInterceptorsContainerFactoryArgs := interceptorscontainer.MetaInterceptorsContainerFactoryArgs{
		ShardCoordinator:       shardCoordinator,
		NodesCoordinator:       nodesCoordinator,
		Messenger:              network.NetMessenger,
		Store:                  data.Store,
		ProtoMarshalizer:       dataCore.InternalMarshalizer,
		TxSignMarshalizer:      dataCore.TxSignMarshalizer,
		Hasher:                 dataCore.Hasher,
		MultiSigner:            crypto.MultiSigner,
		DataPool:               data.Datapool,
		Accounts:               state.AccountsAdapter,
		AddressPubkeyConverter: state.AddressPubkeyConverter,
		SingleSigner:           crypto.TxSingleSigner,
		BlockSingleSigner:      crypto.SingleSigner,
		KeyGen:                 crypto.TxSignKeyGen,
		BlockKeyGen:            crypto.BlockSignKeyGen,
		MaxTxNonceDeltaAllowed: core.MaxTxNonceDeltaAllowed,
		TxFeeHandler:           economics,
		BlackList:              headerBlackList,
		HeaderSigVerifier:      headerSigVerifier,
		ChainID:                dataCore.ChainID,
		SizeCheckDelta:         sizeCheckDelta,
		ValidityAttester:       validityAttester,
		EpochStartTrigger:      epochStartTrigger,
		WhiteListHandler:       whiteListHandler,
		AntifloodHandler:       network.InputAntifloodHandler,
		NonceConverter:         dataCore.Uint64ByteSliceConverter,
	}
	interceptorContainerFactory, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(metaInterceptorsContainerFactoryArgs)
	if err != nil {
		return nil, nil, err
	}

	return interceptorContainerFactory, headerBlackList, nil
}

func newShardResolverContainerFactory(
	shardCoordinator sharding.Coordinator,
	data *Data,
	core *Core,
	network *Network,
	sizeCheckDelta uint32,
	numConcurrentResolverJobs int32,
) (dataRetriever.ResolversContainerFactory, error) {

	dataPacker, err := partitioning.NewSimpleDataPacker(core.InternalMarshalizer)
	if err != nil {
		return nil, err
	}

	resolversContainerFactoryArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:           shardCoordinator,
		Messenger:                  network.NetMessenger,
		Store:                      data.Store,
		Marshalizer:                core.InternalMarshalizer,
		DataPools:                  data.Datapool,
		Uint64ByteSliceConverter:   core.Uint64ByteSliceConverter,
		DataPacker:                 dataPacker,
		TriesContainer:             core.TriesContainer,
		SizeCheckDelta:             sizeCheckDelta,
		InputAntifloodHandler:      network.InputAntifloodHandler,
		OutputAntifloodHandler:     network.OutputAntifloodHandler,
		NumConcurrentResolvingJobs: numConcurrentResolverJobs,
	}
	resolversContainerFactory, err := resolverscontainer.NewShardResolversContainerFactory(resolversContainerFactoryArgs)
	if err != nil {
		return nil, err
	}

	return resolversContainerFactory, nil
}

func newMetaResolverContainerFactory(
	shardCoordinator sharding.Coordinator,
	data *Data,
	core *Core,
	network *Network,
	sizeCheckDelta uint32,
	numConcurrentResolverJobs int32,
) (dataRetriever.ResolversContainerFactory, error) {
	dataPacker, err := partitioning.NewSimpleDataPacker(core.InternalMarshalizer)
	if err != nil {
		return nil, err
	}

	resolversContainerFactoryArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:           shardCoordinator,
		Messenger:                  network.NetMessenger,
		Store:                      data.Store,
		Marshalizer:                core.InternalMarshalizer,
		DataPools:                  data.Datapool,
		Uint64ByteSliceConverter:   core.Uint64ByteSliceConverter,
		DataPacker:                 dataPacker,
		TriesContainer:             core.TriesContainer,
		SizeCheckDelta:             sizeCheckDelta,
		InputAntifloodHandler:      network.InputAntifloodHandler,
		OutputAntifloodHandler:     network.OutputAntifloodHandler,
		NumConcurrentResolvingJobs: numConcurrentResolverJobs,
	}
	resolversContainerFactory, err := resolverscontainer.NewMetaResolversContainerFactory(resolversContainerFactoryArgs)
	if err != nil {
		return nil, err
	}
	return resolversContainerFactory, nil
}

func generateGenesisHeadersAndApplyInitialBalances(args *processComponentsFactoryArgs) (map[uint32]data.HeaderHandler, error) {
	coreComponents := args.coreData
	stateComponents := args.state
	dataComponents := args.data
	shardCoordinator := args.shardCoordinator
	nodesSetup := args.nodesConfig
	genesisParser := args.genesisParser
	economicsData := args.economicsData

	validatorStatsRootHash, err := stateComponents.PeerAccounts.RootHash()
	if err != nil {
		return nil, err
	}

	arg := genesisProcess.ArgsGenesisBlockCreator{
		GenesisTime:              uint64(nodesSetup.StartTime),
		StartEpochNum:            args.startEpochNum,
		Accounts:                 stateComponents.AccountsAdapter,
		PubkeyConv:               stateComponents.AddressPubkeyConverter,
		InitialNodesSetup:        nodesSetup,
		Economics:                economicsData,
		ShardCoordinator:         shardCoordinator,
		Store:                    dataComponents.Store,
		Blkc:                     dataComponents.Blkc,
		Marshalizer:              coreComponents.InternalMarshalizer,
		Hasher:                   coreComponents.Hasher,
		Uint64ByteSliceConverter: coreComponents.Uint64ByteSliceConverter,
		DataPool:                 dataComponents.Datapool,
		GenesisParser:            genesisParser,
		ValidatorStatsRootHash:   validatorStatsRootHash,
		GasMap:                   args.gasSchedule,
	}

	gbc, err := genesisProcess.NewGenesisBlockCreator(arg)
	if err != nil {
		return nil, err
	}

	return gbc.CreateGenesisBlocks()
}

func newBlockTracker(
	processArgs *processComponentsFactoryArgs,
	headerValidator process.HeaderConstructionValidator,
	requestHandler process.RequestHandler,
	rounder process.Rounder,
	genesisBlocks map[uint32]data.HeaderHandler,
) (process.BlockTracker, error) {

	argBaseTracker := track.ArgBaseTracker{
		Hasher:           processArgs.coreData.Hasher,
		HeaderValidator:  headerValidator,
		Marshalizer:      processArgs.coreData.InternalMarshalizer,
		RequestHandler:   requestHandler,
		Rounder:          rounder,
		ShardCoordinator: processArgs.shardCoordinator,
		Store:            processArgs.data.Store,
		StartHeaders:     genesisBlocks,
		PoolsHolder:      processArgs.data.Datapool,
		WhitelistHandler: processArgs.whiteListHandler,
	}

	if processArgs.shardCoordinator.SelfId() < processArgs.shardCoordinator.NumberOfShards() {
		arguments := track.ArgShardTracker{
			ArgBaseTracker: argBaseTracker,
		}

		return track.NewShardBlockTrack(arguments)
	}

	if processArgs.shardCoordinator.SelfId() == core.MetachainShardId {
		arguments := track.ArgMetaTracker{
			ArgBaseTracker: argBaseTracker,
		}

		return track.NewMetaBlockTrack(arguments)
	}

	return nil, errors.New("could not create block tracker")
}

func newForkDetector(
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	headerBlackList process.BlackListHandler,
	blockTracker process.BlockTracker,
	genesisTime int64,
) (process.ForkDetector, error) {
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return processSync.NewShardForkDetector(rounder, headerBlackList, blockTracker, genesisTime)
	}
	if shardCoordinator.SelfId() == core.MetachainShardId {
		return processSync.NewMetaForkDetector(rounder, headerBlackList, blockTracker, genesisTime)
	}

	return nil, errors.New("could not create fork detector")
}

func newBlockProcessor(
	processArgs *processComponentsFactoryArgs,
	requestHandler process.RequestHandler,
	forkDetector process.ForkDetector,
	epochStartTrigger epochStart.TriggerHandler,
	bootStorer process.BootStorer,
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor,
	headerValidator process.HeaderConstructionValidator,
	blockTracker process.BlockTracker,
	pendingMiniBlocksHandler process.PendingMiniBlocksHandler,
	txLogsProcessor process.TransactionLogProcessor,
) (process.BlockProcessor, error) {

	shardCoordinator := processArgs.shardCoordinator

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return newShardBlockProcessor(
			processArgs.coreComponents.config,
			requestHandler,
			processArgs.shardCoordinator,
			processArgs.nodesCoordinator,
			processArgs.data,
			processArgs.coreData,
			processArgs.state,
			forkDetector,
			processArgs.coreServiceContainer,
			processArgs.economicsData,
			processArgs.rounder,
			epochStartTrigger,
			bootStorer,
			processArgs.gasSchedule,
			processArgs.stateCheckpointModulus,
			headerValidator,
			blockTracker,
			processArgs.minSizeInBytes,
			processArgs.maxSizeInBytes,
			txLogsProcessor,
		)
	}
	if shardCoordinator.SelfId() == core.MetachainShardId {
		return newMetaBlockProcessor(
			requestHandler,
			processArgs.shardCoordinator,
			processArgs.nodesCoordinator,
			processArgs.data,
			processArgs.coreData,
			processArgs.state,
			forkDetector,
			processArgs.coreServiceContainer,
			processArgs.economicsData,
			validatorStatisticsProcessor,
			processArgs.rounder,
			epochStartTrigger,
			bootStorer,
			headerValidator,
			blockTracker,
			pendingMiniBlocksHandler,
			processArgs.stateCheckpointModulus,
			processArgs.crypto.MessageSignVerifier,
			processArgs.gasSchedule,
			processArgs.minSizeInBytes,
			processArgs.maxSizeInBytes,
			processArgs.ratingsData,
			processArgs.nodesConfig,
			txLogsProcessor,
		)
	}

	return nil, errors.New("could not create block processor")
}

func newShardBlockProcessor(
	config *config.Config,
	requestHandler process.RequestHandler,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	data *Data,
	core *Core,
	stateComponents *State,
	forkDetector process.ForkDetector,
	coreServiceContainer serviceContainer.Core,
	economics *economics.EconomicsData,
	rounder consensus.Rounder,
	epochStartTrigger epochStart.TriggerHandler,
	bootStorer process.BootStorer,
	gasSchedule map[string]map[string]uint64,
	stateCheckpointModulus uint,
	headerValidator process.HeaderConstructionValidator,
	blockTracker process.BlockTracker,
	minSizeInBytes uint32,
	maxSizeInBytes uint32,
	txLogsProcessor process.TransactionLogProcessor,
) (process.BlockProcessor, error) {
	argsParser := vmcommon.NewAtArgumentParser()

	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasMap:          gasSchedule,
		MapDNSAddresses: make(map[string]struct{}),
	}
	builtInFuncs, err := builtInFunctions.CreateBuiltInFunctionContainer(argsBuiltIn)
	if err != nil {
		return nil, err
	}

	argsHook := hooks.ArgBlockChainHook{
		Accounts:         stateComponents.AccountsAdapter,
		PubkeyConv:       stateComponents.AddressPubkeyConverter,
		StorageService:   data.Store,
		BlockChain:       data.Blkc,
		ShardCoordinator: shardCoordinator,
		Marshalizer:      core.InternalMarshalizer,
		Uint64Converter:  core.Uint64ByteSliceConverter,
		BuiltInFunctions: builtInFuncs,
	}
	vmFactory, err := shard.NewVMContainerFactory(config.VirtualMachineConfig, economics.MaxGasLimitPerBlock(), gasSchedule, argsHook)
	if err != nil {
		return nil, err
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, err
	}

	interimProcFactory, err := shard.NewIntermediateProcessorsContainerFactory(
		shardCoordinator,
		core.InternalMarshalizer,
		core.Hasher,
		stateComponents.AddressPubkeyConverter,
		data.Store,
		data.Datapool,
		economics,
	)
	if err != nil {
		return nil, err
	}

	interimProcContainer, err := interimProcFactory.Create()
	if err != nil {
		return nil, err
	}

	scForwarder, err := interimProcContainer.Get(dataBlock.SmartContractResultBlock)
	if err != nil {
		return nil, err
	}

	receiptTxInterim, err := interimProcContainer.Get(dataBlock.ReceiptBlock)
	if err != nil {
		return nil, err
	}

	badTxInterim, err := interimProcContainer.Get(dataBlock.InvalidBlock)
	if err != nil {
		return nil, err
	}

	gasHandler, err := preprocess.NewGasComputation(economics)
	if err != nil {
		return nil, err
	}

	txFeeHandler, err := postprocess.NewFeeAccumulator()
	if err != nil {
		return nil, err
	}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:  stateComponents.AddressPubkeyConverter,
		ShardCoordinator: shardCoordinator,
		BuiltInFuncNames: builtInFuncs.Keys(),
		ArgumentParser:   vmcommon.NewAtArgumentParser(),
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	if err != nil {
		return nil, err
	}

	argsNewScProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:      vmContainer,
		ArgsParser:       argsParser,
		Hasher:           core.Hasher,
		Marshalizer:      core.InternalMarshalizer,
		AccountsDB:       stateComponents.AccountsAdapter,
		TempAccounts:     vmFactory.BlockChainHookImpl(),
		PubkeyConv:       stateComponents.AddressPubkeyConverter,
		Coordinator:      shardCoordinator,
		ScrForwarder:     scForwarder,
		TxFeeHandler:     txFeeHandler,
		EconomicsFee:     economics,
		TxTypeHandler:    txTypeHandler,
		GasHandler:       gasHandler,
		BuiltInFunctions: vmFactory.BlockChainHookImpl().GetBuiltInFunctions(),
		TxLogsProcessor:  txLogsProcessor,
	}
	scProcessor, err := smartContract.NewSmartContractProcessor(argsNewScProcessor)
	if err != nil {
		return nil, err
	}

	rewardsTxProcessor, err := rewardTransaction.NewRewardTxProcessor(
		stateComponents.AccountsAdapter,
		stateComponents.AddressPubkeyConverter,
		shardCoordinator,
	)
	if err != nil {
		return nil, err
	}

	transactionProcessor, err := transaction.NewTxProcessor(
		stateComponents.AccountsAdapter,
		core.Hasher,
		stateComponents.AddressPubkeyConverter,
		core.InternalMarshalizer,
		shardCoordinator,
		scProcessor,
		txFeeHandler,
		txTypeHandler,
		economics,
		receiptTxInterim,
		badTxInterim,
	)
	if err != nil {
		return nil, errors.New("could not create transaction statisticsProcessor: " + err.Error())
	}

	blockSizeThrottler, err := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)
	if err != nil {
		return nil, err
	}

	blockSizeComputationHandler, err := preprocess.NewBlockSizeComputation(core.InternalMarshalizer, blockSizeThrottler, maxSizeInBytes)
	if err != nil {
		return nil, err
	}

	balanceComputationHandler, err := preprocess.NewBalanceComputation()
	if err != nil {
		return nil, err
	}

	preProcFactory, err := shard.NewPreProcessorsContainerFactory(
		shardCoordinator,
		data.Store,
		core.InternalMarshalizer,
		core.Hasher,
		data.Datapool,
		stateComponents.AddressPubkeyConverter,
		stateComponents.AccountsAdapter,
		requestHandler,
		transactionProcessor,
		scProcessor,
		scProcessor,
		rewardsTxProcessor,
		economics,
		gasHandler,
		blockTracker,
		blockSizeComputationHandler,
		balanceComputationHandler,
	)
	if err != nil {
		return nil, err
	}

	preProcContainer, err := preProcFactory.Create()
	if err != nil {
		return nil, err
	}

	txCoordinator, err := coordinator.NewTransactionCoordinator(
		core.Hasher,
		core.InternalMarshalizer,
		shardCoordinator,
		stateComponents.AccountsAdapter,
		data.Datapool.MiniBlocks(),
		requestHandler,
		preProcContainer,
		interimProcContainer,
		gasHandler,
		txFeeHandler,
		blockSizeComputationHandler,
		balanceComputationHandler,
	)
	if err != nil {
		return nil, err
	}

	accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDb[state.UserAccountsState] = stateComponents.AccountsAdapter

	argumentsBaseProcessor := block.ArgBaseProcessor{
		AccountsDB:             accountsDb,
		ForkDetector:           forkDetector,
		Hasher:                 core.Hasher,
		Marshalizer:            core.InternalMarshalizer,
		Store:                  data.Store,
		ShardCoordinator:       shardCoordinator,
		NodesCoordinator:       nodesCoordinator,
		Uint64Converter:        core.Uint64ByteSliceConverter,
		RequestHandler:         requestHandler,
		Core:                   coreServiceContainer,
		BlockChainHook:         vmFactory.BlockChainHookImpl(),
		TxCoordinator:          txCoordinator,
		Rounder:                rounder,
		EpochStartTrigger:      epochStartTrigger,
		HeaderValidator:        headerValidator,
		BootStorer:             bootStorer,
		BlockTracker:           blockTracker,
		DataPool:               data.Datapool,
		FeeHandler:             txFeeHandler,
		BlockChain:             data.Blkc,
		StateCheckpointModulus: stateCheckpointModulus,
		BlockSizeThrottler:     blockSizeThrottler,
	}
	arguments := block.ArgShardProcessor{
		ArgBaseProcessor: argumentsBaseProcessor,
	}

	blockProcessor, err := block.NewShardProcessor(arguments)
	if err != nil {
		return nil, errors.New("could not create block statisticsProcessor: " + err.Error())
	}

	err = blockProcessor.SetAppStatusHandler(core.StatusHandler)
	if err != nil {
		return nil, err
	}

	return blockProcessor, nil
}

func newMetaBlockProcessor(
	requestHandler process.RequestHandler,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	data *Data,
	core *Core,
	stateComponents *State,
	forkDetector process.ForkDetector,
	coreServiceContainer serviceContainer.Core,
	economicsData *economics.EconomicsData,
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor,
	rounder consensus.Rounder,
	epochStartTrigger epochStart.TriggerHandler,
	bootStorer process.BootStorer,
	headerValidator process.HeaderConstructionValidator,
	blockTracker process.BlockTracker,
	pendingMiniBlocksHandler process.PendingMiniBlocksHandler,
	stateCheckpointModulus uint,
	messageSignVerifier vm.MessageSignVerifier,
	gasSchedule map[string]map[string]uint64,
	minSizeInBytes uint32,
	maxSizeInBytes uint32,
	ratingsData process.RatingsInfoHandler,
	nodesSetup sharding.GenesisNodesSetupHandler,
	txLogsProcessor process.TransactionLogProcessor,
) (process.BlockProcessor, error) {

	builtInFuncs := builtInFunctions.NewBuiltInFunctionContainer()
	argsHook := hooks.ArgBlockChainHook{
		Accounts:         stateComponents.AccountsAdapter,
		PubkeyConv:       stateComponents.AddressPubkeyConverter,
		StorageService:   data.Store,
		BlockChain:       data.Blkc,
		ShardCoordinator: shardCoordinator,
		Marshalizer:      core.InternalMarshalizer,
		Uint64Converter:  core.Uint64ByteSliceConverter,
		BuiltInFunctions: builtInFuncs, // no built-in functions for meta.
	}
	vmFactory, err := metachain.NewVMContainerFactory(
		argsHook,
		economicsData,
		messageSignVerifier,
		gasSchedule,
		nodesSetup,
	)
	if err != nil {
		return nil, err
	}

	argsParser := vmcommon.NewAtArgumentParser()

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, err
	}

	interimProcFactory, err := metachain.NewIntermediateProcessorsContainerFactory(
		shardCoordinator,
		core.InternalMarshalizer,
		core.Hasher,
		stateComponents.AddressPubkeyConverter,
		data.Store,
		data.Datapool,
	)
	if err != nil {
		return nil, err
	}

	interimProcContainer, err := interimProcFactory.Create()
	if err != nil {
		return nil, err
	}

	scForwarder, err := interimProcContainer.Get(dataBlock.SmartContractResultBlock)
	if err != nil {
		return nil, err
	}

	gasHandler, err := preprocess.NewGasComputation(economicsData)
	if err != nil {
		return nil, err
	}

	txFeeHandler, err := postprocess.NewFeeAccumulator()
	if err != nil {
		return nil, err
	}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:  stateComponents.AddressPubkeyConverter,
		ShardCoordinator: shardCoordinator,
		BuiltInFuncNames: builtInFuncs.Keys(),
		ArgumentParser:   vmcommon.NewAtArgumentParser(),
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	if err != nil {
		return nil, err
	}

	argsNewScProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:      vmContainer,
		ArgsParser:       argsParser,
		Hasher:           core.Hasher,
		Marshalizer:      core.InternalMarshalizer,
		AccountsDB:       stateComponents.AccountsAdapter,
		TempAccounts:     vmFactory.BlockChainHookImpl(),
		PubkeyConv:       stateComponents.AddressPubkeyConverter,
		Coordinator:      shardCoordinator,
		ScrForwarder:     scForwarder,
		TxFeeHandler:     txFeeHandler,
		EconomicsFee:     economicsData,
		TxTypeHandler:    txTypeHandler,
		GasHandler:       gasHandler,
		BuiltInFunctions: vmFactory.BlockChainHookImpl().GetBuiltInFunctions(),
		TxLogsProcessor:  txLogsProcessor,
	}
	scProcessor, err := smartContract.NewSmartContractProcessor(argsNewScProcessor)
	if err != nil {
		return nil, err
	}

	transactionProcessor, err := transaction.NewMetaTxProcessor(
		core.Hasher,
		core.InternalMarshalizer,
		stateComponents.AccountsAdapter,
		stateComponents.AddressPubkeyConverter,
		shardCoordinator,
		scProcessor,
		txTypeHandler,
		economicsData,
	)
	if err != nil {
		return nil, errors.New("could not create transaction processor: " + err.Error())
	}

	blockSizeThrottler, err := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)
	if err != nil {
		return nil, err
	}

	blockSizeComputationHandler, err := preprocess.NewBlockSizeComputation(core.InternalMarshalizer, blockSizeThrottler, maxSizeInBytes)
	if err != nil {
		return nil, err
	}

	balanceComputationHandler, err := preprocess.NewBalanceComputation()
	if err != nil {
		return nil, err
	}

	preProcFactory, err := metachain.NewPreProcessorsContainerFactory(
		shardCoordinator,
		data.Store,
		core.InternalMarshalizer,
		core.Hasher,
		data.Datapool,
		stateComponents.AccountsAdapter,
		requestHandler,
		transactionProcessor,
		scProcessor,
		economicsData,
		gasHandler,
		blockTracker,
		stateComponents.AddressPubkeyConverter,
		blockSizeComputationHandler,
		balanceComputationHandler,
	)
	if err != nil {
		return nil, err
	}

	preProcContainer, err := preProcFactory.Create()
	if err != nil {
		return nil, err
	}

	txCoordinator, err := coordinator.NewTransactionCoordinator(
		core.Hasher,
		core.InternalMarshalizer,
		shardCoordinator,
		stateComponents.AccountsAdapter,
		data.Datapool.MiniBlocks(),
		requestHandler,
		preProcContainer,
		interimProcContainer,
		gasHandler,
		txFeeHandler,
		blockSizeComputationHandler,
		balanceComputationHandler,
	)
	if err != nil {
		return nil, err
	}

	scDataGetter, err := smartContract.NewSCQueryService(vmContainer, economicsData)
	if err != nil {
		return nil, err
	}

	argsStaking := scToProtocol.ArgStakingToPeer{
		PubkeyConv:       stateComponents.ValidatorPubkeyConverter,
		Hasher:           core.Hasher,
		ProtoMarshalizer: core.InternalMarshalizer,
		VmMarshalizer:    core.VmMarshalizer,
		PeerState:        stateComponents.PeerAccounts,
		BaseState:        stateComponents.AccountsAdapter,
		ArgParser:        argsParser,
		CurrTxs:          data.Datapool.CurrentBlockTxs(),
		ScQuery:          scDataGetter,
		RatingsData:      ratingsData,
	}
	smartContractToProtocol, err := scToProtocol.NewStakingToPeer(argsStaking)
	if err != nil {
		return nil, err
	}

	argsEpochStartData := metachainEpochStart.ArgsNewEpochStartData{
		Marshalizer:       core.InternalMarshalizer,
		Hasher:            core.Hasher,
		Store:             data.Store,
		DataPool:          data.Datapool,
		BlockTracker:      blockTracker,
		ShardCoordinator:  shardCoordinator,
		EpochStartTrigger: epochStartTrigger,
		RequestHandler:    requestHandler,
	}
	epochStartDataCreator, err := metachainEpochStart.NewEpochStartData(argsEpochStartData)
	if err != nil {
		return nil, err
	}

	argsEpochEconomics := metachainEpochStart.ArgsNewEpochEconomics{
		Marshalizer:         core.InternalMarshalizer,
		Hasher:              core.Hasher,
		Store:               data.Store,
		ShardCoordinator:    shardCoordinator,
		NodesConfigProvider: nodesCoordinator,
		RewardsHandler:      economicsData,
		RoundTime:           rounder,
	}
	epochEconomics, err := metachainEpochStart.NewEndOfEpochEconomicsDataCreator(argsEpochEconomics)
	if err != nil {
		return nil, err
	}

	rewardsStorage := data.Store.GetStorer(dataRetriever.RewardTransactionUnit)
	miniBlockStorage := data.Store.GetStorer(dataRetriever.MiniBlockUnit)
	argsEpochRewards := metachainEpochStart.ArgsNewRewardsCreator{
		ShardCoordinator: shardCoordinator,
		PubkeyConverter:  stateComponents.AddressPubkeyConverter,
		RewardsStorage:   rewardsStorage,
		MiniBlockStorage: miniBlockStorage,
		Hasher:           core.Hasher,
		Marshalizer:      core.InternalMarshalizer,
		DataPool:         data.Datapool,
	}
	epochRewards, err := metachainEpochStart.NewEpochStartRewardsCreator(argsEpochRewards)
	if err != nil {
		return nil, err
	}

	argsEpochValidatorInfo := metachainEpochStart.ArgsNewValidatorInfoCreator{
		ShardCoordinator: shardCoordinator,
		MiniBlockStorage: miniBlockStorage,
		Hasher:           core.Hasher,
		Marshalizer:      core.InternalMarshalizer,
		DataPool:         data.Datapool,
	}
	validatorInfoCreator, err := metachainEpochStart.NewValidatorInfoCreator(argsEpochValidatorInfo)
	if err != nil {
		return nil, err
	}

	accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDb[state.UserAccountsState] = stateComponents.AccountsAdapter
	accountsDb[state.PeerAccountsState] = stateComponents.PeerAccounts

	argumentsBaseProcessor := block.ArgBaseProcessor{
		AccountsDB:             accountsDb,
		ForkDetector:           forkDetector,
		Hasher:                 core.Hasher,
		Marshalizer:            core.InternalMarshalizer,
		Store:                  data.Store,
		ShardCoordinator:       shardCoordinator,
		NodesCoordinator:       nodesCoordinator,
		Uint64Converter:        core.Uint64ByteSliceConverter,
		RequestHandler:         requestHandler,
		Core:                   coreServiceContainer,
		BlockChainHook:         vmFactory.BlockChainHookImpl(),
		TxCoordinator:          txCoordinator,
		EpochStartTrigger:      epochStartTrigger,
		Rounder:                rounder,
		HeaderValidator:        headerValidator,
		BootStorer:             bootStorer,
		BlockTracker:           blockTracker,
		DataPool:               data.Datapool,
		FeeHandler:             txFeeHandler,
		BlockChain:             data.Blkc,
		StateCheckpointModulus: stateCheckpointModulus,
		BlockSizeThrottler:     blockSizeThrottler,
	}
	arguments := block.ArgMetaProcessor{
		ArgBaseProcessor:             argumentsBaseProcessor,
		SCDataGetter:                 scDataGetter,
		SCToProtocol:                 smartContractToProtocol,
		PendingMiniBlocksHandler:     pendingMiniBlocksHandler,
		EpochStartDataCreator:        epochStartDataCreator,
		EpochEconomics:               epochEconomics,
		EpochRewardsCreator:          epochRewards,
		EpochValidatorInfoCreator:    validatorInfoCreator,
		ValidatorStatisticsProcessor: validatorStatisticsProcessor,
	}

	metaProcessor, err := block.NewMetaProcessor(arguments)
	if err != nil {
		return nil, errors.New("could not create block processor: " + err.Error())
	}

	err = metaProcessor.SetAppStatusHandler(core.StatusHandler)
	if err != nil {
		return nil, err
	}

	return metaProcessor, nil
}

func newValidatorStatisticsProcessor(
	processComponents *processComponentsFactoryArgs,
) (process.ValidatorStatisticsProcessor, error) {

	storageService := processComponents.data.Store

	var peerDataPool peer.DataPool = processComponents.data.Datapool
	if processComponents.shardCoordinator.SelfId() < processComponents.shardCoordinator.NumberOfShards() {
		peerDataPool = processComponents.data.Datapool
	}

	arguments := peer.ArgValidatorStatisticsProcessor{
		PeerAdapter:         processComponents.state.PeerAccounts,
		PubkeyConv:          processComponents.state.ValidatorPubkeyConverter,
		NodesCoordinator:    processComponents.nodesCoordinator,
		ShardCoordinator:    processComponents.shardCoordinator,
		DataPool:            peerDataPool,
		StorageService:      storageService,
		Marshalizer:         processComponents.coreData.InternalMarshalizer,
		StakeValue:          processComponents.economicsData.GenesisNodePrice(),
		Rater:               processComponents.rater,
		MaxComputableRounds: processComponents.maxComputableRounds,
		RewardsHandler:      processComponents.economicsData,
		StartEpoch:          processComponents.startEpochNum,
		NodesSetup:          processComponents.nodesConfig,
	}

	validatorStatisticsProcessor, err := peer.NewValidatorStatisticsProcessor(arguments)
	if err != nil {
		return nil, err
	}

	return validatorStatisticsProcessor, nil
}

// PrepareNetworkShardingCollector will create the network sharding collector and apply it to the network messenger
func PrepareNetworkShardingCollector(
	network *Network,
	config *config.Config,
	nodesCoordinator sharding.NodesCoordinator,
	coordinator sharding.Coordinator,
	epochStartRegistrationHandler epochStart.RegistrationHandler,
	epochShard uint32,
) (*networksharding.PeerShardMapper, error) {

	networkShardingCollector, err := createNetworkShardingCollector(config, nodesCoordinator, epochStartRegistrationHandler, epochShard)
	if err != nil {
		return nil, err
	}

	localId := network.NetMessenger.ID()
	networkShardingCollector.UpdatePeerIdShardId(localId, coordinator.SelfId())

	err = network.NetMessenger.SetPeerShardResolver(networkShardingCollector)
	if err != nil {
		return nil, err
	}

	return networkShardingCollector, nil
}

func createNetworkShardingCollector(
	config *config.Config,
	nodesCoordinator sharding.NodesCoordinator,
	epochStartRegistrationHandler epochStart.RegistrationHandler,
	epochStart uint32,
) (*networksharding.PeerShardMapper, error) {

	cacheConfig := config.PublicKeyPeerId
	cachePkPid, err := createCache(cacheConfig)
	if err != nil {
		return nil, err
	}

	cacheConfig = config.PublicKeyShardId
	cachePkShardId, err := createCache(cacheConfig)
	if err != nil {
		return nil, err
	}

	cacheConfig = config.PeerIdShardId
	cachePidShardId, err := createCache(cacheConfig)
	if err != nil {
		return nil, err
	}

	psm, err := networksharding.NewPeerShardMapper(
		cachePkPid,
		cachePkShardId,
		cachePidShardId,
		nodesCoordinator,
		epochStart,
	)
	if err != nil {
		return nil, err
	}

	epochStartRegistrationHandler.RegisterHandler(psm)

	return psm, nil
}

func createCache(cacheConfig config.CacheConfig) (storage.Cacher, error) {
	return storageUnit.NewCache(storageUnit.CacheType(cacheConfig.Type), cacheConfig.Size, cacheConfig.Shards)
}

// GetSigningParams returns a key generator, a private key, and a public key
func GetSigningParams(
	ctx *cli.Context,
	pubkeyConverter state.PubkeyConverter,
	skName string,
	skIndexName string,
	skPemFileName string,
	suite crypto.Suite,
) (*CryptoParams, error) {

	cryptoParams := &CryptoParams{}
	sk, readPk, err := getSkPk(ctx, pubkeyConverter, skName, skIndexName, skPemFileName)
	if err != nil {
		return nil, err
	}

	cryptoParams.KeyGenerator = signing.NewKeyGenerator(suite)
	cryptoParams.PrivateKey, err = cryptoParams.KeyGenerator.PrivateKeyFromByteArray(sk)
	if err != nil {
		return nil, err
	}

	cryptoParams.PublicKey = cryptoParams.PrivateKey.GeneratePublic()
	if len(readPk) > 0 {

		cryptoParams.PublicKeyBytes, err = cryptoParams.PublicKey.ToByteArray()
		if err != nil {
			return nil, err
		}

		if !bytes.Equal(cryptoParams.PublicKeyBytes, readPk) {
			return nil, errPublicKeyMismatch
		}
	}

	cryptoParams.PublicKeyString = pubkeyConverter.Encode(cryptoParams.PublicKeyBytes)

	return cryptoParams, nil
}

func getSkPk(
	ctx *cli.Context,
	pubkeyConverter state.PubkeyConverter,
	skName string,
	skIndexName string,
	skPemFileName string,
) ([]byte, []byte, error) {

	//if flag is defined, it shall overwrite what was read from pem file
	if ctx.GlobalIsSet(skName) {
		encodedSk := []byte(ctx.GlobalString(skName))
		sk, err := hex.DecodeString(string(encodedSk))

		return sk, nil, err
	}

	skIndex := ctx.GlobalInt(skIndexName)
	encodedSk, pkString, err := core.LoadSkPkFromPemFile(skPemFileName, skIndex)
	if err != nil {
		return nil, nil, err
	}

	skBytes, err := hex.DecodeString(string(encodedSk))
	if err != nil {
		return nil, nil, fmt.Errorf("%w for encoded secret key", err)
	}

	pkBytes, err := pubkeyConverter.Decode(pkString)
	if err != nil {
		return nil, nil, fmt.Errorf("%w for encoded public key %s", err, pkString)
	}

	return skBytes, pkBytes, nil
}
