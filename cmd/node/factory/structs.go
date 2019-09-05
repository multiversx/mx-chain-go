package factory

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	"io"
	"math/big"
	"path/filepath"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/round"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/genesis"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/core/serviceContainer"
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
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	metafactoryDataRetriever "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/metachain"
	shardfactoryDataRetriever "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/shard"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	factoryP2P "github.com/ElrondNetwork/elrond-go/p2p/libp2p/factory"
	"github.com/ElrondNetwork/elrond-go/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	processSync "github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	factoryViews "github.com/ElrondNetwork/elrond-go/statusHandler/factory"
	"github.com/ElrondNetwork/elrond-go/statusHandler/view"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/urfave/cli"
)

const (
	// BlsHashSize specifies the hash size for using bls scheme
	BlsHashSize = 16

	// BlsConsensusType specifies te signature scheme used in the consensus
	BlsConsensusType = "bls"

	// BnConsensusType specifies te signature scheme used in the consensus
	BnConsensusType = "bn"

	// MaxTxsToRequest specifies the maximum number of txs to request
	MaxTxsToRequest = 100
)

var log = logger.DefaultLogger()

// Network struct holds the network components of the Elrond protocol
type Network struct {
	NetMessenger p2p.Messenger
}

// Core struct holds the core components of the Elrond protocol
type Core struct {
	Hasher                   hashing.Hasher
	Marshalizer              marshal.Marshalizer
	Trie                     data.Trie
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	StatusHandler            core.AppStatusHandler
}

// State struct holds the state components of the Elrond protocol
type State struct {
	AddressConverter  state.AddressConverter
	AccountsAdapter   state.AccountsAdapter
	InBalanceForShard map[string]*big.Int
}

// Data struct holds the data components of the Elrond protocol
type Data struct {
	Blkc         data.ChainHandler
	Store        dataRetriever.StorageService
	Datapool     dataRetriever.PoolsHolder
	MetaDatapool dataRetriever.MetaPoolsHolder
}

// Crypto struct holds the crypto components of the Elrond protocol
type Crypto struct {
	TxSingleSigner crypto.SingleSigner
	SingleSigner   crypto.SingleSigner
	MultiSigner    crypto.MultiSigner
	TxSignKeyGen   crypto.KeyGenerator
	TxSignPrivKey  crypto.PrivateKey
	TxSignPubKey   crypto.PublicKey
	InitialPubKeys map[uint32][]string
}

// Process struct holds the process components of the Elrond protocol
type Process struct {
	InterceptorsContainer process.InterceptorsContainer
	ResolversFinder       dataRetriever.ResolversFinder
	Rounder               consensus.Rounder
	ForkDetector          process.ForkDetector
	BlockProcessor        process.BlockProcessor
	BlockTracker          process.BlocksTracker
}

type coreComponentsFactoryArgs struct {
	config   *config.Config
	uniqueID string
}

