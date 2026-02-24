package ntp_test

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	beevikNtp "github.com/beevik/ntp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/ntp"
	"github.com/multiversx/mx-chain-go/testscommon"
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
	st := ntp.NewSyncTime(config.NTPConfig{Hosts: []string{""}, SyncPeriodSeconds: 1, OutOfBoundsThreshold: 200}, queryMock2)

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
	st := ntp.NewSyncTime(config.NTPConfig{Hosts: []string{""}, SyncPeriodSeconds: 1, OutOfBoundsThreshold: 200}, queryMock3)

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
	// TODO fix this test
	t.Skip("rework this test as to not rely on the internet connection")
	t.Parallel()

	ntpConfig := testscommon.NewNTPGoogleConfig()
	ntpOptions := ntp.NewNTPOptions(ntpConfig)
	st := ntp.NewSyncTime(ntpConfig, nil)
	query := st.Query()
	response, err := query(ntpOptions, 0)

	assert.NotNil(t, response)
	assert.Nil(t, err)
}

func TestNtpHostIsChange(t *testing.T) {
	t.Parallel()

	ntpConfig := config.NTPConfig{Hosts: []string{"host1", "host2", "host3"}, SyncPeriodSeconds: 1, OutOfBoundsThreshold: 1200}
	st := ntp.NewSyncTime(ntpConfig, queryMock5)
	st.Sync()

	// HostIndex will be equal with 1 and time offset will be a second
	assert.Equal(t, time.Second, st.ClockOffset())
}

func TestSyncShouldUpdateClockOffsetWhenEnoughResponses(t *testing.T) {
	t.Parallel()

	// queryMock6: host 0 succeeds (ClockOffset = 1s), hosts 1 and 2 fail
	// 10 successful responses out of 30 total (33%) exceeds minResponsesPercent (25%)
	ntpConfig := config.NTPConfig{Hosts: []string{"host1", "host2", "host3"}, SyncPeriodSeconds: 1, OutOfBoundsThreshold: 1200}
	st := ntp.NewSyncTime(ntpConfig, queryMock6)
	st.SetClockOffset(time.Millisecond)
	st.Sync()

	assert.Equal(t, time.Second, st.ClockOffset())
}

func TestGetSleepTime(t *testing.T) {
	t.Parallel()

	syncPeriodSeconds := 3600
	givenTime := time.Duration(syncPeriodSeconds) * time.Second
	st := ntp.NewSyncTime(config.NTPConfig{SyncPeriodSeconds: syncPeriodSeconds, OutOfBoundsThreshold: 200}, nil)
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
			SyncPeriodSeconds:    3600,
			Hosts:                []string{"host1"},
			OutOfBoundsThreshold: 1,
		},
		func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
			time.Sleep(2 * time.Millisecond)

			return &beevikNtp.Response{
				ClockOffset: 1 + time.Millisecond,
			}, nil
		},
	)

	currentValue := 10 * time.Millisecond
	st.SetClockOffset(currentValue)
	st.Sync()

	expValue := 1 + time.Millisecond

	assert.Equal(t, expValue, st.ClockOffset())
}

func TestCallQueryShouldUpdateOnOutOfBoundValuesPositiveIfDurationNotOutOfBounds(t *testing.T) {
	t.Parallel()

	st := ntp.NewSyncTime(
		config.NTPConfig{
			SyncPeriodSeconds:    3600,
			Hosts:                []string{"host1"},
			OutOfBoundsThreshold: 1,
		},
		func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
			return &beevikNtp.Response{
				ClockOffset: 1 + time.Millisecond,
			}, nil
		},
	)

	currentValue := 10 * time.Millisecond
	st.SetClockOffset(currentValue)
	st.Sync()

	assert.NotEqual(t, currentValue, st.ClockOffset())
}

func TestCallQueryShouldUpdateOnOutOfBoundValuesNegative(t *testing.T) {
	t.Parallel()

	st := ntp.NewSyncTime(
		config.NTPConfig{
			SyncPeriodSeconds:    3600,
			Hosts:                []string{"host1"},
			OutOfBoundsThreshold: 2,
		},
		func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
			return &beevikNtp.Response{
				ClockOffset: -2 - 2*time.Millisecond,
			}, nil
		},
	)

	currentValue := 2 * 10 * time.Millisecond
	st.SetClockOffset(currentValue)
	st.Sync()

	expValue := -2 - 2*time.Millisecond

	assert.Equal(t, expValue, st.ClockOffset())
}

