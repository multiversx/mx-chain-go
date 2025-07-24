package txcache

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func Test_isRelayer(t *testing.T) {
	t.Parallel()

	t.Run("should return true", func(t *testing.T) {
		t.Parallel()

		breadcrumb := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    0,
				HasValue: false,
			},
			lastNonce: core.OptionalUint64{
				Value:    0,
				HasValue: false,
			},
			consumedBalance: nil,
		}

		actualRes := breadcrumb.hasUnkownNonce()
		require.True(t, actualRes)
	})

	t.Run("should return false", func(t *testing.T) {
		t.Parallel()

		breadcrumb := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    0,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    0,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		actualRes := breadcrumb.hasUnkownNonce()
		require.False(t, actualRes)
	})
}

func Test_updateLastNonce(t *testing.T) {
	t.Parallel()

	t.Run("should receivedLastNonceNotSet the received lastNonce does not have value", func(t *testing.T) {
		t.Parallel()

		breadcrumb := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    0,
				HasValue: false,
			},
			lastNonce: core.OptionalUint64{
				Value:    1,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		receivedLastNonce := core.OptionalUint64{
			Value:    3,
			HasValue: false,
		}

		err := breadcrumb.updateLastNonce(receivedLastNonce)
		require.Equal(t, errReceivedLastNonceNotSet, err)
		require.Equal(t, uint64(1), breadcrumb.lastNonce.Value)
	})

	t.Run("should return nonce gap", func(t *testing.T) {
		t.Parallel()

		breadcrumb := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    0,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    3,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		receivedLastNonce := core.OptionalUint64{
			Value:    5,
			HasValue: true,
		}

		err := breadcrumb.updateLastNonce(receivedLastNonce)
		require.Equal(t, errNonceGap, err)
	})

	t.Run("should return no err", func(t *testing.T) {
		t.Parallel()

		breadcrumb := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    0,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    3,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		receivedLastNonce := core.OptionalUint64{
			Value:    4,
			HasValue: true,
		}

		err := breadcrumb.updateLastNonce(receivedLastNonce)
		require.Nil(t, err)
		require.Equal(t, uint64(4), breadcrumb.lastNonce.Value)
	})
}

func Test_verifyContinuityBetweenAccountBreadcrumbs(t *testing.T) {
	t.Parallel()

	t.Run("should return true", func(t *testing.T) {
		t.Parallel()

		breadcrumbAlice := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    1,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    2,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		breadcrumbBob := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    3,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    4,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		actualRes := breadcrumbBob.verifyContinuityBetweenAccountBreadcrumbs(&breadcrumbAlice)
		require.True(t, actualRes)
	})

	t.Run("should return false", func(t *testing.T) {
		t.Parallel()

		breadcrumbAlice := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    1,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    2,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		breadcrumbBob := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    2,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    4,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		actualRes := breadcrumbBob.verifyContinuityBetweenAccountBreadcrumbs(&breadcrumbAlice)
		require.False(t, actualRes)
	})
}

func Test_verifyContinuityWithSessionNonce(t *testing.T) {
	t.Parallel()

	t.Run("should return true", func(t *testing.T) {
		t.Parallel()

		breadcrumb := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    1,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    2,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		actualRes := breadcrumb.verifyContinuityWithSessionNonce(1)
		require.True(t, actualRes)
	})

	t.Run("should return false", func(t *testing.T) {
		t.Parallel()

		breadcrumb := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    1,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    2,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		actualRes := breadcrumb.verifyContinuityWithSessionNonce(2)
		require.False(t, actualRes)
	})
}
