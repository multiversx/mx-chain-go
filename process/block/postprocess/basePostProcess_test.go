package postprocess

import (
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/require"
)

func createBaseProcessors(numProcessors int) []*basePostProcessor {
	basePostProcessors := make([]*basePostProcessor, numProcessors)
	for i := 0; i < numProcessors; i++ {
		basePostProcessors[i] = &basePostProcessor{
			hasher:                  sha256.NewSha256(),
			marshalizer:             &marshal.GogoProtoMarshalizer{},
			store:                   nil,
			shardCoordinator:        nil,
			storageType:             0,
			mutInterResultsForBlock: sync.Mutex{},
			interResultsForBlock:    make(map[string]*txInfo),
			mapProcessedResult:      make(map[string]*processedResult),
			intraShardMiniBlock:     nil,
			economicsFee:            nil,
			index:                   0,
		}
	}
	return basePostProcessors
}

func createTxs(hasher hashing.Hasher, num int) ([]data.TransactionHandler, [][]byte) {
	txs := make([]data.TransactionHandler, num)
	txHashes := make([][]byte, num)
	marshaller := &marshal.GogoProtoMarshalizer{}

	for i := 0; i < num; i++ {
		txs[i] = &transaction.Transaction{
			Nonce: uint64(i),
		}
		marshalledTx, _ := marshaller.Marshal(txs[i])
		txHashes[i] = hasher.Compute(string(marshalledTx))
	}

	return txs, txHashes
}

func createScrs(hasher hashing.Hasher, num int) ([]data.TransactionHandler, [][]byte) {
	scrs := make([]data.TransactionHandler, num)
	scrHashes := make([][]byte, num)
	marshaller := &marshal.GogoProtoMarshalizer{}

	for i := 0; i < num; i++ {
		scrs[i] = &smartContractResult.SmartContractResult{
			Nonce:          uint64(i),
			OriginalTxHash: []byte("original tx hash"),
		}
		marshalledTx, _ := marshaller.Marshal(scrs[i])
		scrHashes[i] = hasher.Compute(string(marshalledTx))
	}

	return scrs, scrHashes
}

