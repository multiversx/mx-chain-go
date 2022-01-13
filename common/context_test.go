package common

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// nolint
func TestIsContextDone(t *testing.T) {
	t.Parallel()

	t.Run("valid context tests", func(t *testing.T) {
		numTests := 10
		ctx, cancel := context.WithCancel(context.Background())
		for i := 0; i < numTests; i++ {
			assert.False(t, IsContextDone(ctx))
		}
		cancel()
		for i := 0; i < numTests; i++ {
			assert.True(t, IsContextDone(ctx))
		}
	})
	t.Run("nil context", func(t *testing.T) {
		assert.True(t, IsContextDone(nil))
	})
}

func BenchmarkIsContextDone_WhenDone(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for i := 0; i < b.N; i++ {
		_ = IsContextDone(ctx)
	}
}

func BenchmarkIsContextDone_WhenNotDone(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < b.N; i++ {
		_ = IsContextDone(ctx)
	}
}
