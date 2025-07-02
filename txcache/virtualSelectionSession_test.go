package txcache

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func Test_newVirtualSelectionSession(t *testing.T) {
	t.Parallel()

	session := txcachemocks.NewSelectionSessionMock()
	virtualSession := newVirtualSelectionSession(session)
	require.NotNil(t, virtualSession)
}

func Test_updateVirtualRecord(t *testing.T) {
	t.Parallel()

	t.Run("virtual record has value for nonce", func(t *testing.T) {
		virtualRecord := newVirtualAccountRecord(core.OptionalUint64{
			Value:    1,
			HasValue: true,
		}, big.NewInt(2))

		breadcrumb := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
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

		require.Equal(t, virtualRecord.initialNonce, breadcrumb.initialNonce)
		require.Equal(t, virtualRecord.initialBalance, big.NewInt(2))
		require.Equal(t, virtualRecord.consumedBalance, big.NewInt(3))
	})

	t.Run("virtual record doesn't have value for nonce", func(t *testing.T) {
		virtualRecord := newVirtualAccountRecord(core.OptionalUint64{
			Value:    1,
			HasValue: false,
		}, big.NewInt(2))

		breadcrumb := accountBreadcrumb{
			initialNonce: core.OptionalUint64{
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

		require.Equal(t, virtualRecord.initialNonce, breadcrumb.initialNonce)
		require.Equal(t, virtualRecord.initialBalance, big.NewInt(2))
		require.Equal(t, virtualRecord.consumedBalance, big.NewInt(2))
	})
}
