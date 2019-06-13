package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/config"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/genesis"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/partitioning"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
	blsMultiSig "github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/multisig"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/addressConverters"
	factoryState "github.com/ElrondNetwork/elrond-go-sandbox/data/state/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/factory/containers"
	metafactoryDataRetriever "github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/factory/metachain"
	shardfactoryDataRetriever "github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/factory/shard"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	factoryP2P "github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/smartContract"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/smartContract/hooks"
	processSync "github.com/ElrondNetwork/elrond-go-sandbox/process/sync"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/track"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/storageUnit"
	"github.com/btcsuite/btcd/btcec"
	crypto2 "github.com/libp2p/go-libp2p-crypto"
	"github.com/urfave/cli"
)

type Core struct {
	hasher                   hashing.Hasher
	marshalizer              marshal.Marshalizer
	tr                       trie.PatriciaMerkelTree
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

type State struct {
	addressConverter  state.AddressConverter
	accountsAdapter   state.AccountsAdapter
	inBalanceForShard map[string]*big.Int
}

type Data struct {
	blkc         data.ChainHandler
	store        dataRetriever.StorageService
	datapool     dataRetriever.PoolsHolder
	metaDatapool dataRetriever.MetaPoolsHolder
}

type Crypto struct {
	txSingleSigner crypto.SingleSigner
	singleSigner   crypto.SingleSigner
	multiSigner    crypto.MultiSigner
	txSignKeyGen   crypto.KeyGenerator
	txSignPrivKey  crypto.PrivateKey
	txSignPubKey   crypto.PublicKey
	initialPubKeys map[uint32][]string
}

type Process struct {
	interceptorsContainer process.InterceptorsContainer
	resolversFinder       dataRetriever.ResolversFinder
	rounder               consensus.Rounder
	forkDetector          process.ForkDetector
	blockProcessor        process.BlockProcessor
	blockTracker          process.BlocksTracker
	netMessenger          p2p.Messenger
}

func coreComponentsFactory(config *config.Config) (*Core, error) {
	hasher, err := getHasherFromConfig(config)
	if err != nil {
		return nil, errors.New("could not create hasher: " + err.Error())
	}

	marshalizer, err := getMarshalizerFromConfig(config)
	if err != nil {
		return nil, errors.New("could not create marshalizer: " + err.Error())
	}

	tr, err := getTrie(config.AccountsTrieStorage, hasher)
	if err != nil {
		return nil, errors.New("error creating trie: " + err.Error())
	}
	uint64ByteSliceConverter := uint64ByteSlice.NewBigEndianConverter()

	return &Core{
		hasher:                   hasher,
		marshalizer:              marshalizer,
		tr:                       tr,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
	}, nil
}

func stateComponentsFactory(config *config.Config, genesisConfig *sharding.Genesis, shardCoordinator sharding.Coordinator, core *Core) (*State, error) {
	addressConverter, err := addressConverters.NewPlainAddressConverter(config.Address.Length, config.Address.Prefix)
	if err != nil {
		return nil, errors.New("could not create address converter: " + err.Error())
	}

	accountFactory, err := factoryState.NewAccountFactoryCreator(shardCoordinator)
	if err != nil {
		return nil, errors.New("could not create account factory: " + err.Error())
	}

	accountsAdapter, err := state.NewAccountsDB(core.tr, core.hasher, core.marshalizer, accountFactory)
	if err != nil {
		return nil, errors.New("could not create accounts adapter: " + err.Error())
	}

	inBalanceForShard, err := genesisConfig.InitialNodesBalances(shardCoordinator, addressConverter)
	if err != nil {
		return nil, errors.New("initial balances could not be processed " + err.Error())
	}

	return &State{
		addressConverter:  addressConverter,
		accountsAdapter:   accountsAdapter,
		inBalanceForShard: inBalanceForShard,
	}, nil
}

func dataComponentsFactory(config *config.Config, shardCoordinator sharding.Coordinator, core *Core) (*Data, error) {
	var datapool dataRetriever.PoolsHolder
	var metaDatapool dataRetriever.MetaPoolsHolder
	blkc, err := createBlockChainFromConfig(config, shardCoordinator)
	if err != nil {
		return nil, errors.New("could not create block chain: " + err.Error())
	}

	store, err := createDataStoreFromConfig(config, shardCoordinator)
	if err != nil {
		return nil, errors.New("could not create local data store: " + err.Error())
	}

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		datapool, err = createShardDataPoolFromConfig(config, core.uint64ByteSliceConverter)
		if err != nil {
			return nil, errors.New("could not create shard data pools: " + err.Error())
		}
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		metaDatapool, err = createMetaDataPoolFromConfig(config, core.uint64ByteSliceConverter)
		if err != nil {
			return nil, errors.New("could not create shard data pools: " + err.Error())
		}
	}
	if datapool == nil && metaDatapool == nil {
		return nil, errors.New("could not create data pools: ")
	}

	return &Data{
		blkc:         blkc,
		store:        store,
		datapool:     datapool,
		metaDatapool: metaDatapool,
	}, nil
}

