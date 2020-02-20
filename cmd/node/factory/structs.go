package factory

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/round"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/core/serviceContainer"
	"github.com/ElrondNetwork/elrond-go/core/statistics/softwareVersion"
	factorySoftwareVersion "github.com/ElrondNetwork/elrond-go/core/statistics/softwareVersion/factory"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	blsMultiSig "github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/multisig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/address"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	factoryState "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool/headersCache"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	txpoolFactory "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/txpool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/txpool"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/genesis"
	metachainEpochStart "github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/epochStart/shardchain"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	factoryP2P "github.com/ElrondNetwork/elrond-go/p2p/libp2p/factory"
	"github.com/ElrondNetwork/elrond-go/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/block/poolsCleaner"
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
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	processSync "github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/urfave/cli"
)

const (
	// BlsHashSize specifies the hash size for using bls scheme
	BlsHashSize = 16

	// BlsConsensusType specifies te signature scheme used in the consensus
	BlsConsensusType = "bls"

	// MaxTxsToRequest specifies the maximum number of txs to request
	MaxTxsToRequest = 100
)

//TODO remove this
var log = logger.GetOrCreate("main")

// ErrCreateForkDetector signals that a fork detector could not be created
//TODO: Extract all others error messages from this file in some defined errors
var ErrCreateForkDetector = errors.New("could not create fork detector")

// timeSpanForBadHeaders is the expiry time for an added block header hash
var timeSpanForBadHeaders = time.Minute * 2

// Network struct holds the network components of the Elrond protocol
type Network struct {
	NetMessenger p2p.Messenger
}

// Core struct holds the core components of the Elrond protocol
type Core struct {
	Hasher                   hashing.Hasher
	Marshalizer              marshal.Marshalizer
	TriesContainer           state.TriesHolder
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	StatusHandler            core.AppStatusHandler
	ChainID                  []byte
}

// State struct holds the state components of the Elrond protocol
type State struct {
	AddressConverter    state.AddressConverter
	BLSAddressConverter state.AddressConverter
	PeerAccounts        state.AccountsAdapter
	AccountsAdapter     state.AccountsAdapter
	InBalanceForShard   map[string]*big.Int
}

// Data struct holds the data components of the Elrond protocol
type Data struct {
	Blkc     data.ChainHandler
	Store    dataRetriever.StorageService
	Datapool dataRetriever.PoolsHolder
}

// Crypto struct holds the crypto components of the Elrond protocol
type Crypto struct {
	TxSingleSigner  crypto.SingleSigner
	SingleSigner    crypto.SingleSigner
	MultiSigner     crypto.MultiSigner
	BlockSignKeyGen crypto.KeyGenerator
	TxSignKeyGen    crypto.KeyGenerator
	InitialPubKeys  map[uint32][]string
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
	hasher, err := getHasherFromConfig(args.config)
	if err != nil {
		return nil, errors.New("could not create hasher: " + err.Error())
	}

	marshalizer, err := getMarshalizerFromConfig(args.config)
	if err != nil {
		return nil, errors.New("could not create marshalizer: " + err.Error())
	}

	uint64ByteSliceConverter := uint64ByteSlice.NewBigEndianConverter()

	trieContainer, err := createTries(args, marshalizer, hasher)
	if err != nil {
		return nil, err
	}

	return &Core{
		Hasher:                   hasher,
		Marshalizer:              marshalizer,
		TriesContainer:           trieContainer,
		Uint64ByteSliceConverter: uint64ByteSliceConverter,
		StatusHandler:            statusHandler.NewNilStatusHandler(),
		ChainID:                  args.chainID,
	}, nil
}

func createTries(
	args *coreComponentsFactoryArgs,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
) (state.TriesHolder, error) {

	trieContainer := state.NewDataTriesHolder()

	trieFactoryArgs := factory.TrieFactoryArgs{
		EvictionWaitingListCfg: args.config.EvictionWaitingList,
		SnapshotDbCfg:          args.config.TrieSnapshotDB,
		Marshalizer:            marshalizer,
		Hasher:                 hasher,
		PathManager:            args.pathManager,
		ShardId:                args.shardId,
	}
	trieFactory, err := factory.NewTrieFactory(trieFactoryArgs)
	if err != nil {
		return nil, err
	}

	merkleTrie, err := trieFactory.Create(args.config.AccountsTrieStorage, args.config.StateTriesConfig.AccountsStatePruningEnabled)
	if err != nil {
		return nil, err
	}

	trieContainer.Put([]byte(factory.UserAccountTrie), merkleTrie)

	peerAccountsTrie, err := trieFactory.Create(args.config.PeerAccountsTrieStorage, args.config.StateTriesConfig.PeerStatePruningEnabled)
	if err != nil {
		return nil, err
	}

	trieContainer.Put([]byte(factory.PeerAccountTrie), peerAccountsTrie)

	return trieContainer, nil
}

type stateComponentsFactoryArgs struct {
	config           *config.Config
	genesisConfig    *sharding.Genesis
	shardCoordinator sharding.Coordinator
	core             *Core
	pathManager      storage.PathManagerHandler
}

// NewStateComponentsFactoryArgs initializes the arguments necessary for creating the state components
func NewStateComponentsFactoryArgs(
	config *config.Config,
	genesisConfig *sharding.Genesis,
	shardCoordinator sharding.Coordinator,
	core *Core,
	pathManager storage.PathManagerHandler,
) *stateComponentsFactoryArgs {
	return &stateComponentsFactoryArgs{
		config:           config,
		genesisConfig:    genesisConfig,
		shardCoordinator: shardCoordinator,
		core:             core,
		pathManager:      pathManager,
	}
}

