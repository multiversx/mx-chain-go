package statusHandler_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestNewNodeDetailsHandler(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewNodeDetailsHandler()
	assert.NotNil(t, ndh)
}

func TestNodeDetailsHandler_DetailsMapShouldSetValues(t *testing.T) {
	t.Parallel()
	ndh := statusHandler.NewNodeDetailsHandler()

	key := "test-key"
	value := "test-value"
	// set a metric in order to be able to see if they will be exported in the map
	ndh.SetStringValue(key, value)

	retMap, err := ndh.DetailsMap()
	assert.Nil(t, err)
	assert.Equal(t, value, retMap[key])
}
