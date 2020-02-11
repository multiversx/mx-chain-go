package integrationTests

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	kmultisig "github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/multisig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

type nodeKeys struct {
	TxSignKeyGen     crypto.KeyGenerator
	TxSignSk         crypto.PrivateKey
	TxSignPk         crypto.PublicKey
	TxSignPkBytes    []byte
	BlockSignKeyGen  crypto.KeyGenerator
	BlockSignSk      crypto.PrivateKey
	BlockSignPk      crypto.PublicKey
	BlockSignPkBytes []byte
}

func pubKeysMapFromKeysMap(ncp map[uint32][]*nodeKeys) map[uint32][]string {
	keysMap := make(map[uint32][]string)

	for shardId, keys := range ncp {
		shardKeys := make([]string, len(keys))
		for i, nk := range keys {
			shardKeys[i] = string(nk.BlockSignPkBytes)
		}
		keysMap[shardId] = shardKeys
	}

	return keysMap
}

// CreateProcessorNodesWithNodesCoordinator creates a map of nodes with a valid nodes coordinator implementation
// keeping the consistency of generated keys
func CreateProcessorNodesWithNodesCoordinator(
	nodesPerShard int,
	nbMetaNodes int,
	nbShards int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	seedAddress string,
) map[uint32][]*TestProcessorNode {

	ncp := createNodesCryptoParams(nodesPerShard, nbMetaNodes, uint32(nbShards))
	validatorsMap := genValidators(ncp)
	nodesMap := make(map[uint32][]*TestProcessorNode)
	for shardId, validatorList := range validatorsMap {
		nodesList := make([]*TestProcessorNode, len(validatorList))
		for i, v := range validatorList {
			cache, _ := lrucache.NewCache(10000)
			argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
				ShardConsensusGroupSize: shardConsensusGroupSize,
				MetaConsensusGroupSize:  metaConsensusGroupSize,
				Hasher:                  TestHasher,
				ShardId:                 shardId,
				NbShards:                uint32(nbShards),
				Nodes:                   validatorsMap,
				SelfPublicKey:           v.Address(),
				ConsensusGroupCache:     cache,
			}

			nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
			if err != nil {
				fmt.Println("error creating node coordinator")
			}

			nodesList[i] = newTestProcessorNodeWithCustomNodesCoordinator(
				uint32(nbShards),
				shardId,
				seedAddress,
				nodesCoordinator,
				i,
				ncp,
			)
		}
		nodesMap[shardId] = nodesList
	}

	return nodesMap
}

func createTestSingleSigner() crypto.SingleSigner {
	return &singlesig.SchnorrSigner{}
}

func createNodesCryptoParams(nodesPerShard int, nbMetaNodes int, nbShards uint32) map[uint32][]*nodeKeys {
	suiteBlock := kyber.NewSuitePairingBn256()
	suiteTx := kyber.NewBlakeSHA256Ed25519()

	blockSignKeyGen := signing.NewKeyGenerator(suiteBlock)
	txSignKeyGen := signing.NewKeyGenerator(suiteTx)

	ncp := make(map[uint32][]*nodeKeys)
	for shard := uint32(0); shard < nbShards; shard++ {
		ncp[shard] = createShardNodeKeys(blockSignKeyGen, txSignKeyGen, nodesPerShard)
	}

	ncp[sharding.MetachainShardId] = createShardNodeKeys(blockSignKeyGen, txSignKeyGen, nbMetaNodes)

	return ncp
}

func createShardNodeKeys(
	blockSignKeyGen crypto.KeyGenerator,
	txSignKeyGen crypto.KeyGenerator,
	nodesPerShard int,
) []*nodeKeys {

	keys := make([]*nodeKeys, nodesPerShard)
	for i := 0; i < nodesPerShard; i++ {
		keys[i] = &nodeKeys{
			BlockSignKeyGen: blockSignKeyGen,
			TxSignKeyGen:    txSignKeyGen,
		}
		keys[i].TxSignSk, keys[i].TxSignPk = txSignKeyGen.GeneratePair()
		keys[i].BlockSignSk, keys[i].BlockSignPk = blockSignKeyGen.GeneratePair()

		keys[i].BlockSignPkBytes, _ = keys[i].BlockSignPk.ToByteArray()
		keys[i].TxSignPkBytes, _ = keys[i].TxSignPk.ToByteArray()
	}

	return keys
}

func genValidators(ncp map[uint32][]*nodeKeys) map[uint32][]sharding.Validator {
	validatorsMap := make(map[uint32][]sharding.Validator)

	for shardId, shardNodesKeys := range ncp {
		shardValidators := make([]sharding.Validator, 0)
		for i := 0; i < len(shardNodesKeys); i++ {
			v, _ := sharding.NewValidator(
				big.NewInt(0),
				1,
				shardNodesKeys[i].BlockSignPkBytes,
				shardNodesKeys[i].TxSignPkBytes)
			shardValidators = append(shardValidators, v)
		}
		validatorsMap[shardId] = shardValidators
	}

	return validatorsMap
}

func newTestProcessorNodeWithCustomNodesCoordinator(
	maxShards uint32,
	nodeShardId uint32,
	initialNodeAddr string,
	nodesCoordinator sharding.NodesCoordinator,
	keyIndex int,
	ncp map[uint32][]*nodeKeys,
) *TestProcessorNode {

	shardCoordinator, _ := sharding.NewMultiShardCoordinator(maxShards, nodeShardId)

	messenger := CreateMessengerWithKadDht(context.Background(), initialNodeAddr)
	tpn := &TestProcessorNode{
		ShardCoordinator:  shardCoordinator,
		Messenger:         messenger,
		NodesCoordinator:  nodesCoordinator,
		HeaderSigVerifier: &mock.HeaderSigVerifierStub{},
		ChainID:           ChainID,
	}

	tpn.NodeKeys = &TestKeyPair{
		Pk: ncp[nodeShardId][keyIndex].BlockSignPk,
		Sk: ncp[nodeShardId][keyIndex].BlockSignSk,
	}
	llsig := &kmultisig.KyberMultiSignerBLS{}
	blsHasher := blake2b.Blake2b{HashSize: factory.BlsHashSize}

	pubKeysMap := pubKeysMapFromKeysMap(ncp)
	kp := ncp[nodeShardId][keyIndex]
	var err error
	tpn.MultiSigner, err = multisig.NewBLSMultisig(
		llsig,
		blsHasher,
		pubKeysMap[nodeShardId],
		tpn.NodeKeys.Sk,
		kp.BlockSignKeyGen,
		uint16(keyIndex),
	)
	if err != nil {
		fmt.Printf("error generating multisigner: %s\n", err)
		return nil
	}

	tpn.OwnAccount = &TestWalletAccount{
		SingleSigner:      createTestSingleSigner(),
		BlockSingleSigner: createTestSingleSigner(),
		SkTxSign:          kp.TxSignSk,
		PkTxSign:          kp.TxSignPk,
		PkTxSignBytes:     kp.TxSignPkBytes,
		KeygenTxSign:      kp.TxSignKeyGen,
		KeygenBlockSign:   kp.BlockSignKeyGen,
		Nonce:             0,
		Balance:           nil,
	}
	tpn.OwnAccount.Address, _ = TestAddressConverter.CreateAddressFromPublicKeyBytes(kp.TxSignPkBytes)

	tpn.initDataPools()
	tpn.initTestNode()

	return tpn
}