// StateComponentsFactory creates the state components
func StateComponentsFactory(args *stateComponentsFactoryArgs) (*State, error) {
	addressConverter, err := addressConverters.NewPlainAddressConverter(
		args.config.Address.Length,
		args.config.Address.Prefix,
	)
	if err != nil {
		return nil, errors.New("could not create address converter: " + err.Error())
	}

	blsAddressConverter, err := addressConverters.NewPlainAddressConverter(
		args.config.BLSPublicKey.Length,
		args.config.BLSPublicKey.Prefix,
	)
	if err != nil {
		return nil, errors.New("could not create bls address converter: " + err.Error())
	}

	accountFactory, err := factoryState.NewAccountFactoryCreator(factoryState.UserAccount)
	if err != nil {
		return nil, errors.New("could not create account factory: " + err.Error())
	}

	merkleTrie := args.core.TriesContainer.Get([]byte(factory.UserAccountTrie))
	accountsAdapter, err := state.NewAccountsDB(merkleTrie, args.core.Hasher, args.core.Marshalizer, accountFactory)
	if err != nil {
		return nil, errors.New("could not create accounts adapter: " + err.Error())
	}

	inBalanceForShard, err := args.genesisConfig.InitialNodesBalances(args.shardCoordinator, addressConverter)
	if err != nil {
		return nil, errors.New("initial balances could not be processed " + err.Error())
	}

	accountFactory, err = factoryState.NewAccountFactoryCreator(factoryState.ValidatorAccount)
	if err != nil {
		return nil, errors.New("could not create peer account factory: " + err.Error())
	}

	merkleTrie = args.core.TriesContainer.Get([]byte(factory.PeerAccountTrie))
	peerAdapter, err := state.NewPeerAccountsDB(merkleTrie, args.core.Hasher, args.core.Marshalizer, accountFactory)
	if err != nil {
		return nil, err
	}

	return &State{
		PeerAccounts:        peerAdapter,
		AddressConverter:    addressConverter,
		BLSAddressConverter: blsAddressConverter,
		AccountsAdapter:     accountsAdapter,
		InBalanceForShard:   inBalanceForShard,
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
	blkc, err := createBlockChainFromConfig(args.config, args.shardCoordinator, args.core.StatusHandler)
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

	datapool, err = createDataPoolFromConfig(args)
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
	txSingleSigner := &singlesig.SchnorrSigner{}
	singleSigner, err := createSingleSigner(args.config)
	if err != nil {
		return nil, errors.New("could not create singleSigner: " + err.Error())
	}

	multisigHasher, err := getMultisigHasherFromConfig(args.config)
	if err != nil {
		return nil, errors.New("could not create multisig hasher: " + err.Error())
	}

	currentShardNodesPubKeys, err := args.nodesConfig.InitialNodesPubKeysForShard(args.shardCoordinator.SelfId())
	if err != nil {
		return nil, errors.New("could not start creation of multiSigner: " + err.Error())
	}

	multiSigner, err := createMultiSigner(args.config, multisigHasher, currentShardNodesPubKeys, args.privKey, args.keyGen)
	if err != nil {
		return nil, err
	}

	txSignKeyGen := signing.NewKeyGenerator(kyber.NewBlakeSHA256Ed25519())

	return &Crypto{
		TxSingleSigner:  txSingleSigner,
		SingleSigner:    singleSigner,
		MultiSigner:     multiSigner,
		BlockSignKeyGen: args.keyGen,
		TxSignKeyGen:    txSignKeyGen,
		InitialPubKeys:  initialPubKeys,
	}, nil
}

// NetworkComponentsFactory creates the network components
func NetworkComponentsFactory(p2pConfig *config.P2PConfig, log logger.Logger, core *Core) (*Network, error) {
	var randReader io.Reader
	if p2pConfig.Node.Seed != "" {
		randReader = NewSeedRandReader(core.Hasher.Compute(p2pConfig.Node.Seed))
	} else {
		randReader = rand.Reader
	}

	netMessenger, err := createNetMessenger(p2pConfig, log, randReader)
	if err != nil {
		return nil, err
	}

	return &Network{
		NetMessenger: netMessenger,
	}, nil
}

type processComponentsFactoryArgs struct {
	coreComponents         *coreComponentsFactoryArgs
	genesisConfig          *sharding.Genesis
	economicsData          *economics.EconomicsData
	nodesConfig            *sharding.NodesSetup
	gasSchedule            map[string]map[string]uint64
	syncer                 ntp.SyncTimer
	shardCoordinator       sharding.Coordinator
	nodesCoordinator       sharding.NodesCoordinator
	data                   *Data
	core                   *Core
	crypto                 *Crypto
	state                  *State
	network                *Network
	coreServiceContainer   serviceContainer.Core
	requestedItemsHandler  dataRetriever.RequestedItemsHandler
	epochStartNotifier     EpochStartNotifier
	epochStart             *config.EpochStartConfig
	rater                  sharding.RaterHandler
	startEpochNum          uint32
	sizeCheckDelta         uint32
	stateCheckpointModulus uint
	maxComputableRounds    uint64
}

// NewProcessComponentsFactoryArgs initializes the arguments necessary for creating the process components
func NewProcessComponentsFactoryArgs(
	coreComponents *coreComponentsFactoryArgs,
	genesisConfig *sharding.Genesis,
	economicsData *economics.EconomicsData,
	nodesConfig *sharding.NodesSetup,
	gasSchedule map[string]map[string]uint64,
	syncer ntp.SyncTimer,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	data *Data,
	core *Core,
	crypto *Crypto,
	state *State,
	network *Network,
	coreServiceContainer serviceContainer.Core,
	requestedItemsHandler dataRetriever.RequestedItemsHandler,
	epochStartNotifier EpochStartNotifier,
	epochStart *config.EpochStartConfig,
	startEpochNum uint32,
	rater sharding.RaterHandler,
	sizeCheckDelta uint32,
	stateCheckpointModulus uint,
	maxComputableRounds uint64,
) *processComponentsFactoryArgs {
	return &processComponentsFactoryArgs{
		coreComponents:         coreComponents,
		genesisConfig:          genesisConfig,
		economicsData:          economicsData,
		nodesConfig:            nodesConfig,
		gasSchedule:            gasSchedule,
		syncer:                 syncer,
		shardCoordinator:       shardCoordinator,
		nodesCoordinator:       nodesCoordinator,
		data:                   data,
		core:                   core,
		crypto:                 crypto,
		state:                  state,
		network:                network,
		coreServiceContainer:   coreServiceContainer,
		requestedItemsHandler:  requestedItemsHandler,
		epochStartNotifier:     epochStartNotifier,
		epochStart:             epochStart,
		startEpochNum:          startEpochNum,
		rater:                  rater,
		sizeCheckDelta:         sizeCheckDelta,
		stateCheckpointModulus: stateCheckpointModulus,
		maxComputableRounds:    maxComputableRounds,
	}
}

// ProcessComponentsFactory creates the process components
func ProcessComponentsFactory(args *processComponentsFactoryArgs) (*Process, error) {
	argsHeaderSig := &headerCheck.ArgsHeaderSigVerifier{
		Marshalizer:       args.core.Marshalizer,
		Hasher:            args.core.Hasher,
		NodesCoordinator:  args.nodesCoordinator,
		MultiSigVerifier:  args.crypto.MultiSigner,
		SingleSigVerifier: args.crypto.SingleSigner,
		KeyGen:            args.crypto.BlockSignKeyGen,
	}
	headerSigVerifier, err := headerCheck.NewHeaderSigVerifier(argsHeaderSig)
	if err != nil {
		return nil, err
	}

	rounder, err := round.NewRound(
		time.Unix(args.nodesConfig.StartTime, 0),
		args.syncer.CurrentTime(),
		time.Millisecond*time.Duration(args.nodesConfig.RoundDuration),
		args.syncer)
	if err != nil {
		return nil, err
	}

	resolversContainerFactory, err := newResolverContainerFactory(
		args.shardCoordinator,
		args.data,
		args.core,
		args.network,
		args.sizeCheckDelta,
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

	requestHandler, err := newRequestHandler(resolversFinder, args.shardCoordinator, args.requestedItemsHandler)
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

	validatorStatisticsProcessor, err := newValidatorStatisticsProcessor(args)
	if err != nil {
		return nil, err
	}

	validatorStatsRootHash, err := validatorStatisticsProcessor.RootHash()
	if err != nil {
		return nil, err
	}

	log.Trace("Validator stats created", "validatorStatsRootHash", validatorStatsRootHash)

	genesisBlocks, err := generateGenesisHeadersAndApplyInitialBalances(args)
	if err != nil {
		return nil, err
	}

	err = prepareGenesisBlock(args, genesisBlocks)
	if err != nil {
		return nil, err
	}

	bootStr := args.data.Store.GetStorer(dataRetriever.BootstrapUnit)
	bootStorer, err := bootstrapStorage.NewBootstrapStorer(args.core.Marshalizer, bootStr)
	if err != nil {
		return nil, err
	}

	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:      args.core.Hasher,
		Marshalizer: args.core.Marshalizer,
	}
	headerValidator, err := block.NewHeaderValidator(argsHeaderValidator)
	if err != nil {
		return nil, err
	}

	blockTracker, err := newBlockTracker(
		args,
		headerValidator,
		requestHandler,
		rounder,
		genesisBlocks,
	)
	if err != nil {
		return nil, err
	}

	interceptorContainerFactory, blackListHandler, err := newInterceptorContainerFactory(
		args.shardCoordinator,
		args.nodesCoordinator,
		args.data,
		args.core,
		args.crypto,
		args.state,
		args.network,
		args.economicsData,
		headerSigVerifier,
		args.sizeCheckDelta,
		blockTracker,
		epochStartTrigger,
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
	if args.shardCoordinator.SelfId() == sharding.MetachainShardId {
		pendingMiniBlocksHandler, err = newPendingMiniBlocks(
			args.data.Store,
			args.core.Marshalizer,
			args.data.Datapool,
		)
		if err != nil {
			return nil, err
		}
	}

	forkDetector, err := newForkDetector(
		rounder,
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
		rounder,
		epochStartTrigger,
		bootStorer,
		validatorStatisticsProcessor,
		headerValidator,
		blockTracker,
		pendingMiniBlocksHandler,
	)
	if err != nil {
		return nil, err
	}

	return &Process{
		InterceptorsContainer:    interceptorsContainer,
		ResolversFinder:          resolversFinder,
		Rounder:                  rounder,
		ForkDetector:             forkDetector,
		BlockProcessor:           blockProcessor,
		EpochStartTrigger:        epochStartTrigger,
		BlackListHandler:         blackListHandler,
		BootStorer:               bootStorer,
		HeaderSigVerifier:        headerSigVerifier,
		ValidatorsStatistics:     validatorStatisticsProcessor,
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

	genesisBlockHash, err := core.CalculateHash(args.core.Marshalizer, args.core.Hasher, genesisBlock)
	if err != nil {
		return err
	}

	err = args.data.Blkc.SetGenesisHeader(genesisBlock)
	if err != nil {
		return err
	}

	args.data.Blkc.SetGenesisHeaderHash(genesisBlockHash)

	marshalizedBlock, err := args.core.Marshalizer.Marshal(genesisBlock)
	if err != nil {
		return err
	}

	if args.shardCoordinator.SelfId() == sharding.MetachainShardId {
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

func newRequestHandler(
	resolversFinder dataRetriever.ResolversFinder,
	shardCoordinator sharding.Coordinator,
	requestedItemsHandler dataRetriever.RequestedItemsHandler,
) (process.RequestHandler, error) {
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		requestHandler, err := requestHandlers.NewShardResolverRequestHandler(
			resolversFinder,
			requestedItemsHandler,
			MaxTxsToRequest,
			shardCoordinator.SelfId(),
		)
		if err != nil {
			return nil, err
		}

		return requestHandler, nil
	}

	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		requestHandler, err := requestHandlers.NewMetaResolverRequestHandler(
			resolversFinder,
			requestedItemsHandler,
			MaxTxsToRequest,
		)
		if err != nil {
			return nil, err
		}

		return requestHandler, nil
	}

	return nil, errors.New("could not create new request handler because of wrong shard id")
}

func newEpochStartTrigger(
	args *processComponentsFactoryArgs,
	requestHandler epochStart.RequestHandler,
) (epochStart.TriggerHandler, error) {
	if args.shardCoordinator.SelfId() < args.shardCoordinator.NumberOfShards() {
		argsHeaderValidator := block.ArgsHeaderValidator{
			Hasher:      args.core.Hasher,
			Marshalizer: args.core.Marshalizer,
		}
		headerValidator, err := block.NewHeaderValidator(argsHeaderValidator)
		if err != nil {
			return nil, err
		}

		argEpochStart := &shardchain.ArgsShardEpochStartTrigger{
			Marshalizer:        args.core.Marshalizer,
			Hasher:             args.core.Hasher,
			HeaderValidator:    headerValidator,
			Uint64Converter:    args.core.Uint64ByteSliceConverter,
			DataPool:           args.data.Datapool,
			Storage:            args.data.Store,
			RequestHandler:     requestHandler,
			Epoch:              args.startEpochNum,
			EpochStartNotifier: args.epochStartNotifier,
			Validity:           process.MetaBlockValidity,
			Finality:           process.BlockFinality,
		}
		epochStartTrigger, err := shardchain.NewEpochStartTrigger(argEpochStart)
		if err != nil {
			return nil, errors.New("error creating new start of epoch trigger" + err.Error())
		}

		return epochStartTrigger, nil
	}

	if args.shardCoordinator.SelfId() == sharding.MetachainShardId {
		argEpochStart := &metachainEpochStart.ArgsNewMetaEpochStartTrigger{
			GenesisTime:        time.Unix(args.nodesConfig.StartTime, 0),
			Settings:           args.epochStart,
			Epoch:              args.startEpochNum,
			EpochStartNotifier: args.epochStartNotifier,
			Storage:            args.data.Store,
			Marshalizer:        args.core.Marshalizer,
		}
		epochStartTrigger, err := metachainEpochStart.NewEpochStartTrigger(argEpochStart)
		if err != nil {
			return nil, errors.New("error creating new start of epoch trigger" + err.Error())
		}

		return epochStartTrigger, nil
	}

	return nil, errors.New("error creating new start of epoch trigger because of invalid shard id")
}

type seedRandReader struct {
	index int
	seed  []byte
}

// NewSeedRandReader will return a new instance of a seed-based reader
func NewSeedRandReader(seed []byte) *seedRandReader {
	return &seedRandReader{seed: seed, index: 0}
}

func (srr *seedRandReader) Read(p []byte) (n int, err error) {
	if srr.seed == nil {
		return 0, errors.New("nil seed")
	}
	if len(srr.seed) == 0 {
		return 0, errors.New("empty seed")
	}
	if p == nil {
		return 0, errors.New("nil buffer")
	}
	if len(p) == 0 {
		return 0, errors.New("empty buffer")
	}

	for i := 0; i < len(p); i++ {
		p[i] = srr.seed[srr.index]

		srr.index++
		srr.index %= len(srr.seed)
	}

	return len(p), nil
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

func getHasherFromConfig(cfg *config.Config) (hashing.Hasher, error) {
	switch cfg.Hasher.Type {
	case "sha256":
		return sha256.Sha256{}, nil
	case "blake2b":
		return &blake2b.Blake2b{}, nil
	}

	return nil, errors.New("no hasher provided in config file")
}

func getMarshalizerFromConfig(cfg *config.Config) (marshal.Marshalizer, error) {
	switch cfg.Marshalizer.Type {
	case "json":
		return &marshal.JsonMarshalizer{}, nil
	}

	return nil, errors.New("no marshalizer provided in config file")
}

func createBlockChainFromConfig(config *config.Config, coordinator sharding.Coordinator, ash core.AppStatusHandler) (data.ChainHandler, error) {
	badBlockCache, err := storageUnit.NewCache(
		storageUnit.CacheType(config.BadBlocksCache.Type),
		config.BadBlocksCache.Size,
		config.BadBlocksCache.Shards)
	if err != nil {
		return nil, err
	}

	if coordinator == nil {
		return nil, state.ErrNilShardCoordinator
	}

	if coordinator.SelfId() < coordinator.NumberOfShards() {
		var blockChain *blockchain.BlockChain
		blockChain, err = blockchain.NewBlockChain(badBlockCache)
		if err != nil {
			return nil, err
		}

		err = blockChain.SetAppStatusHandler(ash)
		if err != nil {
			return nil, err
		}

		return blockChain, nil
	}
	if coordinator.SelfId() == sharding.MetachainShardId {
		var blockChain *blockchain.MetaChain
		blockChain, err = blockchain.NewMetaChain(badBlockCache)
		if err != nil {
			return nil, err
		}

		err = blockChain.SetAppStatusHandler(ash)
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
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		return storageServiceFactory.CreateForMeta()
	}
	return nil, errors.New("can not create data store")
}

func createDataPoolFromConfig(args *dataComponentsFactoryArgs) (dataRetriever.PoolsHolder, error) {
	log.Debug("creatingDataPool from config")

	configs := args.config

	txPool, err := txpoolFactory.CreateTxPool(txpool.ArgShardedTxPool{
		Config:         storageFactory.GetCacherFromConfig(configs.TxDataPool),
		MinGasPrice:    args.economicsData.MinGasPrice(),
		NumberOfShards: args.shardCoordinator.NumberOfShards(),
	})
	if err != nil {
		log.Error("error creating txpool")
		return nil, err
	}

	uTxPool, err := shardedData.NewShardedData(storageFactory.GetCacherFromConfig(configs.UnsignedTransactionDataPool))
	if err != nil {
		log.Error("error creating smart contract result pool")
		return nil, err
	}

	rewardTxPool, err := shardedData.NewShardedData(storageFactory.GetCacherFromConfig(configs.RewardTransactionDataPool))
	if err != nil {
		log.Error("error creating reward transaction pool")
		return nil, err
	}

	hdrPool, err := headersCache.NewHeadersPool(configs.HeadersPoolConfig)
	if err != nil {
		log.Error("error creating headers pool")
		return nil, err
	}

	cacherCfg := storageFactory.GetCacherFromConfig(configs.TxBlockBodyDataPool)
	txBlockBody, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		log.Error("error creating txBlockBody")
		return nil, err
	}

	cacherCfg = storageFactory.GetCacherFromConfig(configs.PeerBlockBodyDataPool)
	peerChangeBlockBody, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		log.Error("error creating peerChangeBlockBody")
		return nil, err
	}

	cacherCfg = storageFactory.GetCacherFromConfig(configs.TrieNodesDataPool)
	trieNodes, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		log.Info("error creating trieNodes")
		return nil, err
	}

	currBlockTxs, err := dataPool.NewCurrentBlockPool()
	if err != nil {
		return nil, err
	}

	return dataPool.NewDataPool(
		txPool,
		uTxPool,
		rewardTxPool,
		hdrPool,
		txBlockBody,
		peerChangeBlockBody,
		trieNodes,
		currBlockTxs,
	)
}

func createSingleSigner(config *config.Config) (crypto.SingleSigner, error) {
	switch config.Consensus.Type {
	case BlsConsensusType:
		return &singlesig.BlsSingleSigner{}, nil
	default:
		return nil, errors.New("no consensus type provided in config file")
	}
}

func getMultisigHasherFromConfig(cfg *config.Config) (hashing.Hasher, error) {
	if cfg.Consensus.Type == BlsConsensusType && cfg.MultisigHasher.Type != "blake2b" {
		return nil, errors.New("wrong multisig hasher provided for bls consensus type")
	}

	switch cfg.MultisigHasher.Type {
	case "sha256":
		return sha256.Sha256{}, nil
	case "blake2b":
		if cfg.Consensus.Type == BlsConsensusType {
			return &blake2b.Blake2b{HashSize: BlsHashSize}, nil
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
	case BlsConsensusType:
		blsSigner := &blsMultiSig.KyberMultiSignerBLS{}
		return multisig.NewBLSMultisig(blsSigner, hasher, pubKeys, privateKey, keyGen, uint16(0))
	default:
		return nil, errors.New("no consensus type provided in config file")
	}
}

func createNetMessenger(
	p2pConfig *config.P2PConfig,
	log logger.Logger,
	randReader io.Reader,
) (p2p.Messenger, error) {

	if p2pConfig.Node.Port < 0 {
		return nil, errors.New("cannot start node on port < 0")
	}

	pDiscoveryFactory := factoryP2P.NewPeerDiscovererFactory(*p2pConfig)
	pDiscoverer, err := pDiscoveryFactory.CreatePeerDiscoverer()

	if err != nil {
		return nil, err
	}

	log.Debug("peer discovery", "method", pDiscoverer.Name())

	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), randReader)
	sk := (*libp2pCrypto.Secp256k1PrivateKey)(prvKey)

	nm, err := libp2p.NewNetworkMessenger(
		context.Background(),
		p2pConfig.Node.Port,
		sk,
		nil,
		loadBalancer.NewOutgoingChannelLoadBalancer(),
		pDiscoverer,
		libp2p.ListenAddrWithIp4AndTcp,
		p2pConfig.Node.TargetPeerCount,
	)
	if err != nil {
		return nil, err
	}

	return nm, nil
}