// NewCoreComponentsFactoryArgs initializes the arguments necessary for creating the core components
func NewCoreComponentsFactoryArgs(config *config.Config, uniqueID string) *coreComponentsFactoryArgs {
	return &coreComponentsFactoryArgs{
		config:   config,
		uniqueID: uniqueID,
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

	merkleTrie, err := getTrie(args.config.AccountsTrieStorage, marshalizer, hasher, args.uniqueID)
	if err != nil {
		return nil, errors.New("error creating trie: " + err.Error())
	}
	uint64ByteSliceConverter := uint64ByteSlice.NewBigEndianConverter()

	return &Core{
		Hasher:                   hasher,
		Marshalizer:              marshalizer,
		Trie:                     merkleTrie,
		Uint64ByteSliceConverter: uint64ByteSliceConverter,
		StatusHandler:            statusHandler.NewNilStatusHandler(),
	}, nil
}

type stateComponentsFactoryArgs struct {
	config           *config.Config
	genesisConfig    *sharding.Genesis
	shardCoordinator sharding.Coordinator
	core             *Core
}

// NewStateComponentsFactoryArgs initializes the arguments necessary for creating the state components
func NewStateComponentsFactoryArgs(
	config *config.Config,
	genesisConfig *sharding.Genesis,
	shardCoordinator sharding.Coordinator,
	core *Core,
) *stateComponentsFactoryArgs {
	return &stateComponentsFactoryArgs{
		config:           config,
		genesisConfig:    genesisConfig,
		shardCoordinator: shardCoordinator,
		core:             core,
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

	accountFactory, err := factoryState.NewAccountFactoryCreator(factoryState.UserAccount)
	if err != nil {
		return nil, errors.New("could not create account factory: " + err.Error())
	}

	accountsAdapter, err := state.NewAccountsDB(args.core.Trie, args.core.Hasher, args.core.Marshalizer, accountFactory)
	if err != nil {
		return nil, errors.New("could not create accounts adapter: " + err.Error())
	}

	inBalanceForShard, err := args.genesisConfig.InitialNodesBalances(args.shardCoordinator, addressConverter)
	if err != nil {
		return nil, errors.New("initial balances could not be processed " + err.Error())
	}

	return &State{
		AddressConverter:  addressConverter,
		AccountsAdapter:   accountsAdapter,
		InBalanceForShard: inBalanceForShard,
	}, nil
}

type dataComponentsFactoryArgs struct {
	config           *config.Config
	shardCoordinator sharding.Coordinator
	core             *Core
	uniqueID         string
}

// NewDataComponentsFactoryArgs initializes the arguments necessary for creating the data components
func NewDataComponentsFactoryArgs(
	config *config.Config,
	shardCoordinator sharding.Coordinator,
	core *Core,
	uniqueID string,
) *dataComponentsFactoryArgs {
	return &dataComponentsFactoryArgs{
		config:           config,
		shardCoordinator: shardCoordinator,
		core:             core,
		uniqueID:         uniqueID,
	}
}

// DataComponentsFactory creates the data components
func DataComponentsFactory(args *dataComponentsFactoryArgs) (*Data, error) {
	var datapool dataRetriever.PoolsHolder
	var metaDatapool dataRetriever.MetaPoolsHolder
	blkc, err := createBlockChainFromConfig(args.config, args.shardCoordinator, args.core.StatusHandler)
	if err != nil {
		return nil, errors.New("could not create block chain: " + err.Error())
	}

	store, err := createDataStoreFromConfig(args.config, args.shardCoordinator, args.uniqueID)
	if err != nil {
		return nil, errors.New("could not create local data store: " + err.Error())
	}

	if args.shardCoordinator.SelfId() < args.shardCoordinator.NumberOfShards() {
		datapool, err = createShardDataPoolFromConfig(args.config, args.core.Uint64ByteSliceConverter)
		if err != nil {
			return nil, errors.New("could not create shard data pools: " + err.Error())
		}
	}
	if args.shardCoordinator.SelfId() == sharding.MetachainShardId {
		metaDatapool, err = createMetaDataPoolFromConfig(args.config, args.core.Uint64ByteSliceConverter)
		if err != nil {
			return nil, errors.New("could not create shard data pools: " + err.Error())
		}
	}
	if datapool == nil && metaDatapool == nil {
		return nil, errors.New("could not create data pools: ")
	}

	return &Data{
		Blkc:         blkc,
		Store:        store,
		Datapool:     datapool,
		MetaDatapool: metaDatapool,
	}, nil
}

type cryptoComponentsFactoryArgs struct {
	ctx                          *cli.Context
	config                       *config.Config
	nodesConfig                  *sharding.NodesSetup
	shardCoordinator             sharding.Coordinator
	keyGen                       crypto.KeyGenerator
	privKey                      crypto.PrivateKey
	log                          *logger.Logger
	initialBalancesSkPemFileName string
	txSignSkName                 string
	txSignSkIndexName            string
}

// NewCryptoComponentsFactoryArgs initializes the arguments necessary for creating the crypto components
func NewCryptoComponentsFactoryArgs(
	ctx *cli.Context,
	config *config.Config,
	nodesConfig *sharding.NodesSetup,
	shardCoordinator sharding.Coordinator,
	keyGen crypto.KeyGenerator,
	privKey crypto.PrivateKey,
	log *logger.Logger,
	initialBalancesSkPemFileName string,
	txSignSkName string,
	txSignSkIndexName string,
) *cryptoComponentsFactoryArgs {
	return &cryptoComponentsFactoryArgs{
		ctx:                          ctx,
		config:                       config,
		nodesConfig:                  nodesConfig,
		shardCoordinator:             shardCoordinator,
		keyGen:                       keyGen,
		privKey:                      privKey,
		log:                          log,
		initialBalancesSkPemFileName: initialBalancesSkPemFileName,
		txSignSkName:                 txSignSkName,
		txSignSkIndexName:            txSignSkIndexName,
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

	initialBalancesSkPemFileName := args.ctx.GlobalString(args.initialBalancesSkPemFileName)
	txSignKeyGen, txSignPrivKey, txSignPubKey, err := GetSigningParams(
		args.ctx,
		args.log,
		args.txSignSkName,
		args.txSignSkIndexName,
		initialBalancesSkPemFileName,
		kyber.NewBlakeSHA256Ed25519())
	if err != nil {
		return nil, err
	}
	args.log.Info("Starting with tx sign public key: " + GetPkEncoded(txSignPubKey))

	return &Crypto{
		TxSingleSigner: txSingleSigner,
		SingleSigner:   singleSigner,
		MultiSigner:    multiSigner,
		TxSignKeyGen:   txSignKeyGen,
		TxSignPrivKey:  txSignPrivKey,
		TxSignPubKey:   txSignPubKey,
		InitialPubKeys: initialPubKeys,
	}, nil
}

// NetworkComponentsFactory creates the network components
func NetworkComponentsFactory(p2pConfig *config.P2PConfig, log *logger.Logger, core *Core) (*Network, error) {
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
	genesisConfig        *sharding.Genesis
	rewardsConfig        *config.RewardConfig
	nodesConfig          *sharding.NodesSetup
	syncer               ntp.SyncTimer
	shardCoordinator     sharding.Coordinator
	nodesCoordinator     sharding.NodesCoordinator
	data                 *Data
	core                 *Core
	crypto               *Crypto
	state                *State
	network              *Network
	coreServiceContainer serviceContainer.Core
}

// NewProcessComponentsFactoryArgs initializes the arguments necessary for creating the process components
func NewProcessComponentsFactoryArgs(
	genesisConfig *sharding.Genesis,
	rewardsConfig *config.RewardConfig,
	nodesConfig *sharding.NodesSetup,
	syncer ntp.SyncTimer,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	data *Data,
	core *Core,
	crypto *Crypto,
	state *State,
	network *Network,
	coreServiceContainer serviceContainer.Core,
) *processComponentsFactoryArgs {
	return &processComponentsFactoryArgs{
		genesisConfig:        genesisConfig,
		rewardsConfig:        rewardsConfig,
		nodesConfig:          nodesConfig,
		syncer:               syncer,
		shardCoordinator:     shardCoordinator,
		nodesCoordinator:     nodesCoordinator,
		data:                 data,
		core:                 core,
		crypto:               crypto,
		state:                state,
		network:              network,
		coreServiceContainer: coreServiceContainer,
	}
}

// ProcessComponentsFactory creates the process components
func ProcessComponentsFactory(args *processComponentsFactoryArgs) (*Process, error) {
	interceptorContainerFactory, resolversContainerFactory, err := newInterceptorAndResolverContainerFactory(
		args.shardCoordinator, args.nodesCoordinator, args.data, args.core, args.crypto, args.state, args.network)
	if err != nil {
		return nil, err
	}

	//TODO refactor all these factory calls
	interceptorsContainer, err := interceptorContainerFactory.Create()
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

	rounder, err := round.NewRound(
		time.Unix(args.nodesConfig.StartTime, 0),
		args.syncer.CurrentTime(),
		time.Millisecond*time.Duration(args.nodesConfig.RoundDuration),
		args.syncer)
	if err != nil {
		return nil, err
	}

	forkDetector, err := processSync.NewBasicForkDetector(rounder)
	if err != nil {
		return nil, err
	}

	shardsGenesisBlocks, err := generateGenesisHeadersAndApplyInitialBalances(
		args.core,
		args.state,
		args.shardCoordinator,
		args.nodesConfig,
		args.genesisConfig,
	)
	if err != nil {
		return nil, err
	}

	err = prepareGenesisBlock(args, shardsGenesisBlocks)
	if err != nil {
		return nil, err
	}

	blockProcessor, blockTracker, err := newBlockProcessorAndTracker(
		resolversFinder,
		args.shardCoordinator,
		args.nodesCoordinator,
		args.rewardsConfig,
		args.data,
		args.core,
		args.state,
		forkDetector,
		shardsGenesisBlocks,
		args.nodesConfig,
		args.coreServiceContainer,
	)

	if err != nil {
		return nil, err
	}

	return &Process{
		InterceptorsContainer: interceptorsContainer,
		ResolversFinder:       resolversFinder,
		Rounder:               rounder,
		ForkDetector:          forkDetector,
		BlockProcessor:        blockProcessor,
		BlockTracker:          blockTracker,
	}, nil
}

func prepareGenesisBlock(args *processComponentsFactoryArgs, shardsGenesisBlocks map[uint32]data.HeaderHandler) error {
	genesisBlock, ok := shardsGenesisBlocks[args.shardCoordinator.SelfId()]
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
		log.LogIfError(errNotCritical)

	} else {
		errNotCritical := args.data.Store.Put(dataRetriever.BlockHeaderUnit, genesisBlockHash, marshalizedBlock)
		log.LogIfError(errNotCritical)
	}

	return nil
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
		srr.index = srr.index % len(srr.seed)
	}

	return len(p), nil
}

// CreateStatusHandlerPresenter will return an instance of PresenterStatusHandler
func CreateStatusHandlerPresenter() view.Presenter {
	presenterStatusHandlerFactory := factoryViews.NewPresenterFactory()

	return presenterStatusHandlerFactory.Create()
}

// CreateViews will start an termui console  and will return an object if cannot create and start termuiConsole
func CreateViews(presenter view.Presenter) ([]factoryViews.Viewer, error) {
	viewsFactory, err := factoryViews.NewViewsFactory(presenter)
	if err != nil {
		return nil, err
	}

	views, err := viewsFactory.Create()
	if err != nil {
		return nil, err
	}

	for _, v := range views {
		err = v.Start()
		if err != nil {
			return nil, err
		}
	}

	return views, nil
}

func getHasherFromConfig(cfg *config.Config) (hashing.Hasher, error) {
	switch cfg.Hasher.Type {
	case "sha256":
		return sha256.Sha256{}, nil
	case "blake2b":
		return blake2b.Blake2b{}, nil
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

func getTrie(
	cfg config.StorageConfig,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	uniqueID string,
) (data.Trie, error) {

	accountsTrieStorage, err := storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(cfg.Cache),
		getDBFromConfig(cfg.DB, uniqueID),
		getBloomFromConfig(cfg.Bloom),
	)
	if err != nil {
		return nil, errors.New("error creating accountsTrieStorage: " + err.Error())
	}

	return trie.NewTrie(accountsTrieStorage, marshalizer, hasher)
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
		blockChain, err := blockchain.NewBlockChain(badBlockCache)
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
		blockChain, err := blockchain.NewMetaChain(badBlockCache)
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
	uniqueID string,
) (dataRetriever.StorageService, error) {
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return createShardDataStoreFromConfig(config, shardCoordinator, uniqueID)
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		return createMetaChainDataStoreFromConfig(config, shardCoordinator, uniqueID)
	}
	return nil, errors.New("can not create data store")
}

func createShardDataStoreFromConfig(
	config *config.Config,
	shardCoordinator sharding.Coordinator,
	uniqueID string,
) (dataRetriever.StorageService, error) {

	var headerUnit *storageUnit.Unit
	var peerBlockUnit *storageUnit.Unit
	var miniBlockUnit *storageUnit.Unit
	var txUnit *storageUnit.Unit
	var metachainHeaderUnit *storageUnit.Unit
	var unsignedTxUnit *storageUnit.Unit
	var rewardTxUnit *storageUnit.Unit
	var metaHdrHashNonceUnit *storageUnit.Unit
	var shardHdrHashNonceUnit *storageUnit.Unit
	var err error

	defer func() {
		// cleanup
		if err != nil {
			if headerUnit != nil {
				_ = headerUnit.DestroyUnit()
			}
			if peerBlockUnit != nil {
				_ = peerBlockUnit.DestroyUnit()
			}
			if miniBlockUnit != nil {
				_ = miniBlockUnit.DestroyUnit()
			}
			if txUnit != nil {
				_ = txUnit.DestroyUnit()
			}
			if unsignedTxUnit != nil {
				_ = unsignedTxUnit.DestroyUnit()
			}
			if rewardTxUnit != nil {
				_ = rewardTxUnit.DestroyUnit()
			}
			if metachainHeaderUnit != nil {
				_ = metachainHeaderUnit.DestroyUnit()
			}
			if metaHdrHashNonceUnit != nil {
				_ = metaHdrHashNonceUnit.DestroyUnit()
			}
			if shardHdrHashNonceUnit != nil {
				_ = shardHdrHashNonceUnit.DestroyUnit()
			}
		}
	}()

	txUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.TxStorage.Cache),
		getDBFromConfig(config.TxStorage.DB, uniqueID),
		getBloomFromConfig(config.TxStorage.Bloom))
	if err != nil {
		return nil, err
	}

	unsignedTxUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.UnsignedTransactionStorage.Cache),
		getDBFromConfig(config.UnsignedTransactionStorage.DB, uniqueID),
		getBloomFromConfig(config.UnsignedTransactionStorage.Bloom))
	if err != nil {
		return nil, err
	}

	rewardTxUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.RewardTxStorage.Cache),
		getDBFromConfig(config.RewardTxStorage.DB, uniqueID),
		getBloomFromConfig(config.RewardTxStorage.Bloom))
	if err != nil {
		return nil, err
	}

	miniBlockUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.MiniBlocksStorage.Cache),
		getDBFromConfig(config.MiniBlocksStorage.DB, uniqueID),
		getBloomFromConfig(config.MiniBlocksStorage.Bloom))
	if err != nil {
		return nil, err
	}

	peerBlockUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.PeerBlockBodyStorage.Cache),
		getDBFromConfig(config.PeerBlockBodyStorage.DB, uniqueID),
		getBloomFromConfig(config.PeerBlockBodyStorage.Bloom))
	if err != nil {
		return nil, err
	}

	headerUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.BlockHeaderStorage.Cache),
		getDBFromConfig(config.BlockHeaderStorage.DB, uniqueID),
		getBloomFromConfig(config.BlockHeaderStorage.Bloom))
	if err != nil {
		return nil, err
	}

	metachainHeaderUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.MetaBlockStorage.Cache),
		getDBFromConfig(config.MetaBlockStorage.DB, uniqueID),
		getBloomFromConfig(config.MetaBlockStorage.Bloom))
	if err != nil {
		return nil, err
	}

	metaHdrHashNonceUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.MetaHdrNonceHashStorage.Cache),
		getDBFromConfig(config.MetaHdrNonceHashStorage.DB, uniqueID),
		getBloomFromConfig(config.MetaHdrNonceHashStorage.Bloom),
	)
	if err != nil {
		return nil, err
	}

	shardHdrHashNonceUnit, err = storageUnit.NewShardedStorageUnitFromConf(
		getCacherFromConfig(config.ShardHdrNonceHashStorage.Cache),
		getDBFromConfig(config.ShardHdrNonceHashStorage.DB, uniqueID),
		getBloomFromConfig(config.ShardHdrNonceHashStorage.Bloom),
		shardCoordinator.SelfId(),
	)
	if err != nil {
		return nil, err
	}

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, txUnit)
	store.AddStorer(dataRetriever.MiniBlockUnit, miniBlockUnit)
	store.AddStorer(dataRetriever.PeerChangesUnit, peerBlockUnit)
	store.AddStorer(dataRetriever.BlockHeaderUnit, headerUnit)
	store.AddStorer(dataRetriever.MetaBlockUnit, metachainHeaderUnit)
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, unsignedTxUnit)
	store.AddStorer(dataRetriever.RewardTransactionUnit, rewardTxUnit)
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, metaHdrHashNonceUnit)
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardCoordinator.SelfId())
	store.AddStorer(hdrNonceHashDataUnit, shardHdrHashNonceUnit)

	return store, err
}

