package txcache

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestTxCache_selectTransactionsFromBunchesUsingHeap(t *testing.T) {
	sw := core.NewStopWatch()

	t.Run("numSenders = 1000, numTransactions = 1000", func(t *testing.T) {
		bunches := createBunchesOfTransactionsWithUniformDistribution(1000, 1000)

		sw.Start(t.Name())
		merged := selectTransactionsFromBunchesUsingHeap(bunches, 10_000_000_000)
		sw.Stop(t.Name())

		require.Equal(t, 200001, len(merged))
	})

	for name, measurement := range sw.GetMeasurementsMap() {
		fmt.Printf("%fs (%s)\n", measurement, name)
	}
}

func createBunchesOfTransactionsWithUniformDistribution(nSenders int, nTransactionsPerSender int) []BunchOfTransactions {
	bunches := make([]BunchOfTransactions, 0, nSenders)

	for senderTag := 0; senderTag < nSenders; senderTag++ {
		bunch := make(BunchOfTransactions, 0, nTransactionsPerSender)
		sender := createFakeSenderAddress(senderTag)

		for txNonce := nTransactionsPerSender; txNonce > 0; txNonce-- {
			transactionHash := createFakeTxHash(sender, txNonce)
			transaction := createTx(transactionHash, string(sender), uint64(txNonce))
			bunch = append(bunch, transaction)
		}

		bunches = append(bunches, bunch)
	}

	return bunches
}
