package errChan

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewErrChan(t *testing.T) {
	t.Parallel()

	ec := NewErrChanWrapper()
	assert.False(t, check.IfNil(ec))
	assert.Equal(t, 1, cap(ec.ch))
}

func TestErrChan_WriteInChanNonBlocking(t *testing.T) {
	t.Parallel()

	t.Run("write in a nil channel", func(t *testing.T) {
		t.Parallel()

		ec := NewErrChanWrapper()
		ec.ch = nil
		ec.WriteInChanNonBlocking(fmt.Errorf("err1"))

		assert.Equal(t, 0, len(ec.ch))
	})

	t.Run("write in a closed channel", func(t *testing.T) {
		t.Parallel()

		ec := NewErrChanWrapper()
		ec.Close()
		ec.WriteInChanNonBlocking(fmt.Errorf("err1"))

		assert.Equal(t, 0, len(ec.ch))
	})

	t.Run("should work", func(t *testing.T) {
		expectedErr := fmt.Errorf("err1")
		ec := NewErrChanWrapper()
		ec.WriteInChanNonBlocking(expectedErr)
		ec.WriteInChanNonBlocking(fmt.Errorf("err2"))
		ec.WriteInChanNonBlocking(fmt.Errorf("err3"))

		assert.Equal(t, 1, len(ec.ch))
		assert.Equal(t, expectedErr, <-ec.ch)
		assert.Equal(t, 0, len(ec.ch))
	})
}

func TestErrChan_ReadFromChanNonBlocking(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("err1")
	ec := NewErrChanWrapper()
	ec.ch <- expectedErr

	assert.Equal(t, 1, len(ec.ch))
	assert.Equal(t, expectedErr, ec.ReadFromChanNonBlocking())
	assert.Equal(t, 0, len(ec.ch))
	assert.Nil(t, ec.ReadFromChanNonBlocking())
}

func TestErrChan_Close(t *testing.T) {
	t.Parallel()

	t.Run("close an already closed channel", func(t *testing.T) {
		t.Parallel()

		ec := NewErrChanWrapper()
		ec.Close()

		assert.True(t, ec.closed)
		ec.Close()
	})

	t.Run("close a nil channel", func(t *testing.T) {
		t.Parallel()

		ec := NewErrChanWrapper()
		ec.ch = nil
		ec.Close()

		assert.False(t, ec.closed)
	})
}

func TestErrChan_Len(t *testing.T) {
	t.Parallel()

	ec := NewErrChanWrapper()
	assert.Equal(t, 0, ec.Len())

	ec.ch <- fmt.Errorf("err1")
	assert.Equal(t, 1, ec.Len())

	ec.WriteInChanNonBlocking(fmt.Errorf("err2"))
	assert.Equal(t, 1, ec.Len())
}

func TestErrChan_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	ec := NewErrChanWrapper()
	numOperations := 1000
	numMethods := 2
	wg := sync.WaitGroup{}
	wg.Add(numOperations)
	for i := 0; i < numOperations; i++ {
		go func(idx int) {

			if idx == numOperations-100 {
				ec.Close()
			}

			operation := idx % numMethods
			switch operation {
			case 0:
				ec.WriteInChanNonBlocking(fmt.Errorf("err"))
			case 1:
				_ = ec.ReadFromChanNonBlocking()
			default:
				assert.Fail(t, "invalid numMethods")
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
}
