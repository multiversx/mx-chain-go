package main

import (
	"errors"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/config"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
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
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/storageUnit"
	"github.com/urfave/cli"
)

type Core struct {
	hasher                   hashing.Hasher
	marshalizer              marshal.Marshalizer
	tr                       trie.PatriciaMerkelTree
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

type State struct {
	addressConverter state.AddressConverter
	accountsAdapter  state.AccountsAdapter
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
}

//TODO: refactor in the next phase
//type Process struct {
//	interceptorsContainer process.InterceptorsContainer
//	resolversFinder       dataRetriever.ResolversFinder
//	rounder               consensus.Rounder
//	forkDetector          process.ForkDetector
//	blockProcessor        process.BlockProcessor
//	blockTracker          process.BlocksTracker
//}

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

func stateComponentsFactory(config *config.Config, shardCoordinator sharding.Coordinator, core *Core) (*State, error) {
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

	return &State{
		addressConverter: addressConverter,
		accountsAdapter:  accountsAdapter,
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

func cryptoComponentsFactory(ctx *cli.Context, config *config.Config, nodesConfig *sharding.NodesSetup, shardCoordinator sharding.Coordinator,
	keyGen crypto.KeyGenerator, privKey crypto.PrivateKey, log *logger.Logger,
) (*Crypto, error) {
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

	return &Crypto{
		txSingleSigner: txSingleSigner,
		singleSigner:   singleSigner,
		multiSigner:    multiSigner,
		txSignKeyGen:   txSignKeyGen,
		txSignPrivKey:  txSignPrivKey,
		txSignPubKey:   txSignPubKey,
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
		return createShardDataStoreFromConfig(config, shardCoordinator)
	}
	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		return createMetaChainDataStoreFromConfig(config, shardCoordinator)
	}
	return nil, errors.New("can not create data store")
}

func createShardDataStoreFromConfig(config *config.Config, shardCoordinator sharding.Coordinator) (dataRetriever.StorageService, error) {
	var headerUnit, peerBlockUnit, miniBlockUnit, txUnit, metachainHeaderUnit, metaHdrHashNonceUnit, shardHdrHashNonceUnit *storageUnit.Unit
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

	metaHdrHashNonceUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.MetaHdrNonceHashStorage.Cache),
		getDBFromConfig(config.MetaHdrNonceHashStorage.DB),
		getBloomFromConfig(config.MetaHdrNonceHashStorage.Bloom),
	)
	if err != nil {
		return nil, err
	}

	shardHdrHashNonceUnit, err = storageUnit.NewShardedStorageUnitFromConf(
		getCacherFromConfig(config.ShardHdrNonceHashStorage.Cache),
		getDBFromConfig(config.ShardHdrNonceHashStorage.DB),
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
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, metaHdrHashNonceUnit)
	store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(shardCoordinator.SelfId()), shardHdrHashNonceUnit)

	return store, err
}

func createMetaChainDataStoreFromConfig(config *config.Config, shardCoordinator sharding.Coordinator) (dataRetriever.StorageService, error) {
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

	metaHdrHashNonceUnit, err = storageUnit.NewStorageUnitFromConf(
		getCacherFromConfig(config.MetaHdrNonceHashStorage.Cache),
		getDBFromConfig(config.MetaHdrNonceHashStorage.DB),
		getBloomFromConfig(config.MetaHdrNonceHashStorage.Bloom),
	)
	if err != nil {
		return nil, err
	}

	shardHdrHashNonceUnits = make([]*storageUnit.Unit, shardCoordinator.NumberOfShards())
	for i := uint32(0); i < shardCoordinator.NumberOfShards(); i++ {
		shardHdrHashNonceUnits[i], err = storageUnit.NewShardedStorageUnitFromConf(
			getCacherFromConfig(config.ShardHdrNonceHashStorage.Cache),
			getDBFromConfig(config.ShardHdrNonceHashStorage.DB),
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
		store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit+dataRetriever.UnitType(i), shardHdrHashNonceUnits[i])
	}

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
