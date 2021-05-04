package ntp_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/ntp"
	beevikNtp "github.com/beevik/ntp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var responseMock1 *beevikNtp.Response
var failNtpMock1 = false
var responseMock2 *beevikNtp.Response
var failNtpMock2 = false
var responseMock3 *beevikNtp.Response
var failNtpMock3 = false

var errNtpMock = errors.New("NTP Mock generic error")
var queryMock4Call = 0
var mutex = sync.Mutex{}

func queryMock1(options ntp.NTPOptions, _ int) (*beevikNtp.Response, error) {
	fmt.Printf("Hosts: %s\n", options.Hosts)

	if failNtpMock1 {
		return nil, errNtpMock
	}

	return responseMock1, nil
}

func queryMock2(options ntp.NTPOptions, _ int) (*beevikNtp.Response, error) {
	fmt.Printf("Hosts: %s\n", options.Hosts)

	if failNtpMock2 {
		return nil, errNtpMock
	}

	return responseMock2, nil
}

func queryMock3(options ntp.NTPOptions, _ int) (*beevikNtp.Response, error) {
	fmt.Printf("Hosts: %s\n", options.Hosts)

	if failNtpMock3 {
		return nil, errNtpMock
	}

	return responseMock3, nil
}

func queryMock4(options ntp.NTPOptions, _ int) (*beevikNtp.Response, error) {
	fmt.Printf("Hosts: %s\n", options.Hosts)

	mutex.Lock()
	queryMock4Call++
	mutex.Unlock()

	return nil, errNtpMock
}

func queryMock5(_ ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
	switch hostIndex {
	case 0:
		return nil, errNtpMock
	default:
		return &beevikNtp.Response{ClockOffset: time.Second}, nil
	}
}

func queryMock6(_ ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
	switch hostIndex {
	case 0:
		return &beevikNtp.Response{ClockOffset: time.Second}, nil
	default:
		return nil, errNtpMock
	}
}

func TestHandleErrorInDoSync(t *testing.T) {
	failNtpMock1 = true
	st := ntp.NewSyncTime(config.NTPConfig{Hosts: []string{""}, SyncPeriodSeconds: 1}, queryMock1)

	st.Sync()

	assert.Equal(t, st.ClockOffset(), time.Millisecond*0)

	st.SetClockOffset(1234)

	st.Sync()

	assert.Equal(t, st.ClockOffset(), time.Duration(1234))

}

func TestValueInDoSync(t *testing.T) {
	responseMock2 = &beevikNtp.Response{ClockOffset: 23456}

	failNtpMock2 = false
	st := ntp.NewSyncTime(config.NTPConfig{Hosts: []string{""}, SyncPeriodSeconds: 1}, queryMock2)

	assert.Equal(t, st.ClockOffset(), time.Millisecond*0)
	st.Sync()
	assert.Equal(t, st.ClockOffset(), time.Nanosecond*23456)

	st.SetClockOffset(1234)

	st.Sync()

	assert.Equal(t, st.ClockOffset(), time.Nanosecond*23456)
}

func TestGetOffset(t *testing.T) {
	responseMock3 = &beevikNtp.Response{ClockOffset: 23456}

	failNtpMock3 = false
	st := ntp.NewSyncTime(config.NTPConfig{Hosts: []string{""}, SyncPeriodSeconds: 1}, queryMock3)

	assert.Equal(t, st.ClockOffset(), time.Millisecond*0)
	st.Sync()
	assert.Equal(t, st.ClockOffset(), time.Nanosecond*23456)
	assert.Equal(t, st.ClockOffset(), time.Nanosecond*23456)
}

func TestCallQuery(t *testing.T) {
	st := ntp.NewSyncTime(config.NTPConfig{Hosts: []string{""}, SyncPeriodSeconds: 1}, queryMock4)
	st.StartSyncingTime()

	assert.NotNil(t, st.Query())
	assert.Equal(t, time.Second, st.SyncPeriod())

	// wait a few cycles
	time.Sleep(time.Millisecond * 100)

	mutex.Lock()
	qmc := queryMock4Call
	mutex.Unlock()
	assert.NotEqual(t, qmc, 0)

	fmt.Printf("Current time: %v\n", st.FormattedCurrentTime())

	_ = st.Close()
}

