package configs

import (
	"github.com/multiversx/mx-chain-go/common/configs/dto"
	"github.com/multiversx/mx-chain-go/config"
)

func initCfgVarMap() map[dto.ConfigVariable]configVariableHandler {
	return map[dto.ConfigVariable]configVariableHandler{
		dto.NumFloodingRoundsFastReacting: {
			valueSelector: func(cfg config.ProcessConfigByRound) uint64 {
				return uint64(cfg.NumFloodingRoundsFastReacting)
			},
			defaultValue: defaultNumFloodingRoundsFastReacting,
		},
		dto.NumFloodingRoundsSlowReacting: {
			valueSelector: func(cfg config.ProcessConfigByRound) uint64 {
				return uint64(cfg.NumFloodingRoundsSlowReacting)
			},
			defaultValue: defaultNumFloodingRoundsSlowReacting,
		},
		dto.NumFloodingRoundsOutOfSpecs: {
			valueSelector: func(cfg config.ProcessConfigByRound) uint64 {
				return uint64(cfg.NumFloodingRoundsOutOfSpecs)
			},
			defaultValue: defaultNumFloodingRoundsOutOfSpecs,
		},
		dto.MaxConsecutiveRoundsOfRatingDecrease: {
			valueSelector: func(cfg config.ProcessConfigByRound) uint64 {
				return cfg.MaxConsecutiveRoundsOfRatingDecrease
			},
			defaultValue: defaultMaxConsecutiveRoundsOfRatingDecrease,
		},
	}
}
