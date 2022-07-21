package blockAPI

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBaseAPIBlockProcessor_GetStorageKeyForReceipts(t *testing.T) {
	processor := createBaseBlockProcessor()
	processor.emptyReceiptsHash = []byte("empty-receipts-hash")

	t.Run("when receipts hash is for non-empty receipts", func(t *testing.T) {
		storageKey := processor.getStorageKeyForReceipts([]byte("non-empty-receipts-hash"), []byte("header-hash"))
		require.Equal(t, []byte("non-empty-receipts-hash"), storageKey)
	})

	t.Run("when receipts hash is for empty receipts", func(t *testing.T) {
		storageKey := processor.getStorageKeyForReceipts([]byte("empty-receipts-hash"), []byte("header-hash"))
		require.Equal(t, []byte("header-hash"), storageKey)
	})
}