func TestCallQueryShouldErrIndexOutOfBounds(t *testing.T) {
	t.Parallel()

	st := ntp.NewSyncTime(config.NTPConfig{SyncPeriodSeconds: 3600}, nil)
	query := st.Query()
	response, err := query(ntp.NTPOptions{Hosts: []string{"host1", "host2", "host3"}}, 3)

	assert.Nil(t, response)
	assert.Equal(t, ntp.ErrIndexOutOfBounds, err)
}

func TestCallQueryShouldWork(t *testing.T) {
	//TODO fix this test
	t.Skip("rework this test as to not rely on the internet connection")
	t.Parallel()

	ntpConfig := ntp.NewNTPGoogleConfig()
	ntpOptions := ntp.NewNTPOptions(ntpConfig)
	st := ntp.NewSyncTime(ntpConfig, nil)
	query := st.Query()
	response, err := query(ntpOptions, 0)

	assert.NotNil(t, response)
	assert.Nil(t, err)
}

func TestNtpHostIsChange(t *testing.T) {
	t.Parallel()

	ntpConfig := config.NTPConfig{Hosts: []string{"host1", "host2", "host3"}, SyncPeriodSeconds: 1}
	st := ntp.NewSyncTime(ntpConfig, queryMock5)
	st.Sync()

	//HostIndex will be equal with 1 and time offset will be a second
	assert.Equal(t, time.Second, st.ClockOffset())
}

func TestSyncShouldNotUpdateClockOffset(t *testing.T) {
	t.Parallel()

	ntpConfig := config.NTPConfig{Hosts: []string{"host1", "host2", "host3"}, SyncPeriodSeconds: 1}
	st := ntp.NewSyncTime(ntpConfig, queryMock6)
	st.SetClockOffset(time.Millisecond)
	st.Sync()

	assert.Equal(t, time.Millisecond, st.ClockOffset())
}

func TestGetClockOffsetsWithoutEdges(t *testing.T) {
	t.Parallel()

	st := ntp.NewSyncTime(config.NTPConfig{SyncPeriodSeconds: 1}, nil)

	clockOffsets := make([]time.Duration, 0)
	clockOffsetsWithoutEdges := st.GetClockOffsetsWithoutEdges(clockOffsets)
	require.Equal(t, 0, len(clockOffsetsWithoutEdges))

	clockOffsets = []time.Duration{100}
	clockOffsetsWithoutEdges = st.GetClockOffsetsWithoutEdges(clockOffsets)
	require.Equal(t, 1, len(clockOffsetsWithoutEdges))

	clockOffsets = []time.Duration{100, 54}
	clockOffsetsWithoutEdges = st.GetClockOffsetsWithoutEdges(clockOffsets)
	require.Equal(t, 2, len(clockOffsetsWithoutEdges))
	assert.Equal(t, time.Duration(54), clockOffsetsWithoutEdges[0])
	assert.Equal(t, time.Duration(100), clockOffsetsWithoutEdges[1])

	clockOffsets = []time.Duration{100, 54, 2}
	clockOffsetsWithoutEdges = st.GetClockOffsetsWithoutEdges(clockOffsets)
	require.Equal(t, 3, len(clockOffsetsWithoutEdges))
	assert.Equal(t, time.Duration(2), clockOffsetsWithoutEdges[0])
	assert.Equal(t, time.Duration(54), clockOffsetsWithoutEdges[1])
	assert.Equal(t, time.Duration(100), clockOffsetsWithoutEdges[2])

	clockOffsets = []time.Duration{100, 54, 2, 52}
	clockOffsetsWithoutEdges = st.GetClockOffsetsWithoutEdges(clockOffsets)
	require.Equal(t, 4, len(clockOffsetsWithoutEdges))
	assert.Equal(t, time.Duration(2), clockOffsetsWithoutEdges[0])
	assert.Equal(t, time.Duration(52), clockOffsetsWithoutEdges[1])
	assert.Equal(t, time.Duration(54), clockOffsetsWithoutEdges[2])
	assert.Equal(t, time.Duration(100), clockOffsetsWithoutEdges[3])

	clockOffsets = []time.Duration{100, 54, 12, 52, 16, 1, 70}
	clockOffsetsWithoutEdges = st.GetClockOffsetsWithoutEdges(clockOffsets)
	require.Equal(t, 5, len(clockOffsetsWithoutEdges))
	assert.Equal(t, time.Duration(12), clockOffsetsWithoutEdges[0])
	assert.Equal(t, time.Duration(16), clockOffsetsWithoutEdges[1])
	assert.Equal(t, time.Duration(52), clockOffsetsWithoutEdges[2])
	assert.Equal(t, time.Duration(54), clockOffsetsWithoutEdges[3])
	assert.Equal(t, time.Duration(70), clockOffsetsWithoutEdges[4])
}

