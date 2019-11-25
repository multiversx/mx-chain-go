package integrationTests

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"strconv"
)

const (
	validatorIncreaseRatingStep   = "validatorIncreaseRatingStep"
	validatorDecreaseRatingStep   = "validatorDecreaseRatingStep"
	proposerIncreaseRatingStepKey = "proposerIncreaseRatingStep"
	proposerDecreaseRatingStepKey = "proposerDecreaseRatingStep"
	minRating                     = uint32(1)
	maxRating                     = uint32(10)
)

// CreateNodesWithNodesCoordinator returns a map with nodes per shard each using a real nodes coordinator
func CreateNodesWithNodesCoordinatorWithRater(
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

		ratingValues := make(map[string]int32, 0)
		ratingValues[validatorIncreaseRatingStep] = 1
		ratingValues[validatorDecreaseRatingStep] = -2
		ratingValues[proposerIncreaseRatingStepKey] = 3
		ratingValues[proposerDecreaseRatingStepKey] = -4

		ratingsData, _ := economics.NewRatingsData(uint32(5), uint32(1), uint32(10), "mockRater", ratingValues)

		rater, _ := rating.NewBlockSigningRater(ratingsData)

		nodesCoordinator, err := sharding.NewIndexHashedNodesCoordinatorWithRater(argumentsNodesCoordinator, rater)

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
				rater,
			)
		}
		nodesMap[shardId] = nodesList
	}

	return nodesMap
}
