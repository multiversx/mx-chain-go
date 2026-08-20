package p2p

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"

	communicationP2P "github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// the pubsub tracer only forwards discarded messages to a debugger implementing this interface
var _ communicationP2P.DiscardedMessagesDebugger = (*p2pDebugger)(nil)
var _ communicationP2P.Debugger = (*p2pDebugger)(nil)

func mockPrintFn(string) {}
func shouldCompute() bool {
	return true
}
func shouldNotCompute() bool {
	return false
}

func TestNewP2PDebugger(t *testing.T) {
	t.Parallel()

	pd := NewP2PDebugger("")

	assert.False(t, check.IfNil(pd))
}

// ------- AddIncomingMessage

func TestP2pDebugger_AddIncomingMessageShouldNotProcessWillNotAdd(t *testing.T) {
	t.Parallel()

	pd := newTestP2PDebugger(
		"",
		shouldNotCompute,
		mockPrintFn,
	)

	topic := "topic"
	size := uint64(3857)
	pd.AddIncomingMessage(topic, size, false)

	m := pd.GetClonedMetric(topic)
	assert.Nil(t, m)
}

func TestP2pDebugger_AddIncomingMessage(t *testing.T) {
	t.Parallel()

	pd := newTestP2PDebugger(
		"",
		shouldCompute,
		mockPrintFn,
	)

	topic := "topic"
	size := uint64(3857)
	pd.AddIncomingMessage(topic, size, false)

	m := pd.GetClonedMetric(topic)
	require.NotNil(t, m)

	expectedMetric := &metric{
		topic:                topic,
		incomingSize:         size,
		incomingNum:          1,
		incomingRejectedSize: 0,
		incomingRejectedNum:  0,
		outgoingSize:         0,
		outgoingNum:          0,
		outgoingRejectedSize: 0,
		outgoingRejectedNum:  0,
	}
	assert.Equal(t, expectedMetric, m)

	pd.AddIncomingMessage(topic, size, true)
	m = pd.GetClonedMetric(topic)
	require.NotNil(t, m)

	expectedMetric = &metric{
		topic:                topic,
		incomingSize:         size + size,
		incomingNum:          2,
		incomingRejectedSize: size,
		incomingRejectedNum:  1,
		outgoingSize:         0,
		outgoingNum:          0,
		outgoingRejectedSize: 0,
		outgoingRejectedNum:  0,
	}
	assert.Equal(t, expectedMetric, m)
}

// ------- AddOutgoingMessage

func TestP2pDebugger_AddOutgoingMessageShouldNotProcessWillNotAdd(t *testing.T) {
	t.Parallel()

	pd := newTestP2PDebugger(
		"",
		shouldNotCompute,
		mockPrintFn,
	)

	topic := "topic"
	size := uint64(3857)
	pd.AddOutgoingMessage(topic, size, false)

	m := pd.GetClonedMetric(topic)
	assert.Nil(t, m)
}

func TestP2pDebugger_AddOutgoingMessage(t *testing.T) {
	t.Parallel()

	pd := newTestP2PDebugger(
		"",
		shouldCompute,
		mockPrintFn,
	)

	topic := "topic"
	size := uint64(3857)
	pd.AddOutgoingMessage(topic, size, false)

	m := pd.GetClonedMetric(topic)
	require.NotNil(t, m)

	expectedMetric := &metric{
		topic:                topic,
		incomingSize:         0,
		incomingNum:          0,
		incomingRejectedSize: 0,
		incomingRejectedNum:  0,
		outgoingSize:         size,
		outgoingNum:          1,
		outgoingRejectedSize: 0,
		outgoingRejectedNum:  0,
	}
	assert.Equal(t, expectedMetric, m)

	pd.AddOutgoingMessage(topic, size, true)
	m = pd.GetClonedMetric(topic)
	require.NotNil(t, m)

	expectedMetric = &metric{
		topic:                topic,
		incomingSize:         0,
		incomingNum:          0,
		incomingRejectedSize: 0,
		incomingRejectedNum:  0,
		outgoingSize:         size + size,
		outgoingNum:          2,
		outgoingRejectedSize: size,
		outgoingRejectedNum:  1,
	}
	assert.Equal(t, expectedMetric, m)
}

