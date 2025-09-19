package txcache

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func Test_updateVirtualRecord(t *testing.T) {
	t.Run("breadcrumb doesn't have last nonce", func(t *testing.T) {
		t.Parallel()

		breadcrumb := accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    3,
				HasValue: false,
			},
			lastNonce: core.OptionalUint64{
				Value:    4,
				HasValue: false,
			},
			consumedBalance: big.NewInt(3),
		}

		virtualRecord, err := newVirtualAccountRecord(core.OptionalUint64{
			Value:    1,
			HasValue: true,
		}, big.NewInt(2))
		require.NoError(t, err)

		virtualRecord.updateVirtualRecord(&breadcrumb)
		require.Equal(t, uint64(1), virtualRecord.initialNonce.Value)
	})

	t.Run("virtual record has value for nonce", func(t *testing.T) {
		t.Parallel()

		virtualRecord, err := newVirtualAccountRecord(core.OptionalUint64{
			Value:    1,
			HasValue: true,
		}, big.NewInt(2))
		require.NoError(t, err)

		breadcrumb := accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    3,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    4,
				HasValue: true,
			},
			consumedBalance: big.NewInt(3),
		}

		virtualRecord.updateVirtualRecord(&breadcrumb)

		require.Equal(t, core.OptionalUint64{Value: 5, HasValue: true}, virtualRecord.initialNonce)
		require.Equal(t, big.NewInt(2), virtualRecord.getInitialBalance())
		require.Equal(t, big.NewInt(3), virtualRecord.getConsumedBalance())
	})

	t.Run("virtual record doesn't have value for nonce", func(t *testing.T) {
		t.Parallel()

		virtualRecord, err := newVirtualAccountRecord(core.OptionalUint64{
			Value:    1,
			HasValue: false,
		}, big.NewInt(2))
		require.NoError(t, err)

		breadcrumb := accountBreadcrumb{
			firstNonce: core.OptionalUint64{
				Value:    3,
				HasValue: true,
			},
			lastNonce: core.OptionalUint64{
				Value:    4,
				HasValue: true,
			},
			consumedBalance: big.NewInt(2),
		}

		virtualRecord.updateVirtualRecord(&breadcrumb)

		require.Equal(t, core.OptionalUint64{Value: 5, HasValue: true}, virtualRecord.initialNonce)
		require.Equal(t, big.NewInt(2), virtualRecord.getInitialBalance())
		require.Equal(t, big.NewInt(2), virtualRecord.getConsumedBalance())
	})
}
