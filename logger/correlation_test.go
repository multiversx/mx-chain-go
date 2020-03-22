package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCorrelation_Toggle(t *testing.T) {
	correlation := logCorrelation{}

	correlation.toggle(true)
	require.True(t, correlation.isEnabled())
	correlation.toggle(false)
	require.False(t, correlation.isEnabled())

	// Now with the global setter
	ToggleCorrelation(true)
	require.True(t, globalCorrelation.isEnabled())
	ToggleCorrelation(false)
	require.False(t, globalCorrelation.isEnabled())
}

func TestCorrelation_SettingElements(t *testing.T) {
	correlation := logCorrelation{}

	correlation.setShard("myshard")
	correlation.setEpoch(42)
	correlation.setRound(420)
	correlation.setSubRound("foo")

	require.Equal(t, "myshard", correlation.getShard())
	require.Equal(t, uint32(42), correlation.getEpoch())
	require.Equal(t, int64(420), correlation.getRound())
	require.Equal(t, "foo", correlation.getSubRound())

	// Now with the global setters
	SetCorrelationShard("meta")
	SetCorrelationEpoch(43)
	SetCorrelationRound(430)
	SetCorrelationSubround("bar")

	require.Equal(t, "meta", globalCorrelation.getShard())
	require.Equal(t, uint32(43), globalCorrelation.getEpoch())
	require.Equal(t, int64(430), globalCorrelation.getRound())
	require.Equal(t, "bar", globalCorrelation.getSubRound())
}
