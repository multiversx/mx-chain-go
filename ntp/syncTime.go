package ntp

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/closing"
	"github.com/beevik/ntp"
)

var _ SyncTimer = (*syncTime)(nil)
var _ closing.Closer = (*syncTime)(nil)

var log = logger.GetOrCreate("ntp")

// numRequestsFromHost represents the number of requests to be done from each host
const numRequestsFromHost = 10

// cuttingOutPercent [0, 1) represents the percent of received clock offsets to be removed from the edges (min and max)
const cuttingOutPercent = 0.3

// minResponsesPercent (0, 1] represents the minimum percent of responses, from all requests done, needed to set a new
// clock offset
const minResponsesPercent = 0.25

// maxOffsetPercent [0, 1) represents the maximum percent, from the initial sync period given, which could be added or
// subtracted from it
const maxOffsetPercent = 0.2

// minTimeout represents the minimum time in milliseconds to wait for a response from a host after a NTP request
const minTimeout = 100

const outOfBoundsDuration = time.Second

// NTPOptions defines configuration options for a NTP query
type NTPOptions struct {
	Hosts        []string
	Version      int
	LocalAddress string
	Timeout      time.Duration
	Port         int
}

// NewNTPGoogleConfig creates an NTPConfig object that configures NTP to use a predefined list of hosts. This is useful
// for tests, for example, to avoid loading a configuration file just to have a NTPConfig
func NewNTPGoogleConfig() config.NTPConfig {
	return config.NTPConfig{
		Hosts:               []string{"time.google.com", "time.cloudflare.com", "time.apple.com", "time.windows.com"},
		Port:                123,
		Version:             0,
		TimeoutMilliseconds: 100,
		SyncPeriodSeconds:   3600,
	}
}

// NewNTPOptions creates a new NTPOptions object
func NewNTPOptions(ntpConfig config.NTPConfig) NTPOptions {
	ntpConfig.TimeoutMilliseconds = core.MaxInt(minTimeout, ntpConfig.TimeoutMilliseconds)
	timeout := time.Duration(ntpConfig.TimeoutMilliseconds) * time.Millisecond

	return NTPOptions{
		Hosts:        ntpConfig.Hosts,
		Port:         ntpConfig.Port,
		Version:      ntpConfig.Version,
		LocalAddress: "",
		Timeout:      timeout,
	}
}

// queryNTP wraps beevikntp.QueryWithOptions, in order to use NTPOptions, which contains both Host and Port, unlike
// beevikntp.QueryOptions
func queryNTP(options NTPOptions, hostIndex int) (*ntp.Response, error) {
	if hostIndex >= len(options.Hosts) {
		return nil, ErrIndexOutOfBounds
	}

	queryOptions := ntp.QueryOptions{
		Timeout:      options.Timeout,
		Version:      options.Version,
		LocalAddress: options.LocalAddress,
		Port:         options.Port,
	}

	return ntp.QueryWithOptions(options.Hosts[hostIndex], queryOptions)
}

// syncTime defines an object for time synchronization
type syncTime struct {
	mut         sync.RWMutex
	clockOffset time.Duration
	syncPeriod  time.Duration
	ntpOptions  NTPOptions
	query       func(options NTPOptions, hostIndex int) (*ntp.Response, error)
	cancelFunc  func()
}

// NewSyncTime creates a syncTime object. The customQueryFunc argument allows the caller to set a different NTP-querying
// callback, if desired. If set to nil, then the default queryNTP is used
func NewSyncTime(
	ntpConfig config.NTPConfig,
	customQueryFunc func(options NTPOptions, hostIndex int) (*ntp.Response, error),
) *syncTime {
	queryFunc := customQueryFunc
	if queryFunc == nil {
		queryFunc = queryNTP
	}

	s := syncTime{
		clockOffset: 0,
		syncPeriod:  time.Duration(ntpConfig.SyncPeriodSeconds) * time.Second,
		query:       queryFunc,
		ntpOptions:  NewNTPOptions(ntpConfig),
	}

	return &s
}

// StartSyncingTime method does the time synchronization at every syncPeriod time elapsed. This method should be started on go
// routine
func (s *syncTime) StartSyncingTime() {
	var ctx context.Context
	ctx, s.cancelFunc = context.WithCancel(context.Background())
	go s.startSync(ctx)
}

func (s *syncTime) startSync(ctx context.Context) {
	for {
		s.sync()

		select {
		case <-ctx.Done():
			log.Debug("syncTime's go routine is stopping...")
			return
		case <-time.After(s.getSleepTime()):
		}
	}
}

func (s *syncTime) getSleepTime() time.Duration {
	maxOffset := int64(float64(s.syncPeriod) * maxOffsetPercent)
	maxRandValueToGenerate := maxOffset * 2
	randBigInt, err := rand.Int(rand.Reader, big.NewInt(maxRandValueToGenerate+1))
	if err != nil {
		return s.syncPeriod
	}

	offset := randBigInt.Int64() - maxOffset
	return s.syncPeriod + time.Duration(offset)
}

