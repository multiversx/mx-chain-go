package transactionLog

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

var (
	providedHash = []byte("hash")
	providedTx   = &transaction.Transaction{Nonce: 123}
	providedLogs = []*vmcommon.LogEntry{
		{
			Identifier: []byte("identifier"),
			Address:    []byte("addr"),
			Topics:     [][]byte{[]byte("topic")},
			Data:       [][]byte{[]byte("data")},
		},
	}
)

func TestNewFailedTxLogsAccumulator(t *testing.T) {
	t.Parallel()

	accumulator := NewFailedTxLogsAccumulator()
	require.NotNil(t, accumulator)
}

func TestFailedTxLogsAccumulator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var accumulator *failedTxLogsAccumulator
	require.True(t, accumulator.IsInterfaceNil())

	accumulator = NewFailedTxLogsAccumulator()
	require.False(t, accumulator.IsInterfaceNil())
}

func TestFailedTxLogsAccumulator_GetLogs(t *testing.T) {
	t.Parallel()

	accumulator := NewFailedTxLogsAccumulator()
	tx, logs, ok := accumulator.GetLogs([]byte(""))
	require.False(t, ok)
	require.Nil(t, tx)
	require.Nil(t, logs)

	err := accumulator.SaveLogs(providedHash, providedTx, providedLogs)
	require.NoError(t, err)

	tx, logs, ok = accumulator.GetLogs([]byte("missing hash"))
	require.False(t, ok)
	require.Nil(t, tx)
	require.Nil(t, logs)

	tx, logs, ok = accumulator.GetLogs(providedHash)
	require.True(t, ok)
	require.Equal(t, providedTx, tx)
	require.Equal(t, providedLogs, logs)
}

func TestFailedTxLogsAccumulator_SaveLogs(t *testing.T) {
	t.Parallel()

	t.Run("empty hash should error", func(t *testing.T) {
		t.Parallel()

		accumulator := NewFailedTxLogsAccumulator()
		err := accumulator.SaveLogs([]byte(""), nil, nil)
		require.Equal(t, process.ErrNilTxHash, err)
	})
	t.Run("nil tx should error", func(t *testing.T) {
		t.Parallel()

		accumulator := NewFailedTxLogsAccumulator()
		err := accumulator.SaveLogs(providedHash, nil, nil)
		require.Equal(t, process.ErrNilTransaction, err)
	})
	t.Run("empty logs should return nil", func(t *testing.T) {
		t.Parallel()

		accumulator := NewFailedTxLogsAccumulator()
		err := accumulator.SaveLogs(providedHash, providedTx, nil)
		require.NoError(t, err)
	})
	t.Run("should work and append logs", func(t *testing.T) {
		t.Parallel()

		accumulator := NewFailedTxLogsAccumulator()
		err := accumulator.SaveLogs(providedHash, providedTx, providedLogs)
		require.NoError(t, err)

		providedNewLogs := []*vmcommon.LogEntry{
			{
				Identifier: []byte("identifier 2"),
				Address:    []byte("addr"),
				Topics:     [][]byte{[]byte("topic 2")},
				Data:       [][]byte{[]byte("data 2")},
			},
		}
		err = accumulator.SaveLogs(providedHash, providedTx, providedNewLogs)
		require.NoError(t, err)

		expectedLogs := append(providedLogs, providedNewLogs...)
		receivedTx, receivedLogs, ok := accumulator.GetLogs(providedHash)
		require.True(t, ok)
		require.Equal(t, providedTx, receivedTx)
		require.Equal(t, expectedLogs, receivedLogs)
	})
}

func TestFailedTxLogsAccumulator_Remove(t *testing.T) {
	t.Parallel()

	accumulator := NewFailedTxLogsAccumulator()
	err := accumulator.SaveLogs(providedHash, providedTx, providedLogs)
	require.NoError(t, err)
	_, _, ok := accumulator.GetLogs(providedHash)
	require.True(t, ok)

	accumulator.Remove([]byte("")) // coverage only

	accumulator.Remove(providedHash)
	_, _, ok = accumulator.GetLogs(providedHash)
	require.False(t, ok)
}

func TestTxLogProcessor_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		accumulator := NewFailedTxLogsAccumulator()

		numCalls := 1000
		wg := sync.WaitGroup{}
		wg.Add(numCalls)

		for i := 0; i < numCalls; i++ {
			go func(idx int) {
				switch idx % 3 {
				case 0:
					err := accumulator.SaveLogs(providedHash, providedTx, []*vmcommon.LogEntry{
						{
							Identifier: []byte(fmt.Sprintf("identifier %d", idx)),
							Address:    []byte("addr"),
							Topics:     [][]byte{[]byte(fmt.Sprintf("topic %d", idx))},
							Data:       [][]byte{[]byte(fmt.Sprintf("data %d", idx))},
						},
					})
					require.NoError(t, err)
				case 1:
					_, _, _ = accumulator.GetLogs(providedHash)
				case 2:
					accumulator.Remove(providedHash)
				}

				wg.Done()
			}(i)
		}

		wg.Wait()
	})
}