func createMetaChainDataStoreFromConfig(
	config *config.Config,
	shardCoordinator sharding.Coordinator,
	uniqueID string,
) (dataRetriever.StorageService, error) {
	var peerDataUnit, shardDataUnit, metaBlockUnit, headerUnit, metaHdrHashNonceUnit *storageUnit.Unit
	var shardHdrHashNonceUnits []*storageUnit.Unit
	var err error

	defer func() {
		// cleanup
		if err != nil {
			if peerDataUnit != nil {
				_ = peerDataUnit.DestroyUnit()
			}
			if shardDataUnit != nil {
				_ = shardDataUnit.DestroyUnit()
			}
			if metaBlockUnit != nil {
				_ = metaBlockUnit.DestroyUnit()
			}
			if headerUnit != nil {
				_ = headerUnit.DestroyUnit()
			}
			if metaHdrHashNonceUnit != nil {
				_ = metaHdrHashNonceUnit.DestroyUnit()
			}
			if shardHdrHashNonceUnits != nil {
				for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
					_ = shardHdrHashNonceUnits[i].DestroyUnit()
				}
			}
		}
	}()

	metaBlockUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.MetaBlockStorage.Cache),
		getDBFromConfig(config.MetaBlockStorage.DB, uniqueID),
		getBloomFromConfig(config.MetaBlockStorage.Bloom))
	if err != nil {
		return nil, err
	}

	shardDataUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.ShardDataStorage.Cache),
		getDBFromConfig(config.ShardDataStorage.DB, uniqueID),
		getBloomFromConfig(config.ShardDataStorage.Bloom))
	if err != nil {
		return nil, err
	}

	peerDataUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.PeerDataStorage.Cache),
		getDBFromConfig(config.PeerDataStorage.DB, uniqueID),
		getBloomFromConfig(config.PeerDataStorage.Bloom))
	if err != nil {
		return nil, err
	}

	headerUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.BlockHeaderStorage.Cache),
		getDBFromConfig(config.BlockHeaderStorage.DB, uniqueID),
		getBloomFromConfig(config.BlockHeaderStorage.Bloom))
	if err != nil {
		return nil, err
	}

	metaHdrHashNonceUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.MetaHdrNonceHashStorage.Cache),
		getDBFromConfig(config.MetaHdrNonceHashStorage.DB, uniqueID),
		getBloomFromConfig(config.MetaHdrNonceHashStorage.Bloom),
	)
	if err != nil {
		return nil, err
	}

	shardHdrHashNonceUnits = make([]*storageUnit.Unit, shardCoordinator.NumberOfShards())
	for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
		shardHdrHashNonceUnits[i], err = storageUnit.NewShardedStorageUnitFromConf(
			getCacherFromConfig(config.ShardHdrNonceHashStorage.Cache),
			getDBFromConfig(config.ShardHdrNonceHashStorage.DB, uniqueID),
			getBloomFromConfig(config.ShardHdrNonceHashStorage.Bloom),
			i,
		)
		if err != nil {
			return nil, err
		}
	}

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, metaBlockUnit)
	store.AddStorer(dataRetriever.MetaShardDataUnit, shardDataUnit)
	store.AddStorer(dataRetriever.MetaPeerDataUnit, peerDataUnit)
	store.AddStorer(dataRetriever.BlockHeaderUnit, headerUnit)
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, metaHdrHashNonceUnit)
	for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(i)
		store.AddStorer(hdrNonceHashDataUnit, shardHdrHashNonceUnits[i])
	}

	return store, err
}

