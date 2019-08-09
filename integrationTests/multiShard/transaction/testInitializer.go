package transaction

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"

	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	factoryDataRetriever "github.com/ElrondNetwork/elrond-go/dataRetriever/factory/shard"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var hasher = sha256.Sha256{}
var marshalizer = &marshal.JsonMarshalizer{}
var suite = kyber.NewBlakeSHA256Ed25519()
var singleSigner = &singlesig.SchnorrSigner{}
var keyGen = signing.NewKeyGenerator(suite)
var addrConverter, _ = addressConverters.NewPlainAddressConverter(32, "0x")

type testNode struct {
	node         *node.Node
	messenger    p2p.Messenger
	shardId      uint32
	sk           crypto.PrivateKey
	pk           crypto.PublicKey
	dPool        dataRetriever.PoolsHolder
	resFinder    dataRetriever.ResolversFinder
	txRecv       int32
	mutNeededTxs sync.Mutex
	neededTxs    [][]byte
}

func createTestDataPool(txPool dataRetriever.ShardedDataCacherNotifier) dataRetriever.PoolsHolder {
	if txPool == nil {
		txPool, _ = shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache})
	}

	uTxPool, _ := shardedData.NewShardedData(storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache})
	cacherCfg := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache}
	hdrPool, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache}
	hdrNoncesCacher, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)
	hdrNonces, _ := dataPool.NewNonceSyncMapCacher(hdrNoncesCacher, uint64ByteSlice.NewBigEndianConverter())

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache}
	txBlockBody, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 100000, Type: storageUnit.LRUCache}
	peerChangeBlockBody, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

	cacherCfg = storageUnit.CacheConfig{Size: 10, Type: storageUnit.LRUCache}
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

func generateSkPkInShardAndCreateAccount(
	shardCoordinator sharding.Coordinator,
	targetShardId uint32,
	accntAdapter state.AccountsAdapter,
) (crypto.PrivateKey, crypto.PublicKey) {

	sk, pk := keyGen.GeneratePair()
	for {
		pkBytes, _ := pk.ToByteArray()
		addr, _ := addrConverter.CreateAddressFromPublicKeyBytes(pkBytes)
		if shardCoordinator.ComputeId(addr) == targetShardId {
			if accntAdapter != nil {
				_, _ = accntAdapter.GetAccountWithJournal(addr)
				_, _ = accntAdapter.Commit()
			}
			break
		}
		sk, pk = keyGen.GeneratePair()
	}

	return sk, pk
}

func createNetNode(
	dPool dataRetriever.PoolsHolder,
	accntAdapter state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	targetShardId uint32,
	initialAddr string,
) (
	*node.Node,
	p2p.Messenger,
	crypto.PrivateKey,
	dataRetriever.ResolversFinder) {

	messenger := integrationTests.CreateMessengerWithKadDht(context.Background(), initialAddr)
	sk, pk := generateSkPkInShardAndCreateAccount(shardCoordinator, targetShardId, accntAdapter)

	pkBuff, _ := pk.ToByteArray()
	fmt.Printf("Found pk: %s\n", hex.EncodeToString(pkBuff))

	multiSigner := integrationTests.TestMultiSig
	blkc := integrationTests.CreateShardChain()
	store := integrationTests.CreateShardStore(shardCoordinator.NumberOfShards())
	uint64Converter := uint64ByteSlice.NewBigEndianConverter()
	dataPacker, _ := partitioning.NewSizeDataPacker(marshalizer)

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
		node.WithKeyGen(keyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithBlockChain(blkc),
		node.WithUint64ByteSliceConverter(uint64Converter),
		node.WithMultiSigner(multiSigner),
		node.WithSingleSigner(singleSigner),
		node.WithTxSignPrivKey(sk),
		node.WithTxSignPubKey(pk),
		node.WithInterceptorsContainer(interceptorsContainer),
		node.WithResolversFinder(resolversFinder),
		node.WithDataStore(store),
		node.WithTxSingleSigner(singleSigner),
	)

	return n, messenger, sk, resolversFinder
}

func makeDisplayTable(nodes []*testNode) string {
	header := []string{"pk", "shard ID", "tx cache size", "connections"}

	dataLines := make([]*display.LineData, len(nodes))

	for idx, n := range nodes {
		buffPk, _ := n.pk.ToByteArray()

		dataLines[idx] = display.NewLineData(
			false,
			[]string{
				hex.EncodeToString(buffPk),
				fmt.Sprintf("%d", n.shardId),
				fmt.Sprintf("%d", atomic.LoadInt32(&n.txRecv)),
				fmt.Sprintf("%d / %d", len(n.messenger.ConnectedPeersOnTopic(factory.TransactionTopic+"_"+
					fmt.Sprintf("%d", n.shardId))), len(n.messenger.ConnectedPeers())),
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

func createNode(
	shardId uint32,
	numOfShards int,
	dPool dataRetriever.PoolsHolder,
	skShardId uint32,
	serviceID string,
) *testNode {

	testNode := &testNode{
		dPool:   dPool,
		shardId: uint32(shardId),
	}

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), uint32(shardId))
	accntAdapter, _, _ := integrationTests.CreateAccountsDB(shardCoordinator)
	var n *node.Node
	var mes p2p.Messenger
	var sk crypto.PrivateKey
	var resFinder dataRetriever.ResolversFinder

	n, mes, sk, resFinder = createNetNode(
		testNode.dPool,
		accntAdapter,
		shardCoordinator,
		skShardId,
		serviceID,
	)

	testNode.node = n
	testNode.sk = sk
	testNode.messenger = mes
	testNode.pk = sk.GeneratePublic()
	testNode.resFinder = resFinder
	testNode.dPool.Transactions().RegisterHandler(func(key []byte) {
		atomic.AddInt32(&testNode.txRecv, 1)
	})

	return testNode
}

func createNodesWithNodeSkInShardExceptFirst(
	numOfShards int,
	nodesPerShard int,
	firstSkShardId uint32,
	serviceID string,
) []*testNode {

	//first node generated will have its pk belonging to firstSkShardId
	nodes := make([]*testNode, int(numOfShards)*nodesPerShard)

	idx := 0
	for shardId := 0; shardId < numOfShards; shardId++ {
		for j := 0; j < nodesPerShard; j++ {
			dPool := createTestDataPool(nil)

			skShardId := uint32(shardId)
			isFirstNodeGenerated := shardId == 0 && j == 0
			if isFirstNodeGenerated {
				skShardId = firstSkShardId
			}

			testNode := createNode(uint32(shardId), numOfShards, dPool, skShardId, serviceID)

			nodes[idx] = testNode
			idx++
		}
	}
	return nodes
}