// sync method does the time synchronization and sets the harmonic mean offset difference between local time
// and servers time which have been used in synchronization
func (s *syncTime) sync() {
	clockOffsets := make([]time.Duration, 0)
	for hostIndex := 0; hostIndex < len(s.ntpOptions.Hosts); hostIndex++ {
		for requests := 0; requests < numRequestsFromHost; requests++ {
			response, err := s.query(s.ntpOptions, hostIndex)
			if err != nil {
				log.Debug("sync.query",
					"host", s.ntpOptions.Hosts[hostIndex],
					"port", s.ntpOptions.Port,
					"error", err.Error())

				continue
			}

			log.Trace("sync.query",
				"host", s.ntpOptions.Hosts[hostIndex],
				"reference time", response.ReferenceTime.Format("Mon Jan 2 15:04:05 MST 2006"),
				"time", response.Time.Format("Mon Jan 2 15:04:05 MST 2006"),
				"precision", response.Precision,
				"clock offset", response.ClockOffset,
			)

			clockOffsets = append(clockOffsets, response.ClockOffset)
		}
	}

	numTotalRequests := len(s.ntpOptions.Hosts) * numRequestsFromHost
	minClockOffsetsToAllowUpdate := math.Ceil(float64(numTotalRequests) * minResponsesPercent / (1 - cuttingOutPercent))
	if len(clockOffsets) < int(minClockOffsetsToAllowUpdate) {
		log.Debug("sync.setClockOffset NOT done",
			"clock offsets", len(clockOffsets),
			"min clock offsets to allow update", int(minClockOffsetsToAllowUpdate))

		return
	}

	clockOffsetsWithoutEdges := s.getClockOffsetsWithoutEdges(clockOffsets)
	clockOffsetHarmonicMean := s.getHarmonicMean(clockOffsetsWithoutEdges)
	isOutOfBounds := core.AbsDuration(clockOffsetHarmonicMean)-outOfBoundsDuration > 0
	if isOutOfBounds {
		log.Error("syncTime.sync: clock offset is out of expected bounds",
			"clock offset harmonic mean", clockOffsetHarmonicMean)

		return
	}

	s.setClockOffset(clockOffsetHarmonicMean)

	log.Debug("sync.setClockOffset done",
		"num clock offsets", len(clockOffsets),
		"num clock offsets without edges", len(clockOffsetsWithoutEdges),
		"clock offset harmonic mean", clockOffsetHarmonicMean)
}

func (s *syncTime) getClockOffsetsWithoutEdges(clockOffsets []time.Duration) []time.Duration {
	sort.Slice(clockOffsets, func(i, j int) bool {
		return clockOffsets[i] < clockOffsets[j]
	})

	cuttingOutPercentPerEdge := cuttingOutPercent / 2
	startIndex := math.Floor(float64(len(clockOffsets)) * cuttingOutPercentPerEdge)
	endIndex := math.Ceil(float64(len(clockOffsets)) * (1 - cuttingOutPercentPerEdge))

	return clockOffsets[int(startIndex):int(endIndex)]
}

func (s *syncTime) getHarmonicMean(clockOffsets []time.Duration) time.Duration {
	inverseClockOffsetSum := float64(0)
	for index, clockOffset := range clockOffsets {
		if clockOffset == 0 {
			return time.Duration(0)
		}

		inverseClockOffsetSum += 1 / float64(clockOffset)

		log.Trace("getHarmonicMean",
			"index", index,
			"clock offset", clockOffset,
			"inverse clock offset sum", inverseClockOffsetSum)
	}

	if inverseClockOffsetSum == 0 {
		return time.Duration(0)
	}

	harmonicMean := float64(len(clockOffsets)) / inverseClockOffsetSum
	//TODO: figure out why do we need to add 0.5 here
	return time.Duration(harmonicMean + 0.5)
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

// FormattedCurrentTime method gets the formatted current time on which is added the current offset
func (s *syncTime) FormattedCurrentTime() string {
	return s.formatTime(s.CurrentTime())
}

// formatTime method gets the formatted time for a given time
func (s *syncTime) formatTime(time time.Time) string {
	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ",
		time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute(), time.Second(), time.Nanosecond())
	return str
}

// CurrentTime method gets the current time on which is added the current offset
func (s *syncTime) CurrentTime() time.Time {
	s.mut.RLock()
	currentTime := time.Now().Add(s.clockOffset)
	s.mut.RUnlock()

	return currentTime
}

// Close will close the endless running go routine
func (s *syncTime) Close() error {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *syncTime) IsInterfaceNil() bool {
	return s == nil
}