func cryptoComponentsFactory(
	ctx *cli.Context,
	config *config.Config,
	nodesConfig *sharding.NodesSetup,
	shardCoordinator sharding.Coordinator,
	keyGen crypto.KeyGenerator,
	privKey crypto.PrivateKey,
	log *logger.Logger,
) (*Crypto, error) {
	initialPubKeys := nodesConfig.InitialNodesPubKeys()
	txSingleSigner := &singlesig.SchnorrSigner{}
	singleSigner, err := createSingleSigner(config)
	if err != nil {
		return nil, errors.New("could not create singleSigner: " + err.Error())
	}

	multisigHasher, err := getMultisigHasherFromConfig(config)
	if err != nil {
		return nil, errors.New("could not create multisig hasher: " + err.Error())
	}

	currentShardPubKeys, err := nodesConfig.InitialNodesPubKeysForShard(shardCoordinator.SelfId())
	if err != nil {
		return nil, errors.New("could not start creation of multiSigner: " + err.Error())
	}

	multiSigner, err := createMultiSigner(config, multisigHasher, currentShardPubKeys, privKey, keyGen)
	if err != nil {
		return nil, err
	}

	initialBalancesSkPemFileName := ctx.GlobalString(initialBalancesSkPemFile.Name)
	txSignKeyGen, txSignPrivKey, txSignPubKey, err := getSigningParams(
		ctx,
		log,
		txSignSk.Name,
		txSignSkIndex.Name,
		initialBalancesSkPemFileName,
		kyber.NewBlakeSHA256Ed25519())
	if err != nil {
		return nil, err
	}
	log.Info("Starting with tx sign public key: " + getPkEncoded(txSignPubKey))

	return &Crypto{
		txSingleSigner: txSingleSigner,
		singleSigner:   singleSigner,
		multiSigner:    multiSigner,
		txSignKeyGen:   txSignKeyGen,
		txSignPrivKey:  txSignPrivKey,
		txSignPubKey:   txSignPubKey,
		initialPubKeys: initialPubKeys,
	}, nil
}

func processComponentsFactory(
	genesisConfig *sharding.Genesis,
	nodesConfig *sharding.NodesSetup,
	p2pConfig *config.P2PConfig,
	syncer ntp.SyncTimer,
	shardCoordinator sharding.Coordinator,
	data *Data,
	core *Core,
	crypto *Crypto,
	state *State,
	log *logger.Logger,
) (*Process, error) {

	var randReader io.Reader
	if p2pConfig.Node.Seed != "" {
		randReader = NewSeedRandReader(core.hasher.Compute(p2pConfig.Node.Seed))
	} else {
		randReader = rand.Reader
	}

	netMessenger, err := createNetMessenger(p2pConfig, log, randReader)
	if err != nil {
		return nil, err
	}

	interceptorContainerFactory, resolversContainerFactory, err := getInterceptorAndResolverContainerFactory(
		shardCoordinator, netMessenger, data, core, crypto, state)
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

	resolversFinder, err := containers.NewResolversFinder(resolversContainer, shardCoordinator)
	if err != nil {
		return nil, err
	}

	rounder, err := round.NewRound(
		time.Unix(nodesConfig.StartTime, 0),
		syncer.CurrentTime(),
		time.Millisecond*time.Duration(nodesConfig.RoundDuration),
		syncer)
	if err != nil {
		return nil, err
	}

	forkDetector, err := processSync.NewBasicForkDetector(rounder)
	if err != nil {
		return nil, err
	}

	shardsGenesisBlocks, err := generateGenesisHeadersForInit(
		nodesConfig,
		genesisConfig,
		shardCoordinator,
		state.addressConverter,
		core.hasher,
		core.marshalizer,
	)
	if err != nil {
		return nil, err
	}

	blockProcessor, blockTracker, err := getBlockProcessorAndTracker(resolversFinder, shardCoordinator,
		data, core, state, forkDetector, shardsGenesisBlocks, nodesConfig)
	if err != nil {
		return nil, err
	}

	return &Process{
		interceptorsContainer: interceptorsContainer,
		resolversFinder:       resolversFinder,
		rounder:               rounder,
		forkDetector:          forkDetector,
		blockProcessor:        blockProcessor,
		blockTracker:          blockTracker,
		netMessenger:          netMessenger,
	}, nil
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
		return marshal.JsonMarshalizer{}, nil
	}

	return nil, errors.New("no marshalizer provided in config file")
}

