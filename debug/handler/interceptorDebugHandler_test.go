package handler

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/debug"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var topic = "topic"
var hash = []byte("hash")
var numIntra = 10
var numCross = 9

func createWorkableConfig() config.InterceptorResolverDebugConfig {
	return config.InterceptorResolverDebugConfig{
		Enabled:                    true,
		CacheSize:                  1000,
		EnablePrint:                false,
		IntervalAutoPrintInSeconds: 0,
		NumRequestsThreshold:       0,
		NumResolveFailureThreshold: 0,
	}
}

func mockTimestampHandler() int64 {
	return 22342
}

func TestNewInterceptorResolver_InvalidSizeShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createWorkableConfig()
	cfg.CacheSize = -1
	idh, err := NewInterceptorDebugHandler(cfg)

	assert.True(t, check.IfNil(idh))
	assert.NotNil(t, err)
}

func TestNewInterceptorResolver_InvalidIntervalShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createWorkableConfig()
	cfg.EnablePrint = true
	cfg.IntervalAutoPrintInSeconds = 0
	idh, err := NewInterceptorDebugHandler(cfg)

	assert.True(t, check.IfNil(idh))
	assert.True(t, errors.Is(err, debug.ErrInvalidValue))
}

func TestNewInterceptorResolver_NumResolveFailureThresholdShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createWorkableConfig()
	cfg.EnablePrint = true
	cfg.IntervalAutoPrintInSeconds = 1
	cfg.NumResolveFailureThreshold = 0
	cfg.NumRequestsThreshold = 1
	idh, err := NewInterceptorDebugHandler(cfg)

	assert.True(t, check.IfNil(idh))
	assert.True(t, errors.Is(err, debug.ErrInvalidValue))
}

func TestNewInterceptorResolver_NumRequestsThresholdShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createWorkableConfig()
	cfg.EnablePrint = true
	cfg.IntervalAutoPrintInSeconds = 1
	cfg.NumResolveFailureThreshold = 1
	cfg.NumRequestsThreshold = 0
	idh, err := NewInterceptorDebugHandler(cfg)

	assert.True(t, check.IfNil(idh))
	assert.True(t, errors.Is(err, debug.ErrInvalidValue))
}

func TestNewInterceptorResolver_DebugLineExpirationShouldErr(t *testing.T) {
	t.Parallel()

	cfg := createWorkableConfig()
	cfg.EnablePrint = true
	cfg.IntervalAutoPrintInSeconds = 1
	cfg.NumResolveFailureThreshold = 1
	cfg.NumRequestsThreshold = 1
	cfg.DebugLineExpiration = 0
	idh, err := NewInterceptorDebugHandler(cfg)

	assert.True(t, check.IfNil(idh))
	assert.True(t, errors.Is(err, debug.ErrInvalidValue))
}

func TestNewInterceptorResolver_ShouldWork(t *testing.T) {
	t.Parallel()

	idh, err := NewInterceptorDebugHandler(createWorkableConfig())

	assert.False(t, check.IfNil(idh))
	assert.Nil(t, err)
}

func TestNewInterceptorResolver_EnablePrintShouldWork(t *testing.T) {
	t.Parallel()

	cfg := createWorkableConfig()
	cfg.EnablePrint = true
	cfg.IntervalAutoPrintInSeconds = 1
	cfg.NumResolveFailureThreshold = 1
	cfg.NumRequestsThreshold = 1
	cfg.DebugLineExpiration = 100
	idh, err := NewInterceptorDebugHandler(cfg)

	assert.False(t, check.IfNil(idh))
	assert.Nil(t, err)
}

func TestInterceptorResolver_LogRequestedDataWithFiveIdentifiersShouldWork(t *testing.T) {
	t.Parallel()

	idh, _ := NewInterceptorDebugHandler(createWorkableConfig())
	idh.SetTimeHandler(mockTimestampHandler)
	numIdentifiers := 5
	foundMap := make(map[string]struct{})
	for i := 0; i < numIdentifiers; i++ {
		newTopic := fmt.Sprintf("topic%d", i)
		idh.LogRequestedData(newTopic, [][]byte{hash}, numIntra, numCross)
		foundMap[newTopic] = struct{}{}
	}

	s := strings.Join(idh.Query("*"), "\r\n")
	fmt.Println(s)

	events := idh.Events()
	require.Equal(t, numIdentifiers, len(events))
	for _, ev := range events {
		_, ok := foundMap[ev.topic]
		assert.True(t, ok, "topic: "+ev.topic)
		delete(foundMap, ev.topic)
	}

	assert.Equal(t, 0, len(foundMap))
}

