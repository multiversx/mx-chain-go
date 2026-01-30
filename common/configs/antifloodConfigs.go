package configs

import (
	"errors"
	"fmt"
	"sort"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

const minBanDuration = 1 // second
const minMessages = 1
const minTotalSize = 1 //1Byte
const initNumMessages = 1
const maxPercentReserved = 90.0
const minPercentReserved = 0.0

// ErrEmptyAntifloodConfigsByRound signals that an empty antiflood configs has been provided
var ErrEmptyAntifloodConfigsByRound = errors.New("empty antiflood configs")

type antifloodConfigs struct {
	orderedConfigsByRound []config.AntifloodConfigByRound
	isEnabled             bool
	roundNotifier         process.RoundNotifier
}

// NewAntifloodConfigsHandler creates a new instance of antiflood configs handler
func NewAntifloodConfigsHandler(
	antifoodConfig config.AntifloodConfig,
	roundNotifier process.RoundNotifier,
) (*antifloodConfigs, error) {
	if check.IfNil(roundNotifier) {
		return nil, process.ErrNilRoundNotifier
	}
	err := checkAnifloodConfigsByRound(antifoodConfig.ConfigsByRound)
	if err != nil {
		return nil, err
	}

	ac := &antifloodConfigs{
		orderedConfigsByRound: make([]config.AntifloodConfigByRound, len(antifoodConfig.ConfigsByRound)),
		roundNotifier:         roundNotifier,
		isEnabled:             antifoodConfig.Enabled,
	}

	copy(ac.orderedConfigsByRound, antifoodConfig.ConfigsByRound)
	sort.SliceStable(ac.orderedConfigsByRound, func(i, j int) bool {
		return ac.orderedConfigsByRound[i].Round < ac.orderedConfigsByRound[j].Round
	})

	return ac, nil
}

func checkAnifloodConfigsByRound(configs []config.AntifloodConfigByRound) error {
	if len(configs) == 0 {
		return ErrEmptyAntifloodConfigsByRound
	}

	// check for duplicated configs
	seen := make(map[uint64]struct{})
	for _, cfg := range configs {
		_, exists := seen[cfg.Round]
		if exists {
			return ErrDuplicatedRoundConfig
		}
		seen[cfg.Round] = struct{}{}

		err := checkAntifloodConfig(cfg)
		if err != nil {
			return err
		}
	}

	_, exists := seen[0]
	if !exists {
		return ErrMissingRoundZeroConfig
	}

	return nil
}

func checkAntifloodConfig(conf config.AntifloodConfigByRound) error {
	err := checkFloodPreventerConfig(conf.FastReacting)
	if err != nil {
		return err
	}

	err = checkFloodPreventerConfig(conf.SlowReacting)
	if err != nil {
		return err
	}

	err = checkFloodPreventerConfig(conf.OutOfSpecs)
	if err != nil {
		return err
	}

	return nil
}

func checkFloodPreventerConfig(conf config.FloodPreventerConfig) error {
	if conf.BlackList.ThresholdNumMessagesPerInterval == 0 {
		return fmt.Errorf("%w, thresholdNumReceivedFlood == 0", process.ErrInvalidValue)
	}
	if conf.BlackList.ThresholdSizePerInterval == 0 {
		return fmt.Errorf("%w, thresholdSizeReceivedFlood == 0", process.ErrInvalidValue)
	}
	if conf.BlackList.PeerBanDurationInSeconds < minBanDuration {
		return fmt.Errorf("%w for peerBanDurationInSeconds", process.ErrInvalidValue)
	}
	if conf.BlackList.NumFloodingRounds == 0 {
		return fmt.Errorf("%w, numFloodingRounds", process.ErrInvalidValue)
	}

	if conf.PeerMaxInput.BaseMessagesPerInterval < minMessages {
		return fmt.Errorf("%w, maxMessagesPerPeer: provided %d, minimum %d",
			process.ErrInvalidValue,
			conf.PeerMaxInput.BaseMessagesPerInterval,
			minMessages,
		)
	}
	if conf.PeerMaxInput.TotalSizePerInterval < minTotalSize {
		return fmt.Errorf("%w, maxTotalSizePerPeer: provided %d, minimum %d",
			process.ErrInvalidValue,
			conf.PeerMaxInput.TotalSizePerInterval,
			minTotalSize,
		)
	}
	if conf.ReservedPercent > maxPercentReserved {
		return fmt.Errorf("%w, percentReserved: provided %0.3f, maximum %0.3f",
			process.ErrInvalidValue,
			conf.ReservedPercent,
			maxPercentReserved,
		)
	}
	if conf.ReservedPercent < minPercentReserved {
		return fmt.Errorf("%w, percentReserved: provided %0.3f, minimum %0.3f",
			process.ErrInvalidValue,
			conf.ReservedPercent,
			minPercentReserved,
		)
	}
	if conf.PeerMaxInput.IncreaseFactor.Factor < 0 {
		return fmt.Errorf("%w, increaseFactor is negative: provided %0.3f",
			process.ErrInvalidValue,
			conf.PeerMaxInput.IncreaseFactor.Factor,
		)
	}

	return nil
}

// IsEnabled returns true if antiflood is enabled
func (ac *antifloodConfigs) IsEnabled() bool {
	return ac.isEnabled
}

// GetCurrentConfig returns antiflood config based on the current round
func (ac *antifloodConfigs) GetCurrentConfig() config.AntifloodConfigByRound {
	currentRound := ac.roundNotifier.CurrentRound()

	for i := len(ac.orderedConfigsByRound) - 1; i >= 0; i-- {
		if ac.orderedConfigsByRound[i].Round <= currentRound {
			return ac.orderedConfigsByRound[i]
		}
	}

	// this should not happen, there is already a check in the constructor for missing config
	return config.AntifloodConfigByRound{}
}

// IsInterfaceNil checks if the instance is nil
func (ac *antifloodConfigs) IsInterfaceNil() bool {
	return ac == nil
}
