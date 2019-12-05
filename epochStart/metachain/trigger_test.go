package metachain

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/stretchr/testify/assert"
)

func createMockEpochStartTriggerArguments() *ArgsNewMetaEpochStartTrigger {
	return &ArgsNewMetaEpochStartTrigger{
		GenesisTime: time.Time{},
		Settings: &config.EpochStartConfig{
			MinRoundsBetweenEpochs: 1,
			RoundsPerEpoch:         2,
		},
		Epoch:              0,
		EpochStartNotifier: &mock.EpochStartNotifierStub{},
	}
}

func TestNewEpochStartTrigger_NilArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	epochStartTrigger, err := NewEpochStartTrigger(nil)

	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilArgsNewMetaEpochStartTrigger, err)
}

func TestNewEpochStartTrigger_NilSettingsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings = nil

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilEpochStartSettings, err)
}

func TestNewEpochStartTrigger_InvalidSettingsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings.RoundsPerEpoch = 0

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrInvalidSettingsForEpochStartTrigger, err)
}

func TestNewEpochStartTrigger_NilEpochStartNotifierShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.EpochStartNotifier = nil

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrNilEpochStartNotifier, err)
}

func TestNewEpochStartTrigger_InvalidSettingsShouldErr2(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings.RoundsPerEpoch = 1
	arguments.Settings.MinRoundsBetweenEpochs = 0

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrInvalidSettingsForEpochStartTrigger, err)
}

func TestNewEpochStartTrigger_InvalidSettingsShouldErr3(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings.RoundsPerEpoch = 4
	arguments.Settings.MinRoundsBetweenEpochs = 6

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, epochStartTrigger)
	assert.Equal(t, epochStart.ErrInvalidSettingsForEpochStartTrigger, err)
}

func TestNewEpochStartTrigger_ShouldOk(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()

	epochStartTrigger, err := NewEpochStartTrigger(arguments)
	assert.NotNil(t, epochStartTrigger)
	assert.Nil(t, err)
}

func TestTrigger_Update(t *testing.T) {
	t.Parallel()

	notifierWasCalled := false
	epoch := uint32(0)
	round := uint64(0)
	arguments := createMockEpochStartTriggerArguments()
	arguments.Epoch = epoch
	arguments.EpochStartNotifier = &mock.EpochStartNotifierStub{
		NotifyAllCalled: func(hdr data.HeaderHandler) {
			notifierWasCalled = true
		},
	}
	epochStartTrigger, _ := NewEpochStartTrigger(arguments)

	epochStartTrigger.Update(round)
	round++
	epochStartTrigger.Update(round)
	round++
	epochStartTrigger.Update(round)
	round++
	epochStartTrigger.Update(round)

	ret := epochStartTrigger.IsEpochStart()
	assert.True(t, ret)

	epc := epochStartTrigger.Epoch()
	assert.Equal(t, epoch+1, epc)

	epochStartTrigger.SetProcessed(&block.MetaBlock{
		Round:      round,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{RootHash: []byte("root")}}}})
	ret = epochStartTrigger.IsEpochStart()
	assert.False(t, ret)
	assert.True(t, notifierWasCalled)
}

func TestTrigger_ForceEpochStartIncorrectRoundShouldErr(t *testing.T) {
	t.Parallel()

	round := uint64(1)
	arguments := createMockEpochStartTriggerArguments()
	epochStartTrigger, _ := NewEpochStartTrigger(arguments)

	epochStartTrigger.Update(round)

	err := epochStartTrigger.ForceEpochStart(0)
	assert.Equal(t, epochStart.ErrSavedRoundIsHigherThanInputRound, err)
}

func TestTrigger_ForceEpochStartRoundEqualWithSavedRoundShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	epochStartTrigger, _ := NewEpochStartTrigger(arguments)

	err := epochStartTrigger.ForceEpochStart(0)
	assert.Equal(t, epochStart.ErrForceEpochStartCanBeCalledOnlyOnNewRound, err)
}

func TestTrigger_ForceEpochStartNotEnoughRoundsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings.MinRoundsBetweenEpochs = 2
	epochStartTrigger, _ := NewEpochStartTrigger(arguments)

	err := epochStartTrigger.ForceEpochStart(1)
	assert.Equal(t, epochStart.ErrNotEnoughRoundsBetweenEpochs, err)
}

func TestTrigger_ForceEpochStartShouldOk(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	arguments := createMockEpochStartTriggerArguments()
	arguments.Epoch = epoch
	epochStartTrigger, _ := NewEpochStartTrigger(arguments)

	err := epochStartTrigger.ForceEpochStart(1)
	assert.Nil(t, err)

	newEpoch := epochStartTrigger.Epoch()
	assert.Equal(t, epoch+1, newEpoch)

	isEpochStart := epochStartTrigger.IsEpochStart()
	assert.True(t, isEpochStart)
}
