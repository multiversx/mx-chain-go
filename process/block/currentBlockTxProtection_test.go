package block

import (
	"testing"

	coreBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
)

func TestBaseProcessor_ProtectBlockTransactionsAgainstEviction(t *testing.T) {
	t.Parallel()

	type call struct {
		hashes  [][]byte
		cacheID string
	}
	calls := make([]call, 0)
	releaseCount := 0
	transactionsPool := &testscommon.ShardedDataStub{
		ProtectSetOfDataAgainstEvictionForCurrentBlockCalled: func(hashes [][]byte, cacheID string) {
			calls = append(calls, call{hashes: hashes, cacheID: cacheID})
		},
		ReleaseCurrentBlockTxProtectionCalled: func() {
			releaseCount++
		},
	}
	dataPool := dataRetrieverMock.NewPoolsHolderMock()
	dataPool.SetTransactions(transactionsPool)
	processor := &baseProcessor{dataPool: dataPool}
	txHashes := [][]byte{[]byte("tx1"), []byte("tx2")}

	release := processor.protectBlockTransactionsAgainstEviction(&coreBlock.Body{MiniBlocks: []*coreBlock.MiniBlock{
		{SenderShardID: 0, ReceiverShardID: 1, Type: coreBlock.TxBlock, TxHashes: txHashes},
		{SenderShardID: 0, ReceiverShardID: 1, Type: coreBlock.InvalidBlock, TxHashes: txHashes},
		{SenderShardID: 0, ReceiverShardID: 1, Type: coreBlock.SmartContractResultBlock, TxHashes: txHashes},
	}})

	require.Equal(t, []call{{txHashes, "0_1"}, {txHashes, "0_1"}}, calls)
	release()
	require.Equal(t, 2, releaseCount)
}