func getTrie(cfg config.StorageConfig, hasher hashing.Hasher) (*trie.Trie, error) {
	accountsTrieStorage, err := storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(cfg.Cache),
		getDBFromConfig(cfg.DB),
		getBloomFromConfig(cfg.Bloom),
	)
	if err != nil {
		return nil, errors.New("error creating accountsTrieStorage: " + err.Error())
	}

	dbWriteCache, err := trie.NewDBWriteCache(accountsTrieStorage)
	if err != nil {
		return nil, errors.New("error creating dbWriteCache: " + err.Error())
	}

	return trie.NewTrie(make([]byte, 32), dbWriteCache, hasher)
}

func createBlockChainFromConfig(config *config.Config, coordinator sharding.Coordinator) (data.ChainHandler, error) {
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
		return blockChain, nil
	}
	if coordinator.SelfId() == sharding.MetachainShardId {
		blockChain, err := blockchain.NewMetaChain(badBlockCache)
		if err != nil {
			return nil, err
		}
		return blockChain, nil
	}
	return nil, errors.New("can not create blockchain")
}

func createDataStoreFromConfig(config *config.Config, shardCoordinator sharding.Coordinator) (dataRetriever.StorageService, error) {
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return createShardDataStoreFromConfig(config)
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		return createMetaChainDataStoreFromConfig(config)
	}
	return nil, errors.New("can not create data store")
}

func createShardDataStoreFromConfig(config *config.Config) (dataRetriever.StorageService, error) {
	var headerUnit, peerBlockUnit, miniBlockUnit, txUnit, metachainHeaderUnit *storageUnit.Unit
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
			if metachainHeaderUnit != nil {
				_ = metachainHeaderUnit.DestroyUnit()
			}
		}
	}()

	txUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.TxStorage.Cache),
		getDBFromConfig(config.TxStorage.DB),
		getBloomFromConfig(config.TxStorage.Bloom))
	if err != nil {
		return nil, err
	}

	miniBlockUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.MiniBlocksStorage.Cache),
		getDBFromConfig(config.MiniBlocksStorage.DB),
		getBloomFromConfig(config.MiniBlocksStorage.Bloom))
	if err != nil {
		return nil, err
	}

	peerBlockUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.PeerBlockBodyStorage.Cache),
		getDBFromConfig(config.PeerBlockBodyStorage.DB),
		getBloomFromConfig(config.PeerBlockBodyStorage.Bloom))
	if err != nil {
		return nil, err
	}

	headerUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.BlockHeaderStorage.Cache),
		getDBFromConfig(config.BlockHeaderStorage.DB),
		getBloomFromConfig(config.BlockHeaderStorage.Bloom))
	if err != nil {
		return nil, err
	}

	metachainHeaderUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.MetaBlockStorage.Cache),
		getDBFromConfig(config.MetaBlockStorage.DB),
		getBloomFromConfig(config.MetaBlockStorage.Bloom))
	if err != nil {
		return nil, err
	}

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, txUnit)
	store.AddStorer(dataRetriever.MiniBlockUnit, miniBlockUnit)
	store.AddStorer(dataRetriever.PeerChangesUnit, peerBlockUnit)
	store.AddStorer(dataRetriever.BlockHeaderUnit, headerUnit)
	store.AddStorer(dataRetriever.MetaBlockUnit, metachainHeaderUnit)

	return store, err
}

