package block

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/multisig"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/display"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/multiShard/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/discovery"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/btcsuite/btcd/btcec"
	crypto2 "github.com/libp2p/go-libp2p-crypto"
)

var r *rand.Rand
var testHasher = sha256.Sha256{}
var testMarshalizer = &marshal.JsonMarshalizer{}

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

type testNode struct {
	node             *node.Node
	mesenger         p2p.Messenger
	shardId          uint32
	sk               crypto.PrivateKey
	pk               crypto.PublicKey
	dPool            data.PoolsHolder
	resFinder        process.ResolversFinder
	headersRecv      int32
	miniblocksRecv   int32
	mutHeaders       sync.Mutex
	headersHashes    [][]byte
	mutMiniblocks    sync.Mutex
	miniblocksHashes [][]byte
	metachainHdrRecv int32
}

func createTestBlockChain() *blockchain.BlockChain {
	cfgCache := storage.CacheConfig{Size: 100, Type: storage.LRUCache}
	badBlockCache, _ := storage.NewCache(cfgCache.Type, cfgCache.Size)
	blockChain, _ := blockchain.NewBlockChain(
		badBlockCache,
		createMemUnit(),
		createMemUnit(),
		createMemUnit(),
		createMemUnit(),
		createMemUnit(),
	)

	return blockChain
}

func createMemUnit() storage.Storer {
	cache, _ := storage.NewCache(storage.LRUCache, 10)
	persist, _ := memorydb.New()

	unit, _ := storage.NewStorageUnit(cache, persist)
	return unit
}

func createTestDataPool() data.PoolsHolder {
	txPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100000, Type: storage.LRUCache})
	hdrPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100000, Type: storage.LRUCache})

	cacherCfg := storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	hdrNoncesCacher, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)
	hdrNonces, _ := dataPool.NewNonceToHashCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	txBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	peerChangeBlockBody, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	cacherCfg = storage.CacheConfig{Size: 100000, Type: storage.LRUCache}
	metaBlocks, _ := storage.NewCache(cacherCfg.Type, cacherCfg.Size)

	dPool, _ := dataPool.NewShardedDataPool(
		txPool,
		hdrPool,
		hdrNonces,
		txBlockBody,
		peerChangeBlockBody,
		metaBlocks,
	)

	return dPool
}

func createAccountsDB() *state.AccountsDB {
	dbw, _ := trie.NewDBWriteCache(createMemUnit())
	tr, _ := trie.NewTrie(make([]byte, 32), dbw, sha256.Sha256{})
	adb, _ := state.NewAccountsDB(tr, sha256.Sha256{}, testMarshalizer)
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
	port int,
	dPool data.PoolsHolder,
	accntAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	targetShardId uint32,
	initialAddr string,
) (
	*node.Node,
	p2p.Messenger,
	crypto.PrivateKey,
	process.ResolversFinder) {

	//messenger := createMessenger(context.Background(), port, serviceID)
	messenger := createMessengerWithKadDht(context.Background(), port, initialAddr)
	addrConverter, _ := state.NewPlainAddressConverter(32, "0x")
	suite := kyber.NewBlakeSHA256Ed25519()
	singleSigner := &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()

	for {
		pkBytes, _ := pk.ToByteArray()
		addr, _ := addrConverter.CreateAddressFromPublicKeyBytes(pkBytes)
		if shardCoordinator.ComputeId(addr) == targetShardId {
			break
		}
		sk, pk = keyGen.GeneratePair()
	}

	pkBuff, _ := pk.ToByteArray()
	fmt.Printf("Found pk: %s\n", hex.EncodeToString(pkBuff))

	multiSigner, _ := createMultiSigner(sk, pk, keyGen, testHasher)
	blkc := createTestBlockChain()
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()

	interceptorContainerFactory, _ := factory.NewInterceptorsContainerFactory(
		shardCoordinator,
		messenger,
		blkc,
		testMarshalizer,
		testHasher,
		keyGen,
		singleSigner,
		multiSigner,
		dPool,
		addrConverter,
	)
	interceptorsContainer, err := interceptorContainerFactory.Create()
	if err != nil {
		fmt.Println(err.Error())
	}

	resolversContainerFactory, _ := factory.NewResolversContainerFactory(
		shardCoordinator,
		messenger,
		blkc,
		testMarshalizer,
		dPool,
		uint64Converter,
	)
	resolversContainer, _ := resolversContainerFactory.Create()
	resolversFinder, _ := containers.NewResolversFinder(resolversContainer, shardCoordinator)

	blockProcessor, _ := block.NewBlockProcessor(
		dPool,
		testHasher,
		testMarshalizer,
		&mock.TxProcessorMock{},
		accntAdapter,
		shardCoordinator,
		&mock.ForkDetectorMock{},
		func(destShardID uint32, txHash []byte) {},
	)

	n, err := node.NewNode(
		node.WithMessenger(messenger),
		node.WithMarshalizer(testMarshalizer),
		node.WithHasher(testHasher),
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
		node.WithBlockProcessor(blockProcessor),
	)

	if err != nil {
		fmt.Println(err.Error())
	}

	return n, messenger, sk, resolversFinder
}

