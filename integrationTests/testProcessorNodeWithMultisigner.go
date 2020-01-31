package integrationTests

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	kmultisig "github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/multisig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/headerCheck"
	"github.com/ElrondNetwork/elrond-go/sharding"
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
	llsig := &kmultisig.KyberMultiSignerBLS{}
	blsHasher := blake2b.Blake2b{HashSize: factory.BlsHashSize}

	pubKeysMap := PubKeysMapFromKeysMap(cp.Keys)

	tpn.MultiSigner, _ = multisig.NewBLSMultisig(
		llsig,
		blsHasher,
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
			nodesList[i] = createNode(
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
				nodesCoordinatorFactory,
			)
		}

		for i := range waitingMap[shardId] {
			nodesListWaiting[i] = createNode(
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
				nodesCoordinatorFactory,
			)
		}

		nodesMap[shardId] = append(nodesList, nodesListWaiting...)
	}

	return nodesMap
}

func createNode(
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
	coordinatorFactory NodesCoordinatorFactory,
) *TestProcessorNode {

	initialNodes := createInitialNodes(validatorsMap, waitingMap)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}

	nodesCoordinator := coordinatorFactory.CreateNodesCoordinator(
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
		TestHasher)

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

	waitingMap := make(map[uint32][]sharding.Validator)

	initialNodes := createInitialNodes(validatorsMap, waitingMap)

	for shardId, validatorList := range validatorsMap {
		argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
			ShardConsensusGroupSize: shardConsensusGroupSize,
			MetaConsensusGroupSize:  metaConsensusGroupSize,
			Hasher:                  TestHasher,
			Shuffler:                nodeShuffler,
			EpochStartSubscriber:    epochStartSubscriber,
			ShardId:                 shardId,
			NbShards:                uint32(nbShards),
			EligibleNodes:           validatorsMap,
			WaitingNodes:            make(map[uint32][]sharding.Validator),
			SelfPublicKey:           []byte(strconv.Itoa(int(shardId))),
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

		argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
			ShardConsensusGroupSize: shardConsensusGroupSize,
			MetaConsensusGroupSize:  metaConsensusGroupSize,
			Hasher:                  TestHasher,
			Shuffler:                nodeShuffler,
			EpochStartSubscriber:    epochStartSubscriber,
			ShardId:                 shardId,
			NbShards:                uint32(nbShards),
			EligibleNodes:           validatorsMap,
			WaitingNodes:            waitingMap,
			SelfPublicKey:           []byte(strconv.Itoa(int(shardId))),
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

	// set some randomness
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
		if currentBlockHeader == nil {
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