// ------- continuouslyPrintStatistics

// ------- AddDuplicateMessage

func TestP2pDebugger_AddDuplicateMessageShouldNotProcessWillNotAdd(t *testing.T) {
	t.Parallel()

	pd := newTestP2PDebugger(
		"",
		shouldNotCompute,
		mockPrintFn,
	)

	topic := "topic"
	pd.AddDuplicateMessage(topic, uint64(3857))

	m := pd.GetClonedMetric(topic)
	assert.Nil(t, m)
}

func TestP2pDebugger_AddDuplicateMessage(t *testing.T) {
	t.Parallel()

	pd := newTestP2PDebugger(
		"",
		shouldCompute,
		mockPrintFn,
	)

	topic := "topic"
	size := uint64(3857)
	pd.AddDuplicateMessage(topic, size)

	m := pd.GetClonedMetric(topic)
	require.NotNil(t, m)

	expectedMetric := &metric{
		topic:         topic,
		duplicateSize: size,
		duplicateNum:  1,
	}
	assert.Equal(t, expectedMetric, m)

	pd.AddDuplicateMessage(topic, size)
	m = pd.GetClonedMetric(topic)
	require.NotNil(t, m)

	expectedMetric = &metric{
		topic:         topic,
		duplicateSize: size * 2,
		duplicateNum:  2,
	}
	assert.Equal(t, expectedMetric, m)
}

// ------- AddIgnoredMessage

func TestP2pDebugger_AddIgnoredMessageShouldNotProcessWillNotAdd(t *testing.T) {
	t.Parallel()

	pd := newTestP2PDebugger(
		"",
		shouldNotCompute,
		mockPrintFn,
	)

	topic := "topic"
	pd.AddIgnoredMessage(topic, uint64(3857))

	m := pd.GetClonedMetric(topic)
	assert.Nil(t, m)
}

func TestP2pDebugger_AddIgnoredMessage(t *testing.T) {
	t.Parallel()

	pd := newTestP2PDebugger(
		"",
		shouldCompute,
		mockPrintFn,
	)

	topic := "topic"
	size := uint64(3857)
	pd.AddIgnoredMessage(topic, size)
	pd.AddIgnoredMessage(topic, size)

	m := pd.GetClonedMetric(topic)
	require.NotNil(t, m)

	expectedMetric := &metric{
		topic:       topic,
		ignoredSize: size * 2,
		ignoredNum:  2,
	}
	assert.Equal(t, expectedMetric, m)
}

func TestP2pDebugger_duplicatesAndIgnoredAreCountedSeparately(t *testing.T) {
	t.Parallel()

	pd := newTestP2PDebugger(
		"",
		shouldCompute,
		mockPrintFn,
	)

	topic := "topic"
	pd.AddDuplicateMessage(topic, uint64(10))
	pd.AddIgnoredMessage(topic, uint64(20))

	m := pd.GetClonedMetric(topic)
	require.NotNil(t, m)

	expectedMetric := &metric{
		topic:         topic,
		duplicateSize: 10,
		duplicateNum:  1,
		ignoredSize:   20,
		ignoredNum:    1,
	}
	assert.Equal(t, expectedMetric, m)
}

func TestP2pDebugger_discardedAreReportedInStatsAndReset(t *testing.T) {
	t.Parallel()

	pd := newTestP2PDebugger(
		"",
		shouldCompute,
		mockPrintFn,
	)

	topic := "testTopic"
	size := uint64(5 * 1024) // 5kB
	pd.AddIncomingMessage(topic, size, false)
	pd.AddDuplicateMessage(topic, size)
	pd.AddIgnoredMessage(topic, size)

	str := pd.statsToString(1)
	assert.True(t, strings.Contains(str, "Incoming duplicates (num / size)"))
	assert.True(t, strings.Contains(str, "Incoming ignored (num / size)"))
	assert.True(t, strings.Contains(str, core.ConvertBytes(size)))

	assert.Nil(t, pd.GetClonedMetric(topic))
}

