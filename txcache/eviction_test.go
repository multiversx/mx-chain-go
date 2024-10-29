package txcache

import (
	"math"
	"testing"

	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func TestTxCache_DoEviction_BecauseOfCount(t *testing.T) {
	config := ConfigSourceMe{
		Name:                          "untitled",
		NumChunks:                     16,
		NumBytesThreshold:             maxNumBytesUpperBound,
		NumBytesPerSenderThreshold:    maxNumBytesPerSenderUpperBound,
		CountThreshold:                2,
		CountPerSenderThreshold:       math.MaxUint32,
		NumSendersToPreemptivelyEvict: 2,
	}
	txGasHandler := txcachemocks.NewTxGasHandlerMock()
	cache, err := NewTxCache(config, txGasHandler)
	require.Nil(t, err)
	require.NotNil(t, cache)

	cache.AddTx(createTx([]byte("hash-alice"), "alice", 1).withGasPrice(1 * oneBillion))
	cache.AddTx(createTx([]byte("hash-bob"), "bob", 1).withGasPrice(1 * oneBillion))
	cache.AddTx(createTx([]byte("hash-carol"), "carol", 1).withGasPrice(3 * oneBillion))

	journal := cache.doEviction()
	require.Equal(t, uint32(2), journal.numTxs)

	// Alice and Bob evicted. Carol still there (better score).
	_, ok := cache.GetByTxHash([]byte("hash-carol"))
	require.True(t, ok)
	require.Equal(t, uint64(1), cache.CountSenders())
	require.Equal(t, uint64(1), cache.CountTx())
}

func TestTxCache_DoEviction_BecauseOfSize(t *testing.T) {
	config := ConfigSourceMe{
		Name:                          "untitled",
		NumChunks:                     16,
		CountThreshold:                math.MaxUint32,
		CountPerSenderThreshold:       math.MaxUint32,
		NumBytesThreshold:             1000,
		NumBytesPerSenderThreshold:    maxNumBytesPerSenderUpperBound,
		NumSendersToPreemptivelyEvict: 2,
	}

	txGasHandler := txcachemocks.NewTxGasHandlerMock()
	cache, err := NewTxCache(config, txGasHandler)
	require.Nil(t, err)
	require.NotNil(t, cache)

	cache.AddTx(createTx([]byte("hash-alice"), "alice", 1).withSize(256).withGasLimit(500000))
	cache.AddTx(createTx([]byte("hash-bob"), "bob", 1).withSize(256).withGasLimit(500000))
	cache.AddTx(createTx([]byte("hash-carol"), "carol", 1).withSize(256).withGasLimit(500000).withGasPrice(1.5 * oneBillion))
	cache.AddTx(createTx([]byte("hash-eve"), "eve", 1).withSize(256).withGasLimit(500000).withGasPrice(3 * oneBillion))

	journal := cache.doEviction()
	require.Equal(t, uint32(2), journal.numTxs)

	// Alice and Bob evicted (lower score). Carol and Eve still there.
	_, ok := cache.GetByTxHash([]byte("hash-carol"))
	require.True(t, ok)
	_, ok = cache.GetByTxHash([]byte("hash-eve"))
	require.True(t, ok)
	require.Equal(t, uint64(2), cache.CountSenders())
	require.Equal(t, uint64(2), cache.CountTx())
}

func TestTxCache_DoEviction_DoesNothingWhenAlreadyInProgress(t *testing.T) {
	config := ConfigSourceMe{
		Name:                          "untitled",
		NumChunks:                     1,
		CountThreshold:                0,
		NumSendersToPreemptivelyEvict: 1,
		NumBytesPerSenderThreshold:    maxNumBytesPerSenderUpperBound,
		CountPerSenderThreshold:       math.MaxUint32,
	}

	txGasHandler := txcachemocks.NewTxGasHandlerMock()
	cache, err := NewTxCache(config, txGasHandler)
	require.Nil(t, err)
	require.NotNil(t, cache)

	cache.AddTx(createTx([]byte("hash-alice"), "alice", uint64(1)))

	_ = cache.isEvictionInProgress.SetReturningPrevious()
	journal := cache.doEviction()
	require.Nil(t, journal)
}

// This seems to be the most reasonable "bad-enough" (not worst) scenario to benchmark:
// 25000 senders with 10 transactions each, with default "NumSendersToPreemptivelyEvict".
// ~1 second on average laptop.
func TestTxCache_AddWithEviction_UniformDistribution_25000x10(t *testing.T) {
	config := ConfigSourceMe{
		Name:                          "untitled",
		NumChunks:                     16,
		EvictionEnabled:               true,
		NumBytesThreshold:             1000000000,
		CountThreshold:                240000,
		NumSendersToPreemptivelyEvict: 1000,
		NumBytesPerSenderThreshold:    maxNumBytesPerSenderUpperBound,
		CountPerSenderThreshold:       math.MaxUint32,
	}

	txGasHandler := txcachemocks.NewTxGasHandlerMock()
	numSenders := 25000
	numTxsPerSender := 10

	cache, err := NewTxCache(config, txGasHandler)
	require.Nil(t, err)
	require.NotNil(t, cache)

	addManyTransactionsWithUniformDistribution(cache, numSenders, numTxsPerSender)

	// Sometimes (due to map iteration non-determinism), more eviction happens - one more step of 100 senders.
	require.LessOrEqual(t, uint32(cache.CountTx()), config.CountThreshold)
	require.GreaterOrEqual(t, uint32(cache.CountTx()), config.CountThreshold-config.NumSendersToPreemptivelyEvict*uint32(numTxsPerSender))
}
