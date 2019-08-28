package integrationTests

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/crypto"
	kmultisig "github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/multisig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
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
) *TestProcessorNode {

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	messenger := CreateMessengerWithKadDht(context.Background(), initialNodeAddr)
	tpn := &TestProcessorNode{
		ShardCoordinator: shardCoordinator,
		Messenger:        messenger,
		NodesCoordinator: nodesCoordinator,
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

	tpn.OwnAccount = CreateTestWalletAccount(shardCoordinator, accountShardId)
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
	validatorsMap := GenValidatorsFromPubKeys(pubKeys)
	nodesMap := make(map[uint32][]*TestProcessorNode)
	for shardId, validatorList := range validatorsMap {
		nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(
			shardConsensusGroupSize,
			metaConsensusGroupSize,
			TestHasher,
			shardId,
			uint32(nbShards),
			validatorsMap,
		)

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
) (data.BodyHandler, data.HeaderHandler, [][]byte) {

	nodesCoordinator := nodesMap[shardId][0].NodesCoordinator
	pubKeys, err := nodesCoordinator.GetValidatorsPublicKeys(randomness, round, shardId)
	if err != nil {
		fmt.Println("Error getting the validators public keys: ", err)
	}

	adddresses, err := nodesCoordinator.GetValidatorsRewardsAddresses(randomness, round, shardId)

	// set the consensus reward addresses
	for _, node := range nodesMap[shardId]{
		node.BlockProcessor.SetConsensusRewardAddresses(adddresses)
	}

	consensusNodes := selectTestNodesForPubKeys(nodesMap[shardId], pubKeys)
	// first node is block proposer
	body, header, txHashes := consensusNodes[0].ProposeBlock(round, nonce)
	header.SetPrevRandSeed(randomness)
	header = DoConsensusSigningOnBlock(header, consensusNodes, pubKeys)

	return body, header, txHashes
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
	blockHeaderBytes, _ := TestMarshalizer.Marshal(blockHeader)
	blockHeaderHash := TestHasher.Compute(string(blockHeaderBytes))

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

	return blockHeader
}
