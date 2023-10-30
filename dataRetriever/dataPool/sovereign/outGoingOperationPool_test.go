package sovereign

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewOutGoingOperationPool(t *testing.T) {
	t.Parallel()

	pool := NewOutGoingOperationPool(time.Second)
	require.False(t, pool.IsInterfaceNil())
}

func TestOutGoingOperationsPool_Add_Get_Delete(t *testing.T) {
	t.Parallel()

	pool := NewOutGoingOperationPool(time.Second)

	hash1 := []byte("h1")
	hash2 := []byte("h2")
	hash3 := []byte("h3")

	data1 := []byte("d1")
	data2 := []byte("d2")
	data3 := []byte("d3")

	pool.Add(hash1, data1)
	require.Equal(t, data1, pool.Get(hash1))
	require.Empty(t, pool.Get(hash2))
	require.Empty(t, pool.Get(hash3))

	pool.Add(hash1, data2)
	require.Equal(t, data1, pool.Get(hash1))

	pool.Add(hash2, data2)
	pool.Add(hash3, data3)

	require.Equal(t, data1, pool.Get(hash1))
	require.Equal(t, data2, pool.Get(hash2))
	require.Equal(t, data3, pool.Get(hash3))

	pool.Delete(hash2)
	require.Equal(t, data1, pool.Get(hash1))
	require.Empty(t, pool.Get(hash2))
	require.Equal(t, data3, pool.Get(hash3))

	pool.Delete(hash1)
	pool.Delete(hash1)
	pool.Delete(hash2)
	require.Empty(t, pool.Get(hash1))
	require.Empty(t, pool.Get(hash2))
	require.Equal(t, data3, pool.Get(hash3))
}

func TestOutGoingOperationsPool_GetUnconfirmedOperations(t *testing.T) {
	t.Parallel()

	expiryTime := time.Millisecond * 100
	pool := NewOutGoingOperationPool(expiryTime)

	hash1 := []byte("h1")
	hash2 := []byte("h2")
	hash3 := []byte("h3")

	data1 := []byte("d1")
	data2 := []byte("d2")
	data3 := []byte("d3")

	pool.Add(hash1, data1)
	pool.Add(hash2, data2)
	require.Empty(t, pool.GetUnconfirmedOperations())

	time.Sleep(expiryTime)
	pool.Add(hash3, data3)
	require.Equal(t, [][]byte{data1, data2}, pool.GetUnconfirmedOperations())

	time.Sleep(expiryTime)
	require.Equal(t, [][]byte{data1, data2, data3}, pool.GetUnconfirmedOperations())
}

func TestOutGoingOperationsPool_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	expiryTime := time.Millisecond * 100
	pool := NewOutGoingOperationPool(expiryTime)

	numOperations := 1000
	wg := sync.WaitGroup{}
	wg.Add(numOperations)
	for i := 0; i < numOperations; i++ {

		go func(index int) {
			id := index % 4
			hash := []byte(fmt.Sprintf("hash%d", id))
			data := []byte(fmt.Sprintf("data%d", id))

			switch id {
			case 0:
				pool.Add(hash, data)
			case 1:
				_ = pool.Get(hash)
			case 2:
				pool.Delete(hash)
			case 3:
				_ = pool.GetUnconfirmedOperations()
			default:
				require.Fail(t, "should not get another operation")
			}

			wg.Done()
		}(i)

	}

	wg.Wait()
}