func TestCall_Sync_AcceptedBoundsChecks(t *testing.T) {
	t.Parallel()

	t.Run("response time within accepted bounds, should set new offset", func(t *testing.T) {
		t.Parallel()

		st := ntp.NewSyncTime(
			config.NTPConfig{
				SyncPeriodSeconds:    3600,
				Hosts:                []string{"host1"},
				OutOfBoundsThreshold: 2,
			},
			func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
				return &beevikNtp.Response{
					ClockOffset: 1 * time.Millisecond,
				}, nil
			},
		)

		currentValue := 3 * time.Millisecond
		st.SetClockOffset(currentValue)
		st.Sync()

		expClockOffset := 1 * time.Millisecond
		assert.Equal(t, expClockOffset, st.ClockOffset())
	})

	t.Run("response time within accepted bounds, clock offset not within accepted bounds, should set new offset", func(t *testing.T) {
		t.Parallel()

		st := ntp.NewSyncTime(
			config.NTPConfig{
				SyncPeriodSeconds:    3600,
				Hosts:                []string{"host1"},
				OutOfBoundsThreshold: 2,
			},
			func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
				return &beevikNtp.Response{
					ClockOffset: 4 * time.Millisecond,
				}, nil
			},
		)

		currentValue := 3 * time.Millisecond
		st.SetClockOffset(currentValue)
		st.Sync()

		expClockOffset := 4 * time.Millisecond
		assert.Equal(t, expClockOffset, st.ClockOffset())
	})

	t.Run("response time not within accepted bounds, clock offset not within accepted bounds, should set new offset", func(t *testing.T) {
		t.Parallel()

		st := ntp.NewSyncTime(
			config.NTPConfig{
				SyncPeriodSeconds:    3600,
				Hosts:                []string{"host1"},
				OutOfBoundsThreshold: 2,
			},
			func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
				time.Sleep(5 * time.Millisecond)

				return &beevikNtp.Response{
					ClockOffset: 4 * time.Millisecond,
				}, nil
			},
		)

		currentValue := 3 * time.Millisecond
		st.SetClockOffset(currentValue)
		st.Sync()

		expClockOffset := 4 * time.Millisecond
		assert.Equal(t, expClockOffset, st.ClockOffset())
	})

	t.Run("response time not within accepted bounds, clock offset within accepted bounds, should set new offset", func(t *testing.T) {
		t.Parallel()

		st := ntp.NewSyncTime(
			config.NTPConfig{
				SyncPeriodSeconds:    3600,
				Hosts:                []string{"host1"},
				OutOfBoundsThreshold: 2,
			},
			func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
				time.Sleep(5 * time.Millisecond)

				return &beevikNtp.Response{
					ClockOffset: 1 * time.Millisecond,
				}, nil
			},
		)

		currentValue := 3 * time.Millisecond
		st.SetClockOffset(currentValue)
		st.Sync()

		expClockOffset := 1 * time.Millisecond
		assert.Equal(t, expClockOffset, st.ClockOffset())
	})

	t.Run("no successful response times, should set new offset", func(t *testing.T) {
		t.Parallel()

		st := ntp.NewSyncTime(
			config.NTPConfig{
				SyncPeriodSeconds:    3600,
				Hosts:                []string{"host1"},
				OutOfBoundsThreshold: 2,
			},
			func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
				return nil, errors.New("err")
			},
		)

		currentValue := 3 * time.Millisecond
		st.SetClockOffset(currentValue)
		st.Sync()

		expClockOffset := 3 * time.Millisecond
		assert.Equal(t, expClockOffset, st.ClockOffset())
	})
}

