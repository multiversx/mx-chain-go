package process

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHaveAdditionalTimeShouldUseSeventyMilliseconds(t *testing.T) {
	t.Parallel()

	require.Equal(t, 70*time.Millisecond, additionalTimeForCreatingScheduledMiniBlocks)
	require.True(t, HaveAdditionalTime()())
}