func TestInterceptorResolver_LogRequestedDataSameIdentifierShouldAddRequested(t *testing.T) {
	t.Parallel()

	idh, _ := NewInterceptorDebugHandler(createWorkableConfig())
	idh.SetTimeHandler(mockTimestampHandler)
	idh.LogRequestedData(topic, [][]byte{hash}, numIntra, numCross)
	events := idh.Events()
	require.Equal(t, 1, len(events))
	expected := &event{
		eventType:   requestEvent,
		hash:        hash,
		topic:       topic,
		numReqIntra: numIntra,
		numReqCross: numCross,
		timestamp:   mockTimestampHandler(),
	}

	assert.Equal(t, expected, events[0])

	idh.LogRequestedData(topic, [][]byte{hash}, numIntra, numCross)
	events = idh.Events()
	require.Equal(t, 1, len(events))
	expected = &event{
		eventType:   requestEvent,
		hash:        hash,
		topic:       topic,
		numReqIntra: numIntra * 2,
		numReqCross: numCross * 2,
		timestamp:   mockTimestampHandler(),
	}

	assert.Equal(t, expected, events[0])
	fmt.Println(idh.Query("*"))
}

//------- LogProcessedHashes

func TestInterceptorResolver_LogProcessedHashesNotFoundShouldNotAdd(t *testing.T) {
	t.Parallel()

	idh, _ := NewInterceptorDebugHandler(createWorkableConfig())

	idh.LogProcessedHashes(topic, [][]byte{hash}, nil)

	require.Equal(t, 0, len(idh.Events()))
}

func TestInterceptorResolver_LogProcessedHashesExistingNoErrorShouldRemove(t *testing.T) {
	t.Parallel()

	idh, _ := NewInterceptorDebugHandler(createWorkableConfig())
	idh.LogRequestedData(topic, [][]byte{hash}, numIntra, numCross)
	require.Equal(t, 1, len(idh.Events()))

	idh.LogProcessedHashes(topic, [][]byte{hash}, nil)

	require.Equal(t, 0, len(idh.Events()))
}

func TestInterceptorResolver_LogProcessedHashesExistingWithErrorShouldIncrementProcessed(t *testing.T) {
	t.Parallel()

	idh, _ := NewInterceptorDebugHandler(createWorkableConfig())
	idh.SetTimeHandler(mockTimestampHandler)
	idh.LogRequestedData(topic, [][]byte{hash}, numIntra, numCross)
	require.Equal(t, 1, len(idh.Events()))

	err := errors.New("expected err")
	idh.LogProcessedHashes(topic, [][]byte{hash}, err)

	requests := idh.Events()
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
		timestamp:    mockTimestampHandler(),
	}

	assert.Equal(t, expected, requests[0])
	fmt.Println(idh.Query("*"))
}

//------- LogReceivedHashes

func TestInterceptorResolver_LogReceivedHashesNotFoundShouldNotAdd(t *testing.T) {
	t.Parallel()

	idh, _ := NewInterceptorDebugHandler(createWorkableConfig())

	idh.LogReceivedHashes(topic, [][]byte{hash})

	require.Equal(t, 0, len(idh.Events()))
}

func TestInterceptorResolver_LogReceivedHashesExistingShouldIncrementReceived(t *testing.T) {
	t.Parallel()

	idh, _ := NewInterceptorDebugHandler(createWorkableConfig())
	idh.SetTimeHandler(mockTimestampHandler)
	idh.LogRequestedData(topic, [][]byte{hash}, numIntra, numCross)
	require.Equal(t, 1, len(idh.Events()))

	idh.LogReceivedHashes(topic, [][]byte{hash})

	requests := idh.Events()
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
		timestamp:    mockTimestampHandler(),
	}

	assert.Equal(t, expected, requests[0])
}

//------- LogFailedToResolveData

