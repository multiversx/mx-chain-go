package p2p

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewP2PDebugger(t *testing.T) {
	t.Parallel()

	pd := NewP2PDebugger("")

	assert.False(t, check.IfNil(pd))
}

//------- AddIncomingMessage

func TestP2pDebugger_AddIncomingMessageShouldNotProcessWillNotAdd(t *testing.T) {
	t.Parallel()

	pd := NewP2PDebugger("")
	pd.shouldProcessDataFn = func() bool {
		return false
	}

	topic := "topic"
	size := uint64(3857)
	pd.AddIncomingMessage(topic, size, false)

	m := pd.data[topic]
	assert.Nil(t, m)
}

func TestP2pDebugger_AddIncomingMessage(t *testing.T) {
	t.Parallel()

	pd := NewP2PDebugger("")
	pd.shouldProcessDataFn = func() bool {
		return true
	}

	topic := "topic"
	size := uint64(3857)
	pd.AddIncomingMessage(topic, size, false)

	m := pd.data[topic]
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

	pd := NewP2PDebugger("")
	pd.shouldProcessDataFn = func() bool {
		return false
	}

	topic := "topic"
	size := uint64(3857)
	pd.AddOutgoingMessage(topic, size, false)

	m := pd.data[topic]
	assert.Nil(t, m)
}

func TestP2pDebugger_AddOutgoingMessage(t *testing.T) {
	t.Parallel()

	pd := NewP2PDebugger("")
	pd.shouldProcessDataFn = func() bool {
		return true
	}

	topic := "topic"
	size := uint64(3857)
	pd.AddOutgoingMessage(topic, size, false)

	m := pd.data[topic]
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

//------- doStats

func TestP2pDebugger_doStatsShouldNotPrint(t *testing.T) {
	t.Parallel()

	pd := NewP2PDebugger("")
	pd.shouldProcessDataFn = func() bool {
		return false
	}
	numPrintWasCalled := int32(0)
	pd.printStringFn = func(data string) {
		atomic.AddInt32(&numPrintWasCalled, 1)
	}

	time.Sleep(printTimeOneSecond * 3)

	assert.Equal(t, int32(0), atomic.LoadInt32(&numPrintWasCalled))
}

func TestP2pDebugger_doStatsShouldPrint(t *testing.T) {
	t.Parallel()

	pd := NewP2PDebugger("")
	pd.shouldProcessDataFn = func() bool {
		return true
	}
	numPrintWasCalled := int32(0)
	pd.printStringFn = func(data string) {
		atomic.AddInt32(&numPrintWasCalled, 1)
	}

	time.Sleep(printTimeOneSecond*3 + time.Millisecond*100)

	assert.Equal(t, int32(3), atomic.LoadInt32(&numPrintWasCalled))
}

func TestP2pDebugger_doStatsCloseShouldStop(t *testing.T) {
	t.Parallel()

	pd := NewP2PDebugger("")
	pd.shouldProcessDataFn = func() bool {
		return true
	}
	numPrintWasCalled := int32(0)
	pd.printStringFn = func(data string) {
		atomic.AddInt32(&numPrintWasCalled, 1)
	}

	time.Sleep(printTimeOneSecond*3 + time.Millisecond*100)
	assert.Equal(t, int32(3), atomic.LoadInt32(&numPrintWasCalled))

	err := pd.Close()
	assert.Nil(t, err)

	time.Sleep(printTimeOneSecond*3 + time.Millisecond*100)
	assert.Equal(t, int32(3), atomic.LoadInt32(&numPrintWasCalled))
}

func TestP2pDebugger_doStatsString(t *testing.T) {
	t.Parallel()

	pd := NewP2PDebugger("")
	pd.shouldProcessDataFn = func() bool {
		return true
	}

	topic1 := "testTopic1"
	size1 := uint64(3 * 1048576) //3MB
	topic2 := "testTopic2"
	size2 := uint64(5 * 1024) //5kB
	pd.AddIncomingMessage(topic1, size1, false)
	pd.AddOutgoingMessage(topic2, size2, false)

	str := pd.doStatsString()

	assert.True(t, strings.Contains(str, topic1))
	assert.True(t, strings.Contains(str, core.ConvertBytes(size1)))

	assert.True(t, strings.Contains(str, topic2))
	assert.True(t, strings.Contains(str, core.ConvertBytes(size2)))
}
