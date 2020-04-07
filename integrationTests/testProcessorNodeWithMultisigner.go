package integrationTests

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	mclmultisig "github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/multisig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/headerCheck"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

// NewTestProcessorNodeWithCustomNodesCoordinator returns a new TestProcessorNode instance with custom NodesCoordinator
func NewTestProcessorNodeWithCustomNodesCoordinator(
	maxShards uint32,
	nodeShardId uint32,
	initialNodeAddr string,
	epochStartNotifier notifier.EpochStartNotifier,
	nodesCoordinator sharding.NodesCoordinator,
	cp *CryptoParams,
	keyIndex int,
	ownAccount *TestWalletAccount,
	headerSigVerifier process.InterceptedHeaderSigVerifier,
	initialNodes []*sharding.InitialNode,
) *TestProcessorNode {

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	messenger := CreateMessengerWithKadDht(context.Background(), initialNodeAddr)
	tpn := &TestProcessorNode{
		ShardCoordinator:  shardCoordinator,
		Messenger:         messenger,
		NodesCoordinator:  nodesCoordinator,
		HeaderSigVerifier: headerSigVerifier,
		ChainID:           ChainID,
		InitialNodes:      initialNodes,
	}

	tpn.NodeKeys = cp.Keys[nodeShardId][keyIndex]
	blsHasher := &blake2b.Blake2b{HashSize: hashing.BlsHashSize}
	llsig := &mclmultisig.BlsMultiSigner{Hasher: blsHasher}

	pubKeysMap := PubKeysMapFromKeysMap(cp.Keys)

	tpn.MultiSigner, _ = multisig.NewBLSMultisig(
		llsig,
		pubKeysMap[nodeShardId],
		tpn.NodeKeys.Sk,
		cp.KeyGen,
		0,
	)
	if tpn.MultiSigner == nil {
		fmt.Println("Error generating multisigner")
	}
	accountShardId := nodeShardId
	if nodeShardId == core.MetachainShardId {
		accountShardId = 0
	}

	if ownAccount == nil {
		tpn.OwnAccount = CreateTestWalletAccount(shardCoordinator, accountShardId)
	} else {
		tpn.OwnAccount = ownAccount
	}

	tpn.EpochStartNotifier = epochStartNotifier
	tpn.initDataPools()
	tpn.initTestNode()

	return tpn
}

// CreateNodesWithNodesCoordinator returns a map with nodes per shard each using a real nodes coordinator
func CreateNodesWithNodesCoordinator(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	seedAddress string,
) map[uint32][]*TestProcessorNode {
	return CreateNodesWithNodesCoordinatorWithCacher(nodesPerShard, nbMetaNodes, nbShards, shardConsensusGroupSize, metaConsensusGroupSize, seedAddress)
}

// CreateNodesWithNodesCoordinatorWithCacher returns a map with nodes per shard each using a real nodes coordinator with cacher
func CreateNodesWithNodesCoordinatorWithCacher(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	seedAddress string,
) map[uint32][]*TestProcessorNode {
	coordinatorFactory := &IndexHashedNodesCoordinatorFactory{}
	return CreateNodesWithNodesCoordinatorFactory(nodesPerShard, nbMetaNodes, nbShards, shardConsensusGroupSize, metaConsensusGroupSize, seedAddress, coordinatorFactory)

}

