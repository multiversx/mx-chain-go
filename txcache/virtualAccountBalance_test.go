package txcache

import (
	"math/big"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_virtualAccountBalance_ConcurrentAccumulate(t *testing.T) {
	t.Parallel()

	vb, err := newVirtualAccountBalance(big.NewInt(1000))
	require.Nil(t, err)

	var wg sync.WaitGroup
	count := 100
	addVal := big.NewInt(1)

	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			vb.accumulateConsumedBalance(addVal)
		}()
	}
	wg.Wait()

	// Expected consumed: 100 * 1 = 100
	require.Equal(t, big.NewInt(100), vb.getConsumedBalance())

	// Initial should be unchanged
	require.Equal(t, big.NewInt(1000), vb.getInitialBalance())
}

func Test_virtualAccountBalance_ConcurrentReadWrite(t *testing.T) {
	t.Parallel()

	vb, err := newVirtualAccountBalance(big.NewInt(1000))
	require.Nil(t, err)

	var wg sync.WaitGroup
	count := 100

	wg.Add(2 * count)

	// Writers
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			vb.accumulateConsumedBalance(big.NewInt(1))
		}()
	}

	// Readers
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			_ = vb.getConsumedBalance()
			_ = vb.validateBalance()
		}()
	}

	wg.Wait()

	require.Equal(t, big.NewInt(int64(count)), vb.getConsumedBalance())
}

func Test_virtualAccountBalance_ValidateBalance(t *testing.T) {
	t.Parallel()

	// 1. Valid case
	vb, _ := newVirtualAccountBalance(big.NewInt(100))
	vb.accumulateConsumedBalance(big.NewInt(50))
	require.Nil(t, vb.validateBalance())

	// 2. Exact match
	vb.accumulateConsumedBalance(big.NewInt(50))
	require.Nil(t, vb.validateBalance())

	// 3. Exceeded
	vb.accumulateConsumedBalance(big.NewInt(1))
	require.Equal(t, errExceededBalance, vb.validateBalance())
}
