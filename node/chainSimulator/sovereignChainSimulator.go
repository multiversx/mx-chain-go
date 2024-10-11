package chainSimulator

import (
	"fmt"
)

type sovereignChainSimulator struct {
	*simulator
}

// NewSovereignChainSimulator creates a sovereign chain simulator
func NewSovereignChainSimulator(args ArgsChainSimulator) (*sovereignChainSimulator, error) {
	cs, err := NewBaseChainSimulator(ArgsBaseChainSimulator{
		ArgsChainSimulator:          args,
		ConsensusGroupSize:          args.MinNodesPerShard,
		MetaChainConsensusGroupSize: 0,
	})
	if err != nil {
		return nil, err
	}

	return &sovereignChainSimulator{
		simulator: cs,
	}, nil
}

// GenerateBlocksUntilEpochIsReached will generate blocks until the epoch is reached
func (ss *sovereignChainSimulator) GenerateBlocksUntilEpochIsReached(targetEpoch int32) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	maxNumberOfRounds := 10000
	for idx := 0; idx < maxNumberOfRounds; idx++ {
		ss.incrementRoundOnAllValidators()
		err := ss.allNodesCreateBlocks()
		if err != nil {
			return err
		}

		if ss.isSovereignTargetEpochReached(targetEpoch) {
			return nil
		}
	}
	return fmt.Errorf("exceeded rounds to generate blocks")
}

func (ss *sovereignChainSimulator) isSovereignTargetEpochReached(targetEpoch int32) bool {
	for _, n := range ss.nodes {
		if int32(n.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch()) < targetEpoch {
			return false
		}
	}

	return true
}
