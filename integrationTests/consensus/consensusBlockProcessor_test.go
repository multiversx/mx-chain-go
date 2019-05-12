package consensus

import (
	"context"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/consensus/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	syncFork "github.com/ElrondNetwork/elrond-go-sandbox/process/sync"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	beevikntp "github.com/beevik/ntp"
	"github.com/stretchr/testify/assert"
)

func mockRequestFunction(shardId uint32, txHash []byte) {
	return
}

func createNodeConsensusWithBlockProcessor(
	shardCoordinator sharding.Coordinator,
	shardId uint32,
	selfId uint32,
	initialAddr string,
	consensusSize uint32,
	roundTime uint64,
	privKey crypto.PrivateKey,
	pubKeys []crypto.PublicKey,
	testKeyGen crypto.KeyGenerator,
) (
	*node.Node,
	p2p.Messenger,
	data.ChainHandler) {

	messenger := createMessengerWithKadDht(context.Background(), initialAddr)

	blockChain := createTestBlockChain()
	blockChainMock := &mock.BlockChainMock{BlockChain: blockChain}

	startTime := int64(0)

	singlesigner := &singlesig.SchnorrSigner{}
	singleBlsSigner := &singlesig.BlsSingleSigner{}

	syncer := ntp.NewSyncTime(time.Hour, beevikntp.Query)
	go syncer.StartSync()

	rounder, err := round.NewRound(
		time.Unix(startTime, 0),
		syncer.CurrentTime(),
		time.Millisecond*time.Duration(uint64(roundTime)),
		syncer)

	forkDetector, _ := syncFork.NewBasicForkDetector(rounder)

	hdrResolver := &mock.HeaderResolverMock{}
	mbResolver := &mock.MiniBlocksResolverMock{}
	resolverFinder := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
			if baseTopic == factory.HeadersTopic {
				return hdrResolver, nil
			}
			if baseTopic == factory.MiniBlocksTopic {
				return mbResolver, nil
			}
			return hdrResolver, nil
		},
	}

	inPubKeys := make(map[uint32][]string)
	for _, val := range pubKeys {
		sPubKey, _ := val.ToByteArray()
		inPubKeys[shardId] = append(inPubKeys[shardId], string(sPubKey))
	}

	testHasher := sha256.Sha256{}
	testMarshalizer := &marshal.JsonMarshalizer{}
	testAddressConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")

	accntAdapter := createAccountsDB(testMarshalizer)

	testMultiSig := mock.NewMultiSigner(uint32(consensusSize))
	_ = testMultiSig.Reset(inPubKeys[shardId], uint16(selfId))

	dataPool := createTestShardDataPool()
	metaDataPool := createTestMetaDataPool()
	store := createTestStore()

	var blockProcessor process.BlockProcessor
	if shardId < shardCoordinator.NumberOfShards() {
		blockProcessor, _ = block.NewShardProcessor(dataPool, store, testHasher, testMarshalizer, &mock.TxProcessorMock{},
			accntAdapter, shardCoordinator, forkDetector, mockRequestFunction,
			mockRequestFunction)
	}

	if shardId == sharding.MetachainShardId {
		blockProcessor, _ = block.NewMetaProcessor(accntAdapter, metaDataPool, forkDetector, shardCoordinator,
			testHasher, testMarshalizer, store, mockRequestFunction)
	}

	n, err := node.NewNode(
		node.WithInitialNodesPubKeys(inPubKeys),
		node.WithRoundDuration(uint64(roundTime)),
		node.WithConsensusGroupSize(int(consensusSize)),
		node.WithSyncer(syncer),
		node.WithGenesisTime(time.Unix(startTime, 0)),
		node.WithRounder(rounder),
		node.WithBlsSinglesig(singleBlsSigner),
		node.WithBlsPrivateKey(privKey),
		node.WithForkDetector(forkDetector),
		node.WithMessenger(messenger),
		node.WithMarshalizer(testMarshalizer),
		node.WithHasher(testHasher),
		node.WithAddressConverter(testAddressConverter),
		node.WithAccountsAdapter(accntAdapter),
		node.WithKeyGenerator(testKeyGen),
		node.WithShardCoordinator(shardCoordinator),
		node.WithBlockChain(blockChainMock),
		node.WithMultisig(testMultiSig),
		node.WithSinglesig(singlesigner),
		node.WithPrivateKey(privKey),
		node.WithPublicKey(privKey.GeneratePublic()),
		node.WithBlockProcessor(blockProcessor),
		node.WithDataPool(dataPool),
		node.WithMetaDataPool(metaDataPool),
		node.WithDataStore(store),
		node.WithResolversFinder(resolverFinder),
	)

	if err != nil {
		fmt.Println(err.Error())
	}

	return n, messenger, blockChainMock
}