func newInterceptorContainerFactory(
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	data *Data,
	core *Core,
	crypto *Crypto,
	state *State,
	network *Network,
	economics *economics.EconomicsData,
	headerSigVerifier HeaderSigVerifierHandler,
	sizeCheckDelta uint32,
	validityAttester process.ValidityAttester,
	epochStartTrigger process.EpochStartTriggerHandler,
) (process.InterceptorsContainerFactory, process.BlackListHandler, error) {

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return newShardInterceptorContainerFactory(
			shardCoordinator,
			nodesCoordinator,
			data,
			core,
			crypto,
			state,
			network,
			economics,
			headerSigVerifier,
			sizeCheckDelta,
			validityAttester,
			epochStartTrigger,
		)
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		return newMetaInterceptorContainerFactory(
			shardCoordinator,
			nodesCoordinator,
			data,
			core,
			crypto,
			network,
			state,
			economics,
			headerSigVerifier,
			sizeCheckDelta,
			validityAttester,
			epochStartTrigger,
		)
	}

	return nil, nil, errors.New("could not create interceptor container factory")
}

func newResolverContainerFactory(
	shardCoordinator sharding.Coordinator,
	data *Data,
	core *Core,
	network *Network,
	sizeCheckDelta uint32,
) (dataRetriever.ResolversContainerFactory, error) {

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return newShardResolverContainerFactory(
			shardCoordinator,
			data,
			core,
			network,
			sizeCheckDelta,
		)
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		return newMetaResolverContainerFactory(
			shardCoordinator,
			data,
			core,
			network,
			sizeCheckDelta,
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
) (process.InterceptorsContainerFactory, process.BlackListHandler, error) {

	headerBlackList := timecache.NewTimeCache(timeSpanForBadHeaders)
	shardInterceptorsContainerFactoryArgs := interceptorscontainer.ShardInterceptorsContainerFactoryArgs{
		Accounts:               state.AccountsAdapter,
		ShardCoordinator:       shardCoordinator,
		NodesCoordinator:       nodesCoordinator,
		Messenger:              network.NetMessenger,
		Store:                  data.Store,
		Marshalizer:            dataCore.Marshalizer,
		Hasher:                 dataCore.Hasher,
		KeyGen:                 crypto.TxSignKeyGen,
		BlockSignKeyGen:        crypto.BlockSignKeyGen,
		SingleSigner:           crypto.TxSingleSigner,
		BlockSingleSigner:      crypto.SingleSigner,
		MultiSigner:            crypto.MultiSigner,
		DataPool:               data.Datapool,
		AddrConverter:          state.AddressConverter,
		MaxTxNonceDeltaAllowed: core.MaxTxNonceDeltaAllowed,
		TxFeeHandler:           economics,
		BlackList:              headerBlackList,
		HeaderSigVerifier:      headerSigVerifier,
		ChainID:                dataCore.ChainID,
		SizeCheckDelta:         sizeCheckDelta,
		ValidityAttester:       validityAttester,
		EpochStartTrigger:      epochStartTrigger,
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
) (process.InterceptorsContainerFactory, process.BlackListHandler, error) {
	headerBlackList := timecache.NewTimeCache(timeSpanForBadHeaders)
	metaInterceptorsContainerFactoryArgs := interceptorscontainer.MetaInterceptorsContainerFactoryArgs{
		ShardCoordinator:       shardCoordinator,
		NodesCoordinator:       nodesCoordinator,
		Messenger:              network.NetMessenger,
		Store:                  data.Store,
		Marshalizer:            dataCore.Marshalizer,
		Hasher:                 dataCore.Hasher,
		MultiSigner:            crypto.MultiSigner,
		DataPool:               data.Datapool,
		Accounts:               state.AccountsAdapter,
		AddrConverter:          state.AddressConverter,
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
) (dataRetriever.ResolversContainerFactory, error) {

	dataPacker, err := partitioning.NewSimpleDataPacker(core.Marshalizer)
	if err != nil {
		return nil, err
	}

	resolversContainerFactoryArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:         shardCoordinator,
		Messenger:                network.NetMessenger,
		Store:                    data.Store,
		Marshalizer:              core.Marshalizer,
		DataPools:                data.Datapool,
		Uint64ByteSliceConverter: core.Uint64ByteSliceConverter,
		DataPacker:               dataPacker,
		TriesContainer:           core.TriesContainer,
		SizeCheckDelta:           sizeCheckDelta,
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
) (dataRetriever.ResolversContainerFactory, error) {
	dataPacker, err := partitioning.NewSimpleDataPacker(core.Marshalizer)
	if err != nil {
		return nil, err
	}

	resolversContainerFactoryArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:         shardCoordinator,
		Messenger:                network.NetMessenger,
		Store:                    data.Store,
		Marshalizer:              core.Marshalizer,
		DataPools:                data.Datapool,
		Uint64ByteSliceConverter: core.Uint64ByteSliceConverter,
		DataPacker:               dataPacker,
		TriesContainer:           core.TriesContainer,
		SizeCheckDelta:           sizeCheckDelta,
	}
	resolversContainerFactory, err := resolverscontainer.NewMetaResolversContainerFactory(resolversContainerFactoryArgs)
	if err != nil {
		return nil, err
	}
	return resolversContainerFactory, nil
}

func generateGenesisHeadersAndApplyInitialBalances(args *processComponentsFactoryArgs) (map[uint32]data.HeaderHandler, error) {
	//TODO change this rudimentary startup for metachain nodes
	// Talk between Adrian, Robert and Iulian, did not want it to be discarded:
	// --------------------------------------------------------------------
	// Adrian: "This looks like a workaround as the metchain should not deal with individual accounts, but shards data.
	// What I was thinking was that the genesis on metachain (or pre-genesis block) is the nodes allocation to shards,
	// with 0 state root for every shard, as there is no balance yet.
	// Then the shards start operating as they get the initial node allocation, maybe we can do consensus on the
	// genesis as well, I think this would be actually good as then everything is signed and agreed upon.
	// The genesis shard blocks need to be then just the state root, I think we already have that in genesis,
	// so shard nodes can go ahead with individually creating the block, but then run consensus on this.
	// Then this block is sent to metachain who updates the state root of every shard and creates the metablock for
	// the genesis of each of the shards (this is actually the same thing that would happen at new epoch start)."

	coreComponents := args.core
	stateComponents := args.state
	dataComponents := args.data
	shardCoordinator := args.shardCoordinator
	nodesSetup := args.nodesConfig
	genesisConfig := args.genesisConfig
	economicsData := args.economicsData

	genesisBlocks := make(map[uint32]data.HeaderHandler)

	validatorStatsRootHash, err := stateComponents.PeerAccounts.RootHash()
	if err != nil {
		return nil, err
	}

	for shardId := uint32(0); shardId < shardCoordinator.NumberOfShards(); shardId++ {
		isCurrentShard := shardId == shardCoordinator.SelfId()
		if isCurrentShard {
			continue
		}

		var newShardCoordinator sharding.Coordinator
		var accountsAdapter state.AccountsAdapter
		newShardCoordinator, accountsAdapter, err = createInMemoryShardCoordinatorAndAccount(
			coreComponents,
			shardCoordinator.NumberOfShards(),
			shardId,
		)
		if err != nil {
			return nil, err
		}

		var genesisBlock data.HeaderHandler
		genesisBlock, err = createGenesisBlockAndApplyInitialBalances(
			accountsAdapter,
			newShardCoordinator,
			stateComponents.AddressConverter,
			genesisConfig,
			uint64(nodesSetup.StartTime),
			validatorStatsRootHash,
		)
		if err != nil {
			return nil, err
		}

		genesisBlocks[shardId] = genesisBlock
	}

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		var genesisBlockForCurrentShard data.HeaderHandler
		genesisBlockForCurrentShard, err = createGenesisBlockAndApplyInitialBalances(
			stateComponents.AccountsAdapter,
			shardCoordinator,
			stateComponents.AddressConverter,
			genesisConfig,
			uint64(nodesSetup.StartTime),
			validatorStatsRootHash,
		)
		if err != nil {
			return nil, err
		}

		genesisBlocks[shardCoordinator.SelfId()] = genesisBlockForCurrentShard
	}

	argsMetaGenesis := genesis.ArgsMetaGenesisBlockCreator{
		GenesisTime:              uint64(nodesSetup.StartTime),
		Accounts:                 stateComponents.AccountsAdapter,
		AddrConv:                 stateComponents.AddressConverter,
		NodesSetup:               nodesSetup,
		ShardCoordinator:         shardCoordinator,
		Store:                    dataComponents.Store,
		Blkc:                     dataComponents.Blkc,
		Marshalizer:              coreComponents.Marshalizer,
		Hasher:                   coreComponents.Hasher,
		Uint64ByteSliceConverter: coreComponents.Uint64ByteSliceConverter,
		DataPool:                 dataComponents.Datapool,
		Economics:                economicsData,
		ValidatorStatsRootHash:   validatorStatsRootHash,
	}

	if shardCoordinator.SelfId() != sharding.MetachainShardId {
		var newShardCoordinator sharding.Coordinator
		var newAccounts state.AccountsAdapter
		newShardCoordinator, newAccounts, err = createInMemoryShardCoordinatorAndAccount(
			coreComponents,
			shardCoordinator.NumberOfShards(),
			sharding.MetachainShardId,
		)
		if err != nil {
			return nil, err
		}

		newStore, newBlkc, errPoolCreation := createInMemoryStoreBlkc(newShardCoordinator)
		if errPoolCreation != nil {
			return nil, errPoolCreation
		}

		argsMetaGenesis.ShardCoordinator = newShardCoordinator
		argsMetaGenesis.Accounts = newAccounts
		argsMetaGenesis.Store = newStore
		argsMetaGenesis.Blkc = newBlkc
	}

	genesisBlock, err := genesis.CreateMetaGenesisBlock(
		argsMetaGenesis,
	)
	if err != nil {
		return nil, err
	}

	log.Debug("MetaGenesisBlock created",
		"roothash", genesisBlock.GetRootHash(),
		"validatorStatsRootHash", genesisBlock.GetValidatorStatsRootHash(),
	)

	genesisBlocks[sharding.MetachainShardId] = genesisBlock

	return genesisBlocks, nil
}

func createInMemoryStoreBlkc(
	shardCoordinator sharding.Coordinator,
) (dataRetriever.StorageService, data.ChainHandler, error) {

	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, 10, 1)
	blkc, err := blockchain.NewMetaChain(cache)
	if err != nil {
		return nil, nil, err
	}

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, createMemUnit())
	store.AddStorer(dataRetriever.TransactionUnit, createMemUnit())
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, createMemUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, createMemUnit())
	for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(i)
		store.AddStorer(hdrNonceHashDataUnit, createMemUnit())
	}
	store.AddStorer(dataRetriever.HeartbeatUnit, createMemUnit())

	return store, blkc, nil
}

