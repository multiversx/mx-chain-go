package integrationTests

import (
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// TestBootstrapper extends the Bootstrapper interface with some functions intended to be used only in tests
// as it simplifies the reproduction of edge cases
type TestBootstrapper interface {
	process.Bootstrapper
	RollBack(revertUsingForkNonce bool) error
	SetProbableHighestNonce(nonce uint64)
}

// TestEpochStartTrigger extends the epochStart trigger interface with some functions intended to by used only
// in tests as it simplifies the reproduction of test scenarios
type TestEpochStartTrigger interface {
	epochStart.TriggerHandler
	GetRoundsPerEpoch() uint64
	SetTrigger(triggerHandler epochStart.TriggerHandler)
	SetRoundsPerEpoch(roundsPerEpoch uint64)
}

type NodesCoordinatorFactory interface {
	CreateNodesCoordinator(
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
	) sharding.NodesCoordinator
}
