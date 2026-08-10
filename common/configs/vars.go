package configs

import (
	"github.com/multiversx/mx-chain-go/common/configs/dto"
	"github.com/multiversx/mx-chain-go/config"
)

func initCfgVarMap() map[dto.ConfigVariable]configVariableHandler {
	return map[dto.ConfigVariable]configVariableHandler{
		dto.MaxConsecutiveRoundsOfRatingDecrease: {
			valueSelector: func(cfg config.ProcessConfigByRound) uint64 {
				return cfg.MaxConsecutiveRoundsOfRatingDecrease
			},
			defaultValue: defaultMaxConsecutiveRoundsOfRatingDecrease,
		},
		dto.MaxRoundsOfInactivityAccepted: {
			valueSelector: func(cfg config.ProcessConfigByRound) uint64 {
				return cfg.MaxRoundsOfInactivityAccepted
			},
			defaultValue: defaultMaxRoundsOfInactivityAccepted,
		},
	}
}
