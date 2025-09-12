package holders

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlockchainInfo_GetValues(t *testing.T) {
	t.Parallel()

	t.Run("default values", func(t *testing.T) {
		t.Parallel()

		chainInfo := NewBlockchainInfo(nil, nil, 0)
		require.NotNil(t, chainInfo)
		require.Equal(t, uint64(0), chainInfo.GetCurrentNonce())
		require.Nil(t, chainInfo.GetLatestExecutedBlockHash())
	})

	t.Run("same values", func(t *testing.T) {
		t.Parallel()

		chainInfo := NewBlockchainInfo([]byte("blockHash0"), []byte("blockHash1"), 2)
		require.NotNil(t, chainInfo)
		require.Equal(t, uint64(2), chainInfo.GetCurrentNonce())
		require.Equal(t, "blockHash0", string(chainInfo.GetLatestExecutedBlockHash()))
		require.Equal(t, "blockHash1", string(chainInfo.GetLatestCommittedBlockHash()))
	})
}

func TestBlockchainInfo_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var chainInfo *blockchainInfo
	require.True(t, chainInfo.IsInterfaceNil())

	chainInfo = NewBlockchainInfo(nil, nil, 0)
	require.False(t, chainInfo.IsInterfaceNil())
}
