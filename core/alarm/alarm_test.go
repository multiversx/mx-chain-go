package alarm_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/alarm"
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/stretchr/testify/require"
)

const testAlarm = "alarm1"

func TestNewAlarmScheduler(t *testing.T) {
	t.Parallel()

	alarmScheduler := alarm.NewAlarmScheduler()
	require.NotNil(t, alarmScheduler)
}

func TestAlarmScheduler_Add(t *testing.T) {
	t.Parallel()

	var calledString atomic.String
	var nbCalls atomic.Counter

	cb := func(alarmID string) {
		calledString.Set(alarmID)
		nbCalls.Increment()
	}

	alarmID := testAlarm
	alarmExpiry := time.Millisecond * 50
	alarmScheduler := alarm.NewAlarmScheduler()

	alarmScheduler.Add(cb, alarmExpiry, alarmID)
	time.Sleep(alarmExpiry * 3)

	require.Equal(t, alarmID, calledString.Get())
	require.Equal(t, int64(1), nbCalls.Get())
}

func TestAlarmScheduler_AddMultipleAlarmsExpireCorrectOrder(t *testing.T) {
	t.Parallel()

	var calledString atomic.String
	var nbCalls atomic.Counter

	cb := func(alarmID string) {
		str := calledString.Get()
		calledString.Set(str + alarmID)
		nbCalls.Increment()
	}

	alarmScheduler := alarm.NewAlarmScheduler()

	nbAlarms := 10
	deltaAlarmExpiry := time.Millisecond * 50
	alarmIDs := make([]string, nbAlarms)
	expectedStr := ""

	for i := 0; i < nbAlarms; i++ {
		alarmIDs[i] = "alarm" + strconv.Itoa(i)
		alarmScheduler.Add(cb, deltaAlarmExpiry*time.Duration(i+1), alarmIDs[i])
		expectedStr += alarmIDs[i]
	}

	time.Sleep(deltaAlarmExpiry * time.Duration(nbAlarms+2))

	require.Equal(t, expectedStr, calledString.Get())
	require.Equal(t, int64(nbAlarms), nbCalls.Get())
}

func TestAlarmScheduler_CancelBeforeExpiry(t *testing.T) {
	t.Parallel()

	var calledString atomic.String
	var nbCalls atomic.Counter

	cb := func(alarmID string) {
		calledString.Set(alarmID)
		nbCalls.Increment()
	}

	alarmID := testAlarm
	alarmExpiry := time.Millisecond * 50
	alarmScheduler := alarm.NewAlarmScheduler()

	alarmScheduler.Add(cb, alarmExpiry, alarmID)
	time.Sleep(alarmExpiry / 2)
	alarmScheduler.Cancel(alarmID)
	time.Sleep(alarmExpiry * 3)

	require.Equal(t, "", calledString.Get())
	require.Equal(t, int64(0), nbCalls.Get())
}

func TestAlarmScheduler_CancelTwice(t *testing.T) {
	t.Parallel()

	var calledString atomic.String
	var nbCalls atomic.Counter

	cb := func(alarmID string) {
		calledString.Set(alarmID)
		nbCalls.Increment()
	}

	alarmID := testAlarm
	alarmExpiry := time.Millisecond * 50
	alarmScheduler := alarm.NewAlarmScheduler()

	alarmScheduler.Add(cb, alarmExpiry, alarmID)
	time.Sleep(alarmExpiry / 2)
	alarmScheduler.Cancel(alarmID)
	alarmScheduler.Cancel(alarmID)
	time.Sleep(alarmExpiry * 3)

	require.Equal(t, "", calledString.Get())
	require.Equal(t, int64(0), nbCalls.Get())
}

func TestAlarmScheduler_CancelAfterCanceledFinished(t *testing.T) {
	t.Parallel()

	var calledString atomic.String
	var nbCalls atomic.Counter

	cb := func(alarmID string) {
		calledString.Set(alarmID)
		nbCalls.Increment()
	}

	alarmID := testAlarm
	alarmExpiry := time.Millisecond * 50
	alarmScheduler := alarm.NewAlarmScheduler()

	alarmScheduler.Add(cb, alarmExpiry, alarmID)
	time.Sleep(alarmExpiry / 3)
	alarmScheduler.Cancel(alarmID)
	time.Sleep(alarmExpiry / 3)
	alarmScheduler.Cancel(alarmID)
	time.Sleep(alarmExpiry * 3)

	require.Equal(t, "", calledString.Get())
	require.Equal(t, int64(0), nbCalls.Get())
}

