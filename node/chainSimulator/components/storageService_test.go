package components

import (
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/stretchr/testify/require"
)

func TestCreateStore(t *testing.T) {
	t.Parallel()

	store := CreateStore(2)
	require.NotNil(t, store)

	expectedUnits := []dataRetriever.UnitType{
		dataRetriever.TransactionUnit,
		dataRetriever.MiniBlockUnit,
		dataRetriever.MetaBlockUnit,
		dataRetriever.PeerChangesUnit,
		dataRetriever.BlockHeaderUnit,
		dataRetriever.UnsignedTransactionUnit,
		dataRetriever.RewardTransactionUnit,
		dataRetriever.MetaHdrNonceHashDataUnit,
		dataRetriever.BootstrapUnit,
		dataRetriever.StatusMetricsUnit,
		dataRetriever.ReceiptsUnit,
		dataRetriever.ScheduledSCRsUnit,
		dataRetriever.TxLogsUnit,
		dataRetriever.UserAccountsUnit,
		dataRetriever.PeerAccountsUnit,
		dataRetriever.ESDTSuppliesUnit,
		dataRetriever.RoundHdrHashDataUnit,
		dataRetriever.MiniblocksMetadataUnit,
		dataRetriever.MiniblockHashByTxHashUnit,
		dataRetriever.EpochByHashUnit,
		dataRetriever.ResultsHashesByTxHashUnit,
		dataRetriever.TrieEpochRootHashUnit,
		dataRetriever.ShardHdrNonceHashDataUnit,
		dataRetriever.UnitType(101), // shard 2
	}

	all := store.GetAllStorers()
	require.Equal(t, len(expectedUnits), len(all))

	for i := 0; i < len(expectedUnits); i++ {
		unit, err := store.GetStorer(expectedUnits[i])
		require.NoError(t, err)
		require.NotNil(t, unit)
	}
}
