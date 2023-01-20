package throttle_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process/throttle"
	"github.com/stretchr/testify/assert"
)

const minSizeInBytes = uint32(core.MegabyteSize * 10 / 100)
const maxSizeInBytes = uint32(core.MegabyteSize * 90 / 100)
const testMaxSizeInBytes = maxSizeInBytes / 1000 * 900
const testMediumSizeInBytes = maxSizeInBytes / 1000 * 650
const testHalfSizeInBytes = maxSizeInBytes / 1000 * 450

func TestNewBlockSizeThrottle_ShouldWork(t *testing.T) {
	bst, err := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	assert.Nil(t, err)
	assert.NotNil(t, bst)
}

func TestBlockSizeThrottle_CurrentMaxSizeToAdd(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	maxSize := maxSizeInBytes - 1
	bst.SetCurrentMaxSize(maxSize)
	assert.Equal(t, maxSize, bst.GetCurrentMaxSize())
}

func TestBlockSizeThrottle_Add(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	maxSize := testMaxSizeInBytes
	bst.SetCurrentMaxSize(maxSize)

	round := uint64(2)
	size := testMediumSizeInBytes
	bst.Add(round, size)

	assert.Equal(t, maxSize, bst.CurrentMaxSizeInLastSizeAdded())
	assert.Equal(t, size, bst.SizeInLastSizeAdded())
	assert.Equal(t, round, bst.RoundInLastSizeAdded())
	assert.False(t, bst.SucceedInLastSizeAdded())
}

func TestBlockSizeThrottle_SucceedInLastSizeAdded(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	round := uint64(2)
	size := testMediumSizeInBytes
	bst.Add(round, size)
	assert.False(t, bst.SucceedInLastSizeAdded())

	bst.Succeed(round + 1)
	assert.False(t, bst.SucceedInLastSizeAdded())

	bst.Succeed(round)
	assert.True(t, bst.SucceedInLastSizeAdded())
}

func TestBlockSizeThrottle_SucceedInSizeAdded(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	round := uint64(2)
	size := testMediumSizeInBytes
	bst.Add(round, size)
	bst.Add(round+1, size)
	assert.False(t, bst.SucceedInSizeAdded(0))
	assert.False(t, bst.SucceedInSizeAdded(1))

	bst.Succeed(round + 1)
	assert.False(t, bst.SucceedInSizeAdded(0))
	assert.True(t, bst.SucceedInSizeAdded(1))

	bst.Succeed(round)
	assert.True(t, bst.SucceedInSizeAdded(0))
	assert.True(t, bst.SucceedInSizeAdded(1))
}

func TestBlockSizeThrottle_ComputeCurrentMaxSizeShouldNotSetCurrentMaxSizeWhenStatisticListIsEmpty(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	bst.ComputeCurrentMaxSize()
	assert.Equal(t, maxSizeInBytes, bst.GetCurrentMaxSize())
}

func TestBlockSizeThrottle_ComputeCurrentMaxSizeShouldSetCurrentMaxSizeToMaxSizeWhenLastActionSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	lastActionMaxSize := maxSizeInBytes
	bst.SetCurrentMaxSize(lastActionMaxSize)
	bst.Add(2, minSizeInBytes)
	bst.SetSucceed(2, true)
	bst.ComputeCurrentMaxSize()

	assert.Equal(t, maxSizeInBytes, bst.GetCurrentMaxSize())
}

func TestBlockSizeThrottle_ComputeCurrentMaxSizeShouldSetCurrentMaxSizeToAnIncreasedValueWhenLastActionSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	lastActionMaxSize1 := maxSizeInBytes
	bst.SetCurrentMaxSize(lastActionMaxSize1)
	bst.Add(2, minSizeInBytes)
	bst.SetSucceed(2, true)
	bst.ComputeCurrentMaxSize()
	increasedValue := lastActionMaxSize1 + uint32(float32(maxSizeInBytes-lastActionMaxSize1)*bst.JumpAboveFactor())
	assert.Equal(t, increasedValue, bst.GetCurrentMaxSize())

	bst.SetSucceed(2, false)
	lastActionMaxSize2 := testMediumSizeInBytes
	bst.SetCurrentMaxSize(lastActionMaxSize2)
	bst.Add(3, minSizeInBytes)
	bst.SetSucceed(3, true)
	bst.ComputeCurrentMaxSize()
	increasedValue = lastActionMaxSize2 + uint32(float32(lastActionMaxSize1-lastActionMaxSize2)*bst.JumpAboveFactor())
	assert.Equal(t, increasedValue, bst.GetCurrentMaxSize())

	bst.SetSucceed(2, true)
	bst.ComputeCurrentMaxSize()
	increasedValue = lastActionMaxSize2 + uint32(float32(maxSizeInBytes-lastActionMaxSize2)*bst.JumpAboveFactor())
	assert.Equal(t, increasedValue, bst.GetCurrentMaxSize())
}

