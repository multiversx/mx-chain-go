package accumulator_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/accumulator"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var timeout = time.Second * 2

func TestNewTimeAccumulator_InvalidMaxWaitTimeShouldErr(t *testing.T) {
	t.Parallel()

	ta, err := accumulator.NewTimeAccumulator(accumulator.MinimumAlowedTime-1, 0)

	assert.True(t, check.IfNil(ta))
	assert.True(t, errors.Is(err, core.ErrInvalidValue))
}

func TestNewTimeAccumulator_InvalidMaxOffsetShouldErr(t *testing.T) {
	t.Parallel()

	ta, err := accumulator.NewTimeAccumulator(accumulator.MinimumAlowedTime, -1)

	assert.True(t, check.IfNil(ta))
	assert.True(t, errors.Is(err, core.ErrInvalidValue))
}

func TestNewTimeAccumulator_ShouldWork(t *testing.T) {
	t.Parallel()

	ta, err := accumulator.NewTimeAccumulator(accumulator.MinimumAlowedTime, 0)

	assert.False(t, check.IfNil(ta))
	assert.Nil(t, err)
}

//------- AddData

func TestTimeAccumulator_AddDataShouldWorkEvenIfTheChanIsBlocked(t *testing.T) {
	t.Parallel()

	chDone := make(chan struct{})
	allowedTime := time.Millisecond * 100
	ta, _ := accumulator.NewTimeAccumulator(allowedTime, 0)
	go func() {
		ta.AddData(struct{}{})
		time.Sleep(allowedTime * 3)
		ta.AddData(struct{}{})
		ta.AddData(struct{}{})

		existing := ta.Data()
		assert.Equal(t, 2, len(existing))

		chDone <- struct{}{}
	}()

	select {
	case <-chDone:
	case <-time.After(timeout):
		assert.Fail(t, "test did not finish in a reasonable time span. "+
			"Maybe problems with the used mutexes?")
	}
}

//------- eviction

func TestTimeAccumulator_EvictionShouldStopWhenCloseIsCalled(t *testing.T) {
	t.Parallel()

	allowedTime := time.Millisecond * 100
	ta, _ := accumulator.NewTimeAccumulator(allowedTime, 0)

	ta.AddData(struct{}{})
	time.Sleep(allowedTime * 3)

	_ = ta.Close()
	time.Sleep(allowedTime)

	ch := ta.OutputChannel()
	items, ok := <-ch

	assert.False(t, ok)
	assert.Equal(t, 0, len(items))
}

func TestTimeAccumulator_EvictionDuringWaitShouldStopWhenCloseIsCalled(t *testing.T) {
	t.Parallel()

	allowedTime := time.Millisecond * 100
	ta, _ := accumulator.NewTimeAccumulator(allowedTime, 0)
	ta.AddData(struct{}{})

	_ = ta.Close()
	time.Sleep(allowedTime)

	ch := ta.OutputChannel()
	items, ok := <-ch

	assert.False(t, ok)
	assert.Equal(t, 0, len(items))
}

func TestTimeAccumulator_EvictionShouldPreserveTheOrder(t *testing.T) {
	t.Parallel()

	allowedTime := time.Millisecond * 100
	ta, _ := accumulator.NewTimeAccumulator(allowedTime, 0)

	data := []interface{}{"data1", "data2", "data3"}
	for _, d := range data {
		ta.AddData(d)
	}
	time.Sleep(allowedTime * 3)

	ch := ta.OutputChannel()
	items, ok := <-ch

	require.True(t, ok)
	assert.Equal(t, data, items)
}

func TestTimeAccumulator_EvictionWithOffsetShouldPreserveTheOrder(t *testing.T) {
	t.Parallel()

	allowedTime := time.Millisecond * 100
	ta, _ := accumulator.NewTimeAccumulator(allowedTime, time.Millisecond)

	data := []interface{}{"data1", "data2", "data3"}
	for _, d := range data {
		ta.AddData(d)
	}
	time.Sleep(allowedTime * 3)

	ch := ta.OutputChannel()
	items, ok := <-ch

	require.True(t, ok)
	assert.Equal(t, data, items)
}

//------- computeWaitTime

func TestTimeAccumulator_ComputeWaitTimeWithMaxOffsetZeroShouldRetMaxWaitTime(t *testing.T) {
	t.Parallel()

	allowedTime := time.Millisecond * 56
	ta, _ := accumulator.NewTimeAccumulator(allowedTime, 0)

	assert.Equal(t, allowedTime, ta.ComputeWaitTime())
}

func TestTimeAccumulator_ComputeWaitTimeShouldWork(t *testing.T) {
	t.Parallel()

	allowedTime := time.Millisecond * 56
	maxOffset := time.Millisecond * 12
	ta, _ := accumulator.NewTimeAccumulator(allowedTime, maxOffset)

	numComputations := 10000
	for i := 0; i < numComputations; i++ {
		waitTime := ta.ComputeWaitTime()
		isInInterval := waitTime >= allowedTime-maxOffset &&
			waitTime <= allowedTime

		assert.True(t, isInInterval)
	}
}
