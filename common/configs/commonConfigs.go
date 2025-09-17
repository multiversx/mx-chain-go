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

type commonConfigs struct {
	orderedEpochStartConfigByEpoch []config.EpochStartConfigByEpoch
}

// NewCommonConfigsHandler creates a new process configs by epoch component
func NewCommonConfigsHandler(
	configsByEpoch []config.EpochStartConfigByEpoch,
) (*commonConfigs, error) {
	err := checkCommonConfigsByEpoch(configsByEpoch)
	if err != nil {
		return nil, err
	}

	esc := &commonConfigs{
		orderedEpochStartConfigByEpoch: make([]config.EpochStartConfigByEpoch, len(configsByEpoch)),
	}

	// sort the config values in ascending order
	copy(esc.orderedEpochStartConfigByEpoch, configsByEpoch)
	sort.SliceStable(esc.orderedEpochStartConfigByEpoch, func(i, j int) bool {
		return esc.orderedEpochStartConfigByEpoch[i].EnableEpoch < esc.orderedEpochStartConfigByEpoch[j].EnableEpoch
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

// IsInterfaceNil checks if the instance is nil
func (cc *commonConfigs) IsInterfaceNil() bool {
	return cc == nil
}
