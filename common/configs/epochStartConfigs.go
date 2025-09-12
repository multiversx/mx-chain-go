package configs

import (
	"errors"
	"sort"

	"github.com/multiversx/mx-chain-go/config"
)

const (
	defaultGracePeriodRounds = 25
)

// ErrEmptyEpochStartConfigsByEpoch signals that an empty process configs by epoch has been provided
var ErrEmptyEpochStartConfigsByEpoch = errors.New("empty process configs by epoch")

// ErrEmptyEpochStartConfigsByEpoch signals that an empty process configs by round has been provided
var ErrEmptyEpochStartConfigsByRound = errors.New("empty process configs by round")

type epochStartConfigs struct {
	orderedConfigByEpoch []config.EpochStartConfigByEpoch
}

// NewEpochStartConfigsHandler creates a new process configs by epoch component
func NewEpochStartConfigsHandler(
	configsByEpoch []config.EpochStartConfigByEpoch,
) (*epochStartConfigs, error) {
	err := checkEpochStartConfigsByEpoch(configsByEpoch)
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

func checkEpochStartConfigsByEpoch(configsByEpoch []config.EpochStartConfigByEpoch) error {
	if len(configsByEpoch) == 0 {
		return ErrEmptyEpochStartConfigsByEpoch
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
func (esc *epochStartConfigs) GetGracePeriodRounds(epoch uint32) uint32 {
	for i := len(esc.orderedConfigByEpoch) - 1; i >= 0; i-- {
		if esc.orderedConfigByEpoch[i].EnableEpoch <= epoch {
			return esc.orderedConfigByEpoch[i].GracePeriodRounds
		}
	}

	return defaultGracePeriodRounds // this should not happen
}

// IsInterfaceNil checks if the instance is nil
func (esc *epochStartConfigs) IsInterfaceNil() bool {
	return esc == nil
}
