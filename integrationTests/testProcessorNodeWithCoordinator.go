package integrationTests

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/shardingmock"
	vic "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
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

// CreateProcessorNodesWithNodesCoordinator creates a map of nodes with a valid nodes coordinator implementation
// keeping the consistency of generated keys
func CreateProcessorNodesWithNodesCoordinator(
	rewardsAddrsAssignments map[uint32][]uint32,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
) (map[uint32][]*TestProcessorNode, uint32) {

	ncp, nbShards := createNodesCryptoParams(rewardsAddrsAssignments)
	cp := CreateCryptoParams(len(ncp[0]), len(ncp[core.MetachainShardId]), nbShards, 1)
	pubKeys := PubKeysMapFromNodesKeysMap(cp.NodesKeys)
	validatorsMap := GenValidatorsFromPubKeys(pubKeys, nbShards)
	validatorsMapForNodesCoordinator, _ := nodesCoordinator.NodesInfoToValidators(validatorsMap)

	cpWaiting := CreateCryptoParams(1, 1, nbShards, 1)
	pubKeysWaiting := PubKeysMapFromNodesKeysMap(cpWaiting.NodesKeys)
	waitingMap := GenValidatorsFromPubKeys(pubKeysWaiting, nbShards)
	waitingMapForNodesCoordinator, _ := nodesCoordinator.NodesInfoToValidators(waitingMap)

	nodesSetup := &mock.NodesSetupStub{InitialNodesInfoCalled: func() (m map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, m2 map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
		return validatorsMap, waitingMap
	}}

	ncp, numShards := createNodesCryptoParams(rewardsAddrsAssignments)

	completeNodesList := make([]Connectable, 0)
	nodesMap := make(map[uint32][]*TestProcessorNode)
	for shardId, validatorList := range validatorsMap {
		nodesList := make([]*TestProcessorNode, len(validatorList))
		for i, v := range validatorList {
			lruCache, _ := cache.NewLRUCache(10000)
			argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
				ChainParametersHandler: &shardingmock.ChainParametersHandlerStub{
					ChainParametersForEpochCalled: func(_ uint32) (config.ChainParametersByEpochConfig, error) {
						return config.ChainParametersByEpochConfig{
							ShardConsensusGroupSize:     uint32(shardConsensusGroupSize),
							MetachainConsensusGroupSize: uint32(metaConsensusGroupSize),
						}, nil
					},
				},
				Marshalizer:         TestMarshalizer,
				Hasher:              TestHasher,
				ShardIDAsObserver:   shardId,
				NbShards:            numShards,
				EligibleNodes:       validatorsMapForNodesCoordinator,
				WaitingNodes:        waitingMapForNodesCoordinator,
				SelfPublicKey:       v.PubKeyBytes(),
				ConsensusGroupCache: lruCache,
				ShuffledOutHandler:  &mock.ShuffledOutHandlerStub{},
				ChanStopNode:        endProcess.GetDummyEndProcessChannel(),
				IsFullArchive:       false,
				EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
				ValidatorInfoCacher: &vic.ValidatorInfoCacherStub{},
			}

			nodesCoordinatorInstance, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
			if err != nil {
				fmt.Println("error creating node coordinator")
			}

			multiSigner, err := createMultiSigner(*cp)
			if err != nil {
				log.Error("error generating multisigner: %s\n", err)
				return nil, 0
			}

			kp := ncp[shardId][i]

			ownAccount := &TestWalletAccount{
				SingleSigner:      TestSingleSigner,
				BlockSingleSigner: TestSingleSigner,
				SkTxSign:          kp.TxSignSk,
				PkTxSign:          kp.TxSignPk,
				PkTxSignBytes:     kp.TxSignPkBytes,
				KeygenTxSign:      kp.TxSignKeyGen,
				KeygenBlockSign:   kp.BlockSignKeyGen,
				Nonce:             0,
				Balance:           nil,
			}
			ownAccount.Address = kp.TxSignPkBytes

			nodesList[i] = NewTestProcessorNode(ArgTestProcessorNode{
				MaxShards:            numShards,
				NodeShardId:          shardId,
				TxSignPrivKeyShardId: shardId,
				NodeKeys: &TestNodeKeys{
					MainKey: &TestKeyPair{
						Sk: kp.BlockSignSk,
						Pk: kp.BlockSignPk,
					},
				},
				NodesSetup:       nodesSetup,
				NodesCoordinator: nodesCoordinatorInstance,
				MultiSigner:      multiSigner,
				OwnAccount:       ownAccount,
			})

			completeNodesList = append(completeNodesList, nodesList[i])
		}
		nodesMap[shardId] = nodesList
	}

	ConnectNodes(completeNodesList)

	return nodesMap, numShards
}

func createNodesCryptoParams(rewardsAddrsAssignments map[uint32][]uint32) (map[uint32][]*nodeKeys, uint32) {
	numShards := uint32(0)
	suiteBlock := mcl.NewSuiteBLS12()
	suiteTx := ed25519.NewEd25519()

	blockSignKeyGen := signing.NewKeyGenerator(suiteBlock)
	txSignKeyGen := signing.NewKeyGenerator(suiteTx)

	// we need to first precompute the num shard ID
	for shardID := range rewardsAddrsAssignments {
		foundAHigherShardID := shardID != core.MetachainShardId && shardID > numShards
		if foundAHigherShardID {
			numShards = shardID
		}
	}
	// we need to increment this as the numShards is actually the max shard at this moment
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
		if shardCoordinator.ComputeId(pkBytes) == addrShardID {
			break
		}
		sk, pk = keyGen.GeneratePair()
	}

	return sk, pk
}
