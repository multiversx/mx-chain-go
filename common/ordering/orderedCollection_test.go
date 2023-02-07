package ordering_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common/ordering"
	"github.com/stretchr/testify/require"
)

var (
	zero  = []byte("zero")
	one   = []byte("one")
	two   = []byte("two")
	three = []byte("three")
)

func TestNewOrderedCollection(t *testing.T) {
	t.Parallel()

	oc := ordering.NewOrderedCollection()
	require.NotNil(t, oc)
	require.Equal(t, 0, oc.Len())
}

func TestOrderedCollection_Add(t *testing.T) {
	t.Parallel()

	oc := ordering.NewOrderedCollection()
	oc.Add(zero)
	require.Equal(t, 1, oc.Len())
	oc.Add(one)
	require.Equal(t, 2, oc.Len())
	oc.Add(two)
	require.Equal(t, 3, oc.Len())
}

func TestOrderedCollection_GetOrder(t *testing.T) {
	t.Parallel()

	oc := ordering.NewOrderedCollection()

	order, err := oc.GetOrder(zero)
	require.Equal(t, ordering.ErrItemNotFound, err)
	require.Zero(t, order)

	oc.Add(zero)
	oc.Add(one)
	oc.Add(two)

	order, err = oc.GetOrder(zero)
	require.Nil(t, err)
	require.Equal(t, 0, order)

	order, err = oc.GetOrder(one)
	require.Nil(t, err)
	require.Equal(t, 1, order)

	order, err = oc.GetOrder(two)
	require.Nil(t, err)
	require.Equal(t, 2, order)
}

func TestOrderedCollection_GetItemAtIndex(t *testing.T) {
	t.Parallel()

	oc := ordering.NewOrderedCollection()

	item, err := oc.GetItemAtIndex(0)
	require.Equal(t, ordering.ErrIndexOutOfBounds, err)
	require.Nil(t, item)

	oc.Add(zero)
	oc.Add(one)
	oc.Add(two)

	item, err = oc.GetItemAtIndex(0)
	require.Nil(t, err)
	require.Equal(t, zero, item)

	item, err = oc.GetItemAtIndex(1)
	require.Nil(t, err)
	require.Equal(t, one, item)

	item, err = oc.GetItemAtIndex(2)
	require.Nil(t, err)
	require.Equal(t, two, item)
}

func TestOrderedCollection_GetItems(t *testing.T) {
	t.Parallel()

	oc := ordering.NewOrderedCollection()

	items := oc.GetItems()
	require.Equal(t, 0, len(items))

	oc.Add(zero)
	oc.Add(one)
	oc.Add(two)

	items = oc.GetItems()
	require.Equal(t, 3, len(items))
	require.Equal(t, zero, items[0])
	require.Equal(t, one, items[1])
	require.Equal(t, two, items[2])
}

func TestOrderedCollection_Remove(t *testing.T) {
	t.Parallel()

	oc := ordering.NewOrderedCollection()
	require.Equal(t, 0, oc.Len())

	oc.Remove(zero)
	require.Equal(t, 0, oc.Len())

	oc.Add(zero)
	require.Equal(t, 1, oc.Len())
	// add duplicate should not add
	oc.Add(zero)
	require.Equal(t, 1, oc.Len())

	oc.Remove(zero)
	require.Equal(t, 0, oc.Len())

	oc.Add(zero)
	oc.Add(one)
	oc.Add(two)
	require.Equal(t, 3, oc.Len())

	oc.Remove(one)
	require.Equal(t, 2, oc.Len())

	order, err := oc.GetOrder(zero)
	require.Nil(t, err)
	require.Equal(t, 0, order)

	elem, err := oc.GetItemAtIndex(uint32(order))
	require.Nil(t, err)
	require.Equal(t, zero, elem)

	_, err = oc.GetOrder(one)
	require.Equal(t, ordering.ErrItemNotFound, err)

	order, err = oc.GetOrder(two)
	require.Nil(t, err)
	require.Equal(t, 1, order)

	elem, err = oc.GetItemAtIndex(uint32(order))
	require.Nil(t, err)
	require.Equal(t, two, elem)

	oc.Remove(zero)
	require.Equal(t, 1, oc.Len())

	oc.Remove(two)
	require.Equal(t, 0, oc.Len())
	elem, err = oc.GetItemAtIndex(uint32(0))
	require.Equal(t, ordering.ErrIndexOutOfBounds, err)
	require.Empty(t, elem)
}

