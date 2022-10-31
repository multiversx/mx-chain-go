package integrationTests

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	vic "github.com/ElrondNetwork/elrond-go/testscommon/validatorInfoCacher"
)

// ArgIndexHashedNodesCoordinatorFactory -
type ArgIndexHashedNodesCoordinatorFactory struct {
	nodesPerShard           int
	nbMetaNodes             int
	shardConsensusGroupSize int
	metaConsensusGroupSize  int
	shardId                 uint32
	nbShards                int
	validatorsMap           map[uint32][]nodesCoordinator.Validator
	waitingMap              map[uint32][]nodesCoordinator.Validator
	keyIndex                int
	cp                      *CryptoParams
	epochStartSubscriber    nodesCoordinator.EpochStartEventNotifier
	hasher                  hashing.Hasher
	consensusGroupCache     nodesCoordinator.Cacher
	bootStorer              storage.Storer
}

// IndexHashedNodesCoordinatorFactory -
type IndexHashedNodesCoordinatorFactory struct {
}

// CreateNodesCoordinator -
func (tpn *IndexHashedNodesCoordinatorFactory) CreateNodesCoordinator(arg ArgIndexHashedNodesCoordinatorFactory) nodesCoordinator.NodesCoordinator {

	keys := arg.cp.Keys[arg.shardId][arg.keyIndex]
	pubKeyBytes, _ := keys.Pk.ToByteArray()

	nodeShufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           uint32(arg.nodesPerShard),
		NodesMeta:            uint32(arg.nbMetaNodes),
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
		EnableEpochsHandler:  &testscommon.EnableEpochsHandlerStub{},
	}
	nodeShuffler, _ := nodesCoordinator.NewHashValidatorsShuffler(nodeShufflerArgs)
	argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
		ShardConsensusGroupSize: arg.shardConsensusGroupSize,
		MetaConsensusGroupSize:  arg.metaConsensusGroupSize,
		Marshalizer:             TestMarshalizer,
		Hasher:                  arg.hasher,
		Shuffler:                nodeShuffler,
		EpochStartNotifier:      arg.epochStartSubscriber,
		ShardIDAsObserver:       arg.shardId,
		NbShards:                uint32(arg.nbShards),
		EligibleNodes:           arg.validatorsMap,
		WaitingNodes:            arg.waitingMap,
		SelfPublicKey:           pubKeyBytes,
		ConsensusGroupCache:     arg.consensusGroupCache,
		BootStorer:              arg.bootStorer,
		ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
		ChanStopNode:            endProcess.GetDummyEndProcessChannel(),
		NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
		IsFullArchive:           false,
		EnableEpochsHandler: &testscommon.EnableEpochsHandlerStub{
			RefactorPeersMiniBlocksEnableEpochField: UnreachableEpoch,
		},
		ValidatorInfoCacher: &vic.ValidatorInfoCacherStub{},
	}
	nodesCoord, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
	if err != nil {
		fmt.Println("Error creating node coordinator")
	}

	return nodesCoord
}

// IndexHashedNodesCoordinatorWithRaterFactory -
type IndexHashedNodesCoordinatorWithRaterFactory struct {
	sharding.PeerAccountListAndRatingHandler
}

// CreateNodesCoordinator is used for creating a nodes coordinator in the integration tests
// based on the provided parameters
func (ihncrf *IndexHashedNodesCoordinatorWithRaterFactory) CreateNodesCoordinator(
	arg ArgIndexHashedNodesCoordinatorFactory,
) nodesCoordinator.NodesCoordinator {
	keys := arg.cp.Keys[arg.shardId][arg.keyIndex]
	pubKeyBytes, _ := keys.Pk.ToByteArray()

	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           uint32(arg.nodesPerShard),
		NodesMeta:            uint32(arg.nbMetaNodes),
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
		EnableEpochsHandler: &testscommon.EnableEpochsHandlerStub{
			IsWaitingListFixFlagEnabledField:      true,
			IsBalanceWaitingListsFlagEnabledField: true,
		},
	}
	nodeShuffler, _ := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	argumentsNodesCoordinator := nodesCoordinator.ArgNodesCoordinator{
		ShardConsensusGroupSize: arg.shardConsensusGroupSize,
		MetaConsensusGroupSize:  arg.metaConsensusGroupSize,
		Marshalizer:             TestMarshalizer,
		Hasher:                  arg.hasher,
		Shuffler:                nodeShuffler,
		EpochStartNotifier:      arg.epochStartSubscriber,
		ShardIDAsObserver:       arg.shardId,
		NbShards:                uint32(arg.nbShards),
		EligibleNodes:           arg.validatorsMap,
		WaitingNodes:            arg.waitingMap,
		SelfPublicKey:           pubKeyBytes,
		ConsensusGroupCache:     arg.consensusGroupCache,
		BootStorer:              arg.bootStorer,
		ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
		ChanStopNode:            endProcess.GetDummyEndProcessChannel(),
		NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
		IsFullArchive:           false,
		EnableEpochsHandler: &testscommon.EnableEpochsHandlerStub{
			IsWaitingListFixFlagEnabledField:        true,
			RefactorPeersMiniBlocksEnableEpochField: UnreachableEpoch,
		},
		ValidatorInfoCacher: &vic.ValidatorInfoCacherStub{},
	}

	baseCoordinator, err := nodesCoordinator.NewIndexHashedNodesCoordinator(argumentsNodesCoordinator)
	if err != nil {
		log.Debug("Error creating node coordinator")
	}

	nodesCoord, err := nodesCoordinator.NewIndexHashedNodesCoordinatorWithRater(baseCoordinator, ihncrf.PeerAccountListAndRatingHandler)
	if err != nil {
		log.Debug("Error creating node coordinator")
	}

	return &NodesWithRater{
		NodesCoordinator: nodesCoord,
		rater:            ihncrf.PeerAccountListAndRatingHandler,
	}
}

// NodesWithRater -
type NodesWithRater struct {
	nodesCoordinator.NodesCoordinator
	rater sharding.PeerAccountListAndRatingHandler
}

// IsInterfaceNil -
func (nwr *NodesWithRater) IsInterfaceNil() bool {
	return nwr == nil
}