func TestSyncTime_ForceSync(t *testing.T) {
	t.Parallel()

	t.Run("ForceSync should work", func(t *testing.T) {
		t.Parallel()

		numCalls := 0

		st := ntp.NewSyncTime(
			config.NTPConfig{
				SyncPeriodSeconds:    3600,
				Hosts:                []string{"host1"},
				OutOfBoundsThreshold: 2,
			},
			func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
				numCalls++

				time.Sleep(1 * time.Millisecond)

				return &beevikNtp.Response{
					ClockOffset: 1 * time.Millisecond,
				}, nil
			},
		)

		currentValue := 3 * time.Millisecond
		st.SetClockOffset(currentValue)

		st.ForceSync()

		time.Sleep(time.Duration(ntp.NumRequestsFromHost+5) * time.Millisecond)

		expClockOffset := 1 * time.Millisecond
		assert.Equal(t, expClockOffset, st.ClockOffset())

		require.Equal(t, ntp.NumRequestsFromHost, numCalls)
	})

	t.Run("TriggerSync should not trigger multiple times", func(t *testing.T) {
		t.Parallel()

		numCalls := &atomic.Uint32{}

		st := ntp.NewSyncTime(
			config.NTPConfig{
				SyncPeriodSeconds:    3600,
				Hosts:                []string{"host1"},
				OutOfBoundsThreshold: 2,
			},
			func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
				numCalls.Add(1)

				time.Sleep(2 * time.Millisecond)

				return &beevikNtp.Response{
					ClockOffset: 1 * time.Millisecond,
				}, nil
			},
		)

		currentValue := 3 * time.Millisecond
		st.SetClockOffset(currentValue)

		// multiple calls should trigger multiple times
		st.TriggerSync()
		st.TriggerSync()
		st.TriggerSync()
		st.TriggerSync()

		expClockOffset := 1 * time.Millisecond
		assert.Equal(t, expClockOffset, st.ClockOffset())

		// should not trigger multiple times due to cooldown
		require.Equal(t, ntp.NumRequestsFromHost, int(numCalls.Load()))
	})

	t.Run("direct trigger should not trigger if already in progress", func(t *testing.T) {
		t.Parallel()

		numCalls := &atomic.Uint32{}

		st := ntp.NewSyncTime(
			config.NTPConfig{
				SyncPeriodSeconds:    3600,
				Hosts:                []string{"host1"},
				OutOfBoundsThreshold: 2,
			},
			func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
				numCalls.Add(1)

				time.Sleep(10 * time.Millisecond)

				return &beevikNtp.Response{
					ClockOffset: 1 * time.Millisecond,
				}, nil
			},
		)

		// multiple ForceSync calls should not trigger if already syncing
		go st.TriggerSync()
		st.ForceSync()
		st.ForceSync()
		st.ForceSync()

		time.Sleep(time.Duration(ntp.NumRequestsFromHost*10+10) * time.Millisecond)

		require.Equal(t, ntp.NumRequestsFromHost, int(numCalls.Load()))
	})

	t.Run("ForceSync should skip during cooldown", func(t *testing.T) {
		t.Parallel()

		numCalls := &atomic.Uint32{}

		st := ntp.NewSyncTime(
			config.NTPConfig{
				SyncPeriodSeconds:    3600,
				Hosts:                []string{"host1"},
				OutOfBoundsThreshold: 2,
			},
			func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
				numCalls.Add(1)

				return &beevikNtp.Response{
					ClockOffset: 1 * time.Millisecond,
				}, nil
			},
		)

		// set lastSyncTime to now, so cooldown is active
		st.SetLastSyncTime(time.Now())

		st.ForceSync()

		time.Sleep(50 * time.Millisecond)

		require.Equal(t, uint32(0), numCalls.Load())
	})

	t.Run("ForceSync should work after cooldown expires", func(t *testing.T) {
		t.Parallel()

		numCalls := &atomic.Uint32{}

		st := ntp.NewSyncTime(
			config.NTPConfig{
				SyncPeriodSeconds:    3600,
				Hosts:                []string{"host1"},
				OutOfBoundsThreshold: 2,
			},
			func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
				numCalls.Add(1)

				return &beevikNtp.Response{
					ClockOffset: 1 * time.Millisecond,
				}, nil
			},
		)

		// set lastSyncTime in the past, so cooldown has expired
		st.SetLastSyncTime(time.Now().Add(-ntp.SyncCooldownDuration - time.Second))

		st.ForceSync()

		time.Sleep(50 * time.Millisecond)

		require.Equal(t, uint32(ntp.NumRequestsFromHost), numCalls.Load())
	})

	t.Run("triggerSync should skip during cooldown", func(t *testing.T) {
		t.Parallel()

		numCalls := &atomic.Uint32{}

		st := ntp.NewSyncTime(
			config.NTPConfig{
				SyncPeriodSeconds:    3600,
				Hosts:                []string{"host1"},
				OutOfBoundsThreshold: 2,
			},
			func(options ntp.NTPOptions, hostIndex int) (*beevikNtp.Response, error) {
				numCalls.Add(1)

				return &beevikNtp.Response{
					ClockOffset: 1 * time.Millisecond,
				}, nil
			},
		)

		// set lastSyncTime to now, so cooldown is active
		st.SetLastSyncTime(time.Now())

		// triggerSync should still work regardless of cooldown
		st.TriggerSync()

		require.Equal(t, uint32(0), numCalls.Load())
	})
}