func createMessengerWithKadDht(ctx context.Context, port int, initialAddr string) p2p.Messenger {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*crypto2.Secp256k1PrivateKey)(prvKey)

	libP2PMes, err := libp2p.NewNetworkMessenger(
		ctx,
		port,
		sk,
		nil,
		loadBalancer.NewOutgoingChannelLoadBalancer(),
		discovery.NewKadDhtPeerDiscoverer(time.Second, "test", []string{initialAddr}),
	)
	if err != nil {
		fmt.Println(err.Error())
	}

	return libP2PMes
}

func getConnectableAddress(mes p2p.Messenger) string {
	for _, addr := range mes.Addresses() {
		if strings.Contains(addr, "circuit") {
			continue
		}
		return addr
	}
	return ""
}

func makeDisplayTable(nodes []*testNode) string {
	header := []string{"pk", "shard ID", "headers", "miniblocks", "metachain headers", "connections"}
	dataLines := make([]*display.LineData, len(nodes))
	for idx, n := range nodes {
		buffPk, _ := n.pk.ToByteArray()

		dataLines[idx] = display.NewLineData(
			false,
			[]string{
				hex.EncodeToString(buffPk),
				fmt.Sprintf("%d", n.shardId),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.headersRecv)),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.miniblocksRecv)),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.metachainHdrRecv)),
				fmt.Sprintf("%d / %d", len(n.mesenger.ConnectedPeersOnTopic(factory.TransactionTopic+"_"+
					fmt.Sprintf("%d", n.shardId))), len(n.mesenger.ConnectedPeers())),
			},
		)
	}
	table, _ := display.CreateTableString(header, dataLines)
	return table
}

func displayAndStartNodes(nodes []*testNode) {
	for _, n := range nodes {
		skBuff, _ := n.sk.ToByteArray()
		pkBuff, _ := n.pk.ToByteArray()

		fmt.Printf("Shard ID: %v, sk: %s, pk: %s\n",
			n.shardId,
			hex.EncodeToString(skBuff),
			hex.EncodeToString(pkBuff),
		)
		_ = n.node.Start()
		_ = n.node.P2PBootstrap()
	}
}

func createNodes(
	startingPort int,
	numOfShards int,
	nodesPerShard int,
	serviceID string,
) []*testNode {

	//first node generated will have is pk belonging to firstSkShardId
	nodes := make([]*testNode, int(numOfShards)*nodesPerShard)

	idx := 0
	for shardId := 0; shardId < numOfShards; shardId++ {
		for j := 0; j < nodesPerShard; j++ {
			testNode := &testNode{
				dPool:   createTestDataPool(),
				shardId: uint32(shardId),
			}

			shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), uint32(shardId))
			accntAdapter := createAccountsDB()
			n, mes, sk, resFinder := createNetNode(
				startingPort+idx,
				testNode.dPool,
				accntAdapter,
				shardCoordinator,
				testNode.shardId,
				serviceID,
			)

			testNode.node = n
			testNode.sk = sk
			testNode.mesenger = mes
			testNode.pk = sk.GeneratePublic()
			testNode.resFinder = resFinder
			testNode.dPool.Headers().RegisterHandler(func(key []byte) {
				atomic.AddInt32(&testNode.headersRecv, 1)
				testNode.mutHeaders.Lock()
				testNode.headersHashes = append(testNode.headersHashes, key)
				testNode.mutHeaders.Unlock()
			})
			testNode.dPool.MiniBlocks().RegisterHandler(func(key []byte) {
				atomic.AddInt32(&testNode.miniblocksRecv, 1)
				testNode.mutMiniblocks.Lock()
				testNode.miniblocksHashes = append(testNode.miniblocksHashes, key)
				testNode.mutMiniblocks.Unlock()
			})
			testNode.dPool.MetaBlocks().RegisterHandler(func(key []byte) {
				atomic.AddInt32(&testNode.metachainHdrRecv, 1)
			})

			nodes[idx] = testNode
			idx++
		}
	}

	return nodes
}

func getMiniBlocksHashesFromShardIds(body dataBlock.Body, shardIds ...uint32) [][]byte {
	hashes := make([][]byte, 0)

	for _, miniblock := range body {
		for _, shardId := range shardIds {
			if miniblock.ReceiverShardID == shardId {
				buff, _ := testMarshalizer.Marshal(miniblock)
				hashes = append(hashes, testHasher.Compute(string(buff)))
			}
		}
	}

	return hashes
}

func equalSlices(slice1 [][]byte, slice2 [][]byte) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	//check slice1 has all elements in slice2
	for _, buff1 := range slice1 {
		found := false
		for _, buff2 := range slice2 {
			if bytes.Equal(buff1, buff2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	//check slice2 has all elements in slice1
	for _, buff2 := range slice2 {
		found := false
		for _, buff1 := range slice1 {
			if bytes.Equal(buff1, buff2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func uint32InSlice(searched uint32, list []uint32) bool {
	for _, val := range list {
		if val == searched {
			return true
		}
	}
	return false
}
