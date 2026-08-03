package core

import "github.com/multiversx/mx-chain-go/config"

// ValidateSupernovaActivationTuple -
func ValidateSupernovaActivationTuple(cfg config.Config, supernovaEpoch uint32, supernovaRound uint64) error {
	return validateSupernovaActivationTuple(cfg, supernovaEpoch, supernovaRound)
}
