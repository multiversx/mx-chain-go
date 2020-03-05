package integrationTests

import (
	"context"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	kmultisig "github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/multisig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go/hashing"
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
	rewardsAddrsAssignments map[uint32][]uint32,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	seedAddress string,
) (map[uint32][]*TestProcessorNode, uint32) {

	ncp, nbShards := createNodesCryptoParams(rewardsAddrsAssignments)
	cp := CreateCryptoParams(len(ncp[0]), len(ncp[core.MetachainShardId]), nbShards)
	pubKeys := PubKeysMapFromKeysMap(cp.Keys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, nbShards)

	cpWaiting := CreateCryptoParams(1, 1, nbShards)
	pubKeysWaiting := PubKeysMapFromKeysMap(cpWaiting.Keys)
	waitingMap := GenValidatorsFromPubKeys(pubKeysWaiting, nbShards)

	ncp, numShards := createNodesCryptoParams(rewardsAddrsAssignments)

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
				NbShards:                numShards,
				EligibleNodes:           validatorsMap,
				WaitingNodes:            waitingMap,
				SelfPublicKey:           v.Address(),
				ConsensusGroupCache:     cache,
			}

			nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
			if err != nil {
				fmt.Println("error creating node coordinator")
			}

			nodesList[i] = newTestProcessorNodeWithCustomNodesCoordinator(
				numShards,
				shardId,
				seedAddress,
				nodesCoordinator,
				i,
				ncp,
			)
		}
		nodesMap[shardId] = nodesList
	}

	return nodesMap, numShards
}

func createTestSingleSigner() crypto.SingleSigner {
	return &singlesig.SchnorrSigner{}
}

func createNodesCryptoParams(rewardsAddrsAssignments map[uint32][]uint32) (map[uint32][]*nodeKeys, uint32) {
	numShards := uint32(0)
	suiteBlock := kyber.NewSuitePairingBn256()
	suiteTx := kyber.NewBlakeSHA256Ed25519()

	blockSignKeyGen := signing.NewKeyGenerator(suiteBlock)
	txSignKeyGen := signing.NewKeyGenerator(suiteTx)

	//we need to first precompute the num shard ID
	for shardID := range rewardsAddrsAssignments {
		foundAHigherShardID := shardID != core.MetachainShardId && shardID > numShards
		if foundAHigherShardID {
			numShards = shardID
		}
	}
	//we need to increment this as the numShards is actually the max shard at this moment
	numShards++

	ncp := make(map[uint32][]*nodeKeys)
	for shardID, assignments := range rewardsAddrsAssignments {
		ncp[shardID] = createShardNodeKeys(blockSignKeyGen, txSignKeyGen, assignments, shardID, numShards)
	}

	return ncp, numShards
}

func createShardNodeKeys(
	blockSignKeyGen crypto.KeyGenerator,
	txSignKeyGen crypto.KeyGenerator,
	assignments []uint32,
	shardID uint32,
	numShards uint32,
) []*nodeKeys {
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(numShards, shardID)
	keys := make([]*nodeKeys, len(assignments))
	for i, addrShardID := range assignments {
		keys[i] = &nodeKeys{
			BlockSignKeyGen: blockSignKeyGen,
			TxSignKeyGen:    txSignKeyGen,
		}
		keys[i].TxSignSk, keys[i].TxSignPk = generateSkAndPkInShard(keys[i].TxSignKeyGen, shardCoordinator, addrShardID)
		keys[i].BlockSignSk, keys[i].BlockSignPk = blockSignKeyGen.GeneratePair()

		keys[i].BlockSignPkBytes, _ = keys[i].BlockSignPk.ToByteArray()
		keys[i].TxSignPkBytes, _ = keys[i].TxSignPk.ToByteArray()
	}

	return keys
}

func generateSkAndPkInShard(
	keyGen crypto.KeyGenerator,
	shardCoordinator sharding.Coordinator,
	addrShardID uint32,
) (crypto.PrivateKey, crypto.PublicKey) {
	sk, pk := keyGen.GeneratePair()
	for {
		pkBytes, _ := pk.ToByteArray()
		addr, _ := TestAddressConverter.CreateAddressFromPublicKeyBytes(pkBytes)
		if shardCoordinator.ComputeId(addr) == addrShardID {
			break
		}
		sk, pk = keyGen.GeneratePair()
	}

	return sk, pk
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
	blsHasher := &blake2b.Blake2b{HashSize: hashing.BlsHashSize}

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
