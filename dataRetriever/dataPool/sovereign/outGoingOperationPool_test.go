package sovereign

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/sovereign"
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
	hash4 := []byte("h4")

	data1 := []byte("d1")
	data2 := []byte("d2")
	data3 := []byte("d3")
	data4 := []byte("d4")

	outGoingOperationsHash1 := []byte("h11h22")
	bridgeData1 := &sovereign.BridgeOutGoingData{
		Hash: outGoingOperationsHash1,
		OutGoingOperations: []*sovereign.OutGoingOperation{
			{
				Hash: hash1,
				Data: data1,
			},
			{
				Hash: hash2,
				Data: data2,
			},
		},
	}
	outGoingOperationsHash2 := []byte("h33")
	bridgeData2 := &sovereign.BridgeOutGoingData{
		Hash: outGoingOperationsHash2,
		OutGoingOperations: []*sovereign.OutGoingOperation{
			{
				Hash: hash3,
				Data: data3,
			},
		},
	}
	outGoingOperationsHash3 := []byte("44")
	bridgeData3 := &sovereign.BridgeOutGoingData{
		Hash: outGoingOperationsHash3,
		OutGoingOperations: []*sovereign.OutGoingOperation{
			{
				Hash: hash4,
				Data: data4,
			},
		},
	}

	pool.Add(bridgeData1)
	require.Equal(t, bridgeData1, pool.Get(outGoingOperationsHash1))
	require.Empty(t, pool.Get(outGoingOperationsHash2))
	require.Empty(t, pool.Get(outGoingOperationsHash3))

	pool.Add(bridgeData2)
	require.Equal(t, bridgeData1, pool.Get(outGoingOperationsHash1))
	require.Equal(t, bridgeData2, pool.Get(outGoingOperationsHash2))
	require.Empty(t, pool.Get(outGoingOperationsHash3))

	pool.Add(bridgeData1)
	pool.Add(bridgeData2)
	require.Equal(t, bridgeData1, pool.Get(outGoingOperationsHash1))
	require.Equal(t, bridgeData2, pool.Get(outGoingOperationsHash2))
	require.Empty(t, pool.Get(outGoingOperationsHash3))

	pool.Add(bridgeData3)
	require.Equal(t, bridgeData1, pool.Get(outGoingOperationsHash1))
	require.Equal(t, bridgeData2, pool.Get(outGoingOperationsHash2))
	require.Equal(t, bridgeData3, pool.Get(outGoingOperationsHash3))

	pool.Delete(outGoingOperationsHash2)
	require.Equal(t, bridgeData1, pool.Get(outGoingOperationsHash1))
	require.Empty(t, pool.Get(outGoingOperationsHash2))
	require.Equal(t, bridgeData3, pool.Get(outGoingOperationsHash3))

	pool.Delete(outGoingOperationsHash1)
	pool.Delete(outGoingOperationsHash1)
	pool.Delete(outGoingOperationsHash2)
	require.Empty(t, pool.Get(outGoingOperationsHash1))
	require.Empty(t, pool.Get(outGoingOperationsHash2))
	require.Equal(t, bridgeData3, pool.Get(outGoingOperationsHash3))
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

	outGoingOperationsHash1 := []byte("h11h22")
	bridgeData1 := &sovereign.BridgeOutGoingData{
		Hash: outGoingOperationsHash1,
		OutGoingOperations: []*sovereign.OutGoingOperation{
			{
				Hash: hash1,
				Data: data1,
			},
		},
	}
	outGoingOperationsHash2 := []byte("h33")
	bridgeData2 := &sovereign.BridgeOutGoingData{
		Hash: outGoingOperationsHash2,
		OutGoingOperations: []*sovereign.OutGoingOperation{
			{
				Hash: hash2,
				Data: data2,
			},
		},
	}
	outGoingOperationsHash3 := []byte("44")
	bridgeData3 := &sovereign.BridgeOutGoingData{
		Hash: outGoingOperationsHash3,
		OutGoingOperations: []*sovereign.OutGoingOperation{
			{
				Hash: hash3,
				Data: data3,
			},
		},
	}

	pool.Add(bridgeData1)
	pool.Add(bridgeData2)
	require.Empty(t, pool.GetUnconfirmedOperations())

	time.Sleep(expiryTime)
	pool.Add(bridgeData3)
	require.Equal(t, []*sovereign.BridgeOutGoingData{bridgeData1, bridgeData2}, pool.GetUnconfirmedOperations())

	time.Sleep(expiryTime)
	require.Equal(t, []*sovereign.BridgeOutGoingData{bridgeData1, bridgeData2, bridgeData3}, pool.GetUnconfirmedOperations())
}