// CreateNodesWithNodesCoordinatorFactory returns a map with nodes per shard each using a real nodes coordinator
func CreateNodesWithNodesCoordinatorFactory(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	seedAddress string,
	nodesCoordinatorFactory NodesCoordinatorFactory,
) map[uint32][]*TestProcessorNode {
	cp := CreateCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards))
	pubKeys := PubKeysMapFromKeysMap(cp.Keys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, uint32(nbShards))

	cpWaiting := CreateCryptoParams(1, 1, uint32(nbShards))
	pubKeysWaiting := PubKeysMapFromKeysMap(cpWaiting.Keys)
	waitingMap := GenValidatorsFromPubKeys(pubKeysWaiting, uint32(nbShards))

	nodesMap := make(map[uint32][]*TestProcessorNode)

	for shardId, validatorList := range validatorsMap {
		nodesList := make([]*TestProcessorNode, len(validatorList))
		nodesListWaiting := make([]*TestProcessorNode, len(waitingMap[shardId]))

		for i := range validatorList {
			dataCache, _ := lrucache.NewCache(10000)
			nodesList[i] = CreateNode(
				nodesPerShard,
				nbMetaNodes,
				shardConsensusGroupSize,
				metaConsensusGroupSize,
				shardId,
				nbShards,
				validatorsMap,
				waitingMap,
				i,
				seedAddress,
				cp,
				dataCache,
				nodesCoordinatorFactory,
			)
		}

		for i := range waitingMap[shardId] {
			dataCache, _ := lrucache.NewCache(10000)
			nodesListWaiting[i] = CreateNode(
				nodesPerShard,
				nbMetaNodes,
				shardConsensusGroupSize,
				metaConsensusGroupSize,
				shardId,
				nbShards,
				validatorsMap,
				waitingMap,
				i,
				seedAddress,
				cpWaiting,
				dataCache,
				nodesCoordinatorFactory,
			)
		}

		nodesMap[shardId] = append(nodesList, nodesListWaiting...)
	}

	return nodesMap
}

func CreateNode(
	nodesPerShard int,
	nbMetaNodes int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	shardId uint32,
	nbShards int,
	validatorsMap map[uint32][]sharding.Validator,
	waitingMap map[uint32][]sharding.Validator,
	keyIndex int,
	seedAddress string,
	cp *CryptoParams,
	cache sharding.Cacher,
	coordinatorFactory NodesCoordinatorFactory,
) *TestProcessorNode {

	initialNodes := createInitialNodes(validatorsMap, waitingMap)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := CreateMemUnit()

	argFactory := ArgIndexHashedNodesCoordinatorFactory{
		nodesPerShard,
		nbMetaNodes,
		shardConsensusGroupSize,
		metaConsensusGroupSize,
		shardId,
		nbShards,
		validatorsMap,
		waitingMap,
		keyIndex,
		cp,
		epochStartSubscriber,
		TestHasher,
		cache,
		bootStorer,
	}

	nodesCoordinator := coordinatorFactory.CreateNodesCoordinator(argFactory)

	return NewTestProcessorNodeWithCustomNodesCoordinator(
		uint32(nbShards),
		shardId,
		seedAddress,
		epochStartSubscriber,
		nodesCoordinator,
		cp,
		keyIndex,
		nil,
		&mock.HeaderSigVerifierStub{},
		initialNodes,
	)
}

func createInitialNodes(validatorsMap map[uint32][]sharding.Validator, waitingMap map[uint32][]sharding.Validator) []*sharding.InitialNode {
	initialNodes := make([]*sharding.InitialNode, 0)

	for _, pks := range validatorsMap {
		for _, validator := range pks {
			n := &sharding.InitialNode{
				PubKey:   core.ToHex(validator.PubKey()),
				Address:  core.ToHex(validator.Address()),
				NodeInfo: sharding.NodeInfo{},
			}
			initialNodes = append(initialNodes, n)
		}
	}

	for _, pks := range waitingMap {
		for _, validator := range pks {
			n := &sharding.InitialNode{
				PubKey:   core.ToHex(validator.PubKey()),
				Address:  core.ToHex(validator.Address()),
				NodeInfo: sharding.NodeInfo{},
			}
			initialNodes = append(initialNodes, n)
		}
	}

	sort.Slice(initialNodes, func(i, j int) bool {
		return bytes.Compare([]byte(initialNodes[i].PubKey), []byte(initialNodes[j].PubKey)) > 0
	})
	return initialNodes
}

