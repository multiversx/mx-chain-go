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

func BenchmarkWasContextClosed_WhenClosed(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for i := 0; i < b.N; i++ {
		_ = WasContextClosed(ctx)
	}
}

func BenchmarkWasContextClosed_WhenNotClosed(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < b.N; i++ {
		_ = WasContextClosed(ctx)
	}
}
