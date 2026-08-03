package core

import (
	"time"

	"github.com/multiversx/mx-chain-go/config"
)

// ValidateSupernovaActivationTuple -
func ValidateSupernovaActivationTuple(cfg config.Config, supernovaEpoch uint32, supernovaRound uint64) error {
	return validateSupernovaActivationTuple(cfg, supernovaEpoch, supernovaRound)
}

// SupernovaGenesisTime -
func (cc *coreComponents) SupernovaGenesisTime() time.Time {
	return cc.supernovaGenesisTime
}