func TestBasePostProcessor_InitAddAndRemove(t *testing.T) {
	numInstances := 1
	numTxs := 10000
	numScrs := 10000
	headerHash := []byte("headerHash")
	miniBlockHash := []byte("miniBlockHash")
	_, txHashes := createTxs(sha256.NewSha256(), numTxs)
	scrs, scrHash := createScrs(sha256.NewSha256(), numScrs)

	t.Run("InitProcessedResults for header, miniblock and txs, remove one tx, then miniblock", func(t *testing.T) {
		basePreProcs := createBaseProcessors(numInstances)
		basePreProcs[0].InitProcessedResults(headerHash, nil)
		basePreProcs[0].InitProcessedResults(miniBlockHash, headerHash)
		require.Len(t, basePreProcs[0].mapProcessedResult[string(headerHash)].childrenKeys, 1)
		require.Len(t, basePreProcs[0].mapProcessedResult[string(headerHash)].results, 0)

		for i := 0; i < numTxs; i++ {
			basePreProcs[0].InitProcessedResults(txHashes[i], miniBlockHash)
			basePreProcs[0].addIntermediateTxToResultsForBlock(scrs[i], scrHash[i], 0, 1, txHashes[i])
		}
		require.Equal(t, numTxs, len(basePreProcs[0].mapProcessedResult[string(miniBlockHash)].childrenKeys))
		require.Len(t, basePreProcs[0].mapProcessedResult[string(miniBlockHash)].results, 0)

		for i := 0; i < numTxs; i++ {
			require.Len(t, basePreProcs[0].mapProcessedResult[string(txHashes[i])].results, 1)
			require.Len(t, basePreProcs[0].mapProcessedResult[string(txHashes[i])].childrenKeys, 0)
		}

		results := basePreProcs[0].RemoveProcessedResults(txHashes[0])
		require.Len(t, results, 1)

		// miniBlockHash has numTxs-1 childrenKeys, as one was removed
		// each child has one scr registered
		results = basePreProcs[0].RemoveProcessedResults(miniBlockHash)
		require.Len(t, results, numTxs-1)

		// headerHash no longer has childrenKeys and no direct results, so removing it should return an empty slice
		results = basePreProcs[0].RemoveProcessedResults(headerHash)
		require.Len(t, results, 0)
	})
	t.Run("InitProcessedResults for header, miniblock and txs, remove directly the miniblock", func(t *testing.T) {
		basePreProcs := createBaseProcessors(numInstances)
		basePreProcs[0].InitProcessedResults(headerHash, nil)
		basePreProcs[0].InitProcessedResults(miniBlockHash, headerHash)
		require.Len(t, basePreProcs[0].mapProcessedResult[string(headerHash)].childrenKeys, 1)
		require.Len(t, basePreProcs[0].mapProcessedResult[string(headerHash)].results, 0)

		for i := 0; i < numTxs; i++ {
			basePreProcs[0].InitProcessedResults(txHashes[i], miniBlockHash)
			basePreProcs[0].addIntermediateTxToResultsForBlock(scrs[i], scrHash[i], 0, 1, txHashes[i])
		}
		require.Equal(t, numTxs, len(basePreProcs[0].mapProcessedResult[string(miniBlockHash)].childrenKeys))
		require.Len(t, basePreProcs[0].mapProcessedResult[string(miniBlockHash)].results, 0)

		for i := 0; i < numTxs; i++ {
			require.Len(t, basePreProcs[0].mapProcessedResult[string(txHashes[i])].results, 1)
			require.Len(t, basePreProcs[0].mapProcessedResult[string(txHashes[i])].childrenKeys, 0)
		}

		// miniBlockHash has numTxs childrenKeys, each child has one scr registered
		// removing directly the miniBlock should return numTxs results (the scrs)
		results := basePreProcs[0].RemoveProcessedResults(miniBlockHash)
		require.Len(t, results, numTxs)
		for i := 0; i < numTxs; i++ {
			require.Nil(t, basePreProcs[0].mapProcessedResult[string(txHashes[i])])
		}

		// headerHash no longer has childrenKeys and no direct results, so removing it should return an empty slice
		results = basePreProcs[0].RemoveProcessedResults(headerHash)
		require.Len(t, results, 0)
	})

	t.Run("InitProcessedResults for header, miniblock and txs, remove directly the headerhash", func(t *testing.T) {
		basePreProcs := createBaseProcessors(numInstances)
		basePreProcs[0].InitProcessedResults(headerHash, nil)
		basePreProcs[0].InitProcessedResults(miniBlockHash, headerHash)
		require.Len(t, basePreProcs[0].mapProcessedResult[string(headerHash)].childrenKeys, 1)
		require.Len(t, basePreProcs[0].mapProcessedResult[string(headerHash)].results, 0)

		for i := 0; i < numTxs; i++ {
			basePreProcs[0].InitProcessedResults(txHashes[i], miniBlockHash)
			basePreProcs[0].addIntermediateTxToResultsForBlock(scrs[i], scrHash[i], 0, 1, txHashes[i])
		}
		require.Equal(t, numTxs, len(basePreProcs[0].mapProcessedResult[string(miniBlockHash)].childrenKeys))
		require.Len(t, basePreProcs[0].mapProcessedResult[string(miniBlockHash)].results, 0)

		for i := 0; i < numTxs; i++ {
			require.Len(t, basePreProcs[0].mapProcessedResult[string(txHashes[i])].results, 1)
			require.Len(t, basePreProcs[0].mapProcessedResult[string(txHashes[i])].childrenKeys, 0)
		}

		// headerHash has one child, miniBlockHash
		// miniBlockHash has numTxs childrenKeys, each child has one scr registered
		// removing directly the headerHash should return numTxs results (the scrs) for the removed chained childrenKeys
		results := basePreProcs[0].RemoveProcessedResults(headerHash)
		require.Len(t, results, numTxs)
		require.Nil(t, basePreProcs[0].mapProcessedResult[string(miniBlockHash)])

		for i := 0; i < numTxs; i++ {
			require.Nil(t, basePreProcs[0].mapProcessedResult[string(txHashes[i])])
		}
	})
}