// CreateNodesWithNodesCoordinatorAndHeaderSigVerifier returns a map with nodes per shard each using a real nodes coordinator and header sig verifier
func CreateNodesWithNodesCoordinatorAndHeaderSigVerifier(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	seedAddress string,
	signer crypto.SingleSigner,
	keyGen crypto.KeyGenerator,
) map[uint32][]*TestProcessorNode {
	cp := CreateCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards))
	pubKeys := PubKeysMapFromKeysMap(cp.Keys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, uint32(nbShards))
	nodesMap := make(map[uint32][]*TestProcessorNode)
	nodeShuffler := sharding.NewXorValidatorsShuffler(uint32(nodesPerShard), uint32(nbMetaNodes), 0.2, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := CreateMemUnit()
	waitingMap := make(map[uint32][]sharding.Validator)
	initialNodes := createInitialNodes(validatorsMap, waitingMap)

	for shardId, validatorList := range validatorsMap {
		consensusCache, _ := lrucache.NewCache(10000)
		argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
			ShardConsensusGroupSize: shardConsensusGroupSize,
			MetaConsensusGroupSize:  metaConsensusGroupSize,
			Hasher:                  TestHasher,
			Shuffler:                nodeShuffler,
			BootStorer:              bootStorer,
			EpochStartSubscriber:    epochStartSubscriber,
			ShardIDAsObserver:       shardId,
			NbShards:                uint32(nbShards),
			EligibleNodes:           validatorsMap,
			WaitingNodes:            make(map[uint32][]sharding.Validator),
			SelfPublicKey:           []byte(strconv.Itoa(int(shardId))),
			ConsensusGroupCache:     consensusCache,
		}
		nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)

		if err != nil {
			fmt.Println("Error creating node coordinator")
		}

		nodesList := make([]*TestProcessorNode, len(validatorList))
		args := headerCheck.ArgsHeaderSigVerifier{
			Marshalizer:       TestMarshalizer,
			Hasher:            TestHasher,
			NodesCoordinator:  nodesCoordinator,
			MultiSigVerifier:  TestMultiSig,
			SingleSigVerifier: signer,
			KeyGen:            keyGen,
		}
		headerSig, _ := headerCheck.NewHeaderSigVerifier(&args)
		for i := range validatorList {
			nodesList[i] = NewTestProcessorNodeWithCustomNodesCoordinator(
				uint32(nbShards),
				shardId,
				seedAddress,
				epochStartSubscriber,
				nodesCoordinator,
				cp,
				i,
				nil,
				headerSig,
				initialNodes,
			)
		}
		nodesMap[shardId] = nodesList
	}

	return nodesMap
}

// CreateNodesWithNodesCoordinatorKeygenAndSingleSigner returns a map with nodes per shard each using a real nodes coordinator
// and a given single signer for blocks and a given key gen for blocks
func CreateNodesWithNodesCoordinatorKeygenAndSingleSigner(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	seedAddress string,
	singleSigner crypto.SingleSigner,
	keyGenForBlocks crypto.KeyGenerator,
) map[uint32][]*TestProcessorNode {
	cp := CreateCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards))
	pubKeys := PubKeysMapFromKeysMap(cp.Keys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, uint32(nbShards))

	cpWaiting := CreateCryptoParams(2, 2, uint32(nbShards))
	pubKeysWaiting := PubKeysMapFromKeysMap(cpWaiting.Keys)
	waitingMap := GenValidatorsFromPubKeys(pubKeysWaiting, uint32(nbShards))

	nodesMap := make(map[uint32][]*TestProcessorNode)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	nodeShuffler := &mock.NodeShufflerMock{}

	for shardId, validatorList := range validatorsMap {
		initialNodes := createInitialNodes(validatorsMap, waitingMap)
		bootStorer := CreateMemUnit()
		cache, _ := lrucache.NewCache(10000)
		argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
			ShardConsensusGroupSize: shardConsensusGroupSize,
			MetaConsensusGroupSize:  metaConsensusGroupSize,
			Hasher:                  TestHasher,
			Shuffler:                nodeShuffler,
			EpochStartSubscriber:    epochStartSubscriber,
			BootStorer:              bootStorer,
			ShardIDAsObserver:       shardId,
			NbShards:                uint32(nbShards),
			EligibleNodes:           validatorsMap,
			WaitingNodes:            waitingMap,
			SelfPublicKey:           []byte(strconv.Itoa(int(shardId))),
			ConsensusGroupCache:     cache,
		}
		nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)

		if err != nil {
			fmt.Println("Error creating node coordinator")
		}

		nodesList := make([]*TestProcessorNode, len(validatorList))
		shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(nbShards), shardId)
		for i := range validatorList {
			ownAccount := CreateTestWalletAccountWithKeygenAndSingleSigner(
				shardCoordinator,
				shardId,
				singleSigner,
				keyGenForBlocks,
			)

			args := headerCheck.ArgsHeaderSigVerifier{
				Marshalizer:       TestMarshalizer,
				Hasher:            TestHasher,
				NodesCoordinator:  nodesCoordinator,
				MultiSigVerifier:  TestMultiSig,
				SingleSigVerifier: singleSigner,
				KeyGen:            keyGenForBlocks,
			}

			headerSig, _ := headerCheck.NewHeaderSigVerifier(&args)
			nodesList[i] = NewTestProcessorNodeWithCustomNodesCoordinator(
				uint32(nbShards),
				shardId,
				seedAddress,
				epochStartSubscriber,
				nodesCoordinator,
				cp,
				i,
				ownAccount,
				headerSig,
				initialNodes,
			)
		}
		nodesMap[shardId] = nodesList
	}

	return nodesMap
}

