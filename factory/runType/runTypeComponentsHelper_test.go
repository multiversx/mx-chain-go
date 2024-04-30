package runType

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadInitialAccounts(t *testing.T) {
	t.Parallel()

	t.Run("empty path should error", func(t *testing.T) {
		accounts, err := ReadInitialAccounts("")
		require.Nil(t, accounts)
		require.Error(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		accounts, err := ReadInitialAccounts("../../cmd/node/config/genesis.json")
		require.Nil(t, err)
		require.NotNil(t, accounts)
	})
}
