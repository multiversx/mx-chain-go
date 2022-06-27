package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPruningHandler(t *testing.T) {
	t.Parallel()

	ph := NewPruningHandler(EnableDataRemoval)
	assert.True(t, ph.IsPruningEnabled())

	ph = NewPruningHandler(DisableDataRemoval)
	assert.False(t, ph.IsPruningEnabled())
}
