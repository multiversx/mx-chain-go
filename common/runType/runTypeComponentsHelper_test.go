package runType

import (
	"math/big"
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
		accounts, err := ReadInitialAccounts("../../genesis/process/testdata/genesisTest1.json")
		require.Nil(t, err)
		require.NotNil(t, accounts)
		require.True(t, len(accounts) > 0)
		require.Equal(t, "a00102030405060708090001020304050607080900010203040506070809000a", accounts[0].GetAddress())
		require.Equal(t, "b00102030405060708090001020304050607080900010203040506070809000b", accounts[1].GetAddress())
		require.Equal(t, "c00102030405060708090001020304050607080900010203040506070809000c", accounts[2].GetAddress())
		require.Equal(t, big.NewInt(10000), accounts[2].GetSupply())
		require.Equal(t, big.NewInt(0), accounts[2].GetBalanceValue())
		require.Equal(t, big.NewInt(0), accounts[2].GetStakingValue())
		require.Equal(t, "00000000000000000500f080f48551abf03e12e27e20d9f077abedffdccc0102", accounts[2].GetDelegationHandler().GetAddress())
		require.Equal(t, big.NewInt(10000), accounts[2].GetDelegationHandler().GetValue())
	})
}
