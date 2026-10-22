package core

import (
	"time"

	"github.com/multiversx/mx-chain-go/config"
)

// ValidateSupernovaActivationTuple -
func ValidateSupernovaActivationTuple(
	cfg config.Config,
	economicsCfg config.EconomicsConfig,
	ratingsCfg config.RatingsConfig,
	supernovaEpoch uint32,
	supernovaRound uint64,
) error {
	return validateSupernovaActivationTuple(cfg, economicsCfg, ratingsCfg, supernovaEpoch, supernovaRound)
}

// SupernovaGenesisTime -
func (cc *coreComponents) SupernovaGenesisTime() time.Time {
	return cc.supernovaGenesisTime
}
