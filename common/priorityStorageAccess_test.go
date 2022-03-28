package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsStorageAccessValid(t *testing.T) {
	t.Parallel()

	t.Run("invalid storage access type", func(t *testing.T) {
		t.Parallel()

		assert.False(t, IsStorageAccessValid("invalid"))
	})
	t.Run("valid storage access type", func(t *testing.T) {
		t.Parallel()

		assert.True(t, IsStorageAccessValid(LowPriority))
		assert.True(t, IsStorageAccessValid(HighPriority))
		assert.True(t, IsStorageAccessValid(ProcessPriority))
		assert.True(t, IsStorageAccessValid(APIPriority))
		assert.True(t, IsStorageAccessValid(ResolveRequestPriority))
		assert.True(t, IsStorageAccessValid(SnapshotPriority))
		assert.True(t, IsStorageAccessValid(TestPriority))
	})
}