// ProposeBlockWithConsensusSignature proposes
func ProposeBlockWithConsensusSignature(
	shardId uint32,
	nodesMap map[uint32][]*TestProcessorNode,
	round uint64,
	nonce uint64,
	randomness []byte,
	epoch uint32,
) (data.BodyHandler, data.HeaderHandler, [][]byte, []*TestProcessorNode) {
	nodesCoordinator := nodesMap[shardId][0].NodesCoordinator

	pubKeys, err := nodesCoordinator.GetConsensusValidatorsPublicKeys(randomness, round, shardId, epoch)
	if err != nil {
		fmt.Println("Error getting the validators public keys: ", err)
	}

	// select nodes from map based on their pub keys
	consensusNodes := selectTestNodesForPubKeys(nodesMap[shardId], pubKeys)
	// first node is block proposer
	body, header, txHashes := consensusNodes[0].ProposeBlock(round, nonce)
	header.SetPrevRandSeed(randomness)
	header = DoConsensusSigningOnBlock(header, consensusNodes, pubKeys)

	return body, header, txHashes, consensusNodes
}

func selectTestNodesForPubKeys(nodes []*TestProcessorNode, pubKeys []string) []*TestProcessorNode {
	selectedNodes := make([]*TestProcessorNode, len(pubKeys))
	cntNodes := 0

	for i, pk := range pubKeys {
		for _, node := range nodes {
			pubKeyBytes, _ := node.NodeKeys.Pk.ToByteArray()
			if bytes.Equal(pubKeyBytes, []byte(pk)) {
				selectedNodes[i] = node
				cntNodes++
			}
		}
	}

	if cntNodes != len(pubKeys) {
		fmt.Println("Error selecting nodes from public keys")
	}

	return selectedNodes
}

// DoConsensusSigningOnBlock simulates a consensus aggregated signature on the provided block
func DoConsensusSigningOnBlock(
	blockHeader data.HeaderHandler,
	consensusNodes []*TestProcessorNode,
	pubKeys []string,
) data.HeaderHandler {
	// set bitmap for all consensus nodes signing
	bitmap := make([]byte, len(consensusNodes)/8+1)
	for i := range bitmap {
		bitmap[i] = 0xFF
	}

	bitmap[len(consensusNodes)/8] >>= uint8(8 - (len(consensusNodes) % 8))
	blockHeader.SetPubKeysBitmap(bitmap)
	// clear signature, as we need to compute it below
	blockHeader.SetSignature(nil)
	blockHeader.SetPubKeysBitmap(nil)
	blockHeaderHash, _ := core.CalculateHash(TestMarshalizer, TestHasher, blockHeader)

	var msig crypto.MultiSigner
	msigProposer, _ := consensusNodes[0].MultiSigner.Create(pubKeys, 0)
	_, _ = msigProposer.CreateSignatureShare(blockHeaderHash, bitmap)

	for i := 1; i < len(consensusNodes); i++ {
		msig, _ = consensusNodes[i].MultiSigner.Create(pubKeys, uint16(i))
		sigShare, _ := msig.CreateSignatureShare(blockHeaderHash, bitmap)
		_ = msigProposer.StoreSignatureShare(uint16(i), sigShare)
	}

	sig, _ := msigProposer.AggregateSigs(bitmap)
	blockHeader.SetSignature(sig)
	blockHeader.SetPubKeysBitmap(bitmap)
	blockHeader.SetLeaderSignature([]byte("leader sign"))

	return blockHeader
}