func createShardDataPoolFromConfig(
	config *config.Config,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
) (dataRetriever.PoolsHolder, error) {

	log.Info("creatingShardDataPool from config")

	txPool, err := shardedData.NewShardedData(getCacherFromConfig(config.TxDataPool))
	if err != nil {
		log.Info("error creating txpool")
		return nil, err
	}

	uTxPool, err := shardedData.NewShardedData(getCacherFromConfig(config.UnsignedTransactionDataPool))
	if err != nil {
		log.Info("error creating smart contract result pool")
		return nil, err
	}

	rewardTxPool, err := shardedData.NewShardedData(getCacherFromConfig(config.RewardTransactionDataPool))
	if err != nil {
		log.Info("error creating reward transaction pool")
		return nil, err
	}

	cacherCfg := getCacherFromConfig(config.BlockHeaderDataPool)
	hdrPool, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		log.Info("error creating hdrpool")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.MetaBlockBodyDataPool)
	metaBlockBody, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		log.Info("error creating metaBlockBody")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.BlockHeaderNoncesDataPool)
	hdrNoncesCacher, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		log.Info("error creating hdrNoncesCacher")
		return nil, err
	}
	hdrNonces, err := dataPool.NewNonceSyncMapCacher(hdrNoncesCacher, uint64ByteSliceConverter)
	if err != nil {
		log.Info("error creating hdrNonces")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.TxBlockBodyDataPool)
	txBlockBody, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		log.Info("error creating txBlockBody")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.PeerBlockBodyDataPool)
	peerChangeBlockBody, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		log.Info("error creating peerChangeBlockBody")
		return nil, err
	}

	return dataPool.NewShardedDataPool(
		txPool,
		uTxPool,
		rewardTxPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		metaBlockBody,
	)
}