func createNodesConsensusWithBlockProcessor(
	nodesPerShard int,
	consensusSize int,
	roundTime uint64,
	serviceID string,
	shardId uint32,
) []*testNode {

	privKeys, pubKeys, testKeyGen := initialPrivPubKeys(nodesPerShard)
	//first node generated will have is pk belonging to firstSkShardId
	nodes := make([]*testNode, nodesPerShard)

	for i := 0; i < nodesPerShard; i++ {
		testNode := &testNode{
			shardId: uint32(shardId),
		}

		shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(1), uint32(0))
		n, mes, blkc := createNodeConsensusWithBlockProcessor(
			shardCoordinator,
			testNode.shardId,
			uint32(i),
			serviceID,
			uint32(consensusSize),
			roundTime,
			privKeys[i],
			pubKeys,
			testKeyGen,
		)

		testNode.node = n
		testNode.node = n
		testNode.sk = privKeys[i]
		testNode.mesenger = mes
		testNode.pk = pubKeys[i]
		testNode.blkc = blkc

		nodes[i] = testNode
	}

	return nodes
}

func initNodesWithBlockProcessor(numNodes, consensusSize, numInvalid uint32, roundTime uint64, shardId uint32) ([]*testNode, p2p.Messenger) {
	fmt.Println("Step 1. Setup nodes...")

	advertiser := createMessengerWithKadDht(context.Background(), "")
	advertiser.Bootstrap()

	nodes := createNodesConsensusWithBlockProcessor(
		int(numNodes),
		int(consensusSize),
		roundTime,
		getConnectableAddress(advertiser),
		shardId,
	)
	displayAndStartNodes(nodes)

	if numInvalid < numNodes {
		for i := uint32(0); i < numInvalid; i++ {
			mockBlockChain, _ := nodes[i].blkc.(*mock.BlockChainMock)

			mockBlockChain.GetCurrentBlockBodyCalled = func() data.BodyHandler {
				return nil
			}
			mockBlockChain.GetGenesisHeaderHashCalled = func() []byte {
				return []byte("wrong genesis header hash")
			}
		}
	}

	return nodes, advertiser
}

func TestConsensusBlockProcessorFullConsensus(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodes := uint32(21)
	consensusSize := uint32(21)
	numInvalid := uint32(0)
	roundTime := uint64(4000)
	shardId := uint32(0)
	numCommBlock := uint32(10)
	nodes, advertiser := initNodesWithBlockProcessor(numNodes, consensusSize, numInvalid, roundTime, shardId)

	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second * 1)

	mutex := &sync.Mutex{}
	combinedMap := make(map[uint64]uint32, 0)
	minNonce := ^uint64(0)
	maxNonce := uint64(0)
	for _, n := range nodes {
		mockBlockChain, _ := n.blkc.(*mock.BlockChainMock)

		mockBlockChain.SetCurrentBlockHeaderCalled = func(handler data.HeaderHandler) error {
			mutex.Lock()

			err := mockBlockChain.BlockChain.SetCurrentBlockHeader(handler)

			combinedMap[handler.GetNonce()] = handler.GetRound()
			if minNonce > handler.GetNonce() {
				minNonce = handler.GetNonce()
			}
			if maxNonce < handler.GetNonce() {
				maxNonce = handler.GetNonce()
			}

			mutex.Unlock()

			return err
		}

		_ = n.node.StartConsensus()
	}

	time.Sleep(time.Second * 20)
	chDone := make(chan bool, 0)
	go checkBlockProposedForEachNonce(numCommBlock, combinedMap, &minNonce, &maxNonce, mutex, chDone, t)

	select {
	case <-chDone:
	case <-time.After(180 * time.Second):
		assert.Fail(t, "consensus too slow, not working %d %d")
		return
	}
}

