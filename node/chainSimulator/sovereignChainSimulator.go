package chainSimulator

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
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

// ForceResetValidatorStatisticsCache will force the reset of the cache used for the validators statistics endpoint
func (ss *sovereignChainSimulator) ForceResetValidatorStatisticsCache() error {
	return ss.GetNodeHandler(core.SovereignChainShardId).GetProcessComponents().ValidatorsProvider().ForceUpdate()
}

// ForceChangeOfEpoch will force the change of current epoch
// This method will call the epoch change trigger and generate block till a new epoch is reached
func (ss *sovereignChainSimulator) ForceChangeOfEpoch() error {
	log.Info("force change of epoch")

	ss.mutex.Lock()

	node := ss.nodes[core.SovereignChainShardId]
	err := node.ForceChangeOfEpoch()
	if err != nil {
		ss.mutex.Unlock()
		return fmt.Errorf("force change of epoch error-%w", err)
	}

	epoch := node.GetProcessComponents().EpochStartTrigger().Epoch()
	ss.mutex.Unlock()

	return ss.GenerateBlocksUntilEpochIsReached(int32(epoch + 1))
}
