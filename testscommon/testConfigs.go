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

// GetDefaultHeaderVersionConfig -
func GetDefaultHeaderVersionConfig() config.VersionsConfig {
	return config.VersionsConfig{
		DefaultVersion: "default",
		VersionsByEpochs: []config.VersionByEpochs{
			{
				StartEpoch: 0,
				Version:    "*",
			},
			{
				StartEpoch: 1,
				Version:    "2",
			},
		},
		Cache: config.CacheConfig{
			Name:     "VersionsCache",
			Type:     "LRU",
			Capacity: 100,
		},
	}
}
