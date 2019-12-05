package antiflood_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood"
	"github.com/stretchr/testify/assert"
)

//------- NewCountersMap

func TestNewCountersMap_NilCacherShouldErr(t *testing.T) {
	t.Parallel()

	cm, err := antiflood.NewCountersMap(nil, antiflood.MinOperations)

	assert.True(t, check.IfNil(cm))
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestNewCountersMap_LowerMinOperationsShouldErr(t *testing.T) {
	t.Parallel()

	cm, err := antiflood.NewCountersMap(&mock.CacherStub{}, antiflood.MinOperations-1)

	assert.True(t, check.IfNil(cm))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewCountersMap_EqualMinOperationsShouldWork(t *testing.T) {
	t.Parallel()

	cm, err := antiflood.NewCountersMap(&mock.CacherStub{}, antiflood.MinOperations)

	assert.False(t, check.IfNil(cm))
	assert.Nil(t, err)
}

func TestNewCountersMap_HigherMinOperationsShouldWork(t *testing.T) {
	t.Parallel()

	cm, err := antiflood.NewCountersMap(&mock.CacherStub{}, antiflood.MinOperations+1)

	assert.False(t, check.IfNil(cm))
	assert.Nil(t, err)
}

//------- TryIncrement

func TestCountersMap_TryIncrementIdentifierNotPresentPutOneAndReturnTrue(t *testing.T) {
	t.Parallel()

	putWasCalled := false
	cm, _ := antiflood.NewCountersMap(
		&mock.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				valInt, isInt := value.(int)
				if isInt && valInt == 1 {
					putWasCalled = true
				}

				return false
			},
		},
		antiflood.MinOperations)

	ok := cm.TryIncrement("identifier")

	assert.True(t, ok)
	assert.True(t, putWasCalled)
}

func TestCountersMap_TryIncrementNotIntCounterShouldPutOneAndReturnTrue(t *testing.T) {
	t.Parallel()

	putWasCalled := false
	cm, _ := antiflood.NewCountersMap(
		&mock.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return "bad value", true
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				valInt, isInt := value.(int)
				if isInt && valInt == 1 {
					putWasCalled = true
				}

				return false
			},
		},
		antiflood.MinOperations)

	ok := cm.TryIncrement("identifier")

	assert.True(t, ok)
	assert.True(t, putWasCalled)
}

func TestCountersMap_TryIncrementUnderMaxValueShouldIncrementAndReturnTrue(t *testing.T) {
	t.Parallel()

	putWasCalled := false
	existingValue := antiflood.MinOperations
	cm, _ := antiflood.NewCountersMap(
		&mock.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return existingValue, true
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				valInt, isInt := value.(int)
				if isInt && valInt == existingValue+1 {
					putWasCalled = true
				}

				return false
			},
		},
		antiflood.MinOperations+10)

	ok := cm.TryIncrement("identifier")

	assert.True(t, ok)
	assert.True(t, putWasCalled)
}

func TestCountersMap_TryIncrementEqualMaxValueShouldNotPutAndReturnFalse(t *testing.T) {
	t.Parallel()

	existingValue := antiflood.MinOperations + 10
	cm, _ := antiflood.NewCountersMap(
		&mock.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return existingValue, true
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				assert.Fail(t, "should have not called put")

				return false
			},
		},
		antiflood.MinOperations+10)

	ok := cm.TryIncrement("identifier")

	assert.False(t, ok)
}

func TestCountersMap_TryIncrementOverMaxValueShouldNotPutAndReturnFalse(t *testing.T) {
	t.Parallel()

	existingValue := antiflood.MinOperations + 11
	cm, _ := antiflood.NewCountersMap(
		&mock.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return existingValue, true
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				assert.Fail(t, "should have not called put")

				return false
			},
		},
		antiflood.MinOperations+10)

	ok := cm.TryIncrement("identifier")

	assert.False(t, ok)
}

func TestCountersMap_TryIncrementShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	cm, _ := antiflood.NewCountersMap(
		mock.NewCacherMock(),
		antiflood.MinOperations)
	numIterations := 1000
	wg := sync.WaitGroup{}
	wg.Add(numIterations)
	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			ok := cm.TryIncrement(fmt.Sprintf("%d", idx))
			assert.True(t, ok)
			wg.Done()
		}(i)
	}

	wg.Wait()
}

//------- Reset

func TestCountersMap_ResetShouldCallCacherClear(t *testing.T) {
	t.Parallel()

	clearCalled := false
	cm, _ := antiflood.NewCountersMap(
		&mock.CacherStub{
			ClearCalled: func() {
				clearCalled = true
			},
		},
		antiflood.MinOperations)

	cm.Reset()

	assert.True(t, clearCalled)
}

func TestCountersMap_TryIncrementAndResetShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	cm, _ := antiflood.NewCountersMap(
		mock.NewCacherMock(),
		antiflood.MinOperations)
	numIterations := 1000
	wg := sync.WaitGroup{}
	wg.Add(numIterations + numIterations/10)
	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			ok := cm.TryIncrement(fmt.Sprintf("%d", idx))
			assert.True(t, ok)
			wg.Done()
		}(i)

		if i%10 == 0 {
			go func() {
				cm.Reset()
				wg.Done()
			}()
		}
	}

	wg.Wait()
}
