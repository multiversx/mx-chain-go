package txcache

import (
	"math"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func Test_newGlobalAccountBreadcrumb(t *testing.T) {
	t.Parallel()

	gab := newGlobalAccountBreadcrumb()
	require.NotNil(t, gab)
	require.Equal(t, big.NewInt(0), gab.consumedBalance)

	require.Equal(t, uint64(0), gab.lastNonce.Value)
	require.False(t, gab.lastNonce.HasValue)

	require.Equal(t, uint64(math.MaxUint64), gab.firstNonce.Value)
	require.False(t, gab.firstNonce.HasValue)
}

func Test_updateOnAddedAccountBreadcrumb(t *testing.T) {
	t.Parallel()

	t.Run("should work when each breadcrumb is a sender breadcrumb", func(t *testing.T) {
		t.Parallel()

		accountBreadcrumbsForAlice := []*accountBreadcrumb{
			{
				firstNonce: core.OptionalUint64{
					Value:    10,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    16,
					HasValue: true,
				},
				consumedBalance: big.NewInt(10),
			},
			{
				firstNonce: core.OptionalUint64{
					Value:    17,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    23,
					HasValue: true,
				},
				consumedBalance: big.NewInt(5),
			},
			{
				firstNonce: core.OptionalUint64{
					Value:    24,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    31,
					HasValue: true,
				},
				consumedBalance: big.NewInt(5),
			},
		}

		gab := newGlobalAccountBreadcrumb()
		for _, breadcrumb := range accountBreadcrumbsForAlice {
			gab.updateOnAddedAccountBreadcrumb(breadcrumb)
		}

		require.Equal(t, big.NewInt(20), gab.consumedBalance)

		require.True(t, gab.firstNonce.HasValue)
		require.Equal(t, uint64(10), gab.firstNonce.Value)

		require.True(t, gab.lastNonce.HasValue)
		require.Equal(t, uint64(31), gab.lastNonce.Value)
	})

	t.Run("should work when some breadcrumbs are fee payer breadcrumbs", func(t *testing.T) {
		t.Parallel()

		accountBreadcrumbsForAlice := []*accountBreadcrumb{
			{
				firstNonce: core.OptionalUint64{
					Value:    0,
					HasValue: false,
				},
				lastNonce: core.OptionalUint64{
					Value:    0,
					HasValue: false,
				},
				consumedBalance: big.NewInt(10),
			},
			{
				firstNonce: core.OptionalUint64{
					Value:    0,
					HasValue: false,
				},
				lastNonce: core.OptionalUint64{
					Value:    0,
					HasValue: false,
				},
				consumedBalance: big.NewInt(10),
			},
			{
				firstNonce: core.OptionalUint64{
					Value:    17,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    23,
					HasValue: true,
				},
				consumedBalance: big.NewInt(5),
			},
			{
				firstNonce: core.OptionalUint64{
					Value:    0,
					HasValue: true,
				},
				lastNonce: core.OptionalUint64{
					Value:    0,
					HasValue: true,
				},
				consumedBalance: big.NewInt(5),
			},
		}

		gab := newGlobalAccountBreadcrumb()
		for _, breadcrumb := range accountBreadcrumbsForAlice {
			gab.updateOnAddedAccountBreadcrumb(breadcrumb)
		}

		require.Equal(t, big.NewInt(30), gab.consumedBalance)

		require.True(t, gab.firstNonce.HasValue)
		require.Equal(t, uint64(17), gab.firstNonce.Value)

		require.True(t, gab.lastNonce.HasValue)
		require.Equal(t, uint64(23), gab.lastNonce.Value)
	})
}

func Test_updateOnRemoveAccountBreadcrumbOnExecutedBlock(t *testing.T) {
	t.Parallel()

	t.Run("should work when each breadcrumb is a sender breadcrumb", func(t *testing.T) {
		t.Parallel()

		breadcrumb1 := &accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    10,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    16,
				HasValue: true,
			},
			consumedBalance: big.NewInt(10),
		}
		breadcrumb2 := &accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    17,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    23,
				HasValue: true,
			},
			consumedBalance: big.NewInt(5),
		}
		breadcrumb3 := &accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    24,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    31,
				HasValue: true,
			},
			consumedBalance: big.NewInt(5),
		}

		accountBreadcrumbsForAlice := []*accountBreadcrumb{
			breadcrumb1,
			breadcrumb2,
			breadcrumb3,
		}

		gab := newGlobalAccountBreadcrumb()
		for _, breadcrumb := range accountBreadcrumbsForAlice {
			gab.updateOnAddedAccountBreadcrumb(breadcrumb)
		}

		require.Equal(t, big.NewInt(20), gab.consumedBalance)

		require.True(t, gab.firstNonce.HasValue)
		require.Equal(t, uint64(10), gab.firstNonce.Value)

		require.True(t, gab.lastNonce.HasValue)
		require.Equal(t, uint64(31), gab.lastNonce.Value)

		shouldBeDeleted, err := gab.updateOnRemoveAccountBreadcrumbOnExecutedBlock(breadcrumb1)
		require.NoError(t, err)
		require.False(t, shouldBeDeleted)

		require.Equal(t, big.NewInt(10), gab.consumedBalance)
		require.True(t, gab.firstNonce.HasValue)
		require.Equal(t, uint64(17), gab.firstNonce.Value)

		shouldBeDeleted, err = gab.updateOnRemoveAccountBreadcrumbOnExecutedBlock(breadcrumb3)
		require.NoError(t, err)
		require.False(t, shouldBeDeleted)

		require.Equal(t, big.NewInt(5), gab.consumedBalance)
		require.False(t, gab.firstNonce.HasValue)
		require.Equal(t, uint64(math.MaxUint64), gab.firstNonce.Value)
		require.Equal(t, uint64(0), gab.lastNonce.Value)

		shouldBeDeleted, err = gab.updateOnRemoveAccountBreadcrumbOnExecutedBlock(breadcrumb2)
		require.NoError(t, err)
		require.True(t, shouldBeDeleted)
		require.Equal(t, uint64(math.MaxUint64), gab.firstNonce.Value)
		require.Equal(t, uint64(0), gab.lastNonce.Value)
	})

	t.Run("should work when there are some fee payer breadcrumbs", func(t *testing.T) {
		t.Parallel()

		breadcrumb0 := &accountBreadcrumb{
			consumedBalance: big.NewInt(10),
		}
		breadcrumb1 := &accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    10,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    16,
				HasValue: true,
			},
			consumedBalance: big.NewInt(10),
		}
		breadcrumb2 := &accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    17,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    23,
				HasValue: true,
			},
			consumedBalance: big.NewInt(5),
		}
		breadcrumb3 := &accountBreadcrumb{
			consumedBalance: big.NewInt(10),
		}
		breadcrumb4 := &accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    24,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    31,
				HasValue: true,
			},
			consumedBalance: big.NewInt(5),
		}

		accountBreadcrumbsForAlice := []*accountBreadcrumb{
			breadcrumb0,
			breadcrumb1,
			breadcrumb2,
			breadcrumb3,
			breadcrumb4,
		}

		gab := newGlobalAccountBreadcrumb()
		for _, breadcrumb := range accountBreadcrumbsForAlice {
			gab.updateOnAddedAccountBreadcrumb(breadcrumb)
		}

		require.Equal(t, big.NewInt(40), gab.consumedBalance)

		require.True(t, gab.firstNonce.HasValue)
		require.Equal(t, uint64(10), gab.firstNonce.Value)

		require.True(t, gab.lastNonce.HasValue)
		require.Equal(t, uint64(31), gab.lastNonce.Value)

		shouldBeDeleted, err := gab.updateOnRemoveAccountBreadcrumbOnExecutedBlock(breadcrumb0)
		require.NoError(t, err)
		require.False(t, shouldBeDeleted)

		require.Equal(t, big.NewInt(30), gab.consumedBalance)
		require.True(t, gab.firstNonce.HasValue)
		require.Equal(t, uint64(10), gab.firstNonce.Value)
		require.Equal(t, uint64(31), gab.lastNonce.Value)

		shouldBeDeleted, err = gab.updateOnRemoveAccountBreadcrumbOnExecutedBlock(breadcrumb1)
		require.NoError(t, err)
		require.False(t, shouldBeDeleted)

		require.Equal(t, big.NewInt(20), gab.consumedBalance)
		require.True(t, gab.firstNonce.HasValue)
		require.Equal(t, uint64(17), gab.firstNonce.Value)

		shouldBeDeleted, err = gab.updateOnRemoveAccountBreadcrumbOnExecutedBlock(breadcrumb4)
		require.NoError(t, err)
		require.False(t, shouldBeDeleted)

		require.Equal(t, big.NewInt(15), gab.consumedBalance)
		require.False(t, gab.firstNonce.HasValue)
		require.Equal(t, uint64(math.MaxUint64), gab.firstNonce.Value)
		require.Equal(t, uint64(0), gab.lastNonce.Value)

		shouldBeDeleted, err = gab.updateOnRemoveAccountBreadcrumbOnExecutedBlock(breadcrumb2)
		require.NoError(t, err)
		require.False(t, shouldBeDeleted)

		require.Equal(t, uint64(math.MaxUint64), gab.firstNonce.Value)
		require.Equal(t, uint64(0), gab.lastNonce.Value)

		shouldBeDeleted, err = gab.updateOnRemoveAccountBreadcrumbOnExecutedBlock(breadcrumb3)
		require.NoError(t, err)
		require.True(t, shouldBeDeleted)
	})
}

