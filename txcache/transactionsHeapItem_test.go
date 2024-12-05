package txcache

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/multiversx/mx-chain-storage-go/types"
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
		require.Nil(t, item.senderState)
		require.Equal(t, 0, item.currentTransactionIndex)
		require.Equal(t, bunch[0], item.currentTransaction)
		require.Equal(t, uint64(42), item.currentTransactionNonce)
		require.Nil(t, item.latestSelectedTransaction)
		require.Equal(t, big.NewInt(0), item.consumedBalance)
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
	require.Equal(t, "50000000000000", item.consumedBalance.String())

	ok := item.gotoNextTransaction()
	require.True(t, ok)

	selected = item.selectCurrentTransaction()
	require.Equal(t, b, selected)
	require.Equal(t, b, item.latestSelectedTransaction)
	require.Equal(t, 43, int(item.latestSelectedTransactionNonce))
	require.Equal(t, "100000000000000", item.consumedBalance.String())

	ok = item.gotoNextTransaction()
	require.False(t, ok)
}

func TestTransactionsHeapItem_detectInitialGap(t *testing.T) {
	a := createTx([]byte("tx-1"), "alice", 42)
	b := createTx([]byte("tx-2"), "alice", 43)

	t.Run("unknown", func(t *testing.T) {
		item, err := newTransactionsHeapItem(bunchOfTransactions{a, b})
		require.NoError(t, err)

		require.False(t, item.detectInitialGap())
	})

	t.Run("known, without gap", func(t *testing.T) {
		item, err := newTransactionsHeapItem(bunchOfTransactions{a, b})
		require.NoError(t, err)

		item.senderState = &types.AccountState{
			Nonce: 42,
		}

		require.False(t, item.detectInitialGap())
	})

	t.Run("known, without gap", func(t *testing.T) {
		item, err := newTransactionsHeapItem(bunchOfTransactions{a, b})
		require.NoError(t, err)

		item.senderState = &types.AccountState{
			Nonce: 41,
		}

		require.True(t, item.detectInitialGap())
	})
}

