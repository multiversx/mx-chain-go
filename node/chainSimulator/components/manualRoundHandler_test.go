package components

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewManualRoundHandler(t *testing.T) {
	t.Parallel()

	handler := NewManualRoundHandler(100, time.Second, 0)
	require.NotNil(t, handler)
}

func TestManualRoundHandler_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var handler *manualRoundHandler
	require.True(t, handler.IsInterfaceNil())

	handler = NewManualRoundHandler(100, time.Second, 0)
	require.False(t, handler.IsInterfaceNil())
}

func TestManualRoundHandler_Operations(t *testing.T) {
	t.Parallel()

	genesisTime := time.Now()
	providedIndex := int64(0)
	providedRoundDuration := time.Second
	handler := NewManualRoundHandler(genesisTime.Unix(), providedRoundDuration, providedIndex)
	require.Equal(t, providedIndex, handler.Index())
	handler.IncrementIndex()
	require.Equal(t, providedIndex+1, handler.Index())
	expectedTimestamp := time.Unix(handler.genesisTimeStamp, 0).Add(providedRoundDuration)
	require.Equal(t, expectedTimestamp, handler.TimeStamp())
	require.Equal(t, providedRoundDuration, handler.TimeDuration())
	providedMaxTime := time.Minute
	require.Equal(t, providedMaxTime, handler.RemainingTime(time.Now(), providedMaxTime))
	require.False(t, handler.BeforeGenesis())
	handler.UpdateRound(time.Now(), time.Now()) // for coverage only
}
