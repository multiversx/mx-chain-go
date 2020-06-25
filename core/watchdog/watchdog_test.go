package watchdog_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/core/watchdog"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/stretchr/testify/assert"
)

func TestNewWatchdog_NilAlarmSchedulerShouldErr(t *testing.T) {
	t.Parallel()

	w, err := watchdog.NewWatchdog(nil, make(chan endProcess.ArgEndProcess, 1))
	assert.True(t, check.IfNil(w))
	assert.Equal(t, watchdog.ErrNilAlarmScheduler, err)
}

func TestNewWatchdog_NilChanShouldErr(t *testing.T) {
	t.Parallel()

	w, err := watchdog.NewWatchdog(&mock.AlarmSchedulerStub{}, nil)
	assert.True(t, check.IfNil(w))
	assert.Equal(t, watchdog.ErrNilEndProcessChan, err)
}

func TestWatchdog_Set(t *testing.T) {
	t.Parallel()

	addCalled := false
	alarmScheduler := &mock.AlarmSchedulerStub{
		AddCalled: func(f func(alarmID string), duration time.Duration, s string) {
			addCalled = true
		},
	}
	w, _ := watchdog.NewWatchdog(alarmScheduler, make(chan endProcess.ArgEndProcess, 1))

	w.Set(func(alarmID string) {}, time.Second, "alarm")

	assert.True(t, addCalled)
}

func TestWatchdog_SetDefault(t *testing.T) {
	t.Parallel()

	alarm := "testComponent"
	channel := make(chan endProcess.ArgEndProcess, 1)
	addCalled := false
	alarmScheduler := &mock.AlarmSchedulerStub{
		AddCalled: func(f func(alarmID string), duration time.Duration, s string) {
			f(alarm)
			addCalled = true
		},
	}
	w, _ := watchdog.NewWatchdog(alarmScheduler, channel)

	w.SetDefault(time.Second, alarm)

	arg := <-channel
	expectedReason := "alarm " + alarm + " has expired"
	expectedDescription := "the " + alarm + " is stuck"

	assert.Equal(t, expectedReason, arg.Reason)
	assert.Equal(t, expectedDescription, arg.Description)
	assert.True(t, addCalled)
}

func TestWatchdog_Stop(t *testing.T) {
	t.Parallel()

	stopCalled := false
	alarmScheduler := &mock.AlarmSchedulerStub{
		CancelCalled: func(s string) {
			stopCalled = true
		},
	}
	w, _ := watchdog.NewWatchdog(alarmScheduler, make(chan endProcess.ArgEndProcess, 1))

	w.Stop("alarm")

	assert.True(t, stopCalled)
}

func TestWatchdog_Reset(t *testing.T) {
	t.Parallel()

	resetCalled := false
	alarmScheduler := &mock.AlarmSchedulerStub{
		ResetCalled: func(s string) {
			resetCalled = true
		},
	}
	w, _ := watchdog.NewWatchdog(alarmScheduler, make(chan endProcess.ArgEndProcess, 1))

	w.Reset("alarm")

	assert.True(t, resetCalled)
}
