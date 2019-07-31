package block

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	llsig "github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/multisig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	factoryDataRetriever "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/shard"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"math/big"
)

var r io.Reader

type keyPair struct {
	sk crypto.PrivateKey
	pk crypto.PublicKey
}

type cryptoParams struct {
	keyGen       crypto.KeyGenerator
	keys         map[uint32][]*keyPair
	singleSigner crypto.SingleSigner
}

func genValidatorsFromPubKeys(pubKeysMap map[uint32][]string) map[uint32][]sharding.Validator {
	validatorsMap := make(map[uint32][]sharding.Validator)

	for shardId, shardNodesPks := range pubKeysMap {
		shardValidators := make([]sharding.Validator, 0)
		for i := 0; i < len(shardNodesPks); i++ {
			v, _ := sharding.NewValidator(big.NewInt(0), 1, []byte(shardNodesPks[i]))
			shardValidators = append(shardValidators, v)
		}
		validatorsMap[shardId] = shardValidators
	}

	return validatorsMap
}

func createCryptoParams(nodesPerShard int, nbMetaNodes int, nbShards int) *cryptoParams {
	suite := kyber.NewSuitePairingBn256()
	singleSigner := &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)

	keysMap := make(map[uint32][]*keyPair)
	keyPairs := make([]*keyPair, nodesPerShard)
	for shardId := 0; shardId < nbShards; shardId++ {
		for n := 0; n < nodesPerShard; n++ {
			kp := &keyPair{}
			kp.sk, kp.pk = keyGen.GeneratePair()
			keyPairs[n] = kp
		}
		keysMap[uint32(shardId)] = keyPairs
	}

	keyPairs = make([]*keyPair, nbMetaNodes)
	for n := 0; n < nbMetaNodes; n++ {
		kp := &keyPair{}
		kp.sk, kp.pk = keyGen.GeneratePair()
		keyPairs[n] = kp
	}
	keysMap[sharding.MetachainShardId] = keyPairs

	params := &cryptoParams{
		keys:         keysMap,
		keyGen:       keyGen,
		singleSigner: singleSigner,
	}

	return params
}

func pubKeysMapFromKeysMap(keyPairMap map[uint32][]*keyPair) map[uint32][]string {
	keysMap := make(map[uint32][]string, 0)

	for shardId, pairList := range keyPairMap {
		shardKeys := make([]string, len(pairList))
		for i, pair := range pairList {
			bytes, _ := pair.pk.ToByteArray()
			shardKeys[i] = string(bytes)
		}
		keysMap[shardId] = shardKeys
	}

	return keysMap
}

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func createTestBlockChain() data.ChainHandler {
	cfgCache := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	badBlockCache, _ := storageUnit.NewCache(cfgCache.Type, cfgCache.Size, cfgCache.Shards)

	blockChain, _ := blockchain.NewBlockChain(
		badBlockCache,
	)

	return blockChain
}

func createTestStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, createMemUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, createMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, createMemUnit())
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, createMemUnit())
	store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, createMemUnit())

	return store
}

func createMemUnit() storage.Storer {
	cache, _ := storageUnit.NewCache(storageUnit.LRUCache, 10, 1)
	persist, _ := memorydb.New()

	unit, _ := storageUnit.NewStorageUnit(cache, persist)
	return unit
}

func createTestDataPool() dataRetriever.PoolsHolder {
	txPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache})
	uTxPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache})
	cacherCfg := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	hdrPool, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	hdrNoncesCacher, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	hdrNonces, _ := dataPool.NewNonceSyncMapCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	txBlockBody, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	peerChangeBlockBody, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache}
	metaBlocks, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	dPool, _ := dataPool.NewShardedDataPool(
		txPool,
		uTxPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		metaBlocks,
	)

	return dPool
}