func createMetaDataPoolFromConfig(
	config *config.Config,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
) (dataRetriever.MetaPoolsHolder, error) {
	cacherCfg := getCacherFromConfig(config.MetaBlockBodyDataPool)
	metaBlockBody, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		log.Info("error creating metaBlockBody")
		return nil, err
	}

	miniBlockHashes, err := shardedData.NewShardedData(getCacherFromConfig(config.MiniBlockHeaderHashesDataPool))
	if err != nil {
		log.Info("error creating miniBlockHashes")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.ShardHeadersDataPool)
	shardHeaders, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		log.Info("error creating shardHeaders")
		return nil, err
	}

	headersNoncesCacher, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		log.Info("error creating shard headers nonces pool")
		return nil, err
	}
	headersNonces, err := dataPool.NewNonceSyncMapCacher(headersNoncesCacher, uint64ByteSliceConverter)
	if err != nil {
		log.Info("error creating shard headers nonces pool")
		return nil, err
	}

	return dataPool.NewMetaDataPool(metaBlockBody, miniBlockHashes, shardHeaders, headersNonces)
}

func createSingleSigner(config *config.Config) (crypto.SingleSigner, error) {
	switch config.Consensus.Type {
	case BlsConsensusType:
		return &singlesig.BlsSingleSigner{}, nil
	case BnConsensusType:
		return &singlesig.SchnorrSigner{}, nil
	}

	return nil, errors.New("no consensus type provided in config file")
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
			return blake2b.Blake2b{HashSize: BlsHashSize}, nil
		}
		return blake2b.Blake2b{}, nil
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
	case BnConsensusType:
		return multisig.NewBelNevMultisig(hasher, pubKeys, privateKey, keyGen, uint16(0))
	}

	return nil, errors.New("no consensus type provided in config file")
}

