package trieMetricsCollector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDisabledTrieMetricsCollector(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, NewDisabledTrieMetricsCollector())
}

func TestDisabledTrieMetricsCollector_SetMaxDepthDoesNotPanic(t *testing.T) {
	t.Parallel()

	collector := NewDisabledTrieMetricsCollector()
	collector.SetMaxDepth(5)

	// No assertion needed, just checking that it doesn't panic
}

func TestDisabledTrieMetricsCollector_GetMaxDepthReturnsZero(t *testing.T) {
	t.Parallel()

	collector := NewDisabledTrieMetricsCollector()
	maxDepth := collector.GetMaxDepth()

	assert.Equal(t, uint32(0), maxDepth)
}

func TestDisabledTrieMetricsCollector_AddSizeLoadedInMemDoesNotPanic(t *testing.T) {
	t.Parallel()

	collector := NewDisabledTrieMetricsCollector()
	collector.AddSizeLoadedInMem(100)

	// No assertion needed, just checking that it doesn't panic
}

func TestDisabledTrieMetricsCollector_GetSizeLoadedInMemReturnsZero(t *testing.T) {
	t.Parallel()

	collector := NewDisabledTrieMetricsCollector()
	size := collector.GetSizeLoadedInMem()

	assert.Equal(t, 0, size)
}
