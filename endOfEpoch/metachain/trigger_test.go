package metachain

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/stretchr/testify/assert"
)

func createMockEpochStartTriggerArguments() *ArgsNewMetaEpochStartTrigger {
	return &ArgsNewMetaEpochStartTrigger{
		Rounder:     &mock.RounderStub{},
		GenesisTime: time.Time{},
		Settings: &config.EpochStartConfig{
			MinRoundsBetweenEpochs: 1,
			RoundsPerEpoch:         2,
		},
		Epoch: 0,
	}
}

func TestNewEpochStartTrigger_NilArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	neoet, err := NewEpochStartTrigger(nil)

	assert.Nil(t, neoet)
	assert.Equal(t, epochStart.ErrNilArgsNewMetaEpochStartTrigger, err)
}

func TestNewEpochStartTrigger_NilRounderShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Rounder = nil

	neoet, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, neoet)
	assert.Equal(t, epochStart.ErrNilRounder, err)
}

func TestNewEpochStartTrigger_NilSettingsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings = nil

	neoet, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, neoet)
	assert.Equal(t, epochStart.ErrNilEpochStartSettings, err)
}

func TestNewEpochStartTrigger_InvalidSettingsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings.RoundsPerEpoch = 0

	neoet, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, neoet)
	assert.Equal(t, epochStart.ErrInvalidSettingsForEpochStartTrigger, err)
}

func TestNewEpochStartTrigger_InvalidSettingsShouldErr2(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings.RoundsPerEpoch = 1
	arguments.Settings.MinRoundsBetweenEpochs = 0

	neoet, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, neoet)
	assert.Equal(t, epochStart.ErrInvalidSettingsForEpochStartTrigger, err)
}

func TestNewEpochStartTrigger_InvalidSettingsShouldErr3(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings.RoundsPerEpoch = 4
	arguments.Settings.MinRoundsBetweenEpochs = 6

	neoet, err := NewEpochStartTrigger(arguments)
	assert.Nil(t, neoet)
	assert.Equal(t, epochStart.ErrInvalidSettingsForEpochStartTrigger, err)
}

func TestNewEpochStartTrigger_ShouldOk(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()

	neoet, err := NewEpochStartTrigger(arguments)
	assert.NotNil(t, neoet)
	assert.Nil(t, err)
}

func TestTrigger_Update(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	round := int64(0)
	arguments := createMockEpochStartTriggerArguments()
	arguments.Epoch = epoch
	neoet, _ := NewEpochStartTrigger(arguments)

	neoet.Update(round)
	round++
	neoet.Update(round)
	round++
	neoet.Update(round)
	round++
	neoet.Update(round)

	ret := neoet.IsEpochStart()
	assert.True(t, ret)

	epc := neoet.Epoch()
	assert.Equal(t, epoch+1, epc)

	neoet.Processed()
	ret = neoet.IsEpochStart()
	assert.False(t, ret)
}

func TestTrigger_ForceEpochStartIncorrectRoundShouldErr(t *testing.T) {
	t.Parallel()

	round := int64(1)
	arguments := createMockEpochStartTriggerArguments()
	neoet, _ := NewEpochStartTrigger(arguments)

	neoet.Update(round)

	err := neoet.ForceEpochStart(0)
	assert.Equal(t, epochStart.ErrSavedRoundIsHigherThanInputRound, err)
}

func TestTrigger_ForceEpochStartRoundEqualWithSavedRoundShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	neoet, _ := NewEpochStartTrigger(arguments)

	err := neoet.ForceEpochStart(0)
	assert.Equal(t, epochStart.ErrForceEpochStartCanBeCalledOnlyOnNewRound, err)
}

func TestTrigger_ForceEpochStartNotEnoughRoundsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEpochStartTriggerArguments()
	arguments.Settings.MinRoundsBetweenEpochs = 2
	neoet, _ := NewEpochStartTrigger(arguments)

	err := neoet.ForceEpochStart(1)
	assert.Equal(t, epochStart.ErrNotEnoughRoundsBetweenEpochs, err)
}

func TestTrigger_ForceEpochStartShouldOk(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	arguments := createMockEpochStartTriggerArguments()
	arguments.Epoch = epoch
	neoet, _ := NewEpochStartTrigger(arguments)

	err := neoet.ForceEpochStart(1)
	assert.Nil(t, err)

	newEpoch := neoet.Epoch()
	assert.Equal(t, epoch+1, newEpoch)

	isEpochStart := neoet.IsEpochStart()
	assert.True(t, isEpochStart)
}
