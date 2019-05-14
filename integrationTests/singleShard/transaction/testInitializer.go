package transaction

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/factory/containers"
	factoryDataRetriever "github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/factory/shard"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/btcsuite/btcd/btcec"
	crypto2 "github.com/libp2p/go-libp2p-crypto"
)

var r *rand.Rand

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func createTestBlockChain() data.ChainHandler {
	cfgCache := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	badBlockCache, _ := storage.NewCache(cfgCache.Type, cfgCache.Size)
	blockChain, _ := blockchain.NewBlockChain(
		badBlockCache,
	)

	return blockChain
}

func createMemUnit() storage.Storer {
	cache, _ := storage.NewCache(storage.LRUCache, 10)
	persist, _ := memorydb.New()
	unit, _ := storage.NewStorageUnit(cache, persist)

	return unit
}

func createTestStore() dataRetriever.StorageService {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, createMemUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, createMemUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, createMemUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, createMemUnit())

	return store
}

func createTestDataPool() dataRetriever.PoolsHolder {
	txPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100000, Type: storage.LRUCache})
	cacherCfg := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	hdrPool, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	hdrNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)
	hdrNonces, _ := dataPool.NewNonceToHashCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	txBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	peerChangeBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	metaHdrNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)
	metaHdrNonces, _ := dataPool.NewNonceToHashCacher(metaHdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())
	metaBlocks, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 10, Type: storage.LRUCache}

	dPool, _ := dataPool.NewShardedDataPool(
		txPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		metaBlocks,
		metaHdrNonces,
	)

	return dPool
}

func createDummyHexAddress(chars int) string {
	if chars < 1 {
		return ""
	}

	var characters = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

	buff := make([]byte, chars)
	for i := 0; i < chars; i++ {
		buff[i] = characters[r.Int()%16]
	}

	return string(buff)
}

func createAccountsDB() *state.AccountsDB {
	marsh := &marshal.JsonMarshalizer{}

	dbw, _ := trie.NewDBWriteCache(createMemUnit())
	tr, _ := trie.NewTrie(make([]byte, 32), dbw, sha256.Sha256{})
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, marsh, &mock.AccountsFactoryStub{
		CreateAccountCalled: func(address state.AddressContainer, tracker state.AccountTracker) (wrapper state.AccountHandler, e error) {
			return state.NewAccount(address, tracker)
		},
	})

	return adb
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
	multiSigner, err := multisig.NewBelNevMultisig(hasher, publicKeys, privateKey, keyGen, 0)

	return multiSigner, err
}

func createNetNode(
	dPool dataRetriever.PoolsHolder,
	accntAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
) (
	*node.Node,
	p2p.Messenger,
	crypto.PrivateKey,
	dataRetriever.ResolversFinder) {

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	messenger := createMessenger(context.Background())

	addrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")

	suite := kyber.NewBlakeSHA256Ed25519()
	singleSigner := &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()
	multiSigner, _ := createMultiSigner(sk, pk, keyGen, hasher)
	blkc := createTestBlockChain()
	store := createTestStore()
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	interceptorContainerFactory, _ := shard.NewInterceptorsContainerFactory(
		shardCoordinator,
		messenger,
		store,
		marshalizer,
		hasher,
		keyGen,
		singleSigner,
		multiSigner,
		dPool,
		addrConverter,
		&mock.ChronologyValidatorMock{},
		nil,
	)
	interceptorsContainer, _ := interceptorContainerFactory.Create()

	resolversContainerFactory, _ := factoryDataRetriever.NewResolversContainerFactory(
		shardCoordinator,
		messenger,
		store,
		marshalizer,
		dPool,
		uint64Converter,
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
		node.WithKeyGenerator(keyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithBlockChain(blkc),
		node.WithUint64ByteSliceConverter(uint64Converter),
		node.WithMultisig(multiSigner),
		node.WithSinglesig(singleSigner),
		node.WithPrivateKey(sk),
		node.WithPublicKey(pk),
		node.WithInterceptorsContainer(interceptorsContainer),
		node.WithResolversFinder(resolversFinder),
		node.WithDataStore(store),
	)

	return n, messenger, sk, resolversFinder
}

func createMessenger(ctx context.Context) p2p.Messenger {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*crypto2.Secp256k1PrivateKey)(prvKey)

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