// On local machine, seems like average query time is ~35ms, e.g.:
// Avg response time from host: time.google.com is 42.928837ms
// Avg response time from host: time.cloudflare.com is 13.877162ms
// Avg response time from host: time.apple.com is 37.257168ms
// Avg response time from host: time.windows.com is 48.33448ms
// Global average response time is 35.599412ms
func TestCallQueryShouldWorkMeasurements(t *testing.T) {
	t.Skip("use this test only for local benchmarks, not for remote tests, since it relies on internet connection")
	t.Parallel()

	ntpConfig := testscommon.NewNTPGoogleConfig()
	ntpOptions := ntp.NewNTPOptions(ntpConfig)
	st := ntp.NewSyncTime(ntpConfig, nil)

	query := st.Query()

	numRequestsFromHost := 10
	timeDurations := make(map[string][]time.Duration, len(ntpOptions.Hosts))

	var totalGlobalDuration time.Duration
	var totalRequests int

	for hostIndex := 0; hostIndex < len(ntpOptions.Hosts); hostIndex++ {
		for requests := 0; requests < numRequestsFromHost; requests++ {

			hostName := ntpOptions.Hosts[hostIndex]
			startTime := time.Now()
			response, err := query(ntpOptions, hostIndex)
			duration := time.Since(startTime)

			fmt.Printf("-> Query from host: %s function execution time: %s\n", hostName, duration)

			require.NotNil(t, response)
			require.Nil(t, err)

			timeDurations[hostName] = append(timeDurations[hostName], duration)

			totalGlobalDuration += duration
			totalRequests++
		}
	}

	for _, hostName := range ntpOptions.Hosts {
		durations := timeDurations[hostName]
		var totalDuration time.Duration
		for _, d := range durations {
			totalDuration += d
		}
		avgTimePerHost := totalDuration / time.Duration(len(durations))
		fmt.Printf("Avg response time from host: %s is %s\n", hostName, avgTimePerHost)
	}

	avgGlobalTime := totalGlobalDuration / time.Duration(totalRequests)
	fmt.Printf("Global average response time is %s\n", avgGlobalTime)
}