func createNetMessenger(
	p2pConfig *config.P2PConfig,
	log *logger.Logger,
	randReader io.Reader,
) (p2p.Messenger, error) {

	if p2pConfig.Node.Port < 0 {
		return nil, errors.New("cannot start node on port < 0")
	}

	pDiscoveryFactory := factoryP2P.NewPeerDiscovererCreator(*p2pConfig)
	pDiscoverer, err := pDiscoveryFactory.CreatePeerDiscoverer()

	if err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf("Starting with peer discovery: %s", pDiscoverer.Name()))

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
	)
	if err != nil {
		return nil, err
	}

	return nm, nil
}

func newInterceptorAndResolverContainerFactory(
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	data *Data,
	core *Core,
	crypto *Crypto,
	state *State,
	network *Network,
) (process.InterceptorsContainerFactory, dataRetriever.ResolversContainerFactory, error) {
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return newShardInterceptorAndResolverContainerFactory(
			shardCoordinator,
			nodesCoordinator,
			data,
			core,
			crypto,
			state,
			network,
		)
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		return newMetaInterceptorAndResolverContainerFactory(
			shardCoordinator,
			nodesCoordinator,
			data,
			core,
			crypto,
			network,
		)
	}

	return nil, nil, errors.New("could not create interceptor and resolver container factory")
}

func newShardInterceptorAndResolverContainerFactory(
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	data *Data,
	core *Core,
	crypto *Crypto,
	state *State,
	network *Network,
) (process.InterceptorsContainerFactory, dataRetriever.ResolversContainerFactory, error) {
	//TODO add a real chronology validator and remove null chronology validator
	interceptorContainerFactory, err := shard.NewInterceptorsContainerFactory(
		shardCoordinator,
		nodesCoordinator,
		network.NetMessenger,
		data.Store,
		core.Marshalizer,
		core.Hasher,
		crypto.TxSignKeyGen,
		crypto.TxSingleSigner,
		crypto.MultiSigner,
		data.Datapool,
		state.AddressConverter,
	)
	if err != nil {
		return nil, nil, err
	}

	dataPacker, err := partitioning.NewSizeDataPacker(core.Marshalizer)
	if err != nil {
		return nil, nil, err
	}

	resolversContainerFactory, err := shardfactoryDataRetriever.NewResolversContainerFactory(
		shardCoordinator,
		network.NetMessenger,
		data.Store,
		core.Marshalizer,
		data.Datapool,
		core.Uint64ByteSliceConverter,
		dataPacker,
	)
	if err != nil {
		return nil, nil, err
	}

	return interceptorContainerFactory, resolversContainerFactory, nil
}

func newMetaInterceptorAndResolverContainerFactory(
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	data *Data,
	core *Core,
	crypto *Crypto,
	network *Network,
) (process.InterceptorsContainerFactory, dataRetriever.ResolversContainerFactory, error) {
	//TODO add a real chronology validator and remove null chronology validator
	interceptorContainerFactory, err := metachain.NewInterceptorsContainerFactory(
		shardCoordinator,
		nodesCoordinator,
		network.NetMessenger,
		data.Store,
		core.Marshalizer,
		core.Hasher,
		crypto.MultiSigner,
		data.MetaDatapool,
	)
	if err != nil {
		return nil, nil, err
	}
	resolversContainerFactory, err := metafactoryDataRetriever.NewResolversContainerFactory(
		shardCoordinator,
		network.NetMessenger,
		data.Store,
		core.Marshalizer,
		data.MetaDatapool,
		core.Uint64ByteSliceConverter,
	)
	if err != nil {
		return nil, nil, err
	}
	return interceptorContainerFactory, resolversContainerFactory, nil
}

func generateGenesisHeadersAndApplyInitialBalances(
	coreComponents *Core,
	stateComponents *State,
	shardCoordinator sharding.Coordinator,
	nodesSetup *sharding.NodesSetup,
	genesisConfig *sharding.Genesis,
) (map[uint32]data.HeaderHandler, error) {
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

	shardsGenesisBlocks := make(map[uint32]data.HeaderHandler)

	for shardId := uint32(0); shardId < shardCoordinator.NumberOfShards(); shardId++ {
		isCurrentShard := shardId == shardCoordinator.SelfId()
		if isCurrentShard {
			continue
		}

		newShardCoordinator, account, err := createInMemoryShardCoordinatorAndAccount(
			coreComponents,
			shardCoordinator.NumberOfShards(),
			shardId,
		)
		if err != nil {
			return nil, err
		}

		genesisBlock, err := createGenesisBlockAndApplyInitialBalances(
			account,
			newShardCoordinator,
			stateComponents.AddressConverter,
			genesisConfig,
			uint64(nodesSetup.StartTime),
		)
		if err != nil {
			return nil, err
		}

		shardsGenesisBlocks[shardId] = genesisBlock
	}

	genesisBlockForCurrentShard, err := createGenesisBlockAndApplyInitialBalances(
		stateComponents.AccountsAdapter,
		shardCoordinator,
		stateComponents.AddressConverter,
		genesisConfig,
		uint64(nodesSetup.StartTime),
	)
	if err != nil {
		return nil, err
	}

	shardsGenesisBlocks[shardCoordinator.SelfId()] = genesisBlockForCurrentShard

	genesisBlock, err := genesis.CreateMetaGenesisBlock(
			uint64(nodesSetup.StartTime),
			nodesSetup.InitialNodesPubKeys(),
		)

	if err != nil {
		return nil, err
	}

	shardsGenesisBlocks[sharding.MetachainShardId] = genesisBlock

	return shardsGenesisBlocks, nil
}

