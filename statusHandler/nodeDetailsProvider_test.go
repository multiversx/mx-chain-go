package statusHandler_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestNewNodeDetailsHandler(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewNodeDetailsProvider()
	assert.NotNil(t, ndh)
}

func TestNodeDetailsProvider_Increment(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewNodeDetailsProvider()
	key1 := "test-key1"
	ndh.SetUInt64Value(key1, 0)
	ndh.Increment(key1)

	retMap, err := ndh.DetailsMap()
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), retMap[key1])
}

func TestNodeDetailsProvider_Decrement(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewNodeDetailsProvider()
	key := "test-key2"
	ndh.SetUInt64Value(key, 2)
	ndh.Decrement(key)

	retMap, err := ndh.DetailsMap()
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), retMap[key])
}

func TestNodeDetailsProvider_SetInt64Value(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewNodeDetailsProvider()
	key := "test-key3"
	value := int64(5)
	ndh.SetInt64Value(key, value)
	ndh.Decrement(key)

	retMap, err := ndh.DetailsMap()
	assert.Nil(t, err)
	assert.Equal(t, value, retMap[key])
}

func TestNodeDetailsProvider_SetStringValue(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewNodeDetailsProvider()
	key := "test-key4"
	value := "value"
	ndh.SetStringValue(key, value)
	ndh.Decrement(key)

	retMap, err := ndh.DetailsMap()
	assert.Nil(t, err)
	assert.Equal(t, value, retMap[key])
}