func TestGetHarmonicMean(t *testing.T) {
	t.Parallel()

	st := ntp.NewSyncTime(config.NTPConfig{SyncPeriodSeconds: 1}, nil)

	clockOffsets := make([]time.Duration, 0)
	harmonicMean := st.GetHarmonicMean(clockOffsets)
	assert.Equal(t, time.Duration(0), harmonicMean)

	clockOffsets = []time.Duration{2, 0, 3}
	harmonicMean = st.GetHarmonicMean(clockOffsets)
	assert.Equal(t, time.Duration(0), harmonicMean)

	// harmonic mean for 4, 1, 4 is equal with: 3 / (1/4 + 1/1 + 1/4) = 3 / 1.5 = 2
	clockOffsets = []time.Duration{4, 1, 4}
	harmonicMean = st.GetHarmonicMean(clockOffsets)
	assert.Equal(t, time.Duration(2), harmonicMean)
}

func TestGetSleepTime(t *testing.T) {
	t.Parallel()

	syncPeriodSeconds := 3600
	givenTime := time.Duration(syncPeriodSeconds) * time.Second
	st := ntp.NewSyncTime(config.NTPConfig{SyncPeriodSeconds: syncPeriodSeconds}, nil)
	minSleepTime := time.Duration(float64(givenTime) - float64(givenTime)*0.2)
	maxSleepTime := time.Duration(float64(givenTime) + float64(givenTime)*0.2)

	fmt.Printf("given time = %d\nmin time = %d\nmax time = %d\n\n", givenTime, minSleepTime, maxSleepTime)

	for i := 0; i < 1000; i++ {
		sleepTime := st.GetSleepTime()
		fmt.Printf("%d\n", sleepTime)
		assert.True(t, sleepTime >= minSleepTime && sleepTime <= maxSleepTime)
	}
}

func TestCallQueryShouldNotUpdateOnOutOfBoundValuesPositive(t *testing.T) {
	t.Parallel()

	st := ntp.NewSyncTime(
		config.NTPConfig{
			SyncPeriodSeconds: 3600,
			Hosts:             []string{"host1"},
		},
		func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
			return &beevikNtp.Response{
				ClockOffset: ntp.OutOfBoundsDuration + time.Nanosecond,
			}, nil
		},
	)

	currentValue := time.Microsecond
	st.SetClockOffset(currentValue)
	st.Sync()

	assert.Equal(t, currentValue, st.ClockOffset())
}

func TestCallQueryShouldNotUpdateOnOutOfBoundValuesNegative(t *testing.T) {
	t.Parallel()

	st := ntp.NewSyncTime(
		config.NTPConfig{
			SyncPeriodSeconds: 3600,
			Hosts:             []string{"host1"},
		},
		func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
			return &beevikNtp.Response{
				ClockOffset: -ntp.OutOfBoundsDuration - 2*time.Nanosecond,
			}, nil
		},
	)

	currentValue := time.Microsecond
	st.SetClockOffset(currentValue)
	st.Sync()

	assert.Equal(t, currentValue, st.ClockOffset())
}
