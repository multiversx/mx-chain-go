package consensus

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
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

func createConsensusWithMultiSigBN(
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
	*mock.BlockProcessorMock,
	data.ChainHandler) {

	messenger := createMessengerWithKadDht(context.Background(), initialAddr)
	rootHash := []byte("roothash")

	blockProcessor := &mock.BlockProcessorMock{
		CreateGenesisBlockCalled: func(balances map[string]*big.Int) (handler data.HeaderHandler, e error) {
			header := &dataBlock.Header{
				Nonce:         0,
				ShardId:       shardId,
				BlockBodyType: dataBlock.StateBlock,
				Signature:     rootHash,
				RootHash:      rootHash,
				PrevRandSeed:  rootHash,
				RandSeed:      rootHash,
			}

			return header, nil
		},
		ProcessBlockCalled: func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
			_ = blockChain.SetCurrentBlockHeader(header)
			_ = blockChain.SetCurrentBlockBody(body)
			return nil
		},
		RevertAccountStateCalled: func() {
		},
		CreateBlockCalled: func(round int32, haveTime func() bool) (handler data.BodyHandler, e error) {
			return &dataBlock.Body{}, nil
		},
		CreateBlockHeaderCalled: func(body data.BodyHandler, round int32, haveTime func() bool) (handler data.HeaderHandler, e error) {
			return &dataBlock.Header{Round: uint32(round)}, nil
		},
		MarshalizedDataToBroadcastCalled: func(header data.HeaderHandler, body data.BodyHandler) (bytes map[uint32][]byte, bytes2 map[uint32][][]byte, e error) {
			mrsData := make(map[uint32][]byte)
			mrsTxs := make(map[uint32][][]byte)
			return mrsData, mrsTxs, nil
		},
	}

	blockProcessor.CommitBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
		blockProcessor.NrCommitBlockCalled++
		_ = blockChain.SetCurrentBlockHeader(header)
		_ = blockChain.SetCurrentBlockBody(body)
		return nil
	}

	blockChain := createTestBlockChain()

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

	multiSigner, _ := multisig.NewBelNevMultisig(testHasher, inPubKeys[shardId], privKey, testKeyGen, uint16(selfId))

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
		node.WithBlockChain(blockChain),
		node.WithMultisig(multiSigner),
		node.WithSinglesig(singlesigner),
		node.WithPrivateKey(privKey),
		node.WithPublicKey(privKey.GeneratePublic()),
		node.WithBlockProcessor(blockProcessor),
		node.WithDataPool(createTestShardDataPool()),
		node.WithDataStore(createTestStore()),
		node.WithResolversFinder(resolverFinder),
	)

	if err != nil {
		fmt.Println(err.Error())
	}

	return n, messenger, blockProcessor, blockChain
}

func createNodesConsensusMultiSigBN(
	nodesPerShard int,
	consensusSize int,
	roundTime uint64,
	serviceID string,
) []*testNode {

	privKeys, pubKeys, testKeyGen := initialPrivPubKeys(nodesPerShard)
	//first node generated will have is pk belonging to firstSkShardId
	nodes := make([]*testNode, nodesPerShard)

	for i := 0; i < nodesPerShard; i++ {
		testNode := &testNode{
			shardId: uint32(0),
		}

		shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(1), uint32(0))
		n, mes, blkProcessor, blkc := createConsensusWithMultiSigBN(
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
		testNode.blkProcessor = blkProcessor
		testNode.blkc = blkc

		nodes[i] = testNode
	}

	return nodes
}

func initNodesWithMultiSigAndTest(numNodes, consensusSize, numInvalid uint32, roundTime uint64, numCommBlock uint32) ([]*testNode, p2p.Messenger, *sync.Map) {
	fmt.Println("Step 1. Setup nodes...")

	advertiser := createMessengerWithKadDht(context.Background(), "")
	advertiser.Bootstrap()

	concMap := &sync.Map{}

	nodes := createNodesConsensusMultiSigBN(
		int(numNodes),
		int(consensusSize),
		roundTime,
		getConnectableAddress(advertiser),
	)
	displayAndStartNodes(nodes)

	if numInvalid < numNodes {
		for i := uint32(0); i < numInvalid; i++ {
			nodes[i].blkProcessor.ProcessBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				return process.ErrInvalidBlockHash
			}
		}
	}

	return nodes, advertiser, concMap
}

