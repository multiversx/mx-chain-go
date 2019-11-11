package metachain

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch/mock"
	"github.com/stretchr/testify/assert"
)

func createMockEndOfEpochTriggerArguments() *ArgsNewMetaEndOfEpochTrigger {
	return &ArgsNewMetaEndOfEpochTrigger{
		Rounder:     &mock.RounderMock{},
		GenesisTime: time.Time{},
		Settings: &config.EndOfEpochConfig{
			MinRoundsBetweenEpochs: 1,
			RoundsPerEpoch:         2,
		},
		Epoch: 0,
	}
}

func TestNewEndOfEpochTrigger_NilArgumentsShouldErr(t *testing.T) {
	t.Parallel()

	neoet, err := NewEndOfEpochTrigger(nil)

	assert.Nil(t, neoet)
	assert.Equal(t, endOfEpoch.ErrNilArgsNewMetaEndOfEpochTrigger, err)
}

func TestNewEndOfEpochTrigger_NilRounderShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEndOfEpochTriggerArguments()
	arguments.Rounder = nil

	neoet, err := NewEndOfEpochTrigger(arguments)
	assert.Nil(t, neoet)
	assert.Equal(t, endOfEpoch.ErrNilRounder, err)
}

func TestNewEndOfEpochTrigger_NilSettingsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEndOfEpochTriggerArguments()
	arguments.Settings = nil

	neoet, err := NewEndOfEpochTrigger(arguments)
	assert.Nil(t, neoet)
	assert.Equal(t, endOfEpoch.ErrNilEndOfEpochSettings, err)
}

func TestNewEndOfEpochTrigger_InvalidSettingsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEndOfEpochTriggerArguments()
	arguments.Settings.RoundsPerEpoch = 0

	neoet, err := NewEndOfEpochTrigger(arguments)
	assert.Nil(t, neoet)
	assert.Equal(t, endOfEpoch.ErrInvalidSettingsForEndOfEpochTrigger, err)
}

func TestNewEndOfEpochTrigger_InvalidSettingsShouldErr2(t *testing.T) {
	t.Parallel()

	arguments := createMockEndOfEpochTriggerArguments()
	arguments.Settings.RoundsPerEpoch = 1
	arguments.Settings.MinRoundsBetweenEpochs = 0

	neoet, err := NewEndOfEpochTrigger(arguments)
	assert.Nil(t, neoet)
	assert.Equal(t, endOfEpoch.ErrInvalidSettingsForEndOfEpochTrigger, err)
}

func TestNewEndOfEpochTrigger_InvalidSettingsShouldErr3(t *testing.T) {
	t.Parallel()

	arguments := createMockEndOfEpochTriggerArguments()
	arguments.Settings.RoundsPerEpoch = 4
	arguments.Settings.MinRoundsBetweenEpochs = 6

	neoet, err := NewEndOfEpochTrigger(arguments)
	assert.Nil(t, neoet)
	assert.Equal(t, endOfEpoch.ErrInvalidSettingsForEndOfEpochTrigger, err)
}

func TestNewEndOfEpochTrigger_ShouldOk(t *testing.T) {
	t.Parallel()

	arguments := createMockEndOfEpochTriggerArguments()

	neoet, err := NewEndOfEpochTrigger(arguments)
	assert.NotNil(t, neoet)
	assert.Nil(t, err)
}

func TestTrigger_Update(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	round := int64(0)
	arguments := createMockEndOfEpochTriggerArguments()
	arguments.Epoch = epoch
	neoet, _ := NewEndOfEpochTrigger(arguments)

	neoet.Update(round)
	round++
	neoet.Update(round)
	round++
	neoet.Update(round)
	round++
	neoet.Update(round)

	ret := neoet.IsEndOfEpoch()
	assert.True(t, ret)

	epc := neoet.Epoch()
	assert.Equal(t, epoch+1, epc)

	neoet.Processed()
	ret = neoet.IsEndOfEpoch()
	assert.False(t, ret)
}

func TestTrigger_ForceEndOfEpochIncorrectRoundShouldErr(t *testing.T) {
	t.Parallel()

	round := int64(1)
	arguments := createMockEndOfEpochTriggerArguments()
	neoet, _ := NewEndOfEpochTrigger(arguments)

	neoet.Update(round)

	err := neoet.ForceEndOfEpoch(0)
	assert.Equal(t, endOfEpoch.ErrSavedRoundIsHigherThanInputRound, err)
}

func TestTrigger_ForceEndOfEpochRoundEqualWithSavedRoundShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEndOfEpochTriggerArguments()
	neoet, _ := NewEndOfEpochTrigger(arguments)

	err := neoet.ForceEndOfEpoch(0)
	assert.Equal(t, endOfEpoch.ErrForceEndOfEpochCannotBeCalledOnNewRound, err)
}

func TestTrigger_ForceEndOfEpochNotEnoughRoundsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEndOfEpochTriggerArguments()
	arguments.Settings.MinRoundsBetweenEpochs = 2
	neoet, _ := NewEndOfEpochTrigger(arguments)

	err := neoet.ForceEndOfEpoch(1)
	assert.Equal(t, endOfEpoch.ErrNotEnoughRoundsBetweenEpochs, err)
}

func TestTrigger_ForceEndOfEpochShouldOk(t *testing.T) {
	t.Parallel()

	epoch := uint32(0)
	arguments := createMockEndOfEpochTriggerArguments()
	arguments.Epoch = epoch
	neoet, _ := NewEndOfEpochTrigger(arguments)

	err := neoet.ForceEndOfEpoch(1)
	assert.Nil(t, err)

	newEpoch := neoet.Epoch()
	assert.Equal(t, epoch+1, newEpoch)

	isEndOfEpoch := neoet.IsEndOfEpoch()
	assert.True(t, isEndOfEpoch)
}