func createGenesisBlockAndApplyInitialBalances(
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	addressConverter state.AddressConverter,
	genesisConfig *sharding.Genesis,
	startTime uint64,
	validatorStatsRootHash []byte,
) (data.HeaderHandler, error) {

	initialBalances, err := genesisConfig.InitialNodesBalances(shardCoordinator, addressConverter)
	if err != nil {
		return nil, err
	}

	return genesis.CreateShardGenesisBlockFromInitialBalances(
		accounts,
		shardCoordinator,
		addressConverter,
		initialBalances,
		startTime,
		validatorStatsRootHash,
	)
}

func createInMemoryShardCoordinatorAndAccount(
	coreComponents *Core,
	numOfShards uint32,
	shardId uint32,
) (sharding.Coordinator, state.AccountsAdapter, error) {

	newShardCoordinator, err := sharding.NewMultiShardCoordinator(numOfShards, shardId)
	if err != nil {
		return nil, nil, err
	}

	accountFactory, err := factoryState.NewAccountFactoryCreator(factoryState.UserAccount)
	if err != nil {
		return nil, nil, err
	}

	accounts, err := generateInMemoryAccountsAdapter(
		accountFactory,
		coreComponents.Hasher,
		coreComponents.Marshalizer,
	)
	if err != nil {
		return nil, nil, err
	}

	return newShardCoordinator, accounts, nil
}

