package integrationTests

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	kmultisig "github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/multisig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go/data"
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
	nodesCoordinator sharding.NodesCoordinator,
	cp *CryptoParams,
	keyIndex int,
	ownAccount *TestWalletAccount,
	headerSigVerifier process.InterceptedHeaderSigVerifier,
) *TestProcessorNode {

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	messenger := CreateMessengerWithKadDht(context.Background(), initialNodeAddr)
	tpn := &TestProcessorNode{
		ShardCoordinator:  shardCoordinator,
		Messenger:         messenger,
		NodesCoordinator:  nodesCoordinator,
		HeaderSigVerifier: headerSigVerifier,
		ChainID:           IntegrationTestsChainID,
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
	if nodeShardId == sharding.MetachainShardId {
		accountShardId = 0
	}

	if ownAccount == nil {
		tpn.OwnAccount = CreateTestWalletAccount(shardCoordinator, accountShardId)
	} else {
		tpn.OwnAccount = ownAccount
	}
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
	cp := CreateCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards))
	pubKeys := PubKeysMapFromKeysMap(cp.Keys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, uint32(nbShards))
	nodesMap := make(map[uint32][]*TestProcessorNode)
	for shardId, validatorList := range validatorsMap {
		argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
			ShardConsensusGroupSize: shardConsensusGroupSize,
			MetaConsensusGroupSize:  metaConsensusGroupSize,
			Hasher:                  TestHasher,
			ShardId:                 shardId,
			NbShards:                uint32(nbShards),
			Nodes:                   validatorsMap,
			SelfPublicKey:           []byte(strconv.Itoa(int(shardId))),
		}
		nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)

		if err != nil {
			fmt.Println("Error creating node coordinator")
		}

		nodesList := make([]*TestProcessorNode, len(validatorList))
		for i := range validatorList {
			nodesList[i] = NewTestProcessorNodeWithCustomNodesCoordinator(
				uint32(nbShards),
				shardId,
				seedAddress,
				nodesCoordinator,
				cp,
				i,
				nil,
				&mock.HeaderSigVerifierStub{},
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
	nodesMap := make(map[uint32][]*TestProcessorNode)
	for shardId, validatorList := range validatorsMap {
		argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
			ShardConsensusGroupSize: shardConsensusGroupSize,
			MetaConsensusGroupSize:  metaConsensusGroupSize,
			Hasher:                  TestHasher,
			ShardId:                 shardId,
			NbShards:                uint32(nbShards),
			Nodes:                   validatorsMap,
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
				keyGenForBlocks)
			args := headerCheck.ArgsHeaderSigVerifier{
				Marshalizer:       TestMarshalizer,
				Hasher:            TestHasher,
				NodesCoordinator:  nodesCoordinator,
				MultiSigVerifier:  TestMultiSig,
				SingleSigVerifier: singleSigner,
				KeyGen:            keyGenForBlocks}
			headerSig, _ := headerCheck.NewHeaderSigVerifier(&args)
			nodesList[i] = NewTestProcessorNodeWithCustomNodesCoordinator(
				uint32(nbShards),
				shardId,
				seedAddress,
				nodesCoordinator,
				cp,
				i,
				ownAccount,
				headerSig,
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
) (data.BodyHandler, data.HeaderHandler, [][]byte, []*TestProcessorNode) {

	nodesCoordinator := nodesMap[shardId][0].NodesCoordinator
	pubKeys, err := nodesCoordinator.GetValidatorsPublicKeys(randomness, round, shardId)
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
	prevRandomness map[uint32][]byte,
	nodesMap map[uint32][]*TestProcessorNode,
) (
	map[uint32]data.BodyHandler,
	map[uint32]data.HeaderHandler,
	map[uint32][]*TestProcessorNode,
	map[uint32][]byte,
) {

	body := make(map[uint32]data.BodyHandler)
	header := make(map[uint32]data.HeaderHandler)
	consensusNodes := make(map[uint32][]*TestProcessorNode)
	newRandomness := make(map[uint32][]byte)

	// propose blocks
	for i := range nodesMap {
		body[i], header[i], _, consensusNodes[i] = ProposeBlockWithConsensusSignature(i, nodesMap, round, nonce, prevRandomness[i])
		newRandomness[i] = header[i].GetRandSeed()
	}

	// propagate blocks
	for i := range nodesMap {
		consensusNodes[i][0].BroadcastBlock(body[i], header[i])
		consensusNodes[i][0].CommitBlock(body[i], header[i])
	}

	time.Sleep(2 * time.Second)

	return body, header, consensusNodes, newRandomness
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
