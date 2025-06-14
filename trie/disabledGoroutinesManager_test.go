package trie

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledGoroutinesManager(t *testing.T) {
	t.Parallel()

	d := NewDisabledGoroutinesManager()
	assert.False(t, check.IfNil(d))
}

func TestDisabledGoroutinesManager_ShouldContinueProcessing(t *testing.T) {
	t.Parallel()

	d := NewDisabledGoroutinesManager()
	assert.True(t, d.ShouldContinueProcessing())

	d.SetError(errors.New("error"))
	assert.False(t, d.ShouldContinueProcessing())
}

func TestDisabledGoroutinesManager_CanStartGoRoutine(t *testing.T) {
	t.Parallel()

	d := NewDisabledGoroutinesManager()
	assert.False(t, d.CanStartGoRoutine())
}

func TestDisabledGoroutinesManager_SetAndGetError(t *testing.T) {
	t.Parallel()

	d := NewDisabledGoroutinesManager()
	assert.Nil(t, d.GetError())

	err := errors.New("error")
	d.SetError(err)
	assert.Equal(t, err, d.GetError())
}