func newBlockTracker(
	processArgs *processComponentsFactoryArgs,
	headerValidator process.HeaderConstructionValidator,
	requestHandler process.RequestHandler,
	rounder consensus.Rounder,
	genesisBlocks map[uint32]data.HeaderHandler,
) (process.BlockTracker, error) {

	argBaseTracker := track.ArgBaseTracker{
		Hasher:           processArgs.core.Hasher,
		HeaderValidator:  headerValidator,
		Marshalizer:      processArgs.core.Marshalizer,
		RequestHandler:   requestHandler,
		Rounder:          rounder,
		ShardCoordinator: processArgs.shardCoordinator,
		Store:            processArgs.data.Store,
		StartHeaders:     genesisBlocks,
	}

	if processArgs.shardCoordinator.SelfId() < processArgs.shardCoordinator.NumberOfShards() {
		arguments := track.ArgShardTracker{
			ArgBaseTracker: argBaseTracker,
			PoolsHolder:    processArgs.data.Datapool,
		}

		return track.NewShardBlockTrack(arguments)
	}

	if processArgs.shardCoordinator.SelfId() == sharding.MetachainShardId {
		arguments := track.ArgMetaTracker{
			ArgBaseTracker: argBaseTracker,
			PoolsHolder:    processArgs.data.Datapool,
		}

		return track.NewMetaBlockTrack(arguments)
	}

	return nil, errors.New("could not create block tracker")
}

