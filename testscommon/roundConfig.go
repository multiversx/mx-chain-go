package testscommon

import "github.com/multiversx/mx-chain-go/config"

// GetDefaultRoundsConfig -
func GetDefaultRoundsConfig() config.RoundConfig {
	return config.RoundConfig{
		RoundActivations: map[string]config.ActivationRoundByName{
			"DisableAsyncCallV1": {
				Round: "18446744073709551615",
			},
		},
	}
}