func TestConsensusBlockProcessorNotEnoughValidators(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodes := uint32(21)
	consensusSize := uint32(21)
	numInvalid := uint32(7)
	roundTime := uint64(2000)
	nodes, advertiser := initNodesWithBlockProcessor(numNodes, consensusSize, numInvalid, roundTime, 0)

	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second * 1)
	mutex := &sync.Mutex{}
	combinedMap := make(map[uint64]uint32, 0)
	minNonce := ^uint64(0)
	maxNonce := uint64(0)
	for _, n := range nodes {
		mockBlockChain, _ := n.blkc.(*mock.BlockChainMock)

		mockBlockChain.SetCurrentBlockHeaderCalled = func(handler data.HeaderHandler) error {
			mutex.Lock()

			err := mockBlockChain.BlockChain.SetCurrentBlockHeader(handler)

			combinedMap[handler.GetNonce()] = handler.GetRound()
			if minNonce > handler.GetNonce() {
				minNonce = handler.GetNonce()
			}
			if maxNonce < handler.GetNonce() {
				maxNonce = handler.GetNonce()
			}

			mutex.Unlock()

			return err
		}

		_ = n.node.StartConsensus()
	}

	time.Sleep(time.Second * 60)

	mutex.Lock()
	assert.Equal(t, uint64(0), maxNonce)
	mutex.Unlock()
}

func TestConsensusMetaProcessorFullConsensus(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	//numNodes := uint32(21)
	//consensusSize := uint32(21)
	//numInvalid := uint32(0)
	//roundTime := uint64(4000)
	//shardId := sharding.MetachainShardId
	//numCommBlock := uint32(10)
	//nodes, advertiser := initNodesWithBlockProcessor(numNodes, consensusSize, numInvalid, roundTime, shardId)
	//
	//defer func() {
	//	advertiser.Close()
	//	for _, n := range nodes {
	//		n.node.Stop()
	//	}
	//}()
	//
	//// delay for bootstrapping and topic announcement
	//fmt.Println("Start consensus...")
	//time.Sleep(time.Second * 1)
	//mutex := &sync.Mutex{}
	//combinedMap := make(map[uint64]uint32, 0)
	//minNonce := ^uint64(0)
	//maxNonce := uint64(0)
	//for _, n := range nodes {
	//	mockBlockChain, _ := n.blkc.(*mock.BlockChainMock)
	//
	//	mockBlockChain.SetCurrentBlockHeaderCalled = func(handler data.HeaderHandler) error {
	//		mutex.Lock()
	//
	//		err := mockBlockChain.BlockChain.SetCurrentBlockHeader(handler)
	//
	//		combinedMap[handler.GetNonce()] = handler.GetRound()
	//		if minNonce > handler.GetNonce() {
	//			minNonce = handler.GetNonce()
	//		}
	//		if maxNonce < handler.GetNonce() {
	//			maxNonce = handler.GetNonce()
	//		}
	//
	//		mutex.Unlock()
	//
	//		return err
	//	}
	//
	//	_ = n.node.StartConsensus()
	//}
	//
	//time.Sleep(time.Second * 20)
	//chDone := make(chan bool, 0)
	//go checkBlockProposedForEachNonce(numCommBlock, combinedMap, &minNonce, &maxNonce, mutex, chDone, t)
	//
	//select {
	//case <-chDone:
	//case <-time.After(180 * time.Second):
	//	assert.Fail(t, "consensus too slow, not working %d %d")
	//	return
	//}
}

func TestConsensusMetaProcessorNotEnoughValidators(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodes := uint32(21)
	consensusSize := uint32(21)
	numInvalid := uint32(7)
	roundTime := uint64(2000)
	shardId := sharding.MetachainShardId
	nodes, advertiser := initNodesWithBlockProcessor(numNodes, consensusSize, numInvalid, roundTime, shardId)

	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second * 1)
	mutex := &sync.Mutex{}
	combinedMap := make(map[uint64]uint32, 0)
	minNonce := ^uint64(0)
	maxNonce := uint64(0)
	for _, n := range nodes {
		mockBlockChain, _ := n.blkc.(*mock.BlockChainMock)

		mockBlockChain.SetCurrentBlockHeaderCalled = func(handler data.HeaderHandler) error {
			mutex.Lock()

			err := mockBlockChain.BlockChain.SetCurrentBlockHeader(handler)

			combinedMap[handler.GetNonce()] = handler.GetRound()
			if minNonce > handler.GetNonce() {
				minNonce = handler.GetNonce()
			}
			if maxNonce < handler.GetNonce() {
				maxNonce = handler.GetNonce()
			}

			mutex.Unlock()

			return err
		}

		_ = n.node.StartConsensus()
	}

	time.Sleep(time.Second * 60)

	mutex.Lock()
	assert.Equal(t, uint64(0), maxNonce)
	mutex.Unlock()
}
