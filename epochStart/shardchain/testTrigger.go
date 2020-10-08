package shardchain

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

// SetRoundsPerEpoch does nothing as trigger in shards is not done by chronology
func (t *TestTrigger) SetRoundsPerEpoch(_ uint64) {
}

// SetMinRoundsBetweenEpochs does nothing as trigger in shards is not done by chronology
func (t *TestTrigger) SetMinRoundsBetweenEpochs(_ uint64) {
}

// GetRoundsPerEpoch gets the number of rounds per epoch
func (t *TestTrigger) GetRoundsPerEpoch() uint64 {
	return 0
}

// SetEpoch sets the current epoch for the testTrigger
func (t *TestTrigger) SetEpoch(epoch uint32) {
	t.epoch = epoch
	t.metaEpoch = epoch
}
