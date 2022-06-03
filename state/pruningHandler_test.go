package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPruningHandler(t *testing.T) {
	t.Parallel()

	ph := NewPruningHandler(true)
	assert.True(t, ph.IsPruningEnabled())

	ph = NewPruningHandler(false)
	assert.False(t, ph.IsPruningEnabled())
}
