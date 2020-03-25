package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProfile_GetCurrentProfile(t *testing.T) {
	SetLogLevel("foobar:TRACE")
	ToggleCorrelation(true)
	ToggleLoggerName(true)

	profile := GetCurrentProfile()
	require.Equal(t, "foobar:TRACE", profile.LogLevelPatterns)
	require.True(t, profile.WithCorrelation)
	require.True(t, profile.WithLoggerName)
}

func TestProfile_MarshalUnmarshal(t *testing.T) {
	profile := Profile{
		LogLevelPatterns: "bar:INFO",
		WithCorrelation:  true,
		WithLoggerName:   false,
	}

	json, err := profile.Marshal()
	require.Nil(t, err)

	profile, err = UnmarshalProfile([]byte(json))
	require.Nil(t, err)

	require.Equal(t, "bar:INFO", profile.LogLevelPatterns)
	require.True(t, profile.WithCorrelation)
	require.False(t, profile.WithLoggerName)
}

func TestProfile_Apply(t *testing.T) {
	profile := Profile{
		LogLevelPatterns: "bar:INFO",
		WithCorrelation:  true,
		WithLoggerName:   false,
	}

	profile.Apply()

	require.Equal(t, "bar:INFO", GetLogLevelPattern())
	require.True(t, IsEnabledCorrelation())
	require.False(t, IsEnabledLoggerName())
}
