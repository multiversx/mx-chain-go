package resolver

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var topic = "topic"
var hash = []byte("hash")
var numIntra = 10
var numCross = 9

//------- NewInterceptorResolver

func createWorkableConfig() config.InterceptorResolverDebugConfig {
	return config.InterceptorResolverDebugConfig{
		Enabled:                    true,
		CacheSize:                  1000,
		EnableAutoPrint:            false,
		IntervalAutoPrintInSeconds: 0,
		NumRequestsThreshold:       0,
		NumResolveFailureThreshold: 0,
	}
}

func TestNewInterceptorResolver_InvalidSizeShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createWorkableConfig()
	cfg.CacheSize = -1
	ir, err := NewInterceptorResolver(cfg)

	assert.True(t, check.IfNil(ir))
	assert.NotNil(t, err)
}

func TestNewInterceptorResolver_InvalidIntervalShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createWorkableConfig()
	cfg.EnableAutoPrint = true
	cfg.IntervalAutoPrintInSeconds = 0
	ir, err := NewInterceptorResolver(cfg)

	assert.True(t, check.IfNil(ir))
	assert.True(t, errors.Is(err, debug.ErrInvalidValue))
}

func TestNewInterceptorResolver_NumResolveFailureThresholdShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createWorkableConfig()
	cfg.EnableAutoPrint = true
	cfg.IntervalAutoPrintInSeconds = 1
	cfg.NumResolveFailureThreshold = 0
	cfg.NumRequestsThreshold = 1
	ir, err := NewInterceptorResolver(cfg)

	assert.True(t, check.IfNil(ir))
	assert.True(t, errors.Is(err, debug.ErrInvalidValue))
}

func TestNewInterceptorResolver_NumRequestsThresholdShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createWorkableConfig()
	cfg.EnableAutoPrint = true
	cfg.IntervalAutoPrintInSeconds = 1
	cfg.NumResolveFailureThreshold = 1
	cfg.NumRequestsThreshold = 0
	ir, err := NewInterceptorResolver(cfg)

	assert.True(t, check.IfNil(ir))
	assert.True(t, errors.Is(err, debug.ErrInvalidValue))
}

func TestNewInterceptorResolver_DebugLineExpirationShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createWorkableConfig()
	cfg.EnableAutoPrint = true
	cfg.IntervalAutoPrintInSeconds = 1
	cfg.NumResolveFailureThreshold = 1
	cfg.NumRequestsThreshold = 1
	cfg.DebugLineExpiration = 0
	ir, err := NewInterceptorResolver(cfg)

	assert.True(t, check.IfNil(ir))
	assert.True(t, errors.Is(err, debug.ErrInvalidValue))
}

func TestNewInterceptorResolver_ShouldWork(t *testing.T) {
	t.Parallel()

	ir, err := NewInterceptorResolver(createWorkableConfig())

	assert.False(t, check.IfNil(ir))
	assert.Nil(t, err)
}

func TestNewInterceptorResolver_AutoprintShouldWork(t *testing.T) {
	t.Parallel()

	cfg := createWorkableConfig()
	cfg.EnableAutoPrint = true
	cfg.IntervalAutoPrintInSeconds = 1
	cfg.NumResolveFailureThreshold = 1
	cfg.NumRequestsThreshold = 1
	cfg.DebugLineExpiration = 100
	ir, err := NewInterceptorResolver(cfg)

	assert.False(t, check.IfNil(ir))
	assert.Nil(t, err)
}

//------- RequestedData

func TestInterceptorResolver_RequestedDataWithFiveIdentifiersShouldWork(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(createWorkableConfig())
	numIdentifiers := 5
	foundMap := make(map[string]struct{})
	for i := 0; i < numIdentifiers; i++ {
		newTopic := fmt.Sprintf("topic%d", i)
		ir.RequestedData(newTopic, hash, numIntra, numCross)
		foundMap[newTopic] = struct{}{}
	}

	s := strings.Join(ir.Query("*"), "\r\n")
	fmt.Println(s)

	events := ir.Events()
	require.Equal(t, numIdentifiers, len(events))
	for _, ev := range events {
		_, ok := foundMap[ev.topic]
		assert.True(t, ok, "topic: "+ev.topic)
		delete(foundMap, ev.topic)
	}

	assert.Equal(t, 0, len(foundMap))
}

func TestInterceptorResolver_RequestedDataSameIdentifierShouldAddRequested(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(createWorkableConfig())
	ir.RequestedData(topic, hash, numIntra, numCross)
	events := ir.Events()
	require.Equal(t, 1, len(events))
	expected := &event{
		eventType:   requestEvent,
		hash:        hash,
		topic:       topic,
		numReqIntra: numIntra,
		numReqCross: numCross,
	}

	assert.Equal(t, expected, events[0])

	ir.RequestedData(topic, hash, numIntra, numCross)
	events = ir.Events()
	require.Equal(t, 1, len(events))
	expected = &event{
		eventType:   requestEvent,
		hash:        hash,
		topic:       topic,
		numReqIntra: numIntra * 2,
		numReqCross: numCross * 2,
	}

	assert.Equal(t, expected, events[0])
	fmt.Println(ir.Query("*"))
}

//------- ProcessedHash

func TestInterceptorResolver_ProcessedHashNotFoundShouldNotAdd(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(createWorkableConfig())

	ir.ProcessedHash(topic, hash, nil)

	require.Equal(t, 0, len(ir.Events()))
}