func TestGetMedianOffset(t *testing.T) {
	t.Parallel()

	st := ntp.NewSyncTime(config.NTPConfig{SyncPeriodSeconds: 1, OutOfBoundsThreshold: 200}, nil)

	t.Run("empty slice should return error", func(t *testing.T) {
		t.Parallel()

		offset, err := st.GetMedianOffset([]time.Duration{})
		require.Equal(t, ntp.ErrNoClockOffsets, err)
		require.Equal(t, time.Duration(0), offset)
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()

		offset, err := st.GetMedianOffset([]time.Duration{42 * time.Millisecond})
		require.Nil(t, err)
		require.Equal(t, 42*time.Millisecond, offset)
	})

	t.Run("single negative element", func(t *testing.T) {
		t.Parallel()

		offset, err := st.GetMedianOffset([]time.Duration{-500 * time.Microsecond})
		require.Nil(t, err)
		require.Equal(t, -500*time.Microsecond, offset)
	})

	t.Run("odd count returns true middle element", func(t *testing.T) {
		t.Parallel()

		// Sorted: 10, 20, 30 \u2192 middle index = 1 \u2192 20
		clockOffsets := []time.Duration{30 * time.Millisecond, 10 * time.Millisecond, 20 * time.Millisecond}
		offset, err := st.GetMedianOffset(clockOffsets)
		require.Nil(t, err)
		require.Equal(t, 20*time.Millisecond, offset)
	})

	t.Run("even count returns average of two middle elements", func(t *testing.T) {
		t.Parallel()

		// Sorted: 10, 20, 30, 40 \u2192 median = (20 + 30) / 2 = 25
		clockOffsets := []time.Duration{40 * time.Millisecond, 10 * time.Millisecond, 30 * time.Millisecond, 20 * time.Millisecond}
		offset, err := st.GetMedianOffset(clockOffsets)
		require.Nil(t, err)
		require.Equal(t, 25*time.Millisecond, offset)
	})

	t.Run("all identical values", func(t *testing.T) {
		t.Parallel()

		val := 100 * time.Microsecond
		clockOffsets := []time.Duration{val, val, val, val, val}
		offset, err := st.GetMedianOffset(clockOffsets)
		require.Nil(t, err)
		require.Equal(t, val, offset)
	})

	t.Run("all negative values", func(t *testing.T) {
		t.Parallel()

		// Sorted: -50, -30, -20, -10, -5 -> middle index = 2 -> -20
		clockOffsets := []time.Duration{
			-10 * time.Millisecond,
			-50 * time.Millisecond,
			-20 * time.Millisecond,
			-5 * time.Millisecond,
			-30 * time.Millisecond,
		}
		offset, err := st.GetMedianOffset(clockOffsets)
		require.Nil(t, err)
		require.Equal(t, -20*time.Millisecond, offset)
	})

	t.Run("mixed positive and negative values", func(t *testing.T) {
		t.Parallel()

		// Sorted: -30, -10, 5, 20, 40 -> middle index = 2 -> 5
		clockOffsets := []time.Duration{
			20 * time.Millisecond,
			-10 * time.Millisecond,
			40 * time.Millisecond,
			-30 * time.Millisecond,
			5 * time.Millisecond,
		}
		offset, err := st.GetMedianOffset(clockOffsets)
		require.Nil(t, err)
		require.Equal(t, 5*time.Millisecond, offset)
	})

	t.Run("two elements returns average of both", func(t *testing.T) {
		t.Parallel()

		// Sorted: -6, 10 -> median = (-6 + 10) / 2 = 2
		clockOffsets := []time.Duration{10 * time.Millisecond, -6 * time.Millisecond}
		offset, err := st.GetMedianOffset(clockOffsets)
		require.Nil(t, err)
		require.Equal(t, 2*time.Millisecond, offset)
	})

	t.Run("zero offset among values", func(t *testing.T) {
		t.Parallel()

		// Sorted: -10, 0, 10 -> middle index = 1 -> 0
		clockOffsets := []time.Duration{10 * time.Millisecond, 0, -10 * time.Millisecond}
		offset, err := st.GetMedianOffset(clockOffsets)
		require.Nil(t, err)
		require.Equal(t, time.Duration(0), offset)
	})

	t.Run("realistic mixed positive and negative NTP offsets", func(t *testing.T) {
		t.Parallel()

		// 30 elements (even count): median = average of sorted[14] and sorted[15]
		// sorted[14] = -709033ns, sorted[15] = -706902ns -> (-709033 + -706902) / 2 = -707967
		expectedValue := -707967 * time.Nanosecond
		clockOffsets := []time.Duration{
			-1855712 * time.Nanosecond,
			-1621517 * time.Nanosecond,
			-1682624 * time.Nanosecond,
			-1732382 * time.Nanosecond,
			-1793740 * time.Nanosecond,
			-1739692 * time.Nanosecond,
			-1791143 * time.Nanosecond,
			-1680870 * time.Nanosecond,
			-1674741 * time.Nanosecond,
			-1678761 * time.Nanosecond,
			431740 * time.Nanosecond,
			384421 * time.Nanosecond,
			496821 * time.Nanosecond,
			289701 * time.Nanosecond,
			505729 * time.Nanosecond,
			551695 * time.Nanosecond,
			264902 * time.Nanosecond,
			336397 * time.Nanosecond,
			426982 * time.Nanosecond,
			349654 * time.Nanosecond,
			-717224 * time.Nanosecond,
			-706902 * time.Nanosecond,
			-709033 * time.Nanosecond,
			-613281 * time.Nanosecond,
			-705814 * time.Nanosecond,
			-691355 * time.Nanosecond,
			-602491 * time.Nanosecond,
			-733157 * time.Nanosecond,
			-754736 * time.Nanosecond,
			-732048 * time.Nanosecond,
		}

		offset, err := st.GetMedianOffset(clockOffsets)
		require.Nil(t, err)
		require.Equal(t, expectedValue, offset)
	})

	t.Run("outlier does not affect median", func(t *testing.T) {
		t.Parallel()

		// Sorted: -1, 10, 11, 12, 1000000 -> middle index = 2 -> 11
		clockOffsets := []time.Duration{
			10 * time.Millisecond,
			12 * time.Millisecond,
			1000000 * time.Millisecond,
			-1 * time.Millisecond,
			11 * time.Millisecond,
		}
		offset, err := st.GetMedianOffset(clockOffsets)
		require.Nil(t, err)
		require.Equal(t, 11*time.Millisecond, offset)
	})
}
