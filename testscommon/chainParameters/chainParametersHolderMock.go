package chainParameters

import (
	"github.com/multiversx/mx-chain-go/config"
)

var testChainParams = config.ChainParametersByEpochConfig{
	RoundDuration:               6000,
	Hysteresis:                  0,
	EnableEpoch:                 0,
	ShardConsensusGroupSize:     1,
	ShardMinNumNodes:            1,
	MetachainConsensusGroupSize: 1,
	MetachainMinNumNodes:        1,
	Adaptivity:                  false,
}

// ChainParametersHolderMock -
type ChainParametersHolderMock struct {
}

// CurrentChainParameters -
func (c *ChainParametersHolderMock) CurrentChainParameters() config.ChainParametersByEpochConfig {
	return testChainParams
}

// AllChainParameters -
func (c *ChainParametersHolderMock) AllChainParameters() []config.ChainParametersByEpochConfig {
	return []config.ChainParametersByEpochConfig{
		testChainParams,
	}
}

// ChainParametersForEpoch -
func (c *ChainParametersHolderMock) ChainParametersForEpoch(_ uint32) (config.ChainParametersByEpochConfig, error) {
	return testChainParams, nil
}

// IsInterfaceNil -
func (c *ChainParametersHolderMock) IsInterfaceNil() bool {
	return c == nil
}