// DoConsensusSigningOnBlock simulates a consensus aggregated signature on the provided block
func SimulateDoConsensusSigningOnBlock(
	blockHeader data.HeaderHandler,
	pubKeys []string,
	proposer *TestProcessorNode,
) data.HeaderHandler {
	pubKeysLen := len(pubKeys)
	signersLen := pubKeysLen*2/3 + 1
	// set bitmap for all consensus nodes signing
	bitmap := make([]byte, pubKeysLen/8+1)
	for i := 0; i < signersLen/8+1; i++ {
		bitmap[i] = 0xFF
	}

	bitmap[signersLen/8] >>= uint8(8 - (signersLen % 8))
	blockHeader.SetPubKeysBitmap(bitmap)
	// clear signature, as we need to compute it below
	blockHeader.SetSignature(nil)
	blockHeader.SetPubKeysBitmap(nil)
	blockHeaderHash, _ := core.CalculateHash(TestMarshalizer, TestHasher, blockHeader)

	//var msig crypto.MultiSigner
	msigProposer, _ := proposer.MultiSigner.Create(pubKeys, 0)
	_, _ = msigProposer.CreateSignatureShare(blockHeaderHash, bitmap)

	//for i := 1; i < len(pubKeys); i++ {
	//	msig, _ = proposer.MultiSigner.Create(pubKeys, uint16(i))
	//	sigShare, _ := msig.CreateSignatureShare(blockHeaderHash, bitmap)
	//	_ = msigProposer.StoreSignatureShare(uint16(i), sigShare)
	//}

	sig, _ := msigProposer.AggregateSigs(bitmap)
	blockHeader.SetSignature(sig)
	blockHeader.SetPubKeysBitmap(bitmap)
	blockHeader.SetLeaderSignature([]byte("leader sign"))

	return blockHeader
}

// AllShardsProposeBlock simulates each shard selecting a consensus group and proposing/broadcasting/committing a block
func AllShardsProposeBlock(
	round uint64,
	nonce uint64,
	nodesMap map[uint32][]*TestProcessorNode,
) (
	map[uint32]data.BodyHandler,
	map[uint32]data.HeaderHandler,
	map[uint32][]*TestProcessorNode,
) {

	body := make(map[uint32]data.BodyHandler)
	header := make(map[uint32]data.HeaderHandler)
	consensusNodes := make(map[uint32][]*TestProcessorNode)
	newRandomness := make(map[uint32][]byte)

	// propose blocks
	for i := range nodesMap {
		currentBlockHeader := nodesMap[i][0].BlockChain.GetCurrentBlockHeader()
		if check.IfNil(currentBlockHeader) {
			currentBlockHeader = nodesMap[i][0].BlockChain.GetGenesisHeader()
		}

		// TODO: remove if start of epoch block needs to be validated by the new epoch nodes
		epoch := currentBlockHeader.GetEpoch()
		prevRandomness := currentBlockHeader.GetRandSeed()
		body[i], header[i], _, consensusNodes[i] = ProposeBlockWithConsensusSignature(
			i, nodesMap, round, nonce, prevRandomness, epoch,
		)
		newRandomness[i] = header[i].GetRandSeed()
	}

	// propagate blocks
	for i := range nodesMap {
		consensusNodes[i][0].BroadcastBlock(body[i], header[i])
		consensusNodes[i][0].CommitBlock(body[i], header[i])
	}

	time.Sleep(2 * time.Second)

	return body, header, consensusNodes
}

// AllShardsProposeBlock simulates each shard selecting a consensus group and proposing/broadcasting/committing a block
func SimulateAllShardsProposeBlock(
	round uint64,
	nonce uint64,
	nodesMap map[uint32][]*TestProcessorNode,
) (
	map[uint32]data.BodyHandler,
	map[uint32]data.HeaderHandler,
	map[uint32][]*TestProcessorNode,
) {

	bodyMap := make(map[uint32]data.BodyHandler)
	headerMap := make(map[uint32]data.HeaderHandler)
	consensusNodesMap := make(map[uint32][]*TestProcessorNode)
	newRandomness := make(map[uint32][]byte)

	wg := &sync.WaitGroup{}

	wg.Add(len(nodesMap))
	mutMaps := &sync.Mutex{}
	// propose blocks
	for i := range nodesMap {
		currentNode := nodesMap[i][0]
		go createBlock(currentNode, i, round, nonce, bodyMap, headerMap, newRandomness, mutMaps, wg)
	}
	wg.Wait()

	time.Sleep(1 * time.Millisecond)

	return bodyMap, headerMap, consensusNodesMap
}

