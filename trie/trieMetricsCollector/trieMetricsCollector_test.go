package trieMetricsCollector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTrieMetricsCollector(t *testing.T) {
	t.Parallel()

	collector := NewTrieMetricsCollector()
	assert.NotNil(t, collector)
	assert.Equal(t, 0, collector.maxDepth)
	assert.Equal(t, 0, collector.sizeLoadedInMem)
}

func TestTrieMetricsCollector_SetDepth(t *testing.T) {
	t.Parallel()

	collector := NewTrieMetricsCollector()
	collector.SetDepth(5)

	assert.Equal(t, 5, collector.maxDepth)
	collector.SetDepth(3)
	assert.Equal(t, 5, collector.maxDepth) // Should not change since 3 < 5
	collector.SetDepth(7)
	assert.Equal(t, 7, collector.maxDepth) // Should update to 7 since it's greater than 5
}

func TestTrieMetricsCollector_GetMaxDepth(t *testing.T) {
	t.Parallel()

	collector := NewTrieMetricsCollector()
	assert.Equal(t, uint32(0), collector.GetMaxDepth())

	collector.SetDepth(10)
	assert.Equal(t, uint32(10), collector.GetMaxDepth())

	collector.SetDepth(5)
	assert.Equal(t, uint32(10), collector.GetMaxDepth()) // Should not change since 5 < 10
}

func TestTrieMetricsCollector_AddSizeLoadedInMem(t *testing.T) {
	t.Parallel()

	collector := NewTrieMetricsCollector()
	collector.AddSizeLoadedInMem(100)
	assert.Equal(t, 100, collector.sizeLoadedInMem)

	collector.AddSizeLoadedInMem(50)
	assert.Equal(t, 150, collector.sizeLoadedInMem) // Should accumulate size
}

func TestTrieMetricsCollector_GetSizeLoadedInMem(t *testing.T) {
	t.Parallel()

	collector := NewTrieMetricsCollector()
	assert.Equal(t, 0, collector.GetSizeLoadedInMem())

	collector.AddSizeLoadedInMem(200)
	assert.Equal(t, 200, collector.GetSizeLoadedInMem())

	collector.AddSizeLoadedInMem(300)
	assert.Equal(t, 500, collector.GetSizeLoadedInMem()) // Should accumulate size
}
