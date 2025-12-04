package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrackerInfo_NewTrackerInfo(t *testing.T) {
	t.Parallel()

	diagnosis := NewTrackerDiagnosis(0, 0)
	require.NotNil(t, diagnosis)
}

func TestTrackerInfo_GetNumTrackedBlocks(t *testing.T) {
	t.Parallel()

	diagnosis := NewTrackerDiagnosis(16, 8)
	require.Equal(t, uint64(16), diagnosis.GetNumTrackedBlocks())
}

func TestTrackerInfo_GetNumTrackedAccounts(t *testing.T) {
	t.Parallel()

	diagnosis := NewTrackerDiagnosis(16, 8)
	require.Equal(t, uint64(8), diagnosis.GetNumTrackedAccounts())
}

func TestTrackerInfo_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var td *trackerDiagnosis
	require.True(t, td.IsInterfaceNil())

	td = NewTrackerDiagnosis(0, 0)
	require.False(t, td.IsInterfaceNil())
}