func TestAlarmScheduler_CancelAfterAddImmediately(t *testing.T) {
	t.Parallel()

	var calledString atomic.String
	var nbCalls atomic.Counter

	cb := func(alarmID string) {
		calledString.Set(alarmID)
		nbCalls.Increment()
	}

	alarmID := testAlarm
	alarmExpiry := time.Millisecond * 50
	alarmScheduler := alarm.NewAlarmScheduler()

	alarmScheduler.Add(cb, alarmExpiry, alarmID)
	alarmScheduler.Cancel(alarmID)
	time.Sleep(alarmExpiry * 3)

	require.Equal(t, "", calledString.Get())
	require.Equal(t, int64(0), nbCalls.Get())
}

func TestAlarmScheduler_CancelAlarmFromMultipleScheduled(t *testing.T) {
	t.Parallel()

	var calledString atomic.String
	var nbCalls atomic.Counter

	cb := func(alarmID string) {
		str := calledString.Get()
		calledString.Set(str + alarmID)
		nbCalls.Increment()
	}

	alarmScheduler := alarm.NewAlarmScheduler()

	nbAlarms := 10
	canceledAlarmIndex := 5
	deltaAlarmExpiry := time.Millisecond * 50
	alarmIDs := make([]string, nbAlarms)
	expectedStr := ""

	for i := 0; i < nbAlarms; i++ {
		alarmIDs[i] = "alarm" + strconv.Itoa(i)
		alarmScheduler.Add(cb, deltaAlarmExpiry*time.Duration(i+1), alarmIDs[i])
		if i != canceledAlarmIndex {
			expectedStr += alarmIDs[i]
		}
	}

	time.Sleep(deltaAlarmExpiry)
	alarmScheduler.Cancel(alarmIDs[canceledAlarmIndex])
	time.Sleep(deltaAlarmExpiry * time.Duration(nbAlarms+1))

	require.Equal(t, expectedStr, calledString.Get())
	require.Equal(t, int64(nbAlarms-1), nbCalls.Get())
}

func TestAlarmScheduler_Close(t *testing.T) {
	t.Parallel()

	var calledString atomic.String
	var nbCalls atomic.Counter

	cb := func(alarmID string) {
		str := calledString.Get()
		calledString.Set(str + alarmID)
		nbCalls.Increment()
	}

	alarmScheduler := alarm.NewAlarmScheduler()

	nbAlarms := 10
	deltaAlarmExpiry := time.Millisecond * 50
	alarmIDs := make([]string, nbAlarms)
	expectedStr := ""

	for i := 0; i < nbAlarms; i++ {
		alarmIDs[i] = "alarm" + strconv.Itoa(i)
		alarmScheduler.Add(cb, deltaAlarmExpiry*time.Duration(i+1), alarmIDs[i])
		expectedStr += alarmIDs[i]
	}

	time.Sleep(deltaAlarmExpiry / 2)
	alarmScheduler.Close()
	time.Sleep(deltaAlarmExpiry * time.Duration(nbAlarms))

	require.Equal(t, "", calledString.Get())
	require.Equal(t, int64(0), nbCalls.Get())
}

func TestAlarmScheduler_Reset(t *testing.T) {
	t.Parallel()

	var calledString atomic.String
	var nbCalls atomic.Counter

	cb := func(alarmID string) {
		calledString.Set(alarmID)
		nbCalls.Increment()
	}

	numResets := 5
	alarmID := testAlarm
	alarmExpiry := time.Millisecond * 50
	alarmScheduler := alarm.NewAlarmScheduler()

	alarmScheduler.Add(cb, alarmExpiry, alarmID)
	for i := 0; i < numResets; i++ {
		time.Sleep(alarmExpiry / 2)
		alarmScheduler.Reset(alarmID)
	}
	alarmScheduler.Cancel(alarmID)

	require.Equal(t, "", calledString.Get())
	require.Equal(t, int64(0), nbCalls.Get())
}
