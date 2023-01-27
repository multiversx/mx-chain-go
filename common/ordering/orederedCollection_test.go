package ordering_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common/ordering"
	"github.com/stretchr/testify/require"
)

func TestNewOrderedCollection(t *testing.T) {
	oc := ordering.NewOrderedCollection()
	require.NotNil(t, oc)
	require.Equal(t, 0, oc.Len())
}

func TestOrderedCollection_Add(t *testing.T) {
	oc := ordering.NewOrderedCollection()
	oc.Add("zero")
	require.Equal(t, 1, oc.Len())
	oc.Add("one")
	require.Equal(t, 2, oc.Len())
	oc.Add("two")
	require.Equal(t, 3, oc.Len())
}

func TestOrderedCollection_GetOrder(t *testing.T) {
	oc := ordering.NewOrderedCollection()

	order, err := oc.GetOrder("zero")
	require.Equal(t, ordering.ErrItemNotFound, err)
	require.Zero(t, order)

	oc.Add("zero")
	oc.Add("one")
	oc.Add("two")

	order, err = oc.GetOrder("zero")
	require.Nil(t, err)
	require.Equal(t, 0, order)

	order, err = oc.GetOrder("one")
	require.Nil(t, err)
	require.Equal(t, 1, order)

	order, err = oc.GetOrder("two")
	require.Nil(t, err)
	require.Equal(t, 2, order)
}

func TestOrderedCollection_GetItemAtIndex(t *testing.T) {
	oc := ordering.NewOrderedCollection()

	item, err := oc.GetItemAtIndex(0)
	require.Equal(t, ordering.ErrIndexOutOfBounds, err)
	require.Equal(t, "", item)

	oc.Add("zero")
	oc.Add("one")
	oc.Add("two")

	item, err = oc.GetItemAtIndex(0)
	require.Nil(t, err)
	require.Equal(t, "zero", item)

	item, err = oc.GetItemAtIndex(1)
	require.Nil(t, err)
	require.Equal(t, "one", item)

	item, err = oc.GetItemAtIndex(2)
	require.Nil(t, err)
	require.Equal(t, "two", item)
}

func TestOrderedCollection_GetItems(t *testing.T) {
	oc := ordering.NewOrderedCollection()

	items := oc.GetItems()
	require.Equal(t, 0, len(items))

	oc.Add("zero")
	oc.Add("one")
	oc.Add("two")

	items = oc.GetItems()
	require.Equal(t, 3, len(items))
	require.Equal(t, "zero", items[0])
	require.Equal(t, "one", items[1])
	require.Equal(t, "two", items[2])
}

func TestOrderedCollection_Remove(t *testing.T) {
	oc := ordering.NewOrderedCollection()
	require.Equal(t, 0, oc.Len())

	oc.Remove("zero")
	require.Equal(t, 0, oc.Len())

	oc.Add("zero")
	require.Equal(t, 1, oc.Len())
	// add duplicate should not add
	oc.Add("zero")
	require.Equal(t, 1, oc.Len())

	oc.Remove("zero")
	require.Equal(t, 0, oc.Len())

	oc.Add("zero")
	oc.Add("one")
	oc.Add("two")
	require.Equal(t, 3, oc.Len())

	oc.Remove("one")
	require.Equal(t, 2, oc.Len())

	order, err := oc.GetOrder("zero")
	require.Nil(t, err)
	require.Equal(t, 0, order)

	elem, err := oc.GetItemAtIndex(uint32(order))
	require.Nil(t, err)
	require.Equal(t, "zero", elem)

	_, err = oc.GetOrder("one")
	require.Equal(t, ordering.ErrItemNotFound, err)

	order, err = oc.GetOrder("two")
	require.Nil(t, err)
	require.Equal(t, 1, order)

	elem, err = oc.GetItemAtIndex(uint32(order))
	require.Nil(t, err)
	require.Equal(t, "two", elem)

	oc.Remove("zero")
	require.Equal(t, 1, oc.Len())

	oc.Remove("two")
	require.Equal(t, 0, oc.Len())
	elem, err = oc.GetItemAtIndex(uint32(0))
	require.Equal(t, ordering.ErrIndexOutOfBounds, err)
	require.Empty(t, elem)
}

func TestOrderedCollection_Clear(t *testing.T) {
	oc := ordering.NewOrderedCollection()
	require.Equal(t, 0, oc.Len())

	oc.Add("zero")
	oc.Add("one")
	oc.Add("two")
	require.Equal(t, 3, oc.Len())

	oc.Clear()
	require.Equal(t, 0, oc.Len())
}

func TestOrderedCollection_Contains(t *testing.T) {
	oc := ordering.NewOrderedCollection()
	require.Equal(t, 0, oc.Len())

	contains := oc.Contains("zero")
	require.False(t, contains)

	oc.Add("zero")
	oc.Add("one")
	oc.Add("two")
	require.Equal(t, 3, oc.Len())

	contains = oc.Contains("zero")
	require.True(t, contains)

	contains = oc.Contains("one")
	require.True(t, contains)

	contains = oc.Contains("two")
	require.True(t, contains)

	contains = oc.Contains("three")
	require.False(t, contains)
}

func TestOrderedCollection_IsInterfaceNil(t *testing.T) {
	oc := ordering.GetNilOrderedCollection()
	require.True(t, check.IfNil(oc))
}