func TestBlockSizeThrottle_ComputeCurrentMaxSizeShouldSetCurrentMaxSizeToMinSizeWhenLastActionNotSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	lastActionMaxSize := minSizeInBytes
	bst.SetCurrentMaxSize(lastActionMaxSize)
	bst.Add(2, minSizeInBytes)
	bst.SetSucceed(2, false)
	bst.ComputeCurrentMaxSize()

	assert.Equal(t, minSizeInBytes, bst.GetCurrentMaxSize())
}

func TestBlockSizeThrottle_ComputeCurrentMaxSizeShouldSetCurrentMaxSizeToADecreasedValueWhenLastActionNotSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	lastActionMaxSize1 := core.MaxUint32(testMediumSizeInBytes, minSizeInBytes)
	bst.SetCurrentMaxSize(lastActionMaxSize1)
	bst.Add(2, minSizeInBytes)
	bst.SetSucceed(2, false)
	bst.ComputeCurrentMaxSize()
	decreasedValue := lastActionMaxSize1 - uint32(float32(lastActionMaxSize1-minSizeInBytes)*bst.JumpBelowFactor())
	assert.Equal(t, decreasedValue, bst.GetCurrentMaxSize())

	bst.SetSucceed(2, true)
	lastActionMaxSize2 := core.MaxUint32(testMaxSizeInBytes, minSizeInBytes)
	bst.SetCurrentMaxSize(lastActionMaxSize2)
	bst.Add(3, minSizeInBytes)
	bst.SetSucceed(3, false)
	bst.ComputeCurrentMaxSize()
	decreasedValue = lastActionMaxSize2 - uint32(float32(lastActionMaxSize2-lastActionMaxSize1)*bst.JumpBelowFactor())
	assert.Equal(t, decreasedValue, bst.GetCurrentMaxSize())

	bst.SetSucceed(2, false)
	bst.ComputeCurrentMaxSize()
	decreasedValue = lastActionMaxSize2 - uint32(float32(lastActionMaxSize2-minSizeInBytes)*bst.JumpBelowFactor())
	assert.Equal(t, decreasedValue, bst.GetCurrentMaxSize())
}

func TestBlockSizeThrottle_GetMaxSizeWhenSucceedShouldReturnMaxSize(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	maxSizeWhenSucceed := bst.GetMaxSizeWhenSucceed(maxSizeInBytes)
	assert.Equal(t, maxSizeInBytes, maxSizeWhenSucceed)
}

func TestBlockSizeThrottle_GetMaxSizeWhenSucceedShouldReturnMaxSizeUsedWithoutSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	maxSizeUsedWithoutSucceed := testMaxSizeInBytes
	bst.SetCurrentMaxSize(maxSizeUsedWithoutSucceed)
	bst.Add(2, minSizeInBytes)
	jumpAbovePercent := float32(bst.JumpAbovePercent()+1) / float32(100)
	maxSizeWhenSucceed := bst.GetMaxSizeWhenSucceed(uint32(float32(maxSizeUsedWithoutSucceed) * jumpAbovePercent))

	assert.Equal(t, maxSizeUsedWithoutSucceed, maxSizeWhenSucceed)
}

func TestBlockSizeThrottle_GetMaxSizeWhenSucceedShouldIncreaseCurrentMaxSizeWithAtLeastOneUnit(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	maxSizeUsedWithoutSucceed := core.MinUint32(minSizeInBytes+1, maxSizeInBytes)
	bst.SetCurrentMaxSize(maxSizeUsedWithoutSucceed)
	bst.Add(2, minSizeInBytes)
	maxSizeWhenSucceed := bst.GetMaxSizeWhenSucceed(minSizeInBytes)

	assert.Equal(t, core.MinUint32(minSizeInBytes+1, maxSizeInBytes), maxSizeWhenSucceed)
}

func TestBlockSizeThrottle_GetMaxSizeWhenSucceedShouldIncreaseCurrentMaxSize(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	maxSizeUsedWithoutSucceed := testMaxSizeInBytes
	bst.SetCurrentMaxSize(maxSizeUsedWithoutSucceed)
	bst.Add(2, minSizeInBytes)
	lastActionMaxSize := uint32(float32(maxSizeUsedWithoutSucceed) * 0.5)
	maxSizeWhenSucceed := bst.GetMaxSizeWhenSucceed(lastActionMaxSize)
	increasedValue := lastActionMaxSize + uint32(float32(maxSizeUsedWithoutSucceed-lastActionMaxSize)*0.5)

	assert.Equal(t, increasedValue, maxSizeWhenSucceed)
}

func TestBlockSizeThrottle_GetCloserAboveCurrentMaxSizeUsedWithoutSucceedShouldReturnMaxSize(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	assert.Equal(t, maxSizeInBytes, bst.GetCloserAboveCurrentMaxSizeUsedWithoutSucceed(testMaxSizeInBytes-1))

	maxSizeUsedWithoutSucceed := testMaxSizeInBytes
	bst.SetCurrentMaxSize(maxSizeUsedWithoutSucceed)
	bst.Add(2, minSizeInBytes)
	assert.Equal(t, maxSizeInBytes, bst.GetCloserAboveCurrentMaxSizeUsedWithoutSucceed(testMaxSizeInBytes))

	bst.SetSucceed(2, true)
	assert.Equal(t, maxSizeInBytes, bst.GetCloserAboveCurrentMaxSizeUsedWithoutSucceed(testMaxSizeInBytes-1))
}

