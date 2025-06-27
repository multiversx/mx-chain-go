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

		actualRes := breadcrumb.isRelayer()
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

		actualRes := breadcrumb.isRelayer()
		require.False(t, actualRes)
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

func Test_breadCrumbIsContinuous(t *testing.T) {
	t.Parallel()

	// when breadcrumb is relayer
	t.Run("relayer should be continuous", func(t *testing.T) {
		t.Parallel()

		breadcrumb := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
				Value:    1,
				HasValue: false,
			},
			lastNonce: core.OptionalUint64{
				Value:    2,
				HasValue: false,
			},
			consumedBalance: nil,
		}

		skippedSenders := make(map[string]struct{})
		sendersInContinuityWithSessionNonce := make(map[string]struct{})
		accountPreviousBreadcrumb := make(map[string]*accountBreadcrumb)

		actualRes := breadcrumb.isContinuous("bob", 3,
			skippedSenders, sendersInContinuityWithSessionNonce, accountPreviousBreadcrumb)
		require.True(t, actualRes)
	})

	// when certain account is sender for the first time in the chain of tracked blocks
	t.Run("sender not continuous with session nonce", func(t *testing.T) {
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

		skippedSenders := make(map[string]struct{})
		sendersInContinuityWithSessionNonce := make(map[string]struct{})
		accountPreviousBreadcrumb := make(map[string]*accountBreadcrumb)

		actualRes := breadcrumbAlice.isContinuous("alice", 3,
			skippedSenders, sendersInContinuityWithSessionNonce, accountPreviousBreadcrumb)

		require.False(t, actualRes)
	})

	t.Run("sender continuous with session nonce", func(t *testing.T) {
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

		skippedSenders := make(map[string]struct{})
		sendersInContinuityWithSessionNonce := make(map[string]struct{})
		accountPreviousBreadcrumb := make(map[string]*accountBreadcrumb)

		actualRes := breadcrumbAlice.isContinuous("alice", 1,
			skippedSenders, sendersInContinuityWithSessionNonce, accountPreviousBreadcrumb)

		require.True(t, actualRes)
	})

	// when address was already a sender in the chain of tracked blocks
	t.Run("sender continuous with previous account breadcrumb ", func(t *testing.T) {
		breadcrumbAlice1 := accountBreadcrumb{
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

		breadcrumbAlice2 := accountBreadcrumb{
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

		skippedSenders := make(map[string]struct{})
		sendersInContinuityWithSessionNonce := map[string]struct{}{
			"alice": {},
		}
		accountPreviousBreadcrumb := map[string]*accountBreadcrumb{
			"alice": &breadcrumbAlice1,
		}

		actualRes := breadcrumbAlice2.isContinuous("alice", 1,
			skippedSenders, sendersInContinuityWithSessionNonce, accountPreviousBreadcrumb)

		require.True(t, actualRes)
	})
}
