package txcache

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func TestNewTransactionsHeapItem(t *testing.T) {
	t.Run("empty bunch", func(t *testing.T) {
		item, err := newTransactionsHeapItem(nil)
		require.Nil(t, item)
		require.Equal(t, errEmptyBunchOfTransactions, err)
	})

	t.Run("non-empty bunch", func(t *testing.T) {
		bunch := bunchOfTransactions{
			createTx([]byte("tx-1"), "alice", 42),
		}

		item, err := newTransactionsHeapItem(bunch)
		require.NotNil(t, item)
		require.Nil(t, err)

		require.Equal(t, []byte("alice"), item.sender)
		require.Equal(t, bunch, item.bunch)
		require.Equal(t, 0, item.currentTransactionIndex)
		require.Equal(t, bunch[0], item.currentTransaction)
		require.Equal(t, uint64(42), item.currentTransactionNonce)
		require.Nil(t, item.latestSelectedTransaction)
	})
}

func TestTransactionsHeapItem_selectTransaction(t *testing.T) {
	host := txcachemocks.NewMempoolHostMock()

	a := createTx([]byte("tx-1"), "alice", 42)
	b := createTx([]byte("tx-2"), "alice", 43)
	a.precomputeFields(host)
	b.precomputeFields(host)

	item, err := newTransactionsHeapItem(bunchOfTransactions{a, b})
	require.NoError(t, err)

	selected := item.selectCurrentTransaction()
	require.Equal(t, a, selected)
	require.Equal(t, a, item.latestSelectedTransaction)
	require.Equal(t, 42, int(item.latestSelectedTransactionNonce))

	ok := item.gotoNextTransaction()
	require.True(t, ok)

	selected = item.selectCurrentTransaction()
	require.Equal(t, b, selected)
	require.Equal(t, b, item.latestSelectedTransaction)
	require.Equal(t, 43, int(item.latestSelectedTransactionNonce))

	ok = item.gotoNextTransaction()
	require.False(t, ok)
}

func TestTransactionsHeapItem_detectInitialGap(t *testing.T) {
	a := createTx([]byte("tx-1"), "alice", 42)
	b := createTx([]byte("tx-2"), "alice", 43)

	t.Run("known, without gap", func(t *testing.T) {
		item, err := newTransactionsHeapItem(bunchOfTransactions{a, b})
		require.NoError(t, err)
		require.False(t, item.detectInitialGap(42))
	})

	t.Run("known, without gap", func(t *testing.T) {
		item, err := newTransactionsHeapItem(bunchOfTransactions{a, b})
		require.NoError(t, err)
		require.True(t, item.detectInitialGap(41))
	})
}

func TestTransactionsHeapItem_detectMiddleGap(t *testing.T) {
	a := createTx([]byte("tx-1"), "alice", 42)
	b := createTx([]byte("tx-2"), "alice", 43)
	c := createTx([]byte("tx-3"), "alice", 44)

	t.Run("known, without gap", func(t *testing.T) {
		item := &transactionsHeapItem{}
		item.latestSelectedTransaction = a
		item.latestSelectedTransactionNonce = 42
		item.currentTransaction = b
		item.currentTransactionNonce = 43

		require.False(t, item.detectMiddleGap())
	})

	t.Run("known, without gap", func(t *testing.T) {
		item := &transactionsHeapItem{}
		item.latestSelectedTransaction = a
		item.latestSelectedTransactionNonce = 42
		item.currentTransaction = c
		item.currentTransactionNonce = 44

		require.True(t, item.detectMiddleGap())
	})
}

func TestTransactionsHeapItem_detectLowerNonce(t *testing.T) {
	a := createTx([]byte("tx-1"), "alice", 42)
	b := createTx([]byte("tx-2"), "alice", 43)

	t.Run("known, good", func(t *testing.T) {
		item, err := newTransactionsHeapItem(bunchOfTransactions{a, b})
		require.NoError(t, err)
		require.False(t, item.detectLowerNonce(42))
	})

	t.Run("known, lower", func(t *testing.T) {
		item, err := newTransactionsHeapItem(bunchOfTransactions{a, b})
		require.NoError(t, err)
		require.True(t, item.detectLowerNonce(44))
	})
}

func TestTransactionsHeapItem_detectNonceDuplicate(t *testing.T) {
	a := createTx([]byte("tx-1"), "alice", 42)
	b := createTx([]byte("tx-2"), "alice", 43)
	c := createTx([]byte("tx-3"), "alice", 42)

	t.Run("unknown", func(t *testing.T) {
		item := &transactionsHeapItem{}
		item.latestSelectedTransaction = nil
		require.False(t, item.detectNonceDuplicate())
	})

	t.Run("no duplicates", func(t *testing.T) {
		item := &transactionsHeapItem{}
		item.latestSelectedTransaction = a
		item.latestSelectedTransactionNonce = 42
		item.currentTransaction = b
		item.currentTransactionNonce = 43

		require.False(t, item.detectNonceDuplicate())
	})

	t.Run("duplicates", func(t *testing.T) {
		item := &transactionsHeapItem{}
		item.latestSelectedTransaction = a
		item.latestSelectedTransactionNonce = 42
		item.currentTransaction = c
		item.currentTransactionNonce = 42

		require.True(t, item.detectNonceDuplicate())
	})
}

func TestTransactionsHeapItem_detectIncorrectlyGuarded(t *testing.T) {
	t.Run("is correctly guarded", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newSelectionSessionWrapper(session)

		session.IsIncorrectlyGuardedCalled = func(tx data.TransactionHandler) bool {
			return false
		}

		item, err := newTransactionsHeapItem(bunchOfTransactions{createTx([]byte("tx-1"), "alice", 42)})
		require.NoError(t, err)

		require.False(t, item.detectIncorrectlyGuarded(sessionWrapper))
	})

	t.Run("is incorrectly guarded", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		session.IsIncorrectlyGuardedCalled = func(tx data.TransactionHandler) bool {
			return true
		}
		sessionWrapper := newSelectionSessionWrapper(session)

		item, err := newTransactionsHeapItem(bunchOfTransactions{createTx([]byte("tx-1"), "alice", 42)})
		require.NoError(t, err)

		require.True(t, item.detectIncorrectlyGuarded(sessionWrapper))
	})
}
