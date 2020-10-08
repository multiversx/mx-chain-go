package metachain

import "github.com/ElrondNetwork/elrond-go/epochStart"

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
	t.roundsPerEpoch = roundsPerEpoch
	if t.minRoundsBetweenEpochs > t.roundsPerEpoch {
		t.minRoundsBetweenEpochs = t.roundsPerEpoch - 1
	}
}

// SetMinRoundsBetweenEpochs sets the minimum number of round between epochs
func (t *TestTrigger) SetMinRoundsBetweenEpochs(minRoundsPerEpoch uint64) {
	t.minRoundsBetweenEpochs = minRoundsPerEpoch
	if t.minRoundsBetweenEpochs > t.roundsPerEpoch {
		t.minRoundsBetweenEpochs = t.roundsPerEpoch - 1
	}
}

// GetRoundsPerEpoch gets the number of rounds per epoch
func (t *TestTrigger) GetRoundsPerEpoch() uint64 {
	return t.roundsPerEpoch
}

// SetEpoch sets the current epoch for the testTrigger
func (t *TestTrigger) SetEpoch(epoch uint32) {
	t.epoch = epoch
}