func TestTransactionsHeapItem_detectMiddleGap(t *testing.T) {
	a := createTx([]byte("tx-1"), "alice", 42)
	b := createTx([]byte("tx-2"), "alice", 43)
	c := createTx([]byte("tx-3"), "alice", 44)

	t.Run("unknown", func(t *testing.T) {
		item := &transactionsHeapItem{}
		item.latestSelectedTransaction = nil
		require.False(t, item.detectInitialGap())
	})

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

func TestTransactionsHeapItem_detectWillFeeExceedBalance(t *testing.T) {
	host := txcachemocks.NewMempoolHostMock()

	a := createTx([]byte("tx-1"), "alice", 42)
	b := createTx([]byte("tx-2"), "alice", 43)
	c := createTx([]byte("tx-3"), "alice", 44).withValue(big.NewInt(1000000000000000000))
	d := createTx([]byte("tx-4"), "alice", 45)

	a.precomputeFields(host)
	b.precomputeFields(host)
	c.precomputeFields(host)
	d.precomputeFields(host)

	t.Run("unknown", func(t *testing.T) {
		item, err := newTransactionsHeapItem(bunchOfTransactions{a, b})

		require.NoError(t, err)
		require.False(t, item.detectWillFeeExceedBalance())
	})

	t.Run("known, not exceeded, then exceeded (a)", func(t *testing.T) {
		item, err := newTransactionsHeapItem(bunchOfTransactions{a, b})
		require.NoError(t, err)

		item.senderState = &types.AccountState{
			Balance: big.NewInt(50000000000001),
		}

		require.False(t, item.detectWillFeeExceedBalance())

		_ = item.selectCurrentTransaction()
		_ = item.gotoNextTransaction()

		require.Equal(t, "50000000000000", item.consumedBalance.String())
		require.True(t, item.detectWillFeeExceedBalance())
	})

	t.Run("known, not exceeded, then exceeded (b)", func(t *testing.T) {
		item, err := newTransactionsHeapItem(bunchOfTransactions{a, b, c, d})
		require.NoError(t, err)

		item.senderState = &types.AccountState{
			Balance: big.NewInt(1000000000000000000 + 2*50000000000000 + 1),
		}

		require.False(t, item.detectWillFeeExceedBalance())

		// Select "a", move to "b".
		_ = item.selectCurrentTransaction()
		_ = item.gotoNextTransaction()

		require.Equal(t, "50000000000000", item.consumedBalance.String())
		require.False(t, item.detectWillFeeExceedBalance())

		// Select "b", move to "c".
		_ = item.selectCurrentTransaction()
		_ = item.gotoNextTransaction()

		require.Equal(t, "100000000000000", item.consumedBalance.String())
		require.False(t, item.detectWillFeeExceedBalance())

		// Select "c", move to "d".
		_ = item.selectCurrentTransaction()
		_ = item.gotoNextTransaction()

		require.Equal(t, "1000150000000000000", item.consumedBalance.String())
		require.True(t, item.detectWillFeeExceedBalance())
	})
}

func TestTransactionsHeapItem_detectLowerNonce(t *testing.T) {
	a := createTx([]byte("tx-1"), "alice", 42)
	b := createTx([]byte("tx-2"), "alice", 43)

	t.Run("unknown", func(t *testing.T) {
		item, err := newTransactionsHeapItem(bunchOfTransactions{a, b})
		require.NoError(t, err)

		require.False(t, item.detectInitialGap())
	})

	t.Run("known, good", func(t *testing.T) {
		item, err := newTransactionsHeapItem(bunchOfTransactions{a, b})
		require.NoError(t, err)

		item.senderState = &types.AccountState{
			Nonce: 42,
		}

		require.False(t, item.detectLowerNonce())
	})

	t.Run("known, lower", func(t *testing.T) {
		item, err := newTransactionsHeapItem(bunchOfTransactions{a, b})
		require.NoError(t, err)

		item.senderState = &types.AccountState{
			Nonce: 44,
		}

		require.True(t, item.detectLowerNonce())
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
		session.IsIncorrectlyGuardedCalled = func(tx data.TransactionHandler) bool {
			return false
		}

		item, err := newTransactionsHeapItem(bunchOfTransactions{createTx([]byte("tx-1"), "alice", 42)})
		require.NoError(t, err)

		require.False(t, item.detectIncorrectlyGuarded(session))
	})

	t.Run("is incorrectly guarded", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		session.IsIncorrectlyGuardedCalled = func(tx data.TransactionHandler) bool {
			return true
		}

		item, err := newTransactionsHeapItem(bunchOfTransactions{createTx([]byte("tx-1"), "alice", 42)})
		require.NoError(t, err)

		require.True(t, item.detectIncorrectlyGuarded(session))
	})
}

func TestTransactionsHeapItem_requestAccountStateIfNecessary(t *testing.T) {
	session := txcachemocks.NewSelectionSessionMock()

	noncesByAddress := session.AccountStateByAddress
	noncesByAddress["alice"] = &types.AccountState{
		Nonce:   7,
		Balance: big.NewInt(1000000000000000000),
	}
	noncesByAddress["bob"] = &types.AccountState{
		Nonce:   42,
		Balance: big.NewInt(1000000000000000000),
	}

	a := &transactionsHeapItem{
		sender: []byte("alice"),
	}

	b := &transactionsHeapItem{
		sender: []byte("bob"),
	}

	c := &transactionsHeapItem{}

	_ = a.requestAccountStateIfNecessary(session)
	_ = b.requestAccountStateIfNecessary(session)

	require.Equal(t, uint64(7), a.senderState.Nonce)
	require.Equal(t, uint64(42), b.senderState.Nonce)
	require.Nil(t, c.senderState)
}
