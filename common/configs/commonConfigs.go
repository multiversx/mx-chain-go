package configs

import (
	"errors"
	"sort"

	"github.com/multiversx/mx-chain-go/config"
)

const (
	defaultGracePeriodRounds                             = 25
	defaultExtraDelayForRequestBlockInfoInMs             = 3000
	defaultMaxRoundsWithoutCommittedStartInEpochBlock    = 50
	defaultNumRoundsToWaitBeforeSignalingChronologyStuck = 10
)

const expectedSubroundsTimingCount = 4

var defaultConsensusConfigByRound = config.ConsensusConfigByRound{
	EnableRound: 0,
	SubroundsTiming: []config.SubroundTiming{
		{StartTime: 0.0, EndTime: 0.05},
		{StartTime: 0.05, EndTime: 0.25},
		{StartTime: 0.25, EndTime: 0.85},
		{StartTime: 0.85, EndTime: 0.95},
	},
	ProcessingThresholdPercent: 85,
}

// ErrEmptyCommonConfigsByEpoch signals that an empty common configs by epoch has been provided
var ErrEmptyCommonConfigsByEpoch = errors.New("empty common configs by epoch")

// ErrEmptyCommonConfigsByRound signals that an empty common configs by round has been provided
var ErrEmptyCommonConfigsByRound = errors.New("empty common configs by round")

type commonConfigs struct {
	orderedEpochStartConfigByEpoch []config.EpochStartConfigByEpoch
	orderedEpochStartConfigByRound []config.EpochStartConfigByRound
	orderedConsensusConfigByEpoch  []config.ConsensusConfigByEpoch
	orderedConsensusConfigByRound  []config.ConsensusConfigByRound
}

// NewCommonConfigsHandler creates a new process configs by epoch component
func NewCommonConfigsHandler(
	configsByEpoch []config.EpochStartConfigByEpoch,
	configsByRound []config.EpochStartConfigByRound,
	consensusConfigByEpoch []config.ConsensusConfigByEpoch,
	consensusConfigByRound []config.ConsensusConfigByRound,
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

	err = checkConsensusConfigsByRound(consensusConfigByRound)
	if err != nil {
		return nil, err
	}

	cc := &commonConfigs{
		orderedEpochStartConfigByEpoch: make([]config.EpochStartConfigByEpoch, len(configsByEpoch)),
		orderedEpochStartConfigByRound: make([]config.EpochStartConfigByRound, len(configsByRound)),
		orderedConsensusConfigByEpoch:  make([]config.ConsensusConfigByEpoch, len(consensusConfigByEpoch)),
		orderedConsensusConfigByRound:  make([]config.ConsensusConfigByRound, len(consensusConfigByRound)),
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

	copy(cc.orderedConsensusConfigByRound, consensusConfigByRound)
	sort.SliceStable(cc.orderedConsensusConfigByRound, func(i, j int) bool {
		return cc.orderedConsensusConfigByRound[i].EnableRound < cc.orderedConsensusConfigByRound[j].EnableRound
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

func checkConsensusConfigsByRound(configsByRound []config.ConsensusConfigByRound) error {
	if len(configsByRound) == 0 {
		return ErrEmptyConsensusConfigsByRound
	}

	// check for duplicated rounds
	seen := make(map[uint64]struct{})
	for _, cfg := range configsByRound {
		_, exists := seen[cfg.EnableRound]
		if exists {
			return ErrDuplicatedRoundConfig
		}
		seen[cfg.EnableRound] = struct{}{}

		if err := checkSubroundsTiming(cfg); err != nil {
			return err
		}
	}

	_, exists := seen[0]
	if !exists {
		return ErrMissingRoundZeroConfig
	}

	return nil
}

func checkSubroundsTiming(cfg config.ConsensusConfigByRound) error {
	// the slice must contain exactly one entry per subround
	if len(cfg.SubroundsTiming) != expectedSubroundsTimingCount {
		return ErrInvalidSubroundsTimingCount
	}

	// all values must be non-negative and each subround must have start < end
	for _, t := range cfg.SubroundsTiming {
		if t.StartTime < 0 || t.EndTime < 0 {
			return ErrNegativeSubroundTiming
		}
		if t.StartTime >= t.EndTime {
			return ErrInvalidSubroundTimingRange
		}
	}

	// subrounds must be ordered, contiguous (no gaps), and non-overlapping
	for i := 1; i < len(cfg.SubroundsTiming); i++ {
		if cfg.SubroundsTiming[i].StartTime != cfg.SubroundsTiming[i-1].EndTime {
			return ErrOverlappingSubroundTiming
		}
	}

	// all boundary values must be < 1.0
	for _, t := range cfg.SubroundsTiming {
		if t.StartTime >= 1.0 || t.EndTime >= 1.0 {
			return ErrSubroundTimingExceedsRound
		}
	}

	// ProcessingThresholdPercent must be in (0, 100]
	if cfg.ProcessingThresholdPercent == 0 || cfg.ProcessingThresholdPercent > 100 {
		return ErrInvalidProcessingThreshold
	}

	return nil
}

// GetGracePeriodRoundsByEpoch returns the grace period rounds by epoch
func (cc *commonConfigs) GetGracePeriodRoundsByEpoch(epoch uint32) uint32 {
	for i := len(cc.orderedEpochStartConfigByEpoch) - 1; i >= 0; i-- {
		if cc.orderedEpochStartConfigByEpoch[i].EnableEpoch <= epoch {
			return cc.orderedEpochStartConfigByEpoch[i].GracePeriodRounds
		}
	}

	return defaultGracePeriodRounds // this should not happen
}

// GetExtraDelayForRequestBlockInfoInMs returns the extra delay for request block info by epoch
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

// GetSubroundsTimingByRound returns the subrounds timing configuration active for the given round
func (cc *commonConfigs) GetSubroundsTimingByRound(round uint64) config.ConsensusConfigByRound {
	for i := len(cc.orderedConsensusConfigByRound) - 1; i >= 0; i-- {
		if cc.orderedConsensusConfigByRound[i].EnableRound <= round {
			return cc.orderedConsensusConfigByRound[i]
		}
	}

	return defaultConsensusConfigByRound // this should not happen
}

// GetActiveTimingBoundaryRound returns the EnableRound of the ConsensusConfigByRound entry that is active for the given round
func (cc *commonConfigs) GetActiveTimingBoundaryRound(round uint64) uint64 {
	for i := len(cc.orderedConsensusConfigByRound) - 1; i >= 0; i-- {
		if cc.orderedConsensusConfigByRound[i].EnableRound <= round {
			return cc.orderedConsensusConfigByRound[i].EnableRound
		}
	}

	return 0
}

// IsInterfaceNil checks if the instance is nil
func (cc *commonConfigs) IsInterfaceNil() bool {
	return cc == nil
}
