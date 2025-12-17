package configs

import (
	"errors"
	"fmt"
	"sort"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

const minRoundsToKeepUnprocessedData = uint64(1)

const (
	defaultMaxMetaNoncesBehind                    = 15
	defaultMaxMetaNoncesBehindForGlobalStuck      = 30
	defaultMaxShardNoncesBehind                   = 15
	defaultMaxRoundsWithoutNewBlockReceived       = 10
	defaultMaxRoundsWithoutCommittedBlock         = 10
	defaultRoundModulusTriggerWhenSyncIsStuck     = 20
	defaultMaxSyncWithErrorsAllowed               = 20
	defaultMaxRoundsToKeepUnprocessedMiniBlocks   = 3000
	defaultMaxRoundsToKeepUnprocessedTransactions = 3000
)

// ErrEmptyProcessConfigsByEpoch signals that an empty process configs by epoch has been provided
var ErrEmptyProcessConfigsByEpoch = errors.New("empty process configs by epoch")

// ErrEmptyProcessConfigsByRound signals that an empty process configs by round has been provided
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
		err := checkRoundConfigValues(cfg)
		if err != nil {
			return err
		}

		seen[cfg.EnableRound] = struct{}{}
	}

	_, exists := seen[0]
	if !exists {
		return ErrMissingRoundZeroConfig
	}

	return nil
}

func checkRoundConfigValues(cfg config.ProcessConfigByRound) error {
	if cfg.MaxRoundsToKeepUnprocessedTransactions < minRoundsToKeepUnprocessedData {
		return fmt.Errorf("%w for MaxRoundsToKeepUnprocessedTransactions, received %d, min expected %d",
			process.ErrInvalidValue, cfg.MaxRoundsToKeepUnprocessedTransactions, minRoundsToKeepUnprocessedData)
	}
	if cfg.MaxRoundsToKeepUnprocessedMiniBlocks < minRoundsToKeepUnprocessedData {
		return fmt.Errorf("%w for MaxRoundsToKeepUnprocessedMiniBlocks, received %d, min expected %d",
			process.ErrInvalidValue, cfg.MaxRoundsToKeepUnprocessedMiniBlocks, minRoundsToKeepUnprocessedData)
	}

	return nil
}

// GetMaxMetaNoncesBehindByEpoch returns the max meta nonces behind by epoch
func (pce *processConfigsByEpoch) GetMaxMetaNoncesBehindByEpoch(epoch uint32) uint32 {
	for i := len(pce.orderedConfigByEpoch) - 1; i >= 0; i-- {
		if pce.orderedConfigByEpoch[i].EnableEpoch <= epoch {
			return pce.orderedConfigByEpoch[i].MaxMetaNoncesBehind
		}
	}

	return defaultMaxMetaNoncesBehind // this should not happen
}

// GetMaxMetaNoncesBehindForGlobalStuckByEpoch returns the max meta nonces behind for global stuck by epoch
func (pce *processConfigsByEpoch) GetMaxMetaNoncesBehindForGlobalStuckByEpoch(epoch uint32) uint32 {
	for i := len(pce.orderedConfigByEpoch) - 1; i >= 0; i-- {
		if pce.orderedConfigByEpoch[i].EnableEpoch <= epoch {
			return pce.orderedConfigByEpoch[i].MaxMetaNoncesBehindForGlobalStuck
		}
	}

	return defaultMaxMetaNoncesBehindForGlobalStuck // this should not happen
}

// GetMaxShardNoncesBehindByEpoch returns the max meta nonces behind for global stuck by epoch
func (pce *processConfigsByEpoch) GetMaxShardNoncesBehindByEpoch(epoch uint32) uint32 {
	for i := len(pce.orderedConfigByEpoch) - 1; i >= 0; i-- {
		if pce.orderedConfigByEpoch[i].EnableEpoch <= epoch {
			return pce.orderedConfigByEpoch[i].MaxShardNoncesBehind
		}
	}

	return defaultMaxShardNoncesBehind // this should not happen
}

func getConfigValueByRound[T any](
	configs []config.ProcessConfigByRound,
	round uint64,
	selector func(config.ProcessConfigByRound) T,
	defaultValue T,
) T {
	for i := len(configs) - 1; i >= 0; i-- {
		if configs[i].EnableRound <= round {
			return selector(configs[i])
		}
	}

	return defaultValue
}

// GetMaxRoundsWithoutNewBlockReceivedByRound returns max rounds without new block received by epoch
func (pce *processConfigsByEpoch) GetMaxRoundsWithoutNewBlockReceivedByRound(round uint64) uint32 {
	return getConfigValueByRound(
		pce.orderedConfigByRound,
		round,
		func(cfg config.ProcessConfigByRound) uint32 {
			return cfg.MaxRoundsWithoutNewBlockReceived
		},
		defaultMaxRoundsWithoutNewBlockReceived,
	)
}

// GetMaxRoundsWithoutCommittedBlock returns max rounds without commited block
func (pce *processConfigsByEpoch) GetMaxRoundsWithoutCommittedBlock(round uint64) uint32 {
	return getConfigValueByRound(
		pce.orderedConfigByRound,
		round,
		func(cfg config.ProcessConfigByRound) uint32 {
			return cfg.MaxRoundsWithoutCommittedBlock
		},
		defaultMaxRoundsWithoutCommittedBlock,
	)
}

// GetRoundModulusTriggerWhenSyncIsStuck returns round modulus when sync is stuck
func (pce *processConfigsByEpoch) GetRoundModulusTriggerWhenSyncIsStuck(round uint64) uint32 {
	return getConfigValueByRound(
		pce.orderedConfigByRound,
		round,
		func(cfg config.ProcessConfigByRound) uint32 {
			return cfg.RoundModulusTriggerWhenSyncIsStuck
		},
		defaultRoundModulusTriggerWhenSyncIsStuck,
	)
}

// GetMaxSyncWithErrorsAllowed returns max allowed sync errors until an error event is triggered
func (pce *processConfigsByEpoch) GetMaxSyncWithErrorsAllowed(round uint64) uint32 {
	return getConfigValueByRound(
		pce.orderedConfigByRound,
		round,
		func(cfg config.ProcessConfigByRound) uint32 {
			return cfg.MaxSyncWithErrorsAllowed
		},
		defaultMaxSyncWithErrorsAllowed,
	)
}

// GetMaxRoundsToKeepUnprocessedMiniBlocks returns max rounds to keep unprocessed mini blocks based on round
func (pce *processConfigsByEpoch) GetMaxRoundsToKeepUnprocessedMiniBlocks(round uint64) uint64 {
	return getConfigValueByRound(
		pce.orderedConfigByRound,
		round,
		func(cfg config.ProcessConfigByRound) uint64 {
			return cfg.MaxRoundsToKeepUnprocessedMiniBlocks
		},
		defaultMaxRoundsToKeepUnprocessedMiniBlocks,
	)
}

// GetMaxRoundsToKeepUnprocessedTransactions returns max rounds to keep unprocessed transaction blocks based on round
func (pce *processConfigsByEpoch) GetMaxRoundsToKeepUnprocessedTransactions(round uint64) uint64 {
	return getConfigValueByRound(
		pce.orderedConfigByRound,
		round,
		func(cfg config.ProcessConfigByRound) uint64 {
			return cfg.MaxRoundsToKeepUnprocessedTransactions
		},
		defaultMaxRoundsToKeepUnprocessedTransactions,
	)
}

// IsInterfaceNil checks if the instance is nil
func (pce *processConfigsByEpoch) IsInterfaceNil() bool {
	return pce == nil
}