func TestBlockSizeThrottle_GetCloserAboveCurrentMaxSizeUsedWithoutSucceedShouldReturnMaxSizeUsedWithoutSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	maxSizeUsedWithoutSucceed := testMaxSizeInBytes
	bst.SetCurrentMaxSize(maxSizeUsedWithoutSucceed)
	bst.Add(2, minSizeInBytes)
	assert.Equal(t, maxSizeUsedWithoutSucceed, bst.GetCloserAboveCurrentMaxSizeUsedWithoutSucceed(testMaxSizeInBytes-1))
}

func TestBlockSizeThrottle_GetMaxSizeWhenNotSucceedShouldReturnMinSize(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	maxSizeWhenSucceed := bst.GetMaxSizeWhenNotSucceed(minSizeInBytes)
	assert.Equal(t, minSizeInBytes, maxSizeWhenSucceed)
}

func TestBlockSizeThrottle_GetMaxSizeWhenNotSucceedShouldReturnMaxSizeUsedWithSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	maxSizeUsedWithSucceed := testMaxSizeInBytes
	bst.SetCurrentMaxSize(maxSizeUsedWithSucceed)
	bst.Add(2, minSizeInBytes)
	bst.SetSucceed(2, true)
	jumpBelowPercent := float32(bst.JumpBelowPercent()+1) / float32(100)
	maxSizeWhenNotSucceed := bst.GetMaxSizeWhenNotSucceed(uint32(float32(maxSizeUsedWithSucceed) / jumpBelowPercent))

	assert.Equal(t, maxSizeUsedWithSucceed, maxSizeWhenNotSucceed)
}

func TestBlockSizeThrottle_GetMaxSizeWhenNotSucceedShouldDecreaseCurrentMaxSizeWithAtLeastOneUnit(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	maxSizeUsedWithSucceed := minSizeInBytes
	bst.SetCurrentMaxSize(maxSizeUsedWithSucceed)
	bst.Add(2, minSizeInBytes)
	bst.SetSucceed(2, true)
	maxSizeWhenNotSucceed := bst.GetMaxSizeWhenNotSucceed(minSizeInBytes + 1)

	assert.Equal(t, minSizeInBytes, maxSizeWhenNotSucceed)
}

func TestBlockSizeThrottle_GetMaxSizeWhenNotSucceedShouldDecreaseCurrentMaxSize(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	maxSizeUsedWithSucceed := core.MaxUint32(testHalfSizeInBytes, minSizeInBytes)
	bst.SetCurrentMaxSize(maxSizeUsedWithSucceed)
	bst.Add(2, minSizeInBytes)
	bst.SetSucceed(2, true)
	lastActionMaxSize := uint32(float32(maxSizeUsedWithSucceed) / 0.5)
	maxSizeWhenNotSucceed := bst.GetMaxSizeWhenNotSucceed(lastActionMaxSize)
	decreasedValue := lastActionMaxSize - uint32(float32(lastActionMaxSize-maxSizeUsedWithSucceed)*0.5)

	assert.Equal(t, decreasedValue, maxSizeWhenNotSucceed)
}

func TestBlockSizeThrottle_GetCloserBelowCurrentMaxSizeUsedWithSucceedShouldReturnMinSize(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	assert.Equal(t, minSizeInBytes, bst.GetCloserBelowCurrentMaxSizeUsedWithSucceed(testMaxSizeInBytes+1))

	maxSizeUsedWithSucceed := testMaxSizeInBytes
	bst.SetCurrentMaxSize(maxSizeUsedWithSucceed)
	bst.Add(2, minSizeInBytes)
	bst.SetSucceed(2, true)
	assert.Equal(t, minSizeInBytes, bst.GetCloserBelowCurrentMaxSizeUsedWithSucceed(testMaxSizeInBytes))

	bst.SetSucceed(2, false)
	assert.Equal(t, minSizeInBytes, bst.GetCloserBelowCurrentMaxSizeUsedWithSucceed(testMaxSizeInBytes+1))
}

func TestBlockSizeThrottle_GetCloserBelowCurrentMaxSizeUsedWithSucceedShouldReturnMaxSizeUsedWithSucceed(t *testing.T) {
	bst, _ := throttle.NewBlockSizeThrottle(minSizeInBytes, maxSizeInBytes)

	maxSizeUsedWithSucceed := testMaxSizeInBytes
	bst.SetCurrentMaxSize(maxSizeUsedWithSucceed)
	bst.Add(2, minSizeInBytes)
	bst.SetSucceed(2, true)
	assert.Equal(t, maxSizeUsedWithSucceed, bst.GetCloserBelowCurrentMaxSizeUsedWithSucceed(testMaxSizeInBytes+1))
}
