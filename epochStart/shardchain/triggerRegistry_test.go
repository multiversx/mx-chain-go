package shardchain

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/assert"
)

func TestTrigger_LoadStateAfterSave(t *testing.T) {
	t.Parallel()

	epoch := uint32(5)
	arguments := createMockShardEpochStartTriggerArguments()
	arguments.Epoch = epoch
	bootStorer := mock.NewStorerMock()

	arguments.Storage = &mock.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return bootStorer
		},
	}
	epochStartTrigger1, _ := NewEpochStartTrigger(arguments)
	// create a copy
	epochStartTrigger2 := *epochStartTrigger1

	key := []byte("key")
	epochStartTrigger1.triggerStateKey = key
	epochStartTrigger1.epoch = 10
	epochStartTrigger1.currentRoundIndex = 800
	epochStartTrigger1.epochStartRound = 650
	epochStartTrigger1.epochMetaBlockHash = []byte("meta block hash")
	epochStartTrigger1.isEpochStart = false
	epochStartTrigger1.epochFinalityAttestingRound = 680
	err := epochStartTrigger1.saveState(key)
	assert.Nil(t, err)
	assert.NotEqual(t, epochStartTrigger1, &epochStartTrigger2)

	err = epochStartTrigger2.LoadState(key)
	assert.Nil(t, err)
	assert.Equal(t, epochStartTrigger1, &epochStartTrigger2)
}
