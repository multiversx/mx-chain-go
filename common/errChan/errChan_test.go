package errChan

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewErrChan(t *testing.T) {
	t.Parallel()

	ec := NewErrChan()
	assert.False(t, check.IfNil(ec))
	assert.Equal(t, 1, cap(ec.ch))
}

func TestErrChan_WriteInChanNonBlocking(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("err1")
	ec := NewErrChan()
	ec.WriteInChanNonBlocking(expectedErr)
	ec.WriteInChanNonBlocking(fmt.Errorf("err2"))
	ec.WriteInChanNonBlocking(fmt.Errorf("err3"))

	assert.Equal(t, 1, len(ec.ch))
	assert.Equal(t, expectedErr, <-ec.ch)
	assert.Equal(t, 0, len(ec.ch))
}

func TestErrChan_ReadFromChanNonBlocking(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("err1")
	ec := NewErrChan()
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

		ec := NewErrChan()
		ec.Close()

		assert.True(t, ec.closed)
		ec.Close()
	})

	t.Run("close a nil channel", func(t *testing.T) {
		t.Parallel()

		ec := NewErrChan()
		ec.ch = nil
		ec.Close()

		assert.False(t, ec.closed)
	})
}

func TestErrChan_Len(t *testing.T) {
	t.Parallel()

	ec := NewErrChan()
	assert.Equal(t, 0, ec.Len())

	ec.ch <- fmt.Errorf("err1")
	assert.Equal(t, 1, ec.Len())

	ec.WriteInChanNonBlocking(fmt.Errorf("err2"))
	assert.Equal(t, 1, ec.Len())
}
