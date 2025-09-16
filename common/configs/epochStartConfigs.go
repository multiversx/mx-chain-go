package configs

import (
	"errors"
	"sort"

	"github.com/multiversx/mx-chain-go/config"
)

const (
	defaultGracePeriodRounds                 = 25
	defaultExtraDelayForRequestBlockInfoInMs = 3000
)

// ErrEmptyCommonConfigsByEpoch signals that an empty process configs by epoch has been provided
var ErrEmptyCommonConfigsByEpoch = errors.New("empty process configs by epoch")

// ErrEmptyCommonConfigsByEpoch signals that an empty process configs by round has been provided
var ErrEmptyCommonConfigsByRound = errors.New("empty process configs by round")

type epochStartConfigs struct {
	orderedConfigByEpoch []config.EpochStartConfigByEpoch
}

// NewCommonConfigsHandler creates a new process configs by epoch component
func NewCommonConfigsHandler(
	configsByEpoch []config.EpochStartConfigByEpoch,
) (*epochStartConfigs, error) {
	err := checkCommonConfigsByEpoch(configsByEpoch)
	if err != nil {
		return nil, err
	}

	esc := &epochStartConfigs{
		orderedConfigByEpoch: make([]config.EpochStartConfigByEpoch, len(configsByEpoch)),
	}

	// sort the config values in ascending order
	copy(esc.orderedConfigByEpoch, configsByEpoch)
	sort.SliceStable(esc.orderedConfigByEpoch, func(i, j int) bool {
		return esc.orderedConfigByEpoch[i].EnableEpoch < esc.orderedConfigByEpoch[j].EnableEpoch
	})

	return esc, nil
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

// GetMaxMetaNoncesBehind returns the max meta nonces behind by epoch
func (esc *epochStartConfigs) GetGracePeriodRoundsByEpoch(epoch uint32) uint32 {
	for i := len(esc.orderedConfigByEpoch) - 1; i >= 0; i-- {
		if esc.orderedConfigByEpoch[i].EnableEpoch <= epoch {
			return esc.orderedConfigByEpoch[i].GracePeriodRounds
		}
	}

	return defaultGracePeriodRounds // this should not happen
}

// GetExtraDelayForRequestBlockInfoInMs returns the max meta nonces behind by epoch
func (esc *epochStartConfigs) GetExtraDelayForRequestBlockInfoInMs(epoch uint32) uint32 {
	for i := len(esc.orderedConfigByEpoch) - 1; i >= 0; i-- {
		if esc.orderedConfigByEpoch[i].EnableEpoch <= epoch {
			return esc.orderedConfigByEpoch[i].ExtraDelayForRequestBlockInfoInMilliseconds
		}
	}

	return defaultExtraDelayForRequestBlockInfoInMs // this should not happen
}

// IsInterfaceNil checks if the instance is nil
func (esc *epochStartConfigs) IsInterfaceNil() bool {
	return esc == nil
}
