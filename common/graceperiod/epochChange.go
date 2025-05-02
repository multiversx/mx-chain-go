package graceperiod

import (
	"errors"
	"sort"

	"github.com/multiversx/mx-chain-go/config"
)

var errEmptyGracePeriodByEpochConfig = errors.New("empty grace period by epoch config")
var errDuplicatedEpochConfig = errors.New("duplicated epoch config")
var errMissingEpochZeroConfig = errors.New("missing configuration for epoch 0")

// epochChangeGracePeriod holds the grace period configuration for epoch changes
type epochChangeGracePeriod struct {
	orderedConfigByEpoch []config.EpochChangeGracePeriodByEpoch
}

// NewEpochChangeGracePeriod creates a new instance of epochChangeGracePeriod
func NewEpochChangeGracePeriod(
	gracePeriodByEpoch []config.EpochChangeGracePeriodByEpoch,
) (*epochChangeGracePeriod, error) {
	if len(gracePeriodByEpoch) == 0 {
		return nil, errEmptyGracePeriodByEpochConfig
	}
	// check for duplicated configs
	seen := make(map[uint32]struct{})
	for _, cfg := range gracePeriodByEpoch {
		_, exists := seen[cfg.EnableEpoch]
		if exists {
			return nil, errDuplicatedEpochConfig
		}
		seen[cfg.EnableEpoch] = struct{}{}
	}

	// should have a config for epoch 0
	_, exists := seen[0]
	if !exists {
		return nil, errMissingEpochZeroConfig
	}

	ecgp := &epochChangeGracePeriod{
		orderedConfigByEpoch: make([]config.EpochChangeGracePeriodByEpoch, len(gracePeriodByEpoch)),
	}

	// sort the config values in ascending order
	copy(ecgp.orderedConfigByEpoch, gracePeriodByEpoch)
	sort.SliceStable(ecgp.orderedConfigByEpoch, func(i, j int) bool {
		return ecgp.orderedConfigByEpoch[i].EnableEpoch < ecgp.orderedConfigByEpoch[j].EnableEpoch
	})

	return ecgp, nil
}

// GetGracePeriodForEpoch returns the grace period for the given epoch
func (ecgp *epochChangeGracePeriod) GetGracePeriodForEpoch(epoch uint32) (uint32, error) {
	for i := len(ecgp.orderedConfigByEpoch) - 1; i >= 0; i-- {
		if ecgp.orderedConfigByEpoch[i].EnableEpoch <= epoch {
			return ecgp.orderedConfigByEpoch[i].GracePeriodInRounds, nil
		}
	}

	return 0, errEmptyGracePeriodByEpochConfig
}

// IsInterfaceNil checks if the instance is nil
func (ecgp *epochChangeGracePeriod) IsInterfaceNil() bool {
	return ecgp == nil
}