func createGenesisBlockAndApplyInitialBalances(
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	addressConverter state.AddressConverter,
	genesisConfig *sharding.Genesis,
	startTime uint64,
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

	accounts := generateInMemoryAccountsAdapter(
		accountFactory,
		coreComponents.Hasher,
		coreComponents.Marshalizer,
	)

	return newShardCoordinator, accounts, nil
}

func newBlockProcessorAndTracker(
	resolversFinder dataRetriever.ResolversFinder,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	rewardsConfig *config.RewardConfig,
	data *Data,
	core *Core,
	state *State,
	forkDetector process.ForkDetector,
	shardsGenesisBlocks map[uint32]data.HeaderHandler,
	nodesConfig *sharding.NodesSetup,
	coreServiceContainer serviceContainer.Core,
) (process.BlockProcessor, process.BlocksTracker, error) {

	if rewardsConfig.CommunityAddress == "" || rewardsConfig.BurnAddress == "" {
		return nil, nil, errors.New("rewards configuration missing")
	}

	communityAddress, _ := hex.DecodeString(rewardsConfig.CommunityAddress)
	burnAddress, _ := hex.DecodeString(rewardsConfig.BurnAddress)

	// TODO: construct this correctly on the PR
	specialAddressHolder, err := address.NewSpecialAddressHolder(
		communityAddress,
		burnAddress,
		state.AddressConverter,
		shardCoordinator)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove nodesConfig as no longer needed with nodes coordinator available
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return newShardBlockProcessorAndTracker(
			resolversFinder,
			shardCoordinator,
			nodesCoordinator,
			specialAddressHolder,
			data,
			core,
			state,
			forkDetector,
			shardsGenesisBlocks,
			nodesConfig,
			coreServiceContainer,
		)
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		return newMetaBlockProcessorAndTracker(
			resolversFinder,
			shardCoordinator,
			nodesCoordinator,
			specialAddressHolder,
			data,
			core,
			state,
			forkDetector,
			shardsGenesisBlocks,
			coreServiceContainer,
		)
	}

	return nil, nil, errors.New("could not create block processor and tracker")
}

func newShardBlockProcessorAndTracker(
	resolversFinder dataRetriever.ResolversFinder,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	specialAddressHandler process.SpecialAddressHandler,
	data *Data,
	core *Core,
	state *State,
	forkDetector process.ForkDetector,
	shardsGenesisBlocks map[uint32]data.HeaderHandler,
	nodesConfig *sharding.NodesSetup,
	coreServiceContainer serviceContainer.Core,
) (process.BlockProcessor, process.BlocksTracker, error) {
	argsParser, err := smartContract.NewAtArgumentParser()
	if err != nil {
		return nil, nil, err
	}

	vmFactory, err := shard.NewVMContainerFactory(state.AccountsAdapter, state.AddressConverter)
	if err != nil {
		return nil, nil, err
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, nil, err
	}

	if err != nil {
		return nil, nil, err
	}

	interimProcFactory, err := shard.NewIntermediateProcessorsContainerFactory(
		shardCoordinator,
		core.Marshalizer,
		core.Hasher,
		state.AddressConverter,
		specialAddressHandler,
		data.Store,
		data.Datapool,
	)
	if err != nil {
		return nil, nil, err
	}

	interimProcContainer, err := interimProcFactory.Create()
	if err != nil {
		return nil, nil, err
	}

	scForwarder, err := interimProcContainer.Get(dataBlock.SmartContractResultBlock)
	if err != nil {
		return nil, nil, err
	}

	rewardsTxInterim, err := interimProcContainer.Get(dataBlock.RewardsBlockType)
	if err != nil {
		return nil, nil, err
	}

	rewardsTxHandler, ok := rewardsTxInterim.(process.TransactionFeeHandler)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	scProcessor, err := smartContract.NewSmartContractProcessor(
		vmContainer,
		argsParser,
		core.Hasher,
		core.Marshalizer,
		state.AccountsAdapter,
		vmFactory.VMAccountsDB(),
		state.AddressConverter,
		shardCoordinator,
		scForwarder,
		rewardsTxHandler,
	)
	if err != nil {
		return nil, nil, err
	}

	requestHandler, err := requestHandlers.NewShardResolverRequestHandler(
		resolversFinder,
		factory.TransactionTopic,
		factory.UnsignedTransactionTopic,
		factory.RewardsTransactionTopic,
		factory.MiniBlocksTopic,
		factory.MetachainBlocksTopic,
		MaxTxsToRequest,
	)
	if err != nil {
		return nil, nil, err
	}

	rewardsTxProcessor, err := rewardTransaction.NewRewardTxProcessor(
		state.AccountsAdapter,
		state.AddressConverter,
		shardCoordinator,
		rewardsTxInterim,
	)
	if err != nil {
		return nil, nil, err
	}

	txTypeHandler, err := coordinator.NewTxTypeHandler(state.AddressConverter, shardCoordinator, state.AccountsAdapter)
	if err != nil {
		return nil, nil, err
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
	)
	if err != nil {
		return nil, nil, errors.New("could not create transaction processor: " + err.Error())
	}

	blockTracker, err := track.NewShardBlockTracker(
		data.Datapool,
		core.Marshalizer,
		shardCoordinator,
		data.Store,
	)
	if err != nil {
		return nil, nil, err
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
	)
	if err != nil {
		return nil, nil, err
	}

	preProcContainer, err := preProcFactory.Create()
	if err != nil {
		return nil, nil, err
	}

	txCoordinator, err := coordinator.NewTransactionCoordinator(
		shardCoordinator,
		state.AccountsAdapter,
		data.Datapool,
		requestHandler,
		preProcContainer,
		interimProcContainer,
	)
	if err != nil {
		return nil, nil, err
	}

	blockProcessor, err := block.NewShardProcessor(
		coreServiceContainer,
		data.Datapool,
		data.Store,
		core.Hasher,
		core.Marshalizer,
		state.AccountsAdapter,
		shardCoordinator,
		nodesCoordinator,
		specialAddressHandler,
		forkDetector,
		blockTracker,
		shardsGenesisBlocks,
		requestHandler,
		txCoordinator,
		core.Uint64ByteSliceConverter,
	)
	if err != nil {
		return nil, nil, errors.New("could not create block processor: " + err.Error())
	}

	err = blockProcessor.SetAppStatusHandler(core.StatusHandler)
	if err != nil {
		return nil, nil, err
	}

	return blockProcessor, blockTracker, nil
}

