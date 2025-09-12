package txcache

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func Test_continuousBreadcrumbs(t *testing.T) {
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

		validator := newBreadcrumbValidator()

		actualRes := validator.isContinuousBreadcrumb("bob", 0, &breadcrumb)
		require.True(t, actualRes)
	})

	// when certain account is sender for the first time in the chain of tracked blocks
	t.Run("sender not continuous with session nonce", func(t *testing.T) {
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

		validator := newBreadcrumbValidator()
		actualRes := validator.continuousWithSessionNonce("alice", 3, &breadcrumbAlice)
		require.False(t, actualRes)
	})

	t.Run("sender continuous with session nonce", func(t *testing.T) {
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

		validator := newBreadcrumbValidator()

		actualRes := validator.isContinuousBreadcrumb("alice", 1, &breadcrumbAlice)
		require.True(t, actualRes)

		_, ok := validator.sendersInContinuityWithSessionNonce["alice"]
		require.True(t, ok)

		actualBreadcrumb, ok := validator.accountPreviousBreadcrumb["alice"]
		require.True(t, ok)
		require.Equal(t, &breadcrumbAlice, actualBreadcrumb)
	})

	// when address was already a sender in the chain of tracked blocks
	t.Run("sender continuous with previous account breadcrumb ", func(t *testing.T) {
		t.Parallel()

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

		validator := newBreadcrumbValidator()

		validator.accountPreviousBreadcrumb = map[string]*accountBreadcrumb{
			"alice": &breadcrumbAlice1,
		}

		actualRes := validator.continuousWithPreviousBreadcrumb("alice", 3, &breadcrumbAlice2)
		require.True(t, actualRes)

		actualBreadcrumb, ok := validator.accountPreviousBreadcrumb["alice"]
		require.True(t, ok)
		require.Equal(t, &breadcrumbAlice2, actualBreadcrumb)
	})

	t.Run("sender is not continuous with previous account breadcrumb ", func(t *testing.T) {
		t.Parallel()

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
				Value:    4,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    4,
				HasValue: true,
			},
			consumedBalance: nil,
		}

		validator := newBreadcrumbValidator()

		validator.accountPreviousBreadcrumb = map[string]*accountBreadcrumb{
			"alice": &breadcrumbAlice1,
		}
		validator.sendersInContinuityWithSessionNonce = map[string]struct{}{
			"alice": {},
		}

		actualRes := validator.continuousWithPreviousBreadcrumb("alice", 1, &breadcrumbAlice2)
		require.False(t, actualRes)
	})
}
