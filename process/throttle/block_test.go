package throttle_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/throttle"
	"github.com/stretchr/testify/assert"
)

func TestNewBlockSizeThrottle_ShouldWork(t *testing.T) {
	bst, err := throttle.NewBlockSizeThrottle()

	assert.Nil(t, err)
	assert.NotNil(t, bst)
}

func TestBlockSizeThrottle_MaxItemsToAdd(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxItems := uint32(100)
	bst.SetMaxItems(maxItems)
	assert.Equal(t, maxItems, bst.MaxItemsToAdd())
}

func TestBlockSizeThrottle_Add(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxItems := uint32(12000)
	bst.SetMaxItems(maxItems)

	round := uint64(2)
	items := uint32(10000)
	bst.Add(round, items)

	assert.Equal(t, maxItems, bst.MaxItemsInLastItemAdded())
	assert.Equal(t, items, bst.ItemsInLastItemAdded())
	assert.Equal(t, round, bst.RoundInLastItemAdded())
	assert.False(t, bst.SucceedInLastItemAdded())
}

func TestBlockSizeThrottle_SucceedInLastItemAdded(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	round := uint64(2)
	items := uint32(10000)
	bst.Add(round, items)
	assert.False(t, bst.SucceedInLastItemAdded())

	bst.Succeed(round + 1)
	assert.False(t, bst.SucceedInLastItemAdded())

	bst.Succeed(round)
	assert.True(t, bst.SucceedInLastItemAdded())
}

func TestBlockSizeThrottle_SucceedInItemAdded(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	round := uint64(2)
	items := uint32(10000)
	bst.Add(round, items)
	bst.Add(round+1, items)
	assert.False(t, bst.SucceedInItemAdded(0))
	assert.False(t, bst.SucceedInItemAdded(1))

	bst.Succeed(round + 1)
	assert.False(t, bst.SucceedInItemAdded(0))
	assert.True(t, bst.SucceedInItemAdded(1))

	bst.Succeed(round)
	assert.True(t, bst.SucceedInItemAdded(0))
	assert.True(t, bst.SucceedInItemAdded(1))
}

func TestBlockSizeThrottle_ComputeMaxItemsShouldNotSetMaxItemsWhenStatisticListIsEmpty(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	bst.ComputeMaxItems()
	assert.Equal(t, uint32(process.MaxItemsInBlock), bst.MaxItemsToAdd())
}

func TestBlockSizeThrottle_ComputeMaxItemsShouldSetMaxItemsToMaxItemsInBlockWhenLastActionSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	lastActionMaxItems := uint32(process.MaxItemsInBlock)
	bst.SetMaxItems(lastActionMaxItems)
	bst.Add(2, 0)
	bst.SetSucceed(2, true)
	bst.ComputeMaxItems()

	assert.Equal(t, uint32(process.MaxItemsInBlock), bst.MaxItemsToAdd())
}

func TestBlockSizeThrottle_ComputeMaxItemsShouldSetMaxItemsToAnIncreasedValueWhenLastActionSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	lastActionMaxItems1 := uint32(12000)
	bst.SetMaxItems(lastActionMaxItems1)
	bst.Add(2, 0)
	bst.SetSucceed(2, true)
	bst.ComputeMaxItems()
	increasedValue := lastActionMaxItems1 + uint32(float32(process.MaxItemsInBlock-lastActionMaxItems1)*bst.JumpAboveFactor())
	assert.Equal(t, increasedValue, bst.MaxItemsToAdd())

	bst.SetSucceed(2, false)
	lastActionMaxItems2 := uint32(10000)
	bst.SetMaxItems(lastActionMaxItems2)
	bst.Add(3, 0)
	bst.SetSucceed(3, true)
	bst.ComputeMaxItems()
	increasedValue = lastActionMaxItems2 + uint32(float32(lastActionMaxItems1-lastActionMaxItems2)*bst.JumpAboveFactor())
	assert.Equal(t, increasedValue, bst.MaxItemsToAdd())

	bst.SetSucceed(2, true)
	bst.ComputeMaxItems()
	increasedValue = lastActionMaxItems2 + uint32(float32(process.MaxItemsInBlock-lastActionMaxItems2)*bst.JumpAboveFactor())
	assert.Equal(t, increasedValue, bst.MaxItemsToAdd())
}

func TestBlockSizeThrottle_ComputeMaxItemsShouldSetMaxItemsToMinItemsInBlockWhenLastActionNotSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	lastActionMaxItems := uint32(process.MinItemsInBlock)
	bst.SetMaxItems(lastActionMaxItems)
	bst.Add(2, 0)
	bst.SetSucceed(2, false)
	bst.ComputeMaxItems()

	assert.Equal(t, uint32(process.MinItemsInBlock), bst.MaxItemsToAdd())
}

