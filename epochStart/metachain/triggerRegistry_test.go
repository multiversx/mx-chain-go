package metachain

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

func cloneTrigger(t *trigger) *trigger {
	rt := &trigger{}

	rt.isEpochStart = t.isEpochStart
	rt.epoch = t.epoch
	rt.currentRound = t.currentRound
	rt.epochFinalityAttestingRound = t.epochFinalityAttestingRound
	rt.currEpochStartRound = t.currEpochStartRound
	rt.prevEpochStartRound = t.prevEpochStartRound
	rt.roundsPerEpoch = t.roundsPerEpoch
	rt.minRoundsBetweenEpochs = t.minRoundsBetweenEpochs
	rt.epochStartMetaHash = t.epochStartMetaHash
	rt.triggerStateKey = t.triggerStateKey
	rt.epochStartTime = t.epochStartTime
	rt.epochStartNotifier = t.epochStartNotifier
	rt.triggerStorage = t.triggerStorage
	rt.metaHeaderStorage = t.metaHeaderStorage
	rt.marshalizer = t.marshalizer
	rt.hasher = t.hasher
	rt.appStatusHandler = t.appStatusHandler
	rt.nextEpochStartRound = t.nextEpochStartRound

	return rt
}

func TestTrigger_LoadStateAfterSave(t *testing.T) {
	t.Parallel()

	epoch := uint32(5)
	arguments := createMockEpochStartTriggerArguments()
	arguments.Epoch = epoch
	bootStorer := mock.NewStorerMock()

	arguments.Storage = &mock.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return bootStorer
		},
	}
	epochStartTrigger1, _ := NewEpochStartTrigger(arguments)
	// create a copy
	epochStartTrigger2 := cloneTrigger(epochStartTrigger1)

	key := []byte("key")
	epochStartTrigger1.triggerStateKey = key
	epochStartTrigger1.epoch = 6
	epochStartTrigger1.currentRound = 1000
	epochStartTrigger1.epochFinalityAttestingRound = 998
	epochStartTrigger1.currEpochStartRound = 800
	epochStartTrigger1.prevEpochStartRound = 650
	err := epochStartTrigger1.saveState(key)
	assert.Nil(t, err)
	assert.NotEqual(t, epochStartTrigger1, epochStartTrigger2)

	err = epochStartTrigger2.LoadState(key)
	assert.Nil(t, err)
	assert.Equal(t, epochStartTrigger1, epochStartTrigger2)
}
