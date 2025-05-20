package trie

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/trie/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewGoroutinesManager(t *testing.T) {
	t.Parallel()

	t.Run("nil throttler", func(t *testing.T) {
		t.Parallel()

		manager, err := NewGoroutinesManager(nil, nil, nil, "")
		assert.Nil(t, manager)
		assert.Equal(t, ErrNilThrottler, err)
	})
	t.Run("nil error channel", func(t *testing.T) {
		t.Parallel()

		manager, err := NewGoroutinesManager(&mock.ThrottlerStub{}, nil, nil, "")
		assert.Nil(t, manager)
		assert.Equal(t, ErrNilBufferedErrChan, err)
	})
	t.Run("nil chan close", func(t *testing.T) {
		t.Parallel()

		manager, err := NewGoroutinesManager(&mock.ThrottlerStub{}, errChan.NewErrChanWrapper(), nil, "")
		assert.Nil(t, manager)
		assert.Equal(t, ErrNilChanClose, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		manager, err := NewGoroutinesManager(&mock.ThrottlerStub{}, errChan.NewErrChanWrapper(), make(chan struct{}), "")
		assert.NotNil(t, manager)
		assert.Nil(t, err)
		assert.True(t, manager.canProcess)
	})
}

func TestGoroutinesManager_ShouldContinueProcessing(t *testing.T) {
	t.Parallel()

	t.Run("should return false if chan is closed", func(t *testing.T) {
		t.Parallel()

		closeChan := make(chan struct{})
		manager, _ := NewGoroutinesManager(&mock.ThrottlerStub{}, errChan.NewErrChanWrapper(), closeChan, "")

		close(closeChan)
		assert.False(t, manager.ShouldContinueProcessing())
	})
	t.Run("should return true if canProcess", func(t *testing.T) {
		t.Parallel()

		closeChan := make(chan struct{})
		manager, _ := NewGoroutinesManager(&mock.ThrottlerStub{}, errChan.NewErrChanWrapper(), closeChan, "")

		assert.True(t, manager.ShouldContinueProcessing())
	})
	t.Run("should return false if !canProcess", func(t *testing.T) {
		t.Parallel()

		manager := &goroutinesManager{
			chanClose:  make(chan struct{}),
			canProcess: false,
		}

		assert.False(t, manager.ShouldContinueProcessing())
	})
}

func TestGoroutinesManager_CanStartGoRoutine(t *testing.T) {
	t.Parallel()

	t.Run("should return false if !throttler.CanProcess", func(t *testing.T) {
		t.Parallel()

		throttler := &mock.ThrottlerStub{
			CanProcessCalled: func() bool {
				return false
			},
		}
		manager, _ := NewGoroutinesManager(throttler, errChan.NewErrChanWrapper(), make(chan struct{}), "")

		assert.False(t, manager.CanStartGoRoutine())
	})
	t.Run("should return true if throttler.CanProcess", func(t *testing.T) {
		t.Parallel()

		startProcessingCalled := false
		throttler := &mock.ThrottlerStub{
			CanProcessCalled: func() bool {
				return true
			},
			StartProcessingCalled: func() {
				startProcessingCalled = true
			},
		}
		manager, _ := NewGoroutinesManager(throttler, errChan.NewErrChanWrapper(), make(chan struct{}), "")

		assert.True(t, manager.CanStartGoRoutine())
		assert.True(t, startProcessingCalled)
	})
}

func TestGoroutinesManager_EndGoRoutineProcessing(t *testing.T) {
	t.Parallel()

	endProcessingCalled := false
	throttler := &mock.ThrottlerStub{
		EndProcessingCalled: func() {
			endProcessingCalled = true
		},
	}
	manager, _ := NewGoroutinesManager(throttler, errChan.NewErrChanWrapper(), make(chan struct{}), "")
	manager.EndGoRoutineProcessing()
	assert.True(t, endProcessingCalled)
}

func TestGoroutinesManager_SetError(t *testing.T) {
	t.Parallel()

	t.Run("should set error", func(t *testing.T) {
		expectedErr := errors.New("error")
		errCh := errChan.NewErrChanWrapper()
		manager, _ := NewGoroutinesManager(&mock.ThrottlerStub{}, errCh, make(chan struct{}), "")

		err := errCh.ReadFromChanNonBlocking()
		assert.Nil(t, err)

		manager.SetError(expectedErr)

		err = errCh.ReadFromChanNonBlocking()
		assert.Equal(t, expectedErr, err)
		assert.False(t, manager.canProcess)
	})
	t.Run("set multiple errors should not block", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("error")
		anotherErr := errors.New("another error")
		errCh := errChan.NewErrChanWrapper()
		manager, _ := NewGoroutinesManager(&mock.ThrottlerStub{}, errCh, make(chan struct{}), "")

		err := errCh.ReadFromChanNonBlocking()
		assert.Nil(t, err)

		manager.SetError(expectedErr)
		manager.SetError(anotherErr)

		err = errCh.ReadFromChanNonBlocking()
		assert.Equal(t, expectedErr, err)
		assert.False(t, manager.canProcess)
	})
}

func TestGoroutinesManager_GetError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("error")
	errCh := errChan.NewErrChanWrapper()
	manager, _ := NewGoroutinesManager(&mock.ThrottlerStub{}, errCh, make(chan struct{}), "")

	err := manager.GetError()
	assert.Nil(t, err)

	manager.SetError(expectedErr)

	err = manager.GetError()
	assert.Equal(t, expectedErr, err)
}
