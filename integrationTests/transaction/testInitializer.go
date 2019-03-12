package transaction

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2/multisig"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2/singlesig"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/btcsuite/btcd/btcec"
	crypto2 "github.com/libp2p/go-libp2p-crypto"
)

var r io.Reader

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

type testInitializer struct {
}

func (ti *testInitializer) createTestBlockChain() *blockchain.BlockChain {

	cfgCache := storage.CacheConfig{Size: 100, Type: storage.LRUCache}

	badBlockCache, _ := storage.NewCache(cfgCache.Type, cfgCache.Size)

	blockChain, _ := blockchain.NewBlockChain(
		badBlockCache,
		ti.createMemUnit(),
		ti.createMemUnit(),
		ti.createMemUnit(),
		ti.createMemUnit())

	return blockChain
}

func (ti *testInitializer) createMemUnit() storage.Storer {
	cache, _ := storage.NewCache(storage.LRUCache, 10)
	persist, _ := memorydb.New()

	unit, _ := storage.NewStorageUnit(cache, persist)
	return unit
}

func (ti *testInitializer) createTestDataPool() data.TransientDataHolder {
	txPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100000, Type: storage.LRUCache})
	hdrPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100000, Type: storage.LRUCache})

	cacherCfg := storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	hdrNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)
	hdrNonces, _ := dataPool.NewNonceToHashCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	txBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	peerChangeBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	dPool, _ := dataPool.NewDataPool(
		txPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
	)

	return dPool
}

func (ti *testInitializer) createDummyHexAddress(chars int) string {
	if chars < 1 {
		return ""
	}

	var characters = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

	rdm := rand.New(rand.NewSource(time.Now().UnixNano()))

	buff := make([]byte, chars)
	for i := 0; i < chars; i++ {
		buff[i] = characters[rdm.Int()%16]
	}

	return string(buff)
}

func (ti *testInitializer) createAccountsDB() *state.AccountsDB {
	marsh := &marshal.JsonMarshalizer{}

	dbw, _ := trie.NewDBWriteCache(ti.createMemUnit())
	tr, _ := trie.NewTrie(make([]byte, 32), dbw, sha256.Sha256{})
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, marsh)

	return adb
}

func (ti *testInitializer) createMultiSigner(
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

func (ti *testInitializer) createNetNode(port int, dPool data.TransientDataHolder, accntAdapter state.AccountsAdapter) (
	*node.Node,
	p2p.Messenger,
	crypto.PrivateKey,
	process.InterceptorsResolversFactory) {

	hasher := sha256.Sha256{}
	marshalizer := &marshal.JsonMarshalizer{}

	messenger := ti.createMessenger(context.Background(), port)

	addrConverter, _ := state.NewPlainAddressConverter(32, "0x")

	suite := kv2.NewBlakeSHA256Ed25519()
	singleSigner := &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()
	multiSigner, _ := ti.createMultiSigner(sk, pk, keyGen, hasher)
	blkc := ti.createTestBlockChain()
	shardCoordinator := &sharding.OneShardCoordinator{}
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	pFactory, _ := factory.NewInterceptorsResolversCreator(factory.InterceptorsResolversConfig{
		ResolverContainer:        containers.NewResolversContainer(),
		Messenger:                messenger,
		Blockchain:               blkc,
		DataPool:                 dPool,
		ShardCoordinator:         shardCoordinator,
		AddrConverter:            addrConverter,
		Hasher:                   hasher,
		Marshalizer:              marshalizer,
		MultiSigner:              multiSigner,
		SingleSigner:             singleSigner,
		KeyGen:                   keyGen,
		Uint64ByteSliceConverter: uint64Converter,
	})

	interceptorContainerFactory, _ := factory.NewInterceptorsContainerFactory(
		shardCoordinator,
		messenger,
		blkc,
		marshalizer,
		hasher,
		keyGen,
		singleSigner,
		multiSigner,
		dPool,
		addrConverter,
	)
	interceptorsContainer, _ := interceptorContainerFactory.Create()

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
		node.WithInterceptorsResolversFactory(pFactory),
		node.WithInterceptorsContainer(interceptorsContainer),
	)

	_ = pFactory.CreateResolvers()

	return n, messenger, sk, pFactory
}

func (ti *testInitializer) createMessenger(ctx context.Context, port int) p2p.Messenger {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*crypto2.Secp256k1PrivateKey)(prvKey)

	libP2PMes, err := libp2p.NewNetworkMessenger(
		ctx,
		port,
		sk,
		nil,
		loadBalancer.NewOutgoingChannelLoadBalancer(),
		discovery.NewNullDiscoverer())

	if err != nil {
		fmt.Println(err.Error())
	}

	return libP2PMes
}

func (ti *testInitializer) getConnectableAddress(mes p2p.Messenger) string {
	for _, addr := range mes.Addresses() {
		if strings.Contains(addr, "circuit") {
			continue
		}

		return addr
	}

	return ""
}
