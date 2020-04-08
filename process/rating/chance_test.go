package rating

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectionChance_GettersShouldWork(t *testing.T) {
	chancePercent := uint32(1)
	maxThreshold := uint32(2)
	sc := selectionChance{chancePercentage: chancePercent, maxThreshold: maxThreshold}

	assert.Equal(t, chancePercent, sc.GetChancePercentage())
	assert.Equal(t, maxThreshold, sc.GetMaxThreshold())
	assert.False(t, sc.IsInterfaceNil())
}
