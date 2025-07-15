package main

import (
	"testing"

	"github.com/klauspost/cpuid/v2"
	"github.com/stretchr/testify/require"
)

func TestParseFeatures(t *testing.T) {
	t.Parallel()

	t.Run("unknown feature id should fail", func(t *testing.T) {
		t.Parallel()

		features := []string{
			cpuid.FeatureID(cpuid.UNKNOWN).String(),
		}

		retFeatures, err := parseFeatures(features)
		require.Nil(t, retFeatures)
		require.Error(t, err)
	})

	t.Run("non existing flag should fail", func(t *testing.T) {
		t.Parallel()

		features := []string{
			"non existing feature id",
		}

		retFeatures, err := parseFeatures(features)
		require.Nil(t, retFeatures)
		require.Error(t, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		features := []string{
			cpuid.FeatureID(cpuid.SSE).String(),
			cpuid.FeatureID(cpuid.SSE2).String(),
		}
		retFeatures, err := parseFeatures(features)
		require.Nil(t, err)

		expFeatures := []cpuid.FeatureID{
			cpuid.SSE,
			cpuid.SSE2,
		}

		require.Equal(t, expFeatures, retFeatures)
	})
}
