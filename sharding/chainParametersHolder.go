package sharding

import (
	"fmt"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/config"
)

type chainParametersHolder struct {
	currentChainParameters config.ChainParametersByEpochConfig
	chainParameters        []config.ChainParametersByEpochConfig
	mutOperations          sync.RWMutex
}

// ArgsChainParametersHolder holds the arguments needed for creating a new chainParametersHolder
type ArgsChainParametersHolder struct {
	EpochNotifier   EpochNotifier
	ChainParameters []config.ChainParametersByEpochConfig
}

// NewChainParametersHolder returns a new instance of chainParametersHolder
func NewChainParametersHolder(args ArgsChainParametersHolder) (*chainParametersHolder, error) {
	err := validateArgs(args)
	if err != nil {
		return nil, err
	}

	chainParameters := args.ChainParameters
	sort.SliceStable(chainParameters, func(i, j int) bool {
		return chainParameters[i].EnableEpoch < chainParameters[j].EnableEpoch
	})

	paramsHolder := &chainParametersHolder{
		currentChainParameters: args.ChainParameters[0],
		chainParameters:        args.ChainParameters,
	}

	args.EpochNotifier.RegisterNotifyHandler(paramsHolder)

	return paramsHolder, nil
}

func validateArgs(args ArgsChainParametersHolder) error {
	if check.IfNil(args.EpochNotifier) {
		return ErrNilEpochNotifier
	}

	if len(args.ChainParameters) == 0 {
		return ErrMissingChainParameters
	}

	for idx, chainParameters := range args.ChainParameters {
		if chainParameters.ShardConsensusGroupSize < 1 {
			return fmt.Errorf("%w for chain parameters with index %d", ErrNegativeOrZeroConsensusGroupSize, idx)
		}
		if chainParameters.ShardMinNumNodes < chainParameters.ShardConsensusGroupSize {
			return fmt.Errorf("%w for chain parameters with index %d", ErrMinNodesPerShardSmallerThanConsensusSize, idx)
		}
		if chainParameters.MetachainConsensusGroupSize < 1 {
			return fmt.Errorf("%w for chain parameters with index %d", ErrNegativeOrZeroConsensusGroupSize, idx)
		}
		if chainParameters.MetachainMinNumNodes < chainParameters.MetachainConsensusGroupSize {
			return fmt.Errorf("%w for chain parameters with index %d", ErrMinNodesPerShardSmallerThanConsensusSize, idx)
		}
	}

	return nil
}

// EpochConfirmed is called at each epoch change event
func (c *chainParametersHolder) EpochConfirmed(epoch uint32, _ uint64) {
	c.mutOperations.Lock()
	defer c.mutOperations.Unlock()

	matchingVersionForNewEpoch := c.getMatchingChainParametersUnprotected(epoch)
	if matchingVersionForNewEpoch.EnableEpoch == c.currentChainParameters.EnableEpoch {
		return
	}

	c.currentChainParameters = matchingVersionForNewEpoch
	log.Debug("updated chainParametersHolder current chain parameters",
		"round duration", matchingVersionForNewEpoch.RoundDuration,
		"shard consensus group size", matchingVersionForNewEpoch.ShardConsensusGroupSize,
		"shard min num nodes", matchingVersionForNewEpoch.ShardMinNumNodes,
		"metachain consensus group size", matchingVersionForNewEpoch.MetachainConsensusGroupSize,
		"metachain min num nodes", matchingVersionForNewEpoch.MetachainMinNumNodes,
		"shard consensus group size", matchingVersionForNewEpoch.ShardConsensusGroupSize,
		"hysteresis", matchingVersionForNewEpoch.Hysteresis,
		"adaptivity", matchingVersionForNewEpoch.Adaptivity,
	)
}

// CurrentChainParameters will return the chain parameters that are active at the moment of calling
func (c *chainParametersHolder) CurrentChainParameters() config.ChainParametersByEpochConfig {
	c.mutOperations.RLock()
	defer c.mutOperations.RUnlock()

	return c.currentChainParameters
}

// AllChainParameters will return the entire slice of chain parameters configuration
func (c *chainParametersHolder) AllChainParameters() []config.ChainParametersByEpochConfig {
	c.mutOperations.RLock()
	defer c.mutOperations.RUnlock()

	return c.chainParameters
}

// ChainParametersForEpoch will return the corresponding chain parameters for the provided epoch
func (c *chainParametersHolder) ChainParametersForEpoch(epoch uint32) config.ChainParametersByEpochConfig {
	c.mutOperations.RLock()
	defer c.mutOperations.RUnlock()

	return c.getMatchingChainParametersUnprotected(epoch)
}

func (c *chainParametersHolder) getMatchingChainParametersUnprotected(epoch uint32) config.ChainParametersByEpochConfig {
	chainParametersForEpoch := c.chainParameters[0]
	for _, chainParams := range c.chainParameters {
		if chainParams.EnableEpoch > epoch {
			break
		}

		chainParametersForEpoch = chainParams
	}

	return chainParametersForEpoch
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *chainParametersHolder) IsInterfaceNil() bool {
	return c == nil
}