func TestP2pDebugger_continuouslyPrintStatisticsShouldNotPrint(t *testing.T) {
	t.Parallel()

	numPrintWasCalled := int32(0)
	_ = newTestP2PDebugger(
		"",
		shouldNotCompute,
		func(data string) {
			atomic.AddInt32(&numPrintWasCalled, 1)
		},
	)

	time.Sleep(printInterval * 3)

	assert.Equal(t, int32(0), atomic.LoadInt32(&numPrintWasCalled))
}

func TestP2pDebugger_continuouslyPrintStatisticsShouldPrint(t *testing.T) {
	t.Parallel()

	numPrintWasCalled := int32(0)
	_ = newTestP2PDebugger(
		"",
		shouldCompute,
		func(data string) {
			atomic.AddInt32(&numPrintWasCalled, 1)
		},
	)

	time.Sleep(printInterval*3 + time.Millisecond*100)

	assert.Equal(t, int32(3), atomic.LoadInt32(&numPrintWasCalled))
}

func TestP2pDebugger_continuouslyPrintStatisticsCloseShouldStop(t *testing.T) {
	t.Parallel()

	numPrintWasCalled := int32(0)
	pd := newTestP2PDebugger(
		"",
		shouldCompute,
		func(data string) {
			atomic.AddInt32(&numPrintWasCalled, 1)
		},
	)

	time.Sleep(printInterval*3 + time.Millisecond*100)
	assert.Equal(t, int32(3), atomic.LoadInt32(&numPrintWasCalled))

	err := pd.Close()
	assert.Nil(t, err)

	time.Sleep(printInterval*3 + time.Millisecond*100)
	assert.Equal(t, int32(3), atomic.LoadInt32(&numPrintWasCalled))
}

func TestP2pDebugger_statsToStringSortsByBytesOnTheWire(t *testing.T) {
	t.Parallel()

	pd := newTestP2PDebugger(
		"",
		shouldCompute,
		mockPrintFn,
	)

	// mostlyDuplicates carries fewer accepted bytes but far more bytes on the wire
	pd.AddIncomingMessage("mostlyDuplicates", 100, false)
	pd.AddDuplicateMessage("mostlyDuplicates", 10000)
	pd.AddIncomingMessage("mostlyAccepted", 1000, false)

	str := pd.statsToString(1)

	assert.Less(t, strings.Index(str, "mostlyDuplicates"), strings.Index(str, "mostlyAccepted"))
}

func TestP2pDebugger_statsToStringIgnoredIsNotDoubleCounted(t *testing.T) {
	t.Parallel()

	pd := newTestP2PDebugger(
		"",
		shouldCompute,
		mockPrintFn,
	)

	// the ignored bytes are part of the incoming ones, so they must not push the topic up the table
	pd.AddIncomingMessage("allIgnored", 1000, false)
	pd.AddIgnoredMessage("allIgnored", 1000)
	pd.AddIncomingMessage("plainTraffic", 1001, false)

	str := pd.statsToString(1)

	assert.Less(t, strings.Index(str, "plainTraffic"), strings.Index(str, "allIgnored"))
}

func TestP2pDebugger_statsToString(t *testing.T) {
	t.Parallel()

	pd := newTestP2PDebugger(
		"",
		shouldCompute,
		mockPrintFn,
	)

	topic1 := "testTopic1"
	size1 := uint64(3 * 1048576) // 3MB
	topic2 := "testTopic2"
	size2 := uint64(5 * 1024) // 5kB
	pd.AddIncomingMessage(topic1, size1, false)
	pd.AddOutgoingMessage(topic2, size2, false)

	str := pd.statsToString(1)

	assert.True(t, strings.Contains(str, topic1))
	assert.True(t, strings.Contains(str, core.ConvertBytes(size1)))

	assert.True(t, strings.Contains(str, topic2))
	assert.True(t, strings.Contains(str, core.ConvertBytes(size2)))
}
