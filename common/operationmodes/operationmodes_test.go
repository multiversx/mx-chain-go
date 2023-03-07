package operationmodes

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckOperationModes(t *testing.T) {
	t.Parallel()

	t.Run("invalid operation mode", func(t *testing.T) {
		t.Parallel()

		res, err := ParseOperationModes("invalid op")
		require.Empty(t, res)
		require.Equal(t, "invalid operation mode <invalid op>", err.Error())

		res, err = ParseOperationModes(fmt.Sprintf("%s,%s", OperationModeDbLookupExtension, "invalid op"))
		require.Empty(t, res)
		require.Equal(t, "invalid operation mode <invalid op>", err.Error())
	})

	t.Run("bad config", func(t *testing.T) {
		t.Parallel()

		res, err := ParseOperationModes(fmt.Sprintf("%s,%s", OperationModeDbLookupExtension, OperationModeHistoricalBalances))
		require.Empty(t, res)
		require.Equal(t, "operation-mode flag cannot contain both db-lookup-extension and historical-balances", err.Error())

		res, err = ParseOperationModes(fmt.Sprintf("%s,%s", OperationModeSnapshotlessObserver, OperationModeHistoricalBalances))
		require.Empty(t, res)
		require.Equal(t, "operation-mode flag cannot contain both snapshotless-observer and historical-balances", err.Error())

		res, err = ParseOperationModes(fmt.Sprintf("%s,%s", OperationModeSnapshotlessObserver, OperationModeFullArchive))
		require.Empty(t, res)
		require.Equal(t, "operation-mode flag cannot contain both snapshotless-observer and full-archive", err.Error())
	})

	t.Run("ok config", func(t *testing.T) {
		t.Parallel()

		res, err := ParseOperationModes(OperationModeDbLookupExtension)
		require.Equal(t, []string{OperationModeDbLookupExtension}, res)
		require.NoError(t, err)

		res, err = ParseOperationModes(fmt.Sprintf("%s,%s", OperationModeDbLookupExtension, OperationModeFullArchive))
		require.Equal(t, []string{OperationModeDbLookupExtension, OperationModeFullArchive}, res)
		require.NoError(t, err)

		res, err = ParseOperationModes("")
		require.Empty(t, res)
		require.NoError(t, err)
	})
}
