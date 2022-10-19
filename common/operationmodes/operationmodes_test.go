package operationmodes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckOperationModes(t *testing.T) {
	t.Parallel()

	t.Run("bad config", func(t *testing.T) {
		t.Parallel()

		require.Equal(t,
			"operation-mode flag cannot contain both db-lookup-extension and historical-balances",
			CheckOperationModes([]string{OperationModeDbLookupExtension, OperationModeHistoricalBalances}).Error(),
		)

		require.Equal(t,
			"operation-mode flag cannot contain both lite-observer and historical-balances",
			CheckOperationModes([]string{OperationModeLiteObserver, OperationModeHistoricalBalances}).Error(),
		)

		require.Equal(t,
			"operation-mode flag cannot contain both lite-observer and full-archive",
			CheckOperationModes([]string{OperationModeLiteObserver, OperationModeFullArchive}).Error(),
		)

		require.Equal(t,
			"operation-mode flag cannot contain both lite-observer and full-archive",
			CheckOperationModes([]string{OperationModeFullArchive, OperationModeLiteObserver}).Error(),
		)
	})

	t.Run("ok config", func(t *testing.T) {
		t.Parallel()

		require.NoError(t,
			CheckOperationModes([]string{OperationModeDbLookupExtension}),
		)
		require.NoError(t,
			CheckOperationModes([]string{OperationModeDbLookupExtension, OperationModeImportDb}),
		)
	})
}