func createMetaChainDataStoreFromConfig(config *config.Config) (dataRetriever.StorageService, error) {
	var peerDataUnit, shardDataUnit, metaBlockUnit, headerUnit *storageUnit.Unit
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
		}
	}()

	metaBlockUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.MetaBlockStorage.Cache),
		getDBFromConfig(config.MetaBlockStorage.DB),
		getBloomFromConfig(config.MetaBlockStorage.Bloom))
	if err != nil {
		return nil, err
	}

	shardDataUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.ShardDataStorage.Cache),
		getDBFromConfig(config.ShardDataStorage.DB),
		getBloomFromConfig(config.ShardDataStorage.Bloom))
	if err != nil {
		return nil, err
	}

	peerDataUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.PeerDataStorage.Cache),
		getDBFromConfig(config.PeerDataStorage.DB),
		getBloomFromConfig(config.PeerDataStorage.Bloom))
	if err != nil {
		return nil, err
	}

	headerUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.BlockHeaderStorage.Cache),
		getDBFromConfig(config.BlockHeaderStorage.DB),
		getBloomFromConfig(config.BlockHeaderStorage.Bloom))
	if err != nil {
		return nil, err
	}

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, metaBlockUnit)
	store.AddStorer(dataRetriever.MetaShardDataUnit, shardDataUnit)
	store.AddStorer(dataRetriever.MetaPeerDataUnit, peerDataUnit)
	store.AddStorer(dataRetriever.BlockHeaderUnit, headerUnit)

	return store, err
}

func createShardDataPoolFromConfig(
	config *config.Config,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
) (dataRetriever.PoolsHolder, error) {

	fmt.Println("creatingShardDataPool from config")

	txPool, err := shardedData.NewShardedData(getCacherFromConfig(config.TxDataPool))
	if err != nil {
		fmt.Println("error creating txpool")
		return nil, err
	}

	cacherCfg := getCacherFromConfig(config.BlockHeaderDataPool)
	hdrPool, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating hdrpool")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.MetaBlockBodyDataPool)
	metaBlockBody, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating metaBlockBody")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.BlockHeaderNoncesDataPool)
	hdrNoncesCacher, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating hdrNoncesCacher")
		return nil, err
	}
	hdrNonces, err := dataPool.NewNonceToHashCacher(hdrNoncesCacher, uint64ByteSliceConverter)
	if err != nil {
		fmt.Println("error creating hdrNonces")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.TxBlockBodyDataPool)
	txBlockBody, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating txBlockBody")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.PeerBlockBodyDataPool)
	peerChangeBlockBody, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating peerChangeBlockBody")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.MetaHeaderNoncesDataPool)
	metaBlockNoncesCacher, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating metaBlockNoncesCacher")
		return nil, err
	}
	metaBlockNonces, err := dataPool.NewNonceToHashCacher(metaBlockNoncesCacher, uint64ByteSliceConverter)
	if err != nil {
		fmt.Println("error creating metaBlockNonces")
		return nil, err
	}

	return dataPool.NewShardedDataPool(
		txPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		metaBlockBody,
		metaBlockNonces,
	)
}

func createMetaDataPoolFromConfig(
	config *config.Config,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
) (dataRetriever.MetaPoolsHolder, error) {
	cacherCfg := getCacherFromConfig(config.MetaBlockBodyDataPool)
	metaBlockBody, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating metaBlockBody")
		return nil, err
	}

	miniBlockHashes, err := shardedData.NewShardedData(getCacherFromConfig(config.MiniBlockHeaderHashesDataPool))
	if err != nil {
		fmt.Println("error creating miniBlockHashes")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.ShardHeadersDataPool)
	shardHeaders, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating shardHeaders")
		return nil, err
	}

	shardHeadersNoncesCacher, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating shard headers nonces pool")
		return nil, err
	}
	shardHeadersNonces, err := dataPool.NewNonceToHashCacher(shardHeadersNoncesCacher, uint64ByteSliceConverter)
	if err != nil {
		fmt.Println("error creating shard headers nonces pool")
		return nil, err
	}

	cacherCfg = getCacherFromConfig(config.MetaHeaderNoncesDataPool)
	metaBlockNoncesCacher, err := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	if err != nil {
		fmt.Println("error creating metaBlockNoncesCacher")
		return nil, err
	}
	metaBlockNonces, err := dataPool.NewNonceToHashCacher(metaBlockNoncesCacher, uint64ByteSliceConverter)
	if err != nil {
		fmt.Println("error creating metaBlockNonces")
		return nil, err
	}

	return dataPool.NewMetaDataPool(metaBlockBody, miniBlockHashes, shardHeaders, metaBlockNonces, shardHeadersNonces)
}