func TestBlockSizeThrottle_ComputeMaxItemsShouldSetMaxItemsToADecreasedValueWhenLastActionNotSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	lastActionMaxItems1 := uint32(core.Max(12000, process.MinItemsInBlock))
	bst.SetMaxItems(lastActionMaxItems1)
	bst.Add(2, 0)
	bst.SetSucceed(2, false)
	bst.ComputeMaxItems()
	decreasedValue := lastActionMaxItems1 - uint32(float32(lastActionMaxItems1-process.MinItemsInBlock)*bst.JumpBelowFactor())
	assert.Equal(t, decreasedValue, bst.MaxItemsToAdd())

	bst.SetSucceed(2, true)
	lastActionMaxItems2 := uint32(core.Max(14000, process.MinItemsInBlock))
	bst.SetMaxItems(lastActionMaxItems2)
	bst.Add(3, 0)
	bst.SetSucceed(3, false)
	bst.ComputeMaxItems()
	decreasedValue = lastActionMaxItems2 - uint32(float32(lastActionMaxItems2-lastActionMaxItems1)*bst.JumpBelowFactor())
	assert.Equal(t, decreasedValue, bst.MaxItemsToAdd())

	bst.SetSucceed(2, false)
	bst.ComputeMaxItems()
	decreasedValue = lastActionMaxItems2 - uint32(float32(lastActionMaxItems2-process.MinItemsInBlock)*bst.JumpBelowFactor())
	assert.Equal(t, decreasedValue, bst.MaxItemsToAdd())
}

func TestBlockSizeThrottle_GetMaxItemsWhenSucceedShouldReturnMaxItemsInBlock(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxItemsWhenSucceed := bst.GetMaxItemsWhenSucceed(process.MaxItemsInBlock)
	assert.Equal(t, uint32(process.MaxItemsInBlock), maxItemsWhenSucceed)
}

func TestBlockSizeThrottle_GetMaxItemsWhenSucceedShouldReturnNoOfMaxItemsUsedWithoutSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxItemsUsedWithoutSucceed := uint32(14000)
	bst.SetMaxItems(maxItemsUsedWithoutSucceed)
	bst.Add(2, 0)
	jumpAbovePercent := float32(bst.JumpAbovePercent()+1) / float32(100)
	maxItemsWhenSucceed := bst.GetMaxItemsWhenSucceed(uint32(float32(maxItemsUsedWithoutSucceed) * jumpAbovePercent))

	assert.Equal(t, maxItemsUsedWithoutSucceed, maxItemsWhenSucceed)
}

func TestBlockSizeThrottle_GetMaxItemsWhenSucceedShouldIncreaseMaxItemsWithAtLeastOneUnit(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxItemsUsedWithoutSucceed := uint32(core.Min(process.MinItemsInBlock+1, process.MaxItemsInBlock))
	bst.SetMaxItems(maxItemsUsedWithoutSucceed)
	bst.Add(2, 0)
	maxItemsWhenSucceed := bst.GetMaxItemsWhenSucceed(process.MinItemsInBlock)

	assert.Equal(t, uint32(core.Min(process.MinItemsInBlock+1, process.MaxItemsInBlock)), maxItemsWhenSucceed)
}

func TestBlockSizeThrottle_GetMaxItemsWhenSucceedShouldIncreaseMaxItems(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxItemsUsedWithoutSucceed := uint32(14000)
	bst.SetMaxItems(maxItemsUsedWithoutSucceed)
	bst.Add(2, 0)
	lastActionMaxItems := uint32(float32(maxItemsUsedWithoutSucceed) * 0.5)
	maxItemsWhenSucceed := bst.GetMaxItemsWhenSucceed(lastActionMaxItems)
	increasedValue := lastActionMaxItems + uint32(float32(maxItemsUsedWithoutSucceed-lastActionMaxItems)*0.5)

	assert.Equal(t, increasedValue, maxItemsWhenSucceed)
}

func TestBlockSizeThrottle_GetCloserAboveMaxItemsUsedWithoutSucceedShouldReturnMaxItemsInBlock(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	assert.Equal(t, uint32(process.MaxItemsInBlock), bst.GetCloserAboveMaxItemsUsedWithoutSucceed(13999))

	maxItemsUsedWithoutSucceed := uint32(14000)
	bst.SetMaxItems(maxItemsUsedWithoutSucceed)
	bst.Add(2, 0)
	assert.Equal(t, uint32(process.MaxItemsInBlock), bst.GetCloserAboveMaxItemsUsedWithoutSucceed(14000))

	bst.SetSucceed(2, true)
	assert.Equal(t, uint32(process.MaxItemsInBlock), bst.GetCloserAboveMaxItemsUsedWithoutSucceed(13999))
}

