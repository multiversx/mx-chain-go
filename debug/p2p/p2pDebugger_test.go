package p2p

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

//------- AddIncomingMessage

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

//------- AddOutgoingMessage

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

//------- continuouslyPrintStatistics

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

func TestP2pDebugger_statsToString(t *testing.T) {
	t.Parallel()

	pd := newTestP2PDebugger(
		"",
		shouldCompute,
		mockPrintFn,
	)

	topic1 := "testTopic1"
	size1 := uint64(3 * 1048576) //3MB
	topic2 := "testTopic2"
	size2 := uint64(5 * 1024) //5kB
	pd.AddIncomingMessage(topic1, size1, false)
	pd.AddOutgoingMessage(topic2, size2, false)

	str := pd.statsToString(1)

	assert.True(t, strings.Contains(str, topic1))
	assert.True(t, strings.Contains(str, core.ConvertBytes(size1)))

	assert.True(t, strings.Contains(str, topic2))
	assert.True(t, strings.Contains(str, core.ConvertBytes(size2)))
}
