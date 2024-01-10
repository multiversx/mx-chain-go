package dataRetriever

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitType_String(t *testing.T) {
	t.Parallel()

	ut := TransactionUnit
	require.Equal(t, "TransactionUnit", ut.String())
	ut = MiniBlockUnit
	require.Equal(t, "MiniBlockUnit", ut.String())
	ut = PeerChangesUnit
	require.Equal(t, "PeerChangesUnit", ut.String())
	ut = BlockHeaderUnit
	require.Equal(t, "BlockHeaderUnit", ut.String())
	ut = MetaBlockUnit
	require.Equal(t, "MetaBlockUnit", ut.String())
	ut = UnsignedTransactionUnit
	require.Equal(t, "UnsignedTransactionUnit", ut.String())
	ut = RewardTransactionUnit
	require.Equal(t, "RewardTransactionUnit", ut.String())
	ut = MetaHdrNonceHashDataUnit
	require.Equal(t, "MetaHdrNonceHashDataUnit", ut.String())
	ut = HeartbeatUnit
	require.Equal(t, "HeartbeatUnit", ut.String())
	ut = BootstrapUnit
	require.Equal(t, "BootstrapUnit", ut.String())
	ut = StatusMetricsUnit
	require.Equal(t, "StatusMetricsUnit", ut.String())
	ut = TxLogsUnit
	require.Equal(t, "TxLogsUnit", ut.String())
	ut = MiniblocksMetadataUnit
	require.Equal(t, "MiniblocksMetadataUnit", ut.String())
	ut = EpochByHashUnit
	require.Equal(t, "EpochByHashUnit", ut.String())
	ut = MiniblockHashByTxHashUnit
	require.Equal(t, "MiniblockHashByTxHashUnit", ut.String())
	ut = ReceiptsUnit
	require.Equal(t, "ReceiptsUnit", ut.String())
	ut = ResultsHashesByTxHashUnit
	require.Equal(t, "ResultsHashesByTxHashUnit", ut.String())
	ut = TrieEpochRootHashUnit
	require.Equal(t, "TrieEpochRootHashUnit", ut.String())
	ut = ESDTSuppliesUnit
	require.Equal(t, "ESDTSuppliesUnit", ut.String())
	ut = RoundHdrHashDataUnit
	require.Equal(t, "RoundHdrHashDataUnit", ut.String())
	ut = UserAccountsUnit
	require.Equal(t, "UserAccountsUnit", ut.String())
	ut = PeerAccountsUnit
	require.Equal(t, "PeerAccountsUnit", ut.String())
	ut = ScheduledSCRsUnit
	require.Equal(t, "ScheduledSCRsUnit", ut.String())

	ut = 200
	require.Equal(t, "ShardHdrNonceHashDataUnit100", ut.String())

	ut = 99
	require.Equal(t, "unknown type 99", ut.String())
}