func TestOutGoingOperationsPool_ConfirmOperation(t *testing.T) {
	t.Parallel()

	pool := NewOutGoingOperationPool(time.Microsecond)

	hash1 := []byte("h1")
	hash2 := []byte("h2")
	hash3 := []byte("h3")
	hash4 := []byte("h3")

	data1 := []byte("d1")
	data2 := []byte("d2")
	data3 := []byte("d3")
	data4 := []byte("d3")

	outGoingOperationsHash1 := []byte("h11h22")
	outGoingOperationsHash2 := []byte("h33")

	bridgeData1 := &sovereign.BridgeOutGoingData{
		Hash: outGoingOperationsHash1,
		OutGoingOperations: []*sovereign.OutGoingOperation{
			{
				Hash: hash1,
				Data: data1,
			},
			{
				Hash: hash2,
				Data: data2,
			},
			{
				Hash: hash3,
				Data: data3,
			},
		},
	}

	bridgeData2 := &sovereign.BridgeOutGoingData{
		Hash: outGoingOperationsHash2,
		OutGoingOperations: []*sovereign.OutGoingOperation{
			{
				Hash: hash4,
				Data: data4,
			},
		},
	}

	pool.Add(bridgeData1)
	pool.Add(bridgeData2)

	err := pool.ConfirmOperation(outGoingOperationsHash1, hash2)
	require.Nil(t, err)

	err = pool.ConfirmOperation(outGoingOperationsHash1, hash2)
	require.ErrorIs(t, err, errHashOfBridgeOpNotFound)
	require.True(t, strings.Contains(err.Error(), hex.EncodeToString(hash2)))

	bridgeData := pool.Get(outGoingOperationsHash1)
	require.Equal(t, &sovereign.BridgeOutGoingData{
		Hash: outGoingOperationsHash1,
		OutGoingOperations: []*sovereign.OutGoingOperation{
			{
				Hash: hash1,
				Data: data1,
			},
			{
				Hash: hash3,
				Data: data3,
			},
		},
	}, bridgeData)

	err = pool.ConfirmOperation(outGoingOperationsHash1, hash1)
	require.Nil(t, err)

	bridgeData = pool.Get(outGoingOperationsHash1)
	require.Equal(t, &sovereign.BridgeOutGoingData{
		Hash: outGoingOperationsHash1,
		OutGoingOperations: []*sovereign.OutGoingOperation{
			{
				Hash: hash3,
				Data: data3,
			},
		},
	}, bridgeData)

	err = pool.ConfirmOperation(outGoingOperationsHash1, hash3)
	require.Nil(t, err)
	require.Nil(t, pool.Get(outGoingOperationsHash1))

	err = pool.ConfirmOperation(outGoingOperationsHash1, hash1)
	require.ErrorIs(t, err, errHashOfHashesNotFound)
	require.True(t, strings.Contains(err.Error(), hex.EncodeToString(outGoingOperationsHash1)))

	err = pool.ConfirmOperation(outGoingOperationsHash1, hash2)
	require.ErrorIs(t, err, errHashOfHashesNotFound)
	require.True(t, strings.Contains(err.Error(), hex.EncodeToString(outGoingOperationsHash1)))

	err = pool.ConfirmOperation(outGoingOperationsHash1, hash3)
	require.ErrorIs(t, err, errHashOfHashesNotFound)
	require.True(t, strings.Contains(err.Error(), hex.EncodeToString(outGoingOperationsHash1)))

	bridgeData = pool.Get(outGoingOperationsHash2)
	require.Equal(t, &sovereign.BridgeOutGoingData{
		Hash: outGoingOperationsHash2,
		OutGoingOperations: []*sovereign.OutGoingOperation{
			{
				Hash: hash4,
				Data: data4,
			},
		},
	}, bridgeData)

	err = pool.ConfirmOperation(outGoingOperationsHash2, hash4)
	require.Nil(t, err)
	require.Nil(t, pool.Get(outGoingOperationsHash2))

	require.Empty(t, pool.GetUnconfirmedOperations())
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
				pool.Add(&sovereign.BridgeOutGoingData{
					Hash: hash,
					OutGoingOperations: []*sovereign.OutGoingOperation{
						{
							Hash: hash,
							Data: data,
						},
					},
				})
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