func TestConsensusMultisigFullConsensus(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodes := uint32(21)
	consensusSize := uint32(21)
	numInvalid := uint32(0)
	roundTime := uint64(4000)
	numCommBlock := uint32(10)
	nodes, advertiser, _ := initNodesWithMultiSigAndTest(numNodes, consensusSize, numInvalid, roundTime, numCommBlock)

	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second * 1)

	combinedMap := make(map[uint32]uint64)
	totalCalled := 0
	mutex := &sync.Mutex{}

	for _, n := range nodes {
		n.blkProcessor.CommitBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
			n.blkProcessor.NrCommitBlockCalled++
			_ = blockChain.SetCurrentBlockHeader(header)
			_ = blockChain.SetCurrentBlockBody(body)

			mutex.Lock()
			combinedMap[header.GetRound()] = header.GetNonce()
			totalCalled += 1
			mutex.Unlock()

			return nil
		}
		_ = n.node.StartConsensus()
	}

	time.Sleep(time.Second * 20)
	chDone := make(chan bool, 0)
	go checkBlockProposedEveryRound(numCommBlock, combinedMap, mutex, chDone, t)

	select {
	case <-chDone:
	case <-time.After(180 * time.Second):
		mutex.Lock()
		fmt.Println("combined map: \n", combinedMap)
		assert.Fail(t, "consensus too slow, not working %d %d")
		mutex.Unlock()
		return
	}
}

func TestConsensusWithMultiSigTestValidatorsAtLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodes := uint32(21)
	consensusSize := uint32(21)
	numInvalid := uint32(6)
	roundTime := uint64(4000)
	numCommBlock := uint32(10)
	nodes, advertiser, _ := initNodesWithMultiSigAndTest(numNodes, consensusSize, numInvalid, roundTime, numCommBlock)

	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second * 1)

	combinedMap := make(map[uint64]uint32)
	totalCalled := 0
	mutex := &sync.Mutex{}
	maxNonce := uint64(0)
	minNonce := ^uint64(0)
	for _, n := range nodes {
		n.blkProcessor.CommitBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
			n.blkProcessor.NrCommitBlockCalled++
			_ = blockChain.SetCurrentBlockHeader(header)
			_ = blockChain.SetCurrentBlockBody(body)

			mutex.Lock()
			combinedMap[header.GetNonce()] = header.GetRound()
			totalCalled += 1

			if maxNonce < header.GetNonce() {
				maxNonce = header.GetNonce()
			}

			if minNonce < header.GetNonce() {
				minNonce = header.GetNonce()
			}

			mutex.Unlock()

			return nil
		}
		_ = n.node.StartConsensus()
	}

	time.Sleep(time.Second * 20)
	chDone := make(chan bool, 0)
	go checkBlockProposedForEachNonce(numCommBlock, combinedMap, &minNonce, &maxNonce, mutex, chDone, t)

	select {
	case <-chDone:
	case <-time.After(180 * time.Second):
		mutex.Lock()
		fmt.Println("combined map: \n", combinedMap)
		assert.Fail(t, "consensus too slow, not working %d %d")
		mutex.Unlock()
		return
	}
}

func TestConsensusMultiSignNotEnoughValidators(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodes := uint32(21)
	consensusSize := uint32(21)
	numInvalid := uint32(7)
	roundTime := uint64(2000)
	nodes, advertiser, _ := initNodesWithMultiSigAndTest(numNodes, consensusSize, numInvalid, roundTime, 10)

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
	maxNonce := uint64(0)
	minNonce := ^uint64(0)
	for _, n := range nodes {
		n.blkProcessor.CommitBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
			n.blkProcessor.NrCommitBlockCalled++
			_ = blockChain.SetCurrentBlockHeader(header)
			_ = blockChain.SetCurrentBlockBody(body)

			mutex.Lock()
			if maxNonce < header.GetNonce() {
				maxNonce = header.GetNonce()
			}

			if minNonce < header.GetNonce() {
				minNonce = header.GetNonce()
			}
			mutex.Unlock()

			return nil
		}
		_ = n.node.StartConsensus()
	}

	waitTime := time.Second * 60
	fmt.Println("Run for 60 seconds...")
	time.Sleep(waitTime)

	assert.Equal(t, uint64(0), maxNonce)
}
