package config

// SovereignEpochConfig will hold sovereign epoch configurations
type SovereignEpochConfig struct {
	SovereignEnableEpochs              SovereignEnableEpochs
	SovereignChainSpecificEnableEpochs SovereignChainSpecificEnableEpochs
}

// SovereignEnableEpochs will hold the configuration for sovereign activation epochs
type SovereignEnableEpochs struct{}

// SovereignChainSpecificEnableEpochs will hold the configuration for sovereign chain specific activation epochs
type SovereignChainSpecificEnableEpochs struct{}
