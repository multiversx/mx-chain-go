package integrationTests

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type IndexHashedNodesCoordinatorFactory struct {
}

func (tpn *IndexHashedNodesCoordinatorFactory) CreateNodesCoordinator(
	nodesPerShard int,
	nbMetaNodes int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	shardId uint32,
	nbShards int,
	validatorsMap map[uint32][]sharding.Validator,
	waitingMap map[uint32][]sharding.Validator,
	keyIndex int,
	cp *CryptoParams,
	epochStartSubscriber sharding.EpochStartSubscriber,
	hasher hashing.Hasher,
) sharding.NodesCoordinator {
	nodeKeys := cp.Keys[shardId][keyIndex]
	pubKeyBytes, _ := nodeKeys.Pk.ToByteArray()

	nodeShuffler := sharding.NewXorValidatorsShuffler(uint32(nodesPerShard), uint32(nbMetaNodes), 0.2, false)
	argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: shardConsensusGroupSize,
		MetaConsensusGroupSize:  metaConsensusGroupSize,
		Hasher:                  hasher,
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		ShardId:                 shardId,
		NbShards:                uint32(nbShards),
		EligibleNodes:           validatorsMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           pubKeyBytes,
	}
	nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)

	if err != nil {
		fmt.Println("Error creating node coordinator")
	}

	return nodesCoordinator
}

type IndexHashedNodesCoordinatorWithRaterFactory struct {
	sharding.RaterHandler
}

func (ihncrf *IndexHashedNodesCoordinatorWithRaterFactory) CreateNodesCoordinator(
	nodesPerShard int,
	nbMetaNodes int,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	shardId uint32,
	nbShards int,
	validatorsMap map[uint32][]sharding.Validator,
	waitingMap map[uint32][]sharding.Validator,
	keyIndex int,
	cp *CryptoParams,
	epochStartSubscriber sharding.EpochStartSubscriber,
	hasher hashing.Hasher,
) sharding.NodesCoordinator {
	nodeKeys := cp.Keys[shardId][keyIndex]
	pubKeyBytes, _ := nodeKeys.Pk.ToByteArray()

	nodeShuffler := sharding.NewXorValidatorsShuffler(uint32(nodesPerShard), uint32(nbMetaNodes), 0.2, false)
	argumentsNodesCoordinator := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: shardConsensusGroupSize,
		MetaConsensusGroupSize:  metaConsensusGroupSize,
		Hasher:                  hasher,
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		ShardId:                 shardId,
		NbShards:                uint32(nbShards),
		EligibleNodes:           validatorsMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           pubKeyBytes,
	}

	baseCoordinator, err := sharding.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)

	if err != nil {
		fmt.Println("Error creating node coordinator")
	}

	nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinatorWithRater(baseCoordinator, ihncrf.RaterHandler)

	if err != nil {
		fmt.Println("Error creating node coordinator")
	}

	return &NodesWithRater{
		NodesCoordinator: nodesCoordinator,
		RaterHandler:     ihncrf.RaterHandler,
	}
}

type NodesWithRater struct {
	sharding.NodesCoordinator
	sharding.RaterHandler
}

func (nwr *NodesWithRater) IsInterfaceNil() bool {
	return nwr == nil
}
