package common

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWasContextClosed(t *testing.T) {
	t.Parallel()

	numTests := 10
	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < numTests; i++ {
		assert.False(t, WasContextClosed(ctx))
	}
	cancel()
	for i := 0; i < numTests; i++ {
		assert.True(t, WasContextClosed(ctx))
	}
}