func TestInterceptorResolver_ProcessedHashExistingNoErrorShouldRemove(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(createWorkableConfig())
	ir.RequestedData(topic, hash, numIntra, numCross)
	require.Equal(t, 1, len(ir.Events()))

	ir.ProcessedHash(topic, hash, nil)

	require.Equal(t, 0, len(ir.Events()))
}

func TestInterceptorResolver_ProcessedHashExistingWithErrorShouldIncrementProcessed(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(createWorkableConfig())
	ir.RequestedData(topic, hash, numIntra, numCross)
	require.Equal(t, 1, len(ir.Events()))

	err := errors.New("expected err")
	ir.ProcessedHash(topic, hash, err)

	requests := ir.Events()
	require.Equal(t, 1, len(requests))

	expected := &event{
		eventType:    requestEvent,
		hash:         hash,
		topic:        topic,
		numReqIntra:  numIntra,
		numReqCross:  numCross,
		lastErr:      err,
		numProcessed: 1,
		numReceived:  0,
	}

	assert.Equal(t, expected, requests[0])
	fmt.Println(ir.Query("*"))
}

//------- ReceivedHash

func TestInterceptorResolver_ReceivedHashNotFoundShouldNotAdd(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(createWorkableConfig())

	ir.ReceivedHash(topic, hash)

	require.Equal(t, 0, len(ir.Events()))
}

func TestInterceptorResolver_ReceiveddHashExistingShouldIncrementReceived(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(createWorkableConfig())
	ir.RequestedData(topic, hash, numIntra, numCross)
	require.Equal(t, 1, len(ir.Events()))

	ir.ReceivedHash(topic, hash)

	requests := ir.Events()
	require.Equal(t, 1, len(requests))

	expected := &event{
		eventType:    requestEvent,
		hash:         hash,
		topic:        topic,
		numReqIntra:  numIntra,
		numReqCross:  numCross,
		lastErr:      nil,
		numProcessed: 0,
		numReceived:  1,
	}

	assert.Equal(t, expected, requests[0])
}

//------- FailedToResolveData

func TestInterceptorResolver_FailedToResolveDataShouldWork(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(createWorkableConfig())
	ir.FailedToResolveData(topic, hash, nil)

	require.Equal(t, 1, len(ir.Events()))
	expected := &event{
		eventType:    resolveEvent,
		hash:         hash,
		topic:        topic,
		numReqIntra:  0,
		numReqCross:  0,
		lastErr:      nil,
		numProcessed: 0,
		numReceived:  1,
	}
	assert.Equal(t, expected, ir.Events()[0])

	ir.FailedToResolveData(topic, hash, nil)
	require.Equal(t, 1, len(ir.Events()))
	expected = &event{
		eventType:    resolveEvent,
		hash:         hash,
		topic:        topic,
		numReqIntra:  0,
		numReqCross:  0,
		lastErr:      nil,
		numProcessed: 0,
		numReceived:  2,
	}
	assert.Equal(t, expected, ir.Events()[0])
}

func TestInterceptorResolver_FailedToResolveDataAndRequestedDataShouldWork(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(createWorkableConfig())
	ir.FailedToResolveData(topic, hash, nil)

	assert.Equal(t, 1, len(ir.Events()))

	ir.RequestedData(topic, hash, numIntra, numCross)

	assert.Equal(t, 2, len(ir.Events()))
	fmt.Println(ir.Query("*"))
}

//------- functions

func TestInterceptorResolver_Query(t *testing.T) {
	t.Parallel()

	topic1 := "topic1"
	topic2 := "aaaa"
	ir, _ := NewInterceptorResolver(createWorkableConfig())
	ir.RequestedData(topic1, hash, numIntra, numCross)
	ir.RequestedData(topic2, hash, numIntra, numCross)

	assert.Equal(t, 0, len(ir.Query("not a topic")))
	assert.Equal(t, 1, len(ir.Query(topic1)))
	assert.Equal(t, 1, len(ir.Query(topic2)))
	assert.Equal(t, 2, len(ir.Query("*")))
}

func TestInterceptorResolver_Enabled(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(createWorkableConfig())

	assert.True(t, ir.Enabled())
}

func TestInterceptorResolver_GetStringEventsShouldWork(t *testing.T) {
	t.Parallel()

	ir, _ := NewInterceptorResolver(createWorkableConfig())
	assert.Equal(t, 0, len(ir.getStringEvents(100)))

	ir.FailedToResolveData(topic, hash, nil)
	ir.FailedToResolveData(topic, hash, nil)

	ir.RequestedData(topic, hash, 1, 1)

	assert.Equal(t, 2, len(ir.getStringEvents(100)))
}

func TestInterceptorResolver_NumPrintsShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numPrintCalls := uint32(0)
	cfg := createWorkableConfig()
	cfg.EnableAutoPrint = true
	cfg.IntervalAutoPrintInSeconds = 1
	cfg.NumResolveFailureThreshold = 1
	cfg.NumRequestsThreshold = 1
	cfg.DebugLineExpiration = 2
	ir, _ := NewInterceptorResolver(cfg)
	ir.printEventHandler = func(data string) {
		atomic.AddUint32(&numPrintCalls, 1)
	}

	ir.FailedToResolveData(topic, hash, nil)
	ir.FailedToResolveData(topic, hash, nil)

	time.Sleep(time.Second * 5)

	assert.Equal(t, uint32(cfg.DebugLineExpiration), atomic.LoadUint32(&numPrintCalls))
}