func createBlock(currentNode *TestProcessorNode,
	i uint32,
	round uint64,
	nonce uint64,
	bodyMap map[uint32]data.BodyHandler,
	headerMap map[uint32]data.HeaderHandler,
	newRandomness map[uint32][]byte,
	mutex *sync.Mutex,
	wg *sync.WaitGroup) {
	currentBlockHeader := currentNode.BlockChain.GetCurrentBlockHeader()
	if check.IfNil(currentBlockHeader) {
		currentBlockHeader = currentNode.BlockChain.GetGenesisHeader()
	}

	// TODO: remove if start of epoch block needs to be validated by the new epoch nodes
	epoch := currentBlockHeader.GetEpoch()
	prevRandomness := currentBlockHeader.GetRandSeed()

	nodesCoordinator := currentNode.NodesCoordinator

	pubKeys, err := nodesCoordinator.GetConsensusValidatorsPublicKeys(prevRandomness, round, i, epoch)
	if err != nil {
		fmt.Println("Error getting the validators public keys: ", err)
	}

	// first node is block proposer
	var body data.BodyHandler
	var header data.HeaderHandler
	for i := 0; i < 10; i++ {
		body, header, _ = currentNode.ProposeBlock(round, nonce)
		if body != nil && header != nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	header.SetPrevRandSeed(prevRandomness)
	header = SimulateDoConsensusSigningOnBlock(header, pubKeys, currentNode)

	mutex.Lock()
	bodyMap[i] = body
	headerMap[i] = header
	//consensusNodesMap[i] = consensusNodes
	newRandomness[i] = headerMap[i].GetRandSeed()
	mutex.Unlock()

	currentNode.CommitBlock(body, header)

	wg.Done()
}

// SyncAllShardsWithRoundBlock enforces all nodes in each shard synchronizing the block for the given round
func SyncAllShardsWithRoundBlock(
	t *testing.T,
	nodesMap map[uint32][]*TestProcessorNode,
	indexProposers map[uint32]int,
	round uint64,
) {
	for shard, nodeList := range nodesMap {
		SyncBlock(t, nodeList, []int{indexProposers[shard]}, round)
	}
	time.Sleep(2 * time.Second)
}

// SyncAllShardsWithRoundBlock enforces all nodes in each shard synchronizing the block for the given round
func SimulateSyncAllShardsWithRoundBlock(
	nodesMap map[uint32][]*TestProcessorNode,
	headerMap map[uint32]data.HeaderHandler,
	bodyMap map[uint32]data.BodyHandler,
) {
	for shard, nodeList := range nodesMap {
		for _, node := range nodeList {
			for _, header := range headerMap {
				if header.GetShardID() == shard {
					continue
				}
				marshalizedHeader, _ := TestMarshalizer.Marshal(header)
				headerHash := TestHasher.Compute(string(marshalizedHeader))

				if shard == core.MetachainShardId {
					node.DataPool.Headers().AddHeader(headerHash, header)
				} else {
					if header.GetShardID() == core.MetachainShardId {
						node.DataPool.Headers().AddHeader(headerHash, header)
					}
				}
			}

			for _, body := range bodyMap {
				actualBody := body.(*block.Body)
				for _, miniblocks := range actualBody.MiniBlocks {
					if miniblocks.ReceiverShardID != shard ||
						miniblocks.SenderShardID != shard ||
						miniblocks.SenderShardID != core.AllShardId ||
						miniblocks.ReceiverShardID != core.AllShardId {
						continue
					}
					marshalizedHeader, _ := TestMarshalizer.Marshal(miniblocks)
					headerHash := TestHasher.Compute(string(marshalizedHeader))
					node.DataPool.MiniBlocks().Put(headerHash, miniblocks)
				}
			}
		}
	}
}
