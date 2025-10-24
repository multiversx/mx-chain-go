package configs

import (
	"errors"
	"sort"

	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const (
	defaultGracePeriodRounds                             = 25
	defaultExtraDelayForRequestBlockInfoInMs             = 3000
	defaultMaxRoundsWithoutCommittedStartInEpochBlock    = 50
	defaultNumRoundsToWaitBeforeSignalingChronologyStuck = 10
)

// ErrEmptyCommonConfigsByEpoch signals that an empty common configs by epoch has been provided
var ErrEmptyCommonConfigsByEpoch = errors.New("empty common configs by epoch")

// ErrEmptyCommonConfigsByRound signals that an empty common configs by round has been provided
var ErrEmptyCommonConfigsByRound = errors.New("empty common configs by epoch")

type commonConfigs struct {
	orderedEpochStartConfigByEpoch []config.EpochStartConfigByEpoch
	orderedEpochStartConfigByRound []config.EpochStartConfigByRound
	orderedConsensusConfigByEpoch  []config.ConsensusConfigByEpoch
}

// NewCommonConfigsHandler creates a new process configs by epoch component
func NewCommonConfigsHandler(
	configsByEpoch []config.EpochStartConfigByEpoch,
	configsByRound []config.EpochStartConfigByRound,
	consensusConfigByEpoch []config.ConsensusConfigByEpoch,
) (*commonConfigs, error) {
	err := checkCommonConfigsByEpoch(configsByEpoch)
	if err != nil {
		return nil, err
	}
	err = checkConsensusConfigsByEpoch(consensusConfigByEpoch)
	if err != nil {
		return nil, err
	}

	err = checkCommonConfigsByRound(configsByRound)
	if err != nil {
		return nil, err
	}

	cc := &commonConfigs{
		orderedEpochStartConfigByEpoch: make([]config.EpochStartConfigByEpoch, len(configsByEpoch)),
		orderedEpochStartConfigByRound: make([]config.EpochStartConfigByRound, len(configsByRound)),
		orderedConsensusConfigByEpoch:  make([]config.ConsensusConfigByEpoch, len(consensusConfigByEpoch)),
	}

	// sort the config values in ascending order
	copy(cc.orderedEpochStartConfigByEpoch, configsByEpoch)
	sort.SliceStable(cc.orderedEpochStartConfigByEpoch, func(i, j int) bool {
		return cc.orderedEpochStartConfigByEpoch[i].EnableEpoch < cc.orderedEpochStartConfigByEpoch[j].EnableEpoch
	})

	copy(cc.orderedEpochStartConfigByRound, configsByRound)
	sort.SliceStable(cc.orderedEpochStartConfigByRound, func(i, j int) bool {
		return cc.orderedEpochStartConfigByRound[i].EnableRound < cc.orderedEpochStartConfigByRound[j].EnableRound
	})

	copy(cc.orderedConsensusConfigByEpoch, consensusConfigByEpoch)
	sort.SliceStable(cc.orderedConsensusConfigByEpoch, func(i, j int) bool {
		return cc.orderedConsensusConfigByEpoch[i].EnableEpoch < cc.orderedConsensusConfigByEpoch[j].EnableEpoch
	})

	return cc, nil
}

func checkCommonConfigsByEpoch(configsByEpoch []config.EpochStartConfigByEpoch) error {
	if len(configsByEpoch) == 0 {
		return ErrEmptyCommonConfigsByEpoch
	}

	// check for duplicated configs
	seen := make(map[uint32]struct{})
	for _, cfg := range configsByEpoch {
		_, exists := seen[cfg.EnableEpoch]
		if exists {
			return ErrDuplicatedEpochConfig
		}
		seen[cfg.EnableEpoch] = struct{}{}
	}

	_, exists := seen[0]
	if !exists {
		return ErrMissingEpochZeroConfig
	}

	return nil
}

func checkCommonConfigsByRound(configsByRound []config.EpochStartConfigByRound) error {
	if len(configsByRound) == 0 {
		return ErrEmptyCommonConfigsByRound
	}

	// check for duplicated configs
	seen := make(map[uint64]struct{})
	for _, cfg := range configsByRound {
		_, exists := seen[cfg.EnableRound]
		if exists {
			return ErrDuplicatedRoundConfig
		}
		seen[cfg.EnableRound] = struct{}{}
	}

	_, exists := seen[0]
	if !exists {
		return ErrMissingRoundZeroConfig
	}

	return nil
}

func checkConsensusConfigsByEpoch(configsByEpoch []config.ConsensusConfigByEpoch) error {
	if len(configsByEpoch) == 0 {
		return ErrEmptyCommonConfigsByEpoch
	}

	// check for duplicated configs
	seen := make(map[uint32]struct{})
	for _, cfg := range configsByEpoch {
		_, exists := seen[cfg.EnableEpoch]
		if exists {
			return ErrDuplicatedEpochConfig
		}
		seen[cfg.EnableEpoch] = struct{}{}
	}

	_, exists := seen[0]
	if !exists {
		return ErrMissingEpochZeroConfig
	}

	return nil
}

// GetMaxMetaNoncesBehind returns the max meta nonces behind by epoch
func (cc *commonConfigs) GetGracePeriodRoundsByEpoch(epoch uint32) uint32 {
	for i := len(cc.orderedEpochStartConfigByEpoch) - 1; i >= 0; i-- {
		if cc.orderedEpochStartConfigByEpoch[i].EnableEpoch <= epoch {
			return cc.orderedEpochStartConfigByEpoch[i].GracePeriodRounds
		}
	}

	return defaultGracePeriodRounds // this should not happen
}

// GetExtraDelayForRequestBlockInfoInMs returns the max meta nonces behind by epoch
func (cc *commonConfigs) GetExtraDelayForRequestBlockInfoInMs(epoch uint32) uint32 {
	for i := len(cc.orderedEpochStartConfigByEpoch) - 1; i >= 0; i-- {
		if cc.orderedEpochStartConfigByEpoch[i].EnableEpoch <= epoch {
			return cc.orderedEpochStartConfigByEpoch[i].ExtraDelayForRequestBlockInfoInMilliseconds
		}
	}

	return defaultExtraDelayForRequestBlockInfoInMs // this should not happen
}

// GetMaxRoundsWithoutCommittedStartInEpochBlockInRound returns max rounds without commited start in epoch block
func (cc *commonConfigs) GetMaxRoundsWithoutCommittedStartInEpochBlockInRound(round uint64) uint32 {
	for i := len(cc.orderedEpochStartConfigByRound) - 1; i >= 0; i-- {
		if cc.orderedEpochStartConfigByRound[i].EnableRound <= round {
			return cc.orderedEpochStartConfigByRound[i].MaxRoundsWithoutCommittedStartInEpochBlock
		}
	}

	return defaultMaxRoundsWithoutCommittedStartInEpochBlock // this should not happen
}

// GetNumRoundsToWaitBeforeSignalingChronologyStuck returns number of rounds to wait before signaling chronology stuck
func (cc *commonConfigs) GetNumRoundsToWaitBeforeSignalingChronologyStuck(epoch uint32) uint32 {
	for i := len(cc.orderedConsensusConfigByEpoch) - 1; i >= 0; i-- {
		if cc.orderedConsensusConfigByEpoch[i].EnableEpoch <= epoch {
			return cc.orderedConsensusConfigByEpoch[i].NumRoundsToWaitBeforeSignalingChronologyStuck
		}
	}

	return defaultNumRoundsToWaitBeforeSignalingChronologyStuck // this should not happen
}

// IsInterfaceNil checks if the instance is nil
func (cc *commonConfigs) IsInterfaceNil() bool {
	return cc == nil
}

// SetActivationRound -
func (cc *commonConfigs) SetActivationRound(round uint64, log logger.Logger) {
	nr := len(cc.orderedEpochStartConfigByRound)
	log.Info("commonConfigs.SetActivationRound", "enableRound", round, "oldRound", cc.orderedEpochStartConfigByRound[nr-1].EnableRound)
	cc.orderedEpochStartConfigByRound[nr-1].EnableRound = round
}