func newPendingMiniBlocks(
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	dataPool dataRetriever.PoolsHolder,
) (process.PendingMiniBlocksHandler, error) {

	miniBlockHeaderStore := store.GetStorer(dataRetriever.MiniBlockHeaderUnit)
	if check.IfNil(miniBlockHeaderStore) {
		return nil, errors.New("could not create pending miniblocks handler because of empty miniblock header store")
	}

	metaBlocksStore := store.GetStorer(dataRetriever.MetaBlockUnit)
	if check.IfNil(metaBlocksStore) {
		return nil, errors.New("could not create pending miniblocks handler because of empty metablock store")
	}

	argsPendingMiniBlocks := &metachainEpochStart.ArgsPendingMiniBlocks{
		Marshalizer:      marshalizer,
		Storage:          miniBlockHeaderStore,
		MetaBlockPool:    dataPool.Headers(),
		MetaBlockStorage: metaBlocksStore,
	}
	pendingMiniBlocks, err := metachainEpochStart.NewPendingMiniBlocks(argsPendingMiniBlocks)
	if err != nil {
		return nil, err
	}

	return pendingMiniBlocks, nil
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
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		return processSync.NewMetaForkDetector(rounder, headerBlackList, blockTracker, genesisTime)
	}

	return nil, ErrCreateForkDetector
}

func newBlockProcessor(
	processArgs *processComponentsFactoryArgs,
	requestHandler process.RequestHandler,
	forkDetector process.ForkDetector,
	rounder consensus.Rounder,
	epochStartTrigger epochStart.TriggerHandler,
	bootStorer process.BootStorer,
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor,
	headerValidator process.HeaderConstructionValidator,
	blockTracker process.BlockTracker,
	pendingMiniBlocksHandler process.PendingMiniBlocksHandler,
) (process.BlockProcessor, error) {

	shardCoordinator := processArgs.shardCoordinator
	nodesCoordinator := processArgs.nodesCoordinator

	communityAddr := processArgs.economicsData.CommunityAddress()
	burnAddr := processArgs.economicsData.BurnAddress()
	if communityAddr == "" || burnAddr == "" {
		return nil, errors.New("rewards configuration missing")
	}

	communityAddress, err := hex.DecodeString(communityAddr)
	if err != nil {
		return nil, err
	}

	burnAddress, err := hex.DecodeString(burnAddr)
	if err != nil {
		return nil, err
	}

	specialAddressHolder, err := address.NewSpecialAddressHolder(
		communityAddress,
		burnAddress,
		processArgs.state.AddressConverter,
		shardCoordinator,
		nodesCoordinator,
	)
	if err != nil {
		return nil, err
	}

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return newShardBlockProcessor(
			requestHandler,
			processArgs.shardCoordinator,
			processArgs.nodesCoordinator,
			specialAddressHolder,
			processArgs.data,
			processArgs.core,
			processArgs.state,
			forkDetector,
			processArgs.coreServiceContainer,
			processArgs.economicsData,
			rounder,
			epochStartTrigger,
			bootStorer,
			processArgs.gasSchedule,
			processArgs.stateCheckpointModulus,
			headerValidator,
			blockTracker,
		)
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		return newMetaBlockProcessor(
			requestHandler,
			processArgs.shardCoordinator,
			processArgs.nodesCoordinator,
			specialAddressHolder,
			processArgs.data,
			processArgs.core,
			processArgs.state,
			forkDetector,
			processArgs.coreServiceContainer,
			processArgs.economicsData,
			validatorStatisticsProcessor,
			rounder,
			epochStartTrigger,
			bootStorer,
			headerValidator,
			blockTracker,
			pendingMiniBlocksHandler,
		)
	}

	return nil, errors.New("could not create block processor")
}