func createMultiSigner(
	privateKey crypto.PrivateKey,
	publicKey crypto.PublicKey,
	keyGen crypto.KeyGenerator,
	hasher hashing.Hasher,
) (crypto.MultiSigner, error) {

	publicKeys := make([]string, 1)
	pubKey, _ := publicKey.ToByteArray()
	publicKeys[0] = string(pubKey)
	llsigner := &llsig.KyberMultiSignerBLS{}
	multiSigner, err := multisig.NewBLSMultisig(llsigner, hasher, publicKeys, privateKey, keyGen, 0)

	return multiSigner, err
}

func createAccountsDB() *state.AccountsDB {
	marsh := &marshal.JsonMarshalizer{}
	hasher := sha256.Sha256{}
	store := createMemUnit()

	tr, _ := trie.NewTrie(store, marsh, hasher)
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, marsh, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address state.AddressContainer, tracker state.AccountTracker) (wrapper state.AccountHandler, e error) {
			return state.NewAccount(address, tracker)
		},
	})

	return adb
}

func createNetNode(
	dPool dataRetriever.PoolsHolder,
	store dataRetriever.StorageService,
	accntAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
	params *cryptoParams,
	keysIndex int,
) (
	*node.Node,
	p2p.Messenger,
	crypto.MultiSigner,
	dataRetriever.ResolversFinder) {

	hasher := sha256.Sha256{}
	hasherSigning := &mock.HasherSpongeMock{}
	marshalizer := &marshal.JsonMarshalizer{}

	messenger := createMessenger(context.Background())

	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")
	keyPair := params.keys[shardCoordinator.SelfId()][keysIndex]
	multiSigner, _ := createMultiSigner(keyPair.sk, keyPair.pk, params.keyGen, hasherSigning)
	blkc := createTestBlockChain()
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	dataPacker, _ := partitioning.NewSizeDataPacker(marshalizer)

	interceptorContainerFactory, _ := shard.NewInterceptorsContainerFactory(
		shardCoordinator,
		nodesCoordinator,
		messenger,
		store,
		marshalizer,
		hasher,
		params.keyGen,
		params.singleSigner,
		multiSigner,
		dPool,
		addrConverter,
	)
	interceptorsContainer, _ := interceptorContainerFactory.Create()

	resolversContainerFactory, _ := factoryDataRetriever.NewResolversContainerFactory(
		shardCoordinator,
		messenger,
		store,
		marshalizer,
		dPool,
		uint64Converter,
		dataPacker,
	)
	resolversContainer, _ := resolversContainerFactory.Create()
	resolversFinder, _ := containers.NewResolversFinder(resolversContainer, shardCoordinator)

	n, _ := node.NewNode(
		node.WithMessenger(messenger),
		node.WithMarshalizer(marshalizer),
		node.WithHasher(hasher),
		node.WithDataPool(dPool),
		node.WithAddressConverter(addrConverter),
		node.WithAccountsAdapter(accntAdapter),
		node.WithSingleSigner(params.singleSigner),
		node.WithMultiSigner(multiSigner),
		node.WithKeyGen(params.keyGen),
		node.WithTxSignPrivKey(keyPair.sk),
		node.WithTxSignPubKey(keyPair.pk),
		node.WithShardCoordinator(shardCoordinator),
		node.WithBlockChain(blkc),
		node.WithUint64ByteSliceConverter(uint64Converter),
		node.WithInterceptorsContainer(interceptorsContainer),
		node.WithResolversFinder(resolversFinder),
		node.WithDataStore(store),
	)

	return n, messenger, multiSigner, resolversFinder
}

func createMessenger(ctx context.Context) p2p.Messenger {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*libp2pCrypto.Secp256k1PrivateKey)(prvKey)

	libP2PMes, err := libp2p.NewNetworkMessengerOnFreePort(
		ctx,
		sk,
		nil,
		loadBalancer.NewOutgoingChannelLoadBalancer(),
		discovery.NewNullDiscoverer())

	if err != nil {
		fmt.Println(err.Error())
	}

	return libP2PMes
}

func getConnectableAddress(mes p2p.Messenger) string {
	for _, addr := range mes.Addresses() {
		if strings.Contains(addr, "circuit") || strings.Contains(addr, "169.254") {
			continue
		}

		return addr
	}

	return ""
}
