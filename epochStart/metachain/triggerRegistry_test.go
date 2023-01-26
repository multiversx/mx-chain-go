package metachain

import (
	"encoding/json"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
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
	rt.marshaller = t.marshaller
	rt.hasher = t.hasher
	rt.appStatusHandler = t.appStatusHandler
	rt.nextEpochStartRound = t.nextEpochStartRound
	rt.validatorInfoPool = t.validatorInfoPool

	return rt
}

func TestTrigger_LoadStateAfterSave(t *testing.T) {
	t.Parallel()

	epoch := uint32(5)
	arguments := createMockEpochStartTriggerArguments()
	arguments.Epoch = epoch
	bootStorer := genericMocks.NewStorerMock()

	arguments.Storage = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return bootStorer, nil
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
	require.Nil(t, err)
	require.NotEqual(t, epochStartTrigger1, epochStartTrigger2)

	err = epochStartTrigger2.LoadState(key)
	require.Nil(t, err)
	require.Equal(t, epochStartTrigger1, epochStartTrigger2)
}

func TestTrigger_LoadStateBackwardsCompatibility(t *testing.T) {
	t.Parallel()

	epoch := uint32(5)
	arguments := createMockEpochStartTriggerArguments()
	arguments.Epoch = epoch
	bootStorer := genericMocks.NewStorerMock()

	arguments.Storage = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return bootStorer, nil
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

	trig := createLegacyTriggerRegistryFromTrigger(epochStartTrigger1)
	d, err := json.Marshal(trig)
	require.Nil(t, err)
	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)

	err = bootStorer.Put(trigInternalKey, d)
	require.Nil(t, err)

	err = epochStartTrigger2.LoadState(key)
	require.Nil(t, err)
	require.Equal(t, epochStartTrigger1, epochStartTrigger2)
}

type legacyTriggerRegistry struct {
	Epoch                       uint32
	CurrentRound                uint64
	EpochFinalityAttestingRound uint64
	CurrEpochStartRound         uint64
	PrevEpochStartRound         uint64
	EpochStartMetaHash          []byte
	EpochStartMeta              data.HeaderHandler
}

func createLegacyTriggerRegistryFromTrigger(t *trigger) *legacyTriggerRegistry {
	metaBlock, _ := t.epochStartMeta.(*block.MetaBlock)
	return &legacyTriggerRegistry{
		Epoch:                       t.epoch,
		CurrentRound:                t.currentRound,
		EpochFinalityAttestingRound: t.epochFinalityAttestingRound,
		CurrEpochStartRound:         t.currEpochStartRound,
		PrevEpochStartRound:         t.prevEpochStartRound,
		EpochStartMetaHash:          t.epochStartMetaHash,
		EpochStartMeta:              metaBlock,
	}
}