func newShardBlockProcessor(
	requestHandler process.RequestHandler,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	specialAddressHandler process.SpecialAddressHandler,
	data *Data,
	core *Core,
	state *State,
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
) (process.BlockProcessor, error) {
	argsParser, err := vmcommon.NewAtArgumentParser()
	if err != nil {
		return nil, err
	}

	argsHook := hooks.ArgBlockChainHook{
		Accounts:         state.AccountsAdapter,
		AddrConv:         state.AddressConverter,
		StorageService:   data.Store,
		BlockChain:       data.Blkc,
		ShardCoordinator: shardCoordinator,
		Marshalizer:      core.Marshalizer,
		Uint64Converter:  core.Uint64ByteSliceConverter,
	}
	vmFactory, err := shard.NewVMContainerFactory(economics.MaxGasLimitPerBlock(), gasSchedule, argsHook)
	if err != nil {
		return nil, err
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, err
	}

	interimProcFactory, err := shard.NewIntermediateProcessorsContainerFactory(
		shardCoordinator,
		core.Marshalizer,
		core.Hasher,
		state.AddressConverter,
		specialAddressHandler,
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

	rewardsTxInterim, err := interimProcContainer.Get(dataBlock.RewardsBlock)
	if err != nil {
		return nil, err
	}

	rewardsTxHandler, ok := rewardsTxInterim.(process.TransactionFeeHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	internalTransactionProducer, ok := rewardsTxInterim.(process.InternalTransactionProducer)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	receiptTxInterim, err := interimProcContainer.Get(dataBlock.ReceiptBlock)
	if err != nil {
		return nil, err
	}

	badTxInterim, err := interimProcContainer.Get(dataBlock.InvalidBlock)
	if err != nil {
		return nil, err
	}

	txTypeHandler, err := coordinator.NewTxTypeHandler(state.AddressConverter, shardCoordinator, state.AccountsAdapter)
	if err != nil {
		return nil, err
	}

	gasHandler, err := preprocess.NewGasComputation(economics)
	if err != nil {
		return nil, err
	}

	scProcessor, err := smartContract.NewSmartContractProcessor(
		vmContainer,
		argsParser,
		core.Hasher,
		core.Marshalizer,
		state.AccountsAdapter,
		vmFactory.BlockChainHookImpl(),
		state.AddressConverter,
		shardCoordinator,
		scForwarder,
		rewardsTxHandler,
		economics,
		txTypeHandler,
		gasHandler,
	)
	if err != nil {
		return nil, err
	}

	rewardsTxProcessor, err := rewardTransaction.NewRewardTxProcessor(
		state.AccountsAdapter,
		state.AddressConverter,
		shardCoordinator,
		rewardsTxInterim,
	)
	if err != nil {
		return nil, err
	}

	transactionProcessor, err := transaction.NewTxProcessor(
		state.AccountsAdapter,
		core.Hasher,
		state.AddressConverter,
		core.Marshalizer,
		shardCoordinator,
		scProcessor,
		rewardsTxHandler,
		txTypeHandler,
		economics,
		receiptTxInterim,
		badTxInterim,
	)
	if err != nil {
		return nil, errors.New("could not create transaction statisticsProcessor: " + err.Error())
	}

	miniBlocksCompacter, err := preprocess.NewMiniBlocksCompaction(economics, shardCoordinator, gasHandler)
	if err != nil {
		return nil, err
	}

	preProcFactory, err := shard.NewPreProcessorsContainerFactory(
		shardCoordinator,
		data.Store,
		core.Marshalizer,
		core.Hasher,
		data.Datapool,
		state.AddressConverter,
		state.AccountsAdapter,
		requestHandler,
		transactionProcessor,
		scProcessor,
		scProcessor,
		rewardsTxProcessor,
		internalTransactionProducer,
		economics,
		miniBlocksCompacter,
		gasHandler,
		blockTracker,
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
		core.Marshalizer,
		shardCoordinator,
		state.AccountsAdapter,
		data.Datapool.MiniBlocks(),
		requestHandler,
		preProcContainer,
		interimProcContainer,
		gasHandler,
	)
	if err != nil {
		return nil, err
	}

	txPoolsCleaner, err := poolsCleaner.NewTxsPoolsCleaner(
		state.AccountsAdapter,
		shardCoordinator,
		data.Datapool,
		state.AddressConverter,
		economics,
	)
	if err != nil {
		return nil, err
	}

	argumentsBaseProcessor := block.ArgBaseProcessor{
		Accounts:              state.AccountsAdapter,
		ForkDetector:          forkDetector,
		Hasher:                core.Hasher,
		Marshalizer:           core.Marshalizer,
		Store:                 data.Store,
		ShardCoordinator:      shardCoordinator,
		NodesCoordinator:      nodesCoordinator,
		SpecialAddressHandler: specialAddressHandler,
		Uint64Converter:       core.Uint64ByteSliceConverter,
		RequestHandler:        requestHandler,
		Core:                  coreServiceContainer,
		BlockChainHook:        vmFactory.BlockChainHookImpl(),
		TxCoordinator:         txCoordinator,
		Rounder:               rounder,
		EpochStartTrigger:     epochStartTrigger,
		HeaderValidator:       headerValidator,
		BootStorer:            bootStorer,
		BlockTracker:          blockTracker,
		DataPool:              data.Datapool,
	}
	arguments := block.ArgShardProcessor{
		ArgBaseProcessor:       argumentsBaseProcessor,
		TxsPoolsCleaner:        txPoolsCleaner,
		StateCheckpointModulus: stateCheckpointModulus,
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
	specialAddressHandler process.SpecialAddressHandler,
	data *Data,
	core *Core,
	state *State,
	forkDetector process.ForkDetector,
	coreServiceContainer serviceContainer.Core,
	economics *economics.EconomicsData,
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor,
	rounder consensus.Rounder,
	epochStartTrigger epochStart.TriggerHandler,
	bootStorer process.BootStorer,
	headerValidator process.HeaderConstructionValidator,
	blockTracker process.BlockTracker,
	pendingMiniBlocksHandler process.PendingMiniBlocksHandler,
) (process.BlockProcessor, error) {

	argsHook := hooks.ArgBlockChainHook{
		Accounts:         state.AccountsAdapter,
		AddrConv:         state.AddressConverter,
		StorageService:   data.Store,
		BlockChain:       data.Blkc,
		ShardCoordinator: shardCoordinator,
		Marshalizer:      core.Marshalizer,
		Uint64Converter:  core.Uint64ByteSliceConverter,
	}
	vmFactory, err := metachain.NewVMContainerFactory(argsHook, economics)
	if err != nil {
		return nil, err
	}

	argsParser, err := vmcommon.NewAtArgumentParser()
	if err != nil {
		return nil, err
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, err
	}

	interimProcFactory, err := metachain.NewIntermediateProcessorsContainerFactory(
		shardCoordinator,
		core.Marshalizer,
		core.Hasher,
		state.AddressConverter,
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

	txTypeHandler, err := coordinator.NewTxTypeHandler(state.AddressConverter, shardCoordinator, state.AccountsAdapter)
	if err != nil {
		return nil, err
	}

	gasHandler, err := preprocess.NewGasComputation(economics)
	if err != nil {
		return nil, err
	}

	scProcessor, err := smartContract.NewSmartContractProcessor(
		vmContainer,
		argsParser,
		core.Hasher,
		core.Marshalizer,
		state.AccountsAdapter,
		vmFactory.BlockChainHookImpl(),
		state.AddressConverter,
		shardCoordinator,
		scForwarder,
		&metachain.TransactionFeeHandler{},
		economics,
		txTypeHandler,
		gasHandler,
	)
	if err != nil {
		return nil, err
	}

	transactionProcessor, err := transaction.NewMetaTxProcessor(
		state.AccountsAdapter,
		state.AddressConverter,
		shardCoordinator,
		scProcessor,
		txTypeHandler,
		economics,
	)
	if err != nil {
		return nil, errors.New("could not create transaction processor: " + err.Error())
	}

	miniBlocksCompacter, err := preprocess.NewMiniBlocksCompaction(economics, shardCoordinator, gasHandler)
	if err != nil {
		return nil, err
	}

	preProcFactory, err := metachain.NewPreProcessorsContainerFactory(
		shardCoordinator,
		data.Store,
		core.Marshalizer,
		core.Hasher,
		data.Datapool,
		state.AccountsAdapter,
		requestHandler,
		transactionProcessor,
		scProcessor,
		economics,
		miniBlocksCompacter,
		gasHandler,
		blockTracker,
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
		core.Marshalizer,
		shardCoordinator,
		state.AccountsAdapter,
		data.Datapool.MiniBlocks(),
		requestHandler,
		preProcContainer,
		interimProcContainer,
		gasHandler,
	)
	if err != nil {
		return nil, err
	}

	scDataGetter, err := smartContract.NewSCQueryService(vmContainer, economics.MaxGasLimitPerBlock())
	if err != nil {
		return nil, err
	}

	argsStaking := scToProtocol.ArgStakingToPeer{
		AdrConv:     state.BLSAddressConverter,
		Hasher:      core.Hasher,
		Marshalizer: core.Marshalizer,
		PeerState:   state.PeerAccounts,
		BaseState:   state.AccountsAdapter,
		ArgParser:   argsParser,
		CurrTxs:     data.Datapool.CurrentBlockTxs(),
		ScQuery:     scDataGetter,
	}
	smartContractToProtocol, err := scToProtocol.NewStakingToPeer(argsStaking)
	if err != nil {
		return nil, err
	}

	argumentsBaseProcessor := block.ArgBaseProcessor{
		Accounts:                     state.AccountsAdapter,
		ForkDetector:                 forkDetector,
		Hasher:                       core.Hasher,
		Marshalizer:                  core.Marshalizer,
		Store:                        data.Store,
		ShardCoordinator:             shardCoordinator,
		NodesCoordinator:             nodesCoordinator,
		SpecialAddressHandler:        specialAddressHandler,
		Uint64Converter:              core.Uint64ByteSliceConverter,
		RequestHandler:               requestHandler,
		Core:                         coreServiceContainer,
		BlockChainHook:               vmFactory.BlockChainHookImpl(),
		TxCoordinator:                txCoordinator,
		ValidatorStatisticsProcessor: validatorStatisticsProcessor,
		EpochStartTrigger:            epochStartTrigger,
		Rounder:                      rounder,
		HeaderValidator:              headerValidator,
		BootStorer:                   bootStorer,
		BlockTracker:                 blockTracker,
		DataPool:                     data.Datapool,
	}
	arguments := block.ArgMetaProcessor{
		ArgBaseProcessor:         argumentsBaseProcessor,
		SCDataGetter:             scDataGetter,
		SCToProtocol:             smartContractToProtocol,
		PeerChangesHandler:       smartContractToProtocol,
		PendingMiniBlocksHandler: pendingMiniBlocksHandler,
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

	initialNodes := processComponents.nodesConfig.InitialNodes
	storageService := processComponents.data.Store

	var peerDataPool peer.DataPool = processComponents.data.Datapool
	if processComponents.shardCoordinator.SelfId() < processComponents.shardCoordinator.NumberOfShards() {
		peerDataPool = processComponents.data.Datapool
	}

	arguments := peer.ArgValidatorStatisticsProcessor{
		InitialNodes:        initialNodes,
		PeerAdapter:         processComponents.state.PeerAccounts,
		AdrConv:             processComponents.state.BLSAddressConverter,
		NodesCoordinator:    processComponents.nodesCoordinator,
		ShardCoordinator:    processComponents.shardCoordinator,
		DataPool:            peerDataPool,
		StorageService:      storageService,
		Marshalizer:         processComponents.core.Marshalizer,
		StakeValue:          processComponents.economicsData.StakeValue(),
		Rater:               processComponents.rater,
		MaxComputableRounds: processComponents.maxComputableRounds,
	}

	validatorStatisticsProcessor, err := peer.NewValidatorStatisticsProcessor(arguments)
	if err != nil {
		return nil, err
	}

	return validatorStatisticsProcessor, nil
}

func generateInMemoryAccountsAdapter(
	accountFactory state.AccountFactory,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) (state.AccountsAdapter, error) {
	trieStorage, err := trie.NewTrieStorageManagerWithoutPruning(createMemUnit())
	if err != nil {
		return nil, err
	}

	tr, err := trie.NewTrie(trieStorage, marshalizer, hasher)
	if err != nil {
		return nil, err
	}

	adb, err := state.NewAccountsDB(tr, hasher, marshalizer, accountFactory)
	if err != nil {
		return nil, err
	}

	return adb, nil
}

func createMemUnit() storage.Storer {
	cache, err := storageUnit.NewCache(storageUnit.LRUCache, 10, 1)
	if err != nil {
		log.Error("error creating cache for mem unit " + err.Error())
		return nil
	}

	unit, err := storageUnit.NewStorageUnit(cache, memorydb.New())
	if err != nil {
		log.Error("error creating unit " + err.Error())
		return nil
	}

	return unit
}

// GetSigningParams returns a key generator, a private key, and a public key
func GetSigningParams(
	ctx *cli.Context,
	skName string,
	skIndexName string,
	skPemFileName string,
	suite crypto.Suite,
) (keyGen crypto.KeyGenerator, privKey crypto.PrivateKey, pubKey crypto.PublicKey, err error) {

	sk, err := getSk(ctx, skName, skIndexName, skPemFileName)
	if err != nil {
		return nil, nil, nil, err
	}

	keyGen = signing.NewKeyGenerator(suite)

	privKey, err = keyGen.PrivateKeyFromByteArray(sk)
	if err != nil {
		return nil, nil, nil, err
	}

	pubKey = privKey.GeneratePublic()

	return keyGen, privKey, pubKey, err
}

// GetPkEncoded returns the encoded public key
func GetPkEncoded(pubKey crypto.PublicKey) string {
	pk, err := pubKey.ToByteArray()
	if err != nil {
		return err.Error()
	}

	return encodeAddress(pk)
}

func encodeAddress(address []byte) string {
	return hex.EncodeToString(address)
}

func decodeAddress(address string) ([]byte, error) {
	return hex.DecodeString(address)
}

func getSk(
	ctx *cli.Context,
	skName string,
	skIndexName string,
	skPemFileName string,
) ([]byte, error) {

	//if flag is defined, it shall overwrite what was read from pem file
	if ctx.GlobalIsSet(skName) {
		encodedSk := []byte(ctx.GlobalString(skName))
		return decodeAddress(string(encodedSk))
	}

	skIndex := ctx.GlobalInt(skIndexName)
	encodedSk, err := core.LoadSkFromPemFile(skPemFileName, skIndex)
	if err != nil {
		return nil, err
	}

	return decodeAddress(string(encodedSk))
}
