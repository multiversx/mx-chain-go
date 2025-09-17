package configs

import (
	"errors"
	"sort"

	"github.com/multiversx/mx-chain-go/config"
)

const (
	defaultMaxMetaNoncesBehind               = 15
	defaultMaxMetaNoncesBehindForGlobalStuck = 30
	defaultMaxShardNoncesBehind              = 15
	defaultMaxRoundsWithoutNewBlockReceived  = 10
	defaultMaxRoundsWithoutCommittedBlock    = 10
)

// ErrEmptyProcessConfigsByEpoch signals that an empty process configs by epoch has been provided
var ErrEmptyProcessConfigsByEpoch = errors.New("empty process configs by epoch")

// ErrEmptyProcessConfigsByEpoch signals that an empty process configs by round has been provided
var ErrEmptyProcessConfigsByRound = errors.New("empty process configs by round")

// ErrDuplicatedEpochConfig signals that a duplicated config section has been provided
var ErrDuplicatedEpochConfig = errors.New("duplicated epoch config")

// ErrDuplicatedRoundConfig signals that a duplicated config section has been provided
var ErrDuplicatedRoundConfig = errors.New("duplicated round config")

// ErrMissingEpochZeroConfig signals that epoch zero configuration is missing
var ErrMissingEpochZeroConfig = errors.New("missing configuration for epoch 0")

// ErrMissingRoundZeroConfig signals that base round configuration is missing
var ErrMissingRoundZeroConfig = errors.New("missing base configuration for round")

// processConfigsByEpoch holds the process configuration for epoch changes
type processConfigsByEpoch struct {
	orderedConfigByEpoch []config.ProcessConfigByEpoch
	orderedConfigByRound []config.ProcessConfigByRound
}

// NewProcessConfigsHandler creates a new process configs by epoch component
func NewProcessConfigsHandler(
	configsByEpoch []config.ProcessConfigByEpoch,
	configsByRound []config.ProcessConfigByRound,
) (*processConfigsByEpoch, error) {
	err := checkConfigsByEpoch(configsByEpoch)
	if err != nil {
		return nil, err
	}

	err = checkConfigsByRound(configsByRound)
	if err != nil {
		return nil, err
	}

	pce := &processConfigsByEpoch{
		orderedConfigByEpoch: make([]config.ProcessConfigByEpoch, len(configsByEpoch)),
		orderedConfigByRound: make([]config.ProcessConfigByRound, len(configsByRound)),
	}

	// sort the config values in ascending order
	copy(pce.orderedConfigByEpoch, configsByEpoch)
	sort.SliceStable(pce.orderedConfigByEpoch, func(i, j int) bool {
		return pce.orderedConfigByEpoch[i].EnableEpoch < pce.orderedConfigByEpoch[j].EnableEpoch
	})

	copy(pce.orderedConfigByRound, configsByRound)
	sort.SliceStable(pce.orderedConfigByRound, func(i, j int) bool {
		return pce.orderedConfigByRound[i].EnableRound < pce.orderedConfigByRound[j].EnableRound
	})

	return pce, nil
}

// TODO: consider extracting common functionality into a base component
func checkConfigsByEpoch(configsByEpoch []config.ProcessConfigByEpoch) error {
	if len(configsByEpoch) == 0 {
		return ErrEmptyProcessConfigsByEpoch
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

func checkConfigsByRound(configsByRound []config.ProcessConfigByRound) error {
	if len(configsByRound) == 0 {
		return ErrEmptyProcessConfigsByRound
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

// GetMaxMetaNoncesBehind returns the max meta nonces behind by epoch
func (pce *processConfigsByEpoch) GetMaxMetaNoncesBehindByEpoch(epoch uint32) uint32 {
	for i := len(pce.orderedConfigByEpoch) - 1; i >= 0; i-- {
		if pce.orderedConfigByEpoch[i].EnableEpoch <= epoch {
			return pce.orderedConfigByEpoch[i].MaxMetaNoncesBehind
		}
	}

	return defaultMaxMetaNoncesBehind // this should not happen
}

// GetMaxMetaNoncesBehindForBlobalStuck returns the max meta nonces behind for global stuck by epoch
func (pce *processConfigsByEpoch) GetMaxMetaNoncesBehindForBlobalStuckByEpoch(epoch uint32) uint32 {
	for i := len(pce.orderedConfigByEpoch) - 1; i >= 0; i-- {
		if pce.orderedConfigByEpoch[i].EnableEpoch <= epoch {
			return pce.orderedConfigByEpoch[i].MaxMetaNoncesBehindForGlobalStuck
		}
	}

	return defaultMaxMetaNoncesBehindForGlobalStuck // this should not happen
}

// GetMaxMetaNoncesBehindForBlobalStuck returns the max meta nonces behind for global stuck by epoch
func (pce *processConfigsByEpoch) GetMaxShardNoncesBehindByEpoch(epoch uint32) uint32 {
	for i := len(pce.orderedConfigByEpoch) - 1; i >= 0; i-- {
		if pce.orderedConfigByEpoch[i].EnableEpoch <= epoch {
			return pce.orderedConfigByEpoch[i].MaxShardNoncesBehind
		}
	}

	return defaultMaxShardNoncesBehind // this should not happen
}

// GetMaxRoundsWithoutNewBlockReceived returns max rounds without new block received by epoch
func (pce *processConfigsByEpoch) GetMaxRoundsWithoutNewBlockReceivedByRound(round uint64) uint32 {
	for i := len(pce.orderedConfigByRound) - 1; i >= 0; i-- {
		if pce.orderedConfigByRound[i].EnableRound <= round {
			return pce.orderedConfigByRound[i].MaxRoundsWithoutNewBlockReceived
		}
	}

	return defaultMaxRoundsWithoutNewBlockReceived // this should not happen
}

// GetMaxRoundsWithoutCommittedBlock returns max rounds without commited block
func (pce *processConfigsByEpoch) GetMaxRoundsWithoutCommittedBlock(round uint64) uint32 {
	for i := len(pce.orderedConfigByRound) - 1; i >= 0; i-- {
		if pce.orderedConfigByRound[i].EnableRound <= round {
			return pce.orderedConfigByRound[i].MaxRoundsWithoutCommittedBlock
		}
	}

	return defaultMaxRoundsWithoutCommittedBlock // this should not happen
}

// IsInterfaceNil checks if the instance is nil
func (pce *processConfigsByEpoch) IsInterfaceNil() bool {
	return pce == nil
}
