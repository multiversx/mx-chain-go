package txcache

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func TestSelectionSessionWrapper_accumulateConsumedBalance(t *testing.T) {
	host := txcachemocks.NewMempoolHostMock()

	t.Run("when sender is fee payer", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newSelectionSessionWrapper(session)

		a := createTx([]byte("a-7"), "a", 7)
		b := createTx([]byte("a-8"), "a", 8).withValue(oneQuintillionBig)

		a.precomputeFields(host)
		b.precomputeFields(host)

		sessionWrapper.accumulateConsumedBalance(a)
		require.Equal(t, "50000000000000", sessionWrapper.getAccountRecord([]byte("a")).consumedBalance.String())

		sessionWrapper.accumulateConsumedBalance(b)
		require.Equal(t, "1000100000000000000", sessionWrapper.getAccountRecord([]byte("a")).consumedBalance.String())
	})

	t.Run("when relayer is fee payer", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newSelectionSessionWrapper(session)

		a := createTx([]byte("a-7"), "a", 7).withRelayer([]byte("b")).withGasLimit(100_000)
		b := createTx([]byte("a-8"), "a", 8).withValue(oneQuintillionBig).withRelayer([]byte("b")).withGasLimit(100_000)

		a.precomputeFields(host)
		b.precomputeFields(host)

		sessionWrapper.accumulateConsumedBalance(a)
		require.Equal(t, "0", sessionWrapper.getAccountRecord([]byte("a")).consumedBalance.String())
		require.Equal(t, "100000000000000", sessionWrapper.getAccountRecord([]byte("b")).consumedBalance.String())

		sessionWrapper.accumulateConsumedBalance(b)
		require.Equal(t, "1000000000000000000", sessionWrapper.getAccountRecord([]byte("a")).consumedBalance.String())
		require.Equal(t, "200000000000000", sessionWrapper.getAccountRecord([]byte("b")).consumedBalance.String())
	})
}

func TestSelectionSessionWrapper_detectWillFeeExceedBalance(t *testing.T) {
	host := txcachemocks.NewMempoolHostMock()

	t.Run("unknown", func(t *testing.T) {
		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newSelectionSessionWrapper(session)

		a := createTx([]byte("tx-1"), "alice", 42)
		a.precomputeFields(host)

		require.False(t, sessionWrapper.detectWillFeeExceedBalance(a))
	})

	t.Run("will not exceed for (a) and (b), but will exceed for (c)", func(t *testing.T) {
		a := createTx([]byte("tx-1"), "alice", 42)
		b := createTx([]byte("tx-2"), "alice", 43).withValue(oneQuintillionBig)
		c := createTx([]byte("tx-3"), "alice", 44).withValue(oneQuintillionBig)

		a.precomputeFields(host)
		b.precomputeFields(host)
		c.precomputeFields(host)

		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newSelectionSessionWrapper(session)

		session.SetBalance([]byte("alice"), big.NewInt(oneQuintillion+50000000000000+1))
		recordAlice := sessionWrapper.getAccountRecord([]byte("alice"))
		require.Equal(t, "0", recordAlice.consumedBalance.String())
		require.Equal(t, "1000050000000000001", recordAlice.initialBalance.String())

		require.False(t, sessionWrapper.detectWillFeeExceedBalance(a))

		sessionWrapper.accumulateConsumedBalance(a)

		require.Equal(t, "50000000000000", recordAlice.consumedBalance.String())

		// Even though, in reality, that will be an invalid (but executable) transaction (insufficient balance).
		require.False(t, sessionWrapper.detectWillFeeExceedBalance(b))

		sessionWrapper.accumulateConsumedBalance(b)

		require.Equal(t, "1000100000000000000", recordAlice.consumedBalance.String())
		require.True(t, sessionWrapper.detectWillFeeExceedBalance(c))
	})

	t.Run("will not exceed for (a) and (b), but will exceed for (c) (with relayed)", func(t *testing.T) {
		a := createTx([]byte("tx-1"), "alice", 42).withRelayer([]byte("carol")).withGasLimit(100_000)
		b := createTx([]byte("tx-2"), "alice", 43).withValue(oneQuintillionBig).withRelayer([]byte("carol")).withGasLimit(100_000)
		c := createTx([]byte("tx-3"), "alice", 44).withValue(oneQuintillionBig).withRelayer([]byte("carol")).withGasLimit(100_000)

		a.precomputeFields(host)
		b.precomputeFields(host)
		c.precomputeFields(host)

		session := txcachemocks.NewSelectionSessionMock()
		sessionWrapper := newSelectionSessionWrapper(session)

		session.SetBalance([]byte("alice"), big.NewInt(oneQuintillion))
		session.SetBalance([]byte("carol"), big.NewInt(100000000000000+100000000000000+1))
		recordAlice := sessionWrapper.getAccountRecord([]byte("alice"))
		recordCarol := sessionWrapper.getAccountRecord([]byte("carol"))
		require.Equal(t, "0", recordAlice.consumedBalance.String())
		require.Equal(t, "1000000000000000000", recordAlice.initialBalance.String())
		require.Equal(t, "0", recordCarol.consumedBalance.String())
		require.Equal(t, "200000000000001", recordCarol.initialBalance.String())

		require.False(t, sessionWrapper.detectWillFeeExceedBalance(a))

		sessionWrapper.accumulateConsumedBalance(a)

		require.Equal(t, "0", recordAlice.consumedBalance.String())
		require.Equal(t, "100000000000000", recordCarol.consumedBalance.String())

		require.False(t, sessionWrapper.detectWillFeeExceedBalance(b))

		sessionWrapper.accumulateConsumedBalance(b)

		require.Equal(t, "1000000000000000000", recordAlice.consumedBalance.String())
		require.Equal(t, "200000000000000", recordCarol.consumedBalance.String())
		require.True(t, sessionWrapper.detectWillFeeExceedBalance(c))
	})
}