func createSingleSigner(config *config.Config) (crypto.SingleSigner, error) {
	switch config.Consensus.Type {
	case blsConsensusType:
		return &singlesig.BlsSingleSigner{}, nil
	case bnConsensusType:
		return &singlesig.SchnorrSigner{}, nil
	}

	return nil, errors.New("no consensus type provided in config file")
}

func getMultisigHasherFromConfig(cfg *config.Config) (hashing.Hasher, error) {
	if cfg.Consensus.Type == blsConsensusType && cfg.MultisigHasher.Type != "blake2b" {
		return nil, errors.New("wrong multisig hasher provided for bls consensus type")
	}

	switch cfg.MultisigHasher.Type {
	case "sha256":
		return sha256.Sha256{}, nil
	case "blake2b":
		if cfg.Consensus.Type == blsConsensusType {
			return blake2b.Blake2b{HashSize: blsHashSize}, nil
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
	case blsConsensusType:
		blsSigner := &blsMultiSig.KyberMultiSignerBLS{}
		return multisig.NewBLSMultisig(blsSigner, hasher, pubKeys, privateKey, keyGen, uint16(0))
	case bnConsensusType:
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
	sk := (*crypto2.Secp256k1PrivateKey)(prvKey)

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

func getInterceptorAndResolverContainerFactory(
	shardCoordinator sharding.Coordinator,
	netMessenger p2p.Messenger,
	data *Data,
	core *Core,
	crypto *Crypto,
	state *State,
) (process.InterceptorsContainerFactory, dataRetriever.ResolversContainerFactory, error) {
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		//TODO add a real chronology validator and remove null chronology validator
		interceptorContainerFactory, err := shard.NewInterceptorsContainerFactory(
			shardCoordinator,
			netMessenger,
			data.store,
			core.marshalizer,
			core.hasher,
			crypto.txSignKeyGen,
			crypto.txSingleSigner,
			crypto.multiSigner,
			data.datapool,
			state.addressConverter,
			&nullChronologyValidator{},
		)
		if err != nil {
			return nil, nil, err
		}

		dataPacker, err := partitioning.NewSizeDataPacker(core.marshalizer)
		if err != nil {
			return nil, nil, err
		}

		resolversContainerFactory, err := shardfactoryDataRetriever.NewResolversContainerFactory(
			shardCoordinator,
			netMessenger,
			data.store,
			core.marshalizer,
			data.datapool,
			core.uint64ByteSliceConverter,
			dataPacker,
		)
		if err != nil {
			return nil, nil, err
		}

		return interceptorContainerFactory, resolversContainerFactory, nil
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		//TODO add a real chronology validator and remove null chronology validator
		interceptorContainerFactory, err := metachain.NewInterceptorsContainerFactory(
			shardCoordinator,
			netMessenger,
			data.store,
			core.marshalizer,
			core.hasher,
			crypto.multiSigner,
			data.metaDatapool,
			&nullChronologyValidator{},
		)
		if err != nil {
			return nil, nil, err
		}
		resolversContainerFactory, err := metafactoryDataRetriever.NewResolversContainerFactory(
			shardCoordinator,
			netMessenger,
			data.store,
			core.marshalizer,
			data.metaDatapool,
			core.uint64ByteSliceConverter,
		)
		if err != nil {
			return nil, nil, err
		}
		return interceptorContainerFactory, resolversContainerFactory, nil
	}
	return nil, nil, errors.New("could not create interceptor and resolver container factory")
}

func generateGenesisHeadersForInit(
	nodesSetup *sharding.NodesSetup,
	genesisConfig *sharding.Genesis,
	shardCoordinator sharding.Coordinator,
	addressConverter state.AddressConverter,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
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
		newShardCoordinator, err := sharding.NewMultiShardCoordinator(shardCoordinator.NumberOfShards(), shardId)
		if err != nil {
			return nil, err
		}

		accountFactory, err := factoryState.NewAccountFactoryCreator(newShardCoordinator)
		if err != nil {
			return nil, err
		}

		accounts := generateInMemoryAccountsAdapter(accountFactory, hasher, marshalizer)
		initialBalances, err := genesisConfig.InitialNodesBalances(newShardCoordinator, addressConverter)
		if err != nil {
			return nil, err
		}

		genesisBlock, err := genesis.CreateShardGenesisBlockFromInitialBalances(
			accounts,
			newShardCoordinator,
			addressConverter,
			initialBalances,
			uint64(nodesSetup.StartTime),
		)
		if err != nil {
			return nil, err
		}

		shardsGenesisBlocks[shardId] = genesisBlock
	}

	if nodesSetup.IsMetaChainActive() {
		genesisBlock, err := genesis.CreateMetaGenesisBlock(uint64(nodesSetup.StartTime), nodesSetup.InitialNodesPubKeys())
		if err != nil {
			return nil, err
		}

		shardsGenesisBlocks[sharding.MetachainShardId] = genesisBlock
	}

	return shardsGenesisBlocks, nil
}

func getBlockProcessorAndTracker(
	resolversFinder dataRetriever.ResolversFinder,
	shardCoordinator sharding.Coordinator,
	data *Data,
	core *Core,
	state *State,
	forkDetector process.ForkDetector,
	shardsGenesisBlocks map[uint32]data.HeaderHandler,
	nodesConfig *sharding.NodesSetup,
) (process.BlockProcessor, process.BlocksTracker, error) {
	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		argsParser, err := smartContract.NewAtArgumentParser()
		if err != nil {
			return nil, nil, err
		}

		vmAccountsDB, err := hooks.NewVMAccountsDB(state.accountsAdapter, state.addressConverter)
		if err != nil {
			return nil, nil, err
		}

		//TODO: change the mock
		scProcessor, err := smartContract.NewSmartContractProcessor(&mock.VMExecutionHandlerStub{}, argsParser,
			core.hasher, core.marshalizer, state.accountsAdapter, vmAccountsDB, state.addressConverter, shardCoordinator)
		if err != nil {
			return nil, nil, err
		}

		requestHandler, err := requestHandlers.NewShardResolverRequestHandler(resolversFinder, factory.TransactionTopic,
			factory.MiniBlocksTopic, factory.MetachainBlocksTopic, maxTxsToRequest)
		if err != nil {
			return nil, nil, err
		}

		transactionProcessor, err := transaction.NewTxProcessor(state.accountsAdapter, core.hasher,
			state.addressConverter, core.marshalizer, shardCoordinator, scProcessor)
		if err != nil {
			return nil, nil, errors.New("could not create transaction processor: " + err.Error())
		}

		blockTracker, err := track.NewShardBlockTracker(data.datapool, core.marshalizer, shardCoordinator, data.store)
		if err != nil {
			return nil, nil, err
		}

		blockProcessor, err := block.NewShardProcessor(
			coreServiceContainer,
			data.datapool,
			data.store,
			core.hasher,
			core.marshalizer,
			transactionProcessor,
			state.accountsAdapter,
			shardCoordinator,
			forkDetector,
			blockTracker,
			shardsGenesisBlocks,
			nodesConfig.MetaChainActive,
			requestHandler,
		)
		if err != nil {
			return nil, nil, errors.New("could not create block processor: " + err.Error())
		}

		return blockProcessor, blockTracker, nil
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		requestHandler, err := requestHandlers.NewMetaResolverRequestHandler(resolversFinder, factory.ShardHeadersForMetachainTopic)
		if err != nil {
			return nil, nil, err
		}

		blockTracker, err := track.NewMetaBlockTracker()
		if err != nil {
			return nil, nil, err
		}

		metaProcessor, err := block.NewMetaProcessor(
			coreServiceContainer,
			state.accountsAdapter,
			data.metaDatapool,
			forkDetector,
			shardCoordinator,
			core.hasher,
			core.marshalizer,
			data.store,
			shardsGenesisBlocks,
			requestHandler,
		)
		if err != nil {
			return nil, nil, errors.New("could not create block processor: " + err.Error())
		}
		return metaProcessor, blockTracker, nil
	}
	return nil, nil, errors.New("could not create block processor and tracker")
}