func newMetaBlockProcessorAndTracker(
	resolversFinder dataRetriever.ResolversFinder,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	specialAddressHandler process.SpecialAddressHandler,
	data *Data,
	core *Core,
	state *State,
	forkDetector process.ForkDetector,
	shardsGenesisBlocks map[uint32]data.HeaderHandler,
	coreServiceContainer serviceContainer.Core,
) (process.BlockProcessor, process.BlocksTracker, error) {

	requestHandler, err := requestHandlers.NewMetaResolverRequestHandler(
		resolversFinder,
		factory.ShardHeadersForMetachainTopic,
	)

	if err != nil {
		return nil, nil, err
	}

	blockTracker, err := track.NewMetaBlockTracker()
	if err != nil {
		return nil, nil, err
	}

	metaProcessor, err := block.NewMetaProcessor(
		coreServiceContainer,
		state.AccountsAdapter,
		data.MetaDatapool,
		forkDetector,
		shardCoordinator,
		nodesCoordinator,
		specialAddressHandler,
		core.Hasher,
		core.Marshalizer,
		data.Store,
		shardsGenesisBlocks,
		requestHandler,
		core.Uint64ByteSliceConverter,
	)
	if err != nil {
		return nil, nil, errors.New("could not create block processor: " + err.Error())
	}

	err = metaProcessor.SetAppStatusHandler(core.StatusHandler)
	if err != nil {
		return nil, nil, err
	}

	return metaProcessor, blockTracker, nil
}

func getCacherFromConfig(cfg config.CacheConfig) storageUnit.CacheConfig {
	return storageUnit.CacheConfig{
		Size:   cfg.Size,
		Type:   storageUnit.CacheType(cfg.Type),
		Shards: cfg.Shards,
	}
}

func getDBFromConfig(cfg config.DBConfig, uniquePath string) storageUnit.DBConfig {
	return storageUnit.DBConfig{
		FilePath:          filepath.Join(uniquePath, cfg.FilePath),
		Type:              storageUnit.DBType(cfg.Type),
		MaxBatchSize:      cfg.MaxBatchSize,
		BatchDelaySeconds: cfg.BatchDelaySeconds,
		MaxOpenFiles:      cfg.MaxOpenFiles,
	}
}

func getBloomFromConfig(cfg config.BloomFilterConfig) storageUnit.BloomConfig {
	var hashFuncs []storageUnit.HasherType
	if cfg.HashFunc != nil {
		hashFuncs = make([]storageUnit.HasherType, 0)
		for _, hf := range cfg.HashFunc {
			hashFuncs = append(hashFuncs, storageUnit.HasherType(hf))
		}
	}

	return storageUnit.BloomConfig{
		Size:     cfg.Size,
		HashFunc: hashFuncs,
	}
}

func generateInMemoryAccountsAdapter(
	accountFactory state.AccountFactory,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) state.AccountsAdapter {

	tr, _ := trie.NewTrie(createMemUnit(), marshalizer, hasher)
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, marshalizer, accountFactory)

	return adb
}

func createMemUnit() storage.Storer {
	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, 10, 1)
	persist, _ := memorydb.New()
	unit, _ := storageUnit.NewStorageUnit(cache, persist)
	return unit
}

// GetSigningParams returns a key generator, a private key, and a public key
func GetSigningParams(
	ctx *cli.Context,
	log *logger.Logger,
	skName string,
	skIndexName string,
	skPemFileName string,
	suite crypto.Suite,
) (keyGen crypto.KeyGenerator, privKey crypto.PrivateKey, pubKey crypto.PublicKey, err error) {

	sk, err := getSk(ctx, log, skName, skIndexName, skPemFileName)
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
	log *logger.Logger,
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
	encodedSk, err := core.LoadSkFromPemFile(skPemFileName, log, skIndex)
	if err != nil {
		return nil, err
	}

	return decodeAddress(string(encodedSk))
}