func Test_updateOnRemoveAccountBreadcrumbOnProposedBlock(t *testing.T) {
	t.Parallel()

	// Currently, this scenario does not exist in the SelectionTracker.
	// At the moment, on the OnProposedBlock flow, we only replace the block with the same nonce, without deleting the ones with greater nonce.
	// However, this will be changed in a future PR.
	t.Run("should work in the following order: replacing and then deleting the ones with greater nonce", func(t *testing.T) {
		t.Parallel()

		breadcrumb0 := &accountBreadcrumb{
			consumedBalance: big.NewInt(10),
		}
		breadcrumb1 := &accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    10,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    16,
				HasValue: true,
			},
			consumedBalance: big.NewInt(10),
		}
		breadcrumb2 := &accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    17,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    23,
				HasValue: true,
			},
			consumedBalance: big.NewInt(5),
		}
		breadcrumb3 := &accountBreadcrumb{
			consumedBalance: big.NewInt(10),
		}
		breadcrumb4 := &accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    24,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    31,
				HasValue: true,
			},
			consumedBalance: big.NewInt(5),
		}
		breadcrumb5 := &accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    32,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    40,
				HasValue: true,
			},
			consumedBalance: big.NewInt(10),
		}

		accountBreadcrumbsForAlice := []*accountBreadcrumb{
			breadcrumb0,
			breadcrumb1,
			breadcrumb2,
			breadcrumb3,
			breadcrumb4,
			breadcrumb5,
		}

		gab := newGlobalAccountBreadcrumb()
		for _, breadcrumb := range accountBreadcrumbsForAlice {
			gab.updateOnAddedAccountBreadcrumb(breadcrumb)
		}

		require.Equal(t, big.NewInt(50), gab.consumedBalance)
		require.Equal(t, uint64(10), gab.firstNonce.Value)
		require.Equal(t, uint64(40), gab.lastNonce.Value)

		// now, remove breadcrumb corresponding to the block with nonce 4
		shouldDelete, err := gab.updateOnRemoveAccountBreadcrumbOnProposedBlock(breadcrumb4)
		require.NoError(t, err)
		require.False(t, shouldDelete)

		require.Equal(t, big.NewInt(45), gab.consumedBalance)
		require.True(t, gab.firstNonce.HasValue)
		require.Equal(t, uint64(10), gab.firstNonce.Value)
		require.True(t, gab.lastNonce.HasValue)
		require.Equal(t, uint64(23), gab.lastNonce.Value)

		// now, remove the ones with greater nonce
		shouldDelete, err = gab.updateOnRemoveAccountBreadcrumbOnProposedBlock(breadcrumb5)
		require.NoError(t, err)
		require.False(t, shouldDelete)

		require.Equal(t, big.NewInt(35), gab.consumedBalance)
		require.True(t, gab.firstNonce.HasValue)
		require.Equal(t, uint64(10), gab.firstNonce.Value)
		require.True(t, gab.lastNonce.HasValue)
		require.Equal(t, uint64(23), gab.lastNonce.Value)

		// now, we create a new breadcrumb which will replace the one with nonce 4 (breadcrumb4)
		breadcrumbToReplaceNonce4 := &accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    24,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    31,
				HasValue: true,
			},
			consumedBalance: big.NewInt(5),
		}
		gab.updateOnAddedAccountBreadcrumb(breadcrumbToReplaceNonce4)
		require.Equal(t, big.NewInt(40), gab.consumedBalance)
		require.True(t, gab.firstNonce.HasValue)
		require.Equal(t, uint64(10), gab.firstNonce.Value)
		require.True(t, gab.lastNonce.HasValue)
		require.Equal(t, uint64(31), gab.lastNonce.Value)
	})

	t.Run("should work without specific order", func(t *testing.T) {
		t.Parallel()

		breadcrumb0 := &accountBreadcrumb{
			consumedBalance: big.NewInt(10),
		}
		breadcrumb1 := &accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    10,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    16,
				HasValue: true,
			},
			consumedBalance: big.NewInt(10),
		}
		breadcrumb2 := &accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    17,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    23,
				HasValue: true,
			},
			consumedBalance: big.NewInt(5),
		}
		breadcrumb3 := &accountBreadcrumb{
			consumedBalance: big.NewInt(10),
		}
		breadcrumb4 := &accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    24,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    31,
				HasValue: true,
			},
			consumedBalance: big.NewInt(5),
		}
		breadcrumb5 := &accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    32,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    40,
				HasValue: true,
			},
			consumedBalance: big.NewInt(10),
		}

		accountBreadcrumbsForAlice := []*accountBreadcrumb{
			breadcrumb0,
			breadcrumb1,
			breadcrumb2,
			breadcrumb3,
			breadcrumb4,
			breadcrumb5,
		}

		gab := newGlobalAccountBreadcrumb()
		for _, breadcrumb := range accountBreadcrumbsForAlice {
			gab.updateOnAddedAccountBreadcrumb(breadcrumb)
		}

		require.Equal(t, big.NewInt(50), gab.consumedBalance)
		require.Equal(t, uint64(10), gab.firstNonce.Value)
		require.Equal(t, uint64(40), gab.lastNonce.Value)

		// in this scenario, we want to replace breadcrumb1

		// remove breadcrumb with greater nonce
		shouldDelete, err := gab.updateOnRemoveAccountBreadcrumbOnProposedBlock(breadcrumb4)
		require.NoError(t, err)
		require.False(t, shouldDelete)

		require.Equal(t, big.NewInt(45), gab.consumedBalance)
		require.True(t, gab.firstNonce.HasValue)
		require.Equal(t, uint64(10), gab.firstNonce.Value)
		require.True(t, gab.lastNonce.HasValue)
		require.Equal(t, uint64(23), gab.lastNonce.Value)

		// remove breadcrumb with greater nonce
		shouldDelete, err = gab.updateOnRemoveAccountBreadcrumbOnProposedBlock(breadcrumb5)
		require.NoError(t, err)
		require.False(t, shouldDelete)

		require.Equal(t, big.NewInt(35), gab.consumedBalance)
		require.True(t, gab.firstNonce.HasValue)
		require.Equal(t, uint64(10), gab.firstNonce.Value)
		require.True(t, gab.lastNonce.HasValue)
		require.Equal(t, uint64(23), gab.lastNonce.Value)

		// now, remove the breadcrumb with nonce 1
		shouldDelete, err = gab.updateOnRemoveAccountBreadcrumbOnProposedBlock(breadcrumb1)
		require.NoError(t, err)
		require.False(t, shouldDelete)

		require.Equal(t, big.NewInt(25), gab.consumedBalance)
		require.False(t, gab.firstNonce.HasValue)
		require.False(t, gab.lastNonce.HasValue)

		shouldDelete, err = gab.updateOnRemoveAccountBreadcrumbOnProposedBlock(breadcrumb2)
		require.NoError(t, err)
		require.False(t, shouldDelete)

		require.Equal(t, big.NewInt(20), gab.consumedBalance)
		require.False(t, gab.firstNonce.HasValue)
		require.False(t, gab.lastNonce.HasValue)

		shouldDelete, err = gab.updateOnRemoveAccountBreadcrumbOnProposedBlock(breadcrumb3)
		require.NoError(t, err)
		require.False(t, shouldDelete)

		require.Equal(t, big.NewInt(10), gab.consumedBalance)
		require.False(t, gab.firstNonce.HasValue)
		require.False(t, gab.lastNonce.HasValue)

		// now, if we would delete the breadcrumb 0, we should receive that the account can be deleted from the global map
		shouldDelete, err = gab.updateOnRemoveAccountBreadcrumbOnProposedBlock(breadcrumb0)
		require.NoError(t, err)
		require.True(t, shouldDelete)
	})
}

func Test_reduceConsumedBalance(t *testing.T) {
	t.Parallel()

	t.Run("should fail", func(t *testing.T) {
		t.Parallel()

		gab := newGlobalAccountBreadcrumb()
		breadcrumb := accountBreadcrumb{
			consumedBalance: big.NewInt(10),
		}

		err := gab.reduceConsumedBalance(&breadcrumb)
		require.Equal(t, errNegativeBalanceForBreadcrumb, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		gab := newGlobalAccountBreadcrumb()
		gab.consumedBalance = big.NewInt(10)
		breadcrumb := accountBreadcrumb{
			consumedBalance: big.NewInt(5),
		}

		err := gab.reduceConsumedBalance(&breadcrumb)
		require.Nil(t, err)
	})
}
