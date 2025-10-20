package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewTrackerInfo(t *testing.T) {
	t.Parallel()

	diagnosis := NewTrackerDiagnosis(0, 0)
	require.NotNil(t, diagnosis)
}

func Test_GetNumTrackedBlocks(t *testing.T) {
	t.Parallel()

	diagnosis := NewTrackerDiagnosis(16, 8)
	require.Equal(t, uint64(16), diagnosis.GetNumTrackedBlocks())
}

func Test_GetNumTrackedAccounts(t *testing.T) {
	t.Parallel()

	diagnosis := NewTrackerDiagnosis(16, 8)
	require.Equal(t, uint64(8), diagnosis.GetNumTrackedAccounts())
}
