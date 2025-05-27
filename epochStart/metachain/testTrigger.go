package metachain

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
)

// TestTrigger extends start of epoch trigger and is used in integration tests as it exposes some functions
// that are not supposed to be used in production code
// Exported functions simplify the reproduction of edge cases
type TestTrigger struct {
	*trigger
}

// SetTrigger sets the start of epoch trigger
func (t *TestTrigger) SetTrigger(triggerHandler epochStart.TriggerHandler) {
	actualTrigger, ok := triggerHandler.(*trigger)
	if !ok {
		return
	}

	t.trigger = actualTrigger
}

// SetRoundsPerEpoch sets the number of round between epochs
func (t *TestTrigger) SetRoundsPerEpoch(roundsPerEpoch uint64) {
	minRoundsBetweenEpochs := t.getMinRoundsBetweenEpochs()
	if minRoundsBetweenEpochs > roundsPerEpoch {
		minRoundsBetweenEpochs = roundsPerEpoch - 1
	}

	t.chainParametersHandler = &chainParameters.ChainParametersHandlerStub{
		CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
			return config.ChainParametersByEpochConfig{
				RoundsPerEpoch:         int64(roundsPerEpoch),
				MinRoundsBetweenEpochs: int64(minRoundsBetweenEpochs),
			}
		},
	}
}

// SetMinRoundsBetweenEpochs sets the minimum number of round between epochs
func (t *TestTrigger) SetMinRoundsBetweenEpochs(minRoundsPerEpoch uint64) {
	roundsPerEpoch := t.getRoundsPerEpoch()
	if minRoundsPerEpoch > roundsPerEpoch {
		minRoundsPerEpoch = roundsPerEpoch - 1
	}

	t.chainParametersHandler = &chainParameters.ChainParametersHandlerStub{
		CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
			return config.ChainParametersByEpochConfig{
				RoundsPerEpoch:         int64(roundsPerEpoch),
				MinRoundsBetweenEpochs: int64(minRoundsPerEpoch),
			}
		},
	}
}

// GetRoundsPerEpoch gets the number of rounds per epoch
func (t *TestTrigger) GetRoundsPerEpoch() uint64 {
	return t.getRoundsPerEpoch()
}

// SetEpoch sets the current epoch for the testTrigger
func (t *TestTrigger) SetEpoch(epoch uint32) {
	t.epoch = epoch
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *TestTrigger) IsInterfaceNil() bool {
	return t == nil
}