func TestInterceptorResolver_LogFailedToResolveDataShouldWork(t *testing.T) {
	t.Parallel()

	idh, _ := NewInterceptorDebugHandler(createWorkableConfig())
	idh.SetTimeHandler(mockTimestampHandler)
	idh.LogFailedToResolveData(topic, hash, nil)

	require.Equal(t, 1, len(idh.Events()))
	expected := &event{
		eventType:    resolveEvent,
		hash:         hash,
		topic:        topic,
		numReqIntra:  0,
		numReqCross:  0,
		lastErr:      nil,
		numProcessed: 0,
		numReceived:  1,
		timestamp:    mockTimestampHandler(),
	}
	assert.Equal(t, expected, idh.Events()[0])

	idh.LogFailedToResolveData(topic, hash, nil)
	require.Equal(t, 1, len(idh.Events()))
	expected = &event{
		eventType:    resolveEvent,
		hash:         hash,
		topic:        topic,
		numReqIntra:  0,
		numReqCross:  0,
		lastErr:      nil,
		numProcessed: 0,
		numReceived:  2,
		timestamp:    mockTimestampHandler(),
	}
	assert.Equal(t, expected, idh.Events()[0])
}

func TestInterceptorResolver_LogFailedToResolveDataAndRequestedDataShouldWork(t *testing.T) {
	t.Parallel()

	idh, _ := NewInterceptorDebugHandler(createWorkableConfig())
	idh.LogFailedToResolveData(topic, hash, nil)

	assert.Equal(t, 1, len(idh.Events()))

	idh.LogRequestedData(topic, [][]byte{hash}, numIntra, numCross)

	assert.Equal(t, 2, len(idh.Events()))
	fmt.Println(idh.Query("*"))
}

//------- LogSucceedToResolveData

func TestInterceptorResolver_LogSucceededToResolveDataShouldWork(t *testing.T) {
	t.Parallel()

	idh, _ := NewInterceptorDebugHandler(createWorkableConfig())
	idh.SetTimeHandler(mockTimestampHandler)
	idh.LogFailedToResolveData(topic, hash, nil)

	require.Equal(t, 1, len(idh.Events()))
	expected := &event{
		eventType:    resolveEvent,
		hash:         hash,
		topic:        topic,
		numReqIntra:  0,
		numReqCross:  0,
		lastErr:      nil,
		numProcessed: 0,
		numReceived:  1,
		timestamp:    mockTimestampHandler(),
	}
	assert.Equal(t, expected, idh.Events()[0])

	idh.LogSucceededToResolveData(topic, hash)
	assert.Equal(t, 0, len(idh.Events()))
}

//------- functions

func TestInterceptorResolver_Query(t *testing.T) {
	t.Parallel()

	topic1 := "topic1"
	topic2 := "aaaa"
	idh, _ := NewInterceptorDebugHandler(createWorkableConfig())
	idh.LogRequestedData(topic1, [][]byte{hash}, numIntra, numCross)
	idh.LogRequestedData(topic2, [][]byte{hash}, numIntra, numCross)

	assert.Equal(t, 0, len(idh.Query("not a topic")))
	assert.Equal(t, 1, len(idh.Query(topic1)))
	assert.Equal(t, 1, len(idh.Query(topic2)))
	assert.Equal(t, 2, len(idh.Query("*")))
}

func TestInterceptorResolver_GetStringEventsShouldWork(t *testing.T) {
	t.Parallel()

	idh, _ := NewInterceptorDebugHandler(createWorkableConfig())
	assert.Equal(t, 0, len(idh.getStringEvents(100)))

	idh.LogFailedToResolveData(topic, hash, nil)
	idh.LogFailedToResolveData(topic, hash, nil)

	idh.LogRequestedData(topic, [][]byte{hash}, 1, 1)

	assert.Equal(t, 2, len(idh.getStringEvents(100)))
}

func TestInterceptorResolver_NumPrintsShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numPrintCalls := uint32(0)
	cfg := createWorkableConfig()
	cfg.EnablePrint = true
	cfg.IntervalAutoPrintInSeconds = 1
	cfg.NumResolveFailureThreshold = 1
	cfg.NumRequestsThreshold = 1
	cfg.DebugLineExpiration = 2
	idh, _ := NewInterceptorDebugHandler(cfg)
	idh.printEventFunc = func(data string) {
		atomic.AddUint32(&numPrintCalls, 1)
	}

	idh.LogFailedToResolveData(topic, hash, nil)
	idh.LogFailedToResolveData(topic, hash, nil)

	time.Sleep(time.Second * 5)

	assert.Equal(t, uint32(cfg.DebugLineExpiration), atomic.LoadUint32(&numPrintCalls))
}
