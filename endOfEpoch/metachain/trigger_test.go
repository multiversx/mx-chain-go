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
		SyncTimer:   &mock.SyncTimerMock{},
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

func TestNewEndOfEpochTrigger_NilSyncTimerShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createMockEndOfEpochTriggerArguments()
	arguments.SyncTimer = nil

	neoet, err := NewEndOfEpochTrigger(arguments)
	assert.Nil(t, neoet)
	assert.Equal(t, endOfEpoch.ErrNilSyncTimer, err)
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

func TestEndOfEpochTrigger_IsEndOfEpochCurrentRoundZeroShouldRetTrue(t *testing.T) {
	t.Parallel()

	arguments := createMockEndOfEpochTriggerArguments()
	arguments.Rounder = &mock.RounderMock{
		UpdateRoundCalled: func(t2 time.Time, t time.Time) {
		},
		IndexCalled: func() int64 {
			return 0
		},
	}
	arguments.SyncTimer = &mock.SyncTimerMock{
		CurrentTimeCalled: func() time.Time {
			return time.Now()
		},
	}
	neoet, _ := NewEndOfEpochTrigger(arguments)

	ret := neoet.IsEndOfEpoch()
	assert.True(t, ret)
}

func TestTrigger_IsEndOfEpochRoundIndexGreaterShouldTrue(t *testing.T) {
	t.Parallel()

	arguments := createMockEndOfEpochTriggerArguments()
	arguments.Rounder = &mock.RounderMock{
		UpdateRoundCalled: func(t2 time.Time, t time.Time) {
		},
		IndexCalled: func() int64 {
			return 10
		},
	}
	arguments.SyncTimer = &mock.SyncTimerMock{
		CurrentTimeCalled: func() time.Time {
			return time.Now()
		},
	}

	neoet, _ := NewEndOfEpochTrigger(arguments)

	ret := neoet.IsEndOfEpoch()
	assert.True(t, ret)
}

func TestTrigger_IsEndOfEpochRoundIndexLessShouldFalse(t *testing.T) {
	t.Parallel()

	arguments := createMockEndOfEpochTriggerArguments()
	arguments.Rounder = &mock.RounderMock{
		UpdateRoundCalled: func(t2 time.Time, t time.Time) {
		},
		IndexCalled: func() int64 {
			return 1
		},
	}
	arguments.SyncTimer = &mock.SyncTimerMock{
		CurrentTimeCalled: func() time.Time {
			return time.Now()
		},
	}

	neoet, _ := NewEndOfEpochTrigger(arguments)

	ret := neoet.IsEndOfEpoch()
	assert.False(t, ret)
}
