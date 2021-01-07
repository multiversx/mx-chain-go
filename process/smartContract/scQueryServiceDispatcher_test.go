package smartContract

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewScQueryServiceDispatcher_NilEmptyListShouldErr(t *testing.T) {
	t.Parallel()

	sqsd, err := NewScQueryServiceDispatcher(nil)
	assert.True(t, check.IfNil(sqsd))
	assert.True(t, errors.Is(err, process.ErrNilOrEmptyList))

	sqsd, err = NewScQueryServiceDispatcher(make([]process.SCQueryService, 0))
	assert.True(t, check.IfNil(sqsd))
	assert.True(t, errors.Is(err, process.ErrNilOrEmptyList))
}

func TestNewScQueryServiceDispatcher_OneElementIsNilShouldErr(t *testing.T) {
	t.Parallel()

	sqsd, err := NewScQueryServiceDispatcher([]process.SCQueryService{
		&mock.ScQueryStub{},
		nil,
		&mock.ScQueryStub{},
	})
	assert.True(t, check.IfNil(sqsd))
	assert.True(t, errors.Is(err, process.ErrNilScQueryElement))
}

func TestNewScQueryServiceDispatcher_ShouldWork(t *testing.T) {
	t.Parallel()

	sqsd, err := NewScQueryServiceDispatcher([]process.SCQueryService{
		&mock.ScQueryStub{},
		&mock.ScQueryStub{},
	})
	assert.False(t, check.IfNil(sqsd))
	assert.Nil(t, err)
	assert.Equal(t, 2, len(sqsd.list))
}

func TestScQueryServiceDispatcher_ExecuteQueryShouldCallInRoundRobinFashion(t *testing.T) {
	t.Parallel()

	calledElement1 := 0
	calledElement2 := 0
	sqsd, _ := NewScQueryServiceDispatcher([]process.SCQueryService{
		&mock.ScQueryStub{
			ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, error) {
				calledElement1++

				return nil, nil
			},
		},
		&mock.ScQueryStub{
			ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, error) {
				calledElement2++

				return nil, nil
			},
		},
	})

	_, _ = sqsd.ExecuteQuery(nil)
	_, _ = sqsd.ExecuteQuery(nil)
	_, _ = sqsd.ExecuteQuery(nil)

	assert.Equal(t, 2, calledElement1)
	assert.Equal(t, 1, calledElement2)
}

func TestScQueryServiceDispatcher_ComputeScCallGasLimitShouldCallInRoundRobinFashion(t *testing.T) {
	t.Parallel()

	calledElement1 := 0
	calledElement2 := 0
	sqsd, _ := NewScQueryServiceDispatcher([]process.SCQueryService{
		&mock.ScQueryStub{
			ComputeScCallGasLimitHandler: func(tx *transaction.Transaction) (uint64, error) {
				calledElement1++

				return 0, nil
			},
		},
		&mock.ScQueryStub{
			ComputeScCallGasLimitHandler: func(tx *transaction.Transaction) (uint64, error) {
				calledElement2++

				return 0, nil
			},
		},
	})

	_, _ = sqsd.ComputeScCallGasLimit(nil)
	_, _ = sqsd.ComputeScCallGasLimit(nil)
	_, _ = sqsd.ComputeScCallGasLimit(nil)

	assert.Equal(t, 2, calledElement1)
	assert.Equal(t, 1, calledElement2)
}

func TestScQueryServiceDispatcher_ShouldWorkInAConcurrentManner(t *testing.T) {
	t.Parallel()

	calledElement1 := uint32(0)
	calledElement2 := uint32(0)
	sqsd, _ := NewScQueryServiceDispatcher([]process.SCQueryService{
		&mock.ScQueryStub{
			ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, error) {
				atomic.AddUint32(&calledElement1, 1)

				return nil, nil
			},
			ComputeScCallGasLimitHandler: func(tx *transaction.Transaction) (uint64, error) {
				atomic.AddUint32(&calledElement1, 1)

				return 0, nil
			},
		},
		&mock.ScQueryStub{
			ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, error) {
				atomic.AddUint32(&calledElement2, 1)

				return nil, nil
			},
			ComputeScCallGasLimitHandler: func(tx *transaction.Transaction) (uint64, error) {
				atomic.AddUint32(&calledElement2, 1)

				return 0, nil
			},
		},
	})

	numCalls := 100
	wg := &sync.WaitGroup{}
	wg.Add(numCalls * 2)
	for i := 0; i < numCalls; i++ {
		go func() {
			_, _ = sqsd.ExecuteQuery(nil)
			wg.Done()
		}()
		go func() {
			_, _ = sqsd.ComputeScCallGasLimit(nil)
			wg.Done()
		}()
	}

	wg.Wait()

	assert.Equal(t, uint32(numCalls), atomic.LoadUint32(&calledElement1))
	assert.Equal(t, uint32(numCalls), atomic.LoadUint32(&calledElement2))
}
