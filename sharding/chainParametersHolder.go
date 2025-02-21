package sharding

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
)

type chainParametersHolder struct {
	currentChainParameters  config.ChainParametersByEpochConfig
	chainParameters         []config.ChainParametersByEpochConfig
	chainParametersNotifier ChainParametersNotifierHandler
	mutOperations           sync.RWMutex
}

// ArgsChainParametersHolder holds the arguments needed for creating a new chainParametersHolder
type ArgsChainParametersHolder struct {
	EpochStartEventNotifier EpochStartEventNotifier
	ChainParameters         []config.ChainParametersByEpochConfig
	ChainParametersNotifier ChainParametersNotifierHandler
}

// NewChainParametersHolder returns a new instance of chainParametersHolder
func NewChainParametersHolder(args ArgsChainParametersHolder) (*chainParametersHolder, error) {
	err := validateArgs(args)
	if err != nil {
		return nil, err
	}

	chainParameters := args.ChainParameters
	// sort the config values in descending order
	sort.SliceStable(chainParameters, func(i, j int) bool {
		return chainParameters[i].EnableEpoch > chainParameters[j].EnableEpoch
	})

	earliestChainParams := chainParameters[len(chainParameters)-1]
	if earliestChainParams.EnableEpoch != 0 {
		return nil, ErrMissingConfigurationForEpochZero
	}

	paramsHolder := &chainParametersHolder{
		currentChainParameters:  earliestChainParams, // will be updated on the epoch notifier handlers
		chainParameters:         args.ChainParameters,
		chainParametersNotifier: args.ChainParametersNotifier,
	}
	args.ChainParametersNotifier.UpdateCurrentChainParameters(earliestChainParams)
	args.EpochStartEventNotifier.RegisterHandler(paramsHolder)

	logInitialConfiguration(args.ChainParameters)

	return paramsHolder, nil
}

func logInitialConfiguration(chainParameters []config.ChainParametersByEpochConfig) {
	logMessage := "initialized chainParametersHolder with the values:\n"
	logLines := make([]string, 0, len(chainParameters))
	for _, params := range chainParameters {
		logLines = append(logLines, fmt.Sprintf("\tenable epoch=%d, round duration=%d, hysteresis=%.2f, shard consensus group size=%d, shard min nodes=%d, meta consensus group size=%d, meta min nodes=%d, adaptivity=%v",
			params.EnableEpoch, params.RoundDuration, params.Hysteresis, params.ShardConsensusGroupSize, params.ShardMinNumNodes, params.MetachainConsensusGroupSize, params.MetachainMinNumNodes, params.Adaptivity))
	}

	logMessage += strings.Join(logLines, "\n")
	log.Debug(logMessage)
}

func validateArgs(args ArgsChainParametersHolder) error {
	if check.IfNil(args.EpochStartEventNotifier) {
		return ErrNilEpochStartEventNotifier
	}
	if len(args.ChainParameters) == 0 {
		return ErrMissingChainParameters
	}
	if check.IfNil(args.ChainParametersNotifier) {
		return ErrNilChainParametersNotifier
	}
	return validateChainParameters(args.ChainParameters)
}

func validateChainParameters(chainParametersConfig []config.ChainParametersByEpochConfig) error {
	for idx, chainParameters := range chainParametersConfig {
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

// EpochStartAction is called when a new epoch is confirmed
func (c *chainParametersHolder) EpochStartAction(header data.HeaderHandler) {
	c.handleEpochChange(header.GetEpoch())
}

// EpochStartPrepare is called when a new epoch is observed, but not yet confirmed. No action is required on this component
func (c *chainParametersHolder) EpochStartPrepare(_ data.HeaderHandler, _ data.BodyHandler) {
}

// NotifyOrder returns the notification order for a start of epoch event
func (c *chainParametersHolder) NotifyOrder() uint32 {
	return common.ChainParametersOrder
}

func (c *chainParametersHolder) handleEpochChange(epoch uint32) {
	c.mutOperations.Lock()
	defer c.mutOperations.Unlock()

	matchingVersionForNewEpoch, err := getMatchingChainParametersUnprotected(epoch, c.chainParameters)
	if err != nil {
		log.Error("chainParametersHolder.EpochConfirmed: cannot get matching chain parameters", "epoch", epoch, "error", err)
		return
	}
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
	c.chainParametersNotifier.UpdateCurrentChainParameters(matchingVersionForNewEpoch)
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

	chainParametersCopy := make([]config.ChainParametersByEpochConfig, len(c.chainParameters))
	copy(chainParametersCopy, c.chainParameters)

	return chainParametersCopy
}

// ChainParametersForEpoch will return the corresponding chain parameters for the provided epoch
func (c *chainParametersHolder) ChainParametersForEpoch(epoch uint32) (config.ChainParametersByEpochConfig, error) {
	c.mutOperations.RLock()
	defer c.mutOperations.RUnlock()

	return getMatchingChainParametersUnprotected(epoch, c.chainParameters)
}

func getMatchingChainParametersUnprotected(epoch uint32, configValues []config.ChainParametersByEpochConfig) (config.ChainParametersByEpochConfig, error) {
	for _, chainParams := range configValues {
		if chainParams.EnableEpoch <= epoch {
			return chainParams, nil
		}
	}

	// should never reach this code, as the config values are checked on the constructor
	return config.ChainParametersByEpochConfig{}, ErrNoMatchingConfigurationFound
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *chainParametersHolder) IsInterfaceNil() bool {
	return c == nil
}