func TestOrderedCollections_RemoveMultiple(t *testing.T) {
	oc := ordering.NewOrderedCollection()
	require.Equal(t, 0, oc.Len())

	oc.RemoveMultiple([][]byte{zero, one, two})
	require.Equal(t, 0, oc.Len())

	oc.Add(zero)
	oc.Add(one)
	oc.Add(two)
	require.Equal(t, 3, oc.Len())

	oc.RemoveMultiple([][]byte{zero, one, two})
	require.Equal(t, 0, oc.Len())

	oc.Add(zero)
	oc.Add(one)
	oc.Add(two)
	require.Equal(t, 3, oc.Len())

	oc.RemoveMultiple([][]byte{one, two})
	require.Equal(t, 1, oc.Len())

	oc.RemoveMultiple([][]byte{zero})
	require.Equal(t, 0, oc.Len())
}

func TestOrderedCollection_Clear(t *testing.T) {
	t.Parallel()

	oc := ordering.NewOrderedCollection()
	require.Equal(t, 0, oc.Len())

	oc.Add(zero)
	oc.Add(one)
	oc.Add(two)
	require.Equal(t, 3, oc.Len())

	oc.Clear()
	require.Equal(t, 0, oc.Len())
}

func TestOrderedCollection_Contains(t *testing.T) {
	t.Parallel()

	oc := ordering.NewOrderedCollection()
	require.Equal(t, 0, oc.Len())

	contains := oc.Contains(zero)
	require.False(t, contains)

	oc.Add(zero)
	oc.Add(one)
	oc.Add(two)
	require.Equal(t, 3, oc.Len())

	contains = oc.Contains(zero)
	require.True(t, contains)

	contains = oc.Contains(one)
	require.True(t, contains)

	contains = oc.Contains(two)
	require.True(t, contains)

	contains = oc.Contains(three)
	require.False(t, contains)
}

func TestOrderedCollection_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	oc := ordering.GetNilOrderedCollection()
	require.True(t, check.IfNil(oc))
}

func TestBaseProcessor_ConcurrentCallsOrderedCollection(t *testing.T) {
	t.Parallel()
	oc := ordering.NewOrderedCollection()

	numCalls := 10000
	wg := &sync.WaitGroup{}
	wg.Add(numCalls)

	numCases := 9
	for i := 0; i < numCalls; i++ {
		go func(idx int) {
			switch idx % numCases {
			case 0:
				for i := 0; i < numCases; i++ {
					oc.Add([]byte(fmt.Sprintf("value_%d", idx+i)))
				}
			case 1:
				_ = oc.GetItems()
			case 2:
				_ = oc.Len()
			case 3:
				for i := 0; i < numCases; i++ {
					_ = oc.Contains([]byte(fmt.Sprintf("value_%d", idx-3+i)))
				}
			case 4:
				for i := 0; i < numCases; i++ {
					_, _ = oc.GetOrder([]byte(fmt.Sprintf("value_%d", idx-4+i)))
				}
			case 5:
				for i := 0; i < numCases; i++ {
					_, _ = oc.GetItemAtIndex(uint32(idx - 5 + i))
				}
			case 6:
				_ = oc.IsInterfaceNil()
			case 7:
				for i := 0; i < numCases; i++ {
					oc.Remove([]byte(fmt.Sprintf("value_%d", idx-7+i)))
				}
			case 8:
				oc.Clear()
			}

			wg.Done()
		}(i)
	}

	wg.Wait()
}