func TestBlockSizeThrottle_GetCloserAboveMaxItemsUsedWithoutSucceedShouldReturnMaxItemsUsedWithoutSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxItemsUsedWithoutSucceed := uint32(14000)
	bst.SetMaxItems(maxItemsUsedWithoutSucceed)
	bst.Add(2, 0)
	assert.Equal(t, maxItemsUsedWithoutSucceed, bst.GetCloserAboveMaxItemsUsedWithoutSucceed(13999))
}

func TestBlockSizeThrottle_GetMaxItemsWhenNotSucceedShouldReturnMinItemsInBlock(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxItemsWhenSucceed := bst.GetMaxItemsWhenNotSucceed(process.MinItemsInBlock)
	assert.Equal(t, uint32(process.MinItemsInBlock), maxItemsWhenSucceed)
}

func TestBlockSizeThrottle_GetMaxItemsWhenNotSucceedShouldReturnNoOfMaxItemsUsedWithSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxItemsUsedWithSucceed := uint32(14000)
	bst.SetMaxItems(maxItemsUsedWithSucceed)
	bst.Add(2, 100)
	bst.SetSucceed(2, true)
	jumpBelowPercent := float32(bst.JumpBelowPercent()+1) / float32(100)
	maxItemsWhenNotSucceed := bst.GetMaxItemsWhenNotSucceed(uint32(float32(maxItemsUsedWithSucceed) / jumpBelowPercent))

	assert.Equal(t, maxItemsUsedWithSucceed, maxItemsWhenNotSucceed)
}

func TestBlockSizeThrottle_GetMaxItemsWhenNotSucceedShouldDecreaseMaxItemsWithAtLeastOneUnit(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxItemsUsedWithSucceed := uint32(process.MinItemsInBlock)
	bst.SetMaxItems(maxItemsUsedWithSucceed)
	bst.Add(2, 0)
	bst.SetSucceed(2, true)
	maxItemsWhenNotSucceed := bst.GetMaxItemsWhenNotSucceed(process.MinItemsInBlock + 1)

	assert.Equal(t, uint32(process.MinItemsInBlock), maxItemsWhenNotSucceed)
}

func TestBlockSizeThrottle_GetMaxItemsWhenNotSucceedShouldDecreaseMaxItems(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxItemsUsedWithSucceed := uint32(core.Max(7000, process.MinItemsInBlock))
	bst.SetMaxItems(maxItemsUsedWithSucceed)
	bst.Add(2, 0)
	bst.SetSucceed(2, true)
	lastActionMaxItems := uint32(float32(maxItemsUsedWithSucceed) / 0.5)
	maxItemsWhenNotSucceed := bst.GetMaxItemsWhenNotSucceed(lastActionMaxItems)
	decreasedValue := lastActionMaxItems - uint32(float32(lastActionMaxItems-maxItemsUsedWithSucceed)*0.5)

	assert.Equal(t, decreasedValue, maxItemsWhenNotSucceed)
}

func TestBlockSizeThrottle_GetCloserBelowMaxItemsUsedWithSucceedShouldReturnMinItemsInBlock(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	assert.Equal(t, uint32(process.MinItemsInBlock), bst.GetCloserBelowMaxItemsUsedWithSucceed(14001))

	maxItemsUsedWithSucceed := uint32(14000)
	bst.SetMaxItems(maxItemsUsedWithSucceed)
	bst.Add(2, 0)
	bst.SetSucceed(2, true)
	assert.Equal(t, uint32(process.MinItemsInBlock), bst.GetCloserBelowMaxItemsUsedWithSucceed(14000))

	bst.SetSucceed(2, false)
	assert.Equal(t, uint32(process.MinItemsInBlock), bst.GetCloserBelowMaxItemsUsedWithSucceed(14001))
}

func TestBlockSizeThrottle_GetCloserBelowMaxItemsUsedWithSucceedShouldReturnMaxItemsUsedWithoutSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle()

	maxItemsUsedWithSucceed := uint32(14000)
	bst.SetMaxItems(maxItemsUsedWithSucceed)
	bst.Add(2, 0)
	bst.SetSucceed(2, true)
	assert.Equal(t, maxItemsUsedWithSucceed, bst.GetCloserBelowMaxItemsUsedWithSucceed(14001))
}
