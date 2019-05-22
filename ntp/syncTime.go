package ntp

import (
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/config"
	"github.com/beevik/ntp"
	beevikntp "github.com/beevik/ntp"
)

// totalRequests defines the number of requests made to determine an accurate clock offset
const totalRequests = 10

// NTPOptions defines configuration options for an NTP query
type NTPOptions struct {
	Host         string
	Version      int
	LocalAddress string
	Timeout      time.Duration
	Port         int
}

// NewNTPOptions creates a new NTPOptions object.
func NewNTPOptions(ntpConfig config.NTPConfig) NTPOptions {
	// TODO Read these values from configurations.
	return NTPOptions{
		Host:         ntpConfig.Host,
		Port:         ntpConfig.Port,
		Version:      ntpConfig.Version,
		LocalAddress: "",
		Timeout:      ntpConfig.Timeout}
}

// queryNTP wraps beevikntp.QueryWithOptions, in order to use NTPOptions, which
// contains both Host and Port, unlike beevikntp.QueryOptions.
func queryNTP(options NTPOptions) (*ntp.Response, error) {
	queryOptions := beevikntp.QueryOptions{
		Timeout:      options.Timeout,
		Version:      options.Version,
		LocalAddress: options.LocalAddress,
		Port:         options.Port}
	return beevikntp.QueryWithOptions(options.Host, queryOptions)
}

// syncTime defines an object for time synchronization
type syncTime struct {
	mut         sync.RWMutex
	clockOffset time.Duration
	syncPeriod  time.Duration
	ntpOptions  NTPOptions
	query       func(options NTPOptions) (*ntp.Response, error)
}

// NewSyncTime creates a syncTime object. The customQueryFunc argument allows
// the caller to set a different NTP-querying callback, if desired. If set to
// nil, then the default queryNTP is used.
func NewSyncTime(ntpConfig config.NTPConfig, syncPeriod time.Duration, customQueryFunc func(options NTPOptions) (*ntp.Response, error)) *syncTime {
	var queryFunc func(options NTPOptions) (*ntp.Response, error)
	if customQueryFunc == nil {
		queryFunc = queryNTP
	} else {
		queryFunc = customQueryFunc
	}
	s := syncTime{
		clockOffset: 0,
		syncPeriod:  syncPeriod,
		query:       queryFunc,
		ntpOptions:  NewNTPOptions(ntpConfig)}
	return &s
}

// StartSync method does the time synchronization at every syncPeriod time elapsed. This should be started
// as a go routine
func (s *syncTime) StartSync() {
	for {
		s.sync()
		time.Sleep(s.syncPeriod)
	}
}

// sync method does the time synchronization and sets the current offset difference between local time
// and server time with which it has done the synchronization
func (s *syncTime) sync() {
	if s.query != nil {
		clockOffsetSum := time.Duration(0)
		succeededRequests := 0

		for i := 0; i < totalRequests; i++ {
			r, err := s.query(s.ntpOptions)

			if err != nil {
				fmt.Println(fmt.Sprintf("NTP Error: %s", err))
				continue
			}

			fmt.Println(fmt.Sprintf("NTP reading: %s", r.Time.Format("Mon Jan 2 15:04:05 MST 2006")))

			succeededRequests++
			clockOffsetSum += r.ClockOffset
		}

		if succeededRequests > 0 {
			averageClockOffset := time.Duration(int64(clockOffsetSum) / int64(succeededRequests))
			s.setClockOffset(averageClockOffset)
		}
	}
}

// ClockOffset method gets the current time offset
func (s *syncTime) ClockOffset() time.Duration {
	s.mut.RLock()
	clockOffset := s.clockOffset
	s.mut.RUnlock()

	return clockOffset
}

func (s *syncTime) setClockOffset(clockOffset time.Duration) {
	s.mut.Lock()
	s.clockOffset = clockOffset
	s.mut.Unlock()
}

// FormattedCurrentTime method gets the formatted current time on which is added a given offset
func (s *syncTime) FormattedCurrentTime() string {
	return s.formatTime(s.CurrentTime())
}

// formatTime method gets the formatted time from a given time
func (s *syncTime) formatTime(time time.Time) string {
	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ", time.Year(), time.Month(), time.Day(), time.Hour(),
		time.Minute(), time.Second(), time.Nanosecond())
	return str
}

// CurrentTime method gets the current time on which is added the current offset
func (s *syncTime) CurrentTime() time.Time {
	return time.Now().Add(s.clockOffset)
}
